// Package store encapsula o acesso ao Neo4j como vector store — equivalente
// a Neo4jEmbeddingStore (LangChain4j) + o uso manual do neo4j-java-driver
// para limpeza de dados na versão Java.
package store

import (
	"context"
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

const (
	Label     = "Chunk"
	IndexName = "tensors_index"
)

// Store é um vector store sobre Neo4j, com nós rotulados Label contendo as
// propriedades "text" e "embedding".
type Store struct {
	driver neo4j.DriverWithContext
}

// Match é um resultado de busca por similaridade: o texto do chunk e o score
// de similaridade de cosseno (0-1, maior é mais similar).
type Match struct {
	Text  string
	Score float64
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

// Close libera a conexão com o Neo4j.
func (s *Store) Close(ctx context.Context) error {
	return s.driver.Close(ctx)
}

// Clear remove todos os chunks e o índice vetorial de uma execução anterior.
func (s *Store) Clear(ctx context.Context) error {
	if _, err := neo4j.ExecuteQuery(ctx, s.driver,
		fmt.Sprintf("MATCH (n:%s) DETACH DELETE n", Label),
		nil, neo4j.EagerResultTransformer); err != nil {
		return err
	}

	// Índice pode não existir na primeira execução — IF EXISTS evita erro.
	_, err := neo4j.ExecuteQuery(ctx, s.driver,
		fmt.Sprintf("DROP INDEX %s IF EXISTS", IndexName),
		nil, neo4j.EagerResultTransformer)
	return err
}

// EnsureIndex cria o índice vetorial (similaridade de cosseno) se ainda não
// existir, com a dimensão informada.
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

// Add armazena um chunk de texto com seu vetor de embedding.
func (s *Store) Add(ctx context.Context, text string, embedding []float64) error {
	query := fmt.Sprintf("CREATE (n:%s {text: $text, embedding: $embedding})", Label)
	_, err := neo4j.ExecuteQuery(ctx, s.driver, query,
		map[string]any{"text": text, "embedding": embedding}, neo4j.EagerResultTransformer)
	return err
}

// Search retorna os topK chunks mais próximos de queryEmbedding por
// similaridade de cosseno, ordenados do mais para o menos similar.
func (s *Store) Search(ctx context.Context, queryEmbedding []float64, topK int) ([]Match, error) {
	query := fmt.Sprintf(
		"CALL db.index.vector.queryNodes('%s', $topK, $embedding) "+
			"YIELD node, score RETURN node.text AS text, score ORDER BY score DESC",
		IndexName,
	)

	result, err := neo4j.ExecuteQuery(ctx, s.driver, query,
		map[string]any{"topK": topK, "embedding": queryEmbedding}, neo4j.EagerResultTransformer)
	if err != nil {
		return nil, err
	}

	matches := make([]Match, 0, len(result.Records))
	for _, record := range result.Records {
		text, _ := record.Get("text")
		score, _ := record.Get("score")
		matches = append(matches, Match{
			Text:  text.(string),
			Score: score.(float64),
		})
	}
	return matches, nil
}
