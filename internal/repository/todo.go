package repository

import (
	"context"

	"github.com/jtprogru/todushka/internal/domain/entity"
)

type StorageTodo interface {
	GetTodoByID(ctx context.Context, todoID int) (entity.Todo, error)
	GetAllTodos(ctx context.Context) ([]entity.Todo, error)
	DeleteTodo(ctx context.Context, todoID int) error
	CreateTodo(ctx context.Context, todo entity.TodoCreate) (entity.Todo, error)
	UpdateTodo(ctx context.Context, todoID int, todo entity.TodoUpdate) (entity.Todo, error)
}

func (r *repo) GetTodoByID(ctx context.Context, todoID int) (entity.Todo, error) {
	todo, err := r.st.GetTodoByID(ctx, todoID)
	if err != nil {
		return entity.Todo{}, err
	}
	return todo, nil
}

func (r *repo) GetAllTodos(ctx context.Context) ([]entity.Todo, error) {
	todos, err := r.st.GetAllTodos(ctx)
	if err != nil {
		return []entity.Todo{}, err
	}
	return todos, nil
}

func (r *repo) DeleteTodo(ctx context.Context, todoID int) error {
	err := r.st.DeleteTodo(ctx, todoID)
	if err != nil {
		return err
	}
	return nil
}

func (r *repo) CreateTodo(ctx context.Context, todo entity.TodoCreate) (entity.Todo, error) {
	t, err := r.st.CreateTodo(ctx, todo)
	if err != nil {
		return entity.Todo{}, err
	}
	return t, nil
}

func (r *repo) UpdateTodo(ctx context.Context, todoID int, todo entity.TodoUpdate) (entity.Todo, error) {
	t, err := r.st.UpdateTodo(ctx, todoID, todo)
	if err != nil {
		return entity.Todo{}, err
	}
	return t, nil
}
