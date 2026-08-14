package com.iadeva.nexustools.service;

import org.springframework.stereotype.Service;

/**
 * Equivalente a tools/obs_tools.py.
 */
@Service
public class ObservabilityService {

    /**
     * Executa uma consulta PromQL no Prometheus para analisar métricas de
     * CPU, memória ou latência — equivalente a query_prometheus_metrics.
     */
    public String queryPrometheusMetrics(String query) {
        String lower = query.toLowerCase();
        if (lower.contains("latency") || lower.contains("duration") || lower.contains("latência")) {
            return "📊 Prometheus Result: Average latency at endpoint '/checkout' is 850ms (HIGH). P99 threshold exceeded.";
        }
        if (lower.contains("error") || lower.contains("taxa de erro")) {
            return "📊 Prometheus Result: 5XX error rate is currently at 12% over the last 5 minutes.";
        }
        return "📊 Prometheus Result: Metrics are well within the normal baseline.";
    }

    /**
     * Consulta o rastreamento distribuído do Jaeger para identificar
     * gargalos de latência em serviços — equivalente a
     * query_jaeger_traces.
     */
    public String queryJaegerTraces(String serviceName) {
        return "🔍 Jaeger Trace: The performance bottleneck for service '" + serviceName
                + "' is located in the database call to PostgreSQL (Span duration: 800ms).";
    }
}
