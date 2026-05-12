package com.iadeva.agent.model;

import java.util.Map;

/**
 * Decisão de planejamento retornada pelo LLM a cada step do loop.
 * Equivalente ao PlanDecision do planner.py no curso Python.
 *
 * @param reasoning       Raciocínio do agente sobre a situação atual
 * @param action          Descrição da ação a ser executada
 * @param tool            Nome da tool a ser chamada (pode ser null se done=true)
 * @param args            Argumentos para a tool
 * @param successCriteria Critério de sucesso para avaliar o resultado
 * @param done            Indica se o agente concluiu a tarefa
 */
public record PlanDecision(
        String reasoning,
        String action,
        String tool,
        Map<String, Object> args,
        String successCriteria,
        boolean done
) {}
