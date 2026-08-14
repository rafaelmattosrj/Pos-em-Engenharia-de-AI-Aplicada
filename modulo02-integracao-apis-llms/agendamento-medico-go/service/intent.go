package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"agendamento-medico/llm"
	"agendamento-medico/model"
)

const intentSystemPromptTemplate = `Você é um assistente de agendamento médico. Analise a mensagem do paciente e extraia as informações.

Profissionais disponíveis:
%s

Retorne um JSON com os campos:
- intent: "schedule", "cancel" ou "unknown"
- patientName: nome do paciente (ou null)
- professionalId: ID do profissional (ou null)
- professionalName: nome do profissional (ou null)
- datetime: data e hora em ISO 8601 (ou null)
- reason: motivo da consulta (ou null)

Responda APENAS com o JSON, sem explicações.
`

// IntentService identifica a intenção da mensagem do paciente usando o LLM
// com saída estruturada — equivalente a IntentService.java.
type IntentService struct {
	Client *llm.ResilientClient
}

// IdentifyIntent chama o LLM e converte a resposta em IntentResult. Qualquer
// falha (chamada ao LLM ou parse do JSON) resulta em intent "unknown", igual
// ao catch-all da versão Java.
func (s *IntentService) IdentifyIntent(ctx context.Context, userMessage string, professionals []model.Professional) model.IntentResult {
	var profLines []string
	for _, p := range professionals {
		profLines = append(profLines, fmt.Sprintf("ID %d: %s (%s)", p.ID, p.Name, p.Specialty))
	}
	systemPrompt := fmt.Sprintf(intentSystemPromptTemplate, strings.Join(profLines, "\n"))

	response, err := s.Client.Call(ctx, systemPrompt, userMessage)
	if err != nil {
		return model.UnknownIntent()
	}

	var result model.IntentResult
	if err := json.Unmarshal([]byte(extractJSON(response)), &result); err != nil {
		return model.UnknownIntent()
	}

	return result
}

// extractJSON remove cercas de código markdown (```json ... ```) que LLMs
// às vezes incluem na resposta, mesmo quando instruídos a responder "APENAS
// com o JSON" — mesma tolerância que o BeanOutputConverter do Spring AI
// aplica antes de desserializar.
func extractJSON(response string) string {
	trimmed := strings.TrimSpace(response)
	if !strings.HasPrefix(trimmed, "```") {
		return trimmed
	}

	trimmed = strings.TrimPrefix(trimmed, "```json")
	trimmed = strings.TrimPrefix(trimmed, "```")
	trimmed = strings.TrimSuffix(trimmed, "```")
	return strings.TrimSpace(trimmed)
}
