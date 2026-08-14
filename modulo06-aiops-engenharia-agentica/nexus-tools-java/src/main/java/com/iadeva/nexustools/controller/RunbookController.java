package com.iadeva.nexustools.controller;

import com.iadeva.nexustools.service.RunbookService;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

/**
 * Equivalente ao endpoint da tool consult_runbook (labs/modulo10_remediation.py).
 */
@RestController
public class RunbookController {

    private final RunbookService service;

    public RunbookController(RunbookService service) {
        this.service = service;
    }

    public record ConsultRunbookRequest(String serviceName) {}

    @PostMapping("/tools/consult-runbook")
    public ToolResponse consultRunbook(@RequestBody ConsultRunbookRequest req) {
        return new ToolResponse(service.consultRunbook(req.serviceName()));
    }
}
