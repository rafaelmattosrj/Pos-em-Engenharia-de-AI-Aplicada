package com.iadeva.agent.tool;

import java.util.Map;

/**
 * Contrato que todas as tools do agente devem implementar.
 * Permite que o Executor despache chamadas de forma genérica pelo nome da tool.
 */
public interface AgentTool {

    /** Identificador único da tool — usado pelo Planner para referenciar a ação */
    String getName();

    /**
     * Executa a tool com os argumentos fornecidos.
     *
     * @param args Mapa de argumentos passados pelo Planner
     * @return Resultado da execução como String (JSON ou texto)
     */
    String execute(Map<String, Object> args);
}
