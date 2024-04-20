package verifier

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

const (
	Failed     = "failed"
	InProgress = "in-progress"
	Success    = "success"
)

type Producer interface {
	Produce(context.Context, []byte) error
	ProduceBatch(ctx context.Context, payloads [][]byte) error
}

type Consumer interface {
	Consume(func(chan [][]byte)) error
}

type Logger interface {
	RecordSent(string)
	RecordProcessed(string)
	Info(string)
	Error(string)

	Infof(string, ...any)

	AddedProcessor()
	RemovedProcessor()
}

type EventState struct {
	ID    string
	Count int // times a single event was consumed
}

type Event struct {
	ID    string
	State string
}

type Verifier struct {
	generatedRecords sync.Map

	failedRecords     sync.Map
	inProgressRecords sync.Map
	successRecords    sync.Map

	counts struct {
		totalGenerated  int32
		totalFailed     int32
		totalInProgress int32
		totalSuccess    int32
	}

	errs    sync.Mutex
	errList []string

	consumer Consumer
	producer Producer
	logger   Logger
}

func New(p Producer, c Consumer, l Logger) *Verifier {
	return &Verifier{
		consumer: c,
		producer: p,
		logger:   l,

		generatedRecords: sync.Map{},

		failedRecords:     sync.Map{},
		inProgressRecords: sync.Map{},
		successRecords:    sync.Map{},

		counts: struct {
			totalGenerated  int32
			totalFailed     int32
			totalInProgress int32
			totalSuccess    int32
		}{},

		errs:    sync.Mutex{},
		errList: []string{},
	}
}

func (v *Verifier) Verify() error {
	v.logger.Info("starting producer-consumer verification")
	err := v.startVerification()
	if err != nil {
		return err
	}

	v.waitForCompletion()
	v.logger.Error("producer-consumer verification completed")

	// v.printResult()

	return nil
}

func (v *Verifier) printResult() {
	log.Printf("%d unexpected errors detected\n", len(v.errList))

	v.generatedRecords.Range(func(key, value interface{}) bool {
		state := value.(string)
		id := key.(string)

		var targetMap *sync.Map

		switch state {
		case Failed:
			targetMap = &v.failedRecords
		case InProgress:
			targetMap = &v.inProgressRecords
		case Success:
			targetMap = &v.successRecords
		default:
			return true
		}

		_, ok := targetMap.Load(id)

		if !ok {
			log.Printf("message with ID %s has not been found in the %s state bucket", id, state)
		}

		return true
	})
}

func (v *Verifier) partitionConsumer(res chan [][]byte) {
	go func() {
		v.logger.AddedProcessor()
		defer v.logger.RemovedProcessor()

		for msgs := range res {
			for _, msg := range msgs {

				var e Event
				if err := json.Unmarshal(msg, &e); err != nil {
					v.addUnexpectedError(err.Error())
					continue
				}

				v.storeProcessedRecord(e.ID, e.State)
			}
		}
	}()
}

func (v *Verifier) startVerification() error {
	err := v.consumer.Consume(v.partitionConsumer)
	if err != nil {
		v.logger.Error("starting the consumer")
		return err
	}

	v.produceMessages()

	return nil
}
func (v *Verifier) produceMessages() {
	for i := 0; i < 1000; i++ {
		ctx := context.Background()

		payloads := make([][]byte, 1000)
		events := []Event{}

		for y := 0; y < 1_000; y++ {
			st := generateRandomState()
			id := generateRandomID()

			event := Event{ID: id, State: st}

			payload, err := json.Marshal(event)
			if err != nil {
				v.addUnexpectedError(err.Error())
				continue
			}

			payloads = append(payloads, payload)
			events = append(events, event)
		}

		err := v.producer.ProduceBatch(ctx, payloads)
		if err != nil {
			v.addUnexpectedError(err.Error())
			continue
		}

		for _, e := range events {
			v.storeSentRecord(e.ID, e.State)
		}

	}
}
func (v *Verifier) storeSentRecord(id, st string) {
	v.generatedRecords.Store(id, st)
	atomic.AddInt32(&v.counts.totalGenerated, 1)
	v.logger.RecordSent(st)
}

func (v *Verifier) storeProcessedRecord(id, st string) {
	var targetMap *sync.Map

	switch st {
	case Failed:
		targetMap = &v.failedRecords
		atomic.AddInt32(&v.counts.totalFailed, 1)
	case InProgress:
		targetMap = &v.inProgressRecords
		atomic.AddInt32(&v.counts.totalInProgress, 1)
	case Success:
		targetMap = &v.successRecords
		atomic.AddInt32(&v.counts.totalSuccess, 1)
	default:
		return
	}

	val, ok := targetMap.LoadOrStore(id, &EventState{ID: id, Count: 1})
	v.logger.RecordProcessed(st)

	if ok {
		eventState := val.(*EventState)
		eventState.Count++

		targetMap.Store(id, eventState)
	}
}

func (v *Verifier) addUnexpectedError(errMsg string) {
	v.errs.Lock()
	defer v.errs.Unlock()
	v.errList = append(v.errList, errMsg)
}

func (v *Verifier) waitForCompletion() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	v.logger.Info("waiting records to be processed")

	tryCount := 0
	maxTries := 60

	for range ticker.C {
		if v.allMessagesProcessed() {
			v.logger.Info("all records has been stored")
			return
		}

		tryCount++
		if tryCount >= maxTries {
			return
		}
	}
}

func (v *Verifier) allMessagesProcessed() bool {
	totalProcessed := atomic.LoadInt32(&v.counts.totalFailed) + atomic.LoadInt32(&v.counts.totalInProgress) + atomic.LoadInt32(&v.counts.totalSuccess)
	return atomic.LoadInt32(&v.counts.totalGenerated) == totalProcessed
}

func generateRandomState() string {
	states := []string{Success, Failed, InProgress}
	return states[rand.Intn(len(states))]
}

func generateRandomID() string {
	return uuid.NewString()
}
