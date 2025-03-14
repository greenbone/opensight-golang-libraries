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
	"strings"
	"time"

	"github.com/greenbone/opensight-golang-libraries/pkg/retryableRequest"
)

const (
	basePath               = "/api/notification-service"
	createNotificationPath = "/notifications"
)

// Client can be used to send notifications
type Client struct {
	httpClient                 *http.Client
	notificationServiceAddress string
	maxRetries                 int
	retryWaitMin               time.Duration
	retryWaitMax               time.Duration
	authentication             Authentication
}

// Config configures the notification service client
type Config struct {
	Address      string
	MaxRetries   int
	RetryWaitMin time.Duration
	RetryWaitMax time.Duration
}

type Authentication struct {
	ClientID     string
	ClientSecret string
	URL          string
}

// NewClient returns a new [Client] with the notification service address (host:port) set.
// As httpClient you can use e.g. [http.DefaultClient].
func NewClient(httpClient *http.Client, config Config, authentication Authentication) *Client {
	return &Client{
		httpClient:                 httpClient,
		notificationServiceAddress: config.Address,
		maxRetries:                 config.MaxRetries,
		retryWaitMin:               config.RetryWaitMin,
		retryWaitMax:               config.RetryWaitMax,
		authentication:             authentication,
	}
}

// CreateNotification sends a notification to the notification service.
// The request is authenticated, serialized, and sent via an HTTP POST request.
// It is retried up to the configured number of retries with an exponential backoff,
// so the function may take some time to return.
func (c *Client) CreateNotification(ctx context.Context, notification Notification) error {
	// Get authentication token
	token, err := c.GetAuthenticationToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get authentication token: %w", err)
	}

	// serialize notification
	notificationModel := toNotificationModel(notification)
	notificationSerialized, err := json.Marshal(notificationModel)
	if err != nil {
		return fmt.Errorf("failed to serialize notification object: %w", err)
	}

	createNotificationEndpoint, err := url.JoinPath(c.notificationServiceAddress, basePath, createNotificationPath)
	if err != nil {
		return fmt.Errorf("invalid url '%s': %w", c.notificationServiceAddress, err)
	}

	req, err := http.NewRequest(http.MethodPost, createNotificationEndpoint, bytes.NewReader(notificationSerialized))
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	response, err := retryableRequest.ExecuteRequestWithRetry(ctx, c.httpClient, req,
		c.maxRetries, c.retryWaitMin, c.retryWaitMax)
	if err == nil {
		response.Body.Close()
	}

	return err
}

// GetAuthenticationToken retrieves an authentication token using client credentials.
// It constructs a form-encoded request, sends it with retry logic, and parses the response.
func (c *Client) GetAuthenticationToken(ctx context.Context) (string, error) {
	data := url.Values{}
	data.Set("client_id", c.authentication.ClientID)
	data.Set("client_secret", c.authentication.ClientSecret)
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest(http.MethodPost, c.authentication.URL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create authentication request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := retryableRequest.ExecuteRequestWithRetry(ctx, c.httpClient, req,
		c.maxRetries, c.retryWaitMin, c.retryWaitMax)
	if err != nil {
		return "", fmt.Errorf("failed to execute authentication request with retry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("authentication request failed with status: %d", resp.StatusCode)
	}

	// parse JSON response to extract the access token
	// only access token is needed from the response
	var authResponse struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return "", fmt.Errorf("failed to parse authentication response: %w", err)
	}

	return authResponse.AccessToken, nil
}
