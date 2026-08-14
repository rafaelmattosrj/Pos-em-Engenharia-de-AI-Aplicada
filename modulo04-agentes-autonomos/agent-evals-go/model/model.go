// Package model define os tipos de domínio do framework de avaliação —
// equivalente a EvalScenario.java, ToolSelectionCase.java,
// ToolSelectionReport.java, MemoryImpactReport.java, BenchmarkMetrics.java
// e BenchmarkReport.java.
package model

import "time"

// EvalScenario representa um cenário de avaliação do dataset.
type EvalScenario struct {
	ID            string   `json:"id"`
	Input         string   `json:"input"`
	Difficulty    string   `json:"difficulty"`
	ExpectedTools []string `json:"expectedTools"`
}

// ToolSelectionCase é um caso de teste para avaliação de seleção de
// ferramentas.
type ToolSelectionCase struct {
	ID             string
	Context        string
	LoopStep       int
	ExpectedTool   string
	ExpectedArgs   map[string]any
	ForbiddenTools []string
}

// ToolSelectionReport é o relatório da avaliação de seleção de ferramentas.
// Passed só é true se ToolSelectionAccuracy >= 0.80 e UnnecessaryCallsRate
// <= 0.10 (defaults).
type ToolSelectionReport struct {
	TotalCases            int     `json:"totalCases"`
	ToolSelectionAccuracy float64 `json:"toolSelectionAccuracy"`
	ArgumentAccuracy      float64 `json:"argumentAccuracy"`
	UnnecessaryCallsRate  float64 `json:"unnecessaryCallsRate"`
	WrongToolRate         float64 `json:"wrongToolRate"`
	Passed                bool    `json:"passed"`
}

// MemoryImpactReport compara execuções com e sem memória.
type MemoryImpactReport struct {
	TotalRuns               int     `json:"totalRuns"`
	RetrievalPrecision      float64 `json:"retrievalPrecision"`
	RetrievalRecall         float64 `json:"retrievalRecall"`
	MemoryUtilization       float64 `json:"memoryUtilization"`
	HallucinationFromMemory float64 `json:"hallucinationFromMemory"`
	DecisionImprovement     float64 `json:"decisionImprovement"`
	LessonQuality           float64 `json:"lessonQuality"`
}

// BenchmarkMetrics são as métricas consolidadas de benchmark para uma
// arquitetura cognitiva específica.
type BenchmarkMetrics struct {
	Architecture       string  `json:"architecture"`
	CompletionRate     float64 `json:"completionRate"`
	AvgSteps           float64 `json:"avgSteps"`
	AvgTokensEstimated int     `json:"avgTokensEstimated"`
	ToolSuccessRate    float64 `json:"toolSuccessRate"`
	ToolCoverage       float64 `json:"toolCoverage"`
}

// BenchmarkReport é o relatório completo de benchmark comparando as 3
// arquiteturas cognitivas.
type BenchmarkReport struct {
	GeneratedAt    time.Time          `json:"generatedAt"`
	TotalScenarios int                `json:"totalScenarios"`
	Results        []BenchmarkMetrics `json:"results"`
	Verdict        map[string]string  `json:"verdict"`
}
