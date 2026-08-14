package com.iadeva.nexustools.service;

import org.springframework.stereotype.Service;

/**
 * Equivalente a tools/chatops_tools.py.
 */
@Service
public class ChatOpsService {

    private static final String MANAGER_APPROVAL_PASSWORD = "GESTOR-APROVA";

    /**
     * Aplica mudanças de infraestrutura via Terraform, exigindo aprovação
     * humana (human-in-the-loop) para operações destrutivas — equivalente
     * a execute_terraform.
     */
    public String executeTerraform(String command, String managerPassword) {
        String lower = command.toLowerCase();
        boolean isDestructive = lower.contains("destruir") || lower.contains("apagar") || lower.contains("destroy");

        if (isDestructive) {
            if (!MANAGER_APPROVAL_PASSWORD.equals(managerPassword)) {
                return "🛑 BLOCKED: Critical action detected! Provide the correct manager_password to proceed.";
            }
            return "✅ APPROVED: Human-in-the-loop validated. Terraform executed successfully.";
        }

        return "✅ SUCCESS: The command '" + command + "' was executed (Low impact).";
    }
}
