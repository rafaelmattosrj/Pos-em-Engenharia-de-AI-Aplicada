package com.iadeva.agent.tool;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.Map;

/**
 * Retorna as últimas linhas de log simuladas de um serviço.
 * Tool simulada — em produção integraria com ELK Stack ou CloudWatch.
 */
@Component
public class GetLogsTool implements AgentTool {

    private final ObjectMapper objectMapper;

    public GetLogsTool(ObjectMapper objectMapper) {
        this.objectMapper = objectMapper;
    }

    @Override
    public String getName() {
        return "getLogs";
    }

    @Override
    public String execute(Map<String, Object> args) {
        String service = (String) args.getOrDefault("service", "api-gateway");
        int lines = args.containsKey("lines") ? ((Number) args.get("lines")).intValue() : 10;

        // Logs simulados — em produção viria do ELK Stack, CloudWatch ou similar
        List<Map<String, String>> logs = List.of(
                Map.of(
                        "timestamp", "2024-01-15T14:29:55Z",
                        "severity", "ERROR",
                        "service", service,
                        "message", "Connection pool exhausted: timeout após 5000ms aguardando conexão disponível"
                ),
                Map.of(
                        "timestamp", "2024-01-15T14:29:58Z",
                        "severity", "ERROR",
                        "service", service,
                        "message", "Database query timeout: SELECT * FROM orders WHERE status='pending' excedeu 3000ms"
                ),
                Map.of(
                        "timestamp", "2024-01-15T14:30:01Z",
                        "severity", "WARN",
                        "service", service,
                        "message", "Retry attempt 3/3 para endpoint /api/payments — falha no upstream"
                ),
                Map.of(
                        "timestamp", "2024-01-15T14:30:05Z",
                        "severity", "ERROR",
                        "service", service,
                        "message", "Unhandled exception: NullPointerException em OrderService.processPayment():142"
                ),
                Map.of(
                        "timestamp", "2024-01-15T14:30:10Z",
                        "severity", "INFO",
                        "service", service,
                        "message", "Deploy v2.3.1 detectado às 14:15 — possível correlação com aumento de erros"
                )
        );

        // Limita ao número de linhas solicitado
        List<Map<String, String>> result = logs.subList(0, Math.min(lines, logs.size()));

        try {
            return objectMapper.writeValueAsString(Map.of(
                    "service", service,
                    "total_logs_returned", result.size(),
                    "logs", result
            ));
        } catch (Exception e) {
            return "{\"error\": \"Falha ao serializar logs\"}";
        }
    }
}
