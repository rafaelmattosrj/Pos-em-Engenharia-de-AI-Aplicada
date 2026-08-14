package com.iadeva.nexustools.controller;

import com.iadeva.nexustools.service.K8sOpsService;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

/**
 * Equivalente aos endpoints das tools de tools/k8s_ops.py.
 */
@RestController
public class K8sOpsController {

    private final K8sOpsService service;

    public K8sOpsController(K8sOpsService service) {
        this.service = service;
    }

    public record GenerateManifestRequest(String appName, int replicas, int port) {}
    public record ApplyManifestRequest(String filename) {}
    public record CanaryMetricsRequest(String metricsData) {}

    @PostMapping("/tools/generate-k8s-manifest")
    public ToolResponse generateK8sManifest(@RequestBody GenerateManifestRequest req) {
        return new ToolResponse(service.generateK8sManifest(req.appName(), req.replicas(), req.port()));
    }

    @PostMapping("/tools/apply-k8s-manifest")
    public ToolResponse applyK8sManifest(@RequestBody ApplyManifestRequest req) {
        return new ToolResponse(service.applyK8sManifest(req.filename()));
    }

    @PostMapping("/tools/analyze-canary-metrics")
    public ToolResponse analyzeCanaryMetrics(@RequestBody CanaryMetricsRequest req) {
        return new ToolResponse(service.analyzeCanaryMetrics(req.metricsData()));
    }
}
