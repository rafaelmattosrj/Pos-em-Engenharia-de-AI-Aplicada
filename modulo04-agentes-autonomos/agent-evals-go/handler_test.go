package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"agent-evals/dataset"
	"agent-evals/evaluator"
	"agent-evals/model"
)

type failingClient struct{}

func (failingClient) Chat(ctx context.Context, userPrompt string) (string, error) {
	return `{"tool": "wrongTool", "args": {}}`, nil
}

func TestRunToolSelectionEval_Returns422WhenFailing(t *testing.T) {
	srv := &server{
		toolSelectionEvaluator: &evaluator.ToolSelectionEvaluator{Client: failingClient{}, MinAccuracy: 0.80, MaxUnnecessary: 0.10},
	}

	req := httptest.NewRequest(http.MethodPost, "/evals/tool-selection", nil)
	rec := httptest.NewRecorder()

	srv.runToolSelectionEval(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("esperava 422, obteve %d", rec.Code)
	}

	var report model.ToolSelectionReport
	json.NewDecoder(rec.Body).Decode(&report)
	if report.Passed {
		t.Error("esperava passed=false no corpo da resposta")
	}
}

func TestRunBenchmark_Returns200(t *testing.T) {
	srv := &server{benchmarkRunner: evaluator.NewBenchmarkRunner(dataset.Load(), t.TempDir())}

	req := httptest.NewRequest(http.MethodPost, "/evals/benchmark", nil)
	rec := httptest.NewRecorder()

	srv.runBenchmark(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d", rec.Code)
	}
}

func TestListReports_EmptyWhenDirectoryMissing(t *testing.T) {
	srv := &server{reportsPath: "./nao-existe-" + t.Name()}

	req := httptest.NewRequest(http.MethodGet, "/evals/reports", nil)
	rec := httptest.NewRecorder()

	srv.listReports(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d", rec.Code)
	}

	var reports []reportInfo
	json.NewDecoder(rec.Body).Decode(&reports)
	if len(reports) != 0 {
		t.Errorf("esperava lista vazia, obteve %v", reports)
	}
}

func TestListReports_ListsGeneratedBenchmarkFile(t *testing.T) {
	reportsDir := t.TempDir()
	runner := evaluator.NewBenchmarkRunner(dataset.Load(), reportsDir)
	runner.Run()

	srv := &server{reportsPath: reportsDir}

	req := httptest.NewRequest(http.MethodGet, "/evals/reports", nil)
	rec := httptest.NewRecorder()

	srv.listReports(rec, req)

	var reports []reportInfo
	json.NewDecoder(rec.Body).Decode(&reports)
	if len(reports) != 1 {
		t.Fatalf("esperava 1 relatorio listado, obteve %d", len(reports))
	}
}
