package openrouter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestComplete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Errorf("Authorization header = %q, want %q", got, "Bearer test-key")
		}

		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("falha ao decodificar request: %v", err)
		}
		if req.Model != "google/gemma-3-27b-it:free" {
			t.Errorf("Model = %q, want %q", req.Model, "google/gemma-3-27b-it:free")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(chatResponse{
			Choices: []chatChoice{
				{Message: chatMessage{Role: "assistant", Content: "resposta simulada"}},
			},
		})
	}))
	defer server.Close()

	client := NewClient("test-key")
	client.BaseURL = server.URL

	got, err := client.Complete(context.Background(), "google/gemma-3-27b-it:free", "pergunta")
	if err != nil {
		t.Fatalf("Complete() erro inesperado: %v", err)
	}
	if got != "resposta simulada" {
		t.Errorf("Complete() = %q, want %q", got, "resposta simulada")
	}
}

func TestComplete_EmptyAPIKey(t *testing.T) {
	client := NewClient("")

	_, err := client.Complete(context.Background(), "modelo-qualquer", "pergunta")
	if err != ErrEmptyAPIKey {
		t.Fatalf("Complete() erro = %v, want %v", err, ErrEmptyAPIKey)
	}
}

func TestComplete_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(chatResponse{
			Error: &apiError{Message: "rate limit excedido"},
		})
	}))
	defer server.Close()

	client := NewClient("test-key")
	client.BaseURL = server.URL

	_, err := client.Complete(context.Background(), "modelo-qualquer", "pergunta")
	if err == nil {
		t.Fatal("Complete() esperava erro, obteve nil")
	}
	if !strings.Contains(err.Error(), "rate limit excedido") {
		t.Errorf("Complete() erro = %v, want conter %q", err, "rate limit excedido")
	}
}
