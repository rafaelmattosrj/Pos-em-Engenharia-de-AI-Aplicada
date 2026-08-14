package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"cognitive-architectures/model"
)

const planningPromptTemplate = `Você é um agente planejador. Analise a tarefa abaixo e crie um plano de execução detalhado.

Retorne APENAS um JSON válido (sem markdown, sem explicações) com a seguinte estrutura:
{
  "steps": [
    {
      "stepNumber": 1,
      "description": "descrição do que será feito",
      "tool": "nome_da_ferramenta",
      "args": {"param1": "valor1"},
      "successCriteria": "critério para considerar o step concluído"
    }
  ]
}

Ferramentas disponíveis: search, calculate, summarize, validate, format, store

Tarefa: %s
`

// PlanExecuteAgent implementa a arquitetura Plan-and-Execute — equivalente a
// PlanExecuteAgent.java (aula 08). Planeja tudo primeiro com uma única
// chamada ao LLM, depois executa deterministicamente sem consultar o LLM
// novamente.
type PlanExecuteAgent struct {
	Client ChatClient
}

// Execute executa a arquitetura Plan-and-Execute para input.
func (a *PlanExecuteAgent) Execute(ctx context.Context, input string) (model.AgentResponse, error) {
	// Fase 1: Planejamento — única chamada ao LLM.
	planPrompt := fmt.Sprintf(planningPromptTemplate, input)
	planJSON, err := a.Client.Chat(ctx, planPrompt)
	if err != nil {
		return model.AgentResponse{}, err
	}
	estimatedTokens := (len(planPrompt) + len(planJSON)) / 4

	plan := parsePlan(planJSON)

	// Fase 2: Execução determinística — sem mais chamadas ao LLM.
	var log strings.Builder
	log.WriteString("# Resultado da Execução do Plano\n\n")
	log.WriteString("**Tarefa:** " + input + "\n\n")

	for _, step := range plan.Steps {
		stepResult := simulateToolExecution(step)
		log.WriteString(fmt.Sprintf("## Step %d: %s\n", step.StepNumber, step.Description))
		log.WriteString("- **Ferramenta:** " + step.Tool + "\n")
		log.WriteString("- **Resultado:** " + stepResult + "\n")
		log.WriteString("- **Critério de aceite:** " + step.SuccessCriteria + "\n\n")
	}

	log.WriteString("## Conclusão\n")
	log.WriteString(fmt.Sprintf("Todos os %d steps do plano foram executados com sucesso.", len(plan.Steps)))

	return model.AgentResponse{
		Result:  log.String(),
		Metrics: model.AgentMetrics{Steps: len(plan.Steps), Tokens: estimatedTokens, Reflections: 0},
	}, nil
}

// parsePlan faz o parse do JSON retornado pelo LLM para ExecutionPlan. Em
// caso de falha, retorna um plano de fallback com step genérico.
func parsePlan(planJSON string) model.ExecutionPlan {
	var raw struct {
		Steps []struct {
			StepNumber      int            `json:"stepNumber"`
			Description     string         `json:"description"`
			Tool            string         `json:"tool"`
			Args            map[string]any `json:"args"`
			SuccessCriteria string         `json:"successCriteria"`
		} `json:"steps"`
	}

	if err := json.Unmarshal([]byte(extractJSON(planJSON)), &raw); err != nil || len(raw.Steps) == 0 {
		return model.ExecutionPlan{Steps: []model.PlanStep{{
			StepNumber:      1,
			Description:     "Executar tarefa diretamente",
			Tool:            "generic",
			Args:            map[string]any{"task": "executar"},
			SuccessCriteria: "tarefa concluída",
		}}}
	}

	steps := make([]model.PlanStep, 0, len(raw.Steps))
	for _, s := range raw.Steps {
		stepNumber := s.StepNumber
		if stepNumber == 0 {
			stepNumber = 1
		}
		description := s.Description
		if description == "" {
			description = "step sem descrição"
		}
		tool := s.Tool
		if tool == "" {
			tool = "generic"
		}
		successCriteria := s.SuccessCriteria
		if successCriteria == "" {
			successCriteria = "step concluído"
		}
		args := s.Args
		if args == nil {
			args = map[string]any{}
		}

		steps = append(steps, model.PlanStep{
			StepNumber:      stepNumber,
			Description:     description,
			Tool:            tool,
			Args:            args,
			SuccessCriteria: successCriteria,
		})
	}

	return model.ExecutionPlan{Steps: steps}
}

// extractJSON extrai o bloco JSON de uma string que pode conter texto extra
// do LLM.
func extractJSON(text string) string {
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start != -1 && end != -1 && end > start {
		return text[start : end+1]
	}
	return text
}

// simulateToolExecution simula a execução de um step de ferramenta sem
// chamar APIs reais. Em produção, aqui seria feito o dispatch para
// ferramentas reais via function calling.
func simulateToolExecution(step model.PlanStep) string {
	switch step.Tool {
	case "search":
		return fmt.Sprintf("Resultados de busca obtidos para: %v", step.Args)
	case "calculate":
		return fmt.Sprintf("Cálculo executado com resultado numérico baseado em: %v", step.Args)
	case "summarize":
		return "Resumo gerado com sucesso para o conteúdo fornecido."
	case "validate":
		return "Validação concluída — dados consistentes."
	case "format":
		return "Conteúdo formatado conforme template especificado."
	case "store":
		return "Dados persistidos com sucesso."
	default:
		return fmt.Sprintf("Step '%s' executado (simulação).", step.Tool)
	}
}
