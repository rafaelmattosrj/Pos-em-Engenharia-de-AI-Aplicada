package service

import (
	"fmt"
	"os"
	"strings"
)

// SecurityScanService — equivalente a tools/security_scan.py.
type SecurityScanService struct{}

// RunCheckovScan executa o scanner de segurança estática Checkov sobre um
// arquivo de infraestrutura — equivalente a run_checkov_scan.
func (SecurityScanService) RunCheckovScan(filename string) string {
	target := filename
	if target == "" {
		target = "main.tf"
	}

	if _, err := os.Stat(target); os.IsNotExist(err) {
		return fmt.Sprintf("❌ Error: File '%s' not found for scanning.", target)
	}

	output, err := runCommand("checkov", "-f", target, "--quiet", "--compact", "--no-guide")
	if err != nil && isCommandNotFoundErr(err) {
		return "⚠️ Error: 'checkov' command-line tool not found. Run 'pip install checkov' in the terminal."
	}
	if err != nil || strings.Contains(output, "FAILED") {
		return fmt.Sprintf("❌ Security Failures Detected by Checkov:\n%s", strings.TrimSpace(output))
	}
	return "✅ Checkov: No vulnerabilities detected. Infrastructure code is secure."
}

// ValidateOpaPolicies simula o motor de decisão de políticas do Open Policy
// Agent (OPA) — equivalente a validate_opa_policies.
func (SecurityScanService) ValidateOpaPolicies(content string) string {
	lower := strings.ToLower(content)

	if !strings.Contains(lower, "us-east-1") {
		return "❌ OPA REJECTED: Violation of rule 'SOBERANIA_DADOS'. Nexus resources must reside in us-east-1."
	}
	if strings.Contains(lower, "t3.large") {
		return "❌ OPA REJECTED: Violation of rule 'COST_CONTROL'. Large instance sizes require manual finance approval."
	}
	if strings.Contains(content, "0.0.0.0/0") {
		return "❌ OPA REJECTED: Violation of rule 'NO_PUBLIC_INGRESS'. Open ingress CIDR ranges are strictly forbidden."
	}
	return "✅ OPA PASSED: Infrastructure code complies with Nexus governance policies."
}
