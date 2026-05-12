package com.iadeva.cognitive.model;

/**
 * Resposta retornada pelo controller após execução do agente.
 *
 * @param result  texto final produzido pelo agente
 * @param metrics métricas de execução coletadas durante o ciclo
 */
public record AgentResponse(String result, AgentMetrics metrics) {

    /**
     * Métricas de execução do agente.
     *
     * @param steps       número de passos (iterações ReAct ou steps do plano)
     * @param tokens      estimativa de tokens consumidos
     * @param reflections número de ciclos de reflexão (0 para ReAct e Plan-Execute)
     */
    public record AgentMetrics(int steps, int tokens, int reflections) {}
}
