package opds

import (
	"database/sql"
	"encoding/xml"
	"sort"
	"time"

	"pagebound/internal/store"
)

type Feed struct {
	XMLName   xml.Name `xml:"feed"`
	XMLNS     string   `xml:"xmlns,attr"`
	XMLNSOPDS string   `xml:"xmlns:opds,attr"`
	Title     string   `xml:"title"`
	ID        string   `xml:"id"`
	Updated   string   `xml:"updated"`
	Author    Author   `xml:"author"`
	Link      []Link   `xml:"link"`
	Entry     []Entry  `xml:"entry"`
}

type Author struct {
	Name string `xml:"name"`
}

type Link struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
	Type string `xml:"type,attr,omitempty"`
}

type Entry struct {
	Title   string   `xml:"title"`
	ID      string   `xml:"id"`
	Updated string   `xml:"updated"`
	Summary string   `xml:"summary,omitempty"`
	Author  *Author  `xml:"author,omitempty"`
	Link    []Link   `xml:"link"`
	Content *Content `xml:"content,omitempty"`
}

type Content struct {
	Type string `xml:"type,attr"`
	Src  string `xml:"src,attr"`
}

func Generate(db *sql.DB, baseURL string) ([]byte, error) {
	books, err := store.ListBooks(db)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC().Format(time.RFC3339)

	feed := &Feed{
		XMLNS:     "http://www.w3.org/2005/Atom",
		XMLNSOPDS: "http://opds-spec.org/2010/catalog",
		Title:     "Pagebound",
		ID:        baseURL + "/opds/catalog.xml",
		Updated:   now,
		Author:    Author{Name: "Pagebound"},
		Link: []Link{
			{Rel: "self", Href: baseURL + "/opds/catalog.xml", Type: "application/atom+xml"},
			{Rel: "start", Href: baseURL + "/opds/catalog.xml", Type: "application/atom+xml"},
		},
		Entry: make([]Entry, 0, len(books)),
	}

	for _, b := range books {
		dlURL := baseURL + "/api/books/" + b.ID + "/download"

		links := []Link{
			{Rel: "http://opds-spec.org/acquisition", Href: dlURL, Type: "application/epub+zip"},
		}

		if b.CoverPath != nil && *b.CoverPath != "" {
			links = append(links, Link{
				Rel:  "http://opds-spec.org/thumbnail",
				Href: baseURL + "/" + *b.CoverPath,
			})
		}

		entry := Entry{
			Title:   b.Title,
			ID:      baseURL + "/books/" + b.ID,
			Updated: b.UpdatedAt,
			Summary: buildSummary(b),
			Author:  entryAuthor(b),
			Link:    links,
			Content: &Content{Type: "application/epub+zip", Src: dlURL},
		}
		feed.Entry = append(feed.Entry, entry)
	}

	sort.Slice(feed.Entry, func(i, j int) bool {
		return feed.Entry[i].Title < feed.Entry[j].Title
	})

	output, err := xml.MarshalIndent(feed, "", "  ")
	if err != nil {
		return nil, err
	}

	return append([]byte(xml.Header), output...), nil
}

func buildSummary(b *store.Book) string {
	var s string
	if b.Publisher != nil && *b.Publisher != "" {
		s = "Publisher: " + *b.Publisher
	}
	if b.Identifier != nil && *b.Identifier != "" {
		if s != "" {
			s += " | "
		}
		s += "ISBN: " + *b.Identifier
	}
	return s
}

func entryAuthor(b *store.Book) *Author {
	if b.Creator != nil && *b.Creator != "" {
		return &Author{Name: *b.Creator}
	}
	return nil
}
