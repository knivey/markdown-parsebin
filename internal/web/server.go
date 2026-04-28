package web

import (
	"html/template"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/knivey/dave-web/internal/db"
)

type multiTemplate struct {
	templates map[string]*template.Template
}

func (t *multiTemplate) Instance(name string, data interface{}) render.Render {
	return render.HTML{
		Template: t.templates[name],
		Name:     name,
		Data:     data,
	}
}

type Server struct {
	Router  *gin.Engine
	DB      db.Store
	baseURL string
}

func NewServer(database db.Store, templateFS fs.FS, staticFS fs.FS, baseURL string) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	r.StaticFS("/static", http.FS(staticFS))

	base := template.Must(template.ParseFS(templateFS, "base.html"))

	listBase, _ := base.Clone()
	listTmpl := template.Must(listBase.ParseFS(templateFS, "list.html"))

	viewBase, _ := base.Clone()
	viewTmpl := template.Must(viewBase.ParseFS(templateFS, "view.html"))

	r.HTMLRender = &multiTemplate{
		templates: map[string]*template.Template{
			"list.html": listTmpl,
			"view.html": viewTmpl,
		},
	}

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
