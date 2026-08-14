package service

import (
	"context"
	"fmt"

	"agendamento-medico/llm"
)

// MessageGeneratorService gera mensagens amigáveis para o paciente —
// equivalente a MessageGeneratorService.java.
type MessageGeneratorService struct {
	Client *llm.ResilientClient
}

func (s *MessageGeneratorService) GenerateSuccessMessage(ctx context.Context, action, patientName, details string) (string, error) {
	return s.Client.Call(ctx,
		"Você é um assistente de clínica médica. Gere uma mensagem amigável e profissional em português.",
		fmt.Sprintf("Gere uma confirmação de %s para o paciente %s. Detalhes: %s", action, patientName, details))
}

func (s *MessageGeneratorService) GenerateErrorMessage(ctx context.Context, errMessage string) (string, error) {
	return s.Client.Call(ctx,
		"Você é um assistente de clínica médica. Gere uma mensagem amigável em português.",
		fmt.Sprintf("Informe ao paciente que houve um problema: %s", errMessage))
}

func (s *MessageGeneratorService) GenerateUnknownIntentMessage(ctx context.Context, originalMessage string) (string, error) {
	return s.Client.Call(ctx,
		"Você é um assistente de clínica médica. Responda de forma amigável em português.",
		originalMessage)
}
