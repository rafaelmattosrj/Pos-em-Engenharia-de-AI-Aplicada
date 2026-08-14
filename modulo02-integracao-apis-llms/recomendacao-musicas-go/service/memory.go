// Package service implementa a lógica de memória de conversa e preferências
// musicais — equivalente a MemoryService.java e PreferencesService.java.
package service

import (
	"context"

	"recomendacao-musicas/model"
	"recomendacao-musicas/repository"
)

// MemoryService gerencia o histórico de conversas persistido — equivalente
// ao PostgresSaver (checkpointer) do LangGraph.
type MemoryService struct {
	Repository *repository.ConversationRepository
}

func (s *MemoryService) AddMessage(ctx context.Context, sessionID, role, content string) error {
	return s.Repository.Save(ctx, sessionID, role, content)
}

func (s *MemoryService) GetHistory(ctx context.Context, sessionID string) ([]model.ConversationMessage, error) {
	return s.Repository.FindBySessionID(ctx, sessionID)
}

func (s *MemoryService) CountMessages(ctx context.Context, sessionID string) (int64, error) {
	return s.Repository.CountBySessionID(ctx, sessionID)
}
