// Package graph orquestra o chat de recomendação de músicas com memória de
// conversa e preferências persistidas — equivalente a
// MusicChatOrchestrator.java (que por sua vez substitui o StateGraph do
// LangGraph com checkpointer + store).
// Fluxo: START → chat → (savePreferences? / summarize?) → END
package graph

import (
	"context"
	"fmt"
	"strings"

	"recomendacao-musicas/llm"
	"recomendacao-musicas/service"
)

const chatSystemPromptTemplate = "Você é um especialista em música. Faça recomendações personalizadas.\n" +
	"Use o contexto do usuário para personalizar suas respostas.\n" +
	"Contexto do usuário: %s\n"

// Orchestrator conduz o fluxo de chat com memória.
type Orchestrator struct {
	Client              *llm.ResilientClient
	MemoryService       *service.MemoryService
	PreferencesService  *service.PreferencesService
	SummarizeAfterCount int64 // default 10, equivalente a app.summarize-after-messages
}

// Chat processa uma mensagem do usuário, gera uma resposta com contexto
// completo (histórico + preferências) e persiste a conversa — equivalente a
// MusicChatOrchestrator.chat.
func (o *Orchestrator) Chat(ctx context.Context, userID, sessionID, userMessage string) (string, error) {
	prefs, err := o.PreferencesService.GetOrCreate(ctx, userID)
	if err != nil {
		return "", err
	}
	history, err := o.MemoryService.GetHistory(ctx, sessionID)
	if err != nil {
		return "", err
	}

	var historyLines []string
	for _, m := range history {
		historyLines = append(historyLines, m.Role+": "+m.Content)
	}
	conversationContext := strings.Join(historyLines, "\n")

	var userContext strings.Builder
	if prefs.ConversationSummary != "" {
		userContext.WriteString("Resumo anterior: " + prefs.ConversationSummary + "\n")
	}
	if prefs.Preferences != "" && prefs.Preferences != "{}" {
		userContext.WriteString("Preferências: " + prefs.Preferences)
	}

	systemPrompt := fmt.Sprintf(chatSystemPromptTemplate, userContext.String())

	userPrompt := userMessage
	if conversationContext != "" {
		userPrompt = "Histórico:\n" + conversationContext + "\n\nMensagem atual: " + userMessage
	}

	response, err := o.Client.Call(ctx, systemPrompt, userPrompt)
	if err != nil {
		return "", err
	}

	if err := o.MemoryService.AddMessage(ctx, sessionID, "user", userMessage); err != nil {
		return "", err
	}
	if err := o.MemoryService.AddMessage(ctx, sessionID, "assistant", response); err != nil {
		return "", err
	}

	messageCount, err := o.MemoryService.CountMessages(ctx, sessionID)
	if err != nil {
		return "", err
	}

	summarizeAfter := o.SummarizeAfterCount
	if summarizeAfter == 0 {
		summarizeAfter = 10
	}

	switch {
	case messageCount >= summarizeAfter:
		fullHistory, err := o.MemoryService.GetHistory(ctx, sessionID)
		if err != nil {
			return "", err
		}
		var allTexts []string
		for _, m := range fullHistory {
			allTexts = append(allTexts, m.Role+": "+m.Content)
		}
		if _, err := o.PreferencesService.Summarize(ctx, userID, allTexts); err != nil {
			return "", err
		}
		if err := o.PreferencesService.ExtractAndSavePreferences(ctx, userID, allTexts); err != nil {
			return "", err
		}

	case containsMusicPreference(userMessage):
		exchange := []string{"user: " + userMessage, "assistant: " + response}
		if err := o.PreferencesService.ExtractAndSavePreferences(ctx, userID, exchange); err != nil {
			return "", err
		}
	}

	return response, nil
}

func containsMusicPreference(message string) bool {
	lower := strings.ToLower(message)
	for _, keyword := range []string{"gosto", "adoro", "prefiro", "curto", "favorit", "ouço"} {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}
