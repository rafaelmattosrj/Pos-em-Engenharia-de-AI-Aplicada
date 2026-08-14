package com.trialforge.reactagent;

import com.fasterxml.jackson.core.type.TypeReference;
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
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

/**
 * Cliente real do modelo, via API nativa do Ollama local (POST /api/chat, sem
 * streaming, com {@code tools}) — equivalente a {@code ollama.chat(...)} usado em
 * react-agent-prototype.js / react_agent_prototype.py. Sem chave de API: Ollama
 * roda 100% local.
 */
public class OllamaReactClient implements ChatClient {

    private final String baseUrl;
    private final String modelo;
    private final HttpClient httpClient;
    private final ObjectMapper mapper = new ObjectMapper();

    public OllamaReactClient(String baseUrl, String modelo) {
        this.baseUrl = baseUrl;
        this.modelo = modelo;
        this.httpClient = HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(10))
                .build();
    }

    @Override
    public ModeloResposta chat(List<ObjectNode> historico, List<ObjectNode> tools) throws IOException, InterruptedException {
        ObjectNode body = mapper.createObjectNode();
        body.put("model", modelo);
        body.put("stream", false);

        ArrayNode messagesNode = body.putArray("messages");
        for (ObjectNode m : historico) {
            messagesNode.add(m);
        }

        if (tools != null && !tools.isEmpty()) {
            ArrayNode toolsNode = body.putArray("tools");
            for (ObjectNode t : tools) {
                toolsNode.add(t);
            }
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
            throw new ModeloHttpException(response.statusCode(), response.body());
        }

        JsonNode parsed = mapper.readTree(response.body());
        JsonNode mensagem = parsed.path("message");
        if (mensagem.isMissingNode() || !mensagem.isObject()) {
            throw new IOException("Ollama retornou resposta sem message: " + response.body());
        }

        ObjectNode mensagemBruta = (ObjectNode) mensagem;
        String content = mensagem.path("content").isMissingNode() ? null : mensagem.path("content").asText();

        List<ChamadaFerramenta> chamadas = new ArrayList<>();
        JsonNode toolCallsNode = mensagem.path("tool_calls");
        if (toolCallsNode.isArray()) {
            for (JsonNode tc : toolCallsNode) {
                JsonNode function = tc.path("function");
                String nome = function.path("name").isMissingNode() ? null : function.path("name").asText();
                JsonNode argsNode = function.path("arguments");
                Map<String, Object> argumentos = argsNode.isMissingNode()
                        ? Map.of()
                        : mapper.convertValue(argsNode, new TypeReference<Map<String, Object>>() {
                        });
                String id = tc.path("id").isMissingNode() ? null : tc.path("id").asText();
                chamadas.add(new ChamadaFerramenta(id, nome, argumentos));
            }
        }

        return new ModeloResposta(mensagemBruta, content, chamadas);
    }
}
