package com.iadeva.agent.core;

import com.iadeva.agent.model.AgentTrace;
import com.iadeva.agent.model.PlanDecision;
import com.iadeva.agent.model.ToolResult;
import com.iadeva.agent.observability.AgentTelemetry;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import java.time.Instant;

/**
 * Agent Loop — orquestra o ciclo completo Percepção→Planejamento→Ação→Avaliação.
 * Equivalente ao agent_loop.py das aulas 03-06 — orquestra o ciclo completo Percepção→Planejamento→Ação→Avaliação.
 *
 * Critérios de parada implementados:
 * 1. done=true na PlanDecision (tarefa concluída pelo agente)
 * 2. maxSteps atingido (limite configurado no AgentContract)
 * 3. maxTime atingido (timeout em segundos)
 * 4. noProgress detectado (mesma tool repetida N vezes consecutivas)
 * 5. CircuitBreaker aberto (3 respostas inválidas do LLM)
 */
@Component
public class AgentLoop {

    private static final Logger log = LoggerFactory.getLogger(AgentLoop.class);

    private final Planner planner;
    private final Executor executor;
    private final CircuitBreaker circuitBreaker;
    private final AgentTelemetry telemetry;
    private final AgentContract contract;

    public AgentLoop(
            Planner planner,
            Executor executor,
            CircuitBreaker circuitBreaker,
            AgentTelemetry telemetry,
            AgentContract contract
    ) {
        this.planner = planner;
        this.executor = executor;
        this.circuitBreaker = circuitBreaker;
        this.telemetry = telemetry;
        this.contract = contract;
    }

    /**
     * Executa o agente loop completo para o input fornecido.
     *
     * @param input Tarefa ou pergunta inicial do usuário
     * @return AgentTrace com o histórico completo da execução
     */
    public AgentTrace run(String input) {
        log.info("[AgentLoop] Iniciando execução para: {}", input);

        // Inicializa estado limpo para esta execução
        AgentState state = new AgentState();
        circuitBreaker.reset();
        telemetry.start(input);

        long startMs = System.currentTimeMillis();
        state.addContext("Tarefa recebida: " + input);

        while (true) {
            int step = state.incrementStep();
            log.info("[AgentLoop] === Step {} ===", step);

            // ------------------------------------------------------------------
            // 1. PERCEPÇÃO — o que o agente sabe até agora
            // ------------------------------------------------------------------
            String perception = state.getAccumulatedContext();

            // ------------------------------------------------------------------
            // 2. PLANEJAMENTO — consulta o LLM
            // ------------------------------------------------------------------
            PlanDecision plan = planner.plan(perception, state, contract);

            // Verifica qualidade da resposta do LLM
            if ("INVALID".equals(plan.action())) {
                circuitBreaker.recordInvalid();
                log.warn("[AgentLoop] Resposta inválida detectada (consecutivas: {})", circuitBreaker.getInvalidCount());
            } else {
                circuitBreaker.recordValid();
            }

            log.info("[AgentLoop] Plano: action={}, tool={}, done={}", plan.action(), plan.tool(), plan.done());

            // ------------------------------------------------------------------
            // 3. EXECUÇÃO — chama a tool
            // ------------------------------------------------------------------
            ToolResult toolResult = null;
            if (!plan.done() && plan.tool() != null) {
                toolResult = executor.execute(plan);
                state.recordToolUsed(plan.tool());
                state.addContext(String.format("\n[Step %d] Tool: %s | Resultado: %s",
                        step, plan.tool(),
                        toolResult.success() ? toolResult.output() : "ERRO: " + toolResult.error()));
            }

            // ------------------------------------------------------------------
            // 4. AVALIAÇÃO — determina se o critério de sucesso foi atingido
            // ------------------------------------------------------------------
            String evaluation = evaluate(plan, toolResult);
            state.addContext("\n[Step " + step + "] Avaliação: " + evaluation);

            // Registra step na telemetria
            telemetry.recordStep(step, perception, plan, toolResult, evaluation);

            // ------------------------------------------------------------------
            // 5. CRITÉRIOS DE PARADA
            // ------------------------------------------------------------------

            // 5a. Tarefa concluída
            if (plan.done()) {
                log.info("[AgentLoop] Agente sinalizou conclusão no step {}", step);
                state.setDone(true);
                break;
            }

            // 5b. Circuit breaker aberto
            if (circuitBreaker.shouldBreak()) {
                log.warn("[AgentLoop] Circuit breaker aberto após {} respostas inválidas", circuitBreaker.getInvalidCount());
                state.addContext("\n[PARADA] Circuit breaker acionado — LLM retornou respostas inválidas repetidamente.");
                break;
            }

            // 5c. Máximo de steps atingido
            if (step >= contract.getMaxSteps()) {
                log.warn("[AgentLoop] Limite de steps atingido: {}", contract.getMaxSteps());
                state.addContext("\n[PARADA] Limite de steps atingido: " + contract.getMaxSteps());
                break;
            }

            // 5d. Timeout
            if (state.elapsedSeconds() >= contract.getMaxTimeSeconds()) {
                log.warn("[AgentLoop] Timeout atingido: {}s", contract.getMaxTimeSeconds());
                state.addContext("\n[PARADA] Timeout atingido: " + contract.getMaxTimeSeconds() + "s");
                break;
            }

            // 5e. Sem progresso (mesma tool repetida)
            if (state.checkNoProgress(contract.getNoProgressSteps())) {
                log.warn("[AgentLoop] Sem progresso por {} steps consecutivos", contract.getNoProgressSteps());
                state.addContext("\n[PARADA] Agente preso — mesma tool repetida sem progresso.");
                break;
            }
        }

        long durationMs = System.currentTimeMillis() - startMs;
        return telemetry.finalize(state, durationMs);
    }

    // -------------------------------------------------------------------------
    // Métodos privados
    // -------------------------------------------------------------------------

    /**
     * Avalia se o resultado da execução atingiu o critério de sucesso do plano.
     */
    private String evaluate(PlanDecision plan, ToolResult toolResult) {
        if (plan.done()) {
            return "Tarefa concluída pelo agente.";
        }
        if (toolResult == null) {
            return "Nenhuma tool executada neste step.";
        }
        if (!toolResult.success()) {
            return "FALHA: " + toolResult.error();
        }

        // Avaliação simples baseada em presença do critério no output
        String criteria = plan.successCriteria();
        if (criteria != null && !criteria.isBlank() && toolResult.output() != null) {
            return "Tool executada com sucesso. Output disponível para próximo planejamento.";
        }

        return "Tool executada. Verificar resultado no próximo step.";
    }
}
