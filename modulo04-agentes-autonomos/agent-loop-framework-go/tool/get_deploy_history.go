package tool

import "encoding/json"

// GetDeployHistoryTool retorna o histórico de deploys recentes simulado de
// um serviço — equivalente a GetDeployHistoryTool.java. Em produção
// integraria com o pipeline CI/CD (Jenkins, GitHub Actions, ArgoCD).
type GetDeployHistoryTool struct{}

func (GetDeployHistoryTool) Name() string { return "getDeployHistory" }

func (GetDeployHistoryTool) Execute(args map[string]any) string {
	service := stringArg(args, "service", "api-gateway")

	deploys := []map[string]any{
		{
			"id": "deploy-20240115-001", "timestamp": "2024-01-15T14:15:00Z", "service": service,
			"version": "v2.3.1", "previous_version": "v2.3.0", "status": "SUCCESS",
			"deployed_by":        "ci-pipeline",
			"changes":            "Refatoração do connection pool e aumento de timeout para queries de pagamento",
			"rollback_available": true,
		},
		{
			"id": "deploy-20240114-003", "timestamp": "2024-01-14T22:30:00Z", "service": service,
			"version": "v2.3.0", "previous_version": "v2.2.9", "status": "SUCCESS",
			"deployed_by":        "rafael.mattos",
			"changes":            "Correção de bug no fluxo de autenticação OAuth",
			"rollback_available": true,
		},
		{
			"id": "deploy-20240113-002", "timestamp": "2024-01-13T10:00:00Z", "service": service,
			"version": "v2.2.9", "previous_version": "v2.2.8", "status": "ROLLED_BACK",
			"deployed_by":        "ci-pipeline",
			"changes":            "Migração de dependência do driver JDBC",
			"rollback_available": false,
			"rollback_reason":    "Aumento de 300% na latência do banco após deploy",
		},
	}

	out, err := json.Marshal(map[string]any{
		"service":       service,
		"total_deploys": len(deploys),
		"last_deploy":   deploys[0]["timestamp"],
		"deploys":       deploys,
	})
	if err != nil {
		return `{"error": "Falha ao serializar histórico de deploys"}`
	}
	return string(out)
}
