package graph

import (
	"context"
	"errors"
	"testing"
)

type fakeNeo4j struct {
	schema      string
	failFirstN  int
	calls       int
	lastQueries []string
}

func (f *fakeNeo4j) GetSchema(ctx context.Context) string { return f.schema }

func (f *fakeNeo4j) Query(ctx context.Context, cypher string) ([]map[string]any, error) {
	f.calls++
	f.lastQueries = append(f.lastQueries, cypher)
	if f.calls <= f.failFirstN {
		return nil, errors.New("syntax error near WHERE")
	}
	return []map[string]any{{"name": "Ana Silva"}}, nil
}

type fakeCypherGenerator struct {
	generated string
	corrected string
	genErr    error
	corrErr   error
}

func (f *fakeCypherGenerator) GenerateCypher(ctx context.Context, question, schema string) (string, error) {
	return f.generated, f.genErr
}

func (f *fakeCypherGenerator) CorrectCypher(ctx context.Context, failedQuery, errMessage, schema string) (string, error) {
	return f.corrected, f.corrErr
}

type fakeAnalyticalResponder struct {
	response      string
	receivedCount int
}

func (f *fakeAnalyticalResponder) GenerateResponse(ctx context.Context, question string, results []map[string]any) (string, error) {
	f.receivedCount = len(results)
	return f.response, nil
}

func TestQuery_SucessoDePrimeira(t *testing.T) {
	neo4j := &fakeNeo4j{schema: "Labels: [Student, Course]"}
	cypherGen := &fakeCypherGenerator{generated: "MATCH (s:Student) RETURN s"}
	responder := &fakeAnalyticalResponder{response: "Há 3 alunos matriculados."}

	o := &Orchestrator{Neo4j: neo4j, CypherGenerator: cypherGen, AnalyticalResponse: responder, MaxCorrectionAttempts: 1}

	answer, err := o.Query(context.Background(), "quantos alunos existem?")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if answer != "Há 3 alunos matriculados." {
		t.Errorf("resposta = %q, esperado %q", answer, "Há 3 alunos matriculados.")
	}
	if neo4j.calls != 1 {
		t.Errorf("esperava 1 chamada a Query, teve %d", neo4j.calls)
	}
}

func TestQuery_AutocorrecaoAposFalha(t *testing.T) {
	neo4j := &fakeNeo4j{schema: "schema", failFirstN: 1}
	cypherGen := &fakeCypherGenerator{generated: "MATCH (s:Studnt) RETURN s", corrected: "MATCH (s:Student) RETURN s"}
	responder := &fakeAnalyticalResponder{response: "ok"}

	o := &Orchestrator{Neo4j: neo4j, CypherGenerator: cypherGen, AnalyticalResponse: responder, MaxCorrectionAttempts: 1}

	answer, err := o.Query(context.Background(), "pergunta")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if answer != "ok" {
		t.Errorf("resposta = %q, esperado %q", answer, "ok")
	}
	if neo4j.calls != 2 {
		t.Fatalf("esperava 2 chamadas a Query (original + corrigida), teve %d", neo4j.calls)
	}
	if neo4j.lastQueries[1] != "MATCH (s:Student) RETURN s" {
		t.Errorf("segunda query = %q, esperada a corrigida", neo4j.lastQueries[1])
	}
	if responder.receivedCount != 1 {
		t.Errorf("responder deveria receber 1 resultado apos a correcao, recebeu %d", responder.receivedCount)
	}
}

func TestQuery_EsgotaTentativasRetornaVazio(t *testing.T) {
	neo4j := &fakeNeo4j{schema: "schema", failFirstN: 10}
	cypherGen := &fakeCypherGenerator{generated: "MATCH (s:X) RETURN s", corrected: "MATCH (s:X) RETURN s"}
	responder := &fakeAnalyticalResponder{response: "sem dados"}

	o := &Orchestrator{Neo4j: neo4j, CypherGenerator: cypherGen, AnalyticalResponse: responder, MaxCorrectionAttempts: 1}

	answer, err := o.Query(context.Background(), "pergunta")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if answer != "sem dados" {
		t.Errorf("resposta = %q, esperado %q", answer, "sem dados")
	}
	// tentativa original + 1 correcao = 2 chamadas, MaxCorrectionAttempts=1 esgota ali
	if neo4j.calls != 2 {
		t.Errorf("esperava 2 chamadas a Query (esgotando MaxCorrectionAttempts=1), teve %d", neo4j.calls)
	}
	if responder.receivedCount != 0 {
		t.Errorf("responder deveria receber lista vazia apos esgotar tentativas, recebeu %d itens", responder.receivedCount)
	}
}

func TestQuery_ErroNaGeracaoDeCypherPropagaSemChamarNeo4j(t *testing.T) {
	neo4j := &fakeNeo4j{schema: "schema"}
	cypherGen := &fakeCypherGenerator{genErr: errors.New("all fallback models failed")}
	responder := &fakeAnalyticalResponder{response: "nao deveria chegar aqui"}

	o := &Orchestrator{Neo4j: neo4j, CypherGenerator: cypherGen, AnalyticalResponse: responder, MaxCorrectionAttempts: 1}

	_, err := o.Query(context.Background(), "pergunta")
	if err == nil {
		t.Fatal("esperava erro quando a geracao de Cypher falha")
	}
	if neo4j.calls != 0 {
		t.Errorf("Neo4j.Query nao deveria ser chamado se a geracao de Cypher falhou, foi chamado %d vez(es)", neo4j.calls)
	}
}
