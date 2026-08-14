package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"trialforge-gateway-prototype/gateway"
)

// Client fala com a API nativa do Ollama. Implementa gateway.Embedder e
// gateway.ChatStreamer.
type Client struct {
	BaseURL        string // ex.: http://localhost:11434
	EmbeddingModel string // nomic-embed-text nos dois originais
	HTTPClient     *http.Client
	MaxTentativas  int
	Timeout        time.Duration
	Log            func(format string, args ...interface{})
}

// NewClient cria um Client com os defaults dos dois originais: 3 tentativas,
// timeout de 20s por tentativa.
func NewClient(baseURL, embeddingModel string) *Client {
	return &Client{
		BaseURL:        baseURL,
		EmbeddingModel: embeddingModel,
		HTTPClient:     &http.Client{},
		MaxTentativas:  maxTentativasPadrao,
		Timeout:        timeoutPadrao,
	}
}

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

func (c *Client) tentativas() int {
	if c.MaxTentativas > 0 {
		return c.MaxTentativas
	}
	return maxTentativasPadrao
}

func (c *Client) timeout() time.Duration {
	if c.Timeout > 0 {
		return c.Timeout
	}
	return timeoutPadrao
}

// ---------- Embeddings reais (base de Semantic Cache e de confiança do RAG) ----------

type embedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type embedResponse struct {
	Embedding []float64 `json:"embedding"`
}

// Embedar chama POST /api/embeddings, com retry+timeout (comRetry).
func (c *Client) Embedar(ctx context.Context, texto string) ([]float64, error) {
	return comRetry(func() ([]float64, error) {
		return c.doEmbed(ctx, texto)
	}, c.tentativas(), c.timeout(), "Embedding", c.Log)
}

func (c *Client) doEmbed(ctx context.Context, texto string) ([]float64, error) {
	payload, err := json.Marshal(embedRequest{Model: c.EmbeddingModel, Prompt: texto})
	if err != nil {
		return nil, fmt.Errorf("ollama: falha ao serializar request de embedding: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/embeddings", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("ollama: falha ao montar request de embedding: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama: falha na requisição de embedding (Ollama está rodando? 'ollama serve'): %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ollama: falha ao ler resposta de embedding: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama: erro da API de embedding (status %d): %s", resp.StatusCode, string(body))
	}

	var parsed embedResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("ollama: resposta de embedding inválida (status %d): %w", resp.StatusCode, err)
	}
	if len(parsed.Embedding) == 0 {
		return nil, errors.New("ollama: resposta de embedding sem vetor")
	}
	return parsed.Embedding, nil
}

// ---------- Geração com streaming (Módulo 4.3) ----------

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type chatStreamLine struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	Done  bool   `json:"done"`
	Error string `json:"error,omitempty"`
}

// ChatStream chama POST /api/chat com stream=true. O retry+timeout
// (comRetry) protege só o ESTABELECIMENTO da conexão (obter os headers da
// resposta) — mesmo alcance do comRetry(() => ollama.chat(...)) no JS e do
// com_retry(lambda: chat(...)) no Python: o Promise/generator ali também só
// representa o início da chamada, não a iteração completa dos chunks. Depois
// de estabelecida, a leitura dos chunks NDJSON não é retentada — um chunk
// corrompido no meio do stream propaga o erro pra cima, igual aos originais
// (que também não têm retry por chunk).
func (c *Client) ChatStream(ctx context.Context, model string, messages []gateway.Message, onChunk func(content string)) error {
	msgs := make([]chatMessage, len(messages))
	for i, m := range messages {
		msgs[i] = chatMessage{Role: m.Role, Content: m.Content}
	}
	payload, err := json.Marshal(chatRequest{Model: model, Messages: msgs, Stream: true})
	if err != nil {
		return fmt.Errorf("ollama: falha ao serializar request de chat: %w", err)
	}

	resp, err := comRetry(func() (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/chat", bytes.NewReader(payload))
		if err != nil {
			return nil, fmt.Errorf("ollama: falha ao montar request de chat: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient().Do(req)
		if err != nil {
			return nil, fmt.Errorf("ollama: falha na requisição de chat (Ollama está rodando? 'ollama serve'): %w", err)
		}
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return nil, fmt.Errorf("ollama: erro da API de chat (status %d): %s", resp.StatusCode, string(body))
		}
		return resp, nil
	}, c.tentativas(), c.timeout(), "Geração de resposta", c.Log)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		linha := bytes.TrimSpace(scanner.Bytes())
		if len(linha) == 0 {
			continue
		}
		var parte chatStreamLine
		if err := json.Unmarshal(linha, &parte); err != nil {
			return fmt.Errorf("ollama: linha de streaming inválida: %w", err)
		}
		if parte.Error != "" {
			return fmt.Errorf("ollama: erro durante streaming: %s", parte.Error)
		}
		// Modelo de raciocínio: chunks de "thinking" vêm num campo separado,
		// nunca em message.content — por isso só repassamos .content, igual
		// aos dois originais.
		if parte.Message.Content != "" {
			onChunk(parte.Message.Content)
		}
		if parte.Done {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("ollama: falha ao ler stream de chat: %w", err)
	}
	return nil
}
