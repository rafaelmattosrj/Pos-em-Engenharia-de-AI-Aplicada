package com.iadeva.neo4j.service;

import com.iadeva.neo4j.ResilientChatClient;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Map;

// Gera resposta em linguagem natural a partir dos dados do Neo4j
// Equivalente ao analyticalResponseNode.ts
@Service
public class AnalyticalResponseService {

    private final ResilientChatClient fallbackClient;

    public AnalyticalResponseService(ResilientChatClient fallbackClient) {
        this.fallbackClient = fallbackClient;
    }

    public String generateResponse(String question, List<Map<String, Object>> dbResults) {
        return fallbackClient.call(
                """
                Você é um analista de dados educacionais. Responda perguntas sobre alunos e cursos.
                Seja direto, use os dados fornecidos, responda em português.
                """,
                """
                Pergunta: %s

                Dados do banco: %s

                Responda de forma clara e útil.
                """.formatted(question, dbResults.toString()));
    }
}
