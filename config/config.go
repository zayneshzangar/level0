package config

import (
	"fmt"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

type (
	DatabaseType string
	Config       struct {
		App   App
		DB    Database
		Front Frontend
		Kafka Kafka
	}

	App struct {
		Name        string `env:"APP_NAME" envDefault:"auth"`
		Port        string `env:"APP_PORT" envDefault:":8080"`
		Version     string `env:"APP_VERSION" envDefault:"1.0.0"`
		Environment string `env:"APP_ENV" envDefault:"dev"`
	}

	Database struct {
		Type     DatabaseType `env:"DB_TYPE" envDefault:"postgres"`
		SType    string       `env:"DB_TYPE",required`
		Host     string       `env:"DB_HOST",required`
		User     string       `env:"DB_USER",required`
		Password string       `env:"DB_PASSWORD",required`
		Name     string       `env:"DB_NAME",required`
		Port     string       `env:"DB_PORT",required`
		Mode     string       `env:"DB_SSLMODE",required`
	}

	Frontend struct {
		Host string `env:"FRONT_HOST",required`
		Port string `env:"FRONT_PORT",required`
	}

	Kafka struct {
		Host    string `env:"KAFKA_HOST",required`
		Port1   string `env:"KAFKA_PORT_1" default:"9091"`
		Port2   string `env:"KAFKA_PORT_2" default:"9092"`
		Port3   string `env:"KAFKA_PORT_3" default:"9093"`
		Topic   string `env:"KAFKA_TOPIC" default:"orders"`
		GroupName string `env:"KAFKA_GROUP_NAME" default:"order-group"`
	}
)

const (
	Postgres DatabaseType = "postgres"
	Mongo    DatabaseType = "mongo"
)

func Load() (*Config, error) {
	// запуск .env
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error while load .env file: %w", err)
	}

	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
