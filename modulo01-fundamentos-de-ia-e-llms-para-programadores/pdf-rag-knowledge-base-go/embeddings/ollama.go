// Package embeddings gera embeddings localmente via Ollama — substitui o
// AllMiniLmL6V2EmbeddingModel (ONNX in-process) da versão Java, já que Go não
// tem um binding ONNX Runtime conveniente sem cgo. Usando o modelo
// "all-minilm" servido localmente pelo Ollama, o resultado continua 100%
// local, sem chamada de API paga nem internet — apenas a forma de hospedar o
// modelo muda (processo separado via HTTP, em vez de in-process).
package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Dimension é a dimensionalidade do vetor produzido pelo modelo all-minilm
// (mesma dimensão do all-MiniLM-L6-v2 usado na versão Java).
const Dimension = 384

// Client gera embeddings chamando a API local do Ollama.
type Client struct {
	BaseURL    string
	Model      string
	HTTPClient *http.Client
}

// NewClient cria um Client apontando para baseURL (ex.: http://localhost:11434)
// usando o modelo informado (ex.: "all-minilm").
func NewClient(baseURL, model string) *Client {
	return &Client{
		BaseURL:    baseURL,
		Model:      model,
		HTTPClient: &http.Client{},
	}
}

type embedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type embedResponse struct {
	Embedding []float64 `json:"embedding"`
}

// Embed retorna o vetor de embedding de text.
func (c *Client) Embed(ctx context.Context, text string) ([]float64, error) {
	body, err := json.Marshal(embedRequest{Model: c.Model, Prompt: text})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("chamando Ollama em %s: %w (rode 'ollama serve' e 'ollama pull %s')", c.BaseURL, err, c.Model)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama retornou status %d ao gerar embedding", resp.StatusCode)
	}

	var out embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if len(out.Embedding) == 0 {
		return nil, fmt.Errorf("Ollama retornou embedding vazio")
	}

	return out.Embedding, nil
}
