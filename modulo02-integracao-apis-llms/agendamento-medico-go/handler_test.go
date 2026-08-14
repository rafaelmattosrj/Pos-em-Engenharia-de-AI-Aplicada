package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"agendamento-medico/graph"
	"agendamento-medico/llm"
	"agendamento-medico/openrouter"
	"agendamento-medico/service"
)

func TestChatHandler_UnknownIntentFallsBackToFriendlyMessage(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]string{"content": "Ola! Como posso ajudar?"}}},
		})
	}))
	defer upstream.Close()

	client := &llm.ResilientClient{
		OpenRouter:     &openrouter.Client{APIKey: "test-key", BaseURL: upstream.URL, HTTPClient: upstream.Client()},
		FallbackModels: []string{"model-a"},
	}

	orchestrator := &graph.Orchestrator{
		IntentService:      &service.IntentService{Client: client},
		AppointmentService: service.NewAppointmentService(),
		MessageGenerator:   &service.MessageGeneratorService{Client: client},
	}

	body, _ := json.Marshal(map[string]string{"question": "oi tudo bem?"})
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	chatHandler(orchestrator).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d: %s", rec.Code, rec.Body.String())
	}

	var resp chatResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Reply != "Ola! Como posso ajudar?" {
		t.Errorf("reply inesperado: %q", resp.Reply)
	}
}

func TestChatHandler_QuestionTooShort(t *testing.T) {
	orchestrator := &graph.Orchestrator{}

	body, _ := json.Marshal(map[string]string{"question": "oi"})
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	chatHandler(orchestrator).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400, obteve %d", rec.Code)
	}
}
