package team

import (
	"fmt"
	"strings"

	"opspilot/domain"
)

// MaxHandoffs é o porte de MAX_HANDOFFS (team/team-graph.ts) -- teto de
// delegações do supervisor por turno (spec 018, US5).
const MaxHandoffs = 8

// Prefixos estáveis, testados e documentados em contracts/trace-handoff.md
// no original.
const (
	CapReachedPrefix      = "teto de handoffs atingido"
	InvalidDecisionPrefix = "decisao invalida do supervisor"
)

// TeamGraph é o porte de team-graph.ts: a mesma máquina de estado (sem
// depender do StateGraph do LangGraph, que não tem equivalente maduro em Go)
// -- supervisor decide o próximo papel, cada papel contribui uma entrada no
// blackboard, até a decisão "done" ou o teto de handoffs (MaxHandoffs). O
// estado da rodada (blackboard, trace, contadores) vive em variáveis locais
// de Run; não há grafo declarativo, só a função de transição explícita.
type TeamGraph struct {
	DecideNext         DecideNextFn
	RoleRunners        map[Role]RoleRunner
	SupervisorLLMCalls int
}

// NewTeamGraph cria o coordenador do time. supervisorLLMCalls é quantas
// chamadas de LLM contar por decisão real do supervisor (0 para fakes
// injetados em teste).
func NewTeamGraph(decideNext DecideNextFn, roleRunners map[Role]RoleRunner, supervisorLLMCalls int) *TeamGraph {
	runners := make(map[Role]RoleRunner, len(roleRunners))
	for role, runner := range roleRunners {
		runners[role] = runner
	}
	return &TeamGraph{DecideNext: decideNext, RoleRunners: runners, SupervisorLLMCalls: supervisorLLMCalls}
}

// TeamGraphResult é o porte de TeamGraph.Result (porte Java).
type TeamGraphResult struct {
	Answer   string
	Trace    []domain.TraceEvent
	LLMCalls int
}

// Run é a função coordenadora que orquestra supervisor + papéis até a
// decisão "done" ou o teto de handoffs -- porte comportamental 1:1 de
// runTeamGraph / supervisorNode / roleNode / doneNode (team-graph.ts).
func (g *TeamGraph) Run(message string) TeamGraphResult {
	var blackboard []BlackboardEntry
	var trace []domain.TraceEvent
	handoffCount := 0
	llmCalls := 0
	lastBrief := ""

	for {
		if handoffCount >= MaxHandoffs {
			content := CapReachedPrefix + ": encerrando com o conteudo do blackboard"
			trace = append(trace, domain.Handoff("supervisor", "done", content))
			lastBrief = ""
			break
		}

		decision, err := g.DecideNext(message, blackboard, handoffCount)
		if err != nil {
			content := InvalidDecisionPrefix + ": " + err.Error()
			trace = append(trace, domain.Handoff("supervisor", "done", content))
			llmCalls += g.SupervisorLLMCalls
			lastBrief = ""
			break
		}

		nextLabel := "done"
		if !decision.Done {
			nextLabel = string(decision.Next)
		}
		trace = append(trace, domain.Handoff("supervisor", nextLabel, decision.Brief))
		llmCalls += g.SupervisorLLMCalls

		if decision.Done {
			lastBrief = decision.Brief
			break
		}

		handoffCount++
		runner, ok := g.RoleRunners[decision.Next]
		if !ok {
			content := fmt.Sprintf("erro no papel %s: nenhum role runner configurado", decision.Next)
			blackboard = append(blackboard, BlackboardEntry{Role: decision.Next, Kind: KindError, Brief: decision.Brief, Content: content})
			trace = append(trace, domain.Observation(string(decision.Next), content))
			continue
		}

		blackboardSnapshot := append([]BlackboardEntry(nil), blackboard...)
		result, err := runner.Run(RoleRunInput{Message: message, Brief: decision.Brief, Blackboard: blackboardSnapshot})
		if err != nil {
			content := fmt.Sprintf("erro no papel %s: %s", decision.Next, err.Error())
			blackboard = append(blackboard, BlackboardEntry{Role: decision.Next, Kind: KindError, Brief: decision.Brief, Content: content})
			trace = append(trace, domain.Observation(string(decision.Next), content))
			continue
		}

		blackboard = append(blackboard, result.Entry)
		trace = append(trace, result.Trace...)
		llmCalls += result.LLMCalls
	}

	var answer string
	trimmedBrief := strings.TrimSpace(lastBrief)
	switch {
	case trimmedBrief != "":
		answer = trimmedBrief
	case len(blackboard) > 0:
		answer = "Resumo do blackboard:\n" + RenderBlackboard(blackboard)
	default:
		answer = "A equipe encerrou sem contribuicoes no blackboard para: " + message
	}
	trace = append(trace, domain.Answer("supervisor", answer))

	return TeamGraphResult{Answer: answer, Trace: trace, LLMCalls: llmCalls}
}
