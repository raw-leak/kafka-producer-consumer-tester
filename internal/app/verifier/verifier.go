package verifier

import (
	"context"
	"encoding/json"
	"fmt"
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
		failedRecords:     sync.Map{},
		inProgressRecords: sync.Map{},
		successRecords:    sync.Map{},

		counts: struct {
			totalGenerated  int32
			totalFailed     int32
			totalInProgress int32
			totalSuccess    int32
		}{
			totalGenerated:  0,
			totalFailed:     0,
			totalInProgress: 0,
			totalSuccess:    0,
		},

		errs:    sync.Mutex{},
		errList: []string{},

		consumer: c,
		producer: p,
		logger:   l,
	}
}

func (v *Verifier) Verify() error {
	v.logger.Error("starting producer-consumer verification")
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
	if len(v.errList) > 0 {
		for _, errMsg := range v.errList {
			fmt.Println("TODO: print the error message", errMsg)
		}
	} else {
		fmt.Println("TODO: print that there are no unexpected errors")
	}

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
			fmt.Println("TODO: indicate the the message with ID X and state Y has no been stored correctly")
		} else {
			fmt.Println("OK ... TODO")
		}

		return true
	})
}

// partitionConsumer will be generated for each individual partition
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
	wg := sync.WaitGroup{}

	err := v.consumer.Consume(v.partitionConsumer)
	if err != nil {
		v.logger.Error("starting the consumer")
		return err
	}

	// produce events
	for i := 0; i < 100_000; i++ {
		ctx := context.Background()

		st := generateRandomState()
		id := generateRandomID()

		event := Event{ID: id, State: st}

		payload, err := json.Marshal(event)
		if err != nil {
			v.addUnexpectedError(err.Error())
			continue
		}

		err = v.producer.Produce(ctx, payload)
		if err != nil {
			v.addUnexpectedError(err.Error())
			continue
		}

		v.storeSentRecord(id, st)
	}

	wg.Wait()

	return nil
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

	v.logger.Info("waiting for records processing")

	tryCount := 0
	maxTries := 10

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
