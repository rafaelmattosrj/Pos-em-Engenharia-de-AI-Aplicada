package main

import (
	"encoding/json"
	"net/http"

	"agendamento-medico/graph"
)

type chatRequest struct {
	Question string `json:"question"`
}

type chatResponse struct {
	Reply string `json:"reply"`
}

const minQuestionLength = 5

// chatHandler retorna o handler para POST /chat — equivalente a
// ChatController.chat.
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

		if len(req.Question) < minQuestionLength {
			writeError(w, http.StatusBadRequest, "question deve ter ao menos 5 caracteres")
			return
		}

		reply, err := orchestrator.Process(r.Context(), req.Question)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(chatResponse{Reply: reply})
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
