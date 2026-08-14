package gateway

import (
	"encoding/json"
	"net/http"
)

// chatRequest espelha o record ChatRequest(@Size(min=5) String question) do
// ChatController.java.
type chatRequest struct {
	Question string `json:"question"`
}

const minQuestionLength = 5

// ChatHandler retorna o handler HTTP para POST /chat.
func ChatHandler(client *ResilientClient) http.HandlerFunc {
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

		response, err := client.Generate(r.Context(), req.Question)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
