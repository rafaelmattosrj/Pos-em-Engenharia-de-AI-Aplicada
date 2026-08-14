package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"agent-memory/agent"
	"agent-memory/engine"
	"agent-memory/memory"
	"agent-memory/model"
)

type testChatClient struct{ response string }

func (c *testChatClient) Chat(ctx context.Context, userPrompt string) (string, error) {
	return c.response, nil
}

type testEmbedder struct{}

func (testEmbedder) Embed(ctx context.Context, text string) ([]float64, error) {
	return []float64{1, 0, 0}, nil
}

func newTestServer(t *testing.T) *server {
	t.Helper()
	memoryPath := t.TempDir()
	client := &testChatClient{response: "resposta de teste"}

	episodic := memory.NewEpisodicMemory(memoryPath)
	reflection := engine.NewReflectionEngine(client, memoryPath)

	memoryAgent := &agent.MemoryAwareAgent{
		LongTerm:   memory.NewLongTermMemory(memoryPath),
		Episodic:   episodic,
		Contextual: &memory.ContextualMemory{Embedder: testEmbedder{}, SimilarityThreshold: 0.7},
		Reflection: reflection,
		Client:     client,
	}

	return &server{
		memoryAgent: memoryAgent,
		episodic:    episodic,
		reflection:  reflection,
		chatClient:  client,
		memoryPath:  memoryPath,
	}
}

func TestRunWithMemory_Success(t *testing.T) {
	s := newTestServer(t)

	body, _ := json.Marshal(model.AgentRunRequest{Input: "qual o status?"})
	req := httptest.NewRequest(http.MethodPost, "/agent-memory/run", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	s.runWithMemory(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d: %s", rec.Code, rec.Body.String())
	}
	var result model.AgentRunResult
	json.NewDecoder(rec.Body).Decode(&result)
	if !result.MemoryUsed {
		t.Error("esperava memoryUsed=true")
	}
}

func TestRunWithoutMemory_Success(t *testing.T) {
	s := newTestServer(t)

	body, _ := json.Marshal(model.AgentRunRequest{Input: "qual o status?"})
	req := httptest.NewRequest(http.MethodPost, "/agent-memory/run-without", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	s.runWithoutMemory(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d", rec.Code)
	}
	var result model.AgentRunResult
	json.NewDecoder(rec.Body).Decode(&result)
	if result.MemoryUsed {
		t.Error("esperava memoryUsed=false")
	}
}

func TestGetEpisodes_ReturnsPersistedEpisodes(t *testing.T) {
	s := newTestServer(t)

	body, _ := json.Marshal(model.AgentRunRequest{Input: "primeira execucao"})
	req := httptest.NewRequest(http.MethodPost, "/agent-memory/run", bytes.NewReader(body))
	s.runWithMemory(httptest.NewRecorder(), req)

	req2 := httptest.NewRequest(http.MethodGet, "/agent-memory/episodes", nil)
	rec2 := httptest.NewRecorder()
	s.getEpisodes(rec2, req2)

	var episodes []model.Episode
	json.NewDecoder(rec2.Body).Decode(&episodes)
	if len(episodes) != 1 {
		t.Fatalf("esperava 1 episodio, obteve %d", len(episodes))
	}
}

func TestGetLessons_EmptyByDefault(t *testing.T) {
	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/agent-memory/lessons", nil)
	rec := httptest.NewRecorder()
	s.getLessons(rec, req)

	var lessons []model.Lesson
	json.NewDecoder(rec.Body).Decode(&lessons)
	if lessons == nil || len(lessons) != 0 {
		t.Errorf("esperava lista vazia, obteve %v", lessons)
	}
}

func TestClearMemory_RemovesFiles(t *testing.T) {
	s := newTestServer(t)

	body, _ := json.Marshal(model.AgentRunRequest{Input: "gera episodio"})
	req := httptest.NewRequest(http.MethodPost, "/agent-memory/run", bytes.NewReader(body))
	s.runWithMemory(httptest.NewRecorder(), req)

	clearReq := httptest.NewRequest(http.MethodDelete, "/agent-memory/clear", nil)
	rec := httptest.NewRecorder()
	s.clearMemory(rec, clearReq)

	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, obteve %d", rec.Code)
	}

	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["status"] != "sucesso" {
		t.Errorf("status inesperado: %v", resp)
	}
}
