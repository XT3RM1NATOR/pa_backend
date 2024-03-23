package config

type EnvMode string

const (
	Dev  EnvMode = "dev"
	Prod EnvMode = "prod"
)

type Config struct {
	Server  Server
	MongoDB MongoDB
	Auth    Auth
	Email   Email
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

	Auth struct {
		JWTSecretKey string `env:"JWT_SECRET_KEY"`
	}

	Email struct {
		SMTPUsername string `env:"SMTP_USERNAME"`
		SMTPPassword string `env:"SMTP_PASSWORD"`
		SMTPHost     string `env:"SMTP_HOST"`
		SMTPPort     string `env:"SMTP_PORT"`
		SenderEmail  string `env:"EMAIL="`
	}
)
