// Package team é o porte da camada multiagente introduzida na Unidade 9
// (09-multi-agent-systems/src/team/): um supervisor decide, a cada rodada,
// qual papel (analista/planejador/executor) age a seguir, cada papel
// contribui uma entrada no "blackboard" compartilhado, até a decisão "done"
// ou o teto de handoffs. Go não tem um LangGraph maduro (StateGraph do
// original) -- aqui a máquina de estado é explícita: TeamGraph guarda o
// estado da rodada em variáveis locais e usa uma função de transição
// (TeamGraph.Run) em vez de nós/arestas declarativos; cada papel é uma struct
// com método Run, orquestrada pela função coordenadora.
package team

import (
	"fmt"
	"strings"
)

// Role é o porte do union type TeamRole (team/supervisor.ts).
type Role string

const (
	RoleAnalista   Role = "analista"
	RolePlanejador Role = "planejador"
	RoleExecutor   Role = "executor"
)

// BlackboardKind é o porte do union type BlackboardKind (team/blackboard.ts).
type BlackboardKind string

const (
	KindFacts     BlackboardKind = "facts"
	KindPlan      BlackboardKind = "plan"
	KindExecution BlackboardKind = "execution"
	KindError     BlackboardKind = "error"
)

// BlackboardEntry é o porte de BlackboardEntry (team/blackboard.ts).
type BlackboardEntry struct {
	Role    Role
	Kind    BlackboardKind
	Brief   string
	Content string
}

// RenderBlackboard é o porte de renderBlackboard (team/blackboard.ts).
func RenderBlackboard(entries []BlackboardEntry) string {
	if len(entries) == 0 {
		return "(blackboard vazio -- nenhuma contribuicao ainda)"
	}
	parts := make([]string, 0, len(entries))
	for i, entry := range entries {
		parts = append(parts, fmt.Sprintf("[%d] %s (%s) -- brief: %s\n%s", i+1, entry.Role, entry.Kind, entry.Brief, entry.Content))
	}
	return strings.Join(parts, "\n\n")
}
