package config

type EnvMode string

const (
	Dev  EnvMode = "dev"
	Prod EnvMode = "prod"
)

type Config struct {
	Server       Server
	MongoDB      MongoDB
	Auth         Auth
	Email        Email
	OAuth2       OAuth2
	Website      Website
	MinIo        MinIo
	Integrations Integrations
}

type (
	MongoDB struct {
		Host                string `env:"DB_HOST"`
		Port                string `env:"DB_PORT"`
		User                string `env:"DB_USER"`
		Password            string `env:"DB_PASSWORD"`
		Database            string `env:"DB_NAME"`
		UserCollection      string `env:"DB_USER_COLLECTION"`
		WorkspaceCollection string `env:"DB_WORKSPACE_COLLECTION"`
		HelpDeskCollection  string `env:"DB_HELPDESK_COLLECTION"`
		ChatCollection      string `env:"DB_HELPDESK_COLLECTION"`
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
	}

	OAuth2 struct {
		StateText string `env:"STATE_TEXT"`

		GoogleClientId     string `env:"GOOGLE_CLIENT_ID"`
		GoogleRedirectURL  string `env:"GOOGLE_REDIRECT_URL"`
		GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET"`

		MetaClientId     string `env:"META_CLIENT_ID"`
		MetaRedirectURL  string `env:"META_REDIRECT_URL"`
		MetaClientSecret string `env:"META_CLIENT_SECRET"`

		TelegramClientId     string `env:"TELEGRAM_CLIENT_ID"`
		TelegramClientSecret string `env:"TELEGRAM_CLIENT_SECRET"`
	}

	Website struct {
		WebURL  string `env:"WEB_URL"`
		BaseURL string `env:"BASE_URL"`
	}

	MinIo struct {
		Endpoint   string `env:"MINIO_ENDPOINT"`
		AccessKey  string `env:"MINIO_ACCESS_KEY"`
		SecretKey  string `env:"MINIO_SECRET_KEY"`
		BucketName string `env:"MINIO_BUCKET_NAME"`
	}

	Integrations struct {
		TelegramBaseURL string `env:"TELEGRAM_BASE_URL"`
		MetaBaseURL     string `env:"META_BASE_URL"`
		WhatsappBaseURL string `env:"WHATSAPP_BASE_URL"`
	}
)
