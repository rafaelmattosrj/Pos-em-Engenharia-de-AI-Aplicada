// Agent Evals — framework de avaliação de agentes: benchmark comparativo
// entre 3 arquiteturas cognitivas, avaliação de seleção de ferramentas e
// avaliação de impacto de memória. Porte Go do projeto Spring Boot
// agent-evals-java.
package main

import (
	"fmt"
	"log"
	"net/http"

	"agent-evals/dataset"
	"agent-evals/evaluator"
	"agent-evals/openai"
)

func main() {
	loadDotEnv(".env")

	port := getEnv("PORT", "8080")
	apiKey := getEnv("OPENAI_API_KEY", "sk-placeholder")
	model := getEnv("OPENAI_MODEL", "gpt-4o-mini")
	reportsPath := getEnv("REPORTS_PATH", "./reports")
	minAccuracy := getEnvFloat("TOOL_SELECTION_MIN_ACCURACY", 0.80)
	maxUnnecessary := getEnvFloat("TOOL_SELECTION_MAX_UNNECESSARY", 0.10)

	ds := dataset.Load()
	client := openai.NewClient(apiKey, model)

	srv := &server{
		benchmarkRunner: evaluator.NewBenchmarkRunner(ds, reportsPath),
		toolSelectionEvaluator: &evaluator.ToolSelectionEvaluator{
			Client: client, MinAccuracy: minAccuracy, MaxUnnecessary: maxUnnecessary,
		},
		memoryEvaluator: evaluator.NewMemoryEvaluator(),
		reportsPath:     reportsPath,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /evals/benchmark", srv.runBenchmark)
	mux.HandleFunc("POST /evals/tool-selection", srv.runToolSelectionEval)
	mux.HandleFunc("POST /evals/memory-impact", srv.runMemoryEval)
	mux.HandleFunc("GET /evals/reports", srv.listReports)

	addr := fmt.Sprintf(":%s", port)
	log.Printf("agent-evals ouvindo em %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
