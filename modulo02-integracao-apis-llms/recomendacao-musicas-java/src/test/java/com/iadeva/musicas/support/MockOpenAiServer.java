package com.iadeva.musicas.support;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.sun.net.httpserver.HttpServer;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.Map;
import java.util.function.Function;

/**
 * Servidor HTTP mock minimalista (JDK puro, com.sun.net.httpserver) que simula o endpoint
 * {@code /v1/chat/completions} de uma API compatível com OpenAI/OpenRouter.
 *
 * Usado para testar os componentes Spring AI (ResilientChatClient, PreferencesService,
 * MusicChatOrchestrator, ChatController) apontando {@code spring.ai.openai.base-url} para
 * {@code localhost}, sem depender de rede ou de uma API key real.
 */
public final class MockOpenAiServer implements AutoCloseable {

    private static final ObjectMapper MAPPER = new ObjectMapper();

    private final HttpServer server;

    private MockOpenAiServer(HttpServer server) {
        this.server = server;
    }

    public String baseUrl() {
        return "http://localhost:" + server.getAddress().getPort();
    }

    @Override
    public void close() {
        server.stop(0);
    }

    /** Responde sempre com o mesmo status/conteúdo, ignorando o corpo da requisição recebida. */
    public static MockOpenAiServer fixedResponse(int status, String content) {
        return start(body -> new Response(status, content));
    }

    /** Handler dinâmico: recebe o corpo decodificado da requisição (model, messages, ...) e decide a resposta. */
    @SuppressWarnings("unchecked")
    public static MockOpenAiServer start(Function<Map<String, Object>, Response> handler) {
        try {
            HttpServer httpServer = HttpServer.create(new InetSocketAddress("localhost", 0), 0);
            httpServer.createContext("/v1/chat/completions", exchange -> {
                Map<String, Object> body = MAPPER.readValue(exchange.getRequestBody(), Map.class);
                Response response = handler.apply(body);

                byte[] payload;
                if (response.status() >= 200 && response.status() < 300) {
                    payload = MAPPER.writeValueAsBytes(Map.of(
                            "id", "mock-1",
                            "object", "chat.completion",
                            "created", 1L,
                            "model", String.valueOf(body.getOrDefault("model", "mock-model")),
                            "choices", List.of(Map.of(
                                    "index", 0,
                                    "message", Map.of("role", "assistant", "content", response.content()),
                                    "finish_reason", "stop"
                            ))
                    ));
                } else {
                    payload = ("{\"error\":{\"message\":\"" + response.content() + "\"}}")
                            .getBytes(StandardCharsets.UTF_8);
                }

                exchange.getResponseHeaders().add("Content-Type", "application/json");
                exchange.sendResponseHeaders(response.status(), payload.length);
                exchange.getResponseBody().write(payload);
                exchange.close();
            });
            httpServer.start();
            return new MockOpenAiServer(httpServer);
        } catch (IOException e) {
            throw new RuntimeException("Falha ao iniciar mock server", e);
        }
    }

    public record Response(int status, String content) {}
}
