// Package openrouter é um cliente mínimo para a API de chat completions do
// OpenRouter, compatível com o formato OpenAI — substitui o
// OpenAiChatModel (LangChain4j) usado na versão Java para chamar o LLM via
// OpenRouter.
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

const baseURL = "https://openrouter.ai/api/v1/chat/completions"

// Client fala com a API de chat completions do OpenRouter.
type Client struct {
	APIKey      string
	Model       string
	Temperature float64
	MaxRetries  int
	HTTPClient  *http.Client

	// BaseURL sobrescreve o endpoint padrão — usado nos testes.
	BaseURL string
}

// NewClient cria um Client com temperature=0.3 e maxRetries=2, mesmos
// defaults usados no OpenAiChatModel da versão Java.
func NewClient(apiKey, model string) *Client {
	return &Client{
		APIKey:      apiKey,
		Model:       model,
		Temperature: 0.3,
		MaxRetries:  2,
		HTTPClient:  &http.Client{Timeout: 60 * time.Second},
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
}

type chatChoice struct {
	Message chatMessage `json:"message"`
}

type chatResponse struct {
	Choices []chatChoice `json:"choices"`
	Error   *apiError    `json:"error,omitempty"`
}

type apiError struct {
	Message string `json:"message"`
}

// ErrEmptyAPIKey é retornado quando nenhuma chave de API foi configurada.
var ErrEmptyAPIKey = errors.New("openrouter: API key vazia — defina OPENROUTER_API_KEY")

// Chat envia prompt como mensagem de usuário e retorna o texto da resposta,
// repetindo a chamada em caso de falha até MaxRetries vezes.
func (c *Client) Chat(ctx context.Context, prompt string) (string, error) {
	if c.APIKey == "" {
		return "", ErrEmptyAPIKey
	}

	attempts := c.MaxRetries + 1
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		answer, err := c.doChat(ctx, prompt)
		if err == nil {
			return answer, nil
		}
		lastErr = err
	}
	return "", lastErr
}

func (c *Client) doChat(ctx context.Context, prompt string) (string, error) {
	reqBody := chatRequest{
		Model:       c.Model,
		Messages:    []chatMessage{{Role: "user", Content: prompt}},
		Temperature: c.Temperature,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("openrouter: falha ao serializar request: %w", err)
	}

	url := c.BaseURL
	if url == "" {
		url = baseURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("openrouter: falha ao montar request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("openrouter: falha na requisição: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("openrouter: falha ao ler resposta: %w", err)
	}

	var parsed chatResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("openrouter: resposta inválida (status %d): %w", resp.StatusCode, err)
	}

	if resp.StatusCode != http.StatusOK {
		if parsed.Error != nil {
			return "", fmt.Errorf("openrouter: erro da API (status %d): %s", resp.StatusCode, parsed.Error.Message)
		}
		return "", fmt.Errorf("openrouter: erro da API (status %d)", resp.StatusCode)
	}

	if len(parsed.Choices) == 0 {
		return "", errors.New("openrouter: resposta sem choices")
	}

	return parsed.Choices[0].Message.Content, nil
}
