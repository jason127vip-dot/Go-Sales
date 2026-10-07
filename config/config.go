package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL  string
	ServerPort   string
	OpenAIAPIKey string
	OpenAIModel  string
}

func LoadConfig() *Config {
	_ = godotenv.Load()

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-6-luna"
	}

	return &Config{
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		ServerPort:   port,
		OpenAIAPIKey: os.Getenv("OPENAI_API_KEY"),
		OpenAIModel:  model,
	}
}
