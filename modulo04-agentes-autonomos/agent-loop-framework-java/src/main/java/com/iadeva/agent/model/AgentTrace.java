package com.iadeva.agent.model;

import java.util.ArrayList;
import java.util.List;

/**
 * Trace completo de uma execução do agente, incluindo todos os steps.
 * Equivalente ao AgentTrace do observability/trace.py no curso Python.
 */
public class AgentTrace {

    private String input;
    private List<StepTrace> steps = new ArrayList<>();
    private String finalResult;
    private int totalSteps;
    private int totalTokensEstimated;
    private long durationMs;

    public AgentTrace() {}

    public AgentTrace(String input) {
        this.input = input;
    }

    // -------------------------------------------------------------------------
    // Inner record: representa um único step do loop
    // -------------------------------------------------------------------------

    /**
     * Trace de um step individual do agent loop.
     *
     * @param stepNumber  Número sequencial do step (começa em 1)
     * @param perception  Contexto acumulado que o agente observou
     * @param plan        Decisão de planejamento do LLM
     * @param toolResult  Resultado da execução da tool
     * @param evaluation  Avaliação do resultado (sucesso/falha/parcial)
     */
    public record StepTrace(
            int stepNumber,
            String perception,
            PlanDecision plan,
            ToolResult toolResult,
            String evaluation
    ) {}

    // -------------------------------------------------------------------------
    // Getters e setters
    // -------------------------------------------------------------------------

    public String getInput() { return input; }
    public void setInput(String input) { this.input = input; }

    public List<StepTrace> getSteps() { return steps; }
    public void setSteps(List<StepTrace> steps) { this.steps = steps; }

    public String getFinalResult() { return finalResult; }
    public void setFinalResult(String finalResult) { this.finalResult = finalResult; }

    public int getTotalSteps() { return totalSteps; }
    public void setTotalSteps(int totalSteps) { this.totalSteps = totalSteps; }

    public int getTotalTokensEstimated() { return totalTokensEstimated; }
    public void setTotalTokensEstimated(int totalTokensEstimated) { this.totalTokensEstimated = totalTokensEstimated; }

    public long getDurationMs() { return durationMs; }
    public void setDurationMs(long durationMs) { this.durationMs = durationMs; }

    /** Adiciona um step ao trace */
    public void addStep(StepTrace step) {
        this.steps.add(step);
    }
}
