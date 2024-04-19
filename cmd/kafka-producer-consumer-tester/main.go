package main

import (
	"fmt"
	"kafka-producer-consumer-tester/config"
	"log"

	"kafka-producer-consumer-tester/internal/app/verifier"
	"kafka-producer-consumer-tester/internal/pkg/consumer"
	"kafka-producer-consumer-tester/internal/pkg/logger"
	"kafka-producer-consumer-tester/internal/pkg/producer"
)

func main() {
	err := run()
	if err != nil {
		fmt.Printf("ERROR: Application has failed to tart %v\n", err)
		panic(err)
	}
}
func run() error {
	log.Println("INFO: Starting application")
	defer log.Println("INFO: Application gracefully stopped")

	logger, err := logger.New()
	if err != nil {
		log.Printf("ERROR: Failed to initialize logger: %v\n", err)
		return err
	}

	defer func() {
		logger.Shutdown()
	}()

	cfg, err := config.Get()
	if err != nil {
		logger.Errorf("loading configuration: %v", err)
		return err
	}

	p, err := producer.New(producer.ProducerConfig{
		Seeds: []string{cfg.Seeds},
		Topic: cfg.Topic,
		Group: cfg.Group,
	}, logger)
	if err != nil {
		logger.Errorf("initializing producer: %v", err)
		return err
	}
	defer func() {
		p.Shutdown()
		logger.Infof("producer shutted down")
	}()

	c := consumer.New(consumer.ConsumerConfig{
		Seeds: []string{cfg.Seeds},
		Topic: cfg.Topic,
		Group: cfg.Group,
	}, logger)
	defer func() {
		c.Shutdown()
		logger.Infof("consumer shutted down")
	}()

	v := verifier.New(p, c, logger)

	err = v.Verify()
	if err != nil {
		logger.Errorf("verification: %v", err)
		return err
	}

	return nil
}
