// Package model define os tipos de domínio persistidos — equivalente às
// entidades JPA ConversationMessage e UserPreferences.
package model

import "time"

// ConversationMessage é uma mensagem persistida da conversa — equivalente ao
// checkpoint do LangGraph, salvo aqui em SQLite.
type ConversationMessage struct {
	ID        int64
	SessionID string
	Role      string // "user" | "assistant"
	Content   string
	CreatedAt time.Time
}

// UserPreferences são as preferências musicais extraídas do usuário —
// equivalente ao PostgresStore do LangGraph, salvo aqui em SQLite.
type UserPreferences struct {
	UserID              string
	Preferences         string // JSON com preferências extraídas
	ConversationSummary string
}

// NewUserPreferences cria as preferências iniciais de um usuário —
// equivalente ao construtor UserPreferences(userId).
func NewUserPreferences(userID string) UserPreferences {
	return UserPreferences{UserID: userID, Preferences: "{}"}
}
