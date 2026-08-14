package evaluator

import (
	"os"
	"testing"

	"agent-evals/dataset"
)

// Cenário 2 de EvalsTest.java: BenchmarkRunner deve gerar relatório com 3
// arquiteturas cognitivas, métricas em faixas realistas e veredicto
// contendo as chaves principais.
func TestRun_GeneratesReportWithThreeArchitectures(t *testing.T) {
	ds := dataset.Load()
	runner := NewBenchmarkRunner(ds, t.TempDir())

	report := runner.Run()

	if report.TotalScenarios != 5 {
		t.Errorf("esperava totalScenarios=5, obteve %d", report.TotalScenarios)
	}
	if len(report.Results) != 3 {
		t.Fatalf("esperava 3 resultados, obteve %d", len(report.Results))
	}

	seen := map[string]bool{}
	for _, m := range report.Results {
		seen[m.Architecture] = true

		if m.CompletionRate < 0 || m.CompletionRate > 1 {
			t.Errorf("%s: completionRate fora da faixa: %f", m.Architecture, m.CompletionRate)
		}
		if m.ToolSuccessRate < 0 || m.ToolSuccessRate > 1 {
			t.Errorf("%s: toolSuccessRate fora da faixa: %f", m.Architecture, m.ToolSuccessRate)
		}
		if m.ToolCoverage < 0 || m.ToolCoverage > 1 {
			t.Errorf("%s: toolCoverage fora da faixa: %f", m.Architecture, m.ToolCoverage)
		}
		if m.AvgSteps <= 0 {
			t.Errorf("%s: avgSteps deveria ser > 0, obteve %f", m.Architecture, m.AvgSteps)
		}
		if m.AvgTokensEstimated <= 0 {
			t.Errorf("%s: avgTokensEstimated deveria ser > 0, obteve %d", m.Architecture, m.AvgTokensEstimated)
		}
	}

	for _, expected := range []string{"ReactAgent", "PlanExecuteAgent", "ReflectionAgent"} {
		if !seen[expected] {
			t.Errorf("esperava arquitetura %q no relatorio", expected)
		}
	}

	for _, key := range []string{"melhor_conclusao", "mais_eficiente", "recomendado_geral"} {
		if _, ok := report.Verdict[key]; !ok {
			t.Errorf("esperava chave %q no veredicto, obteve %v", key, report.Verdict)
		}
	}
}

func TestRun_PersistsReportToDisk(t *testing.T) {
	ds := dataset.Load()
	reportsDir := t.TempDir()
	runner := NewBenchmarkRunner(ds, reportsDir)

	runner.Run()

	entries, err := os.ReadDir(reportsDir)
	if err != nil {
		t.Fatalf("erro lendo diretorio de relatorios: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("esperava 1 arquivo de relatorio salvo, obteve %d", len(entries))
	}
}
