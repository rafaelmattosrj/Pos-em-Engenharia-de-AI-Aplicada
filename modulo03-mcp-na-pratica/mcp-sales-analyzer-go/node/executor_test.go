package node

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"mcp-sales-analyzer/model"
	"mcp-sales-analyzer/openai"
)

func TestExecute_CsvDataTriggersToolCall(t *testing.T) {
	round := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		round++
		bodyBytes, _ := io.ReadAll(r.Body)
		body := string(bodyBytes)
		if round == 1 {
			if !strings.Contains(body, "csvToJson") {
				t.Errorf("esperava tool csvToJson na primeira chamada, body=%s", body)
			}
			json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{{"message": map[string]any{
					"role": "assistant",
					"tool_calls": []map[string]any{
						{"id": "call-1", "type": "function", "function": map[string]string{
							"name": "csvToJson", "arguments": `{"csv":"produto,preco\nA,10"}`,
						}},
					},
				}}},
			})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]string{"content": "produto A custa 10"}}},
		})
	}))
	defer server.Close()

	executor := &ExecutorNode{Client: &openai.Client{APIKey: "test-key", Model: "gpt-4o-mini", BaseURL: server.URL, HTTPClient: server.Client()}}

	result, err := executor.Execute(context.Background(), model.IntentResult{
		DataType: "csv", Question: "qual o preco?", ParsedData: "produto,preco\nA,10",
	})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if result.Report != "produto A custa 10" {
		t.Errorf("report inesperado: %q", result.Report)
	}
	if !contains(result.ToolsUsed, "csvToJson") {
		t.Errorf("esperava csvToJson em toolsUsed, obteve %v", result.ToolsUsed)
	}
	if len(result.ProcessingSteps) != 3 {
		t.Errorf("esperava 3 passos de processamento, obteve %d: %v", len(result.ProcessingSteps), result.ProcessingSteps)
	}
}
