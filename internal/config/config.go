package config

import (
	"log"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort    string `env:"SERVER_PORT,required"`
	BasePath      string `env:"BASE_PATH,required"`
	RequestPath   string `env:"REQUEST_PATH,required"`
	CertFile      string `env:"CERT_FILE,required"`
	KeyFile       string `env:"KEY_FILE,required"`
	LogFile       string `env:"LOG_FILE,required"`
	AllowedClient string `env:"ALLOWED_CLIENT"`
}

func LoadConfig() *Config {
	err := godotenv.Load()

	if err != nil {
		log.Fatalf("Unable to load .env: %e", err)
	}

	config := &Config{}

	if err := env.Parse(config); err != nil {
		log.Fatalf("Unable to load variables from .env: %e", err)
	}

	return config
}
