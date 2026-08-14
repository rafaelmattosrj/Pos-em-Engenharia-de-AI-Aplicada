package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChat_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(chatResponse{
			Choices: []chatChoice{{Message: chatMessage{Role: "assistant", Content: "resposta simulada"}}},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "llama3.2")
	got, err := client.Chat(context.Background(), "pergunta")
	if err != nil {
		t.Fatalf("Chat() erro inesperado: %v", err)
	}
	if got != "resposta simulada" {
		t.Errorf("Chat() = %q, want %q", got, "resposta simulada")
	}
}

func TestChat_RetriesOnFailure(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(chatResponse{Error: &apiError{Message: "falha transitória"}})
			return
		}
		json.NewEncoder(w).Encode(chatResponse{
			Choices: []chatChoice{{Message: chatMessage{Content: "sucesso na 2a tentativa"}}},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "llama3.2")
	client.MaxRetries = 1

	got, err := client.Chat(context.Background(), "pergunta")
	if err != nil {
		t.Fatalf("Chat() erro inesperado: %v", err)
	}
	if got != "sucesso na 2a tentativa" {
		t.Errorf("Chat() = %q, want %q", got, "sucesso na 2a tentativa")
	}
	if attempts != 2 {
		t.Errorf("attempts = %d, want 2", attempts)
	}
}

func TestChat_NoResultAfterRetries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(chatResponse{Error: &apiError{Message: "indisponível"}})
	}))
	defer server.Close()

	client := NewClient(server.URL, "llama3.2")
	client.MaxRetries = 1

	_, err := client.Chat(context.Background(), "pergunta")
	if err == nil {
		t.Fatal("Chat() esperava erro após esgotar tentativas, obteve nil")
	}
}
