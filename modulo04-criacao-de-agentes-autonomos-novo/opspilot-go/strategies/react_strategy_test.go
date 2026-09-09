package strategies

import (
	"strings"
	"testing"

	"opspilot/domain"
	"opspilot/llm"
	"opspilot/tools"
)

type fakeTool struct {
	name    string
	execute func(args map[string]any) (string, error)
}

func (t *fakeTool) Name() string        { return t.name }
func (t *fakeTool) Description() string { return "" }
func (t *fakeTool) Execute(args map[string]any) (string, error) {
	return t.execute(args)
}

func TestAnswersDirectlyWithoutToolsWhenModelHasNoToolCall(t *testing.T) {
	model := llm.NewFakeChatModelScripted([]llm.ModelResponse{llm.TextResponse("Tudo operacional.")})
	strategy := NewReactStrategy(model, nil)

	result, err := strategy.Run(domain.NewStrategyRunInput("status geral?"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Answer != "Tudo operacional." {
		t.Fatalf("unexpected answer: %q", result.Answer)
	}
	if result.Metrics.LLMCalls != 1 {
		t.Fatalf("expected 1 llm call, got %d", result.Metrics.LLMCalls)
	}
}

func TestUsesToolBeforeFinalAnswer(t *testing.T) {
	checkStatus := &fakeTool{name: "check_provider_status", execute: func(args map[string]any) (string, error) {
		return "operacional", nil
	}}
	model := llm.NewFakeChatModelScripted([]llm.ModelResponse{
		{ToolCalls: []llm.ToolCall{{ToolName: "check_provider_status"}}},
		llm.TextResponse("Provedores operacionais."),
	})

	strategy := NewReactStrategy(model, []tools.Tool{checkStatus})
	result, err := strategy.Run(domain.NewStrategyRunInput("provedores estao ok?"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Answer != "Provedores operacionais." {
		t.Fatalf("unexpected answer: %q", result.Answer)
	}
	if result.Metrics.LLMCalls != 2 {
		t.Fatalf("expected 2 llm calls, got %d", result.Metrics.LLMCalls)
	}
	found := false
	for _, event := range result.Trace {
		if strings.Contains(event.Content, "operacional") {
			found = true
		}
	}
	if !found {
		t.Fatal("expected trace to contain tool observation")
	}
}

func TestStopsAtMaxIterationsIfModelNeverAnswers(t *testing.T) {
	loopTool := &fakeTool{name: "loop", execute: func(args map[string]any) (string, error) {
		return "de novo", nil
	}}
	model := llm.NewFakeChatModelFunc(func(messages []llm.ChatMessage) (llm.ModelResponse, error) {
		return llm.ModelResponse{ToolCalls: []llm.ToolCall{{ToolName: "loop"}}}, nil
	})

	strategy := NewReactStrategyWithMaxIterations(model, []tools.Tool{loopTool}, 3)
	result, err := strategy.Run(domain.NewStrategyRunInput("nunca termina"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Answer != "" {
		t.Fatalf("expected empty answer, got %q", result.Answer)
	}
	if result.Metrics.LLMCalls != 3 {
		t.Fatalf("expected 3 llm calls, got %d", result.Metrics.LLMCalls)
	}
}
