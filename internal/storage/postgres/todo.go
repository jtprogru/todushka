package postgres

import (
	"context"

	"github.com/jtprogru/todushka/internal/domain/entity"
)

func (s *storage) GetById(ctx context.Context, todoId int) (entity.Todo, error) {
	query := `SELECT 1;`
	var res string
	s.db.Get(&res, query)
	return entity.Todo{}, nil
}
