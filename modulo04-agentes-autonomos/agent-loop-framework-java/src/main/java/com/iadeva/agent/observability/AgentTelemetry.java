package com.iadeva.agent.observability;

import com.iadeva.agent.core.AgentState;
import com.iadeva.agent.model.AgentTrace;
import com.iadeva.agent.model.PlanDecision;
import com.iadeva.agent.model.ToolResult;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

/**
 * Coleta telemetria de cada step do agente durante a execução.
 * Equivalente ao observability/trace.py — coleta telemetria de cada step do agente.
 *
 * Mantém o AgentTrace em construção e o finaliza ao término do loop.
 */
@Component
public class AgentTelemetry {

    private static final Logger log = LoggerFactory.getLogger(AgentTelemetry.class);

    // Trace da execução atual (recriado a cada chamada de start())
    private AgentTrace currentTrace;

    /**
     * Inicializa a coleta de telemetria para uma nova execução.
     *
     * @param input Tarefa inicial recebida pelo agente
     */
    public void start(String input) {
        this.currentTrace = new AgentTrace(input);
        log.debug("[Telemetria] Iniciando trace para input: {}", input);
    }

    /**
     * Registra os dados de um step individual no trace.
     *
     * @param stepNumber  Número do step atual
     * @param perception  Contexto acumulado até o momento
     * @param plan        Decisão tomada pelo Planner
     * @param toolResult  Resultado da execução da tool (pode ser null)
     * @param evaluation  Avaliação do resultado do step
     */
    public void recordStep(
            int stepNumber,
            String perception,
            PlanDecision plan,
            ToolResult toolResult,
            String evaluation
    ) {
        if (currentTrace == null) {
            log.warn("[Telemetria] recordStep chamado sem start() — ignorando step {}", stepNumber);
            return;
        }

        AgentTrace.StepTrace step = new AgentTrace.StepTrace(
                stepNumber,
                perception,
                plan,
                toolResult,
                evaluation
        );

        currentTrace.addStep(step);
        log.debug("[Telemetria] Step {} registrado: tool={}, done={}",
                stepNumber,
                plan != null ? plan.tool() : "none",
                plan != null && plan.done());
    }

    /**
     * Finaliza o trace com as métricas de duração e resultado final.
     *
     * @param state      Estado final do agente
     * @param durationMs Duração total da execução em milissegundos
     * @return AgentTrace completo e finalizado
     */
    public AgentTrace finalize(AgentState state, long durationMs) {
        if (currentTrace == null) {
            log.error("[Telemetria] finalize chamado sem trace ativo");
            return new AgentTrace("unknown");
        }

        currentTrace.setTotalSteps(state.getCurrentStep());
        currentTrace.setDurationMs(durationMs);

        // Estimativa simples de tokens: ~4 chars por token no contexto acumulado
        int estimatedTokens = state.getAccumulatedContext().length() / 4;
        currentTrace.setTotalTokensEstimated(estimatedTokens);

        // Resultado final: última avaliação ou indicativo de conclusão
        String finalResult = state.isDone()
                ? "Tarefa concluída após " + state.getCurrentStep() + " steps"
                : "Execução interrompida após " + state.getCurrentStep() + " steps (limite/timeout/circuitbreaker)";
        currentTrace.setFinalResult(finalResult);

        log.info("[Telemetria] Trace finalizado: steps={}, durationMs={}, tokensEstimados={}",
                currentTrace.getTotalSteps(), durationMs, estimatedTokens);

        return currentTrace;
    }

    /** Retorna o trace da execução atual (pode ser null antes de start()) */
    public AgentTrace getTrace() {
        return currentTrace;
    }
}
