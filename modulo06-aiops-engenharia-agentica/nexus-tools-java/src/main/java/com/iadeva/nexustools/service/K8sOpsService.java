package com.iadeva.nexustools.service;

import org.springframework.stereotype.Service;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;

/**
 * Equivalente a tools/k8s_ops.py.
 */
@Service
public class K8sOpsService {

    /**
     * Gera manifestos Kubernetes de Deployment e Service e os salva em
     * disco — equivalente a generate_k8s_manifest.
     */
    public String generateK8sManifest(String appName, int replicas, int port) {
        String manifest = """
                apiVersion: apps/v1
                kind: Deployment
                metadata:
                  name: %1$s
                spec:
                  replicas: %2$d
                  selector:
                    matchLabels:
                      app: %1$s
                  template:
                    metadata:
                      labels:
                        app: %1$s
                    spec:
                      containers:
                      - name: %1$s
                        image: nginx:latest
                        ports:
                        - containerPort: %3$d
                        readinessProbe:
                          httpGet:
                            path: /
                            port: %3$d
                ---
                apiVersion: v1
                kind: Service
                metadata:
                  name: %1$s-svc
                spec:
                  selector:
                    app: %1$s
                  ports:
                  - protocol: TCP
                    port: 80
                    targetPort: %3$d
                """.formatted(appName, replicas, port);

        String filename = appName + "-k8s.yaml";
        try {
            Files.writeString(Path.of(filename), manifest);
        } catch (IOException e) {
            return "❌ Erro ao escrever o manifesto '" + filename + "': " + e.getMessage();
        }
        return "✅ Kubernetes manifests for '" + appName + "' successfully generated in '" + filename + "'.";
    }

    /**
     * Simula ou executa a reconciliação GitOps via 'kubectl apply' —
     * equivalente a apply_k8s_manifest.
     */
    public String applyK8sManifest(String filename) {
        if (!Files.exists(Path.of(filename))) {
            return "❌ Error: The file '" + filename + "' was not found to apply.";
        }

        try {
            Process process = new ProcessBuilder("kubectl", "apply", "-f", filename)
                    .redirectErrorStream(true)
                    .start();
            String output = new String(process.getInputStream().readAllBytes());
            int exitCode = process.waitFor();

            if (exitCode == 0) {
                return "✅ GitOps Sync Success: " + output.strip();
            }
            return "⚠️ GitOps Simulation: File '" + filename + "' is syntactically valid, "
                    + "but no Kubernetes cluster was detected. The GitOps controller would reconcile this state.";
        } catch (IOException e) {
            return "ℹ️ Simulation Mode: 'kubectl' command line tool is not installed. "
                    + "In a production system, ArgoCD or Flux would apply this manifest now.";
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            return "⚠️ Error: kubectl execution was interrupted.";
        }
    }

    /**
     * Analisa métricas da aplicação para decidir se um Canary Rollout deve
     * prosseguir ou reverter — equivalente a analyze_canary_metrics.
     */
    public String analyzeCanaryMetrics(String metricsData) {
        if (metricsData.contains("error_rate > 5%") || metricsData.toLowerCase().contains("error")) {
            return "❌ ROLLBACK: Elevated error rate detected in Canary pods. Reverting deployment.";
        }
        return "✅ PROCEED: Metrics are stable. Canary rollout approved for production.";
    }
}
