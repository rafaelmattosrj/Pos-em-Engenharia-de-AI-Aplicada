package agent

import (
	"context"
	"testing"
)

func TestPlanExecuteAgent_ParsesPlanAndExecutesSteps(t *testing.T) {
	client := &stubChatClient{responses: []string{
		`{"steps":[{"stepNumber":1,"description":"buscar dados","tool":"search","args":{"q":"clima"},"successCriteria":"dados encontrados"},{"stepNumber":2,"description":"resumir","tool":"summarize","args":{},"successCriteria":"resumo pronto"}]}`,
	}}
	planAgent := &PlanExecuteAgent{Client: client}

	response, err := planAgent.Execute(context.Background(), "pesquisar e resumir o clima")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if response.Metrics.Steps != 2 {
		t.Errorf("esperava 2 steps, obteve %d", response.Metrics.Steps)
	}
	if response.Metrics.Reflections != 0 {
		t.Errorf("esperava reflections=0, obteve %d", response.Metrics.Reflections)
	}
	if client.calls != 1 {
		t.Errorf("esperava exatamente 1 chamada ao LLM (so planejamento), obteve %d", client.calls)
	}
}

func TestPlanExecuteAgent_FallsBackOnInvalidJSON(t *testing.T) {
	client := &stubChatClient{responses: []string{"isso nao e JSON valido"}}
	planAgent := &PlanExecuteAgent{Client: client}

	response, err := planAgent.Execute(context.Background(), "tarefa qualquer")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if response.Metrics.Steps != 1 {
		t.Errorf("esperava plano de fallback com 1 step, obteve %d", response.Metrics.Steps)
	}
}
