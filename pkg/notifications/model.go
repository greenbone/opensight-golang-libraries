// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package notifications

// Notification is the object which can be sent to the notification service.
// It is defined in notification service REST API: https://github.com/greenbone/opensight-notification-service/tree/main/api/notificationservice
type Notification struct {
	// omit property `Id` here, as it is read only
	Origin       string         `json:"origin"`
	OriginUri    string         `json:"originUri,omitempty"` // can be used to provide a link to the origin
	Timestamp    string         `json:"timestamp" format:"date-time"`
	Title        string         `json:"title"` // can also be seen as the 'type'
	Detail       string         `json:"detail"`
	Level        Level          `json:"level"`
	CustomFields map[string]any `json:"customFields,omitempty"` // can contain arbitrary structured information about the notification
}

// Level describes the severity of the notification
type Level string

const (
	LevelInfo     Level = "info"
	LevelWarning  Level = "warning"
	LevelError    Level = "error"
	LevelCritical Level = "critical"
)
