// Package ollama é um cliente mínimo para a API de chat completions
// OpenAI-compatible exposta pelo Ollama (http://localhost:11434/v1) —
// porte do uso do OpenAiChatModel do LangChain4j na versão Java.
package ollama

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

// Client fala com a API de chat completions do Ollama.
type Client struct {
	BaseURL     string
	APIKey      string // Ollama aceita qualquer valor quando a API OpenAI-compatible está habilitada
	ModelName   string
	Temperature float64
	MaxRetries  int
	HTTPClient  *http.Client
}

// NewClient cria um Client com os defaults usados na versão Java:
// temperature 0.7, maxRetries 1, timeout de 60s.
func NewClient(baseURL, modelName string) *Client {
	return &Client{
		BaseURL:     baseURL,
		APIKey:      "ollama",
		ModelName:   modelName,
		Temperature: 0.7,
		MaxRetries:  1,
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

// Chat envia uma pergunta e retorna a resposta do modelo, retentando até
// c.MaxRetries vezes em caso de falha de rede/erro transitório — equivalente
// a .maxRetries(1) na configuração do OpenAiChatModel.
func (c *Client) Chat(ctx context.Context, pergunta string) (string, error) {
	var lastErr error
	for attempt := 0; attempt <= c.MaxRetries; attempt++ {
		resposta, err := c.doChat(ctx, pergunta)
		if err == nil {
			return resposta, nil
		}
		lastErr = err
	}
	return "", lastErr
}

func (c *Client) doChat(ctx context.Context, pergunta string) (string, error) {
	reqBody := chatRequest{
		Model:       c.ModelName,
		Temperature: c.Temperature,
		Messages:    []chatMessage{{Role: "user", Content: pergunta}},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("ollama: falha ao serializar request: %w", err)
	}

	url := c.BaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("ollama: falha ao montar request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama: falha na requisição (Ollama está rodando? 'ollama serve'): %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ollama: falha ao ler resposta: %w", err)
	}

	var parsed chatResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("ollama: resposta inválida (status %d): %w", resp.StatusCode, err)
	}

	if resp.StatusCode != http.StatusOK {
		if parsed.Error != nil {
			return "", fmt.Errorf("ollama: erro da API (status %d): %s", resp.StatusCode, parsed.Error.Message)
		}
		return "", fmt.Errorf("ollama: erro da API (status %d)", resp.StatusCode)
	}

	if len(parsed.Choices) == 0 {
		return "", errors.New("ollama: resposta sem choices")
	}

	return parsed.Choices[0].Message.Content, nil
}
