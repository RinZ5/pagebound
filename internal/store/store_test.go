package store

import (
	"database/sql"
	"testing"
)

type BookDB = sql.DB

func openTestDB(t *testing.T) *BookDB {
	t.Helper()
	db, err := OpenDB(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestUpsertAndGetBook(t *testing.T) {
	db := openTestDB(t)
	title := "Test Book"
	creator := "Test Author"
	book := &Book{
		Title:    title,
		Creator:  &creator,
		FilePath: "test.epub",
		FileSize: 12345,
	}

	if err := UpsertBook(db, book); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	got, err := GetBook(db, book.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got == nil {
		t.Fatal("book not found")
	}
	if got.Title != title {
		t.Fatalf("expected title %q, got %q", title, got.Title)
	}
	if *got.Creator != creator {
		t.Fatalf("expected creator %q, got %q", creator, *got.Creator)
	}
	if got.FilePath != "test.epub" {
		t.Fatalf("expected file_path test.epub, got %q", got.FilePath)
	}
}

func TestUpsertUpdatesExisting(t *testing.T) {
	db := openTestDB(t)
	creator := "Original"
	book := &Book{
		Title:    "Book",
		Creator:  &creator,
		FilePath: "test.epub",
		FileSize: 100,
	}
	if err := UpsertBook(db, book); err != nil {
		t.Fatalf("first upsert: %v", err)
	}

	newCreator := "Updated"
	book2 := &Book{
		Title:    "Book",
		Creator:  &newCreator,
		FilePath: "test.epub",
		FileSize: 200,
	}
	if err := UpsertBook(db, book2); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	got, _ := GetBook(db, book.ID)
	if got == nil {
		t.Fatal("book not found after upsert")
	}
	if *got.Creator != newCreator {
		t.Fatalf("expected updated creator %q, got %q", newCreator, *got.Creator)
	}
	if got.FileSize != 200 {
		t.Fatalf("expected updated file_size 200, got %d", got.FileSize)
	}
}

func TestListBooks(t *testing.T) {
	db := openTestDB(t)
	books := []*Book{
		{Title: "Beta", FilePath: "b.epub", FileSize: 1},
		{Title: "Alpha", FilePath: "a.epub", FileSize: 2},
		{Title: "Gamma", FilePath: "g.epub", FileSize: 3},
	}
	for _, b := range books {
		if err := UpsertBook(db, b); err != nil {
			t.Fatalf("upsert %s: %v", b.Title, err)
		}
	}

	got, err := ListBooks(db)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 books, got %d", len(got))
	}
	if got[0].Title != "Alpha" {
		t.Fatalf("expected Alpha first (sorted), got %s", got[0].Title)
	}
	if got[1].Title != "Beta" {
		t.Fatalf("expected Beta second, got %s", got[1].Title)
	}
}

func TestDeleteBook(t *testing.T) {
	db := openTestDB(t)
	book := &Book{Title: "Delete Me", FilePath: "delete.epub", FileSize: 1}
	UpsertBook(db, book)

	if err := DeleteBook(db, book.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	got, _ := GetBook(db, book.ID)
	if got != nil {
		t.Fatal("book should be nil after delete")
	}
}
