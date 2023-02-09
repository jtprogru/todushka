package todo

import (
	"context"

	"github.com/jtprogru/todushka/internal/domain/entity"
)

// Методы, которые из REST API позволяют обратиться за чем либо
// т.е. бизнес-логика

type Repository interface {
	GetById(ctx context.Context, todoId int) (entity.Todo, error)
	GetAllTodos(ctx context.Context) ([]entity.Todo, error)
	DeleteTodo(ctx context.Context, todoId int) error
	CreateTodo(ctx context.Context, todo entity.TodoCreate) (entity.Todo, error)
	UpdateTodo(ctx context.Context, todoId int, todo entity.TodoUpdate) (entity.Todo, error)
}

type srv struct {
	repo Repository
}

func New(repo Repository) *srv {
	return &srv{
		repo: repo,
	}
}

func (s *srv) GetById(ctx context.Context, todoId int) (entity.Todo, error) {
	todo, err := s.repo.GetById(ctx, todoId)
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

func (s *srv) DeleteTodo(ctx context.Context, todoId int) error {
	err := s.repo.DeleteTodo(ctx, todoId)
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

func (s *srv) UpdateTodo(ctx context.Context, todoId int, todo entity.TodoUpdate) (entity.Todo, error) {
	t, err := s.repo.UpdateTodo(ctx, todoId, todo)
	if err != nil {
		return entity.Todo{}, err
	}
	return t, nil
}
