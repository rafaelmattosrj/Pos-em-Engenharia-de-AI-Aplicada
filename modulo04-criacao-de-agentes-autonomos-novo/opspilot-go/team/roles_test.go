package team

import (
	"fmt"
	"strings"
	"testing"

	"opspilot/llm"
	"opspilot/tools"
)

type echoTool struct {
	name string
}

func (t *echoTool) Name() string        { return t.name }
func (t *echoTool) Description() string { return "tool de teste" }
func (t *echoTool) Execute(args map[string]any) (string, error) {
	return fmt.Sprintf("resultado de %s %v", t.name, args), nil
}

func TestPlanejadorMakesSingleDirectCall(t *testing.T) {
	model := llm.NewFakeChatModelScripted([]llm.ModelResponse{llm.TextResponse("1. Fazer X\n2. Fazer Y")})
	planejador := NewPlanejadorRunner(model)

	result, err := planejador.Run(RoleRunInput{Message: "mensagem", Brief: "planeje"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.LLMCalls != 1 {
		t.Fatalf("expected 1 llm call, got %d", result.LLMCalls)
	}
	if result.Entry.Kind != KindPlan {
		t.Fatalf("expected PLAN kind, got %s", result.Entry.Kind)
	}
	if !strings.Contains(result.Entry.Content, "Fazer X") {
		t.Fatalf("unexpected content: %s", result.Entry.Content)
	}
	if model.CallCount() != 1 {
		t.Fatalf("expected model called once, got %d", model.CallCount())
	}
}

func TestAnalistaUsesToolThenAnswers(t *testing.T) {
	listIncidents := &echoTool{name: "list_incidents"}
	model := llm.NewFakeChatModelScripted([]llm.ModelResponse{
		{ToolCalls: []llm.ToolCall{{ToolName: "list_incidents"}}},
		llm.TextResponse("2 incidentes abertos, nenhum critico."),
	})

	analista := NewAnalistaRunner(model, []tools.Tool{listIncidents})
	result, err := analista.Run(RoleRunInput{Message: "o que esta pegando?", Brief: "diagnostique"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.LLMCalls != 2 {
		t.Fatalf("expected 2 llm calls, got %d", result.LLMCalls)
	}
	if result.Entry.Kind != KindFacts {
		t.Fatalf("expected FACTS kind, got %s", result.Entry.Kind)
	}
	if !strings.Contains(result.Entry.Content, "nenhum critico") {
		t.Fatalf("unexpected content: %s", result.Entry.Content)
	}
}

func TestExecutorFallsBackToErrorMessageForUnknownTool(t *testing.T) {
	model := llm.NewFakeChatModelScripted([]llm.ModelResponse{
		{ToolCalls: []llm.ToolCall{{ToolName: "tool_inexistente"}}},
		llm.TextResponse("Nao consegui executar a acao."),
	})

	executor := NewExecutorRunner(model, nil)
	result, err := executor.Run(RoleRunInput{Message: "resolva o incidente", Brief: "execute"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, event := range result.Trace {
		if strings.Contains(event.Content, "nao disponivel") {
			found = true
		}
	}
	if !found {
		t.Fatal("expected trace to contain 'nao disponivel' observation")
	}
}
