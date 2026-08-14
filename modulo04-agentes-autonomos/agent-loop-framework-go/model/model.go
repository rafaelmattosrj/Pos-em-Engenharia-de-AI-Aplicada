// Package model define os tipos de domínio do agent loop — equivalente a
// PlanDecision.java, ToolResult.java e AgentTrace.java.
package model

// PlanDecision é a decisão de planejamento retornada pelo LLM a cada step
// do loop.
type PlanDecision struct {
	Reasoning       string         `json:"reasoning"`
	Action          string         `json:"action"`
	Tool            string         `json:"tool"`
	Args            map[string]any `json:"args"`
	SuccessCriteria string         `json:"successCriteria"`
	Done            bool           `json:"done"`
}

// ToolResult é o resultado da execução de uma tool pelo Executor.
type ToolResult struct {
	ToolName string `json:"toolName"`
	Success  bool   `json:"success"`
	Output   string `json:"output,omitempty"`
	Error    string `json:"error,omitempty"`
}

// OkResult cria um ToolResult de sucesso.
func OkResult(toolName, output string) ToolResult {
	return ToolResult{ToolName: toolName, Success: true, Output: output}
}

// FailResult cria um ToolResult de erro.
func FailResult(toolName, errMsg string) ToolResult {
	return ToolResult{ToolName: toolName, Success: false, Error: errMsg}
}

// StepTrace é o trace de um único step do agent loop.
type StepTrace struct {
	StepNumber int          `json:"stepNumber"`
	Perception string       `json:"perception"`
	Plan       PlanDecision `json:"plan"`
	ToolResult *ToolResult  `json:"toolResult"`
	Evaluation string       `json:"evaluation"`
}

// AgentTrace é o trace completo de uma execução do agente.
type AgentTrace struct {
	Input                string      `json:"input"`
	Steps                []StepTrace `json:"steps"`
	FinalResult          string      `json:"finalResult"`
	TotalSteps           int         `json:"totalSteps"`
	TotalTokensEstimated int         `json:"totalTokensEstimated"`
	DurationMs           int64       `json:"durationMs"`
}

// NewAgentTrace cria um AgentTrace vazio para o input informado.
func NewAgentTrace(input string) *AgentTrace {
	return &AgentTrace{Input: input}
}

// AddStep adiciona um step ao trace.
func (t *AgentTrace) AddStep(step StepTrace) {
	t.Steps = append(t.Steps, step)
}
