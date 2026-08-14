// Package gemini é um cliente mínimo para a API REST do Gemini
// (generateContent) — substitui o plugin @genkit-ai/google-genai usado por
// flows.ts. Solicita saída em JSON (responseMimeType: application/json) no
// lugar da validação de schema embutida do Genkit — o parsing do JSON
// retornado é feito pelo chamador (service.BragService).
package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta/models"

// Client fala com a API generateContent do Gemini.
type Client struct {
	APIKey     string
	Model      string
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient cria um Client com timeout de 60s.
func NewClient(apiKey, model string) *Client {
	return &Client{
		APIKey:     apiKey,
		Model:      model,
		BaseURL:    defaultBaseURL,
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
	}
}

type part struct {
	Text string `json:"text"`
}

type content struct {
	Parts []part `json:"parts"`
}

type generationConfig struct {
	Temperature      float64 `json:"temperature"`
	ResponseMIMEType string  `json:"responseMimeType"`
}

type generateRequest struct {
	Contents         []content        `json:"contents"`
	GenerationConfig generationConfig `json:"generationConfig"`
}

type candidate struct {
	Content content `json:"content"`
}

type generateResponse struct {
	Candidates []candidate `json:"candidates"`
	Error      *apiError   `json:"error,omitempty"`
}

type apiError struct {
	Message string `json:"message"`
}

// GenerateJSON gera conteúdo em JSON para o prompt informado, na
// temperatura dada, e retorna o texto JSON bruto (ainda não parseado).
func (c *Client) GenerateJSON(ctx context.Context, prompt string, temperature float64) (string, error) {
	reqBody := generateRequest{
		Contents: []content{{Parts: []part{{Text: prompt}}}},
		GenerationConfig: generationConfig{
			Temperature:      temperature,
			ResponseMIMEType: "application/json",
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("gemini: falha ao serializar request: %w", err)
	}

	baseURL := c.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	url := fmt.Sprintf("%s/%s:generateContent?key=%s", baseURL, c.Model, c.APIKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("gemini: falha ao montar request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini: falha na requisicao: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("gemini: falha ao ler resposta: %w", err)
	}

	var parsed generateResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("gemini: resposta invalida (status %d): %w", resp.StatusCode, err)
	}

	if resp.StatusCode != http.StatusOK {
		if parsed.Error != nil {
			return "", fmt.Errorf("gemini: erro da API (status %d): %s", resp.StatusCode, parsed.Error.Message)
		}
		return "", fmt.Errorf("gemini: erro da API (status %d)", resp.StatusCode)
	}

	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini: resposta sem candidates")
	}

	return parsed.Candidates[0].Content.Parts[0].Text, nil
}
