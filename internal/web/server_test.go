package web

import (
	"bytes"
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
	"github.com/knivey/dave-web/internal/testutil"
)

func setupTestRouter(store db.Store) (*Server, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	tmpl := `{{define "list.html"}}{{if .Stats}}<div class="stats-bar"><span>{{.Stats.Total}}</span><span>{{.Stats.Active}}</span><span>{{.Stats.Expired}}</span></div>{{end}}{{end}}`
	tmpl += `{{define "view.html"}}<h1>{{.Title}}</h1><div>{{.RenderedHTML}}</div>{{end}}`
	r.SetHTMLTemplate(template.Must(template.New("").Parse(tmpl)))

	s := &Server{
		Router:  r,
		DB:      store,
		baseURL: "http://localhost:8080",
	}

	r.GET("/", s.handleList)
	r.GET("/p/:slug", s.handleView)
	r.GET("/p/:slug/raw", s.handleRaw)
	r.POST("/api/pastes", s.requireAPIKey, s.handleAPICreate)

	return s, r
}

func TestList_Empty(t *testing.T) {
	mock := &testutil.MockStore{
		CountPastesFunc: func() (*db.PasteStats, error) {
			return &db.PasteStats{Total: 0, Active: 0, Expired: 0}, nil
		},
	}
	_, r := setupTestRouter(mock)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestList_WithStats(t *testing.T) {
	mock := &testutil.MockStore{
		CountPastesFunc: func() (*db.PasteStats, error) {
			return &db.PasteStats{Total: 5, Active: 3, Expired: 2}, nil
		},
	}
	_, r := setupTestRouter(mock)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "5")
	assert.Contains(t, body, "3")
	assert.Contains(t, body, "2")
}

func TestList_StatsDBError(t *testing.T) {
	mock := &testutil.MockStore{
		CountPastesFunc: func() (*db.PasteStats, error) {
			return nil, fmt.Errorf("stats db error")
		},
	}
	_, r := setupTestRouter(mock)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Error loading stats")
}

func TestView_Found(t *testing.T) {
	mock := &testutil.MockStore{
		GetPasteFunc: func(slug string) (*models.Paste, error) {
			return &models.Paste{
				Slug:      "abc",
				Title:     "Test",
				Content:   "**bold**",
				Rendered:  "<p><strong>bold</strong></p>",
				CreatedAt: time.Now(),
			}, nil
		},
	}
	_, r := setupTestRouter(mock)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/p/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "<strong>bold</strong>")
}

func TestView_NotFound(t *testing.T) {
	mock := &testutil.MockStore{
		GetPasteFunc: func(slug string) (*models.Paste, error) {
			return nil, db.ErrNotFound
		},
	}
	_, r := setupTestRouter(mock)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/p/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestView_EmptyTitle(t *testing.T) {
	mock := &testutil.MockStore{
		GetPasteFunc: func(slug string) (*models.Paste, error) {
			return &models.Paste{
				Slug:     "abc",
				Title:    "",
				Content:  "test",
				Rendered: "<p>test</p>",
			}, nil
		},
	}
	_, r := setupTestRouter(mock)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/p/abc", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Untitled")
}

func TestRaw_Found(t *testing.T) {
	mock := &testutil.MockStore{
		GetPasteFunc: func(slug string) (*models.Paste, error) {
			return &models.Paste{
				Slug:     "abc",
				Content:  "# Raw Content",
				Rendered: "<h1>Raw Content</h1>",
			}, nil
		},
	}
	_, r := setupTestRouter(mock)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/p/abc/raw", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "# Raw Content", w.Body.String())
	assert.Contains(t, w.Header().Get("Content-Type"), "text/plain")
}

func TestRaw_NotFound(t *testing.T) {
	mock := &testutil.MockStore{
		GetPasteFunc: func(slug string) (*models.Paste, error) {
			return nil, db.ErrNotFound
		},
	}
	_, r := setupTestRouter(mock)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/p/nonexistent/raw", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestAPICreate_Success(t *testing.T) {
	mock := &testutil.MockStore{
		GetAPIKeyFunc: func(key string) (*db.APIKey, error) {
			return &db.APIKey{Key: key}, nil
		},
		CreatePasteFunc: func(paste *models.Paste) error {
			assert.Equal(t, "markdown", paste.Language)
			assert.Len(t, paste.Slug, 8)
			assert.Equal(t, "Test", paste.Title)
			assert.Nil(t, paste.ExpiresAt)
			assert.False(t, paste.CreatedAt.IsZero())
			return nil
		},
	}
	_, r := setupTestRouter(mock)

	body := `{"content":"# Hello","title":"Test"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/pastes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "dave_testkey123")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp createPasteResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp.Slug)
	assert.Contains(t, resp.URL, "/p/")
	assert.Contains(t, resp.URL, "http://localhost:8080")
}

func TestAPICreate_WithTTL(t *testing.T) {
	mock := &testutil.MockStore{
		GetAPIKeyFunc: func(key string) (*db.APIKey, error) {
			return &db.APIKey{Key: key}, nil
		},
		CreatePasteFunc: func(paste *models.Paste) error {
			assert.NotNil(t, paste.ExpiresAt, "ExpiresAt should be set when TTL provided")
			assert.True(t, paste.ExpiresAt.After(time.Now().Add(3500*time.Second)), "ExpiresAt should be ~1 hour from now")
			assert.True(t, paste.ExpiresAt.Before(time.Now().Add(3700*time.Second)), "ExpiresAt should be ~1 hour from now")
			return nil
		},
	}
	_, r := setupTestRouter(mock)

	body := `{"content":"expires in 1h","ttl":3600}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/pastes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "dave_testkey123")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAPICreate_ZeroTTL(t *testing.T) {
	mock := &testutil.MockStore{
		GetAPIKeyFunc: func(key string) (*db.APIKey, error) {
			return &db.APIKey{Key: key}, nil
		},
		CreatePasteFunc: func(paste *models.Paste) error {
			assert.Nil(t, paste.ExpiresAt, "ExpiresAt should be nil when TTL is 0")
			return nil
		},
	}
	_, r := setupTestRouter(mock)

	body := `{"content":"no expiry","ttl":0}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/pastes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "dave_testkey123")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestAPICreate_DBError(t *testing.T) {
	mock := &testutil.MockStore{
		GetAPIKeyFunc: func(key string) (*db.APIKey, error) {
			return &db.APIKey{Key: key}, nil
		},
		CreatePasteFunc: func(paste *models.Paste) error {
			return fmt.Errorf("db error")
		},
	}
	_, r := setupTestRouter(mock)

	body := `{"content":"test"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/pastes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "dave_testkey123")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "failed to create paste")
}

func TestAPICreate_NoTitleField(t *testing.T) {
	var capturedPaste *models.Paste
	mock := &testutil.MockStore{
		GetAPIKeyFunc: func(key string) (*db.APIKey, error) {
			return &db.APIKey{Key: key}, nil
		},
		CreatePasteFunc: func(paste *models.Paste) error {
			capturedPaste = paste
			return nil
		},
	}
	_, r := setupTestRouter(mock)

	body := `{"content":"just content"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/pastes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "dave_testkey123")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "", capturedPaste.Title)
}

func TestAPICreate_MissingContent(t *testing.T) {
	mock := &testutil.MockStore{
		GetAPIKeyFunc: func(key string) (*db.APIKey, error) {
			return &db.APIKey{Key: key}, nil
		},
		CreatePasteFunc: func(paste *models.Paste) error { return nil },
	}
	_, r := setupTestRouter(mock)

	body := `{"title":"No Content"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/pastes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "dave_testkey123")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAPICreate_InvalidJSON(t *testing.T) {
	mock := &testutil.MockStore{
		GetAPIKeyFunc: func(key string) (*db.APIKey, error) {
			return &db.APIKey{Key: key}, nil
		},
	}
	_, r := setupTestRouter(mock)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/pastes", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "dave_testkey123")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMiddleware_NoKey(t *testing.T) {
	_, r := setupTestRouter(&testutil.MockStore{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/pastes", bytes.NewBufferString(`{"content":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "missing X-API-Key")
}

func TestMiddleware_BadKey(t *testing.T) {
	mock := &testutil.MockStore{
		GetAPIKeyFunc: func(key string) (*db.APIKey, error) {
			return nil, db.ErrNotFound
		},
	}
	_, r := setupTestRouter(mock)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/pastes", bytes.NewBufferString(`{"content":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "bad_key")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "invalid API key")
}

func TestMiddleware_ValidKey(t *testing.T) {
	mock := &testutil.MockStore{
		GetAPIKeyFunc: func(key string) (*db.APIKey, error) {
			return &db.APIKey{Key: key}, nil
		},
		CreatePasteFunc: func(paste *models.Paste) error {
			return nil
		},
	}
	_, r := setupTestRouter(mock)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/pastes", bytes.NewBufferString(`{"content":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "dave_validkey")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}
