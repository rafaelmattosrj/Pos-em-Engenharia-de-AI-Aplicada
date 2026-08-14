package tool

import "encoding/json"

// GetLogsTool retorna as últimas linhas de log simuladas de um serviço —
// equivalente a GetLogsTool.java. Em produção integraria com ELK Stack ou
// CloudWatch.
type GetLogsTool struct{}

func (GetLogsTool) Name() string { return "getLogs" }

func (GetLogsTool) Execute(args map[string]any) string {
	service := stringArg(args, "service", "api-gateway")
	lines := intArg(args, "lines", 10)

	allLogs := []map[string]string{
		{"timestamp": "2024-01-15T14:29:55Z", "severity": "ERROR", "service": service,
			"message": "Connection pool exhausted: timeout após 5000ms aguardando conexão disponível"},
		{"timestamp": "2024-01-15T14:29:58Z", "severity": "ERROR", "service": service,
			"message": "Database query timeout: SELECT * FROM orders WHERE status='pending' excedeu 3000ms"},
		{"timestamp": "2024-01-15T14:30:01Z", "severity": "WARN", "service": service,
			"message": "Retry attempt 3/3 para endpoint /api/payments — falha no upstream"},
		{"timestamp": "2024-01-15T14:30:05Z", "severity": "ERROR", "service": service,
			"message": "Unhandled exception: NullPointerException em OrderService.processPayment():142"},
		{"timestamp": "2024-01-15T14:30:10Z", "severity": "INFO", "service": service,
			"message": "Deploy v2.3.1 detectado às 14:15 — possível correlação com aumento de erros"},
	}

	if lines < len(allLogs) {
		allLogs = allLogs[:lines]
	}

	out, err := json.Marshal(map[string]any{
		"service":             service,
		"total_logs_returned": len(allLogs),
		"logs":                allLogs,
	})
	if err != nil {
		return `{"error": "Falha ao serializar logs"}`
	}
	return string(out)
}
