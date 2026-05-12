package com.iadeva.cognitive.model;

import java.util.List;
import java.util.Map;

/**
 * Plano de execução gerado pelo PlanExecuteAgent.
 * Contém todos os passos determinados antes de qualquer execução.
 *
 * @param steps lista ordenada de passos a executar
 */
public record ExecutionPlan(List<PlanStep> steps) {

    /**
     * Um passo individual dentro do plano.
     *
     * @param stepNumber      número sequencial (1-based)
     * @param description     descrição em linguagem natural do que será feito
     * @param tool            nome da ferramenta ou ação a invocar
     * @param args            argumentos que serão passados para a ferramenta
     * @param successCriteria critério de aceite para considerar o passo bem-sucedido
     */
    public record PlanStep(
            int stepNumber,
            String description,
            String tool,
            Map<String, Object> args,
            String successCriteria
    ) {}
}
