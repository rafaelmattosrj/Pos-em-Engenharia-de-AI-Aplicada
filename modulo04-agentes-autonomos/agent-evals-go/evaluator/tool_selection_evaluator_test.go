package evaluator

import (
	"context"
	"testing"
)

// alwaysWrongClient simula um agente que sempre escolhe uma ferramenta
// proibida — usado para validar o cenário de reprovação (accuracy < 0.80).
type alwaysWrongClient struct{}

func (alwaysWrongClient) Chat(ctx context.Context, userPrompt string) (string, error) {
	return `{"tool": "notifyTeam", "args": {}}`, nil
}

// perfectClient simula um agente que sempre escolhe a ferramenta correta
// esperada para cada caso, na ordem em que os casos são avaliados.
type perfectClient struct{ call int }

func (c *perfectClient) Chat(ctx context.Context, userPrompt string) (string, error) {
	expected := []string{"getMetrics", "getLogs", "getDeployHistory", "saveIncident", "notifyTeam"}
	tool := expected[c.call%len(expected)]
	c.call++
	return `{"tool": "` + tool + `", "args": {"service": "payment-api"}}`, nil
}

// Cenário 3 de EvalsTest.java (adaptado para exercitar a lógica real de
// avaliação): ToolSelectionReport deve ter passed=false quando accuracy
// fica abaixo do threshold.
func TestEvaluate_FailsWhenAllToolChoicesAreWrong(t *testing.T) {
	evaluator := &ToolSelectionEvaluator{Client: alwaysWrongClient{}, MinAccuracy: 0.80, MaxUnnecessary: 0.10}

	report := evaluator.Evaluate(context.Background())

	if report.Passed {
		t.Error("esperava passed=false quando todas as escolhas estao erradas")
	}
	if report.ToolSelectionAccuracy >= 0.80 {
		t.Errorf("esperava accuracy < 0.80, obteve %f", report.ToolSelectionAccuracy)
	}
	if report.TotalCases != 5 {
		t.Errorf("esperava 5 casos, obteve %d", report.TotalCases)
	}
}

func TestEvaluate_PassesWhenAllToolChoicesAreCorrect(t *testing.T) {
	evaluator := &ToolSelectionEvaluator{Client: &perfectClient{}, MinAccuracy: 0.80, MaxUnnecessary: 0.10}

	report := evaluator.Evaluate(context.Background())

	if !report.Passed {
		t.Errorf("esperava passed=true, obteve report=%+v", report)
	}
	if report.ToolSelectionAccuracy < 0.80 {
		t.Errorf("esperava accuracy >= 0.80, obteve %f", report.ToolSelectionAccuracy)
	}
	if report.UnnecessaryCallsRate > 0.10 {
		t.Errorf("esperava unnecessaryCallsRate <= 0.10, obteve %f", report.UnnecessaryCallsRate)
	}
}

func TestExtractToolName_ParsesToolFromJSON(t *testing.T) {
	name := extractToolName(`{"tool": "getMetrics", "args": {"service": "x"}}`)
	if name != "getMetrics" {
		t.Errorf("esperava 'getMetrics', obteve %q", name)
	}
}

func TestExtractToolName_ReturnsUnknownWhenMissing(t *testing.T) {
	name := extractToolName("resposta sem o campo tool")
	if name != "unknown" {
		t.Errorf("esperava 'unknown', obteve %q", name)
	}
}
