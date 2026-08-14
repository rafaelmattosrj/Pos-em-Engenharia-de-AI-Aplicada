// Package tool implementa as tools disponíveis para o agente — equivalente
// ao pacote tool/ da versão Java (AgentTool + 4 implementações simuladas).
package tool

// AgentTool é o contrato que todas as tools do agente devem implementar,
// permitindo que o Executor despache chamadas de forma genérica pelo nome.
type AgentTool interface {
	Name() string
	Execute(args map[string]any) string
}

// stringArg lê um argumento string com valor default caso ausente ou de
// outro tipo — equivalente a args.getOrDefault(key, default) em Java.
func stringArg(args map[string]any, key, defaultValue string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return defaultValue
}

// intArg lê um argumento numérico com valor default. JSON decodifica
// números como float64, então aceita tanto float64 quanto int.
func intArg(args map[string]any, key string, defaultValue int) int {
	switch v := args[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return defaultValue
	}
}
