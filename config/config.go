package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	ServerPort  string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		ServerPort:  port,
	}
}
