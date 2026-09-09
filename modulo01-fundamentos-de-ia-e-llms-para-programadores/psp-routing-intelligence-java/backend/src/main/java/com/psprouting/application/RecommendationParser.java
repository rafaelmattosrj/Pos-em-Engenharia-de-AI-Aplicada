package com.psprouting.application;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.psprouting.domain.PSP;
import org.springframework.stereotype.Component;

import java.util.ArrayList;
import java.util.List;

/**
 * Decodifica a resposta em texto do LLM no {@link ParsedRecommendation}
 * esperado pelo prompt RAG (ver {@code routing-prompt.txt}):
 *
 * <pre>{@code
 * {
 *   "primary": "NOME_DO_PSP",
 *   "confidence": 0.0,
 *   "reasoning": "...",
 *   "fallback": ["PSP2", "PSP3"]
 * }
 * }</pre>
 *
 * O prompt pede JSON puro ("sem markdown, sem texto adicional"), mas LLMs
 * gratuitos as vezes envolvem a resposta em blocos {@code ```json ... ```}
 * — por isso a extracao remove esses marcadores antes de desserializar, em
 * vez de assumir que o modelo sempre obedece a instrucao a risca.
 */
@Component
public class RecommendationParser {

    private final ObjectMapper objectMapper;

    public RecommendationParser(ObjectMapper objectMapper) {
        this.objectMapper = objectMapper;
    }

    public ParsedRecommendation parse(String rawResponse) {
        String json = stripMarkdownFences(rawResponse);

        JsonNode node;
        try {
            node = objectMapper.readTree(json);
        } catch (Exception e) {
            throw new InvalidLlmResponseException(
                    "Resposta do LLM nao e um JSON valido: " + truncate(rawResponse), e);
        }

        if (!node.hasNonNull("primary") || !node.hasNonNull("confidence")) {
            throw new InvalidLlmResponseException(
                    "Resposta do LLM nao contem os campos obrigatorios (primary/confidence): "
                            + truncate(rawResponse));
        }

        PSP primary = parsePsp(node.get("primary").asText(), rawResponse);
        double confidence = node.get("confidence").asDouble();
        String reasoning = node.hasNonNull("reasoning") ? node.get("reasoning").asText() : "";
        List<PSP> fallback = parseFallback(node.get("fallback"), rawResponse);

        return new ParsedRecommendation(primary, confidence, reasoning, fallback);
    }

    private List<PSP> parseFallback(JsonNode fallbackNode, String rawResponse) {
        List<PSP> fallback = new ArrayList<>();
        if (fallbackNode == null || !fallbackNode.isArray()) {
            return fallback;
        }
        for (JsonNode item : fallbackNode) {
            fallback.add(parsePsp(item.asText(), rawResponse));
        }
        return fallback;
    }

    private PSP parsePsp(String value, String rawResponse) {
        try {
            return PSP.valueOf(value.trim().toUpperCase());
        } catch (IllegalArgumentException e) {
            throw new InvalidLlmResponseException(
                    "PSP desconhecido na resposta do LLM: '" + value + "' — " + truncate(rawResponse), e);
        }
    }

    private static String stripMarkdownFences(String raw) {
        String trimmed = raw.strip();
        if (trimmed.startsWith("```")) {
            int firstNewline = trimmed.indexOf('\n');
            if (firstNewline != -1) {
                trimmed = trimmed.substring(firstNewline + 1);
            }
            int lastFence = trimmed.lastIndexOf("```");
            if (lastFence != -1) {
                trimmed = trimmed.substring(0, lastFence);
            }
        }
        return trimmed.strip();
    }

    private static String truncate(String message) {
        if (message == null) {
            return "";
        }
        return message.length() > 200 ? message.substring(0, 200) : message;
    }
}
