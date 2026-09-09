package team

import (
	"strings"
	"time"

	"opspilot/domain"
)

// TeamStrategy é o porte de TeamStrategy (team/team-strategy.ts): expõe o
// TeamGraph como um domain.ReasoningStrategy, ao lado de strategies.ReactStrategy.
type TeamStrategy struct {
	graph *TeamGraph
}

var _ domain.ReasoningStrategy = (*TeamStrategy)(nil)

func NewTeamStrategy(decideNext DecideNextFn, roleRunners map[Role]RoleRunner, supervisorLLMCalls int) *TeamStrategy {
	return &TeamStrategy{graph: NewTeamGraph(decideNext, roleRunners, supervisorLLMCalls)}
}

func (s *TeamStrategy) Name() string { return "team" }

func (s *TeamStrategy) Run(input domain.StrategyRunInput) (domain.StrategyResult, error) {
	startedAt := time.Now()
	result := s.graph.Run(composeMessage(input))
	return domain.StrategyResult{
		Answer: result.Answer,
		Trace:  result.Trace,
		Metrics: domain.ExecutionMetrics{
			LLMCalls:  result.LLMCalls,
			LatencyMs: time.Since(startedAt).Milliseconds(),
		},
	}, nil
}

func composeMessage(input domain.StrategyRunInput) string {
	if len(input.History) == 0 {
		return input.Message
	}
	var sb strings.Builder
	sb.WriteString("Historico recente da conversa:\n")
	for _, message := range input.History {
		sb.WriteString(message.Role)
		sb.WriteString(": ")
		sb.WriteString(message.Content)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
	sb.WriteString(input.Message)
	return sb.String()
}
