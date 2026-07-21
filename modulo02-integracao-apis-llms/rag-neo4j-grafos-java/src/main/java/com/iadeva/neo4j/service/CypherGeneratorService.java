package com.iadeva.neo4j.service;

import com.iadeva.neo4j.ResilientChatClient;
import org.springframework.stereotype.Service;

// Gera queries Cypher a partir de linguagem natural — equivalente ao cypherGeneratorNode.ts
@Service
public class CypherGeneratorService {

    private final ResilientChatClient fallbackClient;

    public CypherGeneratorService(ResilientChatClient fallbackClient) {
        this.fallbackClient = fallbackClient;
    }

    public String generateCypher(String question, String schema) {
        return fallbackClient.call(
                """
                Você é um especialista em Neo4j Cypher. Converta perguntas em linguagem natural para queries Cypher.

                Schema do banco:
                %s

                Regras:
                - Retorne APENAS a query Cypher, sem explicações, sem markdown
                - Use MATCH, WHERE, RETURN conforme necessário
                - Para relações de compra, use: (s:Student)-[:ENROLLED_IN]->(c:Course)
                - Prefira RETURN com propriedades específicas em vez de nós inteiros
                """.formatted(schema),
                question)
                .trim()
                .replaceAll("```cypher", "")
                .replaceAll("```", "")
                .trim();
    }

    public String correctCypher(String failedQuery, String error, String schema) {
        return fallbackClient.call(
                """
                Corrija a query Cypher que falhou. Retorne APENAS a query corrigida.

                Schema: %s
                """.formatted(schema),
                "Query com erro:\n%s\n\nErro: %s\n\nRetorne a query corrigida:".formatted(failedQuery, error))
                .trim()
                .replaceAll("```cypher", "")
                .replaceAll("```", "")
                .trim();
    }
}
