// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package notifications

import "time"

type Notification struct {
	// omit property `Id` here, as it is read only
	Origin           string    // name of the origin, e.g. `SBOM - React`
	OriginClass      string    // unique identifier for the class of origins, e.g. `/vi/SBOM`
	OriginResourceID string    // optional, together with class it can be used to provide a link to the origin, e.g. `<id of react sbom object>`
	Timestamp        time.Time // client will set timestamp if not set
	Title            string    // can also be seen as the 'type'
	Detail           string
	Level            Level
	CustomFields     map[string]any // optional, can contain arbitrary structured information about the notification
}

// notification is the object which is to the notification service.
// It is defined in notification service REST API: https://github.com/greenbone/opensight-notification-service/tree/main/api/notificationservice
type notificationModel struct {
	// omit property `Id` here, as it is read only
	Origin           string         `json:"origin"`
	OriginClass      string         `json:"originClass"`
	OriginResourceID string         `json:"originResourceID,omitempty"`
	Timestamp        string         `json:"timestamp" format:"date-time"`
	Title            string         `json:"title"`
	Detail           string         `json:"detail"`
	Level            Level          `json:"level"`
	CustomFields     map[string]any `json:"customFields,omitempty"`
}

// Level describes the severity of the notification
type Level string

const (
	LevelInfo    Level = "info"
	LevelWarning Level = "warning"
	LevelError   Level = "error"
)

// toNotificationModel converts a Notification to the rest model for the notification service.
// It also adds the current time as timestamp if it is not set.
func toNotificationModel(n Notification) notificationModel {
	if n.Timestamp.IsZero() {
		n.Timestamp = time.Now().UTC()
	}

	return notificationModel{
		Origin:           n.Origin,
		OriginClass:      n.OriginClass,
		OriginResourceID: n.OriginResourceID,
		Timestamp:        n.Timestamp.UTC().Format(time.RFC3339Nano),
		Title:            n.Title,
		Detail:           n.Detail,
		Level:            n.Level,
		CustomFields:     n.CustomFields,
	}
}

type Origin struct {
	Name  string `json:"name"`
	Class string `json:"class"`
}
