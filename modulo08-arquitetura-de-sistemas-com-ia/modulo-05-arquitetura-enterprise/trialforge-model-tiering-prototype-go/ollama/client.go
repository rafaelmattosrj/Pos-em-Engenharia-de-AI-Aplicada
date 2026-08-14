// Package ollama é um cliente mínimo para as APIs nativas de chat (com
// streaming) e embeddings do Ollama — porte do uso de
// ollama.chat({..., stream: true})/ollama.embeddings(...) feito pelos SDKs
// ollama-js/ollama-python no original deste protótipo.
package ollama

import (
	"bufio"
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

type chatStreamLine struct {
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

// ChatStream gera uma resposta em streaming: cada pedaço de texto recebido é
// repassado a onChunk (equivalente a imprimir cada parte no console, como no
// original), e o texto completo é devolvido no final.
func (c *Client) ChatStream(ctx context.Context, model, systemPrompt, userPrompt string, onChunk func(string)) (string, error) {
	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}
	body, err := json.Marshal(chatRequest{Model: model, Messages: messages, Stream: true})
	if err != nil {
		return "", fmt.Errorf("ollama: falha ao serializar request de chat: %w", err)
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
		return "", fmt.Errorf("ollama: falha ao chamar %s/api/chat (rode 'ollama serve' e 'ollama pull %s'): %w", c.BaseURL, model, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama: status %d ao gerar (stream): %s", resp.StatusCode, string(respBody))
	}

	var rascunho bytes.Buffer
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		linha := scanner.Bytes()
		if len(bytes.TrimSpace(linha)) == 0 {
			continue
		}
		var parsed chatStreamLine
		if err := json.Unmarshal(linha, &parsed); err != nil {
			return "", fmt.Errorf("ollama: linha de stream inválida: %w", err)
		}
		if parsed.Message.Content != "" {
			onChunk(parsed.Message.Content)
			rascunho.WriteString(parsed.Message.Content)
		}
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("ollama: falha ao ler stream: %w", err)
	}

	return rascunho.String(), nil
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/embeddings", bytes.NewReader(body))
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
		return nil, fmt.Errorf("ollama: falha ao chamar %s/api/embeddings: %w", c.BaseURL, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ollama: falha ao ler resposta: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama: status %d ao gerar embedding: %s", resp.StatusCode, string(respBody))
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
