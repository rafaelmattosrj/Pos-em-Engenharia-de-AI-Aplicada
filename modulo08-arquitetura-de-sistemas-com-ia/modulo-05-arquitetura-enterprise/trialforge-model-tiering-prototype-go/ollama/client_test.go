package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestChatStream_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Errorf("esperava path /api/chat, obteve %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/x-ndjson")
		fmt.Fprintln(w, `{"message":{"role":"assistant","content":"Olá"},"done":false}`)
		fmt.Fprintln(w, `{"message":{"role":"assistant","content":", mundo"},"done":false}`)
		fmt.Fprintln(w, `{"message":{"role":"assistant","content":""},"done":true}`)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	var pedacos []string
	resposta, err := client.ChatStream(context.Background(), "gemma4:e2b", "system", "pergunta", func(s string) {
		pedacos = append(pedacos, s)
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if resposta != "Olá, mundo" {
		t.Errorf("esperava \"Olá, mundo\", obteve %q", resposta)
	}
	if len(pedacos) != 2 {
		t.Errorf("esperava 2 pedaços via onChunk, obteve %d", len(pedacos))
	}
}

func TestChatStream_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.ChatStream(context.Background(), "gemma4:e2b", "system", "pergunta", func(s string) {})
	if err == nil {
		t.Fatal("esperava erro para status 500, obteve nil")
	}
}

func TestChatStream_RequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("falha ao decodificar request: %v", err)
		}
		if req.Model != "gemma4:e2b" {
			t.Errorf("esperava modelo gemma4:e2b, obteve %s", req.Model)
		}
		if !req.Stream {
			t.Error("esperava stream=true")
		}
		if len(req.Messages) != 2 || req.Messages[0].Role != "system" || req.Messages[1].Role != "user" {
			t.Errorf("esperava mensagens [system, user], obteve %+v", req.Messages)
		}
		fmt.Fprintln(w, `{"message":{"role":"assistant","content":"ok"},"done":true}`)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	resposta, err := client.ChatStream(context.Background(), "gemma4:e2b", "sys", "user", func(s string) {})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !strings.Contains(resposta, "ok") {
		t.Errorf("esperava conter \"ok\", obteve %q", resposta)
	}
}

func TestEmbed_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embeddings" {
			t.Errorf("esperava path /api/embeddings, obteve %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"embedding":[0.1,0.2,0.3]}`)
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
		fmt.Fprint(w, `{"embedding":[]}`)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.Embed(context.Background(), "nomic-embed-text", "algum texto")
	if err == nil {
		t.Fatal("esperava erro para embedding vazio, obteve nil")
	}
}
