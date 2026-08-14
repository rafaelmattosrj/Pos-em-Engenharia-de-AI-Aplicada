package memory

import (
	"context"
	"math"
	"sort"
)

// Embedder é a única operação de que ContextualMemory depende — extraída
// como interface para permitir um duplo de teste simples, equivalente ao
// mock de EmbeddingModel em Java.
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float64, error)
}

// EmbeddingEntry é uma entrada da memória contextual: conteúdo + embedding +
// metadados.
type EmbeddingEntry struct {
	Content   string
	Embedding []float64
	Metadata  map[string]any
}

// ContextualMemory armazena conteúdo com embeddings e permite busca
// semântica por similaridade de cosseno — equivalente a
// ContextualMemory.java. Sem persistência em disco: reconstruída a cada
// inicialização, igual à versão Java.
type ContextualMemory struct {
	Embedder            Embedder
	SimilarityThreshold float64
	entries             []EmbeddingEntry
}

// Store armazena conteúdo gerando seu embedding automaticamente.
func (m *ContextualMemory) Store(ctx context.Context, content string, metadata map[string]any) error {
	embedding, err := m.Embedder.Embed(ctx, content)
	if err != nil {
		return err
	}
	m.entries = append(m.entries, EmbeddingEntry{Content: content, Embedding: embedding, Metadata: metadata})
	return nil
}

type scoredEntry struct {
	entry EmbeddingEntry
	score float64
}

// Search busca os topK conteúdos mais semanticamente próximos de query,
// filtrando resultados abaixo do threshold de similaridade configurado.
func (m *ContextualMemory) Search(ctx context.Context, query string, topK int) ([]EmbeddingEntry, error) {
	queryEmbedding, err := m.Embedder.Embed(ctx, query)
	if err != nil {
		return nil, err
	}

	var scored []scoredEntry
	for _, entry := range m.entries {
		score := cosineSimilarity(queryEmbedding, entry.Embedding)
		if score >= m.SimilarityThreshold {
			scored = append(scored, scoredEntry{entry: entry, score: score})
		}
	}

	sort.Slice(scored, func(i, j int) bool { return scored[i].score > scored[j].score })

	if topK > len(scored) {
		topK = len(scored)
	}

	result := make([]EmbeddingEntry, topK)
	for i := 0; i < topK; i++ {
		result[i] = scored[i].entry
	}
	return result, nil
}

func cosineSimilarity(a, b []float64) float64 {
	length := len(a)
	if len(b) < length {
		length = len(b)
	}

	var dotProduct, normA, normB float64
	for i := 0; i < length; i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	denominator := math.Sqrt(normA) * math.Sqrt(normB)
	if denominator == 0 {
		return 0
	}
	return dotProduct / denominator
}
