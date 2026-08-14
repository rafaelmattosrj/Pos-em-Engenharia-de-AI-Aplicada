package main

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"agent-memory/agent"
	"agent-memory/engine"
	"agent-memory/memory"
	"agent-memory/model"
)

type server struct {
	memoryAgent *agent.MemoryAwareAgent
	episodic    *memory.EpisodicMemory
	reflection  *engine.ReflectionEngine
	chatClient  agent.ChatClient
	memoryPath  string
}

// runWithMemory trata POST /agent-memory/run — equivalente a
// MemoryAgentController.runWithMemory.
func (s *server) runWithMemory(w http.ResponseWriter, r *http.Request) {
	var req model.AgentRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "corpo da requisicao invalido")
		return
	}

	result, err := s.memoryAgent.Run(r.Context(), req.Input)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// runWithoutMemory trata POST /agent-memory/run-without — equivalente a
// MemoryAgentController.runWithoutMemory.
func (s *server) runWithoutMemory(w http.ResponseWriter, r *http.Request) {
	var req model.AgentRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "corpo da requisicao invalido")
		return
	}

	output, err := s.chatClient.Chat(r.Context(), req.Input)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.AgentRunResult{Output: output, MemoryUsed: false})
}

// getEpisodes trata GET /agent-memory/episodes.
func (s *server) getEpisodes(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, s.episodic.GetEpisodes())
}

// getLessons trata GET /agent-memory/lessons.
func (s *server) getLessons(w http.ResponseWriter, r *http.Request) {
	lessons := s.reflection.Lessons()
	if lessons == nil {
		lessons = []model.Lesson{}
	}
	writeJSON(w, http.StatusOK, lessons)
}

// clearMemory trata DELETE /agent-memory/clear — equivalente a
// MemoryAgentController.clearMemory.
func (s *server) clearMemory(w http.ResponseWriter, r *http.Request) {
	files := []string{"long-term.json", "episodes.json", "lessons.json"}
	var removed []string

	for _, name := range files {
		path := filepath.Join(s.memoryPath, name)
		if err := os.Remove(path); err == nil {
			removed = append(removed, name)
		}
	}

	removedList := ""
	for i, name := range removed {
		if i > 0 {
			removedList += ", "
		}
		removedList += name
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":            "sucesso",
		"arquivosRemovidos": removedList,
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
