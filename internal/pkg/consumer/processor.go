package consumer

import (
	"context"
	"fmt"
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"
)

type tp struct {
	t string
	p int32
}

type processor struct {
	callback  func(chan [][]byte)
	consumers map[tp]*pconsumer
	logger    Logger
	enabled   bool
	wg        *sync.WaitGroup
}

func newProcessor(callback func(chan [][]byte), l Logger) *processor {
	return &processor{
		callback:  callback,
		consumers: make(map[tp]*pconsumer),
		logger:    l,
		enabled:   true,
		wg:        &sync.WaitGroup{},
	}
}

func (p *processor) run(cl *kgo.Client) {
	for p.enabled {
		fetches := cl.PollRecords(context.Background(), 10_000)
		if fetches.IsClientClosed() {
			return
		}

		fetches.EachPartition(func(part kgo.FetchTopicPartition) {
			if len(part.Records) > 0 {
				p.sendToPartition(tp{part.Topic, part.Partition}, part.Records)
			}
		})

		cl.AllowRebalance()
	}
}

func (p *processor) assigned(_ context.Context, cl *kgo.Client, assigned map[string][]int32) {
	for topic, partitions := range assigned {
		for _, partition := range partitions {

			p.logger.AddedPartition()

			pc := newPConsumer(cl, topic, partition, p.logger)

			p.consumers[tp{topic, partition}] = pc

			p.callback(pc.res)

			go pc.consume(p.wg)
		}
	}
}

func (p *processor) lostOrRevoked(_ context.Context, cl *kgo.Client, lost map[string][]int32) {
	var wg sync.WaitGroup
	defer wg.Wait()

	for topic, partitions := range lost {
		for _, partition := range partitions {
			p.logger.Info("terminating goroutine for a partition")

			p.logger.RemovedPartition()

			tp := tp{topic, partition}

			pc := p.consumers[tp]

			fmt.Printf("waiting for work to finish t %s p %d\n", topic, partition)
			wg.Add(1)
			pc.shutdown()

			delete(p.consumers, tp)

			go func() { <-pc.done; wg.Done() }()
		}
	}
}

func (p *processor) sendToPartition(tp tp, records []*kgo.Record) {
	p.consumers[tp].recs <- records
}

func (p *processor) Shutdown() {
	p.enabled = false

	for _, c := range p.consumers {
		c.shutdown()
	}

	p.wg.Wait()
}
