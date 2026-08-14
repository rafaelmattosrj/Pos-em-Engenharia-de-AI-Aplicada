// Package core implementa o agent loop (Percepção→Planejamento→Ação→Avaliação)
// — equivalente aos pacotes core/ e observability/ da versão Java.
package core

import "fmt"

// AgentContract define os limites e a identidade do agente — equivalente a
// AgentContract.java.
type AgentContract struct {
	MaxSteps        int
	MaxTokens       int
	MaxTimeSeconds  int
	NoProgressSteps int
	AgentName       string
	Role            string
}

// NewAgentContract cria um AgentContract com os defaults do curso: 10 steps,
// 50000 tokens, 120s de timeout, 3 steps sem progresso.
func NewAgentContract() AgentContract {
	return AgentContract{
		MaxSteps:        10,
		MaxTokens:       50000,
		MaxTimeSeconds:  120,
		NoProgressSteps: 3,
		AgentName:       "DevOps Agent",
		Role:            "Especialista em diagnóstico e resolução de incidentes",
	}
}

// SystemPrompt retorna o system prompt base do agente, derivado do contrato.
func (c AgentContract) SystemPrompt() string {
	return fmt.Sprintf(`Você é o %s.
Papel: %s

Regras:
- Sempre responda em JSON válido, sem markdown ou blocos de código
- Seja objetivo e direto nas ações
- Prefira ferramentas de diagnóstico antes de propor soluções
- Máximo de %d steps disponíveis — use-os com sabedoria
`, c.AgentName, c.Role, c.MaxSteps)
}
