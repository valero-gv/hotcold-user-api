package http

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed swaggerui/index.html swaggerui/swagger.yaml
var swaggerFS embed.FS

func swaggerFileSystem() http.FileSystem {
	sub, err := fs.Sub(swaggerFS, "swaggerui")
	if err != nil {
		// Fallback to full FS; in practice, err shouldn't happen
		return http.FS(swaggerFS)
	}
	return http.FS(sub)
}
