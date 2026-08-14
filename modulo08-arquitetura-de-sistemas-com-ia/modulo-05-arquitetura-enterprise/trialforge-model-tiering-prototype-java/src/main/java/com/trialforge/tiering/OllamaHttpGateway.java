package com.trialforge.tiering;

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
import java.util.function.Consumer;
import java.util.stream.Stream;

/**
 * Implementação real de {@link OllamaGateway}, via API nativa do Ollama local:
 * POST /api/chat com stream=true (NDJSON, uma linha JSON por pedaço — mesmo
 * mecanismo de {@code for await (const parte of stream)} em JS /
 * {@code for parte in stream} em Python) e POST /api/embeddings.
 */
public class OllamaHttpGateway implements OllamaGateway {

    private final String baseUrl;
    private final HttpClient httpClient;
    private final ObjectMapper mapper = new ObjectMapper();

    public OllamaHttpGateway(String baseUrl) {
        this.baseUrl = baseUrl;
        this.httpClient = HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(10))
                .build();
    }

    @Override
    public String chatStream(String modelo, String systemPrompt, String userPrompt, Consumer<String> onChunk)
            throws IOException, InterruptedException {
        ObjectNode body = mapper.createObjectNode();
        body.put("model", modelo);
        body.put("stream", true);
        ArrayNode messages = body.putArray("messages");
        ObjectNode systemMsg = messages.addObject();
        systemMsg.put("role", "system");
        systemMsg.put("content", systemPrompt);
        ObjectNode userMsg = messages.addObject();
        userMsg.put("role", "user");
        userMsg.put("content", userPrompt);

        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(baseUrl + "/api/chat"))
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(mapper.writeValueAsString(body)))
                .build();

        HttpResponse<Stream<String>> response;
        try {
            response = httpClient.send(request, HttpResponse.BodyHandlers.ofLines());
        } catch (IOException e) {
            throw new IOException("Falha ao chamar o Ollama em " + baseUrl
                    + "/api/chat (rode 'ollama serve' e 'ollama pull " + modelo + "'): " + e.getMessage(), e);
        }

        if (response.statusCode() != 200) {
            throw new IOException("Ollama retornou status " + response.statusCode() + " ao gerar (stream).");
        }

        StringBuilder rascunho = new StringBuilder();
        try (Stream<String> linhas = response.body()) {
            for (String linha : (Iterable<String>) linhas::iterator) {
                if (linha.isBlank()) continue;
                JsonNode parsed = mapper.readTree(linha);
                String pedaco = parsed.path("message").path("content").asText("");
                if (!pedaco.isEmpty()) {
                    onChunk.accept(pedaco);
                    rascunho.append(pedaco);
                }
            }
        }
        return rascunho.toString();
    }

    @Override
    public double[] embed(String modelo, String texto) throws IOException, InterruptedException {
        ObjectNode body = mapper.createObjectNode();
        body.put("model", modelo);
        body.put("prompt", texto);

        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(baseUrl + "/api/embeddings"))
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(mapper.writeValueAsString(body)))
                .build();

        HttpResponse<String> response;
        try {
            response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
        } catch (IOException e) {
            throw new IOException("Falha ao chamar o Ollama em " + baseUrl + "/api/embeddings: " + e.getMessage(), e);
        }

        if (response.statusCode() != 200) {
            throw new IOException("Ollama retornou status " + response.statusCode() + " ao gerar embedding: " + response.body());
        }

        JsonNode parsed = mapper.readTree(response.body());
        JsonNode embeddingNode = parsed.path("embedding");
        if (!embeddingNode.isArray() || embeddingNode.isEmpty()) {
            throw new IOException("Ollama retornou embedding vazio: " + response.body());
        }
        double[] vetor = new double[embeddingNode.size()];
        for (int i = 0; i < vetor.length; i++) {
            vetor[i] = embeddingNode.get(i).asDouble();
        }
        return vetor;
    }
}
