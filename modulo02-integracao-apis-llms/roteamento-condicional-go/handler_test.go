package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"roteamento-condicional/graph"
)

func TestChatHandler_UppercaseCommand(t *testing.T) {
	orchestrator := &graph.Orchestrator{Fallback: &graph.FallbackNode{}}

	body, _ := json.Marshal(map[string]string{"question": "converte para UPPER"})
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	chatHandler(orchestrator).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d", rec.Code)
	}
	if rec.Body.String() != "CONVERTE PARA UPPER" {
		t.Errorf("body inesperado: %q", rec.Body.String())
	}
}

func TestChatHandler_QuestionTooShort(t *testing.T) {
	orchestrator := &graph.Orchestrator{Fallback: &graph.FallbackNode{}}

	body, _ := json.Marshal(map[string]string{"question": "oi"})
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	chatHandler(orchestrator).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400, obteve %d", rec.Code)
	}
}
