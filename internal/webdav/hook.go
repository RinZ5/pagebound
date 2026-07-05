package webdav

import (
	"context"
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"strings"

	"pagebound/internal/epub"
	"pagebound/internal/store"

	"golang.org/x/net/webdav"
)

type HookFS struct {
	webdav.FileSystem
	dir  string
	conn *sql.DB
}

func NewHookFS(fs webdav.FileSystem, dir string, conn *sql.DB) *HookFS {
	return &HookFS{FileSystem: fs, dir: dir, conn: conn}
}

func (h *HookFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	f, err := h.FileSystem.OpenFile(ctx, name, flag, perm)
	if err != nil {
		return nil, err
	}
	if strings.HasSuffix(strings.ToLower(name), ".epub") && flag&os.O_CREATE != 0 {
		rel := strings.TrimPrefix(name, "/")
		return &hookFile{
			File:   f,
			abs:    filepath.Join(h.dir, rel),
			rel:    rel,
			dir:    h.dir,
			covers: filepath.Join(h.dir, "covers"),
			conn:   h.conn,
		}, nil
	}
	return f, nil
}

type hookFile struct {
	webdav.File
	abs    string
	rel    string
	dir    string
	covers string
	conn   *sql.DB
}

func (hf *hookFile) Close() error {
	err := hf.File.Close()
	if err != nil {
		return err
	}
	hf.process()
	return nil
}

func (hf *hookFile) process() {
	m, err := epub.Parse(hf.abs)
	if err != nil {
		log.Printf("webdav: skip %s (not a valid epub): %v", hf.rel, err)
		return
	}

	fi, err := os.Stat(hf.abs)
	if err != nil {
		log.Printf("webdav: stat %s: %v", hf.rel, err)
		return
	}

	bookID := store.BookID(hf.rel)

	os.MkdirAll(hf.covers, 0755)

	var coverRel string
	coverAbs, err := epub.ExtractCover(hf.abs, hf.covers, bookID)
	if err == nil {
		rel, _ := filepath.Rel(hf.dir, coverAbs)
		coverRel = rel
	}

	book := &store.Book{
		Title:      m.Title,
		Creator:    maybe(m.Creator),
		Language:   maybe(m.Language),
		Identifier: maybe(m.Identifier),
		Publisher:  maybe(m.Publisher),
		FilePath:   hf.rel,
		CoverPath:  maybe(coverRel),
		FileSize:   fi.Size(),
	}

	if err := store.UpsertBook(hf.conn, book); err != nil {
		log.Printf("webdav: upsert %s: %v", hf.rel, err)
		return
	}

	log.Printf("webdav: indexed %q by %s", m.Title, hf.rel)
}

func maybe(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
