package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"agent-loop-framework/agentloop"
)

// runHandler retorna o handler para POST /agent/run — equivalente a
// AgentController.run.
func runHandler(loop *agentloop.AgentLoop) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSONError(w, http.StatusBadRequest, "corpo da requisicao invalido")
			return
		}

		input := strings.TrimSpace(body["input"])
		if input == "" {
			writeJSONError(w, http.StatusBadRequest, "Campo 'input' é obrigatório e não pode estar vazio")
			return
		}

		trace := loop.Run(r.Context(), body["input"])

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"result": trace.FinalResult,
			"trace":  trace,
		})
	}
}

// healthHandler retorna o handler para GET /agent/health.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "UP", "service": "agent-loop-framework"})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
