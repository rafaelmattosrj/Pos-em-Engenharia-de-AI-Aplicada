// Package evaluator implementa os avaliadores do framework de evals —
// equivalente ao pacote evaluator/ da versão Java (aulas 09, 12, 15).
package evaluator

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"agent-evals/dataset"
	"agent-evals/model"
)

// BenchmarkRunner executa benchmarks comparativos simulados entre as 3
// arquiteturas cognitivas — equivalente a BenchmarkRunner.java (aula 09).
//
// Os resultados são simulados com variações realistas (seed fixo para
// reprodutibilidade) para fins de avaliação do framework, assim como na
// versão Java — a fonte de aleatoriedade é math/rand (não replica bit-a-bit
// a sequência de java.util.Random, mas preserva as mesmas faixas e
// distribuição estatística).
type BenchmarkRunner struct {
	Dataset     *dataset.EvalDataset
	ReportsPath string
	rng         *rand.Rand
}

// NewBenchmarkRunner cria um BenchmarkRunner com seed fixo (42).
func NewBenchmarkRunner(ds *dataset.EvalDataset, reportsPath string) *BenchmarkRunner {
	return &BenchmarkRunner{Dataset: ds, ReportsPath: reportsPath, rng: rand.New(rand.NewSource(42))}
}

// Run executa o benchmark completo, salva o relatório em disco e o retorna.
func (r *BenchmarkRunner) Run() model.BenchmarkReport {
	scenarios := r.Dataset.Scenarios()

	react := r.simulateArchitecture("ReactAgent", scenarios, 0.82, 4.2, 1800)
	planExecute := r.simulateArchitecture("PlanExecuteAgent", scenarios, 0.88, 3.1, 2400)
	reflection := r.simulateArchitecture("ReflectionAgent", scenarios, 0.76, 5.8, 3200)

	results := []model.BenchmarkMetrics{react, planExecute, reflection}

	report := model.BenchmarkReport{
		GeneratedAt:    time.Now(),
		TotalScenarios: len(scenarios),
		Results:        results,
		Verdict:        buildVerdict(results),
	}

	r.saveReport(report)
	return report
}

func (r *BenchmarkRunner) simulateArchitecture(architecture string, scenarios []model.EvalScenario, baseCompletion, baseSteps float64, baseTokens int) model.BenchmarkMetrics {
	completionRate := clamp(baseCompletion+(r.rng.Float64()-0.5)*0.10, 0, 1)
	avgSteps := max(1.0, baseSteps+(r.rng.Float64()-0.5)*1.0)
	avgTokens := int(float64(baseTokens) + (r.rng.Float64()-0.5)*400)
	toolSuccessRate := clamp(completionRate+0.05+r.rng.Float64()*0.05, 0, 1)

	totalExpected := 0
	for _, s := range scenarios {
		totalExpected += len(s.ExpectedTools)
	}

	var toolCoverage float64
	if totalExpected > 0 {
		covered := float64(totalExpected) * clamp(completionRate+r.rng.Float64()*0.08, 0, 1)
		toolCoverage = covered / float64(totalExpected)
	}

	return model.BenchmarkMetrics{
		Architecture:       architecture,
		CompletionRate:     round(completionRate),
		AvgSteps:           round(avgSteps),
		AvgTokensEstimated: avgTokens,
		ToolSuccessRate:    round(toolSuccessRate),
		ToolCoverage:       round(toolCoverage),
	}
}

// buildVerdict gera o veredicto identificando a arquitetura vencedora em
// cada critério.
func buildVerdict(results []model.BenchmarkMetrics) map[string]string {
	verdict := map[string]string{}

	best := results[0]
	for _, m := range results[1:] {
		if m.CompletionRate > best.CompletionRate {
			best = m
		}
	}
	verdict["melhor_conclusao"] = best.Architecture

	mostEfficient := results[0]
	for _, m := range results[1:] {
		if m.AvgSteps < mostEfficient.AvgSteps {
			mostEfficient = m
		}
	}
	verdict["mais_eficiente"] = mostEfficient.Architecture

	cheapest := results[0]
	for _, m := range results[1:] {
		if m.AvgTokensEstimated < cheapest.AvgTokensEstimated {
			cheapest = m
		}
	}
	verdict["menor_custo"] = cheapest.Architecture

	bestCoverage := results[0]
	for _, m := range results[1:] {
		if m.ToolCoverage > bestCoverage.ToolCoverage {
			bestCoverage = m
		}
	}
	verdict["melhor_cobertura_ferramentas"] = bestCoverage.Architecture

	recommended := results[0]
	for _, m := range results[1:] {
		if overallScore(m) > overallScore(recommended) {
			recommended = m
		}
	}
	verdict["recomendado_geral"] = recommended.Architecture

	return verdict
}

// overallScore é o score ponderado para comparação geral: 40% conclusão +
// 30% cobertura + 30% toolSuccess.
func overallScore(m model.BenchmarkMetrics) float64 {
	return 0.40*m.CompletionRate + 0.30*m.ToolCoverage + 0.30*m.ToolSuccessRate
}

// saveReport persiste o relatório em JSON no diretório configurado. Falhas
// são logadas mas não interrompem a execução — o relatório em memória ainda
// é retornado, igual à versão Java.
func (r *BenchmarkRunner) saveReport(report model.BenchmarkReport) {
	if err := os.MkdirAll(r.ReportsPath, 0o755); err != nil {
		log.Printf("[BenchmarkRunner] Falha ao salvar relatório: %v", err)
		return
	}

	timestamp := report.GeneratedAt.Format("20060102_150405")
	path := filepath.Join(r.ReportsPath, "benchmark_"+timestamp+".json")

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		log.Printf("[BenchmarkRunner] Falha ao salvar relatório: %v", err)
		return
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		log.Printf("[BenchmarkRunner] Falha ao salvar relatório: %v", err)
	}
}

func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func round(value float64) float64 {
	return float64(int64(value*10000+0.5)) / 10000
}
