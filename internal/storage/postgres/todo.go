package postgres

import (
	"context"

	"github.com/jtprogru/todushka/internal/domain/entity"
)

func (s *storage) GetTodoByID(ctx context.Context, todoID int) (entity.Todo, error) {
	query := `select id, summary, description, is_done from todo_items where id=$1;`
	var todo entity.Todo
	err := s.db.Get(&todo, query, todoID)
	if err != nil {
		return todo, err
	}
	return todo, nil
}

func (s *storage) GetAllTodos(ctx context.Context) ([]entity.Todo, error) {
	query := `select id, summary, description, is_done from todo_items;`
	var todos []entity.Todo
	err := s.db.SelectContext(ctx, &todos, query)
	if err != nil {
		return todos, nil
	}
	return todos, nil
}

func (s *storage) DeleteTodo(ctx context.Context, todoID int) error {
	query := `delete from todo_items where id=$1;`
	_, err := s.db.Exec(query, todoID)
	if err != nil {
		return err
	}
	return nil
}

func (s *storage) CreateTodo(ctx context.Context, todo entity.TodoCreate) (entity.Todo, error) {
	query := `insert into todo_items (summary, description) values ($1, $2) returning id;`
	var todoID int
	err := s.db.QueryRow(query, todo.Summary, todo.Description).Scan(&todoID)
	if err != nil {
		return entity.Todo{}, err
	}
	resp := entity.Todo{
		ID:          todoID,
		Summary:     todo.Summary,
		Description: todo.Description,
		IsDone:      false,
	}

	return resp, nil
}

func (s *storage) UpdateTodo(ctx context.Context, todoID int, todo entity.TodoUpdate) (entity.Todo, error) {
	query := `update todo_items
    set summary = $1, description = $2, is_done = $3
    where id = $4 returning id, summary, description, is_done;`
	var resp entity.Todo

	err := s.db.QueryRow(
		query,
		todo.Summary,
		todo.Description,
		todo.IsDone,
		todoID).Scan(&resp.ID, &resp.Summary, &resp.Description, &resp.IsDone)
	if err != nil {
		return resp, err
	}

	return resp, nil
}
