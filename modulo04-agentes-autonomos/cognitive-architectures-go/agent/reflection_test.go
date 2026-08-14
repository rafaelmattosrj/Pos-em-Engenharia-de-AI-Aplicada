package agent

import (
	"context"
	"strconv"
	"strings"
	"testing"
)

// routingChatClient inspeciona o prompt para decidir se é uma chamada de
// geração (ReflectionAgent) ou de avaliação (CritiqueEvaluator) — permite
// testar o ciclo completo com um único duplo de teste.
type routingChatClient struct {
	generationCalls int
	critiqueCalls   int
}

func (c *routingChatClient) Chat(ctx context.Context, userPrompt string) (string, error) {
	if strings.Contains(userPrompt, "avaliador crítico") {
		c.critiqueCalls++
		// Sempre reprova, forcando o loop a rodar ate maxCycles.
		return "CORRECTNESS: 0.4\nCOMPLETENESS: 0.4\nQUALITY: 0.4\nFEEDBACK: Ainda precisa melhorar.", nil
	}
	c.generationCalls++
	return "resposta gerada no ciclo " + strconv.Itoa(c.generationCalls), nil
}

// Cenário 2 de CognitiveArchitecturesTest.java: ReflectionAgent deve
// encerrar após maxCycles mesmo que o score nunca passe no threshold.
func TestReflectionAgent_StopsAfterMaxCyclesWithoutPassing(t *testing.T) {
	client := &routingChatClient{}
	critique := &CritiqueEvaluator{Client: client, Threshold: 0.7}
	reflectionAgent := &ReflectionAgent{Client: client, Critique: critique, MaxCycles: 3}

	response, err := reflectionAgent.Execute(context.Background(), "Tarefa muito difícil")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if response.Metrics.Steps != 3 {
		t.Errorf("esperava steps=3 (maxCycles), obteve %d", response.Metrics.Steps)
	}
	if response.Metrics.Reflections != 2 {
		t.Errorf("esperava reflections=2 (maxCycles-1), obteve %d", response.Metrics.Reflections)
	}
	if response.Result == "" {
		t.Error("esperava resultado nao vazio")
	}
	if client.generationCalls != 3 {
		t.Errorf("esperava 3 chamadas de geracao, obteve %d", client.generationCalls)
	}
	if client.critiqueCalls != 3 {
		t.Errorf("esperava 3 chamadas de critica, obteve %d", client.critiqueCalls)
	}
}

func TestReflectionAgent_StopsEarlyWhenApproved(t *testing.T) {
	callCount := 0
	client := &fakeApprovingClient{approveOnCycle: 2, callCount: &callCount}
	critique := &CritiqueEvaluator{Client: client, Threshold: 0.7}
	reflectionAgent := &ReflectionAgent{Client: client, Critique: critique, MaxCycles: 5}

	response, err := reflectionAgent.Execute(context.Background(), "tarefa simples")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	if response.Metrics.Steps != 2 {
		t.Errorf("esperava parar no ciclo 2, obteve steps=%d", response.Metrics.Steps)
	}
	if response.Metrics.Reflections != 1 {
		t.Errorf("esperava reflections=1, obteve %d", response.Metrics.Reflections)
	}
}

type fakeApprovingClient struct {
	approveOnCycle int
	cycle          int
	callCount      *int
}

func (c *fakeApprovingClient) Chat(ctx context.Context, userPrompt string) (string, error) {
	*c.callCount++
	if strings.Contains(userPrompt, "avaliador crítico") {
		if c.cycle >= c.approveOnCycle {
			return "CORRECTNESS: 0.9\nCOMPLETENESS: 0.9\nQUALITY: 0.9\nFEEDBACK: Otimo.", nil
		}
		return "CORRECTNESS: 0.3\nCOMPLETENESS: 0.3\nQUALITY: 0.3\nFEEDBACK: Fraco.", nil
	}
	c.cycle++
	return "output do ciclo", nil
}
