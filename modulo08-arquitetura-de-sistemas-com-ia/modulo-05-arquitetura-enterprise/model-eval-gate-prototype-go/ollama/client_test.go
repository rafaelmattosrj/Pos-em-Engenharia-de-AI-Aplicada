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
		if r.URL.Path != "/api/chat" {
			t.Errorf("esperava path /api/chat, obteve %s", r.URL.Path)
		}
		var req chatRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Stream {
			t.Error("esperava stream=false")
		}
		json.NewEncoder(w).Encode(chatResponse{Message: Message{Role: "assistant", Content: "resposta gerada"}, Done: true})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	resposta, err := client.Chat(context.Background(), "gemma4:e2b", "system", "pergunta")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resposta != "resposta gerada" {
		t.Errorf("esperava \"resposta gerada\", obteve %q", resposta)
	}
}

func TestChat_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.Chat(context.Background(), "gemma4:e2b", "system", "pergunta")
	if err == nil {
		t.Fatal("esperava erro para status 500, obteve nil")
	}
}

func TestEmbed_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embeddings" {
			t.Errorf("esperava path /api/embeddings, obteve %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(embedResponse{Embedding: []float64{0.1, 0.2, 0.3}})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	vetor, err := client.Embed(context.Background(), "nomic-embed-text", "algum texto")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(vetor) != 3 {
		t.Errorf("esperava vetor de 3 dimensões, obteve %d", len(vetor))
	}
}

func TestEmbed_EmptyEmbedding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(embedResponse{Embedding: nil})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.Embed(context.Background(), "nomic-embed-text", "algum texto")
	if err == nil {
		t.Fatal("esperava erro para embedding vazio, obteve nil")
	}
}
