package web

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) requireAPIKey(c *gin.Context) {
	key := c.GetHeader("X-API-Key")
	if key == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing X-API-Key header"})
		return
	}

	_, err := s.DB.GetAPIKey(key)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
		return
	}

	c.Next()
}
