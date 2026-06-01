package whiteboard

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed static/*
var staticEmbed embed.FS

// StaticFS provides a file system for the embedded static files,
// rooted at the "static" directory.
func StaticFS() http.FileSystem {
	fsys, err := fs.Sub(staticEmbed, "static")
	if err != nil {
		panic(err)
	}
	return http.FS(fsys)
}
