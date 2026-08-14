package tool

import "encoding/json"

// GetMetricsTool retorna métricas de latência simuladas para um serviço —
// equivalente a GetMetricsTool.java. Em produção chamaria um sistema de
// métricas real como Prometheus.
type GetMetricsTool struct{}

func (GetMetricsTool) Name() string { return "getMetrics" }

func (GetMetricsTool) Execute(args map[string]any) string {
	service := stringArg(args, "service", "api-gateway")

	metrics := map[string]any{
		"service":   service,
		"timestamp": "2024-01-15T14:30:00Z",
		"latency_ms": map[string]any{
			"p50": 45,
			"p95": 320,
			"p99": 1850,
			"max": 4200,
		},
		"error_rate_pct":      12.5,
		"requests_per_second": 450,
		"status":              "DEGRADED",
		"alert":               "p99 acima do SLO de 1000ms",
	}

	out, err := json.Marshal(metrics)
	if err != nil {
		return `{"error": "Falha ao serializar métricas"}`
	}
	return string(out)
}
