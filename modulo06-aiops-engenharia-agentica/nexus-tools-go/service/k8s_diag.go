package service

import (
	"fmt"
	"strings"
)

// K8sDiagService — equivalente a tools/k8s_diag.py.
type K8sDiagService struct{}

var remediations = map[string]string{
	"OOMKilled":        "Increase 'resources.limits.memory' to 1Gi in the Deployment spec.",
	"ImagePullBackOff": "Correct the image tag to a valid version or 'latest' in ECR/DockerHub.",
	"CrashLoopBackOff": "Check for missing environment variables (e.g., DB_URL) or Kubernetes Secrets.",
}

// InspectPodFailure analisa logs e eventos de um Pod para diagnosticar
// CrashLoopBackOff ou OOMKilled — equivalente a inspect_pod_failure.
func (K8sDiagService) InspectPodFailure(podName string) string {
	lower := strings.ToLower(podName)
	switch {
	case strings.Contains(lower, "api"):
		return "\n" +
			"EVENTS: \n" +
			"- Warning  BackOff  Back-off restarting failed container\n" +
			"LOGS:\n" +
			"- Error: Cannot connect to database at 10.0.1.5:5432\n" +
			"DIAGNOSIS: Database connectivity failure (Network/Config).\n"
	case strings.Contains(lower, "worker"):
		return "STATUS: Terminated | REASON: OOMKilled | MEMORY_USAGE: 512Mi (Limit: 512Mi)."
	default:
		return fmt.Sprintf("Logs for Pod '%s' look normal, but the Readiness Probe is failing.", podName)
	}
}

// SuggestFix sugere a resolução técnica apropriada no manifesto Kubernetes
// com base no tipo de problema — equivalente a suggest_fix.
func (K8sDiagService) SuggestFix(issueType string) string {
	if fix, ok := remediations[issueType]; ok {
		return fix
	}
	return "Review the Readiness Probe and Liveness Probe configurations in the manifest."
}
