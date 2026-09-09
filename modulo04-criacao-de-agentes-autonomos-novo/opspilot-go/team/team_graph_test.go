package team

import (
	"errors"
	"strings"
	"testing"

	"opspilot/domain"
	"opspilot/tools"
)

type fakeRunner struct {
	role    Role
	kind    BlackboardKind
	content string
}

func (r *fakeRunner) Role() Role          { return r.role }
func (r *fakeRunner) Tools() []tools.Tool { return nil }
func (r *fakeRunner) Run(input RoleRunInput) (RoleRunResult, error) {
	return RoleRunResult{
		Entry:    BlackboardEntry{Role: r.role, Kind: r.kind, Brief: input.Brief, Content: r.content},
		LLMCalls: 1,
	}, nil
}

func allRunners() map[Role]RoleRunner {
	return map[Role]RoleRunner{
		RoleAnalista:   &fakeRunner{role: RoleAnalista, kind: KindFacts, content: "fatos coletados"},
		RolePlanejador: &fakeRunner{role: RolePlanejador, kind: KindPlan, content: "plano de 3 passos"},
		RoleExecutor:   &fakeRunner{role: RoleExecutor, kind: KindExecution, content: "acao executada"},
	}
}

func TestDelegatesThenFinishesWithBriefAsAnswer(t *testing.T) {
	decideNext := func(message string, blackboard []BlackboardEntry, handoffCount int) (SupervisorDecision, error) {
		if len(blackboard) == 0 {
			return SupervisorDecision{Next: RoleAnalista, Brief: "levante os fatos"}, nil
		}
		return DoneDecision("Resumo final do supervisor"), nil
	}
	graph := NewTeamGraph(decideNext, allRunners(), 1)

	result := graph.Run("o que esta acontecendo?")

	if result.Answer != "Resumo final do supervisor" {
		t.Fatalf("unexpected answer: %q", result.Answer)
	}
	if result.LLMCalls != 3 { // 2 chamadas ao supervisor + 1 ao papel
		t.Fatalf("expected 3 llm calls, got %d", result.LLMCalls)
	}
}

func TestFallsBackToBlackboardWhenDoneBriefIsEmpty(t *testing.T) {
	calls := 0
	decideNext := func(message string, blackboard []BlackboardEntry, handoffCount int) (SupervisorDecision, error) {
		if calls == 0 {
			calls++
			return SupervisorDecision{Next: RoleAnalista, Brief: "levante os fatos"}, nil
		}
		return DoneDecision(""), nil
	}
	graph := NewTeamGraph(decideNext, allRunners(), 1)

	result := graph.Run("mensagem")

	if !strings.Contains(result.Answer, "Resumo do blackboard") || !strings.Contains(result.Answer, "fatos coletados") {
		t.Fatalf("unexpected answer: %q", result.Answer)
	}
}

func TestStopsAtHandoffCapEvenIfSupervisorNeverSaysDone(t *testing.T) {
	decideNext := func(message string, blackboard []BlackboardEntry, handoffCount int) (SupervisorDecision, error) {
		return SupervisorDecision{Next: RoleAnalista, Brief: "de novo"}, nil
	}
	graph := NewTeamGraph(decideNext, allRunners(), 1)

	result := graph.Run("mensagem")

	handoffEvents := 0
	for _, event := range result.Trace {
		if event.Type == domain.TraceHandoff {
			handoffEvents++
		}
	}
	if handoffEvents != MaxHandoffs+1 { // +1 é o handoff final pra "done"
		t.Fatalf("expected %d handoff events, got %d", MaxHandoffs+1, handoffEvents)
	}
	if !strings.Contains(result.Answer, "Resumo do blackboard") {
		t.Fatalf("unexpected answer: %q", result.Answer)
	}
}

func TestInvalidSupervisorDecisionEndsGracefully(t *testing.T) {
	decideNext := func(message string, blackboard []BlackboardEntry, handoffCount int) (SupervisorDecision, error) {
		return SupervisorDecision{}, errors.New("json malformado")
	}
	graph := NewTeamGraph(decideNext, allRunners(), 1)

	result := graph.Run("mensagem")

	if !strings.Contains(result.Answer, "A equipe encerrou sem contribuicoes") {
		t.Fatalf("unexpected answer: %q", result.Answer)
	}
	if !strings.Contains(result.Trace[0].Content, InvalidDecisionPrefix) {
		t.Fatalf("expected first trace event to contain invalid decision prefix, got %q", result.Trace[0].Content)
	}
}

func TestEmptyMessageWithNoBlackboardStillProducesAnswer(t *testing.T) {
	decideNext := func(message string, blackboard []BlackboardEntry, handoffCount int) (SupervisorDecision, error) {
		return DoneDecision(""), nil
	}
	graph := NewTeamGraph(decideNext, allRunners(), 1)

	result := graph.Run("mensagem vazia de contribuicoes")

	expected := "A equipe encerrou sem contribuicoes no blackboard para: mensagem vazia de contribuicoes"
	if !strings.Contains(result.Answer, expected) {
		t.Fatalf("unexpected answer: %q", result.Answer)
	}
}
