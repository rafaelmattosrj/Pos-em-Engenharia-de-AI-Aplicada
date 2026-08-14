package com.iadeva.nexustools.controller;

import com.iadeva.nexustools.service.GovernanceService;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

/**
 * Equivalente aos endpoints das tools de tools/governance_tools.py.
 */
@RestController
public class GovernanceController {

    private final GovernanceService service;

    public GovernanceController(GovernanceService service) {
        this.service = service;
    }

    public record TriageRequest(String trivyJsonReport) {}
    public record OptimizeCicdRequest(String workflowYamlContent) {}
    public record FinopsRequest(String currentResourcesInventory) {}

    @PostMapping("/tools/triage-security-vulnerabilities")
    public ToolResponse triageSecurityVulnerabilities(@RequestBody TriageRequest req) {
        return new ToolResponse(service.triageSecurityVulnerabilities(req.trivyJsonReport()));
    }

    @PostMapping("/tools/optimize-cicd-pipeline")
    public ToolResponse optimizeCicdPipeline(@RequestBody OptimizeCicdRequest req) {
        return new ToolResponse(service.optimizeCicdPipeline(req.workflowYamlContent()));
    }

    @PostMapping("/tools/analyze-finops-costs")
    public ToolResponse analyzeFinopsCosts(@RequestBody FinopsRequest req) {
        return new ToolResponse(service.analyzeFinopsCosts(req.currentResourcesInventory()));
    }
}
