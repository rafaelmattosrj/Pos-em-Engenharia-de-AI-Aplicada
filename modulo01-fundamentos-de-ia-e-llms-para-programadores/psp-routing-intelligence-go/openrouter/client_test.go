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
		if req.Temperature != 0.3 {
			t.Errorf("esperava temperature=0.3, obteve %v", req.Temperature)
		}
		json.NewEncoder(w).Encode(chatResponse{
			Choices: []chatChoice{{Message: chatMessage{Role: "assistant", Content: `{"primary":"ADYEN","confidence":0.87,"reasoning":"ok","fallback":[]}`}}},
		})
	}))
	defer server.Close()

	client := &Client{APIKey: "test-key", Model: "test-model", Temperature: 0.3, MaxRetries: 2, HTTPClient: server.Client(), BaseURL: server.URL}

	answer, err := client.Chat(context.Background(), "prompt de roteamento")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if answer == "" {
		t.Errorf("esperava resposta nao vazia")
	}
}

func TestChat_EmptyAPIKey(t *testing.T) {
	client := NewClient("", "test-model")
	_, err := client.Chat(context.Background(), "prompt")
	if err != ErrEmptyAPIKey {
		t.Fatalf("esperava ErrEmptyAPIKey, obteve %v", err)
	}
}

func TestChat_RetriesOnFailure(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(chatResponse{
			Choices: []chatChoice{{Message: chatMessage{Content: "ok apos retry"}}},
		})
	}))
	defer server.Close()

	client := &Client{APIKey: "test-key", Model: "test-model", Temperature: 0.3, MaxRetries: 2, HTTPClient: server.Client(), BaseURL: server.URL}

	answer, err := client.Chat(context.Background(), "prompt")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if answer != "ok apos retry" {
		t.Errorf("esperava 'ok apos retry', obteve %q", answer)
	}
	if attempts != 2 {
		t.Errorf("esperava 2 tentativas, obteve %d", attempts)
	}
}

func TestChat_ErroDaApiPropagaMensagem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(chatResponse{Error: &apiError{Message: "rate limit excedido"}})
	}))
	defer server.Close()

	client := &Client{APIKey: "test-key", Model: "test-model", MaxRetries: 0, HTTPClient: server.Client(), BaseURL: server.URL}

	_, err := client.Chat(context.Background(), "prompt")
	if err == nil {
		t.Fatal("esperava erro, obteve nil")
	}
}
