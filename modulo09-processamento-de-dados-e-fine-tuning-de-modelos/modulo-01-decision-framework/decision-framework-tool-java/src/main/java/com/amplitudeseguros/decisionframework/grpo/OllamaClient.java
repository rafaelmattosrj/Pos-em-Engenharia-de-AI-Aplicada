package com.amplitudeseguros.decisionframework.grpo;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;

/** Cliente mínimo para /api/generate do Ollama local -- equivalente a chamarOllama() em JS. */
public final class OllamaClient {

    private static final String OLLAMA_URL = "http://localhost:11434/api/generate";
    private static final ObjectMapper MAPPER = new ObjectMapper();

    private final HttpClient httpClient = HttpClient.newBuilder()
            .connectTimeout(Duration.ofSeconds(5))
            .build();

    public String chamar(String modelo, String prompt, double temperatura) throws IOException, InterruptedException {
        var corpo = MAPPER.createObjectNode();
        corpo.put("model", modelo);
        corpo.put("prompt", prompt);
        corpo.put("stream", false);
        corpo.putObject("options").put("temperature", temperatura);

        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(OLLAMA_URL))
                .timeout(Duration.ofSeconds(60))
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(corpo.toString()))
                .build();

        HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
        if (response.statusCode() != 200) {
            throw new IOException("Ollama respondeu " + response.statusCode());
        }
        JsonNode dados = MAPPER.readTree(response.body());
        return dados.path("response").asText();
    }
}
