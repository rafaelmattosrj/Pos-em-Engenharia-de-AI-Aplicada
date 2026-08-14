package gateway

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"gateway-openrouter/openrouter"
)

func newOpenRouterStub(t *testing.T, handler http.HandlerFunc) *openrouter.Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &openrouter.Client{APIKey: "test-key", BaseURL: server.URL, HTTPClient: server.Client()}
}

func TestGenerate_FirstModelSucceeds(t *testing.T) {
	calls := 0
	client := newOpenRouterStub(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		json.NewEncoder(w).Encode(map[string]any{
			"model":   "model-a",
			"choices": []map[string]any{{"message": map[string]string{"content": "ok"}}},
		})
	})

	resilient := &ResilientClient{
		OpenRouter:     client,
		FallbackModels: []string{"model-a", "model-b"},
		SystemPrompt:   "sys",
		Temperature:    0.2,
		MaxTokens:      100,
	}

	resp, err := resilient.Generate(context.Background(), "pergunta valida")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resp.Model != "model-a" || resp.Content != "ok" {
		t.Errorf("resposta inesperada: %+v", resp)
	}
	if calls != 1 {
		t.Errorf("esperava 1 chamada, obteve %d", calls)
	}
}

func TestGenerate_FallsBackOnFailure(t *testing.T) {
	var requestedModels []string
	client := newOpenRouterStub(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		model, _ := body["model"].(string)
		requestedModels = append(requestedModels, model)

		if model == "model-a" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"model":   "model-b",
			"choices": []map[string]any{{"message": map[string]string{"content": "resposta do fallback"}}},
		})
	})

	resilient := &ResilientClient{
		OpenRouter:     client,
		FallbackModels: []string{"model-a", "model-b"},
		SystemPrompt:   "sys",
	}

	resp, err := resilient.Generate(context.Background(), "pergunta valida")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resp.Model != "model-b" {
		t.Errorf("esperava fallback para model-b, obteve %q", resp.Model)
	}
	if len(requestedModels) != 2 {
		t.Errorf("esperava 2 modelos tentados, obteve %v", requestedModels)
	}
}

func TestGenerate_AllModelsFail(t *testing.T) {
	client := newOpenRouterStub(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	resilient := &ResilientClient{
		OpenRouter:     client,
		FallbackModels: []string{"model-a", "model-b"},
		SystemPrompt:   "sys",
	}

	_, err := resilient.Generate(context.Background(), "pergunta valida")
	if err == nil {
		t.Fatal("esperava erro quando todos os modelos falham")
	}
}
