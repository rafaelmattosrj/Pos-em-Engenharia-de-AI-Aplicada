package node

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mcp-sales-analyzer/model"
	"mcp-sales-analyzer/openai"
)

func newIntentNode(t *testing.T, llmContent string) *IntentNode {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]string{"content": llmContent}}},
		})
	}))
	t.Cleanup(server.Close)

	return &IntentNode{Client: &openai.Client{APIKey: "test-key", Model: "gpt-4o-mini", BaseURL: server.URL, HTTPClient: server.Client()}}
}

func TestExtract_ValidJSONResponse(t *testing.T) {
	n := newIntentNode(t, `{"dataType":"csv","parsedData":"produto,preco\nA,10","question":"qual o total?","suggestedTools":["csvToJson"]}`)

	result := n.Extract(context.Background(), model.AnalysisRequest{Question: "qual o total?", Data: "produto,preco\nA,10"})
	if result.DataType != "csv" {
		t.Errorf("esperava dataType=csv, obteve %q", result.DataType)
	}
	if len(result.SuggestedTools) != 1 || result.SuggestedTools[0] != "csvToJson" {
		t.Errorf("suggestedTools inesperado: %v", result.SuggestedTools)
	}
}

func TestExtract_MarkdownFencedJSON(t *testing.T) {
	n := newIntentNode(t, "```json\n{\"dataType\":\"json\",\"parsedData\":\"[]\",\"question\":\"q\",\"suggestedTools\":[]}\n```")

	result := n.Extract(context.Background(), model.AnalysisRequest{Question: "q", Data: "[]"})
	if result.DataType != "json" {
		t.Errorf("esperava dataType=json, obteve %q", result.DataType)
	}
}

func TestExtract_InvalidJSONFallsBackToHeuristic(t *testing.T) {
	n := newIntentNode(t, "isso nao e JSON valido")

	result := n.Extract(context.Background(), model.AnalysisRequest{Question: "q", Data: "a,b\n1,2"})
	if result.DataType != "csv" {
		t.Errorf("esperava fallback heuristico dataType=csv, obteve %q", result.DataType)
	}
	if result.ParsedData != "a,b\n1,2" {
		t.Errorf("esperava dados originais preservados no fallback, obteve %q", result.ParsedData)
	}
}

func TestDetectDataType(t *testing.T) {
	cases := map[string]string{
		"":           "unknown",
		"   ":        "unknown",
		"[1,2,3]":    "json",
		`{"a":1}`:    "json",
		"a,b\n1,2":   "csv",
		"nome,idade": "csv",
	}
	for input, expected := range cases {
		if got := detectDataType(input); got != expected {
			t.Errorf("detectDataType(%q) = %q, esperava %q", input, got, expected)
		}
	}
}
