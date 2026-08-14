// Package openrouter é um cliente mínimo para a API de chat completions do
// OpenRouter (https://openrouter.ai), que é compatível com o formato de API da
// OpenAI. Diferente do LangChain4j usado na versão Java, o ecossistema Go não
// tem uma abstração tão madura para LLMs — por isso este cliente fala HTTP
// diretamente, no mesmo espírito do exemplo em shell/curl que este projeto
// também replica (ver README.md).
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
	APIKey     string
	HTTPClient *http.Client
	BaseURL    string

	// Referer e Title preenchem os headers opcionais recomendados pelo
	// OpenRouter (HTTP-Referer e X-Title), usados apenas para identificar a
	// origem da requisição no dashboard de uso.
	Referer string
	Title   string
}

// NewClient cria um Client com timeout padrão de 60s, espelhando o timeout
// configurado no OpenAiChatModel da versão Java.
func NewClient(apiKey string) *Client {
	return &Client{
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
		BaseURL:    baseURL,
	}
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

// ErrEmptyAPIKey é retornado quando nenhuma chave de API foi configurada.
var ErrEmptyAPIKey = errors.New("openrouter: API key vazia — defina OPENROUTER_API_KEY")

// Complete envia uma única mensagem de usuário para o modelo informado e
// retorna o texto da resposta. Equivale a chatModel.generate(mensagem) na
// versão Java.
func (c *Client) Complete(ctx context.Context, model, mensagem string) (string, error) {
	if c.APIKey == "" {
		return "", ErrEmptyAPIKey
	}

	reqBody := chatRequest{
		Model: model,
		Messages: []chatMessage{
			{Role: "user", Content: mensagem},
		},
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
	if c.Referer != "" {
		req.Header.Set("HTTP-Referer", c.Referer)
	}
	if c.Title != "" {
		req.Header.Set("X-Title", c.Title)
	}

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
