package web

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/knivey/dave-web/internal/db"
	"github.com/knivey/dave-web/internal/models"
	"github.com/knivey/dave-web/internal/renderer"
	"github.com/knivey/dave-web/internal/util"
)

const maxTTL = 365 * 24 * time.Hour

type createPasteRequest struct {
	Content string `json:"content" binding:"required"`
	Title   string `json:"title"`
	TTL     int    `json:"ttl"`
}

type createPasteResponse struct {
	Slug string `json:"slug"`
	URL  string `json:"url"`
}

func (s *Server) handleAPICreate(c *gin.Context) {
	var req createPasteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
		return
	}

	if req.TTL < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ttl must be >= 0"})
		return
	}
	if req.TTL > 0 && time.Duration(req.TTL)*time.Second > maxTTL {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("ttl must be <= %d seconds", int(maxTTL.Seconds()))})
		return
	}

	rendered, err := renderer.Render(req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to render markdown"})
		return
	}

	var slug string
	var paste *models.Paste
	for attempt := 0; attempt < 3; attempt++ {
		slug, err = util.GenerateSlug()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate slug"})
			return
		}

		paste = &models.Paste{
			Slug:      slug,
			Title:     req.Title,
			Content:   req.Content,
			Rendered:  rendered,
			CreatedAt: time.Now(),
			Language:  "markdown",
		}

		if req.TTL > 0 {
			expiresAt := time.Now().Add(time.Duration(req.TTL) * time.Second)
			paste.ExpiresAt = &expiresAt
		}

		err = s.DB.CreatePaste(paste)
		if !db.IsDuplicateSlug(err) {
			break
		}
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create paste"})
		return
	}

	c.JSON(http.StatusCreated, createPasteResponse{
		Slug: slug,
		URL:  fmt.Sprintf("%s/p/%s", s.baseURL, slug),
	})
}
