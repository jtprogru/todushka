package api

import (
	"log"
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/jtprogru/todushka/internal/config"
	"github.com/jtprogru/todushka/models/list"
	"github.com/jtprogru/todushka/models/project"
	"github.com/jtprogru/todushka/models/todo"
)

type Api struct {
	cfg *config.ServerConfig
	R   *chi.Mux
	s   *sqlx.DB
}

func (a *Api) Start() {
	log.Fatalln(http.ListenAndServe(a.cfg.Addr, a.R))
}

func New(c *config.ServerConfig, s *sqlx.DB) *Api {
	r := chi.NewRouter()
	// A good base middleware stack
	r.Use(
		middleware.Logger,          // Log API request calls
		middleware.RedirectSlashes, // Redirect slashes to no slash URL versions
		middleware.Recoverer,       // Recover from panics without crashing server
		middleware.RealIP,
		middleware.RequestID,
	)

	r.Route("/v1", func(r chi.Router) {
		r.Mount("/api/todo", todo.Routes())
		r.Mount("/api/list", list.Routes())
		r.Mount("/api/project", project.Routes())
	})

	return &Api{
		cfg: c,
		R:   r,
		s:   s,
	}
}
