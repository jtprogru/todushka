package repository

import (
	"context"

	"github.com/jtprogru/todushka/internal/domain/entity"
)

type Storage interface {
	GetById(ctx context.Context, todoId int) (entity.Todo, error)
	GetAllTodos(ctx context.Context) ([]entity.Todo, error)
	DeleteTodo(ctx context.Context, todoId int) error
}

type repo struct {
	store Storage
}

func New(store Storage) *repo {
	return &repo{
		store: store,
	}
}

func (r *repo) GetById(ctx context.Context, todoId int) (entity.Todo, error) {
	todo, err := r.store.GetById(ctx, todoId)
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

func (r *repo) DeleteTodo(ctx context.Context, todoId int) error {
	err := r.store.DeleteTodo(ctx, todoId)
	if err != nil {
		return err
	}
	return nil
}
