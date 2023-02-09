package postgres

import (
	"context"

	"github.com/jtprogru/todushka/internal/domain/entity"
)

func (s *storage) GetById(ctx context.Context, todoId int) (entity.Todo, error) {
	query := `select id, summary, description, is_done from todo_items where id=$1;`
	var todo entity.Todo
	err := s.db.Get(&todo, query, todoId)
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

func (s *storage) DeleteTodo(ctx context.Context, todoId int) error {
	query := `delete from todo_items where id=$1;`
	_, err := s.db.Exec(query, todoId)
	if err != nil {
		return err
	}
	return nil
}
