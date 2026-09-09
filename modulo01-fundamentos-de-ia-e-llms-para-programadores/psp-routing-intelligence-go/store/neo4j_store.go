// Package store encapsula o acesso ao Neo4j como vector store das
// transacoes historicas — equivalente a Neo4jTransactionRepository
// (Neo4jEmbeddingStore + neo4j-java-driver para limpeza de dados) da versao
// Java. Mesmo pacote/padrao de acesso ao Neo4j reaproveitado de
// pdf-rag-knowledge-base-go/store, adaptado para o dominio de transacoes
// (label "Transaction", indice "psp_routing_index") em vez de chunks de PDF.
package store

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"

	"psp-routing-intelligence/domain"
)

const (
	Label     = "Transaction"
	IndexName = "psp_routing_index"
)

// Store e o vector store de transacoes historicas sobre Neo4j.
type Store struct {
	driver neo4j.DriverWithContext
}

// NewStore conecta ao Neo4j em uri com as credenciais informadas.
func NewStore(ctx context.Context, uri, user, password string) (*Store, error) {
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(user, password, ""))
	if err != nil {
		return nil, err
	}
	if err := driver.VerifyConnectivity(ctx); err != nil {
		return nil, fmt.Errorf("conectando ao Neo4j em %s: %w", uri, err)
	}
	return &Store{driver: driver}, nil
}

// Close libera a conexao com o Neo4j.
func (s *Store) Close(ctx context.Context) error {
	return s.driver.Close(ctx)
}

// Clear remove todas as transacoes historicas armazenadas e o indice
// vetorial de uma execucao anterior — chamado antes de (re)popular o seed.
func (s *Store) Clear(ctx context.Context) error {
	if _, err := neo4j.ExecuteQuery(ctx, s.driver,
		fmt.Sprintf("MATCH (n:%s) DETACH DELETE n", Label),
		nil, neo4j.EagerResultTransformer); err != nil {
		return err
	}

	// Indice pode nao existir na primeira execucao — IF EXISTS evita erro.
	_, err := neo4j.ExecuteQuery(ctx, s.driver,
		fmt.Sprintf("DROP INDEX %s IF EXISTS", IndexName),
		nil, neo4j.EagerResultTransformer)
	return err
}

// EnsureIndex cria o indice vetorial (similaridade de cosseno) se ainda nao
// existir, com a dimensao informada.
func (s *Store) EnsureIndex(ctx context.Context, dimension int) error {
	query := fmt.Sprintf(
		"CREATE VECTOR INDEX %s IF NOT EXISTS FOR (n:%s) ON (n.embedding) "+
			"OPTIONS {indexConfig: {`vector.dimensions`: $dim, `vector.similarity_function`: 'cosine'}}",
		IndexName, Label,
	)
	_, err := neo4j.ExecuteQuery(ctx, s.driver, query,
		map[string]any{"dim": dimension}, neo4j.EagerResultTransformer)
	return err
}

// Store persiste uma transacao historica com seu texto natural, embedding e
// os metadados usados para reconstruir um domain.SimilarCase na busca (psp,
// status, amount, method) — equivalente a Neo4jTransactionRepository.store
// da versao Java (la, via Metadata do LangChain4j; aqui, propriedades do
// no Cypher).
func (s *Store) Store(ctx context.Context, tx domain.HistoricalTransaction, text string, embedding []float64) error {
	query := fmt.Sprintf(
		"CREATE (n:%s {text: $text, embedding: $embedding, psp: $psp, status: $status, amount: $amount, method: $method})",
		Label,
	)
	_, err := neo4j.ExecuteQuery(ctx, s.driver, query, map[string]any{
		"text":      text,
		"embedding": embedding,
		"psp":       string(tx.PSP),
		"status":    string(tx.Status),
		"amount":    tx.Amount,
		"method":    string(tx.Method),
	}, neo4j.EagerResultTransformer)
	return err
}

// FindSimilar retorna os topK casos historicos mais similares a
// queryEmbedding por similaridade de cosseno, ordenados do mais para o
// menos similar — equivalente a Neo4jTransactionRepository.findSimilar.
func (s *Store) FindSimilar(ctx context.Context, queryEmbedding []float64, topK int) ([]domain.SimilarCase, error) {
	query := fmt.Sprintf(
		"CALL db.index.vector.queryNodes('%s', $topK, $embedding) "+
			"YIELD node, score "+
			"RETURN node.psp AS psp, node.status AS status, node.amount AS amount, node.method AS method, score "+
			"ORDER BY score DESC",
		IndexName,
	)

	result, err := neo4j.ExecuteQuery(ctx, s.driver, query,
		map[string]any{"topK": topK, "embedding": queryEmbedding}, neo4j.EagerResultTransformer)
	if err != nil {
		return nil, err
	}

	cases := make([]domain.SimilarCase, 0, len(result.Records))
	for _, record := range result.Records {
		pspRaw, _ := record.Get("psp")
		statusRaw, _ := record.Get("status")
		amountRaw, _ := record.Get("amount")
		methodRaw, _ := record.Get("method")
		scoreRaw, _ := record.Get("score")

		psp, err := domain.ParsePSP(pspRaw.(string))
		if err != nil {
			return nil, fmt.Errorf("registro invalido no Neo4j: %w", err)
		}
		status, err := domain.ParseTransactionStatus(statusRaw.(string))
		if err != nil {
			return nil, fmt.Errorf("registro invalido no Neo4j: %w", err)
		}
		method, err := domain.ParsePaymentMethod(methodRaw.(string))
		if err != nil {
			return nil, fmt.Errorf("registro invalido no Neo4j: %w", err)
		}

		cases = append(cases, domain.SimilarCase{
			PSP:        psp,
			Status:     status,
			Amount:     amountRaw.(float64),
			Method:     method,
			Similarity: scoreRaw.(float64),
		})
	}
	return cases, nil
}
