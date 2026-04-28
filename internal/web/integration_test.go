package web

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/knivey/dave-web/internal/db"
	"github.com/knivey/dave-web/internal/models"
)

//go:embed all:migrations
var testMigrationsFS embed.FS

func newIntegrationDB(t *testing.T) *db.DB {
	t.Helper()
	database, err := db.New(":memory:", testMigrationsFS)
	require.NoError(t, err)
	t.Cleanup(func() { database.Close() })
	return database
}

func integrationServer(t *testing.T, database *db.DB) *Server {
	t.Helper()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())

	tmpl := `{{define "list.html"}}{{if .Stats}}<div class="stats-bar"><span>{{.Stats.Total}}</span><span>{{.Stats.Active}}</span><span>{{.Stats.Expired}}</span></div>{{end}}{{end}}`
	tmpl += `{{define "view.html"}}<h1>{{.Title}}</h1><div>{{.RenderedHTML}}</div><a href="{{.RawURL}}">raw</a>{{end}}`
	r.SetHTMLTemplate(template.Must(template.New("").Parse(tmpl)))

	s := &Server{
		Router:  r,
		DB:      database,
		baseURL: "http://localhost:8080",
	}

	r.GET("/", s.handleList)
	r.GET("/p/:slug", s.handleView)
	r.GET("/p/:slug/raw", s.handleRaw)
	r.POST("/api/pastes", s.requireAPIKey, s.handleAPICreate)

	return s
}

func TestIntegration_FullFlow(t *testing.T) {
	database := newIntegrationDB(t)
	srv := integrationServer(t, database)

	ak, err := database.CreateAPIKey("test")
	require.NoError(t, err)

	body := `{"content":"# Hello\n\n**world**","title":"Test Paste"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/pastes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", ak.Key)
	srv.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp createPasteResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp.Slug)
	assert.Contains(t, resp.URL, resp.Slug)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/p/"+resp.Slug, nil)
	srv.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "<strong>world</strong>")

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/p/"+resp.Slug+"/raw", nil)
	srv.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "# Hello\n\n**world**", w.Body.String())

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/", nil)
	srv.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "1")
}

func TestIntegration_UnauthCreate(t *testing.T) {
	database := newIntegrationDB(t)
	srv := integrationServer(t, database)

	body := `{"content":"test"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/pastes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	srv.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestIntegration_BadKeyCreate(t *testing.T) {
	database := newIntegrationDB(t)
	srv := integrationServer(t, database)

	body := `{"content":"test"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/pastes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "bad_key")
	srv.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestIntegration_ViewNotFound(t *testing.T) {
	database := newIntegrationDB(t)
	srv := integrationServer(t, database)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/p/nonexistent", nil)
	srv.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestIntegration_RawNotFound(t *testing.T) {
	database := newIntegrationDB(t)
	srv := integrationServer(t, database)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/p/nonexistent/raw", nil)
	srv.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestIntegration_ListEmpty(t *testing.T) {
	database := newIntegrationDB(t)
	srv := integrationServer(t, database)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	srv.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "0")
}

func TestIntegration_ListShowsStats(t *testing.T) {
	database := newIntegrationDB(t)
	srv := integrationServer(t, database)

	ak, err := database.CreateAPIKey("test")
	require.NoError(t, err)

	for i := 0; i < 3; i++ {
		body := fmt.Sprintf(`{"content":"paste %d","title":"Paste %d"}`, i, i)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/pastes", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", ak.Key)
		srv.Router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	srv.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "3")
}

func TestIntegration_RevokedKey(t *testing.T) {
	database := newIntegrationDB(t)
	srv := integrationServer(t, database)

	ak, err := database.CreateAPIKey("test")
	require.NoError(t, err)

	body := `{"content":"test"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/pastes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", ak.Key)
	srv.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	require.NoError(t, database.DeleteAPIKey(ak.Key))

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/pastes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", ak.Key)
	srv.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestIntegration_PasteRender(t *testing.T) {
	database := newIntegrationDB(t)
	srv := integrationServer(t, database)

	paste := &models.Paste{
		Slug:      "render1",
		Title:     "Render Test",
		Content:   "# Header\n\n**bold** and _italic_\n\n```go\nfmt.Println()\n```",
		Rendered:  "",
		CreatedAt: time.Now(),
		Language:  "markdown",
	}
	require.NoError(t, database.CreatePaste(paste))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/p/render1", nil)
	srv.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestIntegration_MultiplePastes(t *testing.T) {
	database := newIntegrationDB(t)
	srv := integrationServer(t, database)

	ak, err := database.CreateAPIKey("test")
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		body := fmt.Sprintf(`{"content":"paste %d","title":"Paste %d"}`, i, i)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/pastes", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", ak.Key)
		srv.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	srv.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "5")
}

func TestIntegration_PasteWithEmptyTitle(t *testing.T) {
	database := newIntegrationDB(t)
	srv := integrationServer(t, database)

	ak, err := database.CreateAPIKey("test")
	require.NoError(t, err)

	body := `{"content":"first line content\nsecond line"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/pastes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", ak.Key)
	srv.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var resp createPasteResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/p/"+resp.Slug, nil)
	srv.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Untitled")
}

func TestIntegration_ContentTypes(t *testing.T) {
	database := newIntegrationDB(t)
	srv := integrationServer(t, database)

	ak, err := database.CreateAPIKey("test")
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/pastes", bytes.NewBufferString(`{"content":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", ak.Key)
	srv.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")

	var resp createPasteResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/p/"+resp.Slug, nil)
	srv.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/p/"+resp.Slug+"/raw", nil)
	srv.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")
}

func TestIntegration_DisallowedMethods(t *testing.T) {
	database := newIntegrationDB(t)
	srv := integrationServer(t, database)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/pastes", nil)
	srv.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/p/someslug", nil)
	srv.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestIntegration_CreateWithTTL(t *testing.T) {
	database := newIntegrationDB(t)
	srv := integrationServer(t, database)

	ak, err := database.CreateAPIKey("test")
	require.NoError(t, err)

	body := `{"content":"expires soon","title":"TTL Test","ttl":3600}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/pastes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", ak.Key)
	srv.Router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var resp createPasteResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/", nil)
	srv.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "1")
}
