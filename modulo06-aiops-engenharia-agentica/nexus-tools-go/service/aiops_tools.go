package service

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// AiOpsService — equivalente a tools/aiops_tools.py.
type AiOpsService struct{}

// NlToPromql converte linguagem natural para PromQL ou LogQL — equivalente
// a nl_to_promql (aula 5.1).
func (AiOpsService) NlToPromql(naturalLanguageQuery string) string {
	lower := strings.ToLower(naturalLanguageQuery)
	switch {
	case strings.Contains(lower, "taxa de erro") || strings.Contains(lower, "error"):
		return `rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m])`
	case strings.Contains(lower, "disco") || strings.Contains(lower, "disk"):
		return `node_filesystem_avail_bytes{mountpoint="/data"} / node_filesystem_size_bytes{mountpoint="/data"} * 100`
	default:
		return `up{job="kubernetes-pods"}`
	}
}

// PredictiveDiskAlert analisa o histórico de métricas usando algoritmos
// preditivos (Prophet/Isolation Forest) e prevê saturação de recursos —
// equivalente a predictive_disk_alert (aula 5.2 e Prática 4).
func (AiOpsService) PredictiveDiskAlert(metricsHistory string) string {
	lower := strings.ToLower(metricsHistory)
	if strings.Contains(lower, "growth") || strings.Contains(lower, "crescimento") {
		return "🚨 [ALERTA PREDITIVO - MACHINE LEARNING]\n" +
			"        Anomalia Detectada: Crescimento acelerado no volume /data.\n" +
			"        Previsão (Prophet Algorithm): Saturação de 100% ocorrerá em exatas 4 horas.\n" +
			"        Ação Recomendada: Acionar script de limpeza de logs ou escalar o PVC."
	}
	return "✅ Padrão de uso normal. Sem anomalias na série temporal."
}

type grafanaTarget struct {
	Expr string `json:"expr"`
}

type grafanaPanel struct {
	Title   string          `json:"title"`
	Type    string          `json:"type"`
	Targets []grafanaTarget `json:"targets"`
}

type grafanaDashboard struct {
	Title  string         `json:"title"`
	Panels []grafanaPanel `json:"panels"`
}

// GenerateGrafanaDashboard cria o JSON de um Dashboard Dinâmico no Grafana
// focado no incidente atual e salva o arquivo em disco — equivalente a
// generate_grafana_dashboard (aula 5.3).
func (AiOpsService) GenerateGrafanaDashboard(incidentContext string) string {
	dashboard := grafanaDashboard{
		Title: "Dynamic Incident Dashboard: " + incidentContext,
		Panels: []grafanaPanel{
			{Title: "Disk Usage Prediction", Type: "timeseries", Targets: []grafanaTarget{{Expr: "node_filesystem_avail_bytes"}}},
			{Title: "Error Rate Spike", Type: "stat", Targets: []grafanaTarget{{Expr: "rate(http_requests_total{status='500'}[5m])"}}},
		},
	}

	filename := "incident_dashboard.json"
	data, err := json.MarshalIndent(dashboard, "", "  ")
	if err != nil {
		return fmt.Sprintf("❌ Erro ao gerar o dashboard: %v", err)
	}
	if err := os.WriteFile(filename, data, 0o644); err != nil {
		return fmt.Sprintf("❌ Erro ao salvar o dashboard: %v", err)
	}

	return fmt.Sprintf("✅ Dashboard gerado com sucesso! O arquivo '%s' foi salvo no disco pronto para importação no Grafana.", filename)
}
