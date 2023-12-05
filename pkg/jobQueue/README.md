# jobQueue

This package provides a job queue that can be used to execute a function in a thread-safe manner. The job queue is designed to be used in situations where multiple requests need to be processed, but only one request can be processed at a time.

## Example Usage

Here is an example of how to use the job queue:

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/greenbone/opensight-golang-libraries/jobQueue"
)

func main() {
	// Define the function that will be executed for each request
	execFunc := func() error {
		fmt.Println("Processing request")
		return nil
	}

	// Create a new job queue with the defined function and a context with a 5 second timeout
	q := jobQueue.NewJobQueue(execFunc, context.WithTimeout(context.Background(), 5*time.Second))

	// Add several requests to the queue
	for i := 0; i < 10; i++ {
		req := jobQueue.Request{ID: fmt.Sprintf("Request %d", i)}
		q.AddQueueRequest(req)
	}

	// Wait for the queue to finish processing all requests
	q.Wait()
}
```

In this example, 10 requests are added to the queue, and the function is executed for each request one after the other. The Wait function is called to wait for all requests to be processed before the 
program exits.
