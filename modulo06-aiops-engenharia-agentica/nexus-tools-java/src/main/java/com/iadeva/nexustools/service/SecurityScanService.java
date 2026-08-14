package com.iadeva.nexustools.service;

import org.springframework.stereotype.Service;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;

/**
 * Equivalente a tools/security_scan.py.
 */
@Service
public class SecurityScanService {

    /**
     * Executa o scanner de segurança estática Checkov sobre um arquivo de
     * infraestrutura — equivalente a run_checkov_scan.
     */
    public String runCheckovScan(String filename) {
        String target = filename == null || filename.isBlank() ? "main.tf" : filename;

        if (!Files.exists(Path.of(target))) {
            return "❌ Error: File '" + target + "' not found for scanning.";
        }

        try {
            Process process = new ProcessBuilder("checkov", "-f", target, "--quiet", "--compact", "--no-guide")
                    .redirectErrorStream(true)
                    .start();
            String output = new String(process.getInputStream().readAllBytes());
            int exitCode = process.waitFor();

            if (exitCode != 0 || output.contains("FAILED")) {
                return "❌ Security Failures Detected by Checkov:\n" + output.strip();
            }
            return "✅ Checkov: No vulnerabilities detected. Infrastructure code is secure.";
        } catch (IOException e) {
            return "⚠️ Error: 'checkov' command-line tool not found. Run 'pip install checkov' in the terminal.";
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            return "⚠️ Unexpected error running scanner: execução interrompida.";
        }
    }

    /**
     * Simula o motor de decisão de políticas do Open Policy Agent (OPA) —
     * equivalente a validate_opa_policies.
     */
    public String validateOpaPolicies(String content) {
        String contentLower = content.toLowerCase();

        if (!contentLower.contains("us-east-1")) {
            return "❌ OPA REJECTED: Violation of rule 'SOBERANIA_DADOS'. Nexus resources must reside in us-east-1.";
        }
        if (contentLower.contains("t3.large")) {
            return "❌ OPA REJECTED: Violation of rule 'COST_CONTROL'. Large instance sizes require manual finance approval.";
        }
        if (content.contains("0.0.0.0/0")) {
            return "❌ OPA REJECTED: Violation of rule 'NO_PUBLIC_INGRESS'. Open ingress CIDR ranges are strictly forbidden.";
        }
        return "✅ OPA PASSED: Infrastructure code complies with Nexus governance policies.";
    }
}
