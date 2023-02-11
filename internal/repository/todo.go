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
	return r.st.GetTodoByID(ctx, todoID)
}

func (r *repo) GetAllTodos(ctx context.Context) ([]entity.Todo, error) {
	return r.st.GetAllTodos(ctx)
}

func (r *repo) DeleteTodo(ctx context.Context, todoID int) error {
	return r.st.DeleteTodo(ctx, todoID)
}

func (r *repo) CreateTodo(ctx context.Context, todo entity.TodoCreate) (entity.Todo, error) {
	return r.st.CreateTodo(ctx, todo)
}

func (r *repo) UpdateTodo(ctx context.Context, todoID int, todo entity.TodoUpdate) (entity.Todo, error) {
	return r.st.UpdateTodo(ctx, todoID, todo)
}
