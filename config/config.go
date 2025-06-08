package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"log"
	"strings"
	"time"
)

func LoadConfig(envFiles ...string) *Config {
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
	Server        *ServerConfig
	SessionConfig *SessionConfig     `envconfig:"session"`
	CORS          *CORSConfig        `envconfig:"cors"`
	Token         *TokenConfig       `envconfig:"token"`
	Logger        *LoggerConfig      `envconfig:"logger"`
	GoogleOAuth   *GoogleOAuthConfig `envconfig:"google_oauth"`
	Redis         *RedisConfig       `envconfig:"redis"`
	Postgres      *PostgresConfig    `envconfig:"postgres"`
}

type ServerConfig struct {
	Version string `envconfig:"version" default:"0.0.1"`
	Host    string `envconfig:"host" default:"localhost"`
	Port    string `envconfig:"port" default:"9731"`
}

type SessionConfig struct {
	Secret        string `envconfig:"secret" required:"true"`
	Name          string `envconfig:"name" default:"session"`
	RedisHost     string `envconfig:"redis_host" default:"localhost"`
	RedisPort     uint16 `envconfig:"redis_port" default:"6379"`
	RedisDB       int    `envconfig:"redis_db" default:"0"`
	RedisUsername string `envconfig:"redis_username" default:"default"`
	RedisPassword string `envconfig:"redis_password" default:""`
}

type CORSConfig struct {
	AllowedOrigins   []string `envconfig:"allowed_origins" default:"*"`
	AllowedMethods   []string `envconfig:"allowed_methods" default:"GET,POST,PUT,PATCH,DELETE,OPTIONS"`
	AllowedHeaders   []string `envconfig:"allowed_headers" default:"*"`
	AllowCredentials bool     `envconfig:"allow_credentials" default:"false"`
	ExposedHeaders   []string `envconfig:"exposed_headers" default:"*"`
}

type TokenConfig struct {
	Secret        string        `envconfig:"secret" required:"true"`
	RefreshSecret string        `envconfig:"refresh_secret" required:"true"`
	ShortExpiry   time.Duration `envconfig:"short_expiry" default:"3600s"`   // 1 hour
	LongExpiry    time.Duration `envconfig:"long_expiry" default:"2592000s"` // 30 days
	Issuer        string        `envconfig:"issuer" default:"gookie.io"`
}

type LoggerConfig struct {
	Level   string `envconfig:"log_level" default:"info"`
	Request bool   `envconfig:"log_request" default:"true"`
}

type GoogleOAuthConfig struct {
	ClientID         string `envconfig:"google_client_id" required:"true"`
	ClientSecret     string `envconfig:"google_client_secret" required:"true"`
	TokenEndpoint    string `envconfig:"google_token_endpoint" required:"true"`
	UserInfoEndpoint string `envconfig:"google_userinfo_endpoint" required:"true"`
}

type RedisConfig struct {
	Host     string `envconfig:"redis_host" default:"localhost"`
	Port     uint16 `envconfig:"redis_port" default:"6379"`
	Username string `envconfig:"redis_username" default:"default"`
	Password string `envconfig:"redis_password" default:""`
	DB       int    `envconfig:"redis_db" default:"0"`
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
