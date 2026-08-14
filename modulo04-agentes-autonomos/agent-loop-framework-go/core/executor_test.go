package core

import (
	"testing"

	"agent-loop-framework/model"
	"agent-loop-framework/tool"
)

func TestExecutor_ReturnsErrorForUnknownTool(t *testing.T) {
	executor := NewExecutor(nil)

	plan := model.PlanDecision{
		Tool: "toolQueNaoExiste",
		Args: map[string]any{"param": "valor"},
	}

	result := executor.Execute(plan)

	if result.Success {
		t.Error("esperava success=false para tool desconhecida")
	}
	if result.ToolName != "toolQueNaoExiste" {
		t.Errorf("esperava toolName preservado, obteve %q", result.ToolName)
	}
	if result.Error == "" {
		t.Error("esperava mensagem de erro nao vazia")
	}
	if result.Output != "" {
		t.Errorf("esperava output vazio em caso de erro, obteve %q", result.Output)
	}
}

func TestExecutor_ExecutesRegisteredTool(t *testing.T) {
	executor := NewExecutor([]tool.AgentTool{tool.GetMetricsTool{}})

	result := executor.Execute(model.PlanDecision{Tool: "getMetrics", Args: map[string]any{"service": "api-gateway"}})

	if !result.Success {
		t.Fatalf("esperava sucesso, obteve erro: %s", result.Error)
	}
	if result.Output == "" {
		t.Error("esperava output nao vazio")
	}
}
