package openai

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
		if len(req.Messages) != 2 || req.Messages[0].Role != "system" {
			t.Errorf("mensagens inesperadas: %+v", req.Messages)
		}
		json.NewEncoder(w).Encode(chatResponse{
			Choices: []chatChoice{{Message: chatMessage{Content: "resposta"}}},
		})
	}))
	defer server.Close()

	client := &Client{APIKey: "test-key", Model: "gpt-4o-mini", BaseURL: server.URL, HTTPClient: server.Client()}
	content, err := client.Chat(context.Background(), "system", "user")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if content != "resposta" {
		t.Errorf("esperava 'resposta', obteve %q", content)
	}
}

func TestChat_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(chatResponse{Error: &apiError{Message: "erro interno"}})
	}))
	defer server.Close()

	client := &Client{APIKey: "test-key", Model: "gpt-4o-mini", BaseURL: server.URL, HTTPClient: server.Client()}
	_, err := client.Chat(context.Background(), "system", "user")
	if err == nil {
		t.Fatal("esperava erro para status 500")
	}
}
