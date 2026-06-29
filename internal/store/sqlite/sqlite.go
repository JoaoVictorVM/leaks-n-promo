// Package sqlite implementa store.SeenStore sobre SQLite, usando o driver puro
// em Go (modernc.org/sqlite) para não exigir CGO.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/JoaoVictorVM/leaks-n-promo/internal/store"
)

const schema = `CREATE TABLE IF NOT EXISTS seen_leaks (
	url     TEXT PRIMARY KEY,
	seen_at INTEGER NOT NULL
);`

// Store persiste as URLs já notificadas em um arquivo SQLite.
type Store struct {
	db *sql.DB
}

// Garante em tempo de compilação que Store satisfaz a interface.
var _ store.SeenStore = (*Store)(nil)

// Open abre (ou cria) o banco no caminho informado e garante o schema.
func Open(ctx context.Context, path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("abrindo banco sqlite: %w", err)
	}
	// SQLite não lida bem com escritas concorrentes; serializar via uma única
	// conexão evita erros de "database is locked".
	db.SetMaxOpenConns(1)

	if _, err := db.ExecContext(ctx, schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("criando schema: %w", err)
	}
	return &Store{db: db}, nil
}

// Close fecha a conexão com o banco.
func (s *Store) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("fechando banco sqlite: %w", err)
	}
	return nil
}

// Unseen retorna as URLs ainda não registradas, preservando a ordem de entrada.
func (s *Store) Unseen(ctx context.Context, urls []string) ([]string, error) {
	if len(urls) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(urls))
	args := make([]any, len(urls))
	for i, u := range urls {
		placeholders[i] = "?"
		args[i] = u
	}

	// A parte concatenada são apenas placeholders "?"; os valores vão como args
	// parametrizados, sem risco de injeção.
	query := "SELECT url FROM seen_leaks WHERE url IN (" + strings.Join(placeholders, ",") + ")" //nolint:gosec // só placeholders, valores parametrizados
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("consultando urls vistas: %w", err)
	}
	defer func() { _ = rows.Close() }()

	seen := make(map[string]struct{})
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			return nil, fmt.Errorf("lendo linha: %w", err)
		}
		seen[u] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterando resultados: %w", err)
	}

	out := make([]string, 0, len(urls))
	for _, u := range urls {
		if _, ok := seen[u]; !ok {
			out = append(out, u)
		}
	}
	return out, nil
}

// MarkSeen registra as URLs como já vistas. É idempotente (INSERT OR IGNORE).
func (s *Store) MarkSeen(ctx context.Context, urls []string) error {
	if len(urls) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("iniciando transação: %w", err)
	}

	stmt, err := tx.PrepareContext(ctx, "INSERT OR IGNORE INTO seen_leaks(url, seen_at) VALUES(?, ?)")
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("preparando insert: %w", err)
	}
	defer func() { _ = stmt.Close() }()

	now := time.Now().Unix()
	for _, u := range urls {
		if _, err := stmt.ExecContext(ctx, u, now); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("registrando url: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("confirmando transação: %w", err)
	}
	return nil
}
