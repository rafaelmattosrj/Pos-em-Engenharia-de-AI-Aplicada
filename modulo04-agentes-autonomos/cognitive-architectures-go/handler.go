package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"cognitive-architectures/agent"
	"cognitive-architectures/model"
)

type server struct {
	react       *agent.ReactAgent
	planExecute *agent.PlanExecuteAgent
	reflection  *agent.ReflectionAgent
}

// run trata POST /agents/run — equivalente a AgentController.run, incluindo
// o roteamento por "architecture" e o tratamento gracioso (200 + mensagem)
// para arquitetura desconhecida.
func (s *server) run(w http.ResponseWriter, r *http.Request) {
	var req model.AgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "corpo da requisicao invalido")
		return
	}

	if strings.TrimSpace(req.Architecture) == "" {
		writeJSON(w, http.StatusBadRequest, model.AgentResponse{
			Result: "Campo 'architecture' é obrigatório.",
		})
		return
	}
	if strings.TrimSpace(req.Input) == "" {
		writeJSON(w, http.StatusBadRequest, model.AgentResponse{
			Result: "Campo 'input' é obrigatório.",
		})
		return
	}

	var response model.AgentResponse
	var err error

	switch strings.ToLower(strings.TrimSpace(req.Architecture)) {
	case "react":
		response, err = s.react.Execute(r.Context(), req.Input)
	case "plan-execute":
		response, err = s.planExecute.Execute(r.Context(), req.Input)
	case "reflection":
		response, err = s.reflection.Execute(r.Context(), req.Input)
	default:
		response = model.AgentResponse{
			Result: "Arquitetura desconhecida: '" + req.Architecture + "'. Valores válidos: react, plan-execute, reflection.",
		}
	}

	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, response)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
