// Package agentloop orquestra o ciclo completo
// Percepção→Planejamento→Ação→Avaliação — equivalente a AgentLoop.java.
// Fica em pacote próprio (em vez de dentro de core/) porque depende tanto de
// core quanto de observability, e observability já depende de core.
package agentloop

import (
	"context"
	"fmt"
	"log"
	"time"

	"agent-loop-framework/core"
	"agent-loop-framework/model"
	"agent-loop-framework/observability"
)

// AgentLoop executa o ciclo completo do agente.
type AgentLoop struct {
	Planner        *core.Planner
	Executor       *core.Executor
	CircuitBreaker *core.CircuitBreaker
	Telemetry      *observability.AgentTelemetry
	Contract       core.AgentContract
}

// Run executa o agent loop completo para o input fornecido, com os mesmos
// critérios de parada da versão Java:
//  1. done=true na PlanDecision
//  2. maxSteps atingido
//  3. maxTime atingido
//  4. noProgress detectado (mesma tool repetida N vezes consecutivas)
//  5. CircuitBreaker aberto (3 respostas inválidas do LLM)
func (l *AgentLoop) Run(ctx context.Context, input string) *model.AgentTrace {
	log.Printf("[AgentLoop] Iniciando execução para: %s", input)

	state := core.NewAgentState()
	l.CircuitBreaker.Reset()
	l.Telemetry.Start(input)

	startTime := time.Now()
	state.AddContext("Tarefa recebida: " + input)

	for {
		step := state.IncrementStep()
		log.Printf("[AgentLoop] === Step %d ===", step)

		// 1. PERCEPÇÃO
		perception := state.AccumulatedContext()

		// 2. PLANEJAMENTO
		plan := l.Planner.Plan(ctx, perception, state, l.Contract)

		if plan.Action == "INVALID" {
			l.CircuitBreaker.RecordInvalid()
			log.Printf("[AgentLoop] Resposta inválida detectada (consecutivas: %d)", l.CircuitBreaker.InvalidCount())
		} else {
			l.CircuitBreaker.RecordValid()
		}

		log.Printf("[AgentLoop] Plano: action=%s, tool=%s, done=%v", plan.Action, plan.Tool, plan.Done)

		// 3. EXECUÇÃO
		var toolResult *model.ToolResult
		if !plan.Done && plan.Tool != "" {
			result := l.Executor.Execute(plan)
			toolResult = &result
			state.RecordToolUsed(plan.Tool)

			outcome := result.Output
			if !result.Success {
				outcome = "ERRO: " + result.Error
			}
			state.AddContext(fmt.Sprintf("\n[Step %d] Tool: %s | Resultado: %s", step, plan.Tool, outcome))
		}

		// 4. AVALIAÇÃO
		evaluation := evaluate(plan, toolResult)
		state.AddContext(fmt.Sprintf("\n[Step %d] Avaliação: %s", step, evaluation))

		l.Telemetry.RecordStep(step, perception, plan, toolResult, evaluation)

		// 5. CRITÉRIOS DE PARADA

		if plan.Done {
			log.Printf("[AgentLoop] Agente sinalizou conclusão no step %d", step)
			state.SetDone(true)
			break
		}

		if l.CircuitBreaker.ShouldBreak() {
			log.Printf("[AgentLoop] Circuit breaker aberto após %d respostas inválidas", l.CircuitBreaker.InvalidCount())
			state.AddContext("\n[PARADA] Circuit breaker acionado — LLM retornou respostas inválidas repetidamente.")
			break
		}

		if step >= l.Contract.MaxSteps {
			log.Printf("[AgentLoop] Limite de steps atingido: %d", l.Contract.MaxSteps)
			state.AddContext(fmt.Sprintf("\n[PARADA] Limite de steps atingido: %d", l.Contract.MaxSteps))
			break
		}

		if state.ElapsedSeconds() >= int64(l.Contract.MaxTimeSeconds) {
			log.Printf("[AgentLoop] Timeout atingido: %ds", l.Contract.MaxTimeSeconds)
			state.AddContext(fmt.Sprintf("\n[PARADA] Timeout atingido: %ds", l.Contract.MaxTimeSeconds))
			break
		}

		if state.CheckNoProgress(l.Contract.NoProgressSteps) {
			log.Printf("[AgentLoop] Sem progresso por %d steps consecutivos", l.Contract.NoProgressSteps)
			state.AddContext("\n[PARADA] Agente preso — mesma tool repetida sem progresso.")
			break
		}
	}

	durationMs := time.Since(startTime).Milliseconds()
	return l.Telemetry.Finalize(state, durationMs)
}

// evaluate avalia se o resultado da execução atingiu o critério de sucesso
// do plano — equivalente ao método privado evaluate de AgentLoop.java.
func evaluate(plan model.PlanDecision, toolResult *model.ToolResult) string {
	if plan.Done {
		return "Tarefa concluída pelo agente."
	}
	if toolResult == nil {
		return "Nenhuma tool executada neste step."
	}
	if !toolResult.Success {
		return "FALHA: " + toolResult.Error
	}
	if plan.SuccessCriteria != "" && toolResult.Output != "" {
		return "Tool executada com sucesso. Output disponível para próximo planejamento."
	}
	return "Tool executada. Verificar resultado no próximo step."
}
