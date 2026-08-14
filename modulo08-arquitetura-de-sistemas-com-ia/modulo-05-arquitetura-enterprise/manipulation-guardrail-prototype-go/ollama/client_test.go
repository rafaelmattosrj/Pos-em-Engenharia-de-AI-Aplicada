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
		var req chatRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Model != "gemma4:e2b" {
			t.Errorf("esperava modelo gemma4:e2b, obteve %s", req.Model)
		}
		if req.Stream {
			t.Error("esperava stream=false")
		}

		json.NewEncoder(w).Encode(chatResponse{
			Message: Message{Role: "assistant", Content: "legitima"},
			Done:    true,
		})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	resposta, err := client.Chat(context.Background(), "gemma4:e2b", []Message{
		{Role: "user", Content: "pergunta qualquer"},
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resposta != "legitima" {
		t.Errorf("esperava \"legitima\", obteve %q", resposta)
	}
}

func TestChat_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.Chat(context.Background(), "gemma4:e2b", []Message{{Role: "user", Content: "x"}})
	if err == nil {
		t.Fatal("esperava erro para status 500, obteve nil")
	}
}

func TestChat_RespostaSemConteudo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(chatResponse{Done: true})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.Chat(context.Background(), "gemma4:e2b", []Message{{Role: "user", Content: "x"}})
	if err == nil {
		t.Fatal("esperava erro para resposta sem conteúdo, obteve nil")
	}
}
