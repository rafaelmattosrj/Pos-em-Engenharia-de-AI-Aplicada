package com.iadeva.agent.tool;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.springframework.stereotype.Component;

import java.util.Map;

/**
 * Retorna métricas de latência simuladas para um serviço.
 * Tool simulada — em produção chamaria um sistema de métricas real como Prometheus.
 */
@Component
public class GetMetricsTool implements AgentTool {

    private final ObjectMapper objectMapper;

    public GetMetricsTool(ObjectMapper objectMapper) {
        this.objectMapper = objectMapper;
    }

    @Override
    public String getName() {
        return "getMetrics";
    }

    @Override
    public String execute(Map<String, Object> args) {
        String service = (String) args.getOrDefault("service", "api-gateway");

        // Dados simulados de métricas de latência — em produção viria do Prometheus/Grafana
        Map<String, Object> metrics = Map.of(
                "service", service,
                "timestamp", "2024-01-15T14:30:00Z",
                "latency_ms", Map.of(
                        "p50", 45,
                        "p95", 320,
                        "p99", 1850,
                        "max", 4200
                ),
                "error_rate_pct", 12.5,
                "requests_per_second", 450,
                "status", "DEGRADED",
                "alert", "p99 acima do SLO de 1000ms"
        );

        try {
            return objectMapper.writeValueAsString(metrics);
        } catch (Exception e) {
            return "{\"error\": \"Falha ao serializar métricas\"}";
        }
    }
}
