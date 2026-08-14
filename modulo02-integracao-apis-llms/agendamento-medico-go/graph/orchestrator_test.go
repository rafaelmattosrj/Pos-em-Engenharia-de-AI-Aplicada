package graph

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"agendamento-medico/llm"
	"agendamento-medico/openrouter"
	"agendamento-medico/service"
)

// scriptedLLM retorna, em ordem, uma resposta diferente a cada chamada —
// simula o LLM respondendo primeiro com o IntentResult estruturado e depois
// com a mensagem de confirmação em linguagem natural.
func scriptedLLM(t *testing.T, responses []string) *llm.ResilientClient {
	t.Helper()
	call := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		content := responses[call%len(responses)]
		call++
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]string{"content": content}}},
		})
	}))
	t.Cleanup(server.Close)

	return &llm.ResilientClient{
		OpenRouter:     &openrouter.Client{APIKey: "test-key", BaseURL: server.URL, HTTPClient: server.Client()},
		FallbackModels: []string{"model-a"},
	}
}

func TestProcess_ScheduleSuccess(t *testing.T) {
	client := scriptedLLM(t, []string{
		`{"intent":"schedule","patientName":"Rafael Souza","professionalId":3,"professionalName":"Dra. Carol Gomes","datetime":"2030-01-01T10:00:00Z","reason":"dor de cabeca"}`,
		"Sua consulta foi agendada com sucesso!",
	})

	orch := &Orchestrator{
		IntentService:      &service.IntentService{Client: client},
		AppointmentService: service.NewAppointmentService(),
		MessageGenerator:   &service.MessageGeneratorService{Client: client},
	}

	reply, err := orch.Process(context.Background(), "quero marcar consulta com a Dra Carol")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if reply != "Sua consulta foi agendada com sucesso!" {
		t.Errorf("reply inesperado: %q", reply)
	}
}

func TestProcess_ScheduleConflictGeneratesErrorMessage(t *testing.T) {
	// Mesmo profissional (id=1) e mesmo horario dos agendamentos pre-existentes
	// criados por NewAppointmentService causaria conflito real; aqui simulamos
	// um conflito reservando o horario antes de processar.
	appointmentService := service.NewAppointmentService()
	conflictDate := "2030-06-15T09:00:00Z"
	parsedDate, _ := parseDatetime(&conflictDate)
	if _, err := appointmentService.BookAppointment(1, parsedDate, "Outro Paciente", "motivo"); err != nil {
		t.Fatalf("setup falhou: %v", err)
	}

	client := scriptedLLM(t, []string{
		`{"intent":"schedule","patientName":"Rafael Souza","professionalId":1,"professionalName":"Dr. Alicio da Silva","datetime":"2030-06-15T09:00:00Z","reason":"check-up"}`,
		"Desculpe, esse horario nao esta mais disponivel.",
	})

	orch := &Orchestrator{
		IntentService:      &service.IntentService{Client: client},
		AppointmentService: appointmentService,
		MessageGenerator:   &service.MessageGeneratorService{Client: client},
	}

	reply, err := orch.Process(context.Background(), "quero marcar consulta com o Dr Alicio")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if reply != "Desculpe, esse horario nao esta mais disponivel." {
		t.Errorf("reply inesperado: %q", reply)
	}
}

func TestProcess_UnknownIntentUsesFallbackMessage(t *testing.T) {
	client := scriptedLLM(t, []string{
		"nao consigo entender essa mensagem como JSON",
		"Ola! Em que posso ajudar?",
	})

	orch := &Orchestrator{
		IntentService:      &service.IntentService{Client: client},
		AppointmentService: service.NewAppointmentService(),
		MessageGenerator:   &service.MessageGeneratorService{Client: client},
	}

	reply, err := orch.Process(context.Background(), "oi, tudo bem?")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if reply != "Ola! Em que posso ajudar?" {
		t.Errorf("reply inesperado: %q", reply)
	}
}
