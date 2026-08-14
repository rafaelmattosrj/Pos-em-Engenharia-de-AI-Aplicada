package com.trialforge.guardrail;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ArrayNode;
import com.fasterxml.jackson.databind.node.ObjectNode;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;

/**
 * Cliente real do classificador, via API nativa do Ollama local (POST /api/chat,
 * sem streaming — equivalente a {@code ollama.chat(...)} usado em
 * manipulation-guardrail-prototype.js / .py). Sem chave de API: Ollama roda
 * 100% local.
 */
public class OllamaClassifierClient implements ClassifierClient {

    private final String baseUrl;
    private final String modelo;
    private final HttpClient httpClient;
    private final ObjectMapper mapper = new ObjectMapper();

    public OllamaClassifierClient(String baseUrl, String modelo) {
        this.baseUrl = baseUrl;
        this.modelo = modelo;
        this.httpClient = HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(10))
                .build();
    }

    @Override
    public String classify(String pergunta) throws IOException, InterruptedException {
        ObjectNode body = mapper.createObjectNode();
        body.put("model", modelo);
        body.put("stream", false);

        ArrayNode messages = body.putArray("messages");
        ObjectNode systemMsg = messages.addObject();
        systemMsg.put("role", "system");
        systemMsg.put("content", GuardrailGateway.INSTRUCAO_CLASSIFICADOR);
        ObjectNode userMsg = messages.addObject();
        userMsg.put("role", "user");
        userMsg.put("content", pergunta);

        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(baseUrl + "/api/chat"))
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(mapper.writeValueAsString(body)))
                .build();

        HttpResponse<String> response;
        try {
            response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
        } catch (IOException e) {
            throw new IOException(
                    "Falha ao chamar o Ollama em " + baseUrl + " (rode 'ollama serve' e 'ollama pull " + modelo + "'): "
                            + e.getMessage(),
                    e);
        }

        if (response.statusCode() != 200) {
            throw new IOException("Ollama retornou status " + response.statusCode() + " ao classificar: " + response.body());
        }

        JsonNode parsed = mapper.readTree(response.body());
        JsonNode content = parsed.path("message").path("content");
        if (content.isMissingNode() || content.asText().isBlank()) {
            throw new IOException("Ollama retornou resposta sem message.content: " + response.body());
        }
        return content.asText().trim();
    }
}
