package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

func OpenDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(1)
	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return db, nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS books (
			id          TEXT PRIMARY KEY,
			title       TEXT NOT NULL DEFAULT '',
			creator     TEXT,
			language    TEXT,
			identifier  TEXT,
			description TEXT,
			publisher   TEXT,
			file_path   TEXT NOT NULL UNIQUE,
			cover_path  TEXT,
			file_size   INTEGER NOT NULL DEFAULT 0,
			created_at  TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at  TEXT NOT NULL DEFAULT (datetime('now'))
		);
		CREATE INDEX IF NOT EXISTS idx_books_title   ON books(title);
		CREATE INDEX IF NOT EXISTS idx_books_creator  ON books(creator);
		CREATE INDEX IF NOT EXISTS idx_books_filepath ON books(file_path);
	`)
	return err
}
