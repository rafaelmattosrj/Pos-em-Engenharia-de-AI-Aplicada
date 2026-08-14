package memory

import (
	"context"
	"strconv"
	"testing"
	"time"

	"agent-memory/model"

	"github.com/google/uuid"
)

func TestLongTermMemory_DoesNotDuplicateFactsWithSameContent(t *testing.T) {
	ltm := NewLongTermMemory(t.TempDir())

	ltm.AddFact(model.MemoryFact{
		ID: uuid.NewString(), Content: "O usuário prefere respostas em português",
		Source: "usuario", ConfirmedAt: time.Now(),
	})
	ltm.AddFact(model.MemoryFact{
		ID: uuid.NewString(), Content: "O usuário prefere respostas em português",
		Source: "sistema", ConfirmedAt: time.Now(),
	})

	facts := ltm.GetFacts()
	if len(facts) != 1 {
		t.Fatalf("esperava 1 fato (deduplicado), obteve %d", len(facts))
	}
	if facts[0].Content != "O usuário prefere respostas em português" {
		t.Errorf("conteudo inesperado: %q", facts[0].Content)
	}
}

type fakeEmbedder struct {
	vectors map[string][]float64
}

func (f *fakeEmbedder) Embed(ctx context.Context, text string) ([]float64, error) {
	return f.vectors[text], nil
}

func TestContextualMemory_FiltersByThreshold(t *testing.T) {
	vetorBase := []float64{1.0, 0.0, 0.0}
	vetorSimilar := []float64{0.95, 0.1, 0.0} // cosseno alto
	vetorDistante := []float64{0.0, 0.0, 1.0} // cosseno ~0.0

	embedder := &fakeEmbedder{vectors: map[string][]float64{
		"conteudo relevante":   vetorSimilar,
		"conteudo irrelevante": vetorDistante,
		"query de busca":       vetorBase,
	}}

	mem := &ContextualMemory{Embedder: embedder, SimilarityThreshold: 0.7}

	ctx := context.Background()
	mem.Store(ctx, "conteudo relevante", map[string]any{"tipo": "teste"})
	mem.Store(ctx, "conteudo irrelevante", map[string]any{"tipo": "teste"})

	results, err := mem.Search(ctx, "query de busca", 10)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("esperava 1 resultado acima do threshold, obteve %d", len(results))
	}
	if results[0].Content != "conteudo relevante" {
		t.Errorf("conteudo inesperado: %q", results[0].Content)
	}
}

func TestEpisodicMemory_GetRecentEpisodesRespectsLimit(t *testing.T) {
	em := NewEpisodicMemory(t.TempDir())

	base := time.Now()
	for i := 1; i <= 5; i++ {
		em.AddEpisode(model.Episode{
			ID: uuid.NewString(), Input: "input " + strconv.Itoa(i),
			Steps: []string{"passo 1", "passo 2"}, Outcome: "resultado " + strconv.Itoa(i),
			Timestamp: base.Add(time.Duration(i) * time.Second),
		})
	}

	recent := em.GetRecentEpisodes(3)

	if len(recent) != 3 {
		t.Fatalf("esperava 3 episodios, obteve %d", len(recent))
	}
	if recent[0].Input != "input 5" || recent[1].Input != "input 4" || recent[2].Input != "input 3" {
		t.Errorf("ordem inesperada: %v", []string{recent[0].Input, recent[1].Input, recent[2].Input})
	}
}
