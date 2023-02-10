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
	todo, err := s.repo.GetTodoByID(ctx, todoID)
	if err != nil {
		return entity.Todo{}, err
	}
	return todo, nil
}

func (s *srv) GetAllTodos(ctx context.Context) ([]entity.Todo, error) {
	todos, err := s.repo.GetAllTodos(ctx)
	if err != nil {
		return []entity.Todo{}, err
	}
	return todos, nil
}

func (s *srv) DeleteTodo(ctx context.Context, todoID int) error {
	err := s.repo.DeleteTodo(ctx, todoID)
	if err != nil {
		return err
	}
	return nil
}

func (s *srv) CreateTodo(ctx context.Context, todo entity.TodoCreate) (entity.Todo, error) {
	t, err := s.repo.CreateTodo(ctx, todo)
	if err != nil {
		return entity.Todo{}, err
	}
	return t, nil
}

func (s *srv) UpdateTodo(ctx context.Context, todoID int, todo entity.TodoUpdate) (entity.Todo, error) {
	t, err := s.repo.UpdateTodo(ctx, todoID, todo)
	if err != nil {
		return entity.Todo{}, err
	}
	return t, nil
}
