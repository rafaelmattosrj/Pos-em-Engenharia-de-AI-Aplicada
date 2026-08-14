package com.iadeva.bragbot.gemini;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestClient;

import java.util.Map;

/**
 * Cliente mínimo para a API REST do Gemini (generateContent) — substitui o
 * plugin @genkit-ai/google-genai usado por flows.ts. Solicita saída em JSON
 * (responseMimeType: application/json) no lugar da validação de schema
 * embutida do Genkit — o parsing do JSON retornado é feito pelo chamador.
 */
@Component
public class GeminiClient {

    private static final String BASE_URL = "https://generativelanguage.googleapis.com/v1beta/models";

    private final RestClient restClient;
    private final ObjectMapper objectMapper;
    private final String apiKey;
    private final String model;

    public GeminiClient(
            @Value("${app.gemini.api-key:}") String apiKey,
            @Value("${app.gemini.model:gemini-2.5-flash}") String model,
            ObjectMapper objectMapper) {
        this.apiKey = apiKey;
        this.model = model;
        this.objectMapper = objectMapper;
        this.restClient = RestClient.create();
    }

    /**
     * Gera conteúdo em JSON para o prompt informado, na temperatura dada.
     *
     * @return o texto JSON retornado pelo modelo (ainda não parseado)
     */
    public String generateJson(String prompt, double temperature) {
        Map<String, Object> body = Map.of(
                "contents", new Object[]{
                        Map.of("parts", new Object[]{Map.of("text", prompt)})
                },
                "generationConfig", Map.of(
                        "temperature", temperature,
                        "responseMimeType", "application/json"
                )
        );

        String url = BASE_URL + "/" + model + ":generateContent?key=" + apiKey;

        String rawResponse = restClient.post()
                .uri(url)
                .contentType(org.springframework.http.MediaType.APPLICATION_JSON)
                .body(body)
                .retrieve()
                .body(String.class);

        return extractText(rawResponse);
    }

    private String extractText(String rawResponse) {
        try {
            JsonNode root = objectMapper.readTree(rawResponse);
            return root.path("candidates").path(0)
                    .path("content").path("parts").path(0)
                    .path("text").asText();
        } catch (Exception e) {
            throw new RuntimeException("Falha ao interpretar resposta do Gemini: " + e.getMessage(), e);
        }
    }
}
