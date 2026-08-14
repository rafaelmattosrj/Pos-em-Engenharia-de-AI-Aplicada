// Package ollama é um cliente mínimo para a API nativa de chat do Ollama
// (POST /api/chat, sem streaming) — porte do uso de ollama.chat(...) feito
// pelos SDKs ollama-js/ollama-python nos originais deste protótipo.
package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Message é uma mensagem de chat (role + content), igual ao formato aceito
// pela API do Ollama.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Client fala com a API nativa de chat do Ollama local.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient cria um Client apontando para baseURL (ex.: http://localhost:11434).
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{},
	}
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type chatResponse struct {
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

// Chat envia as mensagens ao modelo indicado e devolve o texto da resposta
// (sem streaming — equivalente a `ollama.chat({model, messages})` nos originais).
func (c *Client) Chat(ctx context.Context, model string, messages []Message) (string, error) {
	body, err := json.Marshal(chatRequest{Model: model, Messages: messages, Stream: false})
	if err != nil {
		return "", fmt.Errorf("ollama: falha ao serializar request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/chat", bytes.NewReader(body))
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
		return "", fmt.Errorf("ollama: falha ao chamar %s (rode 'ollama serve' e 'ollama pull %s'): %w", c.BaseURL, model, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ollama: falha ao ler resposta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama: status %d ao chamar /api/chat: %s", resp.StatusCode, string(respBody))
	}

	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("ollama: resposta inválida: %w", err)
	}
	if parsed.Message.Content == "" {
		return "", fmt.Errorf("ollama: resposta sem message.content: %s", string(respBody))
	}

	return parsed.Message.Content, nil
}
