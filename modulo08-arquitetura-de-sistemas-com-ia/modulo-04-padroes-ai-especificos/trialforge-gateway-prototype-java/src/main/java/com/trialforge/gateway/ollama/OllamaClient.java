package com.trialforge.gateway.ollama;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ArrayNode;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.trialforge.gateway.ChatMessage;
import com.trialforge.gateway.ChatStreamer;
import com.trialforge.gateway.DemoLogger;
import com.trialforge.gateway.Embedder;

import java.io.IOException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.ArrayList;
import java.util.List;
import java.util.function.Consumer;
import java.util.stream.Collectors;
import java.util.stream.Stream;

/**
 * Cliente para a API NATIVA do Ollama (http://localhost:11434/api/...) —
 * usa os mesmos dois endpoints que os pacotes {@code ollama} do npm e do
 * PyPI usam por baixo dos panos: POST /api/embeddings (embedding de um
 * texto) e POST /api/chat (chat, com streaming NDJSON quando
 * {@code stream=true}). Diferente do cliente usado em
 * manipulation-guardrail-prototype-java (so /api/chat, sem streaming), este
 * gateway precisa de embeddings reais e de streaming token a token.
 */
public class OllamaClient implements Embedder, ChatStreamer {

    private static final int MAX_TENTATIVAS_PADRAO = 3; // MAX_TENTATIVAS_OLLAMA nos dois originais
    private static final long TIMEOUT_MS_PADRAO = 20_000; // TIMEOUT_OLLAMA_MS / TIMEOUT_OLLAMA_S

    private final String baseUrl;
    private final String embeddingModel;
    private final HttpClient httpClient;
    private final ObjectMapper mapper = new ObjectMapper();

    private DemoLogger logger;
    private int maxTentativas = MAX_TENTATIVAS_PADRAO;
    private long timeoutMs = TIMEOUT_MS_PADRAO;

    public OllamaClient(String baseUrl, String embeddingModel) {
        this.baseUrl = baseUrl;
        this.embeddingModel = embeddingModel;
        this.httpClient = HttpClient.newBuilder().connectTimeout(Duration.ofSeconds(10)).build();
    }

    public void setLogger(DemoLogger logger) {
        this.logger = logger;
    }

    public void setMaxTentativas(int maxTentativas) {
        this.maxTentativas = maxTentativas;
    }

    public void setTimeoutMs(long timeoutMs) {
        this.timeoutMs = timeoutMs;
    }

    // ---------- Embeddings reais (base de Semantic Cache e de confianca do RAG) ----------

    @Override
    public List<Double> embedar(String texto) throws Exception {
        return RetrySupport.comRetry(() -> doEmbed(texto), maxTentativas, timeoutMs, "Embedding", logger);
    }

    private List<Double> doEmbed(String texto) throws IOException, InterruptedException {
        ObjectNode body = mapper.createObjectNode();
        body.put("model", embeddingModel);
        body.put("prompt", texto);

        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(baseUrl + "/api/embeddings"))
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(body.toString()))
                .build();

        HttpResponse<String> response;
        try {
            response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
        } catch (IOException e) {
            throw new IOException(
                    "ollama: falha na requisição de embedding (Ollama está rodando? 'ollama serve'): " + e.getMessage(), e);
        }

        if (response.statusCode() != 200) {
            throw new IOException("ollama: erro da API de embedding (status " + response.statusCode() + "): " + response.body());
        }

        JsonNode parsed = mapper.readTree(response.body());
        JsonNode embeddingNode = parsed.path("embedding");
        if (!embeddingNode.isArray() || embeddingNode.isEmpty()) {
            throw new IOException("ollama: resposta de embedding sem vetor: " + response.body());
        }
        List<Double> embedding = new ArrayList<>();
        for (JsonNode v : embeddingNode) {
            embedding.add(v.asDouble());
        }
        return embedding;
    }

    // ---------- Geracao com streaming (Modulo 4.3) ----------

    /**
     * O retry+timeout (RetrySupport.comRetry) protege so o ESTABELECIMENTO da
     * conexao (obter os headers da resposta, via httpClient.send com
     * BodyHandlers.ofLines() — que so bloqueia ate os headers chegarem, a
     * leitura do corpo e lazy) — mesmo alcance do
     * {@code comRetry(() => ollama.chat(...))} no JS e do
     * {@code com_retry(lambda: chat(...))} no Python: o Promise/generator ali
     * tambem so representa o inicio da chamada, nao a iteracao completa dos
     * chunks. Depois de estabelecida, a leitura das linhas NDJSON nao e
     * retentada — uma linha corrompida no meio do stream propaga o erro pra
     * cima, igual aos originais (que tambem nao tem retry por chunk).
     */
    @Override
    public void chatStream(String model, List<ChatMessage> messages, Consumer<String> onChunk) throws Exception {
        Stream<String> linhas = RetrySupport.comRetry(
                () -> abrirChatStream(model, messages), maxTentativas, timeoutMs, "Geração de resposta", logger);

        try (linhas) {
            for (String linha : (Iterable<String>) linhas::iterator) {
                if (linha.isBlank()) {
                    continue;
                }
                JsonNode parte = mapper.readTree(linha);
                if (parte.has("error")) {
                    throw new IOException("ollama: erro durante streaming: " + parte.get("error").asText());
                }
                // Modelo de raciocinio: chunks de "thinking" vem num campo
                // separado, nunca em message.content — por isso so repassamos
                // .content, igual aos dois originais.
                String content = parte.path("message").path("content").asText("");
                if (!content.isEmpty()) {
                    onChunk.accept(content);
                }
                if (parte.path("done").asBoolean(false)) {
                    break;
                }
            }
        }
    }

    private Stream<String> abrirChatStream(String model, List<ChatMessage> messages) throws IOException, InterruptedException {
        ObjectNode body = mapper.createObjectNode();
        body.put("model", model);
        body.put("stream", true);
        ArrayNode msgs = body.putArray("messages");
        for (ChatMessage m : messages) {
            ObjectNode node = msgs.addObject();
            node.put("role", m.role());
            node.put("content", m.content());
        }

        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(baseUrl + "/api/chat"))
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(body.toString()))
                .build();

        HttpResponse<Stream<String>> response;
        try {
            response = httpClient.send(request, HttpResponse.BodyHandlers.ofLines());
        } catch (IOException e) {
            throw new IOException(
                    "ollama: falha na requisição de chat (Ollama está rodando? 'ollama serve'): " + e.getMessage(), e);
        }

        if (response.statusCode() != 200) {
            String corpo = response.body().collect(Collectors.joining("\n"));
            throw new IOException("ollama: erro da API de chat (status " + response.statusCode() + "): " + corpo);
        }
        return response.body();
    }
}
