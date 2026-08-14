// Package observability coleta e exporta a telemetria do agent loop —
// equivalente a AgentTelemetry.java e TraceExporter.java.
package observability

import (
	"fmt"
	"log"

	"agent-loop-framework/core"
	"agent-loop-framework/model"
)

// AgentTelemetry coleta telemetria de cada step do agente durante a
// execução — equivalente a AgentTelemetry.java.
type AgentTelemetry struct {
	currentTrace *model.AgentTrace
}

// Start inicializa a coleta de telemetria para uma nova execução.
func (t *AgentTelemetry) Start(input string) {
	t.currentTrace = model.NewAgentTrace(input)
}

// RecordStep registra os dados de um step individual no trace.
func (t *AgentTelemetry) RecordStep(stepNumber int, perception string, plan model.PlanDecision, toolResult *model.ToolResult, evaluation string) {
	if t.currentTrace == nil {
		log.Printf("[Telemetria] RecordStep chamado sem Start() — ignorando step %d", stepNumber)
		return
	}

	t.currentTrace.AddStep(model.StepTrace{
		StepNumber: stepNumber,
		Perception: perception,
		Plan:       plan,
		ToolResult: toolResult,
		Evaluation: evaluation,
	})
}

// Finalize finaliza o trace com as métricas de duração e resultado final.
func (t *AgentTelemetry) Finalize(state *core.AgentState, durationMs int64) *model.AgentTrace {
	if t.currentTrace == nil {
		log.Println("[Telemetria] Finalize chamado sem trace ativo")
		return model.NewAgentTrace("unknown")
	}

	t.currentTrace.TotalSteps = state.CurrentStep()
	t.currentTrace.DurationMs = durationMs

	// Estimativa simples de tokens: ~4 chars por token no contexto acumulado.
	t.currentTrace.TotalTokensEstimated = len(state.AccumulatedContext()) / 4

	if state.Done() {
		t.currentTrace.FinalResult = fmt.Sprintf("Tarefa concluída após %d steps", state.CurrentStep())
	} else {
		t.currentTrace.FinalResult = fmt.Sprintf("Execução interrompida após %d steps (limite/timeout/circuitbreaker)", state.CurrentStep())
	}

	return t.currentTrace
}

// Trace retorna o trace da execução atual (pode ser nil antes de Start()).
func (t *AgentTelemetry) Trace() *model.AgentTrace {
	return t.currentTrace
}
