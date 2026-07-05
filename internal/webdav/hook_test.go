package webdav

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pagebound/internal/store"

	"golang.org/x/net/webdav"
)

func createTestEpub(t *testing.T, hasCover bool) []byte {
	t.Helper()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	mt, _ := zw.CreateHeader(&zip.FileHeader{Name: "mimetype", Method: zip.Store})
	io.WriteString(mt, "application/epub+zip")

	container, _ := zw.Create("META-INF/container.xml")
	container.Write([]byte(`<?xml version="1.0"?>`))
	container.Write([]byte(`<container xmlns="urn:oasis:names:tc:opendocument:xmlns:container" version="1.0">`))
	container.Write([]byte(`<rootfiles><rootfile full-path="content.opf" media-type="application/oebps-package+xml"/></rootfiles>`))
	container.Write([]byte(`</container>`))

	opf := `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="2.0" unique-identifier="bookid">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Integration Test Book</dc:title>
    <dc:creator>Test Author</dc:creator>
    <dc:language>en</dc:language>`

	if hasCover {
		opf += `<meta name="cover" content="cover-image"/>`
	}

	opf += `</metadata>
  <manifest>`

	if hasCover {
		opf += `<item id="cover-image" href="cover.jpg" media-type="image/jpeg"/>`
	}

	opf += `</manifest>
</package>`

	content, _ := zw.Create("content.opf")
	io.WriteString(content, opf)

	if hasCover {
		img := image.NewRGBA(image.Rect(0, 0, 1, 1))
		img.Set(0, 0, color.RGBA{255, 255, 255, 255})
		var pngBuf bytes.Buffer
		png.Encode(&pngBuf, img)
		cv, _ := zw.Create("cover.jpg")
		cv.Write(pngBuf.Bytes())
	}

	zw.Close()
	return buf.Bytes()
}

func TestHookIngestion(t *testing.T) {
	dir := t.TempDir()
	coversDir := filepath.Join(dir, "covers")
	os.MkdirAll(coversDir, 0755)

	db, err := store.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	hook := NewHookFS(webdav.Dir(dir), dir, db)

	davHandler := &webdav.Handler{
		Prefix:     "",
		FileSystem: hook,
		LockSystem: webdav.NewMemLS(),
	}

	mux := http.NewServeMux()
	mux.Handle("/", davHandler)

	server := &http.Server{
		Addr:    "127.0.0.1:0",
		Handler: mux,
	}
	defer server.Close()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	go server.Serve(ln)

	port := ln.Addr().(*net.TCPAddr).Port
	base := fmt.Sprintf("http://127.0.0.1:%d", port)

	epubData := createTestEpub(t, true)

	req, _ := http.NewRequest("PUT", base+"/testbook.epub", bytes.NewReader(epubData))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("put: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != 201 && resp.StatusCode != 204 {
		t.Fatalf("put status: %s", resp.Status)
	}

	book, err := store.GetBookByPath(db, "testbook.epub")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if book == nil {
		t.Fatal("book not found in store after PUT")
	}
	if book.Title != "Integration Test Book" {
		t.Fatalf("expected title %q, got %q", "Integration Test Book", book.Title)
	}
	if *book.Creator != "Test Author" {
		t.Fatalf("expected creator %q, got %q", "Test Author", *book.Creator)
	}
	if book.CoverPath == nil || *book.CoverPath == "" {
		t.Fatal("expected cover path to be set")
	}
}

func TestHookSkipsNonEpub(t *testing.T) {
	dir := t.TempDir()

	db, err := store.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	hook := NewHookFS(webdav.Dir(dir), dir, db)

	davHandler := &webdav.Handler{
		Prefix:     "",
		FileSystem: hook,
		LockSystem: webdav.NewMemLS(),
	}

	mux := http.NewServeMux()
	mux.Handle("/", davHandler)

	server := &http.Server{Addr: "127.0.0.1:0", Handler: mux}
	defer server.Close()
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	go server.Serve(ln)

	port := ln.Addr().(*net.TCPAddr).Port
	base := fmt.Sprintf("http://127.0.0.1:%d", port)

	req, _ := http.NewRequest("PUT", base+"/readme.txt", strings.NewReader("hello"))
	resp, _ := http.DefaultClient.Do(req)
	resp.Body.Close()

	books, _ := store.ListBooks(db)
	if len(books) != 0 {
		t.Fatal("non-epub file should not create a book entry")
	}
}

func TestHookReuploadUpdates(t *testing.T) {
	dir := t.TempDir()

	db, err := store.OpenDB(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	hook := NewHookFS(webdav.Dir(dir), dir, db)

	davHandler := &webdav.Handler{
		Prefix:     "",
		FileSystem: hook,
		LockSystem: webdav.NewMemLS(),
	}

	mux := http.NewServeMux()
	mux.Handle("/", davHandler)

	server := &http.Server{Addr: "127.0.0.1:0", Handler: mux}
	defer server.Close()
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	go server.Serve(ln)

	port := ln.Addr().(*net.TCPAddr).Port
	base := fmt.Sprintf("http://127.0.0.1:%d", port)

	epubData := createTestEpub(t, false)

	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("PUT", base+"/replace.epub", bytes.NewReader(epubData))
		resp, _ := http.DefaultClient.Do(req)
		resp.Body.Close()
	}

	books, _ := store.ListBooks(db)
	if len(books) != 1 {
		t.Fatalf("expected 1 book after re-uploads, got %d", len(books))
	}
	if books[0].FilePath != "replace.epub" {
		t.Fatalf("expected path replace.epub, got %q", books[0].FilePath)
	}
}
