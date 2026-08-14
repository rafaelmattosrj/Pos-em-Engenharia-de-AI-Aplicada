package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"recomendacao-musicas/graph"
	"recomendacao-musicas/llm"
	"recomendacao-musicas/openrouter"
	"recomendacao-musicas/repository"
	"recomendacao-musicas/service"
)

func newTestChatOrchestrator(t *testing.T) *graph.Orchestrator {
	t.Helper()

	db, err := repository.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("erro abrindo banco: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]string{"content": "resposta do assistente"}}},
		})
	}))
	t.Cleanup(upstream.Close)

	client := &llm.ResilientClient{
		OpenRouter:     &openrouter.Client{APIKey: "test-key", BaseURL: upstream.URL, HTTPClient: upstream.Client()},
		FallbackModels: []string{"model-a"},
	}

	return &graph.Orchestrator{
		Client:              client,
		MemoryService:       &service.MemoryService{Repository: &repository.ConversationRepository{DB: db}},
		PreferencesService:  &service.PreferencesService{Repository: &repository.PreferencesRepository{DB: db}, Client: client},
		SummarizeAfterCount: 10,
	}
}

func TestChatHandler_Success(t *testing.T) {
	orchestrator := newTestChatOrchestrator(t)

	body, _ := json.Marshal(map[string]string{"message": "me recomenda algo"})
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	chatHandler(orchestrator).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d: %s", rec.Code, rec.Body.String())
	}

	var resp chatResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Reply != "resposta do assistente" {
		t.Errorf("reply inesperado: %q", resp.Reply)
	}
}

func TestChatHandler_MessageTooShort(t *testing.T) {
	orchestrator := newTestChatOrchestrator(t)

	body, _ := json.Marshal(map[string]string{"message": "oi"})
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	chatHandler(orchestrator).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400, obteve %d", rec.Code)
	}
}
