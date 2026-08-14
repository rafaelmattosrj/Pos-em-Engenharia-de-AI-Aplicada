package com.iadeva.nexustools.controller;

import com.iadeva.nexustools.service.K8sDiagService;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

/**
 * Equivalente aos endpoints das tools de tools/k8s_diag.py.
 */
@RestController
public class K8sDiagController {

    private final K8sDiagService service;

    public K8sDiagController(K8sDiagService service) {
        this.service = service;
    }

    public record PodFailureRequest(String podName) {}
    public record SuggestFixRequest(String issueType) {}

    @PostMapping("/tools/inspect-pod-failure")
    public ToolResponse inspectPodFailure(@RequestBody PodFailureRequest req) {
        return new ToolResponse(service.inspectPodFailure(req.podName()));
    }

    @PostMapping("/tools/suggest-fix")
    public ToolResponse suggestFix(@RequestBody SuggestFixRequest req) {
        return new ToolResponse(service.suggestFix(req.issueType()));
    }
}
