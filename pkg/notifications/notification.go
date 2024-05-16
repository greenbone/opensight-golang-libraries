// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package Notifications provides a client to communicate with the OpenSight Notification Service [github.com/greenbone/opensight-notification-service]
package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/greenbone/opensight-golang-libraries/pkg/retryableRequest"
)

const basePath = "/api/notification-service"
const createNotificationPath = "/notifications"

// Client can be used to send notifications
type Client struct {
	httpClient                 *http.Client
	notificationServiceAddress string
	maxRetries                 int
	retryWaitMin               time.Duration
	retryWaitMax               time.Duration
}

// Config configures the notification service client
type Config struct {
	Address      string
	MaxRetries   int
	RetryWaitMin time.Duration
	RetryWaitMax time.Duration
}

// NewClient returns a new [Client] with the notification service address (host:port) set.
// As httpClient you can use e.g. [http.DefaultClient].
func NewClient(httpClient *http.Client, config Config) *Client {
	return &Client{
		httpClient:                 httpClient,
		notificationServiceAddress: config.Address,
		maxRetries:                 config.MaxRetries,
		retryWaitMin:               config.RetryWaitMin,
		retryWaitMax:               config.RetryWaitMax,
	}
}

// CreateNotification sends a notification to the notification service.
// The request is retried up to the configured number of retries with an exponential backoff.
// So it can take some time until the functions returns.
func (c *Client) CreateNotification(ctx context.Context, notification Notification) error {
	notificationModel := toNotificationModel(notification)

	notificationSerialized, err := json.Marshal(notificationModel)
	if err != nil {
		return fmt.Errorf("failed to serialize notification object: %w", err)
	}

	// create request
	createNotificationEndpoint, err := url.JoinPath(c.notificationServiceAddress, basePath, createNotificationPath)
	if err != nil {
		return fmt.Errorf("invalid url '%s': %w", c.notificationServiceAddress, err)
	}

	request, err := http.NewRequest(http.MethodPost, createNotificationEndpoint, bytes.NewReader(notificationSerialized))
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	response, err := retryableRequest.ExecuteRequestWithRetry(ctx, c.httpClient, request, c.maxRetries, c.retryWaitMin, c.retryWaitMax)
	if err == nil {
		// note: the successful response returns the notification object, but we don't care about its values and omit parsing the body here
		response.Body.Close()
	}
	return err
}
