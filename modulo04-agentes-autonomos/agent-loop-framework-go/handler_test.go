package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"agent-loop-framework/agentloop"
	"agent-loop-framework/core"
	"agent-loop-framework/observability"
	"agent-loop-framework/openai"
	"agent-loop-framework/tool"
)

func TestRunHandler_MissingInputReturns400(t *testing.T) {
	loop := &agentloop.AgentLoop{}

	body, _ := json.Marshal(map[string]string{"input": ""})
	req := httptest.NewRequest(http.MethodPost, "/agent/run", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	runHandler(loop).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400, obteve %d", rec.Code)
	}
}

func TestRunHandler_CompletesWhenLLMSignalsDone(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]string{
				"content": `{"reasoning":"tudo certo","action":"finalizar","tool":null,"args":{},"successCriteria":"N/A","done":true}`,
			}}},
		})
	}))
	defer upstream.Close()

	client := &openai.Client{APIKey: "test-key", Model: "gpt-4o-mini", BaseURL: upstream.URL, HTTPClient: upstream.Client()}

	loop := &agentloop.AgentLoop{
		Planner:        &core.Planner{Client: client},
		Executor:       core.NewExecutor([]tool.AgentTool{tool.GetMetricsTool{}}),
		CircuitBreaker: core.NewCircuitBreaker(),
		Telemetry:      &observability.AgentTelemetry{},
		Contract:       core.NewAgentContract(),
	}

	body, _ := json.Marshal(map[string]string{"input": "verificar status do servico"})
	req := httptest.NewRequest(http.MethodPost, "/agent/run", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	runHandler(loop).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d: %s", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	result, _ := resp["result"].(string)
	if result == "" {
		t.Error("esperava campo 'result' preenchido")
	}
}

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/agent/health", nil)
	rec := httptest.NewRecorder()

	healthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d", rec.Code)
	}
}
