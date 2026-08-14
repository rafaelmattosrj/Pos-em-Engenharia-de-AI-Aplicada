package com.iadeva.nexustools.controller;

import com.iadeva.nexustools.service.PolicyRagService;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

/**
 * Equivalente ao endpoint da tool de tools/policy_rag.py.
 */
@RestController
public class PolicyRagController {

    private final PolicyRagService service;

    public PolicyRagController(PolicyRagService service) {
        this.service = service;
    }

    public record ComplianceRequest(String query) {}

    @PostMapping("/tools/check-compliance-rules")
    public ToolResponse checkComplianceRules(@RequestBody ComplianceRequest req) {
        return new ToolResponse(service.checkComplianceRules(req.query()));
    }
}
