package evaluator

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"agent-evals/model"
)

// ChatClient é a única operação de que ToolSelectionEvaluator depende.
type ChatClient interface {
	Chat(ctx context.Context, userPrompt string) (string, error)
}

var availableTools = []string{"getMetrics", "getLogs", "getDeployHistory", "saveIncident", "notifyTeam"}

var toolNamePattern = regexp.MustCompile(`"tool"\s*:\s*"([^"]+)"`)

const toolSelectionPromptTemplate = `Você é um agente de monitoramento de sistemas. Dado o contexto abaixo, escolha UMA ferramenta.

Ferramentas disponíveis: %s

Contexto: %s
Passo do loop: %d

Responda APENAS com o nome da ferramenta e os argumentos em formato JSON.
Exemplo: {"tool": "getMetrics", "args": {"service": "payment-api"}}
`

// ToolSelectionEvaluator avalia a qualidade das decisões de seleção de
// ferramentas do agente — equivalente a ToolSelectionEvaluator.java
// (aula 12). Usa o ChatClient para simular decisões reais do modelo.
type ToolSelectionEvaluator struct {
	Client         ChatClient
	MinAccuracy    float64
	MaxUnnecessary float64
}

// Evaluate executa a avaliação de tool selection sobre os 5 cenários
// hardcoded (mesmos da versão Java).
func (e *ToolSelectionEvaluator) Evaluate(ctx context.Context) model.ToolSelectionReport {
	cases := buildTestCases()

	var correctTool, correctArgs, unnecessaryCalls, wrongTool int

	for _, tc := range cases {
		response := e.askModel(ctx, tc)
		selected := extractToolName(response)

		toolCorrect := strings.EqualFold(tc.ExpectedTool, selected)
		forbidden := false
		for _, f := range tc.ForbiddenTools {
			if strings.EqualFold(f, selected) {
				forbidden = true
				break
			}
		}

		if toolCorrect {
			correctTool++
			argsOK := true
			lowerResponse := strings.ToLower(response)
			for key := range tc.ExpectedArgs {
				if !strings.Contains(lowerResponse, strings.ToLower(key)) {
					argsOK = false
					break
				}
			}
			if argsOK {
				correctArgs++
			}
		} else {
			wrongTool++
		}

		if forbidden {
			unnecessaryCalls++
		}
	}

	total := len(cases)
	toolAccuracy := float64(correctTool) / float64(total)
	argAccuracy := float64(correctArgs) / float64(total)
	unnecessaryRate := float64(unnecessaryCalls) / float64(total)
	wrongRate := float64(wrongTool) / float64(total)

	minAccuracy := e.minAccuracy()
	maxUnnecessary := e.maxUnnecessary()
	passed := toolAccuracy >= minAccuracy && unnecessaryRate <= maxUnnecessary

	return model.ToolSelectionReport{
		TotalCases:            total,
		ToolSelectionAccuracy: round(toolAccuracy),
		ArgumentAccuracy:      round(argAccuracy),
		UnnecessaryCallsRate:  round(unnecessaryRate),
		WrongToolRate:         round(wrongRate),
		Passed:                passed,
	}
}

func (e *ToolSelectionEvaluator) minAccuracy() float64 {
	if e.MinAccuracy == 0 {
		return 0.80
	}
	return e.MinAccuracy
}

func (e *ToolSelectionEvaluator) maxUnnecessary() float64 {
	if e.MaxUnnecessary == 0 {
		return 0.10
	}
	return e.MaxUnnecessary
}

// askModel consulta o modelo para decidir qual ferramenta usar. Em caso de
// falha na API, retorna "{}" para contabilizar como resposta errada — mesmo
// comportamento defensivo da versão Java.
func (e *ToolSelectionEvaluator) askModel(ctx context.Context, tc model.ToolSelectionCase) string {
	prompt := fmt.Sprintf(toolSelectionPromptTemplate, availableTools, tc.Context, tc.LoopStep)

	response, err := e.Client.Chat(ctx, prompt)
	if err != nil {
		return "{}"
	}
	return response
}

func extractToolName(response string) string {
	matches := toolNamePattern.FindStringSubmatch(response)
	if matches == nil {
		return "unknown"
	}
	return matches[1]
}

// buildTestCases define os 5 casos de teste de tool selection — cada caso
// representa uma situação real de um loop de agente de monitoramento.
func buildTestCases() []model.ToolSelectionCase {
	return []model.ToolSelectionCase{
		{
			ID: "ts-001", Context: "Latência alta detectada em payment-api. Ainda não coletamos dados de métricas.",
			LoopStep: 1, ExpectedTool: "getMetrics",
			ExpectedArgs:   map[string]any{"service": "payment-api", "metric": "latency"},
			ForbiddenTools: []string{"saveIncident", "notifyTeam"},
		},
		{
			ID: "ts-002", Context: "Métricas coletadas mostram CPU em 95%. Preciso entender o que aconteceu antes de abrir incidente.",
			LoopStep: 2, ExpectedTool: "getLogs",
			ExpectedArgs:   map[string]any{"service": "payment-api", "level": "ERROR"},
			ForbiddenTools: []string{"saveIncident", "notifyTeam"},
		},
		{
			ID: "ts-003", Context: "Logs mostram OutOfMemoryError às 14h. Houve algum deploy próximo desse horário?",
			LoopStep: 3, ExpectedTool: "getDeployHistory",
			ExpectedArgs:   map[string]any{"service": "payment-api", "since": "13:00"},
			ForbiddenTools: []string{"saveIncident", "notifyTeam"},
		},
		{
			ID: "ts-004", Context: "Deploy identificado às 13:45. Causa raiz confirmada: memory leak na versão 2.3.1. Devo registrar o incidente.",
			LoopStep: 4, ExpectedTool: "saveIncident",
			ExpectedArgs:   map[string]any{"service": "payment-api", "severity": "high"},
			ForbiddenTools: []string{"getMetrics", "getLogs"},
		},
		{
			ID: "ts-005", Context: "Incidente registrado. Time de on-call precisa ser notificado imediatamente.",
			LoopStep: 5, ExpectedTool: "notifyTeam",
			ExpectedArgs:   map[string]any{"channel": "on-call", "severity": "high"},
			ForbiddenTools: []string{"getMetrics", "getLogs", "getDeployHistory"},
		},
	}
}
