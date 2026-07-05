package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"pagebound/internal/api"
	"pagebound/internal/opds"
	"pagebound/internal/store"
	webdavhook "pagebound/internal/webdav"
	"pagebound/internal/webui"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/webdav"
)

func main() {
	booksDir := os.Getenv("BOOKS_DIR")
	if booksDir == "" {
		booksDir = "./data/books"
	}
	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":8080"
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/library.db"
	}

	if err := os.MkdirAll(booksDir, 0755); err != nil {
		log.Fatalf("failed to create books directory %s: %v", booksDir, err)
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		log.Fatalf("failed to create db directory: %v", err)
	}

	conn, err := store.OpenDB(dbPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer conn.Close()

	inner := webdav.Dir(booksDir)
	hook := webdavhook.NewHookFS(inner, booksDir, conn)

	davHandler := &webdav.Handler{
		Prefix:     "/dav",
		FileSystem: hook,
		LockSystem: webdav.NewMemLS(),
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	apiHandler := api.NewHandler(conn, booksDir)
	r.GET("/api/books", apiHandler.ListBooks)
	r.GET("/api/books/:id/download", apiHandler.DownloadBook)
	r.DELETE("/api/books/:id", apiHandler.DeleteBook)
	r.Static("/covers", filepath.Join(booksDir, "covers"))

	r.GET("/opds/catalog.xml", func(c *gin.Context) {
		scheme := "http"
		if c.Request.TLS != nil {
			scheme = "https"
		}
		baseURL := fmt.Sprintf("%s://%s", scheme, c.Request.Host)
		feed, err := opds.Generate(conn, baseURL)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.Data(http.StatusOK, "application/atom+xml; charset=utf-8", feed)
	})

	assets, _ := fs.Sub(webui.FS, "dist/assets")
	r.StaticFS("/assets", http.FS(assets))
	r.GET("/", func(c *gin.Context) {
		index, _ := webui.FS.ReadFile("dist/index.html")
		c.Data(http.StatusOK, "text/html; charset=utf-8", index)
	})

	r.NoRoute(func(c *gin.Context) {
		index, _ := webui.FS.ReadFile("dist/index.html")
		c.Data(http.StatusOK, "text/html; charset=utf-8", index)
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if strings.HasPrefix(req.URL.Path, "/dav") {
			davHandler.ServeHTTP(w, req)
			return
		}
		r.ServeHTTP(w, req)
	})

	fmt.Printf("Pagebound starting on %s\n", listenAddr)
	fmt.Printf("  WebDAV: http://localhost%s/dav  (Finder: Cmd+K)\n", listenAddr)
	fmt.Printf("  Dashboard: http://localhost%s\n", listenAddr)
	fmt.Printf("  OPDS:    http://localhost%s/opds/catalog.xml\n", listenAddr)
	fmt.Printf("  Books:   %s\n", booksDir)
	fmt.Printf("  DB:      %s\n", dbPath)

	if err := http.ListenAndServe(listenAddr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
