package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cognitive-architectures/agent"
	"cognitive-architectures/model"
)

type spyChatClient struct {
	calls    []string
	response string
}

func (c *spyChatClient) Chat(ctx context.Context, userPrompt string) (string, error) {
	c.calls = append(c.calls, userPrompt)
	return c.response, nil
}

// Cenário 3 de CognitiveArchitecturesTest.java: roteia para ReactAgent
// quando architecture="react", sem tocar nos outros agentes.
func TestRun_RoutesToReactAgent(t *testing.T) {
	reactClient := &spyChatClient{response: "Thought: ok\nAction: FINAL_ANSWER(Resposta ReAct.)"}
	planClient := &spyChatClient{}
	reflectionClient := &spyChatClient{}

	srv := &server{
		react:       &agent.ReactAgent{Client: reactClient, MaxSteps: 10},
		planExecute: &agent.PlanExecuteAgent{Client: planClient},
		reflection: &agent.ReflectionAgent{
			Client:    reflectionClient,
			Critique:  &agent.CritiqueEvaluator{Client: reflectionClient, Threshold: 0.7},
			MaxCycles: 3,
		},
	}

	body, _ := json.Marshal(model.AgentRequest{Architecture: "react", Input: "Qual é a capital do Brasil?"})
	req := httptest.NewRequest(http.MethodPost, "/agents/run", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	srv.run(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d: %s", rec.Code, rec.Body.String())
	}

	var resp model.AgentResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Result != "Resposta ReAct." {
		t.Errorf("resultado inesperado: %q", resp.Result)
	}
	if resp.Metrics.Reflections != 0 {
		t.Errorf("esperava reflections=0, obteve %d", resp.Metrics.Reflections)
	}

	if len(planClient.calls) != 0 || len(reflectionClient.calls) != 0 {
		t.Error("apenas o ReactAgent deveria ter sido chamado")
	}
}

// Cenário 4: arquitetura desconhecida retorna 200 com mensagem de erro no
// body, sem chamar nenhum agente.
func TestRun_UnknownArchitectureReturnsGracefulError(t *testing.T) {
	reactClient := &spyChatClient{}
	planClient := &spyChatClient{}
	reflectionClient := &spyChatClient{}

	srv := &server{
		react:       &agent.ReactAgent{Client: reactClient},
		planExecute: &agent.PlanExecuteAgent{Client: planClient},
		reflection: &agent.ReflectionAgent{
			Client:   reflectionClient,
			Critique: &agent.CritiqueEvaluator{Client: reflectionClient, Threshold: 0.7},
		},
	}

	body, _ := json.Marshal(model.AgentRequest{Architecture: "arquitetura-inexistente", Input: "Alguma tarefa"})
	req := httptest.NewRequest(http.MethodPost, "/agents/run", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	srv.run(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d", rec.Code)
	}

	var resp model.AgentResponse
	json.NewDecoder(rec.Body).Decode(&resp)
	if !strings.Contains(resp.Result, "Arquitetura desconhecida") {
		t.Errorf("esperava mensagem de arquitetura desconhecida, obteve %q", resp.Result)
	}
	if resp.Metrics.Steps != 0 {
		t.Errorf("esperava steps=0, obteve %d", resp.Metrics.Steps)
	}

	if len(reactClient.calls) != 0 || len(planClient.calls) != 0 || len(reflectionClient.calls) != 0 {
		t.Error("nenhum agente deveria ter sido chamado")
	}
}

func TestRun_MissingArchitectureReturns400(t *testing.T) {
	srv := &server{}

	body, _ := json.Marshal(model.AgentRequest{Input: "tarefa"})
	req := httptest.NewRequest(http.MethodPost, "/agents/run", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	srv.run(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400, obteve %d", rec.Code)
	}
}
