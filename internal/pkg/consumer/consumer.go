package consumer

import (
	"context"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Logger interface {
	Info(string)
	Infof(string, ...any)
	Error(string)
	Errorf(string, ...any)

	AddedPartition()
	RemovedPartition()
}

type Consumer struct {
	Seeds      []string
	Topic      string
	Group      string
	logger     Logger
	processors []*processor
}

type ConsumerConfig struct {
	Seeds []string
	Topic string
	Group string
}

func New(cfg ConsumerConfig, l Logger) *Consumer {
	return &Consumer{Seeds: cfg.Seeds, Group: cfg.Group, Topic: cfg.Topic, logger: l}
}

func (c *Consumer) Consume(callback func(chan [][]byte)) error {
	c.logger.Info("initializing consumer")

	p := newProcessor(callback, c.logger)

	cl, err := kgo.NewClient(
		kgo.SeedBrokers(c.Seeds...),
		kgo.ConsumeTopics(c.Topic),
		kgo.ConsumerGroup(c.Group),

		kgo.FetchMinBytes(1_000_000),    // Set minimum fetch bytes to ~1MB
		kgo.FetchMaxBytes(2_000_000),    // Set maximum fetch bytes to ~2MB
		kgo.FetchMaxWait(5*time.Second), // Wait up to 5 seconds if fetch.min.bytes not reached

		kgo.OnPartitionsAssigned(p.assigned),
		kgo.OnPartitionsRevoked(p.lostOrRevoked),
		kgo.OnPartitionsLost(p.lostOrRevoked),
		kgo.BlockRebalanceOnPoll(),
	)
	if err != nil {
		c.logger.Errorf("creating consumer client: %v", err)
		return err
	}
	if err := cl.Ping(context.Background()); err != nil {
		c.logger.Errorf("verifying consumer client connection: %v", err)
		return err
	}

	go p.run(cl)

	return nil
}

func (c *Consumer) Shutdown() {
	c.logger.Info("closing consumer")
	for _, p := range c.processors {
		p.Shutdown()
	}
}
