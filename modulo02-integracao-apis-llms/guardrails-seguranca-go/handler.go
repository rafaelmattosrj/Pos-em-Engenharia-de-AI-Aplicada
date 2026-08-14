package main

import (
	"encoding/json"
	"net/http"

	"guardrails-seguranca/graph"
)

type chatRequest struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

type chatResponse struct {
	Allowed bool   `json:"allowed"`
	Message string `json:"message"`
}

const minMessageLength = 3

// chatHandler retorna o handler para POST /chat — equivalente a
// ChatController.chat, com o default de username "member".
func chatHandler(orchestrator *graph.Orchestrator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "corpo da requisicao invalido")
			return
		}

		if len(req.Message) < minMessageLength {
			writeError(w, http.StatusBadRequest, "message deve ter ao menos 3 caracteres")
			return
		}

		username := req.Username
		if username == "" {
			username = "member"
		}

		result, err := orchestrator.Process(r.Context(), username, req.Message)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(chatResponse{Allowed: result.Allowed, Message: result.Message})
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
