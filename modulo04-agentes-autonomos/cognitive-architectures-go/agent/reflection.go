package agent

import (
	"context"
	"fmt"

	"cognitive-architectures/model"
)

const initialGenerationPromptTemplate = `Responda à seguinte tarefa de forma completa, precisa e bem estruturada:

Tarefa: %s
`

const regenerationPromptTemplate = `Você recebeu feedback sobre sua resposta anterior. Melhore-a incorporando as sugestões.

Tarefa original: %s

Sua resposta anterior:
%s

Feedback do avaliador (score: %.2f/1.0):
- Correção: %.2f
- Completude: %.2f
- Qualidade: %.2f
- Observações: %s

Gere uma resposta MELHORADA que corrija os pontos levantados:
`

// ReflectionAgent implementa a arquitetura Reflection — equivalente a
// ReflectionAgent.java (aula 08). Melhora iterativamente via auto-avaliação:
// gera output, critica, regenera com feedback, até o CritiqueEvaluator
// aprovar ou atingir MaxCycles.
type ReflectionAgent struct {
	Client    ChatClient
	Critique  *CritiqueEvaluator
	MaxCycles int
}

// Execute executa o ciclo de reflexão para input.
func (a *ReflectionAgent) Execute(ctx context.Context, input string) (model.AgentResponse, error) {
	maxCycles := a.MaxCycles
	if maxCycles == 0 {
		maxCycles = 3
	}

	var currentOutput string
	var lastCritique model.CritiqueResult
	hasCritique := false
	reflectionCount := 0
	estimatedTokens := 0

	for cycle := 1; cycle <= maxCycles; cycle++ {
		prompt := buildGenerationPrompt(input, currentOutput, lastCritique, hasCritique)

		output, err := a.Client.Chat(ctx, prompt)
		if err != nil {
			return model.AgentResponse{}, err
		}
		estimatedTokens += (len(prompt) + len(output)) / 4
		currentOutput = output

		critique, err := a.Critique.Evaluate(ctx, input, currentOutput)
		if err != nil {
			return model.AgentResponse{}, err
		}
		lastCritique = critique
		hasCritique = true

		if critique.Passed {
			break
		}

		if cycle < maxCycles {
			reflectionCount++
		}
	}

	finalResult := formatResult(currentOutput, lastCritique, reflectionCount)

	return model.AgentResponse{
		Result:  finalResult,
		Metrics: model.AgentMetrics{Steps: reflectionCount + 1, Tokens: estimatedTokens, Reflections: reflectionCount},
	}, nil
}

// buildGenerationPrompt monta o prompt de geração. Na primeira iteração
// (sem output/crítica anteriores): geração inicial. Nas iterações
// seguintes: inclui o output anterior e o feedback do crítico.
func buildGenerationPrompt(task, previousOutput string, critique model.CritiqueResult, hasCritique bool) string {
	if !hasCritique {
		return fmt.Sprintf(initialGenerationPromptTemplate, task)
	}
	return fmt.Sprintf(regenerationPromptTemplate,
		task, previousOutput, critique.Score, critique.Correctness, critique.Completeness, critique.Quality, critique.Feedback)
}

// formatResult formata o resultado final com metadados de qualidade para
// transparência.
func formatResult(output string, critique model.CritiqueResult, reflections int) string {
	result := output + "\n\n---\n"
	result += fmt.Sprintf("*Qualidade: %.0f%%", critique.Score*100)
	if reflections > 0 {
		result += fmt.Sprintf(" após %d reflexão(ões)", reflections)
	}
	if critique.Passed {
		result += " | Aprovado*"
	} else {
		result += " | Melhor disponível*"
	}
	return result
}
