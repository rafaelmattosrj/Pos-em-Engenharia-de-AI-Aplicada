package com.iadeva.agent.core;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

/**
 * Contrato de comportamento do agente — define os limites e a identidade do agente.
 * Equivalente ao agent.md/loop.md do curso Python — define o contrato de comportamento do agente.
 *
 * Carregado via application.properties, permitindo ajuste por variável de ambiente.
 */
@Component
public class AgentContract {

    @Value("${agent.max-steps:10}")
    private int maxSteps;

    @Value("${agent.max-tokens:50000}")
    private int maxTokens;

    @Value("${agent.max-time-seconds:120}")
    private int maxTimeSeconds;

    @Value("${agent.no-progress-steps:3}")
    private int noProgressSteps;

    private final String agentName = "DevOps Agent";
    private final String role = "Especialista em diagnóstico e resolução de incidentes";

    // -------------------------------------------------------------------------
    // Getters
    // -------------------------------------------------------------------------

    public int getMaxSteps() { return maxSteps; }
    public int getMaxTokens() { return maxTokens; }
    public int getMaxTimeSeconds() { return maxTimeSeconds; }
    public int getNoProgressSteps() { return noProgressSteps; }
    public String getAgentName() { return agentName; }
    public String getRole() { return role; }

    /**
     * Retorna o system prompt base do agente, derivado do contrato.
     */
    public String getSystemPrompt() {
        return String.format("""
                Você é o %s.
                Papel: %s
                
                Regras:
                - Sempre responda em JSON válido, sem markdown ou blocos de código
                - Seja objetivo e direto nas ações
                - Prefira ferramentas de diagnóstico antes de propor soluções
                - Máximo de %d steps disponíveis — use-os com sabedoria
                """, agentName, role, maxSteps);
    }
}
