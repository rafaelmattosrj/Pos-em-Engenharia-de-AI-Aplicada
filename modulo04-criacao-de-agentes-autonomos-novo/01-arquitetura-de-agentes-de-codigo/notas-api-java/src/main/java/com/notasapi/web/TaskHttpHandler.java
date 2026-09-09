package com.notasapi.web;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ArrayNode;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.notasapi.domain.Task;
import com.notasapi.domain.TaskListFilter;
import com.notasapi.service.TaskNotFoundException;
import com.notasapi.service.TaskService;
import com.notasapi.service.TaskValidationException;
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpHandler;

import java.io.IOException;
import java.io.OutputStream;
import java.nio.charset.StandardCharsets;
import java.util.List;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

/**
 * Porte 1:1 das rotas de http/task-routes.ts, usando com.sun.net.httpserver
 * (builtin do JDK) em vez de framework -- ver README para a justificativa
 * (4 rotas simples nao pedem Spring Boot; ver "Flexibilidade de framework"
 * na skill portar-projetos-java-go).
 */
public class TaskHttpHandler implements HttpHandler {

    private static final ObjectMapper MAPPER = new ObjectMapper();
    private static final Pattern COMPLETE_PATH = Pattern.compile("^/tasks/([^/]+)/complete$");
    private static final Pattern TASK_PATH = Pattern.compile("^/tasks/([^/]+)$");

    private final TaskService taskService;

    public TaskHttpHandler(TaskService taskService) {
        this.taskService = taskService;
    }

    @Override
    public void handle(HttpExchange exchange) throws IOException {
        try {
            String method = exchange.getRequestMethod();
            String path = exchange.getRequestURI().getPath();
            String query = exchange.getRequestURI().getQuery();

            if ("POST".equals(method) && "/tasks".equals(path)) {
                ObjectNode body = (ObjectNode) MAPPER.readTree(exchange.getRequestBody());
                String title = body.hasNonNull("title") ? body.get("title").asText() : null;
                Task task = taskService.createTask(title);
                sendJson(exchange, 201, toJson(task));
                return;
            }

            if ("GET".equals(method) && "/tasks".equals(path)) {
                String statusParam = queryParam(query, "status").orElse("all");
                TaskListFilter filter = TaskListFilter.fromWireValue(statusParam);
                List<Task> tasks = taskService.listTasks(filter);
                ArrayNode array = MAPPER.createArrayNode();
                tasks.forEach(task -> array.add(toJson(task)));
                sendJson(exchange, 200, array);
                return;
            }

            Matcher completeMatch = COMPLETE_PATH.matcher(path);
            if ("PATCH".equals(method) && completeMatch.matches()) {
                Task task = taskService.completeTask(decode(completeMatch.group(1)));
                sendJson(exchange, 200, toJson(task));
                return;
            }

            Matcher taskMatch = TASK_PATH.matcher(path);
            if ("DELETE".equals(method) && taskMatch.matches()) {
                taskService.removeTask(decode(taskMatch.group(1)));
                sendEmpty(exchange, 204);
                return;
            }

            ObjectNode notFound = MAPPER.createObjectNode();
            notFound.put("error", "Route not found");
            sendJson(exchange, 404, notFound);
        } catch (TaskValidationException e) {
            ObjectNode body = MAPPER.createObjectNode();
            body.put("error", "Validation failed");
            ArrayNode issues = MAPPER.createArrayNode();
            e.issues().forEach(issues::add);
            body.set("issues", issues);
            sendJson(exchange, 400, body);
        } catch (TaskNotFoundException e) {
            ObjectNode body = MAPPER.createObjectNode();
            body.put("error", e.getMessage());
            sendJson(exchange, 404, body);
        } catch (com.fasterxml.jackson.core.JsonProcessingException e) {
            ObjectNode body = MAPPER.createObjectNode();
            body.put("error", "Invalid JSON body");
            sendJson(exchange, 400, body);
        } catch (RuntimeException e) {
            ObjectNode body = MAPPER.createObjectNode();
            body.put("error", "Internal server error");
            sendJson(exchange, 500, body);
        }
    }

    private static ObjectNode toJson(Task task) {
        ObjectNode node = MAPPER.createObjectNode();
        node.put("id", task.id());
        node.put("title", task.title());
        node.put("status", task.status().wireValue());
        return node;
    }

    private static java.util.Optional<String> queryParam(String query, String name) {
        if (query == null) {
            return java.util.Optional.empty();
        }
        for (String pair : query.split("&")) {
            String[] parts = pair.split("=", 2);
            if (parts.length == 2 && parts[0].equals(name)) {
                return java.util.Optional.of(decode(parts[1]));
            }
        }
        return java.util.Optional.empty();
    }

    private static String decode(String value) {
        return java.net.URLDecoder.decode(value, StandardCharsets.UTF_8);
    }

    private static void sendJson(HttpExchange exchange, int statusCode, Object body) throws IOException {
        byte[] bytes = MAPPER.writeValueAsBytes(body);
        exchange.getResponseHeaders().add("content-type", "application/json; charset=utf-8");
        exchange.sendResponseHeaders(statusCode, bytes.length);
        try (OutputStream os = exchange.getResponseBody()) {
            os.write(bytes);
        }
    }

    private static void sendEmpty(HttpExchange exchange, int statusCode) throws IOException {
        exchange.sendResponseHeaders(statusCode, -1);
        exchange.getResponseBody().close();
    }
}
