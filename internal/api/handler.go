package api

import (
	"database/sql"
	"net/http"
	"os"
	"path/filepath"

	"pagebound/internal/store"

	"github.com/gin-gonic/gin"
)

type bookResponse struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Creator    *string `json:"creator"`
	Identifier *string `json:"identifier"`
	Publisher  *string `json:"publisher"`
	FilePath   string  `json:"file_path"`
	CoverURL   string  `json:"cover_url"`
	FileSize   int64   `json:"file_size"`
	CreatedAt  string  `json:"created_at"`
}

type listResponse struct {
	Books []bookResponse `json:"books"`
	Total int            `json:"total"`
}

func toResponse(b *store.Book) bookResponse {
	r := bookResponse{
		ID:         b.ID,
		Title:      b.Title,
		Creator:    b.Creator,
		Identifier: b.Identifier,
		Publisher:  b.Publisher,
		FilePath:   b.FilePath,
		FileSize:   b.FileSize,
		CreatedAt:  b.CreatedAt,
	}
	if b.CoverPath != nil && *b.CoverPath != "" {
		r.CoverURL = "/" + *b.CoverPath
	}
	return r
}

func NewHandler(db *sql.DB, booksDir string) *Handler {
	return &Handler{db: db, booksDir: booksDir}
}

type Handler struct {
	db       *sql.DB
	booksDir string
}

func (h *Handler) ListBooks(c *gin.Context) {
	books, err := store.ListBooks(h.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := listResponse{Total: len(books), Books: make([]bookResponse, 0, len(books))}
	for _, b := range books {
		resp.Books = append(resp.Books, toResponse(b))
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) DownloadBook(c *gin.Context) {
	id := c.Param("id")
	book, err := store.GetBook(h.db, id)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if book == nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	epubPath := filepath.Join(h.booksDir, book.FilePath)
	c.Header("Content-Type", "application/epub+zip")
	c.File(epubPath)
}

func (h *Handler) DeleteBook(c *gin.Context) {
	id := c.Param("id")

	book, err := store.GetBook(h.db, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if book == nil {
		c.Status(http.StatusNotFound)
		return
	}

	epubPath := filepath.Join(h.booksDir, book.FilePath)
	os.Remove(epubPath)

	if book.CoverPath != nil && *book.CoverPath != "" {
		coverPath := filepath.Join(h.booksDir, *book.CoverPath)
		os.Remove(coverPath)
	}

	if err := store.DeleteBook(h.db, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
