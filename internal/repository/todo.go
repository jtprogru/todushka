package repository

import (
	"context"

	"github.com/jtprogru/todushka/internal/domain/entity"
)

type Storage interface {
	GetById(ctx context.Context, todoId int) (entity.Todo, error)
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
	todo, _ := r.store.GetById(ctx, todoId)
	return todo, nil
}
