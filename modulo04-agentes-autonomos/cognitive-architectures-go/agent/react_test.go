package agent

import (
	"context"
	"testing"
)

func TestReactAgent_StopsOnFinalAnswer(t *testing.T) {
	client := &stubChatClient{responses: []string{
		"Thought: preciso pensar\nAction: FINAL_ANSWER(Brasília é a capital do Brasil)",
	}}
	reactAgent := &ReactAgent{Client: client, MaxSteps: 10}

	response, err := reactAgent.Execute(context.Background(), "Qual a capital do Brasil?")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if response.Result != "Brasília é a capital do Brasil" {
		t.Errorf("resultado inesperado: %q", response.Result)
	}
	if response.Metrics.Steps != 1 {
		t.Errorf("esperava 1 step, obteve %d", response.Metrics.Steps)
	}
}

func TestReactAgent_StopsOnRepeatedAction(t *testing.T) {
	client := &stubChatClient{responses: []string{
		"Thought: tentando\nAction: search(clima)",
	}}
	reactAgent := &ReactAgent{Client: client, MaxSteps: 10}

	response, err := reactAgent.Execute(context.Background(), "tarefa qualquer")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	// Mesma action 3x consecutivas -> encerra no step 3 (com 1 repeticao antes -> na 3a ocorrencia total).
	if response.Metrics.Steps > 4 {
		t.Errorf("esperava parada por loop bem antes do maxSteps, obteve steps=%d", response.Metrics.Steps)
	}
}

func TestReactAgent_ReachesMaxStepsWithoutFinalAnswer(t *testing.T) {
	call := 0
	client := &varyingActionClient{}
	_ = call
	reactAgent := &ReactAgent{Client: client, MaxSteps: 3}

	response, err := reactAgent.Execute(context.Background(), "tarefa sem solucao")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if response.Metrics.Steps != 3 {
		t.Errorf("esperava 3 steps (maxSteps), obteve %d", response.Metrics.Steps)
	}
}

type varyingActionClient struct{ n int }

func (c *varyingActionClient) Chat(ctx context.Context, userPrompt string) (string, error) {
	c.n++
	return "Thought: tentando algo diferente\nAction: tool" + string(rune('A'+c.n)) + "()", nil
}
