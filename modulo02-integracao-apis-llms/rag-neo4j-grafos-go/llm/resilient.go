// Package llm fornece o cliente resiliente compartilhado por
// CypherGeneratorService e AnalyticalResponseService — equivalente a
// ResilientChatClient.java.
package llm

import (
	"context"
	"errors"
	"fmt"

	"rag-neo4j-grafos/openrouter"
)

// ResilientClient tenta, em ordem, cada modelo de FallbackModels até um
// responder com sucesso.
type ResilientClient struct {
	OpenRouter     *openrouter.Client
	FallbackModels []string
	Temperature    float64
	MaxTokens      int
}

// Call envia system+user prompts e retorna o texto da primeira resposta bem
// sucedida — equivalente a ResilientChatClient.call.
func (c *ResilientClient) Call(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	var lastErr error

	for _, model := range c.FallbackModels {
		content, _, err := c.OpenRouter.Chat(ctx, model, systemPrompt, userPrompt, c.Temperature, c.MaxTokens)
		if err != nil {
			lastErr = err
			continue
		}
		return content, nil
	}

	if lastErr == nil {
		lastErr = errors.New("nenhum modelo de fallback configurado")
	}
	return "", fmt.Errorf("all fallback models failed: %w", lastErr)
}
