// Package ollama e um cliente minimo para a API nativa de chat do Ollama
// (POST /api/chat, sem streaming) — equivalente ao pacote npm/pip "ollama"
// usado no original em JS/Python.
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

// Client fala com a API nativa de chat do Ollama local.
type Client struct {
	BaseURL    string
	Modelo     string
	HTTPClient *http.Client
}

// NewClient cria um Client apontando para o baseURL informado
// (ex.: "http://localhost:11434", sem sufixo /api).
func NewClient(baseURL, modelo string) *Client {
	return &Client{
		BaseURL:    baseURL,
		Modelo:     modelo,
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
	}
}

// Message e uma mensagem de chat no formato aceito pela API do Ollama.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type chatResponse struct {
	Message Message `json:"message"`
	Error   string  `json:"error,omitempty"`
}

// Chat envia o historico de mensagens e devolve o texto da resposta do
// assistente — equivalente a resposta.message.content no original.
func (c *Client) Chat(ctx context.Context, mensagens []Message) (string, error) {
	reqBody := chatRequest{Model: c.Modelo, Messages: mensagens, Stream: false}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("ollama: falha ao serializar request: %w", err)
	}

	url := c.BaseURL + "/api/chat"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("ollama: falha ao montar request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf(
			"ollama: falha na requisição (Ollama está rodando? 'ollama serve'): %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ollama: falha ao ler resposta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama: erro da API (status %d): %s", resp.StatusCode, string(body))
	}

	var parsed chatResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("ollama: resposta inválida (status %d): %w", resp.StatusCode, err)
	}

	if parsed.Error != "" {
		return "", errors.New("ollama: erro da API: " + parsed.Error)
	}

	if parsed.Message.Content == "" {
		return "", errors.New("ollama: resposta sem message.content")
	}

	return parsed.Message.Content, nil
}
