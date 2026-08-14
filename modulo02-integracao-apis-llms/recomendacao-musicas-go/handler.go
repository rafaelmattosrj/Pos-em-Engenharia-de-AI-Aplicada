package main

import (
	"encoding/json"
	"net/http"

	"recomendacao-musicas/graph"
)

type chatRequest struct {
	UserID    string `json:"userId"`
	SessionID string `json:"sessionId"`
	Message   string `json:"message"`
}

type chatResponse struct {
	Reply string `json:"reply"`
}

const minMessageLength = 3

// chatHandler retorna o handler para POST /chat — equivalente a
// ChatController.chat, incluindo os defaults "default-user"/"default-session"
// quando userId/sessionId não são informados.
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

		userID := req.UserID
		if userID == "" {
			userID = "default-user"
		}
		sessionID := req.SessionID
		if sessionID == "" {
			sessionID = "default-session"
		}

		reply, err := orchestrator.Chat(r.Context(), userID, sessionID, req.Message)
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
