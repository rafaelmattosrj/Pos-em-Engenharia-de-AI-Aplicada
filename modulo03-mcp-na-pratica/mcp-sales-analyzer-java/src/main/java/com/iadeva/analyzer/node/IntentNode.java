package com.iadeva.analyzer.node;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.iadeva.analyzer.model.AnalysisRequest;
import com.iadeva.analyzer.model.IntentResult;
import org.springframework.ai.chat.client.ChatClient;
import org.springframework.stereotype.Component;

import java.util.List;

// Equivalente ao IntentNode do LangGraph — extrai intenção estruturada antes de executar
@Component
public class IntentNode {

    private final ChatClient chatClient;
    private final ObjectMapper objectMapper = new ObjectMapper();

    public IntentNode(ChatClient.Builder chatClientBuilder) {
        this.chatClient = chatClientBuilder.build();
    }

    /**
     * Analisa o input do usuário e extrai intenção estruturada.
     * Determina o tipo de dado (csv/json), normaliza o conteúdo
     * e sugere quais tools devem ser utilizadas pelo ExecutorNode.
     *
     * @param request AnalysisRequest com question e data bruta
     * @return IntentResult com metadados de intenção e dados normalizados
     */
    public IntentResult extract(AnalysisRequest request) {
        // Detecta o tipo de dado heuristicamente antes de enviar ao LLM
        String dataType = detectDataType(request.data());

        // Prompt estruturado para extração de intenção
        String prompt = """
                Analise os dados e a pergunta abaixo. Responda APENAS com JSON válido, sem markdown.
                
                Formato esperado:
                {
                  "dataType": "csv" ou "json",
                  "parsedData": "<dados normalizados como string>",
                  "question": "<pergunta original>",
                  "suggestedTools": ["<tool1>", "<tool2>"]
                }
                
                Tools disponíveis: csvToJson, analyzeData, summarize
                
                Pergunta: %s
                
                Dados:
                %s
                """.formatted(request.question(), request.data());

        String response = chatClient.prompt()
                .user(prompt)
                .call()
                .content();

        return parseIntentResponse(response, request, dataType);
    }

    /**
     * Detecta heuristicamente se os dados são CSV ou JSON.
     * JSON começa com '[' ou '{'; caso contrário assume CSV.
     */
    private String detectDataType(String data) {
        if (data == null || data.isBlank()) {
            return "unknown";
        }
        String trimmed = data.trim();
        if (trimmed.startsWith("[") || trimmed.startsWith("{")) {
            return "json";
        }
        return "csv";
    }

    /**
     * Parseia a resposta JSON do LLM para IntentResult.
     * Em caso de falha no parse, retorna resultado com valores padrão seguros.
     */
    private IntentResult parseIntentResponse(String response, AnalysisRequest request, String fallbackDataType) {
        try {
            // Remove possíveis marcadores de código markdown
            String cleanJson = response.replaceAll("```json", "").replaceAll("```", "").trim();
            var node = objectMapper.readTree(cleanJson);

            String dataType = node.path("dataType").asText(fallbackDataType);
            String parsedData = node.path("parsedData").asText(request.data());
            String question = node.path("question").asText(request.question());

            List<String> suggestedTools = List.of();
            var toolsNode = node.path("suggestedTools");
            if (toolsNode.isArray()) {
                suggestedTools = objectMapper.readerForListOf(String.class).readValue(toolsNode);
            }

            return new IntentResult(dataType, parsedData, question, suggestedTools);

        } catch (Exception e) {
            // Fallback seguro: retorna intenção básica sem tools sugeridas
            return new IntentResult(
                    fallbackDataType,
                    request.data(),
                    request.question(),
                    List.of()
            );
        }
    }
}
