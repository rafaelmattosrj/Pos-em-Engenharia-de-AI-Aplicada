package evaluator

import "testing"

// Cenário adicional de EvalsTest.java: MemoryEvaluator deve retornar
// métricas entre 0 e 1.
func TestEvaluate_ReturnsMetricsInValidRange(t *testing.T) {
	evaluator := NewMemoryEvaluator()

	report := evaluator.Evaluate()

	if report.TotalRuns <= 0 {
		t.Errorf("esperava totalRuns > 0, obteve %d", report.TotalRuns)
	}

	checks := map[string]float64{
		"retrievalPrecision":      report.RetrievalPrecision,
		"retrievalRecall":         report.RetrievalRecall,
		"memoryUtilization":       report.MemoryUtilization,
		"hallucinationFromMemory": report.HallucinationFromMemory,
		"decisionImprovement":     report.DecisionImprovement,
		"lessonQuality":           report.LessonQuality,
	}
	for name, value := range checks {
		if value < 0 || value > 1 {
			t.Errorf("%s fora da faixa [0,1]: %f", name, value)
		}
	}
}

func TestEvaluateRuns_RespectsCustomRunCount(t *testing.T) {
	evaluator := NewMemoryEvaluator()

	report := evaluator.EvaluateRuns(25)
	if report.TotalRuns != 25 {
		t.Errorf("esperava totalRuns=25, obteve %d", report.TotalRuns)
	}
}
