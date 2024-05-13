// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package Notifications provides a client to communicate with the OpenSight Notification Service [github.com/greenbone/opensight-notification-service]
package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const basePath = "/api/notification-service"
const createNotificationPath = "/notifications"

// Client can be used to send notifications
type Client struct {
	httpClient                 *http.Client
	notificationServiceAddress string
}

// NewClient returns a new [Client] with the notification service address (host:port) set.
// As httpClient you can use e.g. [http.DefaultClient].
func NewClient(httpClient *http.Client, notificationServiceAddress string) *Client {
	return &Client{
		httpClient:                 httpClient,
		notificationServiceAddress: notificationServiceAddress,
	}
}

// CreateNotification sends a notification to the notification service
func (c *Client) CreateNotification(notification Notification) error {
	var notificationSerialized bytes.Buffer
	err := json.NewEncoder(&notificationSerialized).Encode(notification)
	if err != nil {
		return fmt.Errorf("failed to serialize notification object: %w", err)
	}

	// create request
	createNotificationEndpoint, err := url.JoinPath(c.notificationServiceAddress, basePath, createNotificationPath)
	if err != nil {
		return fmt.Errorf("invalid url '%s': %w", c.notificationServiceAddress, err)
	}
	request, err := http.NewRequest(http.MethodPost, createNotificationEndpoint, &notificationSerialized)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	// execute request
	response, err := c.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send request to %s: %w", createNotificationEndpoint, err)
	}
	defer response.Body.Close()

	if response.StatusCode > 299 || response.StatusCode < 200 {
		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			responseBody = []byte("can't display error details, failed to read response body: " + err.Error())
		}
		return fmt.Errorf("request to %s returned error: %s, %s", createNotificationEndpoint, response.Status, string(responseBody))
	}

	// note: the successful response returns the notification object, but we don't care about its values and omit parsing the body here

	return nil
}
