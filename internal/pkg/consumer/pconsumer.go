package consumer

import (
	"context"
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"
)

type pconsumer struct {
	cl        *kgo.Client
	topic     string
	partition int32

	quit chan struct{}
	done chan struct{}
	recs chan []*kgo.Record

	res chan [][]byte

	logger Logger

	consuming bool
}

func newPConsumer(cl *kgo.Client, topic string, partition int32, l Logger) *pconsumer {
	return &pconsumer{
		cl:        cl,
		topic:     topic,
		partition: partition,

		quit: make(chan struct{}),
		done: make(chan struct{}),
		recs: make(chan []*kgo.Record, 5),

		res: make(chan [][]byte),

		logger: l,

		consuming: false,
	}
}

func (pc *pconsumer) consume(wg *sync.WaitGroup) {
	defer close(pc.done)

	pc.logger.Infof("starting partition-consumer for  t %s p %d", pc.topic, pc.partition)
	defer pc.logger.Infof("closing partition-consumer for t %s p %d", pc.topic, pc.partition)

	wg.Add(1)
	defer wg.Done()

	pc.consuming = true

	for {
		select {
		case <-pc.quit:
			return
		case recs := <-pc.recs:
			parsed := [][]byte{}

			for _, record := range recs {
				parsed = append(parsed, record.Value)
			}

			pc.res <- parsed

			err := pc.cl.CommitRecords(context.Background(), recs...)
			if err != nil {
				pc.logger.Errorf("committing offsets with err: %v t: %s p: %d offset %d\n", err, pc.topic, pc.partition, recs[len(recs)-1].Offset+1)
			}
		}
	}
}

func (pc *pconsumer) shutdown() {
	pc.logger.Infof("stopping partition-consumer for t: %s p: %d\n", pc.topic, pc.partition)

	if pc.consuming {
		pc.consuming = false
		close(pc.quit)
		<-pc.done
	}
}
