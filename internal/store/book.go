package store

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
)

func BookID(relPath string) string {
	h := sha256.Sum256([]byte(relPath))
	return fmt.Sprintf("%x", h)
}

func UpsertBook(db *sql.DB, b *Book) error {
	b.ID = BookID(b.FilePath)
	_, err := db.Exec(`
		INSERT INTO books (id, title, creator, language, identifier, description, publisher, file_path, cover_path, file_size)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(file_path) DO UPDATE SET
			title       = COALESCE(NULLIF(excluded.title, ''), books.title),
			creator     = COALESCE(excluded.creator, books.creator),
			language    = COALESCE(excluded.language, books.language),
			identifier  = COALESCE(excluded.identifier, books.identifier),
			description = COALESCE(excluded.description, books.description),
			publisher   = COALESCE(excluded.publisher, books.publisher),
			cover_path  = COALESCE(excluded.cover_path, books.cover_path),
			file_size   = excluded.file_size,
			updated_at  = datetime('now')
	`, b.ID, b.Title, b.Creator, b.Language, b.Identifier, b.Description, b.Publisher, b.FilePath, b.CoverPath, b.FileSize)
	return err
}

func GetBook(db *sql.DB, id string) (*Book, error) {
	row := db.QueryRow(`
		SELECT id, title, creator, language, identifier, description, publisher, file_path, cover_path, file_size, created_at, updated_at
		FROM books WHERE id = ?
	`, id)
	return scanBook(row)
}

func GetBookByPath(db *sql.DB, filePath string) (*Book, error) {
	row := db.QueryRow(`
		SELECT id, title, creator, language, identifier, description, publisher, file_path, cover_path, file_size, created_at, updated_at
		FROM books WHERE file_path = ?
	`, filePath)
	return scanBook(row)
}

func ListBooks(db *sql.DB) ([]*Book, error) {
	rows, err := db.Query(`
		SELECT id, title, creator, language, identifier, description, publisher, file_path, cover_path, file_size, created_at, updated_at
		FROM books ORDER BY title
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []*Book
	for rows.Next() {
		b, err := scanBook(rows)
		if err != nil {
			return nil, err
		}
		books = append(books, b)
	}
	return books, rows.Err()
}

func DeleteBook(db *sql.DB, id string) error {
	_, err := db.Exec(`DELETE FROM books WHERE id = ?`, id)
	return err
}

type scannable interface {
	Scan(dest ...any) error
}

func scanBook(row scannable) (*Book, error) {
	b := &Book{}
	err := row.Scan(&b.ID, &b.Title, &b.Creator, &b.Language, &b.Identifier, &b.Description, &b.Publisher, &b.FilePath, &b.CoverPath, &b.FileSize, &b.CreatedAt, &b.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return b, err
}
