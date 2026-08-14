package observability

import (
	"testing"

	"agent-loop-framework/core"
	"agent-loop-framework/model"
)

func TestTelemetry_StartRecordFinalize(t *testing.T) {
	telemetry := &AgentTelemetry{}
	telemetry.Start("tarefa de teste")

	plan := model.PlanDecision{Tool: "getMetrics", Done: false}
	toolResult := model.OkResult("getMetrics", "{}")
	telemetry.RecordStep(1, "percepcao inicial", plan, &toolResult, "avaliacao ok")

	state := core.NewAgentState()
	state.IncrementStep()
	state.SetDone(true)

	trace := telemetry.Finalize(state, 1234)

	if trace.Input != "tarefa de teste" {
		t.Errorf("input inesperado: %q", trace.Input)
	}
	if len(trace.Steps) != 1 {
		t.Fatalf("esperava 1 step registrado, obteve %d", len(trace.Steps))
	}
	if trace.DurationMs != 1234 {
		t.Errorf("durationMs inesperado: %d", trace.DurationMs)
	}
	if trace.TotalSteps != 1 {
		t.Errorf("totalSteps inesperado: %d", trace.TotalSteps)
	}
}

func TestTelemetry_RecordStepWithoutStartIsIgnored(t *testing.T) {
	telemetry := &AgentTelemetry{}
	telemetry.RecordStep(1, "x", model.PlanDecision{}, nil, "eval")

	if telemetry.Trace() != nil {
		t.Error("trace deveria continuar nil sem Start() ter sido chamado")
	}
}
