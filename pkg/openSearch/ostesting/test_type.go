// SPDX-FileCopyrightText: 2025 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package ostesting

import "time"

// TestType can be used as generic document object for testing
type TestType struct {
	ID               string    `json:"id"` // for easier identification in tests
	Text             string    `json:"text"`
	Keyword          string    `json:"keyword"`
	TextAndKeyword   string    `json:"textAndKeyword"`
	Integer          int       `json:"integer"`
	Float            float32   `json:"float"`
	Boolean          bool      `json:"boolean"`
	DateTimeStr      string    `json:"dateTimeStr,omitempty"`
	DateTime         time.Time `json:"dateTime"`
	KeywordOmitEmpty string    `json:"keywordOmitEmpty,omitempty"`
}

// testTypeMapping is an index mapping for testType
var testTypeMapping string = `{
		"mappings": {
			"properties": {
				"id": {
					"type": "keyword"
				},
				"text": {
					"type": "text"
				},
				"keyword": {
					"type": "keyword"
				},
				"textAndKeyword": {
					"type": "text",
					"fields": {
						"keyword": {
							"type": "keyword",
							"ignore_above": 256
						}
					}
				},
				"integer": {
					"type": "long"
				},
				"float": {
					"type": "float"
				},
				"boolean": {
					"type": "boolean"
				},
				"dateTimeStr": {
					"type": "date"
				},
				"dateTime": {
					"type": "date"
				},
				"keywordOmitEmpty": {
					"type": "keyword"
				}
			}
		}
	}`
