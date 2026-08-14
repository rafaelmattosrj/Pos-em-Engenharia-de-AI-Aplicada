package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"agent-evals/evaluator"
)

type server struct {
	benchmarkRunner        *evaluator.BenchmarkRunner
	toolSelectionEvaluator *evaluator.ToolSelectionEvaluator
	memoryEvaluator        *evaluator.MemoryEvaluator
	reportsPath            string
}

// runBenchmark trata POST /evals/benchmark.
func (s *server) runBenchmark(w http.ResponseWriter, r *http.Request) {
	report := s.benchmarkRunner.Run()
	writeJSON(w, http.StatusOK, report)
}

// runToolSelectionEval trata POST /evals/tool-selection. Retorna 422 se o
// agente não atingir o threshold mínimo de acurácia.
func (s *server) runToolSelectionEval(w http.ResponseWriter, r *http.Request) {
	report := s.toolSelectionEvaluator.Evaluate(r.Context())
	if !report.Passed {
		writeJSON(w, http.StatusUnprocessableEntity, report)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

// runMemoryEval trata POST /evals/memory-impact.
func (s *server) runMemoryEval(w http.ResponseWriter, r *http.Request) {
	report := s.memoryEvaluator.Evaluate()
	writeJSON(w, http.StatusOK, report)
}

type reportInfo struct {
	Nome         string `json:"nome"`
	TamanhoBytes int64  `json:"tamanhoBytes"`
	CriadoEm     string `json:"criadoEm"`
}

// listReports trata GET /evals/reports — lista os relatórios salvos em
// disco com nome e timestamp de criação.
func (s *server) listReports(w http.ResponseWriter, r *http.Request) {
	entries, err := os.ReadDir(s.reportsPath)
	if err != nil {
		writeJSON(w, http.StatusOK, []reportInfo{})
		return
	}

	var reports []reportInfo
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		reports = append(reports, reportInfo{
			Nome:         entry.Name(),
			TamanhoBytes: info.Size(),
			CriadoEm:     info.ModTime().UTC().Format("2006-01-02T15:04:05.000000Z"),
		})
	}

	sort.Slice(reports, func(i, j int) bool { return reports[i].CriadoEm > reports[j].CriadoEm })

	if reports == nil {
		reports = []reportInfo{}
	}
	writeJSON(w, http.StatusOK, reports)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}
