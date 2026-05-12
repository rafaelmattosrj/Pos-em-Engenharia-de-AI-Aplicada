package com.iadeva.cognitive.model;

/**
 * Requisição enviada ao controller para executar um agente.
 *
 * @param architecture tipo da arquitetura: "react", "plan-execute" ou "reflection"
 * @param input        tarefa ou pergunta que o agente deve resolver
 */
public record AgentRequest(String architecture, String input) {}
