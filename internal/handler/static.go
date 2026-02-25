package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

type SPAHandler struct {
	staticDir   string
	indexFile   string
	routePrefix string
	absStaticDir string
}

func NewSPAHandler(staticDir string, routePrefix string) *SPAHandler {
	absDir, err := filepath.Abs(staticDir)
	if err != nil {
		panic(fmt.Sprintf("failed to resolve static dir path %q: %v", staticDir, err))
	}
	return &SPAHandler{
		staticDir:    staticDir,
		indexFile:    "index.html",
		routePrefix:  routePrefix,
		absStaticDir: absDir,
	}
}

func (h *SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Get the path relative to the route prefix
	// First try wildcard param (for Handle("/*")), then fall back to URL path (for NotFound)
	path := chi.URLParam(r, "*")
	if path == "" {
		path = strings.TrimPrefix(r.URL.Path, h.routePrefix)
		path = strings.TrimPrefix(path, "/")
	}

	// Redirect to trailing slash if at root of route (e.g., /portal -> /portal/)
	if path == "" && !strings.HasSuffix(r.URL.Path, "/") {
		http.Redirect(w, r, r.URL.Path+"/", http.StatusMovedPermanently)
		return
	}

	if path == "" {
		path = "/"
	}

	if strings.HasPrefix(path, "api/") {
		http.NotFound(w, r)
		return
	}

	filePath := filepath.Join(h.staticDir, filepath.Clean(path))

	// Path traversal protection
	absFilePath, err := filepath.Abs(filePath)
	if err != nil || (!strings.HasPrefix(absFilePath, h.absStaticDir+string(filepath.Separator)) && absFilePath != h.absStaticDir) {
		log.Warn().Str("path", path).Msg("path traversal attempt blocked")
		http.NotFound(w, r)
		return
	}

	info, err := os.Stat(filePath)
	if err == nil && !info.IsDir() {
		http.ServeFile(w, r, filePath)
		return
	}

	indexPath := filepath.Join(h.staticDir, h.indexFile)
	if _, err := os.Stat(indexPath); err != nil {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, indexPath)
}

func StaticFileServer(staticDir string, routePrefix string) http.Handler {
	return NewSPAHandler(staticDir, routePrefix)
}
