package config

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
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
	Server      *ServerConfig
	Monitor     *MonitorConfig     `envconfig:"monitor"`
	Session     *SessionConfig     `envconfig:"session"`
	CORS        *CORSConfig        `envconfig:"cors"`
	Token       *TokenConfig       `envconfig:"token"`
	Logger      *LoggerConfig      `envconfig:"logger"`
	GoogleOAuth *GoogleOAuthConfig `envconfig:"google_oauth"`
	Redis       *RedisConfig       `envconfig:"redis"`
	Postgres    *PostgresConfig    `envconfig:"postgres"`
}

type ServerConfig struct {
	Version string `envconfig:"version" default:"0.0.1"`
	Host    string `envconfig:"host" default:"localhost"`
	Port    string `envconfig:"port" default:"9731"`
	AESKey  string `envconfig:"aes_key" required:"true"`
}

type MonitorConfig struct {
	ServiceName       string `envconfig:"service_name" default:"gitlab.com/jodworkspace/mvp"`
	CollectorEndpoint string `envconfig:"collector_endpoint" default:"localhost:4317"`
}

type SessionConfig struct {
	CookieSecret string `envconfig:"cookie_secret"`
	Name         string `envconfig:"name" default:"sid"`
	Domain       string `envconfig:"domain" default:"localhost"`
	CookiePath   string `envconfig:"cookie_path" default:"/"`
	MaxAge       int    `envconfig:"max_age" default:"86400"`
	HTTPOnly     bool   `envconfig:"http_only" default:"true"`
	Secure       bool   `envconfig:"secure" default:"false"`

	// Storage configuration
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
	Secret      string        `envconfig:"secret"`
	ShortExpiry time.Duration `envconfig:"short_expiry" default:"3600s"`   // 1 hour
	LongExpiry  time.Duration `envconfig:"long_expiry" default:"2592000s"` // 30 days
	Issuer      string        `envconfig:"issuer" default:"jodworkspace"`
}

type LoggerConfig struct {
	Level   string `envconfig:"log_level" default:"info"`
	Request bool   `envconfig:"log_request" default:"true"`
}

type GoogleOAuthConfig struct {
	ClientID         string `envconfig:"client_id" required:"true"`
	ClientSecret     string `envconfig:"client_secret" required:"true"`
	TokenEndpoint    string `envconfig:"token_endpoint" required:"true"`
	UserInfoEndpoint string `envconfig:"userinfo_endpoint" required:"true"`
}

type RedisConfig struct {
	Host     string `envconfig:"redis_host" default:"localhost"`
	Port     uint16 `envconfig:"redis_port" default:"6379"`
	Username string `envconfig:"redis_username" default:"default"`
	Password string `envconfig:"redis_password" default:""`
	DB       int    `envconfig:"redis_db" default:"0"`
}

type PostgresConfig struct {
	Host     string `envconfig:"host" default:"localhost"`
	Port     uint16 `envconfig:"port" default:"5432"`
	Username string `envconfig:"username" required:"true"`
	Password string `envconfig:"password" required:"true"`
	Database string `envconfig:"database" required:"true"`
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
