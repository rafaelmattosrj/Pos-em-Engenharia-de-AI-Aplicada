package evaluator

import (
	"math/rand"

	"agent-evals/model"
)

const defaultMemoryEvalRuns = 10

// MemoryEvaluator avalia o impacto do uso de memória no desempenho do
// agente — equivalente a MemoryEvaluator.java (aula 15). Simula múltiplas
// execuções com e sem memória e calcula métricas de comparação. Em
// produção, substituir pela execução real do agente com e sem memória
// ativa.
type MemoryEvaluator struct {
	rng *rand.Rand
}

// NewMemoryEvaluator cria um MemoryEvaluator com seed fixo (99).
func NewMemoryEvaluator() *MemoryEvaluator {
	return &MemoryEvaluator{rng: rand.New(rand.NewSource(99))}
}

// Evaluate executa a avaliação com o número padrão de execuções simuladas
// (10).
func (e *MemoryEvaluator) Evaluate() model.MemoryImpactReport {
	return e.EvaluateRuns(defaultMemoryEvalRuns)
}

// EvaluateRuns executa a avaliação com um número configurável de execuções.
func (e *MemoryEvaluator) EvaluateRuns(totalRuns int) model.MemoryImpactReport {
	return model.MemoryImpactReport{
		TotalRuns:               totalRuns,
		RetrievalPrecision:      round(e.simulateMetric(0.84, 0.06)),
		RetrievalRecall:         round(e.simulateMetric(0.79, 0.07)),
		MemoryUtilization:       round(e.simulateMetric(0.71, 0.08)),
		HallucinationFromMemory: round(e.simulateMetric(0.06, 0.03)),
		DecisionImprovement:     round(e.simulateMetric(0.23, 0.09)),
		LessonQuality:           round(e.simulateMetric(0.78, 0.05)),
	}
}

// simulateMetric simula uma métrica com valor base e variação aleatória
// controlada, restrita ao intervalo [0, 1].
func (e *MemoryEvaluator) simulateMetric(base, variance float64) float64 {
	value := base + (e.rng.Float64()-0.5)*variance*2
	return clamp(value, 0, 1)
}
