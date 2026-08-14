package agentloop

import (
	"context"
	"testing"

	"agent-loop-framework/core"
	"agent-loop-framework/observability"
	"agent-loop-framework/tool"
)

// fakeChatClient sempre retorna a mesma decisão de plano válida (nunca
// sinaliza done) — equivalente ao mock de Planner usado no cenário 1 de
// AgentLoopTest.java.
type fakeChatClient struct {
	calls int
}

func (f *fakeChatClient) Chat(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	f.calls++
	return `{"reasoning":"Analisando metricas","action":"Buscar metricas do servico","tool":"getMetrics","args":{"service":"api-gateway"},"successCriteria":"Obter p99 < 1000ms","done":false}`, nil
}

func TestRun_RespectsMaxSteps(t *testing.T) {
	client := &fakeChatClient{}
	contract := core.NewAgentContract()
	contract.MaxSteps = 3
	contract.MaxTimeSeconds = 120
	contract.NoProgressSteps = 10 // desativa deteccao de no-progress

	loop := &AgentLoop{
		Planner:        &core.Planner{Client: client},
		Executor:       core.NewExecutor([]tool.AgentTool{tool.GetMetricsTool{}}),
		CircuitBreaker: core.NewCircuitBreaker(),
		Telemetry:      &observability.AgentTelemetry{},
		Contract:       contract,
	}

	trace := loop.Run(context.Background(), "Diagnosticar degradacao no api-gateway")

	if client.calls != 3 {
		t.Errorf("esperava 3 chamadas ao planner (maxSteps=3), obteve %d", client.calls)
	}
	if trace == nil {
		t.Fatal("trace nao deveria ser nil")
	}
	if trace.TotalSteps != 3 {
		t.Errorf("esperava totalSteps=3, obteve %d", trace.TotalSteps)
	}
}
