package com.iadeva.nexustools.controller;

import com.iadeva.nexustools.service.ChatOpsService;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

/**
 * Equivalente aos endpoints das tools de tools/chatops_tools.py.
 */
@RestController
public class ChatOpsController {

    private final ChatOpsService service;

    public ChatOpsController(ChatOpsService service) {
        this.service = service;
    }

    public record ExecuteTerraformRequest(String command, String managerPassword) {}

    @PostMapping("/tools/execute-terraform")
    public ToolResponse executeTerraform(@RequestBody ExecuteTerraformRequest req) {
        String password = req.managerPassword() != null ? req.managerPassword() : "None";
        return new ToolResponse(service.executeTerraform(req.command(), password));
    }
}
