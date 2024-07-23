// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package retryableRequest provides a function to retry an http request up to a configured number of retries. It builds on [retryablehttp], but works with the default [http.Client].
package retryableRequest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// ExecuteRequestWithRetry executes the given request via the passed http client and retries on failures.
// An error is only returned if there was a non retryable error or the maximum number of retries was reached.
// It uses the retry policy of [retryablehttp.ErrorPropagatedRetryPolicy] and exponential backoff from [retryablehttp.DefaultBackoff].
func ExecuteRequestWithRetry(ctx context.Context, client *http.Client, request *http.Request,
	maxRetries int, retryWaitMin, retryWaitMax time.Duration,
) (*http.Response, error) {
	var errList []error
	for attempt := range maxRetries + 1 {
		requestCopy, err := DeepCopyRequest(request)
		if err != nil {
			return nil, fmt.Errorf("failed to copy request for repeated use: %w", err)
		}

		response, err := client.Do(requestCopy.WithContext(ctx))

		if err == nil && IsOk(response.StatusCode) {
			return response, nil
		}

		// collect errors
		if err != nil {
			errList = append(errList, fmt.Errorf("attempt %d: failed to send request: %w",
				attempt+1, err)) // humans prefer to read one indexed
		} else {
			responseBody, err := io.ReadAll(response.Body)
			response.Body.Close()
			if err != nil {
				responseBody = []byte("can't display error details, failed to read response body: " + err.Error())
			}
			errList = append(errList, fmt.Errorf("attempt %d: request returned error: %s, %s",
				attempt+1, response.Status, string(responseBody)))
		}

		retry, retryErr := retryablehttp.ErrorPropagatedRetryPolicy(ctx, response, err)
		if !retry {
			return nil, fmt.Errorf("failed to execute request to %s, stop retrying after %d attempts due to %w, encountered errors: %w",
				request.URL.String(), attempt+1, retryErr, errors.Join(errList...))
		}

		if attempt < maxRetries {
			waitTime := retryablehttp.DefaultBackoff(retryWaitMin, retryWaitMax, attempt, response)
			time.Sleep(waitTime)
		}
	}
	return nil, fmt.Errorf("failed to execute request to %s after maximum number of %d attempts, encountered errors: %w",
		request.URL.String(), maxRetries+1, errors.Join(errList...))
}

// DeepCopyRequest returns a deep copy of the request. The context of the original request is placed by [context.Background].
// An error indicates a problem with the http request. In that case the passed request should no longer be used.
func DeepCopyRequest(request *http.Request) (copy *http.Request, err error) {
	copy = request.Clone(context.Background())

	// copy body, unlike stated by the docs, `request.Clone` does not deep copy the body
	var buf bytes.Buffer
	_, err = buf.ReadFrom(request.Body)
	if err != nil {
		return nil, fmt.Errorf("could not clone body - failed to read body: %w", err)
	}

	request.Body = io.NopCloser(&buf)
	copy.Body = io.NopCloser(bytes.NewReader(buf.Bytes()))

	return copy, nil
}

// IsOk returns true on a 2xx http status code
func IsOk(statusCode int) bool {
	return statusCode >= 200 && statusCode <= 299
}
