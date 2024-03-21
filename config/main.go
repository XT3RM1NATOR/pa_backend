package config

import (
	"github.com/caarlos0/env/v6"
	_ "github.com/joho/godotenv/autoload"
	"sync"
)

var (
	instance Config
	once     sync.Once
)

func Load() *Config {
	once.Do(func() {
		if err := env.Parse(&instance); err != nil {
			panic(err)
		}
	})

	return &instance
}
