package project

import (
	"context"

	chi "github.com/go-chi/chi/v5"
	"github.com/jtprogru/todushka/internal/domain/entity"
)

type Service interface {
	GetProjectByID(ctx context.Context, projectID int) (entity.Project, error)
	GetAllProjects(ctx context.Context) ([]entity.Project, error)
	DeleteProject(ctx context.Context, projectID int) error
	CreateProject(ctx context.Context, project entity.ProjectCreate) (entity.Project, error)
	UpdateProject(ctx context.Context, projectID int, project entity.ProjectUpdate) (entity.Project, error)
}

type Handler struct {
	srv Service
}

func New(srv Service) *Handler {
	return &Handler{
		srv: srv,
	}
}

func Routes(srv Service) *chi.Mux {
	router := chi.NewRouter()
	handler := New(srv)

	router.Get("/{projectID}", handler.GetProject())
	// router.Put("/{projectID}", handler.UpdateProject())
	router.Get("/", handler.GetAllProjects())
	// router.Delete("/{projectID}", handler.DeleteProject())
	// router.Post("/", handler.CreateProject())
	return router
}
