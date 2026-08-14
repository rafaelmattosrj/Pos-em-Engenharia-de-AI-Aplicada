package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mcp-sales-analyzer/graph"
	"mcp-sales-analyzer/model"
	"mcp-sales-analyzer/node"
	"mcp-sales-analyzer/openai"
)

func TestAnalyzeHandler_Success(t *testing.T) {
	round := 0
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		round++
		switch round {
		case 1: // IntentNode
			json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{{"message": map[string]string{
					"content": `{"dataType":"csv","parsedData":"produto,preco\nA,10","question":"qual o total?","suggestedTools":[]}`,
				}}},
			})
		case 2: // ExecutorNode, resposta final sem tool call
			json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{{"message": map[string]string{"content": "o total e 10"}}},
			})
		}
	}))
	defer upstream.Close()

	client := &openai.Client{APIKey: "test-key", Model: "gpt-4o-mini", BaseURL: upstream.URL, HTTPClient: upstream.Client()}
	orchestrator := &graph.Orchestrator{
		IntentNode:   &node.IntentNode{Client: client},
		ExecutorNode: &node.ExecutorNode{Client: client},
	}

	body, _ := json.Marshal(model.AnalysisRequest{Question: "qual o total?", Data: "produto,preco\nA,10"})
	req := httptest.NewRequest(http.MethodPost, "/analyze", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	analyzeHandler(orchestrator).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d: %s", rec.Code, rec.Body.String())
	}

	var result model.AnalysisResult
	json.NewDecoder(rec.Body).Decode(&result)
	if result.Report != "o total e 10" {
		t.Errorf("report inesperado: %q", result.Report)
	}
}

func TestAnalyzeHandler_MethodNotAllowed(t *testing.T) {
	orchestrator := &graph.Orchestrator{}

	req := httptest.NewRequest(http.MethodGet, "/analyze", nil)
	rec := httptest.NewRecorder()

	analyzeHandler(orchestrator).ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("esperava 405, obteve %d", rec.Code)
	}
}
