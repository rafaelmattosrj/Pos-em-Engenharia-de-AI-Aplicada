package service

import (
	"fmt"
	"strings"
)

// ObservabilityService — equivalente a tools/obs_tools.py.
type ObservabilityService struct{}

// QueryPrometheusMetrics executa uma consulta PromQL no Prometheus para
// analisar métricas de CPU, memória ou latência — equivalente a
// query_prometheus_metrics.
func (ObservabilityService) QueryPrometheusMetrics(query string) string {
	lower := strings.ToLower(query)
	switch {
	case strings.Contains(lower, "latency") || strings.Contains(lower, "duration") || strings.Contains(lower, "latência"):
		return "📊 Prometheus Result: Average latency at endpoint '/checkout' is 850ms (HIGH). P99 threshold exceeded."
	case strings.Contains(lower, "error") || strings.Contains(lower, "taxa de erro"):
		return "📊 Prometheus Result: 5XX error rate is currently at 12% over the last 5 minutes."
	default:
		return "📊 Prometheus Result: Metrics are well within the normal baseline."
	}
}

// QueryJaegerTraces consulta o rastreamento distribuído do Jaeger para
// identificar gargalos de latência em serviços — equivalente a
// query_jaeger_traces.
func (ObservabilityService) QueryJaegerTraces(serviceName string) string {
	return fmt.Sprintf("🔍 Jaeger Trace: The performance bottleneck for service '%s' "+
		"is located in the database call to PostgreSQL (Span duration: 800ms).", serviceName)
}
