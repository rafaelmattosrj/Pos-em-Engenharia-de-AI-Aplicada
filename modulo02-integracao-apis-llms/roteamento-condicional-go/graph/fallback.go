package graph

import (
	"context"
	"errors"
	"fmt"

	"roteamento-condicional/openrouter"
)

const fallbackSystemPrompt = "Você é um assistente útil. Responda em português."

// FallbackNode chama o LLM (com fallback em cadeia entre modelos) quando o
// comando é desconhecido — equivalente a FallbackNode.java +
// ResilientChatClient.java.
type FallbackNode struct {
	OpenRouter     *openrouter.Client
	FallbackModels []string
	Temperature    float64
	MaxTokens      int
}

// Process gera uma resposta livre para state.Output e a define como o novo
// output do estado.
func (n *FallbackNode) Process(ctx context.Context, state State) (State, error) {
	var lastErr error

	for _, model := range n.FallbackModels {
		content, _, err := n.OpenRouter.Chat(ctx, model, fallbackSystemPrompt, state.Output, n.Temperature, n.MaxTokens)
		if err != nil {
			lastErr = err
			continue
		}
		return state.WithOutput(content), nil
	}

	if lastErr == nil {
		lastErr = errors.New("nenhum modelo de fallback configurado")
	}
	return state, fmt.Errorf("all fallback models failed: %w", lastErr)
}
