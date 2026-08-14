package reactagent

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

// OllamaReactClient e o cliente real do modelo, via API nativa do Ollama
// local (POST /api/chat, sem streaming, com "tools") — equivalente a
// ollama.chat(...) usado em react-agent-prototype.js / .py. Sem chave de
// API: Ollama roda 100% local.
type OllamaReactClient struct {
	BaseURL    string
	Modelo     string
	HTTPClient *http.Client
}

// NewOllamaReactClient cria um client apontando pro baseURL informado
// (ex.: "http://localhost:11434", sem sufixo /api).
func NewOllamaReactClient(baseURL, modelo string) *OllamaReactClient {
	return &OllamaReactClient{
		BaseURL:    baseURL,
		Modelo:     modelo,
		HTTPClient: &http.Client{Timeout: 60 * time.Second},
	}
}

type chatRequestBody struct {
	Model    string     `json:"model"`
	Messages []Mensagem `json:"messages"`
	Tools    []Mensagem `json:"tools,omitempty"`
	Stream   bool       `json:"stream"`
}

type chatResponseBody struct {
	Message Mensagem `json:"message"`
	Error   string   `json:"error,omitempty"`
}

// Chat implementa ChatClient chamando a API nativa do Ollama.
func (c *OllamaReactClient) Chat(ctx context.Context, historico []Mensagem, tools []Mensagem) (ModeloResposta, error) {
	reqBody := chatRequestBody{Model: c.Modelo, Messages: historico, Tools: tools, Stream: false}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return ModeloResposta{}, fmt.Errorf("ollama: falha ao serializar request: %w", err)
	}

	url := c.BaseURL + "/api/chat"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return ModeloResposta{}, fmt.Errorf("ollama: falha ao montar request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return ModeloResposta{}, fmt.Errorf(
			"ollama: falha na requisição (Ollama está rodando? 'ollama serve'): %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ModeloResposta{}, fmt.Errorf("ollama: falha ao ler resposta: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return ModeloResposta{}, &ModeloHTTPError{StatusCode: resp.StatusCode, Body: string(body)}
	}

	var parsed chatResponseBody
	if err := json.Unmarshal(body, &parsed); err != nil {
		return ModeloResposta{}, fmt.Errorf("ollama: resposta inválida (status %d): %w", resp.StatusCode, err)
	}

	if parsed.Error != "" {
		return ModeloResposta{}, errors.New("ollama: erro da API: " + parsed.Error)
	}

	if parsed.Message == nil {
		return ModeloResposta{}, errors.New("ollama: resposta sem message")
	}

	content, _ := parsed.Message["content"].(string)

	var chamadas []ChamadaFerramenta
	if rawToolCalls, ok := parsed.Message["tool_calls"].([]interface{}); ok {
		for _, rawTC := range rawToolCalls {
			tc, ok := rawTC.(map[string]interface{})
			if !ok {
				continue
			}
			id, _ := tc["id"].(string)
			function, _ := tc["function"].(map[string]interface{})
			nome, _ := function["name"].(string)
			argumentos, _ := function["arguments"].(map[string]interface{})
			chamadas = append(chamadas, ChamadaFerramenta{ID: id, Nome: nome, Argumentos: argumentos})
		}
	}

	return ModeloResposta{
		MensagemBruta: parsed.Message,
		Content:       content,
		Chamadas:      chamadas,
	}, nil
}
