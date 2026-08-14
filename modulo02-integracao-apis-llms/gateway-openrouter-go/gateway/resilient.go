// Package gateway implementa o roteamento resiliente entre modelos
// (fallback em cadeia) e o handler HTTP do endpoint /chat — equivalente a
// ResilientChatClient + OpenRouterService + ChatController da versão Java.
package gateway

import (
	"context"
	"errors"
	"fmt"

	"gateway-openrouter/openrouter"
)

// Response é o corpo de resposta do endpoint /chat.
type Response struct {
	Model   string `json:"model"`
	Content string `json:"content"`
}

// ResilientClient tenta, em ordem, cada modelo de FallbackModels até um
// responder com sucesso — equivalente a ResilientChatClient.callForResponse.
type ResilientClient struct {
	OpenRouter     *openrouter.Client
	FallbackModels []string
	SystemPrompt   string
	Temperature    float64
	MaxTokens      int
}

// Generate roda a cadeia de fallback e retorna a primeira resposta bem
// sucedida — equivalente a OpenRouterService.generate.
func (c *ResilientClient) Generate(ctx context.Context, question string) (Response, error) {
	var lastErr error

	for _, model := range c.FallbackModels {
		content, usedModel, err := c.OpenRouter.Chat(ctx, model, c.SystemPrompt, question, c.Temperature, c.MaxTokens)
		if err != nil {
			lastErr = err
			continue
		}
		return Response{Model: usedModel, Content: content}, nil
	}

	if lastErr == nil {
		lastErr = errors.New("nenhum modelo de fallback configurado")
	}
	return Response{}, fmt.Errorf("all fallback models failed: %w", lastErr)
}
