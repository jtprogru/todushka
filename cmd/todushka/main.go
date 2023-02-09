package main

import (
	"log"

	"github.com/jtprogru/todushka/internal/api"
	"github.com/jtprogru/todushka/internal/config"
	"github.com/jtprogru/todushka/internal/domain/service/todo"
	"github.com/jtprogru/todushka/internal/repository"
	"github.com/jtprogru/todushka/internal/storage/postgres"
)

func main() {
	cfg := config.New()
	db, err := postgres.NewPostgresDB(cfg.Db)
	if err != nil {
		panic("db ")
	}
	store := postgres.New(db)
	repo := repository.New(store)
	svc := todo.New(repo)
	app := api.New(cfg.Server, svc)
	log.Printf("todushka is running...")
	app.Start()
}
