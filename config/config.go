package config

import (
	"fmt"
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
	config      Config
	once        sync.Once
	configError error
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

// Get reads config from environment. Once.
func Get() (*Config, error) {
	once.Do(func() {
		// Process the environment variables and capture the error
		err := envconfig.Process("", &config)
		if err != nil {
			configError = fmt.Errorf("error processing config: %v", err)
		}
	})

	return &config, configError
}
