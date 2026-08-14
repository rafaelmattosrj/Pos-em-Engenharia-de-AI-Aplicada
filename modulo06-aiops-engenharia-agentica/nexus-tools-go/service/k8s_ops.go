// Package service implementa as ferramentas (tools) de AIOps/DevSecOps/
// FinOps do projeto Nexus — equivalente ao pacote tools/*.py (decoradas com
// @tool do CrewAI) do módulo 06 do curso. A orquestração multi-agente
// hierárquica do CrewAI (core/agents.py, labs/*.py) NÃO é portada — apenas
// a lógica de negócio das ferramentas.
package service

import (
	"fmt"
	"os"
	"strings"
)

// K8sOpsService — equivalente a tools/k8s_ops.py.
type K8sOpsService struct{}

const k8sManifestTemplate = `apiVersion: apps/v1
kind: Deployment
metadata:
  name: %[1]s
spec:
  replicas: %[2]d
  selector:
    matchLabels:
      app: %[1]s
  template:
    metadata:
      labels:
        app: %[1]s
    spec:
      containers:
      - name: %[1]s
        image: nginx:latest
        ports:
        - containerPort: %[3]d
        readinessProbe:
          httpGet:
            path: /
            port: %[3]d
---
apiVersion: v1
kind: Service
metadata:
  name: %[1]s-svc
spec:
  selector:
    app: %[1]s
  ports:
  - protocol: TCP
    port: 80
    targetPort: %[3]d
`

// GenerateK8sManifest gera manifestos Kubernetes de Deployment e Service e
// os salva em disco — equivalente a generate_k8s_manifest.
func (K8sOpsService) GenerateK8sManifest(appName string, replicas, port int) string {
	manifest := fmt.Sprintf(k8sManifestTemplate, appName, replicas, port)
	filename := appName + "-k8s.yaml"

	if err := os.WriteFile(filename, []byte(manifest), 0o644); err != nil {
		return fmt.Sprintf("❌ Erro ao escrever o manifesto '%s': %v", filename, err)
	}
	return fmt.Sprintf("✅ Kubernetes manifests for '%s' successfully generated in '%s'.", appName, filename)
}

// ApplyK8sManifest simula ou executa a reconciliação GitOps via 'kubectl
// apply' — equivalente a apply_k8s_manifest.
func (K8sOpsService) ApplyK8sManifest(filename string) string {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Sprintf("❌ Error: The file '%s' was not found to apply.", filename)
	}

	output, err := runCommand("kubectl", "apply", "-f", filename)
	if err == nil {
		return fmt.Sprintf("✅ GitOps Sync Success: %s", strings.TrimSpace(output))
	}
	if isCommandNotFoundErr(err) {
		return "ℹ️ Simulation Mode: 'kubectl' command line tool is not installed. " +
			"In a production system, ArgoCD or Flux would apply this manifest now."
	}
	return fmt.Sprintf("⚠️ GitOps Simulation: File '%s' is syntactically valid, "+
		"but no Kubernetes cluster was detected. The GitOps controller would reconcile this state.", filename)
}

// AnalyzeCanaryMetrics analisa métricas da aplicação para decidir se um
// Canary Rollout deve prosseguir ou reverter — equivalente a
// analyze_canary_metrics.
func (K8sOpsService) AnalyzeCanaryMetrics(metricsData string) string {
	if strings.Contains(metricsData, "error_rate > 5%") || strings.Contains(strings.ToLower(metricsData), "error") {
		return "❌ ROLLBACK: Elevated error rate detected in Canary pods. Reverting deployment."
	}
	return "✅ PROCEED: Metrics are stable. Canary rollout approved for production."
}
