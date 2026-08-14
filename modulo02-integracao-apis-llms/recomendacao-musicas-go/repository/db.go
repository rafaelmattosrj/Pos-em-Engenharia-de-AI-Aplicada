// Package repository persiste conversas e preferências em SQLite —
// substitui o H2 file-based (JPA/Hibernate) usado na versão Java. SQLite
// (via modernc.org/sqlite, driver puro-Go) mantém a mesma característica de
// banco embarcado, com um único arquivo em disco, sem exigir um servidor de
// banco de dados separado.
package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS conversation_messages (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	session_id TEXT NOT NULL,
	role TEXT NOT NULL,
	content TEXT NOT NULL,
	created_at DATETIME NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_conversation_messages_session ON conversation_messages(session_id);

CREATE TABLE IF NOT EXISTS user_preferences (
	user_id TEXT PRIMARY KEY,
	preferences TEXT,
	conversation_summary TEXT
);
`

// Open abre (ou cria) o banco SQLite em path e garante que o schema exista —
// equivalente a spring.jpa.hibernate.ddl-auto=update.
func Open(ctx context.Context, path string) (*sql.DB, error) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("repository: criando diretorio do banco: %w", err)
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("repository: abrindo banco: %w", err)
	}

	if _, err := db.ExecContext(ctx, schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("repository: aplicando schema: %w", err)
	}

	return db, nil
}
