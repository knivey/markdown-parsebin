package web

import (
	"html/template"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/knivey/dave-web/internal/db"
)

type Server struct {
	Router  *gin.Engine
	DB      db.Store
	baseURL string
}

func NewServer(database db.Store, templateFS fs.FS, staticFS fs.FS, baseURL string) *Server {
	tmpl := template.Must(template.ParseFS(templateFS, "*.html"))

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	r.SetHTMLTemplate(tmpl)
	r.StaticFS("/static", http.FS(staticFS))

	s := &Server{
		Router:  r,
		DB:      database,
		baseURL: baseURL,
	}

	r.GET("/", s.handleList)
	r.GET("/p/:slug", s.handleView)
	r.GET("/p/:slug/raw", s.handleRaw)
	r.POST("/api/pastes", s.requireAPIKey, s.handleAPICreate)

	return s
}

func (s *Server) Run(addr string) error {
	return s.Router.Run(addr)
}
