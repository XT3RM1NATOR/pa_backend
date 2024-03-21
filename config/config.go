package config

type EnvMode string

const (
	Dev  EnvMode = "dev"
	Prod EnvMode = "prod"
)

type Config struct {
	Server  Server
	MongoDB MongoDB
}

type (
	MongoDB struct {
		Host     string `env:"DB_HOST"`
		Port     string `env:"DB_PORT"`
		User     string `env:"DB_USER"`
		Password string `env:"DB_PASSWORD"`
		Database string `env:"DB_NAME"`
	}

	Server struct {
		Environment EnvMode `env:"SERVER_ENVIRONMENT" envDefault:"dev"`
		Port        string  `env:"SERVER_PORT"`
	}
)
