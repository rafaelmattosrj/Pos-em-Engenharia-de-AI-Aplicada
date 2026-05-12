package com.iadeva.agent.core;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.iadeva.agent.model.PlanDecision;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.stereotype.Component;

import java.util.Map;

/**
 * Planner — consulta o LLM e retorna uma PlanDecision estruturada.
 * Equivalente ao planner.py — consulta LLM e retorna PlanDecision estruturado.
 *
 * Recebe a percepção atual do agente e produz a próxima ação a executar.
 */
@Component
public class Planner {

    private static final Logger log = LoggerFactory.getLogger(Planner.class);

    private final ChatClient chatClient;
    private final ObjectMapper objectMapper;

    public Planner(ChatClient.Builder chatClientBuilder, ObjectMapper objectMapper) {
        this.chatClient = chatClientBuilder.build();
        this.objectMapper = objectMapper;
    }

    /**
     * Gera a próxima decisão de planejamento com base na percepção atual.
     *
     * @param perception Contexto acumulado pelo agente até o momento
     * @param state      Estado atual do agente
     * @param contract   Contrato com as regras e limites do agente
     * @return PlanDecision com a próxima ação a executar
     */
    public PlanDecision plan(String perception, AgentState state, AgentContract contract) {
        String prompt = buildPrompt(perception, state, contract);

        log.debug("Consultando LLM para planejamento no step {}", state.getCurrentStep());

        try {
            String response = chatClient
                    .prompt()
                    .system(contract.getSystemPrompt())
                    .user(prompt)
                    .call()
                    .content();

            return parseResponse(response);
        } catch (Exception e) {
            log.error("Erro ao consultar LLM: {}", e.getMessage());
            // Retorna decisão de finalização em caso de erro irrecuperável
            return new PlanDecision(
                    "Erro ao consultar LLM: " + e.getMessage(),
                    "Encerrar execução por erro",
                    null,
                    Map.of(),
                    "N/A",
                    true
            );
        }
    }

    // -------------------------------------------------------------------------
    // Métodos privados
    // -------------------------------------------------------------------------

    private String buildPrompt(String perception, AgentState state, AgentContract contract) {
        return String.format("""
                ## Contexto atual (Percepção)
                %s
                
                ## Situação
                - Step atual: %d de %d
                - Tools já usadas: %s
                - Tempo decorrido: %d segundos
                
                ## Tools disponíveis
                - getMetrics: Obtém métricas de latência (p50, p95, p99) de um serviço. Args: { "service": "nome-do-servico" }
                - getLogs: Obtém últimas linhas de log de um serviço. Args: { "service": "nome-do-servico", "lines": 20 }
                - getDeployHistory: Obtém histórico de deploys recentes. Args: { "service": "nome-do-servico" }
                - saveIncident: Salva um incidente. Args: { "title": "...", "severity": "high|medium|low", "description": "..." }
                
                ## Instrução
                Analise o contexto e decida a próxima ação. Responda APENAS com JSON válido no formato abaixo, sem markdown:
                
                {
                  "reasoning": "Seu raciocínio sobre a situação",
                  "action": "Descrição clara da ação",
                  "tool": "nomeDaTool ou null se tarefa concluída",
                  "args": { "chave": "valor" },
                  "successCriteria": "Como saber que o resultado foi satisfatório",
                  "done": false
                }
                
                Se a tarefa estiver concluída, use "done": true e "tool": null.
                """,
                perception.isBlank() ? "Nenhum contexto ainda — início da execução." : perception,
                state.getCurrentStep(),
                contract.getMaxSteps(),
                state.getLastToolsUsed().isEmpty() ? "nenhuma" : String.join(", ", state.getLastToolsUsed()),
                state.elapsedSeconds()
        );
    }

    @SuppressWarnings("unchecked")
    private PlanDecision parseResponse(String response) {
        try {
            // Remove possíveis blocos de markdown que o LLM possa ter incluído
            String cleaned = response.trim()
                    .replaceAll("```json", "")
                    .replaceAll("```", "")
                    .trim();

            Map<String, Object> json = objectMapper.readValue(cleaned, Map.class);

            return new PlanDecision(
                    getString(json, "reasoning"),
                    getString(json, "action"),
                    getString(json, "tool"),
                    json.containsKey("args") ? (Map<String, Object>) json.get("args") : Map.of(),
                    getString(json, "successCriteria"),
                    Boolean.TRUE.equals(json.get("done"))
            );
        } catch (Exception e) {
            log.warn("Falha ao parsear resposta do LLM: {}. Resposta: {}", e.getMessage(), response);
            // Retorna decisão inválida que será detectada pelo CircuitBreaker
            return new PlanDecision(
                    "Resposta inválida do LLM",
                    "INVALID",
                    null,
                    Map.of(),
                    "N/A",
                    false
            );
        }
    }

    private String getString(Map<String, Object> map, String key) {
        Object value = map.get(key);
        return value != null ? value.toString() : null;
    }
}
