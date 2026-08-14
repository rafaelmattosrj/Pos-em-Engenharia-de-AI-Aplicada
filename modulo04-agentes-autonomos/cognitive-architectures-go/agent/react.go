// Package agent implementa as 3 arquiteturas cognitivas (ReAct,
// Plan-and-Execute, Reflection) e o avaliador crítico usado pela Reflection
// — equivalente ao pacote agent/ da versão Java.
package agent

import (
	"context"
	"fmt"
	"strings"

	"cognitive-architectures/model"
)

// ChatClient é a única operação de que os agentes dependem.
type ChatClient interface {
	Chat(ctx context.Context, userPrompt string) (string, error)
}

const reactPromptTemplate = `Você é um agente ReAct. Resolva a tarefa usando o ciclo Thought → Action → Observation.

Regras:
- Thought: explique seu raciocínio antes de agir
- Action: indique a ação como "Action: <nome_da_acao>(<parametros>)"
  ou "Action: FINAL_ANSWER(<sua_resposta_completa>)" quando tiver a resposta
- Observation: descreva o resultado hipotético da ação

Tarefa: %s

Histórico até agora:
%s

Step %d — produza apenas o próximo Thought, Action e Observation:
`

// ReactAgent implementa a arquitetura ReAct (Reasoning + Acting) —
// equivalente a ReactAgent.java (aula 07). Cada step raciocina antes de
// agir; o loop para quando o LLM emite "Action: FINAL_ANSWER" ou atinge
// MaxSteps.
type ReactAgent struct {
	Client   ChatClient
	MaxSteps int
}

// Execute executa o loop ReAct para input.
func (a *ReactAgent) Execute(ctx context.Context, input string) (model.AgentResponse, error) {
	maxSteps := a.MaxSteps
	if maxSteps == 0 {
		maxSteps = 10
	}

	var history strings.Builder
	var lastAction string
	sameActionCount := 0
	stepCount := 0
	estimatedTokens := 0
	var finalAnswer string

	for step := 1; step <= maxSteps; step++ {
		stepCount = step

		prompt := fmt.Sprintf(reactPromptTemplate, input, historyOrPlaceholder(history.String()), step)

		response, err := a.Client.Chat(ctx, prompt)
		if err != nil {
			return model.AgentResponse{}, err
		}
		estimatedTokens += (len(prompt) + len(response)) / 4

		currentAction := extractField(response, "Action")

		history.WriteString(fmt.Sprintf("\n--- Step %d ---\n%s\n", step, response))

		if strings.HasPrefix(currentAction, "FINAL_ANSWER") {
			finalAnswer = extractFinalAnswer(currentAction)
			break
		}

		if currentAction != "" && currentAction == lastAction {
			sameActionCount++
			if sameActionCount >= 3 {
				finalAnswer = fmt.Sprintf("Não foi possível determinar uma resposta definitiva após %d iterações. Último raciocínio: %s",
					step, extractField(response, "Thought"))
				break
			}
		} else {
			sameActionCount = 0
			lastAction = currentAction
		}
	}

	if finalAnswer == "" {
		finalAnswer = fmt.Sprintf("Limite de %d steps atingido sem resposta final. Histórico parcial disponível.", maxSteps)
	}

	return model.AgentResponse{
		Result:  finalAnswer,
		Metrics: model.AgentMetrics{Steps: stepCount, Tokens: estimatedTokens, Reflections: 0},
	}, nil
}

func historyOrPlaceholder(history string) string {
	if history == "" {
		return "(nenhum)"
	}
	return history
}

// extractField extrai o valor de um campo (ex.: "Action:", "Thought:") da
// resposta do LLM.
func extractField(response, fieldName string) string {
	prefix := fieldName + ":"
	start := strings.Index(response, prefix)
	if start == -1 {
		return ""
	}
	start += len(prefix)
	rest := response[start:]
	if end := strings.Index(rest, "\n"); end != -1 {
		rest = rest[:end]
	}
	return strings.TrimSpace(rest)
}

// extractFinalAnswer extrai o texto de "FINAL_ANSWER(<texto>)".
func extractFinalAnswer(actionValue string) string {
	open := strings.Index(actionValue, "(")
	closeIdx := strings.LastIndex(actionValue, ")")
	if open != -1 && closeIdx != -1 && open < closeIdx {
		return strings.TrimSpace(actionValue[open+1 : closeIdx])
	}
	return strings.TrimSpace(strings.ReplaceAll(actionValue, "FINAL_ANSWER", ""))
}
