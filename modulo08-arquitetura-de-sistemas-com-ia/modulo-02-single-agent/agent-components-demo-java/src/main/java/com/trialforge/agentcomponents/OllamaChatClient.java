package com.trialforge.agentcomponents;

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
import java.util.List;

/**
 * Cliente real do modelo, via API nativa do Ollama local (POST /api/chat, sem
 * streaming) — equivalente a {@code ollama.chat(...)} usado em
 * agent-components-demo.js / agent_components_demo.py. Sem chave de API:
 * Ollama roda 100% local.
 */
public class OllamaChatClient implements ChatClient {

    private final String baseUrl;
    private final String modelo;
    private final HttpClient httpClient;
    private final ObjectMapper mapper = new ObjectMapper();

    public OllamaChatClient(String baseUrl, String modelo) {
        this.baseUrl = baseUrl;
        this.modelo = modelo;
        this.httpClient = HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(10))
                .build();
    }

    @Override
    public String chat(List<Memoria.Mensagem> mensagens) throws IOException, InterruptedException {
        ObjectNode body = mapper.createObjectNode();
        body.put("model", modelo);
        body.put("stream", false);
        ArrayNode messagesNode = body.putArray("messages");
        for (Memoria.Mensagem mensagem : mensagens) {
            ObjectNode m = messagesNode.addObject();
            m.put("role", mensagem.role());
            m.put("content", mensagem.content());
        }

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
                    "Falha ao chamar o Ollama em " + baseUrl + " (rode 'ollama serve' e 'ollama pull "
                            + modelo + "'): " + e.getMessage(), e);
        }

        if (response.statusCode() != 200) {
            throw new IOException("Ollama retornou status " + response.statusCode() + ": " + response.body());
        }

        JsonNode parsed = mapper.readTree(response.body());
        JsonNode content = parsed.path("message").path("content");
        if (content.isMissingNode()) {
            throw new IOException("Ollama retornou resposta sem message.content: " + response.body());
        }
        return content.asText();
    }
}
