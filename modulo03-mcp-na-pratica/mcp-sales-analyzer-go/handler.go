package main

import (
	"encoding/json"
	"net/http"

	"mcp-sales-analyzer/graph"
	"mcp-sales-analyzer/model"
)

// analyzeHandler retorna o handler para POST /analyze — equivalente a
// AnalysisController.analyze.
func analyzeHandler(orchestrator *graph.Orchestrator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req model.AnalysisRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "corpo da requisicao invalido")
			return
		}

		result, err := orchestrator.Analyze(r.Context(), req)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
