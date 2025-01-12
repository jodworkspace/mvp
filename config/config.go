package config

import (
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"log"
)

func NewConfig(envFiles ...string) *Config {
	err := godotenv.Load(envFiles...)
	if err != nil {
		log.Printf("[Warning] config - init - godotenv.Load: %v", err)
	}

	cfg := &Config{}
	err = envconfig.Process("", cfg)
	if err != nil {
		log.Fatalf("config - init - envconfig.Process: %v", err)
	}
	return cfg
}

type Config struct {
	Server      *ServerConfig
	Logger      *LoggerConfig
	GoogleOAuth *GoogleOAuthConfig
	Postgres    *PostgresConfig
	Redis       *RedisConfig
}

type ServerConfig struct {
	Version   string `envconfig:"version" default:"0.0.1"`
	Host      string `envconfig:"host" default:"localhost"`
	Port      string `envconfig:"port" default:"9731"`
	JWTSecret string `envconfig:"jwt_secret"`
}

type LoggerConfig struct {
	Level   string `envconfig:"log_level" default:"info"`
	Request bool   `envconfig:"log_request" default:"true"`
}

type PostgresConfig struct {
	Host     string `envconfig:"pg_host" default:"localhost"`
	Port     uint16 `envconfig:"pg_port" default:"5432"`
	User     string `envconfig:"pg_user" required:"true"`
	Password string `envconfig:"pg_password" required:"true"`
	Database string `envconfig:"pg_database" required:"true"`
	Params   map[string]string
}

type RedisConfig struct {
	Host string `envconfig:"redis_host" default:"localhost"`
	Port uint16 `envconfig:"redis_port" default:"6379"`
}

type GoogleOAuthConfig struct {
	ClientID     string `envconfig:"client_id"`
	ClientSecret string `envconfig:"client_secret"`
	RedirectURI  string `envconfig:"redirect_uri"`
}
