package repository

import (
	"github.com/jmoiron/sqlx"
	"github.com/jtprogru/todushka/models/todo"
)

type ITodo interface {
	GetAll() ([]todo.Todo, error)
	GetById(todoId string) (todo.Todo, error)
}

type Repository struct {
	ITodo
}

func New(db *sqlx.DB) *Repository {
	return &Repository{
		ITodo: NewTodoPostgres(db),
	}
}
