package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Debug      bool          `yaml:"debug"`
	Limiter    limiter       `yaml:"limiter"`
	AppID      int32         `yaml:"app_id"`
	AppSecret  string        `yaml:"app_secret"`
	Server     server        `yaml:"server"`
	DB         db            `yaml:"db"`
	Clients    clientsConfig `yaml:"clients"`
	SMTPServer smtp          `yaml:"smtp_server"`
	CORS       Cors          `yaml:"cors"`
}

type smtp struct {
	Host         string        `yaml:"host" env-required:"true"`
	Port         int           `yaml:"port" env-required:"true"`
	Username     string        `yaml:"username" env-required:"true" env:"SMTP_USERNAME"`
	Password     string        `yaml:"password" env-required:"true" env:"SMTP_PASSWORD"`
	Sender       string        `yaml:"sender" env-required:"true"`
	Timeout      time.Duration `yaml:"timeout" env-default:"5s"`
	ApiToken     string        `yaml:"api_token" env-required:"true" env:"SMTP_API_TOKEN"`
	RetriesCount int           `yaml:"retries_count" env-default:"1"`
}

type limiter struct {
	Enabled bool    `yaml:"enabled"`
	Rps     float64 `yaml:"rps" env-default:"20"`
	Burst   int     `yaml:"burst" env-default:"5"`
}

type client struct {
	Addr         string        `yaml:"addr" env-required:"true"`
	RetryTimeout time.Duration `yaml:"retry_timeout" env-default:"1s"`
	RetriesCount int           `yaml:"retries_count" env-default:"1"`
}

type clientsConfig struct {
	SSO client `yaml:"sso"`
}
type server struct {
	Port string `yaml:"port" env-default:"8000"`
	Host string `yaml:"host" env-default:"locahost"`

	ReadTimeout     time.Duration `yaml:"read_timeout" env-default:"2s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env-default:"2s"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" env-default:"60s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env-default:"5s"`
}

type db struct {
	Driver          string        `yaml:"driver" env-default:"postgres"`
	User            string        `yaml:"user" env-required:"true" env:"DB_USER"`
	Password        string        `yaml:"password" env-required:"true" env:"DB_PASSWORD"`
	Host            string        `yaml:"host" env-required:"true"`
	Port            string        `yaml:"port" env-required:"true"`
	Name            string        `yaml:"name" env-required:"true"`
	MaxConns        int           `yaml:"max_conns" env-default:"25"`
	MaxConnIdleTime time.Duration `yaml:"max_conn_idle_time" env-default:"10m"`
}

type Cors struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
}

func (db *db) GetDsn() string {
	return fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=disable", db.Driver, db.User, db.Password, db.Host, db.Port, db.Name)
}

func MustLoad(configPath string) *Config {
	var cfg Config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic(fmt.Errorf("config file %s not found", configPath))
	}
	err := godotenv.Load()
	if err != nil {
		panic(fmt.Errorf("error loading .env file: %w", err))
	}
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(err)
	}

	return &cfg
}
