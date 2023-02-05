package config

import (
	env "github.com/caitlinelfring/go-env-default"
)

type Config struct {
	Addr string `env:"ADDR" env-default:""`
}

func NewConfig() *Config {
	return &Config{
		Addr: env.GetDefault("ADDR", ":8080"),
	}
}
