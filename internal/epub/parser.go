package epub

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/barsanuphe/epubgo"
)

type Metadata struct {
	Title       string
	Creator     string
	Language    string
	Identifier  string
	Description string
	Publisher   string
}

func Parse(path string) (*Metadata, error) {
	book, err := epubgo.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open epub: %w", err)
	}
	defer book.Close()

	m := &Metadata{}
	if titles, err := book.Metadata("title"); err == nil && len(titles) > 0 {
		m.Title = titles[0]
	}
	if creators, err := book.Metadata("creator"); err == nil && len(creators) > 0 {
		m.Creator = creators[0]
	}
	if langs, err := book.Metadata("language"); err == nil && len(langs) > 0 {
		m.Language = langs[0]
	}
	if ids, err := book.Metadata("identifier"); err == nil && len(ids) > 0 {
		m.Identifier = ids[0]
	}
	if descs, err := book.Metadata("description"); err == nil && len(descs) > 0 {
		m.Description = descs[0]
	}
	if pubs, err := book.Metadata("publisher"); err == nil && len(pubs) > 0 {
		m.Publisher = pubs[0]
	}
	return m, nil
}

func ExtractCover(epubPath, destDir, bookID string) (string, error) {
	book, err := epubgo.Open(epubPath)
	if err != nil {
		return "", fmt.Errorf("open for cover: %w", err)
	}
	defer book.Close()

	// EPUB2: <meta name="cover" content="item-id"/>
	if elems, err := book.MetadataElement("meta"); err == nil {
		for _, e := range elems {
			if e.Attr["name"] == "cover" {
				p, err := readCoverFile(book, e.Content, destDir, bookID)
				if err == nil {
					return p, nil
				}
			}
		}
	}

	// EPUB3 / fallback: try common cover IDs
	for _, id := range []string{"cover", "cover-image", "coverImage"} {
		p, err := readCoverFile(book, id, destDir, bookID)
		if err == nil {
			return p, nil
		}
	}

	return "", fmt.Errorf("no cover found")
}

func readCoverFile(book *epubgo.Epub, id, destDir, bookID string) (string, error) {
	rc, err := book.OpenFileId(id)
	if err != nil {
		return "", err
	}
	defer rc.Close()

	header := make([]byte, 512)
	n, _ := io.ReadFull(rc, header)

	ext := detectExt(header[:n])

	outPath := filepath.Join(destDir, bookID+ext)
	f, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if n > 0 {
		f.Write(header[:n])
	}
	io.Copy(f, rc)

	return outPath, nil
}

func detectExt(buf []byte) string {
	ct := http.DetectContentType(buf)
	switch {
	case strings.Contains(ct, "jpeg"):
		return ".jpg"
	case strings.Contains(ct, "png"):
		return ".png"
	case strings.Contains(ct, "gif"):
		return ".gif"
	case strings.Contains(ct, "webp"):
		return ".webp"
	default:
		return ".img"
	}
}
