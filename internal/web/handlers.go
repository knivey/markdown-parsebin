package web

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/knivey/dave-web/internal/db"
)

func (s *Server) handleList(c *gin.Context) {
	stats, err := s.DB.CountPastes()
	if err != nil {
		c.String(http.StatusInternalServerError, "Error loading stats: %v", err)
		return
	}

	c.HTML(http.StatusOK, "list.html", gin.H{
		"Title": "dave-web",
		"Stats": stats,
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
