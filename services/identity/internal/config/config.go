package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	JWT_SECRET []byte
}

func NewConfig() *Config {
	if err := godotenv.Load(); err != nil {
		panic("Can't load .env file")
	}

	return &Config{
		JWT_SECRET: []byte(getEnvOrPanic("JWT_SECRET")),
	}
}

func getEnvOrPanic(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	panic("You must pass key in .env")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}
