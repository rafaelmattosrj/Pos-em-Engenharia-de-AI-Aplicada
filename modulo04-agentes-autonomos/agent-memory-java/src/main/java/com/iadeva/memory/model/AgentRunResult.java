package com.iadeva.memory.model;

/**
 * Resultado da execução do agente — inclui métricas de uso de memória para comparação.
 */
public record AgentRunResult(
        String output,
        boolean memoryUsed,
        int factsRetrieved,
        int episodesConsidered
) {}
