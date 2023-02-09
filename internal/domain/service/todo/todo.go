package todo

import (
	"context"

	"github.com/jtprogru/todushka/internal/domain/entity"
)

// Методы, которые из REST API позволяют обратиться за чем либо
// т.е. бизнес-логика

type Repository interface {
	GetById(ctx context.Context, todoId int) (entity.Todo, error)
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
	todo, _ := s.repo.GetById(ctx, todoId)
	return todo, nil
}
