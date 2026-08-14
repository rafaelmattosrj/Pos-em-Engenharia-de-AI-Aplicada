package openrouter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChat_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		json.NewDecoder(r.Body).Decode(&req)

		if len(req.Messages) != 2 || req.Messages[0].Role != "system" || req.Messages[1].Role != "user" {
			t.Errorf("mensagens inesperadas: %+v", req.Messages)
		}

		json.NewEncoder(w).Encode(chatResponse{
			Model:   "qwen/qwen3.6-plus:free",
			Choices: []chatChoice{{Message: chatMessage{Content: "resposta"}}},
		})
	}))
	defer server.Close()

	client := &Client{APIKey: "test-key", BaseURL: server.URL, HTTPClient: server.Client()}
	content, model, err := client.Chat(context.Background(), "qwen/qwen3.6-plus:free", "system", "user", 0.2, 100)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if content != "resposta" {
		t.Errorf("esperava 'resposta', obteve %q", content)
	}
	if model != "qwen/qwen3.6-plus:free" {
		t.Errorf("esperava echo do modelo, obteve %q", model)
	}
}

func TestChat_EmptyAPIKey(t *testing.T) {
	client := NewClient("")
	_, _, err := client.Chat(context.Background(), "model", "sys", "user", 0.2, 100)
	if err != ErrEmptyAPIKey {
		t.Fatalf("esperava ErrEmptyAPIKey, obteve %v", err)
	}
}

func TestChat_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(chatResponse{Error: &apiError{Message: "rate limited"}})
	}))
	defer server.Close()

	client := &Client{APIKey: "test-key", BaseURL: server.URL, HTTPClient: server.Client()}
	_, _, err := client.Chat(context.Background(), "model", "sys", "user", 0.2, 100)
	if err == nil {
		t.Fatal("esperava erro para status 429, obteve nil")
	}
}
