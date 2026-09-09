package team

import (
	"encoding/json"
	"fmt"
	"strings"

	"opspilot/llm"
)

// SupervisorSystemPrompt é o porte de SUPERVISOR_SYSTEM_PROMPT
// (team/supervisor-prompt.ts).
const SupervisorSystemPrompt = "Voce e o SUPERVISOR do plantao de incidentes. Decida qual papel age a seguir:\n" +
	"analista (diagnostico factual, so leitura), planejador (transforma fatos em plano,\n" +
	"sem ferramentas) ou executor (executa acoes de incidente). Responda 'done' quando o\n" +
	"pedido do plantonista estiver atendido.\n" +
	"Responda SOMENTE com um JSON no formato {\"next\": \"analista|planejador|executor|done\", \"brief\": \"...\"}."

// SupervisorDecision é o porte de SupervisorDecision (team/supervisor.ts).
// Done == true equivale a next == null no porte Java / next == "done" no
// original TS.
type SupervisorDecision struct {
	Next  Role
	Done  bool
	Brief string
}

// DoneDecision é o porte de SupervisorDecision.done(brief) (porte Java).
func DoneDecision(brief string) SupervisorDecision {
	return SupervisorDecision{Done: true, Brief: brief}
}

// DecideNextFn é o porte do tipo DecideNextFn (team/supervisor.ts).
type DecideNextFn func(message string, blackboard []BlackboardEntry, handoffCount int) (SupervisorDecision, error)

// CreateDecideNext é o porte de createDecideNext (team/supervisor.ts): pede
// ao modelo uma decisão estruturada {next, brief} via prompt (equivalente
// idiomático ao withStructuredOutput do LangChain, que não tem par direto
// fora do ecossistema JS/Python) e faz o parse manual da resposta.
func CreateDecideNext(model llm.ChatModel) DecideNextFn {
	return func(message string, blackboard []BlackboardEntry, handoffCount int) (SupervisorDecision, error) {
		userMessage := fmt.Sprintf(
			"Pedido do plantonista: %s\nDelegacoes ja usadas: %d\n\nBlackboard:\n%s",
			message, handoffCount, RenderBlackboard(blackboard))

		response, err := model.Invoke([]llm.ChatMessage{
			llm.SystemMessage(SupervisorSystemPrompt),
			llm.UserMessage(userMessage),
		}, nil)
		if err != nil {
			return SupervisorDecision{}, err
		}

		return ParseDecision(response.Content)
	}
}

// ParseDecision é o porte de Supervisor.parseDecision (porte Java) --
// exportado para ser testável diretamente, igual ao equivalente Java (que
// precisou ser tornado público pela mesma razão).
func ParseDecision(rawJSON string) (SupervisorDecision, error) {
	var node map[string]any
	if err := json.Unmarshal([]byte(rawJSON), &node); err != nil {
		return SupervisorDecision{}, fmt.Errorf("decisao invalida do supervisor: JSON malformado: %w", err)
	}

	nextRaw, hasNext := node["next"]
	briefRaw, hasBrief := node["brief"]
	if !hasNext || !hasBrief {
		return SupervisorDecision{}, fmt.Errorf("decisao invalida do supervisor: campos 'next'/'brief' ausentes")
	}

	next := fmt.Sprintf("%v", nextRaw)
	brief := fmt.Sprintf("%v", briefRaw)

	if next == "done" {
		return DoneDecision(brief), nil
	}

	role, ok := parseRole(next)
	if !ok {
		return SupervisorDecision{}, fmt.Errorf("decisao invalida do supervisor: papel desconhecido %q", next)
	}
	return SupervisorDecision{Next: role, Brief: brief}, nil
}

func parseRole(raw string) (Role, bool) {
	switch strings.ToLower(raw) {
	case string(RoleAnalista):
		return RoleAnalista, true
	case string(RolePlanejador):
		return RolePlanejador, true
	case string(RoleExecutor):
		return RoleExecutor, true
	default:
		return "", false
	}
}
