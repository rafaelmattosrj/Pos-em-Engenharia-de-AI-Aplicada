// Package graph orquestra o pipeline RAG sobre o Neo4j — equivalente ao
// StateGraph do LangGraph original (TS) e a RagOrchestrator.java. Fluxo:
// getSchema → generateCypher → executeWithCorrection (self-correction) →
// generateResponse.
package graph

import (
	"context"
	"log"
)

// Neo4jService é satisfeito por neo4jservice.Service — abstraído para
// permitir fakes nos testes de self-correction.
type Neo4jService interface {
	Query(ctx context.Context, cypher string) ([]map[string]any, error)
	GetSchema(ctx context.Context) string
}

// CypherGenerator é satisfeito por service.CypherGeneratorService.
type CypherGenerator interface {
	GenerateCypher(ctx context.Context, question, schema string) (string, error)
	CorrectCypher(ctx context.Context, failedQuery, errMessage, schema string) (string, error)
}

// AnalyticalResponder é satisfeito por service.AnalyticalResponseService.
type AnalyticalResponder interface {
	GenerateResponse(ctx context.Context, question string, results []map[string]any) (string, error)
}

// Orchestrator conduz o pipeline RAG completo.
type Orchestrator struct {
	Neo4j                 Neo4jService
	CypherGenerator       CypherGenerator
	AnalyticalResponse    AnalyticalResponder
	MaxCorrectionAttempts int // default efetivo: 1, igual a @Value("${app.max-correction-attempts:1}")
}

// Query executa o pipeline: schema → geração de Cypher → execução com
// autocorreção → resposta analítica — equivalente a RagOrchestrator.query.
func (o *Orchestrator) Query(ctx context.Context, question string) (string, error) {
	log.Printf("Query: %s\n", question)

	schema := o.Neo4j.GetSchema(ctx)

	cypher, err := o.CypherGenerator.GenerateCypher(ctx, question, schema)
	if err != nil {
		return "", err
	}
	log.Printf("Cypher gerado: %s\n", cypher)

	results := o.executeWithCorrection(ctx, cypher, schema, 0)

	answer, err := o.AnalyticalResponse.GenerateResponse(ctx, question, results)
	if err != nil {
		return "", err
	}
	log.Println("Resposta gerada")
	return answer, nil
}

// executeWithCorrection roda a query e, se falhar, pede ao LLM uma versão
// corrigida até MaxCorrectionAttempts vezes; esgotadas as tentativas, retorna
// resultado vazio — equivalente a RagOrchestrator.executeWithCorrection.
func (o *Orchestrator) executeWithCorrection(ctx context.Context, cypher, schema string, attempt int) []map[string]any {
	results, err := o.Neo4j.Query(ctx, cypher)
	if err == nil {
		log.Printf("Query executada: %d resultados\n", len(results))
		return results
	}
	log.Printf("Erro na query (tentativa %d): %s\n", attempt+1, err)

	if attempt < o.MaxCorrectionAttempts {
		corrected, cerr := o.CypherGenerator.CorrectCypher(ctx, cypher, err.Error(), schema)
		if cerr == nil {
			log.Printf("Query corrigida: %s\n", corrected)
			return o.executeWithCorrection(ctx, corrected, schema, attempt+1)
		}
	}

	log.Println("Retornando resultado vazio após falhas")
	return []map[string]any{}
}
