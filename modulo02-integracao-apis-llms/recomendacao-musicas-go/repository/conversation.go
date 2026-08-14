package repository

import (
	"context"
	"database/sql"
	"time"

	"recomendacao-musicas/model"
)

// ConversationRepository persiste mensagens de conversa — equivalente a
// ConversationRepository.java (Spring Data JPA).
type ConversationRepository struct {
	DB *sql.DB
}

// Save insere uma nova mensagem — equivalente a repository.save(new
// ConversationMessage(...)).
func (r *ConversationRepository) Save(ctx context.Context, sessionID, role, content string) error {
	_, err := r.DB.ExecContext(ctx,
		`INSERT INTO conversation_messages (session_id, role, content, created_at) VALUES (?, ?, ?, ?)`,
		sessionID, role, content, time.Now().UTC(),
	)
	return err
}

// FindBySessionID retorna o histórico da sessão em ordem cronológica —
// equivalente a findBySessionIdOrderByCreatedAtAsc.
func (r *ConversationRepository) FindBySessionID(ctx context.Context, sessionID string) ([]model.ConversationMessage, error) {
	rows, err := r.DB.QueryContext(ctx,
		`SELECT id, session_id, role, content, created_at FROM conversation_messages WHERE session_id = ? ORDER BY created_at ASC`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []model.ConversationMessage
	for rows.Next() {
		var m model.ConversationMessage
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

// CountBySessionID retorna quantas mensagens existem na sessão —
// equivalente a countBySessionId.
func (r *ConversationRepository) CountBySessionID(ctx context.Context, sessionID string) (int64, error) {
	var count int64
	err := r.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM conversation_messages WHERE session_id = ?`, sessionID,
	).Scan(&count)
	return count, err
}
