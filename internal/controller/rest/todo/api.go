package todo

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/jtprogru/todushka/internal/domain/entity"
)

type Service interface {
	GetById(ctx context.Context, todoId int) (entity.Todo, error)
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

	router.Get("/{todoID}", handler.GetTodo())
	// router.Delete("/{todoID}", DeleteTodo)
	// router.Post("/", CreateTodo)
	// router.Get("/", GetAllTodos)
	return router
}
