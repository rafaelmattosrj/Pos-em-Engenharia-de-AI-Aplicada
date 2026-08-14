// Package model define os tipos de domínio das arquiteturas cognitivas —
// equivalente a AgentRequest.java, AgentResponse.java, ExecutionPlan.java e
// CritiqueResult.java.
package model

// AgentRequest é a requisição enviada para executar um agente.
type AgentRequest struct {
	Architecture string `json:"architecture"` // "react" | "plan-execute" | "reflection"
	Input        string `json:"input"`
}

// AgentMetrics são as métricas de execução coletadas durante o ciclo.
type AgentMetrics struct {
	Steps       int `json:"steps"`
	Tokens      int `json:"tokens"`
	Reflections int `json:"reflections"`
}

// AgentResponse é a resposta retornada após a execução do agente.
type AgentResponse struct {
	Result  string       `json:"result"`
	Metrics AgentMetrics `json:"metrics"`
}

// PlanStep é um passo individual dentro do plano de execução.
type PlanStep struct {
	StepNumber      int            `json:"stepNumber"`
	Description     string         `json:"description"`
	Tool            string         `json:"tool"`
	Args            map[string]any `json:"args"`
	SuccessCriteria string         `json:"successCriteria"`
}

// ExecutionPlan é o plano de execução gerado pelo PlanExecuteAgent.
type ExecutionPlan struct {
	Steps []PlanStep `json:"steps"`
}

// CritiqueResult é o resultado da avaliação crítica produzida pelo
// CritiqueEvaluator.
type CritiqueResult struct {
	Score        float64 `json:"score"`
	Passed       bool    `json:"passed"`
	Feedback     string  `json:"feedback"`
	Correctness  float64 `json:"correctness"`
	Completeness float64 `json:"completeness"`
	Quality      float64 `json:"quality"`
}
