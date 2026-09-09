// Package openrouter fala com a API de chat completions do OpenRouter
// (compatível com o formato OpenAI) — substitui o ChatModel do Spring AI
// (spring-ai-openai-spring-boot-starter) usado na versão Java.
package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultBaseURL = "https://openrouter.ai/api/v1/chat/completions"

// Client é um cliente HTTP mínimo para a API de chat completions.
type Client struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient cria um Client com timeout de 30s.
func NewClient(apiKey string) *Client {
	return &Client{
		APIKey:     apiKey,
		BaseURL:    defaultBaseURL,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
	MaxTokens   int           `json:"max_tokens"`
}

type chatChoice struct {
	Message chatMessage `json:"message"`
}

type chatResponse struct {
	Model   string       `json:"model"`
	Choices []chatChoice `json:"choices"`
	Error   *apiError    `json:"error,omitempty"`
}

type apiError struct {
	Message string `json:"message"`
}

// ErrEmptyAPIKey é retornado quando nenhuma chave de API foi configurada.
var ErrEmptyAPIKey = errors.New("openrouter: API key vazia — defina OPENROUTER_API_KEY")

// Chat envia system+user prompts ao modelo informado e retorna o conteúdo e
// o nome do modelo que efetivamente respondeu (ecoado pela API).
func (c *Client) Chat(ctx context.Context, model, systemPrompt, userPrompt string, temperature float64, maxTokens int) (content string, usedModel string, err error) {
	if c.APIKey == "" {
		return "", "", ErrEmptyAPIKey
	}

	reqBody := chatRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", "", fmt.Errorf("openrouter: falha ao serializar request: %w", err)
	}

	url := c.BaseURL
	if url == "" {
		url = defaultBaseURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", "", fmt.Errorf("openrouter: falha ao montar request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("openrouter: falha na requisicao: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("openrouter: falha ao ler resposta: %w", err)
	}

	var parsed chatResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", "", fmt.Errorf("openrouter: resposta invalida (status %d): %w", resp.StatusCode, err)
	}

	if resp.StatusCode != http.StatusOK {
		if parsed.Error != nil {
			return "", "", fmt.Errorf("openrouter: erro da API (status %d): %s", resp.StatusCode, parsed.Error.Message)
		}
		return "", "", fmt.Errorf("openrouter: erro da API (status %d)", resp.StatusCode)
	}

	if len(parsed.Choices) == 0 {
		return "", "", errors.New("openrouter: resposta sem choices")
	}

	usedModel = parsed.Model
	if usedModel == "" {
		usedModel = model
	}

	return parsed.Choices[0].Message.Content, usedModel, nil
}
