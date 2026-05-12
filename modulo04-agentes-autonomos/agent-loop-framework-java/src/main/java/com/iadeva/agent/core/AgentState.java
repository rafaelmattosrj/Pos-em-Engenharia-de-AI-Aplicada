package com.iadeva.agent.core;

import java.time.Instant;
import java.util.ArrayList;
import java.util.List;

/**
 * Estado mutável do agente durante uma execução do loop.
 * Mantém contexto acumulado, histórico de tools e contadores de progresso.
 */
public class AgentState {

    private int currentStep = 0;
    private final StringBuilder accumulatedContext = new StringBuilder();
    private final List<String> lastToolsUsed = new ArrayList<>();
    private int consecutiveNoProgress = 0;
    private final Instant startTime = Instant.now();
    private boolean done = false;

    // -------------------------------------------------------------------------
    // Mutações de estado
    // -------------------------------------------------------------------------

    /** Incrementa o step atual e retorna o novo valor */
    public int incrementStep() {
        return ++currentStep;
    }

    /** Adiciona texto ao contexto acumulado */
    public void addContext(String text) {
        accumulatedContext.append("\n").append(text);
    }

    /**
     * Registra a tool utilizada no step atual.
     * Detecta automaticamente se não houve progresso (mesma tool repetida).
     */
    public void recordToolUsed(String toolName) {
        if (!lastToolsUsed.isEmpty() && lastToolsUsed.getLast().equals(toolName)) {
            consecutiveNoProgress++;
        } else {
            consecutiveNoProgress = 0;
        }
        lastToolsUsed.add(toolName);
    }

    /**
     * Verifica se o agente está preso em loop sem progresso.
     *
     * @param noProgressLimit limite configurado em AgentContract
     */
    public boolean checkNoProgress(int noProgressLimit) {
        return consecutiveNoProgress >= noProgressLimit;
    }

    /** Retorna o tempo decorrido desde o início da execução em segundos */
    public long elapsedSeconds() {
        return java.time.Duration.between(startTime, Instant.now()).getSeconds();
    }

    // -------------------------------------------------------------------------
    // Getters e setters
    // -------------------------------------------------------------------------

    public int getCurrentStep() { return currentStep; }
    public String getAccumulatedContext() { return accumulatedContext.toString(); }
    public List<String> getLastToolsUsed() { return lastToolsUsed; }
    public int getConsecutiveNoProgress() { return consecutiveNoProgress; }
    public Instant getStartTime() { return startTime; }
    public boolean isDone() { return done; }
    public void setDone(boolean done) { this.done = done; }
}
