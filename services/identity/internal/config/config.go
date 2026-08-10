package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	JWT_SECRET []byte
}

func NewConfig() *Config {
	_ = godotenv.Load()

	return &Config{
		JWT_SECRET: []byte(getEnvOrPanic("JWT_SECRET")),
	}
}

func getEnvOrPanic(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	msg := "You must pass " + key + " in .env"
	panic(msg)
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
