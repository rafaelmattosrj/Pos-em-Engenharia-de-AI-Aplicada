// Package openai é um cliente mínimo para a API de chat completions e de
// embeddings da OpenAI — substitui o ChatClient/EmbeddingModel (Spring AI)
// usados na versão Java.
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

const (
	defaultChatURL      = "https://api.openai.com/v1/chat/completions"
	defaultEmbeddingURL = "https://api.openai.com/v1/embeddings"
)

// Client fala com a API de chat completions e de embeddings da OpenAI.
type Client struct {
	APIKey         string
	Model          string
	EmbeddingModel string
	ChatURL        string
	EmbeddingURL   string
	HTTPClient     *http.Client
}

// NewClient cria um Client com timeout de 60s.
func NewClient(apiKey, model, embeddingModel string) *Client {
	return &Client{
		APIKey:         apiKey,
		Model:          model,
		EmbeddingModel: embeddingModel,
		ChatURL:        defaultChatURL,
		EmbeddingURL:   defaultEmbeddingURL,
		HTTPClient:     &http.Client{Timeout: 60 * time.Second},
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

// Chat envia userPrompt (sem system prompt, igual ao uso feito em
// ReflectionEngine/MemoryAwareAgent na versão Java) e retorna o texto da
// resposta.
func (c *Client) Chat(ctx context.Context, userPrompt string) (string, error) {
	reqBody := chatRequest{Model: c.Model, Messages: []chatMessage{{Role: "user", Content: userPrompt}}}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("openai: falha ao serializar request: %w", err)
	}

	url := c.ChatURL
	if url == "" {
		url = defaultChatURL
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

type embeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embeddingData struct {
	Embedding []float64 `json:"embedding"`
}

type embeddingResponse struct {
	Data  []embeddingData `json:"data"`
	Error *apiError       `json:"error,omitempty"`
}

// Embed retorna o vetor de embedding de text, usando c.EmbeddingModel
// (default "text-embedding-3-small").
func (c *Client) Embed(ctx context.Context, text string) ([]float64, error) {
	model := c.EmbeddingModel
	if model == "" {
		model = "text-embedding-3-small"
	}

	payload, err := json.Marshal(embeddingRequest{Model: model, Input: text})
	if err != nil {
		return nil, fmt.Errorf("openai: falha ao serializar request de embedding: %w", err)
	}

	url := c.EmbeddingURL
	if url == "" {
		url = defaultEmbeddingURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("openai: falha ao montar request de embedding: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai: falha na requisicao de embedding: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("openai: falha ao ler resposta de embedding: %w", err)
	}

	var parsed embeddingResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("openai: resposta de embedding invalida (status %d): %w", resp.StatusCode, err)
	}

	if resp.StatusCode != http.StatusOK {
		if parsed.Error != nil {
			return nil, fmt.Errorf("openai: erro da API de embedding (status %d): %s", resp.StatusCode, parsed.Error.Message)
		}
		return nil, fmt.Errorf("openai: erro da API de embedding (status %d)", resp.StatusCode)
	}

	if len(parsed.Data) == 0 {
		return nil, fmt.Errorf("openai: resposta de embedding sem dados")
	}

	return parsed.Data[0].Embedding, nil
}
