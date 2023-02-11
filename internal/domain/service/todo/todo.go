package todo

import (
	"context"

	"github.com/jtprogru/todushka/internal/domain/entity"
)

type Repository interface {
	GetTodoByID(ctx context.Context, todoID int) (entity.Todo, error)
	GetAllTodos(ctx context.Context) ([]entity.Todo, error)
	DeleteTodo(ctx context.Context, todoID int) error
	CreateTodo(ctx context.Context, todo entity.TodoCreate) (entity.Todo, error)
	UpdateTodo(ctx context.Context, todoID int, todo entity.TodoUpdate) (entity.Todo, error)
}

type srv struct {
	repo Repository
}

func New(repo Repository) *srv {
	return &srv{
		repo: repo,
	}
}

func (s *srv) GetTodoByID(ctx context.Context, todoID int) (entity.Todo, error) {
	return s.repo.GetTodoByID(ctx, todoID)
}

func (s *srv) GetAllTodos(ctx context.Context) ([]entity.Todo, error) {
	return s.repo.GetAllTodos(ctx)
}

func (s *srv) DeleteTodo(ctx context.Context, todoID int) error {
	return s.repo.DeleteTodo(ctx, todoID)
}

func (s *srv) CreateTodo(ctx context.Context, todo entity.TodoCreate) (entity.Todo, error) {
	return s.repo.CreateTodo(ctx, todo)
}

func (s *srv) UpdateTodo(ctx context.Context, todoID int, todo entity.TodoUpdate) (entity.Todo, error) {
	return s.repo.UpdateTodo(ctx, todoID, todo)
}
