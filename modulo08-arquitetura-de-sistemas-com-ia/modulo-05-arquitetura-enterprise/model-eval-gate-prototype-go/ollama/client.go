// Package ollama é um cliente mínimo para as APIs nativas de chat e embeddings
// do Ollama (POST /api/chat sem streaming, POST /api/embeddings) — porte do
// uso de ollama.chat(...)/ollama.embeddings(...) feito pelos SDKs
// ollama-js/ollama-python nos originais deste protótipo.
package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Message é uma mensagem de chat (role + content).
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Client fala com a API nativa do Ollama local.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient cria um Client apontando para baseURL (ex.: http://localhost:11434).
func NewClient(baseURL string) *Client {
	return &Client{BaseURL: baseURL, HTTPClient: &http.Client{}}
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
// (sem streaming — este protótipo não usa streaming em chat, diferente do
// trialforge-model-tiering-prototype irmão).
func (c *Client) Chat(ctx context.Context, model, systemPrompt, userPrompt string) (string, error) {
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}
	body, err := json.Marshal(chatRequest{Model: model, Messages: messages, Stream: false})
	if err != nil {
		return "", fmt.Errorf("ollama: falha ao serializar request de chat: %w", err)
	}

	respBody, err := c.post(ctx, "/api/chat", body)
	if err != nil {
		return "", err
	}

	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("ollama: resposta de chat inválida: %w", err)
	}
	if parsed.Message.Content == "" {
		return "", fmt.Errorf("ollama: resposta de chat sem message.content: %s", string(respBody))
	}
	return parsed.Message.Content, nil
}

type embedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type embedResponse struct {
	Embedding []float64 `json:"embedding"`
}

// Embed retorna o vetor de embedding de texto, usando o modelo indicado.
func (c *Client) Embed(ctx context.Context, model, texto string) ([]float64, error) {
	body, err := json.Marshal(embedRequest{Model: model, Prompt: texto})
	if err != nil {
		return nil, fmt.Errorf("ollama: falha ao serializar request de embedding: %w", err)
	}

	respBody, err := c.post(ctx, "/api/embeddings", body)
	if err != nil {
		return nil, err
	}

	var parsed embedResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("ollama: resposta de embedding inválida: %w", err)
	}
	if len(parsed.Embedding) == 0 {
		return nil, fmt.Errorf("ollama: embedding vazio: %s", string(respBody))
	}
	return parsed.Embedding, nil
}

func (c *Client) post(ctx context.Context, path string, body []byte) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama: falha ao montar request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama: falha ao chamar %s%s (rode 'ollama serve'): %w", c.BaseURL, path, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ollama: falha ao ler resposta: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama: status %d em %s: %s", resp.StatusCode, path, string(respBody))
	}
	return respBody, nil
}
