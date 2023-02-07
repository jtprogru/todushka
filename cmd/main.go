package main

import (
	"log"

	"github.com/jtprogru/todushka/api"
	"github.com/jtprogru/todushka/internal/config"
	"github.com/jtprogru/todushka/internal/repository"
)

var (
	cfg *config.Config
)

func init() {
	cfg = config.New()
}

func main() {
	db, err := repository.NewPostgresDB(cfg.Db)
	if err != nil {
		log.Fatalf("db initialisation is failed: %v", err.Error())
	}
	app := api.New(cfg.Server, db)
	log.Printf("todushka is running...")
	app.Start()
}
