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
		json.NewEncoder(w).Encode(chatResponse{Choices: []chatChoice{{Message: chatMessage{Content: "resposta"}}}})
	}))
	defer server.Close()

	client := &Client{APIKey: "test-key", Model: "gpt-4o-mini", ChatURL: server.URL, HTTPClient: server.Client()}
	content, err := client.Chat(context.Background(), "pergunta")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if content != "resposta" {
		t.Errorf("esperava 'resposta', obteve %q", content)
	}
}

func TestEmbed_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req embeddingRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Model != "text-embedding-3-small" {
			t.Errorf("modelo inesperado: %q", req.Model)
		}
		json.NewEncoder(w).Encode(embeddingResponse{Data: []embeddingData{{Embedding: []float64{0.1, 0.2, 0.3}}}})
	}))
	defer server.Close()

	client := &Client{APIKey: "test-key", EmbeddingModel: "text-embedding-3-small", EmbeddingURL: server.URL, HTTPClient: server.Client()}
	vector, err := client.Embed(context.Background(), "texto de exemplo")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(vector) != 3 {
		t.Errorf("esperava vetor de 3 dimensoes, obteve %d", len(vector))
	}
}

func TestEmbed_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := &Client{APIKey: "test-key", EmbeddingURL: server.URL, HTTPClient: server.Client()}
	_, err := client.Embed(context.Background(), "texto")
	if err == nil {
		t.Fatal("esperava erro para status 500")
	}
}
