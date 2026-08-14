package com.trialforge.gateway.ollama;

import com.sun.net.httpserver.HttpServer;
import com.trialforge.gateway.ChatMessage;
import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.Test;

import java.io.IOException;
import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.concurrent.CountDownLatch;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.atomic.AtomicInteger;
import java.util.concurrent.atomic.AtomicReference;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

class OllamaClientTest {

    private HttpServer server;

    @AfterEach
    void pararServidor() {
        if (server != null) {
            server.stop(0);
        }
    }

    private String iniciarServidor(com.sun.net.httpserver.HttpHandler handler) throws IOException {
        server = HttpServer.create(new InetSocketAddress("localhost", 0), 0);
        server.createContext("/", handler);
        server.start();
        return "http://localhost:" + server.getAddress().getPort();
    }

    private static void responder(com.sun.net.httpserver.HttpExchange exchange, int status, String corpo) throws IOException {
        byte[] bytes = corpo.getBytes(StandardCharsets.UTF_8);
        exchange.getResponseHeaders().add("Content-Type", "application/json");
        exchange.sendResponseHeaders(status, bytes.length);
        try (OutputStream os = exchange.getResponseBody()) {
            os.write(bytes);
        }
    }

    @Test
    void embedarComSucesso() throws Exception {
        String baseUrl = iniciarServidor(exchange -> {
            assertThat(exchange.getRequestURI().getPath()).isEqualTo("/api/embeddings");
            responder(exchange, 200, "{\"embedding\": [0.1, 0.2, 0.3]}");
        });

        OllamaClient client = new OllamaClient(baseUrl, "nomic-embed-text");
        List<Double> embedding = client.embedar("texto de teste");

        assertThat(embedding).containsExactly(0.1, 0.2, 0.3);
    }

    @Test
    void embedarComRetryAposFalhaTransitoria() throws Exception {
        AtomicInteger tentativas = new AtomicInteger();
        String baseUrl = iniciarServidor(exchange -> {
            if (tentativas.incrementAndGet() < 2) {
                responder(exchange, 500, "erro transitório");
            } else {
                responder(exchange, 200, "{\"embedding\": [1.0]}");
            }
        });

        OllamaClient client = new OllamaClient(baseUrl, "nomic-embed-text");
        List<Double> embedding = client.embedar("texto");

        assertThat(tentativas.get()).isEqualTo(2);
        assertThat(embedding).containsExactly(1.0);
    }

    @Test
    void embedarEsgotaTentativas() throws Exception {
        AtomicInteger tentativas = new AtomicInteger();
        String baseUrl = iniciarServidor(exchange -> {
            tentativas.incrementAndGet();
            responder(exchange, 500, "sempre falha");
        });

        OllamaClient client = new OllamaClient(baseUrl, "nomic-embed-text");
        client.setMaxTentativas(2);

        assertThatThrownBy(() -> client.embedar("texto")).isInstanceOf(Exception.class);
        assertThat(tentativas.get()).isEqualTo(2);
    }

    @Test
    void embedarComTimeout() throws Exception {
        CountDownLatch travar = new CountDownLatch(1);
        String baseUrl = iniciarServidor(exchange -> {
            try {
                travar.await(2, TimeUnit.SECONDS);
            } catch (InterruptedException ignored) {
                Thread.currentThread().interrupt();
            }
            responder(exchange, 200, "{\"embedding\": [1.0]}");
        });

        OllamaClient client = new OllamaClient(baseUrl, "nomic-embed-text");
        client.setMaxTentativas(1);
        client.setTimeoutMs(50);

        assertThatThrownBy(() -> client.embedar("texto")).isInstanceOf(Exception.class);
        travar.countDown();
    }

    @Test
    void chatStreamComSucesso() throws Exception {
        String baseUrl = iniciarServidor(exchange -> {
            assertThat(exchange.getRequestURI().getPath()).isEqualTo("/api/chat");
            String corpo = String.join("\n",
                    "{\"message\":{\"role\":\"assistant\",\"content\":\"Olá\"},\"done\":false}",
                    "{\"message\":{\"role\":\"assistant\",\"content\":\", mundo\"},\"done\":false}",
                    "{\"message\":{\"role\":\"assistant\",\"content\":\"\"},\"done\":true}") + "\n";
            responder(exchange, 200, corpo);
        });

        OllamaClient client = new OllamaClient(baseUrl, "nomic-embed-text");
        StringBuilder recebido = new StringBuilder();
        client.chatStream("gemma4:e2b", List.of(new ChatMessage("user", "oi")), recebido::append);

        assertThat(recebido.toString()).isEqualTo("Olá, mundo");
    }

    @Test
    void chatStreamComErroNaLinha() throws Exception {
        String baseUrl = iniciarServidor(exchange -> responder(exchange, 200, "{\"error\":\"modelo não encontrado\"}\n"));

        OllamaClient client = new OllamaClient(baseUrl, "nomic-embed-text");
        AtomicReference<String> recebido = new AtomicReference<>("");

        assertThatThrownBy(() -> client.chatStream("modelo-inexistente", List.of(), s -> recebido.set(recebido.get() + s)))
                .isInstanceOf(Exception.class);
    }

    @Test
    void chatStreamComStatusNaoOk() throws Exception {
        String baseUrl = iniciarServidor(exchange -> responder(exchange, 503, "indisponível"));

        OllamaClient client = new OllamaClient(baseUrl, "nomic-embed-text");
        client.setMaxTentativas(1);

        assertThatThrownBy(() -> client.chatStream("modelo", List.of(), s -> { }))
                .isInstanceOf(Exception.class);
    }
}
