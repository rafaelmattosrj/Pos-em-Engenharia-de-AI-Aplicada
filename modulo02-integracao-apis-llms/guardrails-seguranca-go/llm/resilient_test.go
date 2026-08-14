package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"guardrails-seguranca/openrouter"
)

func TestCall_FallsBackOnFailure(t *testing.T) {
	var seen []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		model, _ := body["model"].(string)
		seen = append(seen, model)

		if model == "model-a" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]string{"content": "ok"}}},
		})
	}))
	defer server.Close()

	client := &ResilientClient{
		OpenRouter:     &openrouter.Client{APIKey: "test-key", BaseURL: server.URL, HTTPClient: server.Client()},
		FallbackModels: []string{"model-a", "model-b"},
	}

	content, err := client.Call(context.Background(), "sys", "user")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if content != "ok" {
		t.Errorf("esperava 'ok', obteve %q", content)
	}
	if len(seen) != 2 {
		t.Errorf("esperava 2 tentativas, obteve %v", seen)
	}
}

func TestCall_AllModelsFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &ResilientClient{
		OpenRouter:     &openrouter.Client{APIKey: "test-key", BaseURL: server.URL, HTTPClient: server.Client()},
		FallbackModels: []string{"model-a"},
	}

	_, err := client.Call(context.Background(), "sys", "user")
	if err == nil {
		t.Fatal("esperava erro quando todos os modelos falham")
	}
}
