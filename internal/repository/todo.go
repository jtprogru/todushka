package repository

import (
	"context"

	"github.com/jtprogru/todushka/internal/domain/entity"
)

type Storage interface {
	GetByID(ctx context.Context, todoID int) (entity.Todo, error)
	GetAllTodos(ctx context.Context) ([]entity.Todo, error)
	DeleteTodo(ctx context.Context, todoID int) error
	CreateTodo(ctx context.Context, todo entity.TodoCreate) (entity.Todo, error)
	UpdateTodo(ctx context.Context, todoID int, todo entity.TodoUpdate) (entity.Todo, error)
}

type repo struct {
	store Storage
}

func New(store Storage) *repo {
	return &repo{
		store: store,
	}
}

func (r *repo) GetByID(ctx context.Context, todoID int) (entity.Todo, error) {
	todo, err := r.store.GetByID(ctx, todoID)
	if err != nil {
		return entity.Todo{}, err
	}
	return todo, nil
}

func (r *repo) GetAllTodos(ctx context.Context) ([]entity.Todo, error) {
	todos, err := r.store.GetAllTodos(ctx)
	if err != nil {
		return []entity.Todo{}, err
	}
	return todos, nil
}

func (r *repo) DeleteTodo(ctx context.Context, todoID int) error {
	err := r.store.DeleteTodo(ctx, todoID)
	if err != nil {
		return err
	}
	return nil
}

func (r *repo) CreateTodo(ctx context.Context, todo entity.TodoCreate) (entity.Todo, error) {
	t, err := r.store.CreateTodo(ctx, todo)
	if err != nil {
		return entity.Todo{}, err
	}
	return t, nil
}

func (r *repo) UpdateTodo(ctx context.Context, todoID int, todo entity.TodoUpdate) (entity.Todo, error) {
	t, err := r.store.UpdateTodo(ctx, todoID, todo)
	if err != nil {
		return entity.Todo{}, err
	}
	return t, nil
}
