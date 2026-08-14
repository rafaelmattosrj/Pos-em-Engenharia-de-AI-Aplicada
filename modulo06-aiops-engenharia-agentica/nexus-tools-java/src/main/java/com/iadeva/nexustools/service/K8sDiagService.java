package com.iadeva.nexustools.service;

import org.springframework.stereotype.Service;

import java.util.Map;

/**
 * Equivalente a tools/k8s_diag.py.
 */
@Service
public class K8sDiagService {

    private static final Map<String, String> REMEDIATIONS = Map.of(
            "OOMKilled", "Increase 'resources.limits.memory' to 1Gi in the Deployment spec.",
            "ImagePullBackOff", "Correct the image tag to a valid version or 'latest' in ECR/DockerHub.",
            "CrashLoopBackOff", "Check for missing environment variables (e.g., DB_URL) or Kubernetes Secrets."
    );

    /**
     * Analisa logs e eventos de um Pod para diagnosticar CrashLoopBackOff ou
     * OOMKilled — equivalente a inspect_pod_failure.
     */
    public String inspectPodFailure(String podName) {
        String lower = podName.toLowerCase();
        if (lower.contains("api")) {
            return """

                    EVENTS:
                    - Warning  BackOff  Back-off restarting failed container
                    LOGS:
                    - Error: Cannot connect to database at 10.0.1.5:5432
                    DIAGNOSIS: Database connectivity failure (Network/Config).
                    """;
        }
        if (lower.contains("worker")) {
            return "STATUS: Terminated | REASON: OOMKilled | MEMORY_USAGE: 512Mi (Limit: 512Mi).";
        }
        return "Logs for Pod '" + podName + "' look normal, but the Readiness Probe is failing.";
    }

    /**
     * Sugere a resolução técnica apropriada no manifesto Kubernetes com
     * base no tipo de problema — equivalente a suggest_fix.
     */
    public String suggestFix(String issueType) {
        return REMEDIATIONS.getOrDefault(issueType,
                "Review the Readiness Probe and Liveness Probe configurations in the manifest.");
    }
}
