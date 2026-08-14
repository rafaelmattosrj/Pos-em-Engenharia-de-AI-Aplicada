package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"recomendacao-musicas/model"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("erro abrindo banco de teste: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestConversationRepository_SaveAndFind(t *testing.T) {
	db := openTestDB(t)
	repo := &ConversationRepository{DB: db}
	ctx := context.Background()

	repo.Save(ctx, "session-1", "user", "ola")
	repo.Save(ctx, "session-1", "assistant", "oi, tudo bem?")
	repo.Save(ctx, "session-2", "user", "mensagem de outra sessao")

	messages, err := repo.FindBySessionID(ctx, "session-1")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("esperava 2 mensagens, obteve %d", len(messages))
	}
	if messages[0].Role != "user" || messages[1].Role != "assistant" {
		t.Errorf("ordem/roles inesperados: %+v", messages)
	}

	count, err := repo.CountBySessionID(ctx, "session-1")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if count != 2 {
		t.Errorf("esperava count=2, obteve %d", count)
	}
}

func TestPreferencesRepository_SaveAndFind(t *testing.T) {
	db := openTestDB(t)
	repo := &PreferencesRepository{DB: db}
	ctx := context.Background()

	_, err := repo.FindByUserID(ctx, "user-1")
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("esperava sql.ErrNoRows para usuario inexistente, obteve %v", err)
	}

	prefs := model.NewUserPreferences("user-1")
	prefs.Preferences = `{"genero":"rock"}`
	if err := repo.Save(ctx, prefs); err != nil {
		t.Fatalf("erro inesperado ao salvar: %v", err)
	}

	found, err := repo.FindByUserID(ctx, "user-1")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if found.Preferences != `{"genero":"rock"}` {
		t.Errorf("preferencias inesperadas: %q", found.Preferences)
	}

	// Upsert: salvar novamente deve atualizar, nao duplicar.
	found.ConversationSummary = "resumo da conversa"
	if err := repo.Save(ctx, found); err != nil {
		t.Fatalf("erro inesperado ao atualizar: %v", err)
	}
	updated, _ := repo.FindByUserID(ctx, "user-1")
	if updated.ConversationSummary != "resumo da conversa" {
		t.Errorf("summary nao foi atualizado: %q", updated.ConversationSummary)
	}
}
