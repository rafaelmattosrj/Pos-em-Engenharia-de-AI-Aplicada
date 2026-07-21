package com.iadeva.neo4j.graph;

import com.iadeva.neo4j.service.AnalyticalResponseService;
import com.iadeva.neo4j.service.CypherGeneratorService;
import com.iadeva.neo4j.service.Neo4jService;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;

import java.util.Collections;
import java.util.List;
import java.util.Map;

// Orquestrador RAG — equivalente ao StateGraph do LangGraph com self-correction
// Fluxo: extractQuestion → queryPlanner → cypherGenerator → cypherExecutor
//        → (cypherCorrection se falhou) → analyticalResponse → END
@Component
public class RagOrchestrator {

    private final Neo4jService neo4jService;
    private final CypherGeneratorService cypherGenerator;
    private final AnalyticalResponseService analyticalResponse;

    @Value("${app.max-correction-attempts:1}")
    private int maxCorrectionAttempts;

    public RagOrchestrator(
            Neo4jService neo4jService,
            CypherGeneratorService cypherGenerator,
            AnalyticalResponseService analyticalResponse) {
        this.neo4jService = neo4jService;
        this.cypherGenerator = cypherGenerator;
        this.analyticalResponse = analyticalResponse;
    }

    public String query(String question) {
        System.out.printf("Query: %s%n", question);

        // Nó: obter schema
        String schema = neo4jService.getSchema();

        // Nó: gerar Cypher
        String cypher = cypherGenerator.generateCypher(question, schema);
        System.out.printf("Cypher gerado: %s%n", cypher);

        // Nó: executar com autocorreção
        List<Map<String, Object>> results = executeWithCorrection(cypher, question, schema, 0);

        // Nó: gerar resposta analítica
        String answer = analyticalResponse.generateResponse(question, results);
        System.out.printf("Resposta gerada%n");

        return answer;
    }

    // Aresta condicional com self-correction — equivalente ao cypherExecutor + cypherCorrection
    private List<Map<String, Object>> executeWithCorrection(
            String cypher, String question, String schema, int attempt) {
        try {
            List<Map<String, Object>> results = neo4jService.query(cypher);
            System.out.printf("Query executada: %d resultados%n", results.size());
            return results;
        } catch (Exception e) {
            System.out.printf("Erro na query (tentativa %d): %s%n", attempt + 1, e.getMessage());

            if (attempt < maxCorrectionAttempts) {
                // Autocorreção: pede ao LLM para corrigir a query
                String correctedCypher = cypherGenerator.correctCypher(cypher, e.getMessage(), schema);
                System.out.printf("Query corrigida: %s%n", correctedCypher);
                return executeWithCorrection(correctedCypher, question, schema, attempt + 1);
            }

            System.out.println("Retornando resultado vazio após falhas");
            return Collections.emptyList();
        }
    }
}
