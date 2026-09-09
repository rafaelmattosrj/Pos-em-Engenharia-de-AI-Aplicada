package grpo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const ollamaURL = "http://localhost:11434/api/generate"

// OllamaClient é um cliente HTTP mínimo para /api/generate do Ollama local --
// equivalente a chamarOllama() em JS.
type OllamaClient struct {
	HTTPClient *http.Client
}

// NewOllamaClient cria um client com timeout de 60s.
func NewOllamaClient() *OllamaClient {
	return &OllamaClient{HTTPClient: &http.Client{Timeout: 60 * time.Second}}
}

type ollamaRequest struct {
	Model   string             `json:"model"`
	Prompt  string             `json:"prompt"`
	Stream  bool               `json:"stream"`
	Options ollamaRequestOpts  `json:"options"`
}

type ollamaRequestOpts struct {
	Temperature float64 `json:"temperature"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

// Chamar envia o prompt ao modelo informado e retorna o texto gerado.
func (c *OllamaClient) Chamar(modelo, prompt string, temperatura float64) (string, error) {
	payload, err := json.Marshal(ollamaRequest{Model: modelo, Prompt: prompt, Stream: false, Options: ollamaRequestOpts{Temperature: temperatura}})
	if err != nil {
		return "", fmt.Errorf("ollama: falha ao serializar request: %w", err)
	}

	resp, err := c.HTTPClient.Post(ollamaURL, "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("ollama: falha na requisicao: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama respondeu %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ollama: falha ao ler resposta: %w", err)
	}

	var parsed ollamaResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("ollama: resposta invalida: %w", err)
	}
	return parsed.Response, nil
}
