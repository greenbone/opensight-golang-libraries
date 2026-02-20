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
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/greenbone/opensight-golang-libraries/pkg/auth"
	"github.com/greenbone/opensight-golang-libraries/pkg/retryableRequest"
)

const (
	basePath               = "/api/notification-service"
	createNotificationPath = "/notifications"
	registerOriginsPath    = "/origins"
)

// Client can be used to send notifications
type Client struct {
	httpClient                 *http.Client
	authClient                 *auth.KeycloakClient
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
func NewClient(httpClient *http.Client, config Config, authCfg auth.KeycloakConfig) *Client {
	authClient := auth.NewKeycloakClient(httpClient, authCfg)

	return &Client{
		httpClient:                 httpClient,
		authClient:                 authClient,
		notificationServiceAddress: config.Address,
		maxRetries:                 config.MaxRetries,
		retryWaitMin:               config.RetryWaitMin,
		retryWaitMax:               config.RetryWaitMax,
	}
}

// CreateNotification sends a notification to the notification service.
// It is retried up to the configured number of retries with an exponential backoff,
// So it can take some time until the functions returns.
func (c *Client) CreateNotification(ctx context.Context, notification Notification) error {
	token, err := c.authClient.GetToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get authentication token: %w", err)
	}

	notificationModel := toNotificationModel(notification)
	notificationSerialized, err := json.Marshal(notificationModel)
	if err != nil {
		return fmt.Errorf("failed to serialize notification object: %w", err)
	}

	createNotificationEndpoint, err := url.JoinPath(c.notificationServiceAddress, basePath, createNotificationPath)
	if err != nil {
		return fmt.Errorf("invalid url '%s': %w", c.notificationServiceAddress, err)
	}

	req, err := http.NewRequest(http.MethodPost, createNotificationEndpoint,
		bytes.NewReader(notificationSerialized))
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	response, err := retryableRequest.ExecuteRequestWithRetry(ctx, c.httpClient, req,
		c.maxRetries, c.retryWaitMin, c.retryWaitMax)
	if err == nil {
		// note: the successful response returns the notification object, but we don't care about its values and omit parsing the body here
		_ = response.Body.Close()
	}

	return err
}

func (c *Client) RegisterOrigins(ctx context.Context, serviceID string, origins []Origin) error {
	token, err := c.authClient.GetToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get authentication token: %w", err)
	}

	originsSerialized, err := json.Marshal(origins)
	if err != nil {
		return fmt.Errorf("failed to serialize origins: %w", err)
	}

	registerOriginsEndpoint, err := url.JoinPath(c.notificationServiceAddress, basePath, registerOriginsPath, serviceID)
	if err != nil {
		return fmt.Errorf("invalid url '%s': %w", c.notificationServiceAddress, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, registerOriginsEndpoint, bytes.NewReader(originsSerialized))
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	response, err := c.httpClient.Do(req) //nolint:gosec
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode < 200 || response.StatusCode > 299 {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return fmt.Errorf("failed to register origins, status: %s: %w body: %s", response.Status, err, string(body))
		}
		return fmt.Errorf("failed to register origins, status: %s: %s", response.Status, string(body))
	}

	return nil
}
