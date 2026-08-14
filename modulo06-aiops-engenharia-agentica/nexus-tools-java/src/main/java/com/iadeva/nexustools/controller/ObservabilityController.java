package com.iadeva.nexustools.controller;

import com.iadeva.nexustools.service.ObservabilityService;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

/**
 * Equivalente aos endpoints das tools de tools/obs_tools.py.
 */
@RestController
public class ObservabilityController {

    private final ObservabilityService service;

    public ObservabilityController(ObservabilityService service) {
        this.service = service;
    }

    public record PrometheusQueryRequest(String query) {}
    public record JaegerTraceRequest(String serviceName) {}

    @PostMapping("/tools/query-prometheus-metrics")
    public ToolResponse queryPrometheusMetrics(@RequestBody PrometheusQueryRequest req) {
        return new ToolResponse(service.queryPrometheusMetrics(req.query()));
    }

    @PostMapping("/tools/query-jaeger-traces")
    public ToolResponse queryJaegerTraces(@RequestBody JaegerTraceRequest req) {
        return new ToolResponse(service.queryJaegerTraces(req.serviceName()));
    }
}
