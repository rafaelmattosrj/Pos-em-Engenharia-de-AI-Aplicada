package com.trialforge.evalgate;

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
 * Implementação real de {@link OllamaGateway}, via API nativa do Ollama local
 * (POST /api/chat sem streaming, POST /api/embeddings) — equivalente a
 * {@code ollama.chat(...)} / {@code ollama.embeddings(...)} usados nos
 * originais em JS/Python. Sem chave de API: Ollama roda 100% local.
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
    public String chat(String modelo, String systemPrompt, String userPrompt) throws IOException, InterruptedException {
        ObjectNode body = mapper.createObjectNode();
        body.put("model", modelo);
        body.put("stream", false);
        ArrayNode messages = body.putArray("messages");
        ObjectNode systemMsg = messages.addObject();
        systemMsg.put("role", "system");
        systemMsg.put("content", systemPrompt);
        ObjectNode userMsg = messages.addObject();
        userMsg.put("role", "user");
        userMsg.put("content", userPrompt);

        String responseBody = post("/api/chat", body);
        JsonNode parsed = mapper.readTree(responseBody);
        JsonNode content = parsed.path("message").path("content");
        if (content.isMissingNode()) {
            throw new IOException("Ollama retornou resposta sem message.content: " + responseBody);
        }
        return content.asText();
    }

    @Override
    public double[] embed(String modelo, String texto) throws IOException, InterruptedException {
        ObjectNode body = mapper.createObjectNode();
        body.put("model", modelo);
        body.put("prompt", texto);

        String responseBody = post("/api/embeddings", body);
        JsonNode parsed = mapper.readTree(responseBody);
        JsonNode embeddingNode = parsed.path("embedding");
        if (!embeddingNode.isArray() || embeddingNode.isEmpty()) {
            throw new IOException("Ollama retornou embedding vazio: " + responseBody);
        }
        double[] vetor = new double[embeddingNode.size()];
        for (int i = 0; i < vetor.length; i++) {
            vetor[i] = embeddingNode.get(i).asDouble();
        }
        return vetor;
    }

    private String post(String path, ObjectNode body) throws IOException, InterruptedException {
        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(baseUrl + path))
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(mapper.writeValueAsString(body)))
                .build();

        HttpResponse<String> response;
        try {
            response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
        } catch (IOException e) {
            throw new IOException("Falha ao chamar o Ollama em " + baseUrl + path
                    + " (rode 'ollama serve' e verifique se os modelos foram puxados): " + e.getMessage(), e);
        }

        if (response.statusCode() != 200) {
            throw new IOException("Ollama retornou status " + response.statusCode() + " em " + path + ": " + response.body());
        }
        return response.body();
    }
}
