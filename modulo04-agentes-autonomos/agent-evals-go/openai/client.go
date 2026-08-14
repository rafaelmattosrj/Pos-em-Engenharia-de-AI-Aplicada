// Package openai é um cliente mínimo para a API de chat completions da
// OpenAI — substitui o ChatClient (Spring AI) usado na versão Java.
package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultBaseURL = "https://api.openai.com/v1/chat/completions"

// Client fala com a API de chat completions da OpenAI.
type Client struct {
	APIKey     string
	Model      string
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient cria um Client com timeout de 60s.
func NewClient(apiKey, model string) *Client {
	return &Client{APIKey: apiKey, Model: model, BaseURL: defaultBaseURL, HTTPClient: &http.Client{Timeout: 60 * time.Second}}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
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

// Chat envia userPrompt e retorna o texto da resposta. Em caso de erro na
// chamada, retorna string vazia e o erro — o chamador decide como
// contabilizar a falha (equivalente ao catch de ToolSelectionEvaluator.askModel
// que retorna "{}").
func (c *Client) Chat(ctx context.Context, userPrompt string) (string, error) {
	payload, err := json.Marshal(chatRequest{Model: c.Model, Messages: []chatMessage{{Role: "user", Content: userPrompt}}})
	if err != nil {
		return "", fmt.Errorf("openai: falha ao serializar request: %w", err)
	}

	url := c.BaseURL
	if url == "" {
		url = defaultBaseURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("openai: falha ao montar request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai: falha na requisicao: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("openai: falha ao ler resposta: %w", err)
	}

	var parsed chatResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("openai: resposta invalida (status %d): %w", resp.StatusCode, err)
	}

	if resp.StatusCode != http.StatusOK {
		if parsed.Error != nil {
			return "", fmt.Errorf("openai: erro da API (status %d): %s", resp.StatusCode, parsed.Error.Message)
		}
		return "", fmt.Errorf("openai: erro da API (status %d)", resp.StatusCode)
	}

	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("openai: resposta sem choices")
	}

	return parsed.Choices[0].Message.Content, nil
}
