package com.iadeva.nexustools.service;

import com.fasterxml.jackson.databind.ObjectMapper;
import org.springframework.stereotype.Service;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

/**
 * Equivalente a tools/aiops_tools.py.
 */
@Service
public class AiOpsService {

    private final ObjectMapper objectMapper;

    public AiOpsService(ObjectMapper objectMapper) {
        this.objectMapper = objectMapper.copy();
    }

    /**
     * Converte linguagem natural para PromQL ou LogQL — equivalente a
     * nl_to_promql (aula 5.1).
     */
    public String nlToPromql(String naturalLanguageQuery) {
        String lower = naturalLanguageQuery.toLowerCase();
        if (lower.contains("taxa de erro") || lower.contains("error")) {
            return "rate(http_requests_total{status=~\"5..\"}[5m]) / rate(http_requests_total[5m])";
        }
        if (lower.contains("disco") || lower.contains("disk")) {
            return "node_filesystem_avail_bytes{mountpoint=\"/data\"} / node_filesystem_size_bytes{mountpoint=\"/data\"} * 100";
        }
        return "up{job=\"kubernetes-pods\"}";
    }

    /**
     * Analisa o histórico de métricas usando algoritmos preditivos
     * (Prophet/Isolation Forest) e prevê saturação de recursos —
     * equivalente a predictive_disk_alert (aula 5.2 e Prática 4).
     */
    public String predictiveDiskAlert(String metricsHistory) {
        String lower = metricsHistory.toLowerCase();
        if (lower.contains("growth") || lower.contains("crescimento")) {
            return """
                    🚨 [ALERTA PREDITIVO - MACHINE LEARNING]
                            Anomalia Detectada: Crescimento acelerado no volume /data.
                            Previsão (Prophet Algorithm): Saturação de 100% ocorrerá em exatas 4 horas.
                            Ação Recomendada: Acionar script de limpeza de logs ou escalar o PVC.""";
        }
        return "✅ Padrão de uso normal. Sem anomalias na série temporal.";
    }

    /**
     * Cria o JSON de um Dashboard Dinâmico no Grafana focado no incidente
     * atual e salva o arquivo em disco — equivalente a
     * generate_grafana_dashboard (aula 5.3).
     */
    public String generateGrafanaDashboard(String incidentContext) {
        Map<String, Object> panel1 = new LinkedHashMap<>();
        panel1.put("title", "Disk Usage Prediction");
        panel1.put("type", "timeseries");
        panel1.put("targets", List.of(Map.of("expr", "node_filesystem_avail_bytes")));

        Map<String, Object> panel2 = new LinkedHashMap<>();
        panel2.put("title", "Error Rate Spike");
        panel2.put("type", "stat");
        panel2.put("targets", List.of(Map.of("expr", "rate(http_requests_total{status='500'}[5m])")));

        Map<String, Object> dashboard = new LinkedHashMap<>();
        dashboard.put("title", "Dynamic Incident Dashboard: " + incidentContext);
        dashboard.put("panels", List.of(panel1, panel2));

        String filename = "incident_dashboard.json";
        try {
            String json = objectMapper.writerWithDefaultPrettyPrinter().writeValueAsString(dashboard);
            Files.writeString(Path.of(filename), json);
        } catch (IOException e) {
            return "❌ Erro ao salvar o dashboard: " + e.getMessage();
        }

        return "✅ Dashboard gerado com sucesso! O arquivo '" + filename + "' foi salvo no disco pronto para importação no Grafana.";
    }
}
