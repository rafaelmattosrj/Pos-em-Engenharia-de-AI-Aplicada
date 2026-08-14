package embeddings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEmbed_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req embedRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Model != "all-minilm" {
			t.Errorf("esperava modelo all-minilm, obteve %s", req.Model)
		}

		json.NewEncoder(w).Encode(embedResponse{Embedding: make([]float64, Dimension)})
	}))
	defer server.Close()

	client := NewClient(server.URL, "all-minilm")
	vector, err := client.Embed(context.Background(), "tensores em javascript")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(vector) != Dimension {
		t.Errorf("esperava vetor de %d dimensoes, obteve %d", Dimension, len(vector))
	}
}

func TestEmbed_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, "all-minilm")
	_, err := client.Embed(context.Background(), "qualquer texto")
	if err == nil {
		t.Fatal("esperava erro para status 500, obteve nil")
	}
}

func TestEmbed_EmptyEmbedding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(embedResponse{Embedding: nil})
	}))
	defer server.Close()

	client := NewClient(server.URL, "all-minilm")
	_, err := client.Embed(context.Background(), "qualquer texto")
	if err == nil {
		t.Fatal("esperava erro para embedding vazio, obteve nil")
	}
}
