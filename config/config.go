package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"log"
	"strings"
	"time"
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
	JWT         *JWTConfig
	Logger      *LoggerConfig
	HTTPClient  *HTTPClientConfig
	GoogleOAuth *GoogleOAuthConfig
	Redis       *RedisConfig
	Postgres    *PostgresConfig
}

type ServerConfig struct {
	Version string `envconfig:"version" default:"0.0.1"`
	Host    string `envconfig:"host" default:"localhost"`
	Port    string `envconfig:"port" default:"9731"`
}

type HTTPClientConfig struct {
	Timeout time.Duration `envconfig:"timeout" default:"10s"`
}

type JWTConfig struct {
	Secret string        `envconfig:"jwt_secret"`
	Expiry time.Duration `envconfig:"jwt_expire" default:"3600s"`
}

type LoggerConfig struct {
	Level   string `envconfig:"log_level" default:"info"`
	Request bool   `envconfig:"log_request" default:"true"`
}

type GoogleOAuthConfig struct {
	ClientID      string `envconfig:"google_client_id"`
	ClientSecret  string `envconfig:"google_client_secret"`
	TokenEndpoint string `envconfig:"google_token_endpoint"`
}

type RedisConfig struct {
	Host string `envconfig:"redis_host" default:"localhost"`
	Port uint16 `envconfig:"redis_port" default:"6379"`
}

type PostgresConfig struct {
	Host     string `envconfig:"pg_host" default:"localhost"`
	Port     uint16 `envconfig:"pg_port" default:"5432"`
	Username string `envconfig:"pg_username" required:"true"`
	Password string `envconfig:"pg_password" required:"true"`
	Database string `envconfig:"pg_database" required:"true"`
	Params   map[string]string
}

func (c *PostgresConfig) DSN(opts ...map[string]string) string {
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s",
		c.Username,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
	)

	if len(opts) > 0 {
		for k, v := range opts[0] {
			dsn += fmt.Sprintf("&%s=%s", k, v)
		}
	}

	dsn = strings.Replace(dsn, "&", "?", 1)
	return dsn
}
