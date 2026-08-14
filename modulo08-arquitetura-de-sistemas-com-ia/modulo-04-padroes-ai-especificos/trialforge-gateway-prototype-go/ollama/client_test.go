package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"trialforge-gateway-prototype/gateway"
)

func TestEmbedarSucesso(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embeddings" {
			t.Errorf("path = %s, esperado /api/embeddings", r.URL.Path)
		}
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		if body["model"] != "nomic-embed-text" {
			t.Errorf("model = %s, esperado nomic-embed-text", body["model"])
		}
		fmt.Fprint(w, `{"embedding": [0.1, 0.2, 0.3]}`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "nomic-embed-text")
	embedding, err := client.Embedar(context.Background(), "texto de teste")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	esperado := []float64{0.1, 0.2, 0.3}
	for i, v := range esperado {
		if embedding[i] != v {
			t.Errorf("embedding[%d] = %v, esperado %v", i, embedding[i], v)
		}
	}
}

func TestEmbedarRetryAposFalhaTransitoria(t *testing.T) {
	tentativas := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tentativas++
		if tentativas < 2 {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "erro transitório")
			return
		}
		fmt.Fprint(w, `{"embedding": [1.0]}`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "nomic-embed-text")
	embedding, err := client.Embedar(context.Background(), "texto")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if tentativas != 2 {
		t.Errorf("esperado 2 tentativas, obtido %d", tentativas)
	}
	if len(embedding) != 1 || embedding[0] != 1.0 {
		t.Errorf("embedding = %v, esperado [1.0]", embedding)
	}
}

func TestEmbedarEsgotaTentativas(t *testing.T) {
	tentativas := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tentativas++
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "sempre falha")
	}))
	defer server.Close()

	client := NewClient(server.URL, "nomic-embed-text")
	client.MaxTentativas = 2
	_, err := client.Embedar(context.Background(), "texto")
	if err == nil {
		t.Fatal("esperado erro após esgotar tentativas")
	}
	if tentativas != 2 {
		t.Errorf("esperado 2 tentativas, obtido %d", tentativas)
	}
}

func TestEmbedarTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		fmt.Fprint(w, `{"embedding": [1.0]}`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "nomic-embed-text")
	client.MaxTentativas = 1
	client.Timeout = 10 * time.Millisecond
	_, err := client.Embedar(context.Background(), "texto")
	if err == nil {
		t.Fatal("esperado erro de timeout")
	}
}

func TestChatStreamSucesso(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Errorf("path = %s, esperado /api/chat", r.URL.Path)
		}
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["model"] != "gemma4:e2b" {
			t.Errorf("model = %v, esperado gemma4:e2b", body["model"])
		}
		if body["stream"] != true {
			t.Errorf("stream = %v, esperado true", body["stream"])
		}

		w.Header().Set("Content-Type", "application/x-ndjson")
		fmt.Fprintln(w, `{"message":{"role":"assistant","content":"Olá"},"done":false}`)
		fmt.Fprintln(w, `{"message":{"role":"assistant","content":", mundo"},"done":false}`)
		fmt.Fprintln(w, `{"message":{"role":"assistant","content":""},"done":true}`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "nomic-embed-text")
	var recebido string
	err := client.ChatStream(context.Background(), "gemma4:e2b", []gateway.Message{
		{Role: "user", Content: "oi"},
	}, func(content string) {
		recebido += content
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if recebido != "Olá, mundo" {
		t.Errorf("recebido = %q, esperado %q", recebido, "Olá, mundo")
	}
}

func TestChatStreamErroNaLinha(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"error":"modelo não encontrado"}`)
	}))
	defer server.Close()

	client := NewClient(server.URL, "nomic-embed-text")
	err := client.ChatStream(context.Background(), "modelo-inexistente", nil, func(string) {})
	if err == nil {
		t.Fatal("esperado erro quando o stream reporta erro")
	}
}

func TestChatStreamStatusNaoOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprint(w, "indisponível")
	}))
	defer server.Close()

	client := NewClient(server.URL, "nomic-embed-text")
	client.MaxTentativas = 1
	err := client.ChatStream(context.Background(), "modelo", nil, func(string) {})
	if err == nil {
		t.Fatal("esperado erro com status != 200")
	}
}
