// SPDX-FileCopyrightText: 2024-2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package jobQueue

import (
	"context"
	"sync"

	"github.com/rs/zerolog/log"
)

// Request is a request to be processed by the queue and allows to provide an ID for identification
type Request struct {
	ID string
}

// JobQueue is a thread-safe queue of requests to execute a predefined function.
type JobQueue struct {
	reqChan  chan Request
	execFunc func() error
	mu       sync.Mutex
	context  context.Context
}

// NewJobQueue creates a new job queue
// execFunc is the function to be executed for each request that is processed
// context is the context of the caller
func NewJobQueue(execFunc func() error, context context.Context) *JobQueue {
	q := JobQueue{
		reqChan:  make(chan Request, 100),
		execFunc: execFunc,
		context:  context,
	}
	go q.processRequests()
	return &q
}

// AddQueueRequest adds a request to the queue
//
// The job queue is designed to be used in situations where multiple requests of the same type need to be processed,
// but only one request can be processed at a time and only the most recent request needs to be processed.
//
// If a request is added to the queue while another request is being processed, the new request will be added to the queue
// and processed after the current request has finished.
// If there is already a request in the queue, the old request will be considered obsolete and replaced by the new request.
// [jobQueue_test.go](jobQueue_test.go) illustrates this behaviour.
func (q *JobQueue) AddQueueRequest(req Request) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.reqChan <- req
}

func (q *JobQueue) processRequests() {
	for {
		select {
		case <-q.context.Done():
			close(q.reqChan)
			log.Info().Msg("closing JobQueue and consumer")
			return
		case req, ok := <-q.reqChan:
			if !ok {
				log.Warn().Msg("<-q.reqChan returned not ok")
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

func (q *JobQueue) execute(req Request) {
	log.Debug().
		Str("request_id", req.ID).
		Msgf("executing queue request id: %s", req.ID)
	// Call the function
	err := q.execFunc()
	if err != nil {
		log.Error().Err(err).
			Str("request_id", req.ID).
			Msgf("error executing queue request id: %s", req.ID)
	}
	log.Debug().
		Str("request_id", req.ID).
		Msgf("finished queue request id: %s", req.ID)
}
