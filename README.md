# Pagebound

A cozy, self-hosted e-book library server for your home network.
Built for the Raspberry Pi. Lightweight, fast, and delightfully simple.

## Features

- **Drag-and-drop ingestion** — connect via Finder or WebDAV client and drop in `.epub` files
- **Automatic metadata extraction** — title, author, ISBN, publisher parsed on-the-fly, no manual tagging
- **Cover image extraction** — covers displayed in the dashboard and OPDS catalog
- **OPDS 2.0 feed** — browse and download books from any OPDS-compatible e-reader (KOReader, FBReader, etc.)
- **Cozy web dashboard** — warm-toned bookshelf interface to browse, search, download, and manage your library
- **Single binary** — Go backend with embedded Vue 3 frontend, no runtime dependencies beyond the binary
- **SQLite storage** — zero-configuration, zero-maintenance embedded database

## Connect

| Method | URL | Client |
|--------|-----|--------|
| WebDAV | `http://<ip>:8080/dav` | macOS Finder (`Cmd+K`), Cyberduck, Transmit |
| Dashboard | `http://<ip>:8080` | Any web browser |
| OPDS | `http://<ip>:8080/opds/catalog.xml` | KOReader, FBReader, Moon+ Reader, Cool Reader |

## License

MIT
