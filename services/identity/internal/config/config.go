package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type PostgresConfig struct {
	User     string `env:"POSTGRES_USER,required"`
	Password string `env:"POSTGRES_PASSWORD,required"`
	Host     string `env:"POSTGRES_HOST,required"`
	Port     string `env:"POSTGRES_PORT,required"`
	DBName   string `env:"POSTGRES_DB,required"`
}

func (p *PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s",
		p.User,
		p.Password,
		p.Host,
		p.Port,
		p.DBName,
	)
}

type AppConfig struct {
	Host string `env:"HOST"`
	Port string `env:"PORT" envDefault:"8080"`
}

func (c *AppConfig) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

type Config struct {
	JWTSecret []byte `env:"JWTSecret,required"`

	App AppConfig
	DB  PostgresConfig
}

func NewConfig() (*Config, error) {
	_ = godotenv.Load()

	cfg, err := env.ParseAs[Config]()

	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}
