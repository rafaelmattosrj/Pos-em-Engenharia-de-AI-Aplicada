package repository

import (
	"context"
	"database/sql"

	"recomendacao-musicas/model"
)

// PreferencesRepository persiste as preferências musicais por usuário —
// equivalente a PreferencesRepository.java (Spring Data JPA).
type PreferencesRepository struct {
	DB *sql.DB
}

// FindByUserID busca as preferências de userID. Retorna sql.ErrNoRows se não
// existirem — equivalente a Optional<UserPreferences> vazio.
func (r *PreferencesRepository) FindByUserID(ctx context.Context, userID string) (model.UserPreferences, error) {
	var p model.UserPreferences
	var preferences, summary sql.NullString

	err := r.DB.QueryRowContext(ctx,
		`SELECT user_id, preferences, conversation_summary FROM user_preferences WHERE user_id = ?`, userID,
	).Scan(&p.UserID, &preferences, &summary)
	if err != nil {
		return model.UserPreferences{}, err
	}

	p.Preferences = preferences.String
	p.ConversationSummary = summary.String
	return p, nil
}

// Save insere ou atualiza as preferências de um usuário (upsert) —
// equivalente a repository.save(prefs).
func (r *PreferencesRepository) Save(ctx context.Context, p model.UserPreferences) error {
	_, err := r.DB.ExecContext(ctx,
		`INSERT INTO user_preferences (user_id, preferences, conversation_summary) VALUES (?, ?, ?)
		 ON CONFLICT(user_id) DO UPDATE SET preferences = excluded.preferences, conversation_summary = excluded.conversation_summary`,
		p.UserID, p.Preferences, p.ConversationSummary,
	)
	return err
}
