package postgres

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/jtprogru/todushka/internal/config"
	_ "github.com/lib/pq"
)

type storage struct {
	db *sqlx.DB
}

func NewPostgresDB(cfg *config.DBConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.DBName, cfg.Password, cfg.SSLMode))
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func New(db *sqlx.DB) *storage {
	return &storage{
		db: db,
	}
}
