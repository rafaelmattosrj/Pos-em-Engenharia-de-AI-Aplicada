package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"guardrails-seguranca/graph"
	"guardrails-seguranca/llm"
	"guardrails-seguranca/openrouter"
	"guardrails-seguranca/service"
)

func TestChatHandler_MessageTooShort(t *testing.T) {
	orchestrator := &graph.Orchestrator{}

	body, _ := json.Marshal(map[string]string{"message": "oi"})
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	chatHandler(orchestrator).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400, obteve %d", rec.Code)
	}
}

func TestChatHandler_SafeMessagePassesThrough(t *testing.T) {
	call := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call++
		content := "SAFE"
		if call == 2 {
			content = "resposta"
		}
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]string{"content": content}}},
		})
	}))
	defer upstream.Close()

	client := &openrouter.Client{APIKey: "test-key", BaseURL: upstream.URL, HTTPClient: upstream.Client()}
	orchestrator := &graph.Orchestrator{
		Guardrails: &service.GuardrailsService{Client: client, Model: "safeguard-model", Enabled: true},
		Chat:       &service.ChatService{Client: &llm.ResilientClient{OpenRouter: client, FallbackModels: []string{"model-a"}}},
	}

	body, _ := json.Marshal(map[string]string{"message": "qual a previsao do tempo?"})
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	chatHandler(orchestrator).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d: %s", rec.Code, rec.Body.String())
	}

	var resp chatResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if !resp.Allowed || resp.Message != "resposta" {
		t.Errorf("resposta inesperada: %+v", resp)
	}
}
