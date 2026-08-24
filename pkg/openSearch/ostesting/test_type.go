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
// Note: For test full text search filters, it uses a custom english analyzer which preserves negation,
// for the default see https://docs.opensearch.org/latest/analyzers/language-analyzers/english
// and https://docs.opensearch.org/latest/analyzers/token-filters/stop/
var testTypeMapping string = `{
		"settings": {
			"index.max_ngram_diff": 7,
			"analysis": {
				"filter": {
					"english_stop_without_negation": {
						"type": "stop",
						"stopwords": ["a", "an", "and", "are", "as", "at", "be", "but", "by", "for", "if", "in", "into", "is", "it", "of", "on", "or", "such", "that", "the", "their", "then", "there", "these", "they", "this", "to", "was", "will", "with"]
					},
					"english_possessive_stemmer": {
						"type": "stemmer",
						"language": "possessive_english"
					},
					"english_stemmer": {
						"type": "stemmer",
						"language": "english"
					},
					"substring_grams": {
						"type": "ngram",
						"min_gram": 3,
						"max_gram": 10,
						"preserve_original": true
					}
				},
				"analyzer": {
					"english_preserving_negation": {
						"type": "custom",
						"tokenizer": "standard",
						"filter": ["english_possessive_stemmer", "lowercase", "english_stop_without_negation", "english_stemmer"]
					},
					"substring_index": {
						"type": "custom",
						"tokenizer": "standard",
						"filter": ["lowercase", "asciifolding", "english_stop_without_negation", "substring_grams"]
					},
					"substring_search": {
						"type": "custom",
						"tokenizer": "standard",
						"filter": ["lowercase", "asciifolding", "english_stop_without_negation"]
					}
				}
			}
		},
		"mappings": {
			"properties": {
				"id": {
					"type": "keyword"
				},
				"text": {
					"type": "text",
					"analyzer": "english_preserving_negation",
					"fields": {
						"ngram": {
							"type": "text",
							"analyzer": "substring_index",
							"search_analyzer": "substring_search"
						}
					}
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
