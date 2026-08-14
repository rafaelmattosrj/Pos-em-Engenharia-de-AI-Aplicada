package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"recomendacao-musicas/llm"
	"recomendacao-musicas/openrouter"
	"recomendacao-musicas/repository"
)

func newTestDB(t *testing.T) *repository.ConversationRepository {
	t.Helper()
	db, err := repository.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("erro abrindo banco: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return &repository.ConversationRepository{DB: db}
}

func TestMemoryService_AddAndRetrieveHistory(t *testing.T) {
	convRepo := newTestDB(t)
	svc := &MemoryService{Repository: convRepo}
	ctx := context.Background()

	svc.AddMessage(ctx, "s1", "user", "quero recomendacoes de rock")
	svc.AddMessage(ctx, "s1", "assistant", "que tal Led Zeppelin?")

	history, err := svc.GetHistory(ctx, "s1")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("esperava 2 mensagens, obteve %d", len(history))
	}

	count, _ := svc.CountMessages(ctx, "s1")
	if count != 2 {
		t.Errorf("esperava count=2, obteve %d", count)
	}
}

func TestPreferencesService_GetOrCreate(t *testing.T) {
	db, _ := newPreferencesTestSetup(t)
	svc := &PreferencesService{Repository: &repository.PreferencesRepository{DB: db}}
	ctx := context.Background()

	prefs, err := svc.GetOrCreate(ctx, "user-1")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if prefs.Preferences != "{}" {
		t.Errorf("esperava preferences='{}' para usuario novo, obteve %q", prefs.Preferences)
	}

	// Segunda chamada deve retornar o mesmo registro, nao recriar.
	prefs2, err := svc.GetOrCreate(ctx, "user-1")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if prefs2.UserID != prefs.UserID {
		t.Errorf("esperava mesmo usuario, obteve %+v", prefs2)
	}
}

func TestPreferencesService_ExtractAndSavePreferences(t *testing.T) {
	db, server := newPreferencesTestSetup(t)
	defer server.Close()

	client := &llm.ResilientClient{
		OpenRouter:     &openrouter.Client{APIKey: "test-key", BaseURL: server.URL, HTTPClient: server.Client()},
		FallbackModels: []string{"model-a"},
	}
	svc := &PreferencesService{Repository: &repository.PreferencesRepository{DB: db}, Client: client}
	ctx := context.Background()

	err := svc.ExtractAndSavePreferences(ctx, "user-1", []string{"user: gosto de rock classico"})
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	prefs, _ := svc.GetOrCreate(ctx, "user-1")
	if prefs.Preferences != `{"genero":"rock classico"}` {
		t.Errorf("preferencias inesperadas: %q", prefs.Preferences)
	}
}

func newPreferencesTestSetup(t *testing.T) (*sql.DB, *httptest.Server) {
	t.Helper()
	db, err := repository.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("erro abrindo banco: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{{"message": map[string]string{"content": `{"genero":"rock classico"}`}}},
		})
	}))

	return db, server
}
