package config

import (
	"os"

	"github.com/joho/godotenv"
)

type PostgresConfig struct {
	DATABASE_URL string
}

type Config struct {
    ADDRESS_AUTH_SERVICE string
    ADDRESS_CATALOG_SERVICE string
	JWT_SECRET []byte
	DB         PostgresConfig
}

func NewConfig() *Config {
	_ = godotenv.Load()

	return &Config{
        ADDRESS_AUTH_SERVICE: getEnvOrPanic("ADDRESS_AUTH_SERVICE"),
        ADDRESS_CATALOG_SERVICE: getEnvOrPanic("ADDRESS_CATALOG_SERVICE"),
        JWT_SECRET: []byte(getEnvOrPanic("JWT_SECRET")),
		DB: PostgresConfig{
			DATABASE_URL: "postgresql://" + getEnvOrPanic("POSTGRES_USER") + ":" + getEnvOrPanic("POSTGRES_PASSWORD") + "@localhost:5432/" + getEnvOrPanic("POSTGRES_DB"),
		},
	}
}

func getEnvOrPanic(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	msg := "You must pass " + key + " in .env"
	panic(msg)
}
