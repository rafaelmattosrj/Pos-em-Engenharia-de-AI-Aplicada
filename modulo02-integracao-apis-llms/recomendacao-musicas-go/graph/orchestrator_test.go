package graph

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"recomendacao-musicas/llm"
	"recomendacao-musicas/openrouter"
	"recomendacao-musicas/repository"
	"recomendacao-musicas/service"
)

func newTestOrchestrator(t *testing.T, llmContent string, summarizeAfter int64) (*Orchestrator, *int) {
	t.Helper()

	db, err := repository.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("erro abrindo banco: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]string{"content": llmContent}}},
		})
	}))
	t.Cleanup(server.Close)

	client := &llm.ResilientClient{
		OpenRouter:     &openrouter.Client{APIKey: "test-key", BaseURL: server.URL, HTTPClient: server.Client()},
		FallbackModels: []string{"model-a"},
	}

	orch := &Orchestrator{
		Client:              client,
		MemoryService:       &service.MemoryService{Repository: &repository.ConversationRepository{DB: db}},
		PreferencesService:  &service.PreferencesService{Repository: &repository.PreferencesRepository{DB: db}, Client: client},
		SummarizeAfterCount: summarizeAfter,
	}

	return orch, &calls
}

func TestChat_BasicReplyAndPersistsHistory(t *testing.T) {
	orch, _ := newTestOrchestrator(t, "Recomendo Pink Floyd!", 10)
	ctx := context.Background()

	reply, err := orch.Chat(ctx, "user-1", "session-1", "me indique algo de rock progressivo")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if reply != "Recomendo Pink Floyd!" {
		t.Errorf("reply inesperado: %q", reply)
	}

	history, _ := orch.MemoryService.GetHistory(ctx, "session-1")
	if len(history) != 2 {
		t.Fatalf("esperava 2 mensagens persistidas (user+assistant), obteve %d", len(history))
	}
}

func TestChat_MusicPreferenceTriggersExtraction(t *testing.T) {
	orch, calls := newTestOrchestrator(t, "Legal saber que voce gosta disso!", 10)
	ctx := context.Background()

	_, err := orch.Chat(ctx, "user-1", "session-1", "eu gosto muito de jazz")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	// 1 chamada para a resposta do chat + 1 chamada para extrair preferencias.
	if *calls != 2 {
		t.Errorf("esperava 2 chamadas ao LLM (chat + extract), obteve %d", *calls)
	}

	prefs, _ := orch.PreferencesService.GetOrCreate(ctx, "user-1")
	if prefs.Preferences == "{}" {
		t.Error("esperava preferencias extraidas e persistidas")
	}
}

func TestChat_SummarizesAfterThreshold(t *testing.T) {
	orch, _ := newTestOrchestrator(t, "resposta padrao", 2)
	ctx := context.Background()

	if _, err := orch.Chat(ctx, "user-1", "session-1", "primeira mensagem neutra"); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	// Apos a 1a troca ja ha 2 mensagens (user+assistant) >= threshold=2,
	// entao a sumarizacao deve disparar nesta mesma chamada.
	if _, err := orch.Chat(ctx, "user-1", "session-1", "segunda mensagem neutra"); err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	prefs, _ := orch.PreferencesService.GetOrCreate(ctx, "user-1")
	if prefs.ConversationSummary == "" {
		t.Error("esperava um resumo de conversa persistido apos atingir o threshold")
	}
}
