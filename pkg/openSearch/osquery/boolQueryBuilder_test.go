// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package osquery

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aquasecurity/esquery"
	"github.com/greenbone/opensight-golang-libraries/pkg/openSearch/ostesting"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/filter"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/sorting"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// built in openSearch field for sorting
	docField = "_doc"
	// default maximum page size for OpenSearch queries
	maxPageSize = 10_000
)

// testOSClient eases testing against OpenSearch
type testOSClient struct {
	client *opensearchapi.Client
}

func newTestOSClient(client *opensearchapi.Client) (testOSClient, error) {
	return testOSClient{client: client}, nil
}

// ListDocuments sends a `/_search` request to OpenSearch and returns the parsed documents.
func (c testOSClient) ListDocuments(index string, requestBody []byte) (docs []ostesting.TestType, totalResults uint64, err error) {
	// send request
	resp, err := c.client.Search(
		context.Background(),
		&opensearchapi.SearchReq{
			Indices: []string{index},
			Body:    bytes.NewReader(requestBody),
		},
	)
	if err != nil {
		return nil, 0, fmt.Errorf("request failed: request: %v - error: %w", string(requestBody), err)
	}

	// parse response
	hits := resp.Hits.Hits
	docs = make([]ostesting.TestType, len(hits))
	for i, hit := range hits {
		err := json.Unmarshal(hit.Source, &docs[i])
		if err != nil {
			return nil, 0, fmt.Errorf("could not unmarshal document source: %s - error: %w", string(hit.Source), err)
		}
	}
	totalResults = uint64(resp.Hits.Total.Value)

	return docs, totalResults, nil
}

func singleFilter(f filter.RequestField) *filter.Request {
	return &filter.Request{
		Fields: []filter.RequestField{
			f,
		},
		// `Operator` not required/applicable for single filter
	}
}

func TestBoolQueryBuilder_AddFilterRequest(t *testing.T) {
	tester := ostesting.NewTester(t)

	querySettings := QuerySettings{
		FilterFieldMapping: map[string]string{
			"textField":                  "text",
			"keywordField":               "keyword",
			"textAndKeywordKeywordField": "textAndKeyword.keyword",
			"textAndKeywordTextField":    "textAndKeyword",
			"integerField":               "integer",
			"floatField":                 "float",
			"booleanField":               "boolean",
			"dateTimeStrField":           "dateTimeStr",
			"dateTimeField":              "dateTime",
			"keywordOmitEmptyField":      "keywordOmitEmpty",
		},
	}

	// note: for ease of testing, doc values need to be in strictly ascending order
	doc0 := ostesting.TestType{
		ID:          "0",
		DateTimeStr: "1000-01-01T10:00:00.000+02:00", // can't be empty
	}
	doc1 := ostesting.TestType{
		ID:               "1",
		Text:             "1test document number one",
		Keyword:          "keyword1",
		TextAndKeyword:   "1 one, some text which is also a keyword",
		Integer:          1,
		Float:            1.1,
		Boolean:          true,
		DateTimeStr:      "2024-01-23T10:00:00.000Z",
		DateTime:         time.Date(2024, 1, 23, 10, 0, 0, 0, time.UTC),
		KeywordOmitEmpty: "1not empty",
	}
	doc2 := ostesting.TestType{
		ID:               "2",
		Text:             "2test document number two",
		Keyword:          "otherkeyword2",
		TextAndKeyword:   "2 two, some text, also indexed as a keyword",
		Integer:          2,
		Float:            2.2,
		Boolean:          false,
		DateTimeStr:      "2024-02-23T10:00:00.000Z",
		DateTime:         time.Date(2024, 2, 23, 10, 0, 0, 0, time.UTC),
		KeywordOmitEmpty: "2notEmpty",
	}

	allDocs := []ostesting.TestType{doc0, doc1, doc2}

	type testCase struct {
		filterRequest *filter.Request
		wantDocuments []ostesting.TestType
		wantErr       bool // note: considers error during filter translation, after sending to OpenSearch we expect always success in this test
	}

	// general cases irrespective of operator
	tests := map[string]testCase{
		"success with nil filter request": {
			filterRequest: nil,
			wantDocuments: allDocs,
			wantErr:       false,
		},
		"success with empty filter request": {
			filterRequest: &filter.Request{},
			wantDocuments: allDocs,
			wantErr:       false,
		},
		"success without logic operator and single filter request": {
			filterRequest: &filter.Request{
				Fields: []filter.RequestField{
					{
						Name:     "textField",
						Operator: filter.CompareOperatorIsEqualTo,
						Value:    "start",
					},
				},
				// `Operator` not required/applicable for single filter
			},
		},
		"combine filters with OR": {
			filterRequest: &filter.Request{
				Fields: []filter.RequestField{
					{
						Name:     "keywordField",
						Operator: filter.CompareOperatorIsEqualTo,
						Value:    doc1.Keyword,
					},
					{
						Name:     "keywordField",
						Operator: filter.CompareOperatorIsEqualTo,
						Value:    doc2.Keyword,
					},
				},
				Operator: filter.LogicOperatorOr,
			},
			wantDocuments: []ostesting.TestType{doc1, doc2},
		},
		"combine filters with AND": {
			filterRequest: &filter.Request{
				Fields: []filter.RequestField{
					{
						Name:     "integerField",
						Operator: filter.CompareOperatorIsGreaterThan,
						Value:    doc0.Integer,
					},
					{
						Name:     "integerField",
						Operator: filter.CompareOperatorIsLessThan,
						Value:    doc2.Integer,
					},
				},
				Operator: filter.LogicOperatorAnd,
			},
			wantDocuments: []ostesting.TestType{doc1},
		},
		"fail with multiple filters and missing logic operator": {
			filterRequest: &filter.Request{
				Fields: []filter.RequestField{
					{
						Name:     "integerField",
						Operator: filter.CompareOperatorIsGreaterThan,
						Value:    doc0.Integer,
					},
					{
						Name:     "integerField",
						Operator: filter.CompareOperatorIsLessThan,
						Value:    doc2.Integer,
					},
				},
			},
			wantErr: true,
		},
		"fail on empty value": {
			filterRequest: singleFilter(filter.RequestField{
				Name:     "keywordField",
				Operator: filter.CompareOperatorIsEqualTo,
				Value:    nil,
			}),
			wantErr: true,
		},
		"fail on empty slice value": {
			filterRequest: singleFilter(filter.RequestField{
				Name:     "keywordField",
				Operator: filter.CompareOperatorIsEqualTo,
				Value:    []any{},
			}),
			wantErr: true,
		},
		"fail with invalid filter request (empty field name)": {
			filterRequest: &filter.Request{
				Fields: []filter.RequestField{
					{
						Name:     "", // name must not be empty
						Operator: filter.CompareOperatorBeginsWith,
						Value:    "arbitrary",
					},
				},
				Operator: filter.LogicOperatorAnd,
			},
			wantErr: true,
		},
	}
	addTest := func(name string, tc testCase) {
		t.Helper()
		if _, exists := tests[name]; exists {
			t.Fatalf("test case with name %q already exists", name)
		}
		tests[name] = tc
	}

	// empty single values are handled gracefully
	operators := map[filter.CompareOperator][]ostesting.TestType{ // operator to expected documents
		filter.CompareOperatorIsEqualTo:              {doc0},
		filter.CompareOperatorIsNotEqualTo:           {doc1, doc2},
		filter.CompareOperatorIsGreaterThan:          {doc1, doc2},
		filter.CompareOperatorIsGreaterThanOrEqualTo: allDocs,
		filter.CompareOperatorIsLessThan:             {},
		filter.CompareOperatorIsLessThanOrEqualTo:    {doc0},
		filter.CompareOperatorBeginsWith:             allDocs,
		filter.CompareOperatorDoesNotBeginWith:       {},
		filter.CompareOperatorContains:               allDocs,
		filter.CompareOperatorDoesNotContain:         {},
	}
	for operator, wantDocs := range operators {
		addTest(fmt.Sprintf("operator %v: empty single value", operator), testCase{
			filterRequest: singleFilter(filter.RequestField{
				Name:     "keywordField",
				Operator: operator,
				Value:    "",
			}),
			wantDocuments: wantDocs,
		})
	}

	// test cases for single value of different type applicable to most operators
	valueTypes := map[string]any{
		// textField not applicable
		"keywordField":               doc1.Keyword,
		"textAndKeywordKeywordField": doc1.TextAndKeyword,
		"integerField":               doc1.Integer,
		"floatField":                 doc1.Float,
		"booleanField":               doc1.Boolean,
		"dateTimeStrField":           doc1.DateTimeStr,
		"dateTimeField":              doc1.DateTime,
	}
	for fieldName, value := range valueTypes {
		addTest(fmt.Sprintf("operator IsEqualTo: single value (%v)", fieldName), testCase{
			filterRequest: singleFilter(filter.RequestField{
				Name:     fieldName,
				Operator: filter.CompareOperatorIsEqualTo,
				Value:    value,
			}),
			wantDocuments: []ostesting.TestType{doc1},
		})
		addTest(fmt.Sprintf("operator IsNotEqualTo: single value (%v)", fieldName), testCase{
			filterRequest: singleFilter(filter.RequestField{
				Name:     fieldName,
				Operator: filter.CompareOperatorIsNotEqualTo,
				Value:    value,
			}),
			wantDocuments: []ostesting.TestType{doc0, doc2},
		})
		if fieldName != "booleanField" {
			addTest(fmt.Sprintf("operator GreaterThan: single value (%v)", fieldName), testCase{
				filterRequest: singleFilter(filter.RequestField{
					Name:     fieldName,
					Operator: filter.CompareOperatorIsGreaterThan,
					Value:    value,
				}),
				wantDocuments: []ostesting.TestType{doc2},
			})
			addTest(fmt.Sprintf("operator GreaterOrEqualThan: single value (%v)", fieldName), testCase{
				filterRequest: singleFilter(filter.RequestField{
					Name:     fieldName,
					Operator: filter.CompareOperatorIsGreaterThanOrEqualTo,
					Value:    value,
				}),
				wantDocuments: []ostesting.TestType{doc1, doc2},
			})
			addTest(fmt.Sprintf("operator LessThan: single value (%v)", fieldName), testCase{
				filterRequest: singleFilter(filter.RequestField{
					Name:     fieldName,
					Operator: filter.CompareOperatorIsLessThan,
					Value:    value,
				}),
				wantDocuments: []ostesting.TestType{doc0},
			})
			addTest(fmt.Sprintf("operator LessThanOrEqualTo: single value (%v)", fieldName), testCase{
				filterRequest: singleFilter(filter.RequestField{
					Name:     fieldName,
					Operator: filter.CompareOperatorIsLessThanOrEqualTo,
					Value:    value,
				}),
				wantDocuments: []ostesting.TestType{doc0, doc1},
			})
			addTest(fmt.Sprintf("operator AfterDate: single value (%v)", fieldName), testCase{ // so far same as `GreaterThan`
				filterRequest: singleFilter(filter.RequestField{
					Name:     fieldName,
					Operator: filter.CompareOperatorAfterDate,
					Value:    value,
				}),
				wantDocuments: []ostesting.TestType{doc2},
			})
			addTest(fmt.Sprintf("operator BeforDate: single value (%v)", fieldName), testCase{ // so far same as `LessThan`
				filterRequest: singleFilter(filter.RequestField{
					Name:     fieldName,
					Operator: filter.CompareOperatorBeforeDate,
					Value:    value,
				}),
				wantDocuments: []ostesting.TestType{doc0},
			})
		}
	}

	// test cases for multiple values of different type applicable to most operators
	valueTypesMulti := map[string][]any{
		// textField not applicable
		"keywordField":               {doc1.Keyword, doc2.Keyword},
		"textAndKeywordKeywordField": {doc1.TextAndKeyword, doc2.TextAndKeyword},
		"integerField":               {doc1.Integer, doc2.Integer},
		"floatField":                 {doc1.Float, doc2.Float},
		// boolean tested below
		"dateTimeStrField": {doc1.DateTimeStr, doc2.DateTimeStr},
		"dateTimeField":    {doc1.DateTime, doc2.DateTime},
	}
	for fieldName, value := range valueTypesMulti {
		fieldTypeName := strings.TrimSuffix(fieldName, "Field")

		addTest(fmt.Sprintf("operator IsEqualTo: multiple values (%v)", fieldTypeName), testCase{
			filterRequest: singleFilter(filter.RequestField{
				Name:     fieldName,
				Operator: filter.CompareOperatorIsEqualTo,
				Value:    value,
			}),
			wantDocuments: []ostesting.TestType{doc1, doc2},
		})
		addTest(fmt.Sprintf("operator IsNotEqualTo: multiple values (%v)", fieldName), testCase{
			filterRequest: singleFilter(filter.RequestField{
				Name:     fieldName,
				Operator: filter.CompareOperatorIsNotEqualTo,
				Value:    value,
			}),
			wantDocuments: []ostesting.TestType{doc0},
		})
		if fieldName != "booleanField" {
			addTest(fmt.Sprintf("operator GreaterThan: multiple values (%v)", fieldName), testCase{
				filterRequest: singleFilter(filter.RequestField{
					Name:     fieldName,
					Operator: filter.CompareOperatorIsGreaterThan,
					Value:    value,
				}),
				wantDocuments: []ostesting.TestType{doc2},
			})
			addTest(fmt.Sprintf("operator GreaterOrEqualThan: multiple values (%v)", fieldName), testCase{
				filterRequest: singleFilter(filter.RequestField{
					Name:     fieldName,
					Operator: filter.CompareOperatorIsGreaterThanOrEqualTo,
					Value:    value,
				}),
				wantDocuments: []ostesting.TestType{doc1, doc2},
			})
			addTest(fmt.Sprintf("operator LessThan: multiple values (%v)", fieldName), testCase{
				filterRequest: singleFilter(filter.RequestField{
					Name:     fieldName,
					Operator: filter.CompareOperatorIsLessThan,
					Value:    value,
				}),
				wantDocuments: []ostesting.TestType{doc0, doc1},
			})
			addTest(fmt.Sprintf("operator LessThanOrEqualTo: multiple values (%v)", fieldName), testCase{
				filterRequest: singleFilter(filter.RequestField{
					Name:     fieldName,
					Operator: filter.CompareOperatorIsLessThanOrEqualTo,
					Value:    value,
				}),
				wantDocuments: []ostesting.TestType{doc0, doc1, doc2},
			})
			addTest(fmt.Sprintf("operator AfterDate: multiple values (%v)", fieldName), testCase{ // so far same as `GreaterThan`
				filterRequest: singleFilter(filter.RequestField{
					Name:     fieldName,
					Operator: filter.CompareOperatorAfterDate,
					Value:    value,
				}),
				wantDocuments: []ostesting.TestType{doc2},
			})
			addTest(fmt.Sprintf("operator BeforDate: multiple values (%v)", fieldName), testCase{ // so far same as `LessThan`
				filterRequest: singleFilter(filter.RequestField{
					Name:     fieldName,
					Operator: filter.CompareOperatorBeforeDate,
					Value:    value,
				}),
				wantDocuments: []ostesting.TestType{doc0, doc1},
			})
		}
	}
	addTest("operator IsEqualTo: multiple values (booleanField)", testCase{
		filterRequest: singleFilter(filter.RequestField{
			Name:     "booleanField",
			Operator: filter.CompareOperatorIsEqualTo,
			Value:    []any{true, false},
		}),
		wantDocuments: allDocs,
	})

	// string/keyword specific operators, single value
	keywordValueTypes := map[string]string{
		"keywordField":               doc1.Keyword,
		"textAndKeywordKeywordField": doc1.TextAndKeyword,
	}
	for fieldName, value := range keywordValueTypes {
		require.Greater(t, len(value), 1, "test requires value to be longer than 1 character for `contains` operator")
		firstPart := value[:len(value)/2]
		secondPart := value[len(value)/2:]

		addTest(fmt.Sprintf("operator Contains: single value (%v)", fieldName), testCase{
			filterRequest: singleFilter(filter.RequestField{
				Name:     fieldName,
				Operator: filter.CompareOperatorContains,
				Value:    secondPart,
			}),
			wantDocuments: []ostesting.TestType{doc1},
		})
		addTest(fmt.Sprintf("operator DoesNotContain: single value (%v)", fieldName), testCase{
			filterRequest: singleFilter(filter.RequestField{
				Name:     fieldName,
				Operator: filter.CompareOperatorDoesNotContain,
				Value:    secondPart,
			}),
			wantDocuments: []ostesting.TestType{doc0, doc2},
		})
		addTest(fmt.Sprintf("operator BeginsWith: single value (%v)", fieldName), testCase{
			filterRequest: singleFilter(filter.RequestField{
				Name:     fieldName,
				Operator: filter.CompareOperatorBeginsWith,
				Value:    firstPart,
			}),
			wantDocuments: []ostesting.TestType{doc1},
		})
		addTest(fmt.Sprintf("operator DoesNotBeginWith: single value (%v)", fieldName), testCase{
			filterRequest: singleFilter(filter.RequestField{
				Name:     fieldName,
				Operator: filter.CompareOperatorDoesNotBeginWith,
				Value:    firstPart,
			}),
			wantDocuments: []ostesting.TestType{doc0, doc2},
		})
	}

	// string/keyword specific operators, invalid values (only string supported)
	valueTypes = map[string]any{
		"integerField": doc1.Integer,
		"floatField":   doc1.Float,
		"booleanField": doc1.Boolean,
		// dateTimeStrField will only be caught during query execution
		"dateTimeField": doc1.DateTime,
	}
	for fieldName, value := range valueTypes {
		addTest(fmt.Sprintf("operator BeginsWith: invalid value (%v)", fieldName), testCase{
			filterRequest: singleFilter(filter.RequestField{
				Name:     fieldName,
				Operator: filter.CompareOperatorBeginsWith,
				Value:    value,
			}),
			wantErr: true,
		})
		addTest(fmt.Sprintf("operator DoesNotBeginWith: invalid value (%v)", fieldName), testCase{
			filterRequest: singleFilter(filter.RequestField{
				Name:     fieldName,
				Operator: filter.CompareOperatorDoesNotBeginWith,
				Value:    value,
			}),
			wantErr: true,
		})
		addTest(fmt.Sprintf("operator Contains: invalid value (%v)", fieldName), testCase{
			filterRequest: singleFilter(filter.RequestField{
				Name:     fieldName,
				Operator: filter.CompareOperatorContains,
				Value:    value,
			}),
			wantErr: true,
		})
		addTest(fmt.Sprintf("operator DoesNotContain: invalid value (%v)", fieldName), testCase{
			filterRequest: singleFilter(filter.RequestField{
				Name:     fieldName,
				Operator: filter.CompareOperatorDoesNotContain,
				Value:    value,
			}),
			wantErr: true,
		})
	}

	// string/keyword specific operators, multiple values
	keywordValueTypesMulti := map[string][]string{
		"keywordField":               {doc1.Keyword, doc2.Keyword},
		"textAndKeywordKeywordField": {doc1.TextAndKeyword, doc2.TextAndKeyword},
	}
	for fieldName, values := range keywordValueTypesMulti {
		var firstParts, secondParts []any
		for _, value := range values {
			require.Greater(t, len(value), 1, "test requires value to be longer than 1 character for `contains` operator")
			firstParts = append(firstParts, value[:len(value)/2])
			secondParts = append(secondParts, value[len(value)/2:])
		}

		addTest(fmt.Sprintf("operator Contains: multiple values (%v)", fieldName), testCase{
			filterRequest: singleFilter(filter.RequestField{
				Name:     fieldName,
				Operator: filter.CompareOperatorContains,
				Value:    secondParts,
			}),
			wantDocuments: []ostesting.TestType{doc1, doc2},
		})
		addTest(fmt.Sprintf("operator DoesNotContain: multiple values (%v)", fieldName), testCase{
			filterRequest: singleFilter(filter.RequestField{
				Name:     fieldName,
				Operator: filter.CompareOperatorDoesNotContain,
				Value:    secondParts,
			}),
			wantDocuments: []ostesting.TestType{doc0},
		})
		addTest(fmt.Sprintf("operator BeginsWith: multiple values (%v)", fieldName), testCase{
			filterRequest: singleFilter(filter.RequestField{
				Name:     fieldName,
				Operator: filter.CompareOperatorBeginsWith,
				Value:    firstParts,
			}),
			wantDocuments: []ostesting.TestType{doc1, doc2},
		})
		addTest(fmt.Sprintf("operator DoesNotBeginWith: multiple values (%v)", fieldName), testCase{
			filterRequest: singleFilter(filter.RequestField{
				Name:     fieldName,
				Operator: filter.CompareOperatorDoesNotBeginWith,
				Value:    firstParts,
			}),
			wantDocuments: []ostesting.TestType{doc0},
		})
	}

	// TextContains operator
	addTest("operator TextContains: single value", testCase{
		filterRequest: singleFilter(filter.RequestField{
			Name:     "textField",
			Operator: filter.CompareOperatorTextContains,
			Value:    "document one", // all terms need to be present, `one` not present in `doc2.Text``
		}),
		wantDocuments: []ostesting.TestType{doc1},
	})
	addTest("operator TextContains: multiple values", testCase{
		filterRequest: singleFilter(filter.RequestField{
			Name:     "textField",
			Operator: filter.CompareOperatorTextContains,
			Value:    []any{"document one", "two document"},
		}),
		wantDocuments: []ostesting.TestType{doc1, doc2},
	})

	// Exists operator
	addTest("operator Exists", testCase{
		filterRequest: singleFilter(filter.RequestField{
			Name:     "keywordOmitEmptyField",
			Operator: filter.CompareOperatorExists,
		}),
		wantDocuments: []ostesting.TestType{doc1, doc2},
	})
	addTest("operator Exists - empty but set field exist", testCase{
		filterRequest: singleFilter(filter.RequestField{
			Name:     "keywordField",
			Operator: filter.CompareOperatorExists,
		}),
		wantDocuments: []ostesting.TestType{doc0, doc1, doc2},
	})

	// MustNotExists operator
	addTest("operator must not Exists", testCase{
		filterRequest: singleFilter(filter.RequestField{
			Name:     "keywordOmitEmptyField",
			Operator: filter.CompareOperatorDoesNotExist,
		}),
		wantDocuments: []ostesting.TestType{doc0},
	})

	// BetweenDates operator
	addTest("operator BetweenDates (date time string)", testCase{
		filterRequest: singleFilter(filter.RequestField{
			Name:     "dateTimeStrField",
			Operator: filter.CompareOperatorBetweenDates,
			Value:    []string{"2024-01-01T00:00:00.000Z", doc1.DateTimeStr}, // borders are inclusive
		}),
		wantDocuments: []ostesting.TestType{doc1},
	})
	addTest("operator BetweenDates (date time)", testCase{
		filterRequest: singleFilter(filter.RequestField{
			Name:     "dateTimeField",
			Operator: filter.CompareOperatorBetweenDates,
			Value:    []time.Time{doc1.DateTime, doc1.DateTime.Add(5 * time.Hour)}, // borders are inclusive
		}),
		wantDocuments: []ostesting.TestType{doc1},
	})
	addTest("operator BetweenDates (mixed)", testCase{
		filterRequest: singleFilter(filter.RequestField{
			Name:     "dateTimeField",
			Operator: filter.CompareOperatorBetweenDates,
			Value:    []any{doc1.DateTime, doc1.DateTimeStr}, // borders are inclusive
		}),
		wantDocuments: []ostesting.TestType{doc1},
	})
	addTest("operator BetweenDates (invalid value type in slice)", testCase{
		filterRequest: singleFilter(filter.RequestField{
			Name:     "dateTimeField",
			Operator: filter.CompareOperatorBetweenDates,
			Value:    []any{123, 456}, // invalid type
		}),
		wantErr: true,
	})
	addTest("operator BetweenDates (not a slice)", testCase{
		filterRequest: singleFilter(filter.RequestField{
			Name:     "dateTimeField",
			Operator: filter.CompareOperatorBetweenDates,
			Value:    doc1.DateTime, // must be a slice
		}),
		wantErr: true,
	})
	addTest("operator BetweenDates (wrong length of values)", testCase{
		filterRequest: singleFilter(filter.RequestField{
			Name:     "dateTimeField",
			Operator: filter.CompareOperatorBetweenDates,
			Value:    []time.Time{doc1.DateTime}, // only one value
		}),
		wantErr: true,
	})
	addTest("operator BetweenDates (invalid date format)", testCase{
		filterRequest: singleFilter(filter.RequestField{
			Name:     "dateTimeField",
			Operator: filter.CompareOperatorBetweenDates,
			Value:    []string{"not-a-date", doc1.DateTimeStr}, // invalid date format
		}),
		wantErr: true,
	})

	// index setup needed only once, as test cases are only reading data
	_, alias := tester.NewTestTypeIndexAlias(t, "arbitrary")
	tester.CreateDocuments(t, alias, ostesting.ToAnySlice(allDocs), []string{"id0", "id1", "id2"})
	client, err := newTestOSClient(tester.OSClient())
	require.NoError(t, err)

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			queryBuilder := NewBoolQueryBuilder(&querySettings)
			err := queryBuilder.AddFilterRequest(tt.filterRequest)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)

				searchRequest := esquery.Search().
					Size(maxPageSize).
					Sort(docField, esquery.Order(sorting.DirectionAscending))
				searchRequest.Query(queryBuilder.Build())
				requestBody, err := searchRequest.MarshalJSON()
				require.NoError(t, err, "could not marshal search request")

				t.Logf("sending search request to OpenSearch: %s", string(requestBody))
				gotDocuments, totalResults, err := client.ListDocuments(alias, requestBody)
				require.NoError(t, err, "listing documents failed")
				require.Len(t, gotDocuments, int(totalResults),
					"test requires results to be returned on a single page") // catch paging misconfiguration

				assert.ElementsMatch(t, tt.wantDocuments, gotDocuments)
			}
		})
	}
}

func TestBoolQueryBuilder_AddTermFilter(t *testing.T) {
	tester := ostesting.NewTester(t)

	querySettings := QuerySettings{
		FilterFieldMapping: map[string]string{
			"integerField": "integer",
		},
	}

	// note: for ease of testing, doc values need to be in strictly ascending order
	doc0 := ostesting.TestType{
		ID:          "0",
		Integer:     0,
		DateTimeStr: "1000-01-01T00:00:00.000Z", // can't be empty
	}
	doc1 := ostesting.TestType{
		ID:          "1",
		Integer:     1,
		DateTimeStr: "1000-01-01T01:00:00.000Z",
	}
	doc2 := ostesting.TestType{
		ID:          "2",
		Integer:     2,
		DateTimeStr: "1000-01-01T02:00:00.000Z",
	}
	doc3 := ostesting.TestType{
		ID:          "3",
		Integer:     3,
		DateTimeStr: "1000-01-01T03:00:00.000Z",
	}
	doc4 := ostesting.TestType{
		ID:          "4",
		Integer:     4,
		DateTimeStr: "1000-01-01T04:00:00.000Z",
	}

	allDocs := []ostesting.TestType{doc0, doc1, doc2, doc3, doc4}

	// index setup needed only once, as test cases are only reading data
	_, alias := tester.NewTestTypeIndexAlias(t, "arbitrary")
	tester.CreateDocuments(t, alias, ostesting.ToAnySlice(allDocs), []string{
		"id0", "id1", "id2", "id3", "id4",
	})
	client, err := newTestOSClient(tester.OSClient())
	require.NoError(t, err)

	filterRequest := singleFilter(filter.RequestField{ // will filter out doc0
		Name:     "integerField",
		Operator: filter.CompareOperatorIsNotEqualTo,
		Value:    doc0.Integer,
	})

	type termFilter struct {
		fieldName string
		value     any
	}
	type termsFilters struct {
		fieldName string
		values    []any
	}

	tests := map[string]struct {
		termFilters   []termFilter
		termsFilters  []termsFilters
		termFirst     bool // if true, add term(s) filters before adding filter request
		wantDocuments []ostesting.TestType
		wantErr       bool // note: considers error during filter translation, after sending to OpenSearch we expect always success in this test
	}{
		"single term filter works": {
			termFilters: []termFilter{
				{
					fieldName: "integerField",
					value:     doc1.Integer,
				},
			},
			wantDocuments: []ostesting.TestType{doc1},
		},
		"single term filter works (no result)": {
			termFilters: []termFilter{
				{
					fieldName: "integerField",
					value:     doc0.Integer,
				},
			},
			wantDocuments: []ostesting.TestType{}, // doc0 alread filtered out by filterRequest
		},
		"multiple term filter work (no result)": {
			termFilters: []termFilter{
				{
					fieldName: "integerField",
					value:     doc1.Integer,
				},
				{
					fieldName: "integerField",
					value:     doc2.Integer,
				},
			},
			wantDocuments: []ostesting.TestType{}, // no overlap in term filters
		},
		"mutiple terms filter work": {
			termsFilters: []termsFilters{
				{
					fieldName: "integerField",
					values:    []any{doc0.Integer, doc1.Integer, doc2.Integer, doc3.Integer},
				},
				{
					fieldName: "integerField",
					values:    []any{doc0.Integer, doc2.Integer, doc3.Integer, doc4.Integer},
				},
			},
			wantDocuments: []ostesting.TestType{doc2, doc3}, // doc0 already filtered out by filterRequest
		},
		"term and terms filter can be combined": {
			termFilters: []termFilter{
				{
					fieldName: "integerField",
					value:     doc1.Integer,
				},
			},
			termsFilters: []termsFilters{
				{
					fieldName: "integerField",
					values:    []any{doc0.Integer, doc1.Integer, doc2.Integer, doc3.Integer},
				},
			},
			wantDocuments: []ostesting.TestType{doc1},
		},
		"term and terms filter can be combined (added before filter request)": {
			termFilters: []termFilter{
				{
					fieldName: "integerField",
					value:     doc1.Integer,
				},
			},
			termsFilters: []termsFilters{
				{
					fieldName: "integerField",
					values:    []any{doc0.Integer, doc1.Integer, doc2.Integer, doc3.Integer},
				},
			},
			termFirst:     true,
			wantDocuments: []ostesting.TestType{doc1},
		},
		"term and terms filter can be combined (no results)": {
			termFilters: []termFilter{
				{
					fieldName: "integerField",
					value:     doc1.Integer,
				},
			},
			termsFilters: []termsFilters{
				{
					fieldName: "integerField",
					values:    []any{doc0.Integer, doc2.Integer, doc3.Integer},
				},
			},
			wantDocuments: []ostesting.TestType{}, // no overlap in term and terms filter
		},
		"term filter with invalid field name fails": {
			termFilters: []termFilter{
				{
					fieldName: "nonExistingField",
					value:     doc1.Integer,
				},
			},
			wantErr: true,
		},
		"terms filter with invalid field name fails": {
			termsFilters: []termsFilters{
				{
					fieldName: "nonExistingField",
					values:    []any{doc1.Integer, doc2.Integer},
				},
			},
			wantErr: true,
		},
		"terms filter with empty value list fails": {
			termsFilters: []termsFilters{
				{
					fieldName: "integerField",
					values:    []any{},
				},
			},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			queryBuilder := NewBoolQueryBuilder(&querySettings)

			if !tt.termFirst { // to test that order does not matter
				// add always same filter
				err := queryBuilder.AddFilterRequest(filterRequest)
				require.NoError(t, err, "could not add filter request")
			}

			for _, termFilter := range tt.termFilters {
				err := queryBuilder.AddTermFilter(termFilter.fieldName, termFilter.value)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				require.NoError(t, err, "could not add term filter")
			}
			for _, termsFilter := range tt.termsFilters {
				err := queryBuilder.AddTermsFilter(termsFilter.fieldName, termsFilter.values...)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				require.NoError(t, err, "could not add terms filter")
			}

			if tt.termFirst {
				// add always same filter
				err := queryBuilder.AddFilterRequest(filterRequest)
				require.NoError(t, err, "could not add filter request")
			}

			searchRequest := esquery.Search().
				Size(maxPageSize).
				Sort(docField, esquery.Order(sorting.DirectionAscending))
			searchRequest.Query(queryBuilder.Build())
			requestBody, err := searchRequest.MarshalJSON()
			require.NoError(t, err, "could not marshal search request")

			t.Logf("sending search request to OpenSearch: %s", string(requestBody))
			gotDocuments, totalResults, err := client.ListDocuments(alias, requestBody)
			require.NoError(t, err, "listing documents failed")
			require.Len(t, gotDocuments, int(totalResults),
				"test requires results to be returned on a single page") // catch paging misconfiguration

			assert.ElementsMatch(t, tt.wantDocuments, gotDocuments)
		})
	}
}
