package todo

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/jtprogru/todushka/internal/domain/entity"
)

type Service interface {
	GetById(ctx context.Context, todoId int) (entity.Todo, error)
	GetAllTodos(ctx context.Context) ([]entity.Todo, error)
	DeleteTodo(ctx context.Context, todoId int) error
	CreateTodo(ctx context.Context, todo entity.TodoCreate) (entity.Todo, error)
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
	router.Get("/", handler.GetAllTodos())
	router.Delete("/{todoID}", handler.DeleteTodo())
	router.Post("/", handler.CreateTodo())
	return router
}
