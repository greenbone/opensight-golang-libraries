package jobQueue

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
)

// The job queue is a thread-safe queue of import requests.
// When a request is added to the queue, it is processed immediately.
// If a request is actually running, its waiting until the queue is empty
// If several requests are waiting, only the last one is processed

type Request struct {
	ID string
}

type jobQueue struct {
	reqChan  chan Request
	execFunc func() error
	mu       sync.Mutex
	context  context.Context
}

func NewJobQueue(execFunc func() error, context context.Context) *jobQueue {
	q := jobQueue{
		reqChan:  make(chan Request, 100),
		execFunc: execFunc,
		context:  context,
	}
	go q.processRequests()
	return &q
}

func (q *jobQueue) AddQueueRequest(req Request) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.reqChan <- req
}

func (q *jobQueue) processRequests() {
	for {
		select {
		case <-q.context.Done():
			close(q.reqChan)
			log.Info().Msgf("closing jobQueue and consumer")
			return
		case req, ok := <-q.reqChan:
			if !ok {
				log.Warn().Msgf("<-q.reqChan returned not ok")
				return
			}

			// drain the queue and read the latest request
			for len(q.reqChan) > 0 {
				req, ok = <-q.reqChan
				if !ok {
					return
				}
			}

			// execute the latest request
			q.execute(req)
		}
	}
}

func (q *jobQueue) execute(req Request) {
	log.Debug().Msgf("Executing queue request ID: %s\n", req.ID)
	// Call the function
	err := q.execFunc()
	if err != nil {
		log.Error().Msgf("Error executing queue request ID: %s %v\n", req.ID, err)
	}
	log.Debug().Msgf("Finished queue request ID: %s\n", req.ID)
}
