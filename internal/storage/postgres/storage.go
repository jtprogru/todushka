package postgres

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/jtprogru/todushka/internal/config"
)

type storage struct {
	db *sqlx.DB
}

func NewPostgresDB(cfg *config.DbConfig) (*sqlx.DB, error) {
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
