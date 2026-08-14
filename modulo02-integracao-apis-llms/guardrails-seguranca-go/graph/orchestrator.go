// Package graph orquestra o fluxo de segurança — equivalente a
// SafeguardOrchestrator.java (que por sua vez substitui o StateGraph do
// LangGraph). Fluxo: START → guardrails_check → (blocked | chat) → END
package graph

import (
	"context"
	"fmt"
	"log"

	"guardrails-seguranca/service"
)

const systemPromptTemplate = "You are a helpful assistant.\n" +
	"User role: %s\n" +
	"User name: %s\n" +
	"Only users with admin role can access files.\n"

// Result é o resultado do processamento — equivalente ao record
// SafeguardOrchestrator.ChatResult.
type Result struct {
	Allowed bool
	Message string
}

// Orchestrator conduz a verificação de guardrails antes de encaminhar a
// mensagem ao chat principal.
type Orchestrator struct {
	Guardrails *service.GuardrailsService
	Chat       *service.ChatService
}

// Process resolve o usuário, roda a verificação de guardrails e, se seguro,
// encaminha a mensagem ao chat — equivalente a SafeguardOrchestrator.process.
func (o *Orchestrator) Process(ctx context.Context, username, userMessage string) (Result, error) {
	user := service.ResolveUser(username)

	systemPrompt := fmt.Sprintf(systemPromptTemplate, user.Role, user.DisplayName)

	guardrailResult := o.Guardrails.Check(ctx, userMessage, user.Role, user.DisplayName)
	log.Printf("🔒 Guardrails check: safe=%v\n", guardrailResult.Safe)

	if !guardrailResult.Safe {
		blockedMessage := "⛔ Sua mensagem foi bloqueada pelo sistema de segurança. Motivo: " + guardrailResult.Reason
		return Result{Allowed: false, Message: blockedMessage}, nil
	}

	response, err := o.Chat.Chat(ctx, systemPrompt, userMessage)
	if err != nil {
		return Result{}, err
	}
	return Result{Allowed: true, Message: response}, nil
}
