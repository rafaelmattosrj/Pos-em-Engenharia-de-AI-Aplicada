package main

import (
	"encoding/json"
	"net/http"

	"rag-neo4j-grafos/graph"
	"rag-neo4j-grafos/neo4jservice"
)

type chatRequest struct {
	Question string `json:"question"`
}

type chatResponse struct {
	Answer string `json:"answer"`
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

		answer, err := orchestrator.Query(r.Context(), req.Question)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(chatResponse{Answer: answer})
	}
}

// seedHandler retorna o handler para POST /seed — equivalente a
// ChatController.seed.
func seedHandler(neo4jSvc *neo4jservice.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := neo4jSvc.SeedData(r.Context()); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Write([]byte("Dados inseridos com sucesso"))
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
