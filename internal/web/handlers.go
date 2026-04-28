package web

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/knivey/dave-web/internal/db"
)

func (s *Server) handleList(c *gin.Context) {
	pastes, err := s.DB.ListPastes(50)
	if err != nil {
		c.String(http.StatusInternalServerError, "Error loading pastes: %v", err)
		return
	}

	type pasteItem struct {
		Slug      string
		Title     string
		CreatedAt string
		ExpiresAt string
	}

	items := make([]pasteItem, 0, len(pastes))
	for _, p := range pastes {
		title := p.Title
		if title == "" {
			title = "Untitled"
			first := strings.Split(p.Content, "\n")[0]
			if len(first) > 60 {
				first = first[:60] + "..."
			}
			if title == "Untitled" && first != "" {
				title = first
			}
		}
		expires := "never"
		if p.ExpiresAt != nil {
			expires = p.ExpiresAt.Format("2006-01-02 15:04")
		}
		items = append(items, pasteItem{
			Slug:      p.Slug,
			Title:     title,
			CreatedAt: p.CreatedAt.Format("2006-01-02 15:04"),
			ExpiresAt: expires,
		})
	}

	c.HTML(http.StatusOK, "list.html", gin.H{
		"Title":  "dave-web",
		"Pastes": items,
	})
}

func (s *Server) handleView(c *gin.Context) {
	slug := c.Param("slug")
	paste, err := s.DB.GetPaste(slug)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			c.String(http.StatusNotFound, "Paste not found")
		} else {
			c.String(http.StatusInternalServerError, "Error loading paste")
		}
		return
	}

	title := paste.Title
	if title == "" {
		title = "Untitled"
	}

	c.HTML(http.StatusOK, "view.html", gin.H{
		"Title":        title,
		"Paste":        paste,
		"RenderedHTML": template.HTML(paste.Rendered),
		"RawURL":       fmt.Sprintf("/p/%s/raw", paste.Slug),
		"BaseURL":      s.baseURL,
	})
}

func (s *Server) handleRaw(c *gin.Context) {
	slug := c.Param("slug")
	paste, err := s.DB.GetPaste(slug)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			c.String(http.StatusNotFound, "Paste not found")
		} else {
			c.String(http.StatusInternalServerError, "Error loading paste")
		}
		return
	}

	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.String(http.StatusOK, paste.Content)
}
