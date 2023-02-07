package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/jtprogru/todushka/models/todo"
)

type TodoPostgres struct {
	db *sqlx.DB
}

func NewTodoPostgres(db *sqlx.DB) *TodoPostgres {
	return &TodoPostgres{db: db}
}

func (tp *TodoPostgres) GetAll() ([]todo.Todo, error) {
	var todos []todo.Todo
	tp.db.Select(&todos, "SELECT * FROM todo_items")
	return todos, nil
}

func (tp *TodoPostgres) GetById(todoId string) (todo.Todo, error) {
	return todo.Todo{}, nil
}
