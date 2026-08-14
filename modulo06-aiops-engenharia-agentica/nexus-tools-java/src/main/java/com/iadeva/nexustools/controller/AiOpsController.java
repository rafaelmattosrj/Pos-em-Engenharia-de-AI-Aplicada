package com.iadeva.nexustools.controller;

import com.iadeva.nexustools.service.AiOpsService;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

/**
 * Equivalente aos endpoints das tools de tools/aiops_tools.py.
 */
@RestController
public class AiOpsController {

    private final AiOpsService service;

    public AiOpsController(AiOpsService service) {
        this.service = service;
    }

    public record NlToPromqlRequest(String naturalLanguageQuery) {}
    public record PredictiveDiskAlertRequest(String metricsHistory) {}
    public record GrafanaDashboardRequest(String incidentContext) {}

    @PostMapping("/tools/nl-to-promql")
    public ToolResponse nlToPromql(@RequestBody NlToPromqlRequest req) {
        return new ToolResponse(service.nlToPromql(req.naturalLanguageQuery()));
    }

    @PostMapping("/tools/predictive-disk-alert")
    public ToolResponse predictiveDiskAlert(@RequestBody PredictiveDiskAlertRequest req) {
        return new ToolResponse(service.predictiveDiskAlert(req.metricsHistory()));
    }

    @PostMapping("/tools/generate-grafana-dashboard")
    public ToolResponse generateGrafanaDashboard(@RequestBody GrafanaDashboardRequest req) {
        return new ToolResponse(service.generateGrafanaDashboard(req.incidentContext()));
    }
}
