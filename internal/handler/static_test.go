package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSPAHandler(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir, err := os.MkdirTemp("", "spa-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create index.html
	indexContent := "<!DOCTYPE html><html><body>Index</body></html>"
	err = os.WriteFile(filepath.Join(tmpDir, "index.html"), []byte(indexContent), 0644)
	require.NoError(t, err)

	// Create a CSS file
	cssContent := "body { color: black; }"
	err = os.WriteFile(filepath.Join(tmpDir, "styles.css"), []byte(cssContent), 0644)
	require.NoError(t, err)

	// Create a JS file
	jsContent := "console.log('hello');"
	err = os.WriteFile(filepath.Join(tmpDir, "app.js"), []byte(jsContent), 0644)
	require.NoError(t, err)

	handler := NewSPAHandler(tmpDir, "")

	t.Run("serves index.html for root path", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Index")
	})

	t.Run("serves static files", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/styles.css", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "color: black")
	})

	t.Run("serves JS files", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/app.js", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "console.log")
	})

	t.Run("falls back to index.html for unknown paths (SPA routing)", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/dashboard/settings", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "Index")
	})

	t.Run("returns 404 for /api/ paths", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/users", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})

	t.Run("returns 404 for /api prefix", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestSPAHandler_NoIndexFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "spa-test-empty")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	handler := NewSPAHandler(tmpDir, "")

	t.Run("returns 404 when index.html is missing", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestStaticFileServer(t *testing.T) {
	t.Run("returns SPAHandler", func(t *testing.T) {
		handler := StaticFileServer("/tmp/test", "/portal")
		assert.NotNil(t, handler)
		_, ok := handler.(*SPAHandler)
		assert.True(t, ok)
	})
}
