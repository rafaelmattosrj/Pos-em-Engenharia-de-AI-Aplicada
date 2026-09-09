// Package strategies contém a estratégia ReAct de referência (porte
// simplificado de agents/react.ts / strategies/react.ts): loop
// observação->ação single-agent, usada como baseline ao lado da estratégia
// multiagente TeamStrategy (pacote team), que é o foco desta unidade.
package strategies

import (
	"fmt"
	"time"

	"opspilot/domain"
	"opspilot/llm"
	"opspilot/tools"
)

const ReactSystemPrompt = "Voce e o OpsPilot, copiloto de plantao de incidentes. Use as ferramentas disponiveis " +
	"para diagnosticar e agir; responda com a resposta final quando tiver o suficiente."

const defaultMaxIterations = 10

// ReactStrategy é o porte de ReactStrategy.java / react.ts.
type ReactStrategy struct {
	Model         llm.ChatModel
	Tools         []tools.Tool
	MaxIterations int
}

var _ domain.ReasoningStrategy = (*ReactStrategy)(nil)

// NewReactStrategy cria a estratégia com o teto de iterações default (10),
// igual ao construtor de dois argumentos do porte Java.
func NewReactStrategy(model llm.ChatModel, toolset []tools.Tool) *ReactStrategy {
	return NewReactStrategyWithMaxIterations(model, toolset, defaultMaxIterations)
}

func NewReactStrategyWithMaxIterations(model llm.ChatModel, toolset []tools.Tool, maxIterations int) *ReactStrategy {
	return &ReactStrategy{Model: model, Tools: toolset, MaxIterations: maxIterations}
}

func (s *ReactStrategy) Name() string { return "react" }

func (s *ReactStrategy) Run(input domain.StrategyRunInput) (domain.StrategyResult, error) {
	startedAt := time.Now()
	messages := []llm.ChatMessage{
		llm.SystemMessage(ReactSystemPrompt),
		llm.UserMessage(input.Message),
	}
	var trace []domain.TraceEvent
	llmCalls := 0
	answer := ""

	maxIterations := s.MaxIterations
	if maxIterations <= 0 {
		maxIterations = defaultMaxIterations
	}

	for iteration := 0; iteration < maxIterations; iteration++ {
		response, err := s.Model.Invoke(messages, s.Tools)
		llmCalls++
		if err != nil {
			return domain.StrategyResult{}, err
		}

		if len(response.ToolCalls) == 0 {
			answer = response.Content
			trace = append(trace, domain.Answer("react", answer))
			break
		}

		for _, call := range response.ToolCalls {
			trace = append(trace, domain.Action("react", fmt.Sprintf("%s %v", call.ToolName, call.Args)))
			observation := executeTool(s.Tools, call)
			trace = append(trace, domain.Observation("react", observation))
			messages = append(messages, llm.ToolMessage(observation))
		}
	}

	return domain.StrategyResult{
		Answer: answer,
		Trace:  trace,
		Metrics: domain.ExecutionMetrics{
			LLMCalls:  llmCalls,
			LatencyMs: time.Since(startedAt).Milliseconds(),
		},
	}, nil
}

func executeTool(toolset []tools.Tool, call llm.ToolCall) string {
	for _, tool := range toolset {
		if tool.Name() == call.ToolName {
			result, err := tool.Execute(call.Args)
			if err != nil {
				return "Erro: " + err.Error()
			}
			return result
		}
	}
	return fmt.Sprintf("Erro: ferramenta %q nao encontrada.", call.ToolName)
}
