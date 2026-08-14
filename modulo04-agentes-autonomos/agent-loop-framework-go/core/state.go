package core

import (
	"strings"
	"time"
)

// AgentState é o estado mutável do agente durante uma execução do loop —
// equivalente a AgentState.java. Uma nova instância deve ser criada a cada
// execução de AgentLoop.Run.
type AgentState struct {
	currentStep           int
	accumulatedContext    strings.Builder
	lastToolsUsed         []string
	consecutiveNoProgress int
	startTime             time.Time
	done                  bool
}

// NewAgentState cria um AgentState limpo, com o cronômetro iniciado agora.
func NewAgentState() *AgentState {
	return &AgentState{startTime: time.Now()}
}

// IncrementStep incrementa o step atual e retorna o novo valor.
func (s *AgentState) IncrementStep() int {
	s.currentStep++
	return s.currentStep
}

// AddContext adiciona texto ao contexto acumulado.
func (s *AgentState) AddContext(text string) {
	s.accumulatedContext.WriteString("\n")
	s.accumulatedContext.WriteString(text)
}

// RecordToolUsed registra a tool utilizada no step atual, detectando
// automaticamente se não houve progresso (mesma tool repetida).
func (s *AgentState) RecordToolUsed(toolName string) {
	if len(s.lastToolsUsed) > 0 && s.lastToolsUsed[len(s.lastToolsUsed)-1] == toolName {
		s.consecutiveNoProgress++
	} else {
		s.consecutiveNoProgress = 0
	}
	s.lastToolsUsed = append(s.lastToolsUsed, toolName)
}

// CheckNoProgress verifica se o agente está preso em loop sem progresso.
func (s *AgentState) CheckNoProgress(noProgressLimit int) bool {
	return s.consecutiveNoProgress >= noProgressLimit
}

// ElapsedSeconds retorna o tempo decorrido desde o início da execução.
func (s *AgentState) ElapsedSeconds() int64 {
	return int64(time.Since(s.startTime).Seconds())
}

func (s *AgentState) CurrentStep() int           { return s.currentStep }
func (s *AgentState) AccumulatedContext() string { return s.accumulatedContext.String() }
func (s *AgentState) LastToolsUsed() []string    { return s.lastToolsUsed }
func (s *AgentState) ConsecutiveNoProgress() int { return s.consecutiveNoProgress }
func (s *AgentState) Done() bool                 { return s.done }
func (s *AgentState) SetDone(done bool)          { s.done = done }
