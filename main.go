package main

import (
	"log"

	"github.com/jtprogru/todushka/api"
	"github.com/jtprogru/todushka/config"
)

func main() {
	cfg := config.NewConfig()
	api := api.NewApi(cfg)
	log.Printf("todushka is running...")
	api.Start()
	log.Printf("todushka is dead...")
}
