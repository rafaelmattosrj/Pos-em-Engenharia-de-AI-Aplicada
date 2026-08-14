package main

import (
	"encoding/json"
	"io"
	"net/http"

	"roteamento-condicional/graph"
)

type chatRequest struct {
	Question string `json:"question"`
}

const minQuestionLength = 5

// chatHandler retorna o handler para POST /chat. A resposta é texto puro
// (o output processado pelo grafo), espelhando ResponseEntity<String> do
// ChatController.java — não um envelope JSON.
func chatHandler(orchestrator *graph.Orchestrator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "corpo da requisicao invalido", http.StatusBadRequest)
			return
		}

		if len(req.Question) < minQuestionLength {
			http.Error(w, "question deve ter ao menos 5 caracteres", http.StatusBadRequest)
			return
		}

		state, err := orchestrator.Invoke(r.Context(), req.Question)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.WriteString(w, state.Output)
	}
}
