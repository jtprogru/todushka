package api

import (
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jtprogru/todushka/config"
	"github.com/jtprogru/todushka/models/list"
	"github.com/jtprogru/todushka/models/project"
	"github.com/jtprogru/todushka/models/todo"
)

type Api struct {
	cfg *config.Config
	R   *chi.Mux
}

func (a *Api) Start() {
	// Mount the admin sub-router
	http.ListenAndServe(a.cfg.Addr, a.R)
}

func NewApi(c *config.Config) *Api {
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
	}
}
