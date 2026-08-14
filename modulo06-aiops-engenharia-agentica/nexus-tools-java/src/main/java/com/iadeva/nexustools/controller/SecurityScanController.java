package com.iadeva.nexustools.controller;

import com.iadeva.nexustools.service.SecurityScanService;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

/**
 * Equivalente aos endpoints das tools de tools/security_scan.py.
 */
@RestController
public class SecurityScanController {

    private final SecurityScanService service;

    public SecurityScanController(SecurityScanService service) {
        this.service = service;
    }

    public record CheckovScanRequest(String filename) {}
    public record OpaPoliciesRequest(String content) {}

    @PostMapping("/tools/run-checkov-scan")
    public ToolResponse runCheckovScan(@RequestBody CheckovScanRequest req) {
        return new ToolResponse(service.runCheckovScan(req.filename()));
    }

    @PostMapping("/tools/validate-opa-policies")
    public ToolResponse validateOpaPolicies(@RequestBody OpaPoliciesRequest req) {
        return new ToolResponse(service.validateOpaPolicies(req.content()));
    }
}
