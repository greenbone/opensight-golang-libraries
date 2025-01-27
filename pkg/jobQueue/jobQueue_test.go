// SPDX-FileCopyrightText: 2024-2025 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package jobQueue

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type queueTestCase struct {
	test          func(t *testing.T, q *JobQueue)
	expectedCount int
}

func TestExecutor(t *testing.T) {
	tests := map[string]queueTestCase{
		"no request": {
			test: func(t *testing.T, q *JobQueue) {
				time.Sleep(50 * time.Millisecond)
			},
			expectedCount: 0,
		},
		"one request": {
			test: func(t *testing.T, q *JobQueue) {
				q.AddQueueRequest(Request{ID: "1"})
				time.Sleep(50 * time.Millisecond)
			},
			expectedCount: 1,
		},
		"two requests in a row": {
			test: func(t *testing.T, q *JobQueue) {
				q.AddQueueRequest(Request{ID: "1"})
				q.AddQueueRequest(Request{ID: "2"})
				time.Sleep(50 * time.Millisecond)
			},
			expectedCount: 1,
		},
		"add several requests while the first one is working": {
			test: func(t *testing.T, q *JobQueue) {
				q.AddQueueRequest(Request{ID: "1"})

				time.Sleep(50 * time.Millisecond)

				// Request 2 to 4 are added both without any delay

				q.AddQueueRequest(Request{ID: "2"})

				q.AddQueueRequest(Request{ID: "3"})

				q.AddQueueRequest(Request{ID: "4"})

				time.Sleep(50 * time.Millisecond)

				q.AddQueueRequest(Request{ID: "5"})

				time.Sleep(50 * time.Millisecond)
			},
			expectedCount: 3,
		},
	}

	// Here i will add several requests while the first one is working
	for name := range tests {
		t.Run(name, func(t *testing.T) {
			mu := sync.Mutex{}
			count := 0
			running := false
			q := NewJobQueue(jobTestFunc(t, &mu, &running, &count), context.Background())

			tests[name].test(t, q)

			assert.Equal(t, tests[name].expectedCount, count)
		})
	}
}

func jobTestFunc(t *testing.T, mu *sync.Mutex, running *bool, count *int) func() error {
	return func() error {
		mu.Lock()
		if *running {
			t.Errorf("Job is already running!")
		} else {
			*running = true
		}

		*count++
		mu.Unlock()
		time.Sleep(25 * time.Millisecond)

		mu.Lock()
		*running = false
		mu.Unlock()
		return nil
	}
}
