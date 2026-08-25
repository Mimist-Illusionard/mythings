package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	DBParams DatabaseParams
	HTTPPort string
}

type DatabaseParams struct {
	Name string `env:"DB_NAME"`
	Host string `env:"DB_HOST"`
	Port string `env:"DB_PORT"`
	User string `env:"DB_USER"`
	Pass string `env:"DB_PASS"`
}

func New(httpPort, envPath string) (*Config, error) {
	if envPath != "" {
		if err := godotenv.Load(envPath); err != nil {
			return nil, fmt.Errorf("load env: %w", err)
		}
	}

	params := DatabaseParams{}
	if err := env.Parse(&params); err != nil {
		return nil, fmt.Errorf("parse env: %w", err)
	}

	config := &Config{
		HTTPPort: httpPort,
		DBParams: params,
	}

	return config, nil
}
