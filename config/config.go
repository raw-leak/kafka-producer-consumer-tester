package config

import (
	"log"
	"sync"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Seeds string `envconfig:"KAFKA_SEEDS"`
	Topic string `envconfig:"KAFKA_TOPIC"`
	Group string `envconfig:"KAFKA_Group"`
}

var (
	config Config
	once   sync.Once
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

// Get reads config from environment. Once.
func Get() (*Config, error) {
	once.Do(func() {
		err := envconfig.Process("", &config)
		if err != nil {
			log.Fatalf("Error processing config: %v", err)
		}
	})

	return &config, nil
}
