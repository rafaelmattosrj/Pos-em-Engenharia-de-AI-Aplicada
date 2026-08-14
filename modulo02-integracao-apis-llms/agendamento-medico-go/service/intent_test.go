package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"agendamento-medico/llm"
	"agendamento-medico/model"
	"agendamento-medico/openrouter"
)

func newIntentService(t *testing.T, llmContent string) *IntentService {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]string{"content": llmContent}}},
		})
	}))
	t.Cleanup(server.Close)

	return &IntentService{
		Client: &llm.ResilientClient{
			OpenRouter:     &openrouter.Client{APIKey: "test-key", BaseURL: server.URL, HTTPClient: server.Client()},
			FallbackModels: []string{"model-a"},
		},
	}
}

func TestIdentifyIntent_PlainJSON(t *testing.T) {
	svc := newIntentService(t, `{"intent":"schedule","patientName":"Rafael","professionalId":1,"professionalName":"Dr. Alicio","datetime":"2030-01-01T10:00:00Z","reason":"check-up"}`)

	result := svc.IdentifyIntent(context.Background(), "quero agendar", Professionals)
	if result.Intent != "schedule" {
		t.Errorf("esperava intent=schedule, obteve %q", result.Intent)
	}
	if result.PatientName == nil || *result.PatientName != "Rafael" {
		t.Errorf("patientName inesperado: %+v", result.PatientName)
	}
}

func TestIdentifyIntent_MarkdownFencedJSON(t *testing.T) {
	svc := newIntentService(t, "```json\n{\"intent\":\"cancel\"}\n```")

	result := svc.IdentifyIntent(context.Background(), "quero cancelar", Professionals)
	if result.Intent != "cancel" {
		t.Errorf("esperava intent=cancel, obteve %q", result.Intent)
	}
}

func TestIdentifyIntent_InvalidJSONFallsBackToUnknown(t *testing.T) {
	svc := newIntentService(t, "isso nao e JSON")

	result := svc.IdentifyIntent(context.Background(), "mensagem qualquer", Professionals)
	if result.Intent != model.UnknownIntent().Intent {
		t.Errorf("esperava intent=unknown, obteve %q", result.Intent)
	}
}
