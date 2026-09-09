package com.opspilot.web;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ArrayNode;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.opspilot.domain.ReasoningStrategy;
import com.opspilot.domain.StrategyRunInput;
import com.opspilot.domain.StrategyResult;
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;

import java.io.IOException;
import java.io.OutputStream;
import java.util.Map;

/**
 * Porte simplificado de http/server.ts -- so a rota POST /chat, roteando pro
 * nome de estrategia pedido no corpo ({"message": "...", "strategy": "team"}).
 * Sem CORS/streaming/auditoria de request (fora do escopo deste porte).
 */
public class ChatHttpHandler implements HttpHandler {

    private static final ObjectMapper MAPPER = new ObjectMapper();

    private final Map<String, ReasoningStrategy> strategies;
    private final String defaultStrategy;

    public ChatHttpHandler(Map<String, ReasoningStrategy> strategies, String defaultStrategy) {
        this.strategies = strategies;
        this.defaultStrategy = defaultStrategy;
    }

    @Override
    public void handle(HttpExchange exchange) throws IOException {
        if (!"POST".equals(exchange.getRequestMethod()) || !"/chat".equals(exchange.getRequestURI().getPath())) {
            sendJson(exchange, 404, error("Route not found"));
            return;
        }

        ObjectNode body;
        try {
            body = (ObjectNode) MAPPER.readTree(exchange.getRequestBody());
        } catch (IOException e) {
            sendJson(exchange, 400, error("Invalid JSON body"));
            return;
        }

        String message = body.hasNonNull("message") ? body.get("message").asText() : null;
        if (message == null || message.isBlank()) {
            sendJson(exchange, 400, error("Validation failed: message is required"));
            return;
        }

        String strategyName = body.hasNonNull("strategy") ? body.get("strategy").asText() : defaultStrategy;
        ReasoningStrategy strategy = strategies.get(strategyName);
        if (strategy == null) {
            sendJson(exchange, 400, error("Unknown strategy \"" + strategyName + "\""));
            return;
        }

        StrategyResult result = strategy.run(StrategyRunInput.of(message));

        ObjectNode response = MAPPER.createObjectNode();
        response.put("answer", result.answer());
        response.put("strategy", strategy.name());
        ObjectNode metrics = MAPPER.createObjectNode();
        metrics.put("llmCalls", result.metrics().llmCalls());
        metrics.put("latencyMs", result.metrics().latencyMs());
        response.set("metrics", metrics);
        ArrayNode trace = MAPPER.createArrayNode();
        result.trace().forEach(event -> {
            ObjectNode eventNode = MAPPER.createObjectNode();
            eventNode.put("type", event.type().name().toLowerCase());
            eventNode.put("node", event.node());
            eventNode.put("content", event.content());
            if (event.to() != null) {
                eventNode.put("to", event.to());
            }
            trace.add(eventNode);
        });
        response.set("trace", trace);

        sendJson(exchange, 200, response);
    }

    private static ObjectNode error(String message) {
        ObjectNode node = MAPPER.createObjectNode();
        node.put("error", message);
        return node;
    }

    private static void sendJson(HttpExchange exchange, int statusCode, Object body) throws IOException {
        byte[] bytes = MAPPER.writeValueAsBytes(body);
        exchange.getResponseHeaders().add("content-type", "application/json; charset=utf-8");
        exchange.sendResponseHeaders(statusCode, bytes.length);
        try (OutputStream os = exchange.getResponseBody()) {
            os.write(bytes);
        }
    }
}
