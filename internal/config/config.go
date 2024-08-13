package config

import (
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Debug  bool   `yaml:"debug"`
	Server Server `yaml:"server"`
}

type Server struct {
	Port string `yaml:"port" env-default:"8000"`
	Host string `yaml:"host" env-default:"locahost"`

	ReadTimeout  time.Duration `yaml:"read_timeout" env-default:"2s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env-default:"2s"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

func MustLoad(configPath string) *Config {
	var cfg Config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic(fmt.Errorf("config file %s not found", configPath))
	}
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(err)
	}

	return &cfg
}