package agent

import (
	"context"
	"strings"
	"testing"

	"agent-memory/engine"
	"agent-memory/memory"
	"agent-memory/model"
)

type stubChatClient struct {
	response string
}

func (s *stubChatClient) Chat(ctx context.Context, userPrompt string) (string, error) {
	return s.response, nil
}

type stubEmbedder struct{}

func (stubEmbedder) Embed(ctx context.Context, text string) ([]float64, error) {
	return []float64{1, 0, 0}, nil
}

func newTestAgent(t *testing.T, chatResponse string) *MemoryAwareAgent {
	t.Helper()
	client := &stubChatClient{response: chatResponse}
	return &MemoryAwareAgent{
		LongTerm:   memory.NewLongTermMemory(t.TempDir()),
		Episodic:   memory.NewEpisodicMemory(t.TempDir()),
		Contextual: &memory.ContextualMemory{Embedder: stubEmbedder{}, SimilarityThreshold: 0.7},
		Reflection: engine.NewReflectionEngine(client, t.TempDir()),
		Client:     client,
	}
}

func TestRun_ReturnsOutputAndMarksMemoryUsed(t *testing.T) {
	a := newTestAgent(t, "resposta enriquecida com contexto")

	result, err := a.Run(context.Background(), "qual o status do servico?")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !result.MemoryUsed {
		t.Error("esperava memoryUsed=true")
	}
	if result.Output != "resposta enriquecida com contexto" {
		t.Errorf("output inesperado: %q", result.Output)
	}
}

func TestRun_PersistsEpisode(t *testing.T) {
	a := newTestAgent(t, "resposta qualquer")

	a.Run(context.Background(), "primeira pergunta")
	a.Run(context.Background(), "segunda pergunta")

	episodes := a.Episodic.GetEpisodes()
	if len(episodes) != 2 {
		t.Fatalf("esperava 2 episodios persistidos, obteve %d", len(episodes))
	}
}

func TestRun_SecondCallSeesFirstEpisodeInContext(t *testing.T) {
	a := newTestAgent(t, "ok")

	a.Run(context.Background(), "primeira pergunta")
	result, err := a.Run(context.Background(), "segunda pergunta")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.EpisodesConsidered != 1 {
		t.Errorf("esperava 1 episodio considerado na segunda chamada, obteve %d", result.EpisodesConsidered)
	}
}

func TestBuildEnrichedPrompt_IncludesAllSections(t *testing.T) {
	facts := []model.MemoryFact{{Content: "usuario prefere PT-BR", Source: "usuario"}}
	episodes := []model.Episode{{Input: "pergunta anterior", Outcome: "resposta anterior"}}
	context := []memory.EmbeddingEntry{{Content: "trecho semantico relevante"}}

	prompt := buildEnrichedPrompt("pergunta atual", facts, episodes, context)

	for _, expected := range []string{"CONTEXTO SEMÂNTICO", "FATOS CONHECIDOS", "HISTÓRICO RECENTE", "TAREFA ATUAL", "pergunta atual"} {
		if !strings.Contains(prompt, expected) {
			t.Errorf("esperava prompt contendo %q, obteve:\n%s", expected, prompt)
		}
	}
}
