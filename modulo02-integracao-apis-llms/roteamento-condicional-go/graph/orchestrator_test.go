package graph

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"roteamento-condicional/openrouter"
)

func TestInvoke_UppercaseCommand(t *testing.T) {
	orch := &Orchestrator{Fallback: &FallbackNode{}}

	state, err := orch.Invoke(context.Background(), "por favor coloque em UPPER case")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if state.Output != "POR FAVOR COLOQUE EM UPPER CASE" {
		t.Errorf("output inesperado: %q", state.Output)
	}
}

func TestInvoke_LowercaseCommand(t *testing.T) {
	orch := &Orchestrator{Fallback: &FallbackNode{}}

	state, err := orch.Invoke(context.Background(), "LOWER isso por favor")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if state.Output != "lower isso por favor" {
		t.Errorf("output inesperado: %q", state.Output)
	}
}

func TestInvoke_UnknownFallsBackToLLM(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"model":   "model-a",
			"choices": []map[string]any{{"message": map[string]string{"content": "resposta do LLM"}}},
		})
	}))
	defer server.Close()

	orch := &Orchestrator{
		Fallback: &FallbackNode{
			OpenRouter:     &openrouter.Client{APIKey: "test-key", BaseURL: server.URL, HTTPClient: server.Client()},
			FallbackModels: []string{"model-a"},
		},
	}

	state, err := orch.Invoke(context.Background(), "qual a capital do brasil?")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if state.Output != "resposta do LLM" {
		t.Errorf("esperava resposta do LLM, obteve %q", state.Output)
	}
}

func TestInvoke_UnknownAllModelsFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	orch := &Orchestrator{
		Fallback: &FallbackNode{
			OpenRouter:     &openrouter.Client{APIKey: "test-key", BaseURL: server.URL, HTTPClient: server.Client()},
			FallbackModels: []string{"model-a"},
		},
	}

	_, err := orch.Invoke(context.Background(), "pergunta qualquer")
	if err == nil {
		t.Fatal("esperava erro quando todos os modelos falham")
	}
}
