package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	MySQLDSN  string `envconfig:"MYSQL_DSN" required:"true"`
	RabbitURL string `envconfig:"RABBIT_URL" required:"true"`
	Queue     string `envconfig:"RABBIT_QUEUE" default:"wallet_events"`
}

func Load() (*Config, error) {
	// Try to load .env if it's found
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	var c Config
	if err := envconfig.Process("", &c); err != nil {
		return nil, err
	}
	return &c, nil
}
