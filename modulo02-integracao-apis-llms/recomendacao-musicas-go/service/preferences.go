package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"recomendacao-musicas/llm"
	"recomendacao-musicas/model"
	"recomendacao-musicas/repository"
)

const (
	extractPreferencesSystemPrompt = "Extraia preferências musicais desta conversa em JSON.\n" +
		"Inclua: gêneros, artistas, músicas, humor preferido.\n" +
		"Responda APENAS com o JSON, sem explicações.\n"
	summarizeSystemPrompt = "Resuma esta conversa sobre preferências musicais em 3-5 frases."
)

// PreferencesService gerencia preferências musicais persistidas —
// equivalente ao PostgresStore (store) do LangGraph.
type PreferencesService struct {
	Repository *repository.PreferencesRepository
	Client     *llm.ResilientClient
}

// GetOrCreate busca as preferências do usuário, criando um registro vazio se
// ainda não existir — equivalente a getOrCreate.
func (s *PreferencesService) GetOrCreate(ctx context.Context, userID string) (model.UserPreferences, error) {
	prefs, err := s.Repository.FindByUserID(ctx, userID)
	if err == nil {
		return prefs, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return model.UserPreferences{}, err
	}

	prefs = model.NewUserPreferences(userID)
	if err := s.Repository.Save(ctx, prefs); err != nil {
		return model.UserPreferences{}, err
	}
	return prefs, nil
}

// ExtractAndSavePreferences extrai preferências musicais da conversa via LLM
// e as persiste — equivalente a extractAndSavePreferences.
func (s *PreferencesService) ExtractAndSavePreferences(ctx context.Context, userID string, conversationTexts []string) error {
	conversation := strings.Join(conversationTexts, "\n")

	extracted, err := s.Client.Call(ctx, extractPreferencesSystemPrompt, conversation)
	if err != nil {
		return err
	}

	prefs, err := s.GetOrCreate(ctx, userID)
	if err != nil {
		return err
	}
	prefs.Preferences = extracted
	return s.Repository.Save(ctx, prefs)
}

// Summarize sumariza a conversa quando fica muito longa e persiste o resumo
// — equivalente a summarize.
func (s *PreferencesService) Summarize(ctx context.Context, userID string, conversationTexts []string) (string, error) {
	conversation := strings.Join(conversationTexts, "\n")

	summary, err := s.Client.Call(ctx, summarizeSystemPrompt, conversation)
	if err != nil {
		return "", err
	}

	prefs, err := s.GetOrCreate(ctx, userID)
	if err != nil {
		return "", err
	}
	prefs.ConversationSummary = summary
	if err := s.Repository.Save(ctx, prefs); err != nil {
		return "", err
	}
	return summary, nil
}
