package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
)

var (
	instance Config
)

func Load() *Config {
	if err := godotenv.Load(".env"); err != nil {
		panic("Error loading .env file: " + err.Error())
	}

	if err := env.Parse(&instance); err != nil {
		panic(err)
	}
	return &instance
}
