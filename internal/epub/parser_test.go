package epub

import (
	"archive/zip"
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func writeEpub(t *testing.T, dir, opfContent string, coverData []byte) string {
	t.Helper()
	path := filepath.Join(dir, "test.epub")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create epub: %v", err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)

	mt, err := zw.CreateHeader(&zip.FileHeader{
		Name:   "mimetype",
		Method: zip.Store,
	})
	if err != nil {
		t.Fatalf("mimetype header: %v", err)
	}
	if _, err := mt.Write([]byte("application/epub+zip")); err != nil {
		t.Fatalf("mimetype write: %v", err)
	}

	container, _ := zw.Create("META-INF/container.xml")
	container.Write([]byte(`<?xml version="1.0"?>`))
	container.Write([]byte(`<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">`))
	container.Write([]byte(`<rootfiles><rootfile full-path="content.opf" media-type="application/oebps-package+xml"/></rootfiles>`))
	container.Write([]byte(`</container>`))

	opf, _ := zw.Create("content.opf")
	opf.Write([]byte(opfContent))

	if coverData != nil {
		cv, _ := zw.Create("cover.jpg")
		cv.Write(coverData)
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return path
}

func defaultOPF() string {
	return `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="2.0" unique-identifier="bookid">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Test Book Title</dc:title>
    <dc:creator>Test Author</dc:creator>
    <dc:language>en</dc:language>
    <dc:identifier>978-3-16-148410-0</dc:identifier>
    <dc:description>A test book for unit testing</dc:description>
    <dc:publisher>Test Publisher</dc:publisher>
    <meta name="cover" content="cover-image"/>
  </metadata>
  <manifest>
    <item id="cover-image" href="cover.jpg" media-type="image/jpeg"/>
  </manifest>
</package>`
}

func genMinimalJPEG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{255, 255, 255, 255})
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode image: %v", err)
	}
	// Write as .jpg even though it's PNG data — the test just needs a binary file
	return buf.Bytes()
}

func TestParse(t *testing.T) {
	dir := t.TempDir()
	path := writeEpub(t, dir, defaultOPF(), genMinimalJPEG(t))

	m, err := Parse(path)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if m.Title != "Test Book Title" {
		t.Fatalf("title: expected %q, got %q", "Test Book Title", m.Title)
	}
	if m.Creator != "Test Author" {
		t.Fatalf("creator: expected %q, got %q", "Test Author", m.Creator)
	}
	if m.Language != "en" {
		t.Fatalf("language: expected %q, got %q", "en", m.Language)
	}
	if m.Identifier != "978-3-16-148410-0" {
		t.Fatalf("identifier: expected %q, got %q", "978-3-16-148410-0", m.Identifier)
	}
	if m.Description != "A test book for unit testing" {
		t.Fatalf("description: expected %q, got %q", "A test book for unit testing", m.Description)
	}
	if m.Publisher != "Test Publisher" {
		t.Fatalf("publisher: expected %q, got %q", "Test Publisher", m.Publisher)
	}
}

func TestParseMinimal(t *testing.T) {
	opf := `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="2.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>Minimal</dc:title>
  </metadata>
  <manifest/>
</package>`
	dir := t.TempDir()
	path := writeEpub(t, dir, opf, nil)

	m, err := Parse(path)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if m.Title != "Minimal" {
		t.Fatalf("title: expected %q, got %q", "Minimal", m.Title)
	}
	if m.Creator != "" {
		t.Fatalf("expected empty creator, got %q", m.Creator)
	}
}

func TestExtractCover(t *testing.T) {
	dir := t.TempDir()
	coverData := genMinimalJPEG(t)
	path := writeEpub(t, dir, defaultOPF(), coverData)

	coverDest := filepath.Join(dir, "covers")
	os.MkdirAll(coverDest, 0755)

	coverPath, err := ExtractCover(path, coverDest, "book1")
	if err != nil {
		t.Fatalf("extract cover: %v", err)
	}

	if _, err := os.Stat(coverPath); os.IsNotExist(err) {
		t.Fatalf("cover file not created at %s", coverPath)
	}
}

func TestExtractCoverNoMeta(t *testing.T) {
	opf := `<?xml version="1.0"?>
<package xmlns="http://www.idpf.org/2007/opf" version="2.0">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>No Cover</dc:title>
  </metadata>
  <manifest/>
</package>`
	dir := t.TempDir()
	path := writeEpub(t, dir, opf, nil)

	coverDest := filepath.Join(dir, "covers")
	os.MkdirAll(coverDest, 0755)

	_, err := ExtractCover(path, coverDest, "book1")
	if err == nil {
		t.Fatal("expected error for epub without cover")
	}
}
