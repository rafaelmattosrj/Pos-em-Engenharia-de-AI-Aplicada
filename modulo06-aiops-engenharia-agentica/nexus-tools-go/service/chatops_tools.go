package service

import (
	"fmt"
	"strings"
)

// ChatOpsService — equivalente a tools/chatops_tools.py.
type ChatOpsService struct{}

const managerApprovalPassword = "GESTOR-APROVA"

// ExecuteTerraform aplica mudanças de infraestrutura via Terraform,
// exigindo aprovação humana (human-in-the-loop) para operações destrutivas
// — equivalente a execute_terraform.
func (ChatOpsService) ExecuteTerraform(command, managerPassword string) string {
	lower := strings.ToLower(command)
	isDestructive := strings.Contains(lower, "destruir") || strings.Contains(lower, "apagar") || strings.Contains(lower, "destroy")

	if isDestructive {
		if managerPassword != managerApprovalPassword {
			return "🛑 BLOCKED: Critical action detected! Provide the correct manager_password to proceed."
		}
		return "✅ APPROVED: Human-in-the-loop validated. Terraform executed successfully."
	}

	return fmt.Sprintf("✅ SUCCESS: The command '%s' was executed (Low impact).", command)
}
