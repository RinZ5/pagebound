package webui

import "embed"

//go:embed dist/index.html dist/assets/*
var FS embed.FS
