package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChat_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(chatResponse{
			Choices: []chatChoice{{Message: Message{Role: "assistant", Content: "resposta"}}},
		})
	}))
	defer server.Close()

	client := &Client{APIKey: "test-key", Model: "gpt-4o-mini", BaseURL: server.URL, HTTPClient: server.Client()}
	content, err := client.Chat(context.Background(), "pergunta")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if content != "resposta" {
		t.Errorf("esperava 'resposta', obteve %q", content)
	}
}

func TestChatWithTools_NoToolCallReturnsDirectly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(chatResponse{
			Choices: []chatChoice{{Message: Message{Role: "assistant", Content: "relatorio final"}}},
		})
	}))
	defer server.Close()

	client := &Client{APIKey: "test-key", Model: "gpt-4o-mini", BaseURL: server.URL, HTTPClient: server.Client()}
	result, err := client.ChatWithTools(context.Background(), "analise isso", nil, func(name, args string) (string, error) {
		t.Fatal("nao deveria chamar nenhuma tool")
		return "", nil
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result != "relatorio final" {
		t.Errorf("esperava 'relatorio final', obteve %q", result)
	}
}

func TestChatWithTools_ExecutesToolThenReturnsFinalAnswer(t *testing.T) {
	round := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		round++
		if round == 1 {
			json.NewEncoder(w).Encode(chatResponse{
				Choices: []chatChoice{{Message: Message{
					Role: "assistant",
					ToolCalls: []ToolCall{
						{ID: "call-1", Type: "function", Function: ToolCallFunction{Name: "csvToJson", Arguments: `{"csv":"a,b\n1,2"}`}},
					},
				}}},
			})
			return
		}
		json.NewEncoder(w).Encode(chatResponse{
			Choices: []chatChoice{{Message: Message{Role: "assistant", Content: "analise concluida"}}},
		})
	}))
	defer server.Close()

	client := &Client{APIKey: "test-key", Model: "gpt-4o-mini", BaseURL: server.URL, HTTPClient: server.Client()}

	var executedName, executedArgs string
	result, err := client.ChatWithTools(context.Background(), "analise este csv", []Tool{{Type: "function", Function: ToolFunction{Name: "csvToJson"}}},
		func(name, args string) (string, error) {
			executedName, executedArgs = name, args
			return `[{"a":"1","b":"2"}]`, nil
		})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result != "analise concluida" {
		t.Errorf("esperava 'analise concluida', obteve %q", result)
	}
	if executedName != "csvToJson" {
		t.Errorf("esperava tool csvToJson executada, obteve %q", executedName)
	}
	if executedArgs != `{"csv":"a,b\n1,2"}` {
		t.Errorf("argumentos inesperados: %q", executedArgs)
	}
	if round != 2 {
		t.Errorf("esperava 2 rounds de chamada ao modelo, obteve %d", round)
	}
}
