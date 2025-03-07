// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package notifications

import "time"

type Notification struct {
	// omit property `Id` here, as it is read only
	Origin       string
	OriginUri    string // can be used to provide a link to the origin
	Timestamp    time.Time
	Title        string // can also be seen as the 'type'
	Detail       string
	Level        Level
	CustomFields map[string]any // can contain arbitrary structured information about the notification
}

// notification is the object which is to the notification service.
// It is defined in notification service REST API: https://github.com/greenbone/opensight-notification-service/tree/main/api/notificationservice
type notificationModel struct {
	// omit property `Id` here, as it is read only
	Origin       string         `json:"origin"`
	OriginUri    string         `json:"originUri,omitempty"`
	Timestamp    string         `json:"timestamp" format:"date-time"`
	Title        string         `json:"title"`
	Detail       string         `json:"detail"`
	Level        Level          `json:"level"`
	CustomFields map[string]any `json:"customFields,omitempty"`
}

// Level describes the severity of the notification
type Level string

const (
	LevelInfo    Level = "info"
	LevelWarning Level = "warning"
	LevelError   Level = "error"
)

func toNotificationModel(n Notification) notificationModel {
	return notificationModel{
		Origin:       n.Origin,
		OriginUri:    n.OriginUri,
		Timestamp:    n.Timestamp.UTC().Format(time.RFC3339Nano),
		Title:        n.Title,
		Detail:       n.Detail,
		Level:        n.Level,
		CustomFields: n.CustomFields,
	}
}
