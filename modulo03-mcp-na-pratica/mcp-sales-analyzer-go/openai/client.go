// Package openai é um cliente mínimo para a API de chat completions da
// OpenAI, incluindo suporte a function/tool calling — substitui o
// ChatClient (Spring AI) usado na versão Java.
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

const defaultBaseURL = "https://api.openai.com/v1/chat/completions"

// Client fala com a API de chat completions da OpenAI.
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

// Message é uma mensagem no formato do protocolo de chat completions,
// incluindo os papéis "system", "user", "assistant" e "tool".
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

// Tool descreve uma função disponível para o modelo chamar (function
// calling) — equivalente ao schema gerado a partir de @Tool na versão Java.
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Parameters  any    `json:"parameters"`
}

// ToolCall é uma chamada de função solicitada pelo modelo.
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Tools    []Tool    `json:"tools,omitempty"`
}

type chatChoice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type chatResponse struct {
	Choices []chatChoice `json:"choices"`
	Error   *apiError    `json:"error,omitempty"`
}

type apiError struct {
	Message string `json:"message"`
}

// Chat envia uma única mensagem de usuário (sem tools) e retorna o texto da
// resposta — usado pelo IntentNode, que só precisa de extração estruturada
// via prompt, sem function calling.
func (c *Client) Chat(ctx context.Context, prompt string) (string, error) {
	messages := []Message{{Role: "user", Content: prompt}}
	resp, err := c.complete(ctx, messages, nil)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

// ChatWithTools envia userPrompt com as tools disponíveis e executa o loop
// de function calling: sempre que o modelo solicitar uma tool call, executa
// via toolExecutor e reenvia o resultado, até obter uma resposta final em
// texto — equivalente ao comportamento automático de tool-calling do
// ChatClient do Spring AI (usado por ExecutorNode).
func (c *Client) ChatWithTools(ctx context.Context, userPrompt string, tools []Tool, toolExecutor func(name, argsJSON string) (string, error)) (string, error) {
	messages := []Message{{Role: "user", Content: userPrompt}}

	const maxRounds = 5
	for round := 0; round < maxRounds; round++ {
		resp, err := c.complete(ctx, messages, tools)
		if err != nil {
			return "", err
		}

		message := resp.Choices[0].Message
		if len(message.ToolCalls) == 0 {
			return message.Content, nil
		}

		messages = append(messages, message)
		for _, call := range message.ToolCalls {
			result, err := toolExecutor(call.Function.Name, call.Function.Arguments)
			if err != nil {
				result = fmt.Sprintf("erro ao executar tool %s: %v", call.Function.Name, err)
			}
			messages = append(messages, Message{
				Role:       "tool",
				Content:    result,
				ToolCallID: call.ID,
			})
		}
	}

	return "", fmt.Errorf("openai: numero maximo de rounds de tool calling (%d) excedido", maxRounds)
}

func (c *Client) complete(ctx context.Context, messages []Message, tools []Tool) (*chatResponse, error) {
	reqBody := chatRequest{Model: c.Model, Messages: messages, Tools: tools}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("openai: falha ao serializar request: %w", err)
	}

	url := c.BaseURL
	if url == "" {
		url = defaultBaseURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("openai: falha ao montar request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai: falha na requisicao: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("openai: falha ao ler resposta: %w", err)
	}

	var parsed chatResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, fmt.Errorf("openai: resposta invalida (status %d): %w", resp.StatusCode, err)
	}

	if resp.StatusCode != http.StatusOK {
		if parsed.Error != nil {
			return nil, fmt.Errorf("openai: erro da API (status %d): %s", resp.StatusCode, parsed.Error.Message)
		}
		return nil, fmt.Errorf("openai: erro da API (status %d)", resp.StatusCode)
	}

	if len(parsed.Choices) == 0 {
		return nil, fmt.Errorf("openai: resposta sem choices")
	}

	return &parsed, nil
}
