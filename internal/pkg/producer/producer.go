package producer

import (
	"context"
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Logger interface {
	Info(string)
	Infof(string, ...any)
	Error(string)
	Errorf(string, ...any)
}

type Producer struct {
	topic  string
	client *kgo.Client
	logger Logger
}

type ProducerConfig struct {
	Seeds []string
	Topic string
	Group string
}

func New(cfg ProducerConfig, l Logger) (*Producer, error) {
	l.Info("initializing producer")

	cl, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Seeds...),
		kgo.DefaultProduceTopic(cfg.Topic),
	)
	if err != nil {
		l.Errorf("creating producer client: %v", err)
		return nil, err
	}
	if err := cl.Ping(context.Background()); err != nil {
		l.Errorf("verifying producer client connection: %v", err)
		return nil, err
	}

	return &Producer{client: cl, topic: cfg.Topic, logger: l}, nil
}

func (p *Producer) ProduceBatch(ctx context.Context, payloads []string) error {
	var records []*kgo.Record

	for _, payload := range payloads {
		records = append(records, &kgo.Record{Topic: p.topic, Value: []byte(payload)})
	}

	err := p.client.ProduceSync(ctx, records...).FirstErr()
	return err
}

func (p *Producer) Produce(ctx context.Context, payload []byte) (err error) {
	record := &kgo.Record{Value: payload}
	wg := sync.WaitGroup{}

	wg.Add(1)
	p.client.Produce(ctx, record, func(r *kgo.Record, prErr error) {
		err = prErr
		wg.Done()
	})

	wg.Wait()
	return
}

func (p *Producer) Shutdown() {
	p.logger.Info("closing producer")
	p.client.Close()
}
