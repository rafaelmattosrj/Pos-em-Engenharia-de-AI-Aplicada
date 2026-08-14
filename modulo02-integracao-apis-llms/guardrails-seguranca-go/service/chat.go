package service

import (
	"context"

	"guardrails-seguranca/llm"
)

// ChatService é o serviço de chat principal — só recebe inputs já aprovados
// pelo guardrails. Equivalente a ChatService.java.
type ChatService struct {
	Client *llm.ResilientClient
}

func (s *ChatService) Chat(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	return s.Client.Call(ctx, systemPrompt, userMessage)
}
