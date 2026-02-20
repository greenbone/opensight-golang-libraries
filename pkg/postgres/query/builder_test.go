// SPDX-FileCopyrightText: 2024 Greenbone AG <https://greenbone.net>
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package query

import (
	"database/sql"
	"embed"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/greenbone/opensight-golang-libraries/internal/pgtesting"
	"github.com/greenbone/opensight-golang-libraries/pkg/query"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/filter"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/paging"
	"github.com/greenbone/opensight-golang-libraries/pkg/query/sorting"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const unfilteredListTestypesQuery = `SELECT * FROM test_table`

//go:embed test_migrations
var migrationsFS embed.FS

// directory within [migrationsFS] where migration files are located
const migrationDir = "test_migrations"

type TestDoc struct {
	ID       int       `db:"id"`
	String   string    `db:"string"`
	Integer  int       `db:"integer"`
	Float    float32   `db:"float"`
	Boolean  bool      `db:"boolean"`
	DateTime time.Time `db:"date_time"`
}

var fieldMapping = map[string]string{
	"idField":       "id",
	"stringField":   "string",
	"integerField":  "integer",
	"floatField":    "float",
	"booleanField":  "boolean",
	"dateTimeField": "date_time",
}

var sortingTieBreakerColumn = "id"

type TestRepository struct {
	db *sql.DB
}

func NewTestRepository(db *sql.DB) *TestRepository {
	return &TestRepository{db: db}
}

func (r *TestRepository) CreateTestDoc(testType *TestDoc) error {
	_, err := r.db.Exec(
		`INSERT INTO test_table (id, string, integer, float, boolean, date_time)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		testType.ID,
		testType.String,
		testType.Integer,
		testType.Float,
		testType.Boolean,
		testType.DateTime,
	)
	return err
}

func (r *TestRepository) ListTestDocs(query string, args []any) ([]TestDoc, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var testTypes []TestDoc
	for rows.Next() {
		var tt TestDoc
		if err := rows.Scan(&tt.ID, &tt.String, &tt.Integer, &tt.Float, &tt.Boolean, &tt.DateTime); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		testTypes = append(testTypes, tt)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	return testTypes, nil
}

func singleFilter(f filter.RequestField) query.ResultSelector {
	return query.ResultSelector{
		Filter: &filter.Request{
			Fields: []filter.RequestField{
				f,
			},
			// `Operator` not required/applicable for single filter
		},
	}
}

func Test_PostgresQueryBuilder_Build(t *testing.T) {
	doc0 := TestDoc{
		ID: 0,
	}
	doc1 := TestDoc{
		ID:       1,
		String:   "1Test String One",
		Integer:  1,
		Float:    1.1,
		Boolean:  true,
		DateTime: time.Date(2024, 1, 23, 10, 0, 0, 0, time.UTC),
	}
	doc2 := TestDoc{
		ID:       2,
		String:   "2Test String Two",
		Integer:  2,
		Float:    2.2,
		Boolean:  false,
		DateTime: time.Date(2024, 2, 23, 10, 0, 0, 0, time.UTC),
	}

	allDocs := []TestDoc{doc0, doc1, doc2} // all available documents in the default sort order

	type testCase struct {
		resultSelector query.ResultSelector
		wantDocuments  []TestDoc
		wantErr        bool
	}

	tests := map[string]testCase{
		"no filters, no sorting (tie breaker applied), no paging": {
			resultSelector: query.ResultSelector{},
			wantDocuments:  allDocs,
		},
		"pagination: first page": {
			resultSelector: query.ResultSelector{
				Paging: &paging.Request{
					PageSize:  2,
					PageIndex: 0,
				},
				Sorting: &sorting.Request{
					SortColumn:    "idField",
					SortDirection: sorting.DirectionAscending,
				},
			},
			wantDocuments: []TestDoc{doc0, doc1},
		},
		"pagination: second page": {
			resultSelector: query.ResultSelector{
				Paging: &paging.Request{
					PageSize:  2,
					PageIndex: 1,
				},
				Sorting: &sorting.Request{
					SortColumn:    "idField",
					SortDirection: sorting.DirectionAscending,
				},
			},
			wantDocuments: []TestDoc{doc2},
		},
		"pagination: invalid page size": {
			resultSelector: query.ResultSelector{
				Paging: &paging.Request{
					PageSize:  -1,
					PageIndex: 0,
				},
			},
			wantErr: true,
		},
		"pagination: invalid page index": {
			resultSelector: query.ResultSelector{
				Paging: &paging.Request{
					PageSize:  0,
					PageIndex: -1,
				},
			},
			wantErr: true,
		},
		"sorting: tiebreaker is applied": {
			resultSelector: query.ResultSelector{
				Sorting: &sorting.Request{
					SortColumn:    "booleanField",
					SortDirection: sorting.DirectionAscending,
				},
			},
			wantDocuments: []TestDoc{doc0, doc2, doc1}, // doc0 and doc2 have same boolean value (false), so tie breaker (id) decides order
		},
		"sorting: direction is applied": {
			resultSelector: query.ResultSelector{
				Sorting: &sorting.Request{
					SortColumn:    "integerField",
					SortDirection: sorting.DirectionDescending,
				},
			},
			wantDocuments: []TestDoc{doc2, doc1, doc0},
		},
		"sorting: fail on invalid sort column": {
			resultSelector: query.ResultSelector{
				Sorting: &sorting.Request{
					SortColumn:    "nonExistingField",
					SortDirection: sorting.DirectionAscending,
				},
			},
			wantErr: true,
		},
		"filter, sorting and paging combined": {
			resultSelector: query.ResultSelector{
				Filter: &filter.Request{
					Fields: []filter.RequestField{
						{
							Name:     "booleanField",
							Operator: filter.CompareOperatorIsEqualTo,
							Value:    false,
						},
					},
				},
				Sorting: &sorting.Request{
					SortColumn:    "idField",
					SortDirection: sorting.DirectionDescending,
				},
				Paging: &paging.Request{
					PageSize:  1,
					PageIndex: 0,
				},
			},
			wantDocuments: []TestDoc{doc2}, // only doc2 has booleanField = false; sorted by id desc; first page with size 1
		},
		// rest of tests focus on filtering only
		"success without logic operator and single filter": {
			resultSelector: query.ResultSelector{
				Filter: &filter.Request{
					Fields: []filter.RequestField{
						{
							Name:     "idField",
							Operator: filter.CompareOperatorIsEqualTo,
							Value:    doc1.ID,
						},
					},
				},
			},
			wantDocuments: []TestDoc{doc1},
		},
		"accept filter value of type []any": {
			resultSelector: query.ResultSelector{
				Filter: &filter.Request{
					Fields: []filter.RequestField{
						{
							Name:     "idField",
							Operator: filter.CompareOperatorIsEqualTo,
							Value:    []any{doc1.ID},
						},
					},
				},
			},
			wantDocuments: []TestDoc{doc1},
		},
		// rest of tests focus on filtering only
		"accept filter value of specific slice type ([]int)": {
			resultSelector: query.ResultSelector{
				Filter: &filter.Request{
					Fields: []filter.RequestField{
						{
							Name:     "idField",
							Operator: filter.CompareOperatorIsEqualTo,
							Value:    []int{doc1.ID},
						},
					},
				},
			},
			wantDocuments: []TestDoc{doc1},
		},
		"combine filters with OR": {
			resultSelector: query.ResultSelector{
				Filter: &filter.Request{
					Fields: []filter.RequestField{
						{
							Name:     "integerField",
							Operator: filter.CompareOperatorIsEqualTo,
							Value:    doc1.Integer,
						},
						{
							Name:     "integerField",
							Operator: filter.CompareOperatorIsEqualTo,
							Value:    doc2.Integer,
						},
					},
					Operator: filter.LogicOperatorOr,
				},
			},
			wantDocuments: []TestDoc{doc1, doc2},
		},
		"combine filters with AND": {
			resultSelector: query.ResultSelector{
				Filter: &filter.Request{
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
			},
			wantDocuments: []TestDoc{doc1},
		},
		"fail with multiple filters and missing logic operator": {
			resultSelector: query.ResultSelector{
				Filter: &filter.Request{
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
					// Operator missing

				},
			},
			wantErr: true,
		},
		"fail on nil filter value": {
			resultSelector: query.ResultSelector{
				Filter: &filter.Request{
					Fields: []filter.RequestField{
						{
							Name:     "idField",
							Operator: filter.CompareOperatorIsEqualTo,
							Value:    nil,
						},
					},
				},
			},
			wantErr: true,
		},
		"fail on empty slice filter value:": {
			resultSelector: query.ResultSelector{
				Filter: &filter.Request{
					Fields: []filter.RequestField{
						{
							Name:     "idField",
							Operator: filter.CompareOperatorIsEqualTo,
							Value:    []any{},
						},
					},
				},
			},
			wantErr: true,
		},
		"fail with invalid filter (empty field name)": {
			resultSelector: query.ResultSelector{
				Filter: &filter.Request{
					Fields: []filter.RequestField{
						{
							Name:     "",
							Operator: filter.CompareOperatorIsEqualTo,
							Value:    "arbitrary",
						},
					},
				},
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
	operators := map[filter.CompareOperator][]TestDoc{ // operator to expected documents
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
			resultSelector: singleFilter(filter.RequestField{
				Name:     "stringField",
				Operator: operator,
				Value:    "",
			}),
			wantDocuments: wantDocs,
		})
	}

	// test cases for single value of different type applicable to most operators
	valueTypes := map[string]any{
		"stringField":   doc1.String,
		"integerField":  doc1.Integer,
		"floatField":    doc1.Float,
		"booleanField":  doc1.Boolean,
		"dateTimeField": doc1.DateTime,
	}
	for fieldName, value := range valueTypes {
		addTest(fmt.Sprintf("operator IsEqualTo: single value (%v)", fieldName), testCase{
			resultSelector: singleFilter(filter.RequestField{
				Name:     fieldName,
				Operator: filter.CompareOperatorIsEqualTo,
				Value:    value,
			}),
			wantDocuments: []TestDoc{doc1},
		})
		addTest(fmt.Sprintf("operator IsNotEqualTo: single value (%v)", fieldName), testCase{
			resultSelector: singleFilter(filter.RequestField{
				Name:     fieldName,
				Operator: filter.CompareOperatorIsNotEqualTo,
				Value:    value,
			}),
			wantDocuments: []TestDoc{doc0, doc2},
		})
		if fieldName != "booleanField" {
			addTest(fmt.Sprintf("operator GreaterThan: single value (%v)", fieldName), testCase{
				resultSelector: singleFilter(filter.RequestField{
					Name:     fieldName,
					Operator: filter.CompareOperatorIsGreaterThan,
					Value:    value,
				}),
				wantDocuments: []TestDoc{doc2},
			})
			addTest(fmt.Sprintf("operator GreaterOrEqualThan: single value (%v)", fieldName), testCase{
				resultSelector: singleFilter(filter.RequestField{
					Name:     fieldName,
					Operator: filter.CompareOperatorIsGreaterThanOrEqualTo,
					Value:    value,
				}),
				wantDocuments: []TestDoc{doc1, doc2},
			})
			addTest(fmt.Sprintf("operator LessThan: single value (%v)", fieldName), testCase{
				resultSelector: singleFilter(filter.RequestField{
					Name:     fieldName,
					Operator: filter.CompareOperatorIsLessThan,
					Value:    value,
				}),
				wantDocuments: []TestDoc{doc0},
			})
			addTest(fmt.Sprintf("operator LessThanOrEqualTo: single value (%v)", fieldName), testCase{
				resultSelector: singleFilter(filter.RequestField{
					Name:     fieldName,
					Operator: filter.CompareOperatorIsLessThanOrEqualTo,
					Value:    value,
				}),
				wantDocuments: []TestDoc{doc0, doc1},
			})
		}
	}

	// test cases for multiple values of different type applicable to most operators
	valueTypesMulti := map[string][]any{
		"stringField":  {doc1.String, doc2.String},
		"integerField": {doc1.Integer, doc2.Integer},
		"floatField":   {doc1.Float, doc2.Float},
		// boolean tested below
		"dateTimeField": {doc1.DateTime, doc2.DateTime},
	}
	for fieldName, value := range valueTypesMulti {
		fieldTypeName := strings.TrimSuffix(fieldName, "Field")

		addTest(fmt.Sprintf("operator IsEqualTo: multiple values (%v)", fieldTypeName), testCase{
			resultSelector: singleFilter(filter.RequestField{
				Name:     fieldName,
				Operator: filter.CompareOperatorIsEqualTo,
				Value:    value,
			}),
			wantDocuments: []TestDoc{doc1, doc2},
		})
		addTest(fmt.Sprintf("operator IsNotEqualTo: multiple values (%v)", fieldName), testCase{
			resultSelector: singleFilter(filter.RequestField{
				Name:     fieldName,
				Operator: filter.CompareOperatorIsNotEqualTo,
				Value:    value,
			}),
			wantDocuments: []TestDoc{doc0},
		})
		if fieldName != "booleanField" {
			addTest(fmt.Sprintf("operator GreaterThan: multiple values (%v)", fieldName), testCase{
				resultSelector: singleFilter(filter.RequestField{
					Name:     fieldName,
					Operator: filter.CompareOperatorIsGreaterThan,
					Value:    value,
				}),
				wantDocuments: []TestDoc{doc2},
			})
			addTest(fmt.Sprintf("operator GreaterOrEqualThan: multiple values (%v)", fieldName), testCase{
				resultSelector: singleFilter(filter.RequestField{
					Name:     fieldName,
					Operator: filter.CompareOperatorIsGreaterThanOrEqualTo,
					Value:    value,
				}),
				wantDocuments: []TestDoc{doc1, doc2},
			})
			addTest(fmt.Sprintf("operator LessThan: multiple values (%v)", fieldName), testCase{
				resultSelector: singleFilter(filter.RequestField{
					Name:     fieldName,
					Operator: filter.CompareOperatorIsLessThan,
					Value:    value,
				}),
				wantDocuments: []TestDoc{doc0, doc1},
			})
			addTest(fmt.Sprintf("operator LessThanOrEqualTo: multiple values (%v)", fieldName), testCase{
				resultSelector: singleFilter(filter.RequestField{
					Name:     fieldName,
					Operator: filter.CompareOperatorIsLessThanOrEqualTo,
					Value:    value,
				}),
				wantDocuments: []TestDoc{doc0, doc1, doc2},
			})
		}
	}
	addTest("operator IsEqualTo: multiple values (booleanField)", testCase{
		resultSelector: singleFilter(filter.RequestField{
			Name:     "booleanField",
			Operator: filter.CompareOperatorIsEqualTo,
			Value:    []any{true, false},
		}),
		wantDocuments: allDocs,
	})

	// string specific operators, single value
	value := doc1.String
	require.Greater(t, len(value), 1, "test requires value to be longer than 1 character for `contains` operator")
	firstPart := value[:len(value)/2]
	require.NotEqual(t, firstPart, strings.ToLower(firstPart),
		"test requires mixed case, to verify case insensitivity")
	firstPart = strings.ToLower(firstPart)
	secondPart := value[len(value)/2:]
	require.NotEqual(t, secondPart, strings.ToLower(secondPart),
		"test requires mixed case, to verify case insensitivity")
	secondPart = strings.ToLower(secondPart)

	addTest("operator Contains: single value", testCase{
		resultSelector: singleFilter(filter.RequestField{
			Name:     "stringField",
			Operator: filter.CompareOperatorContains,
			Value:    secondPart,
		}),
		wantDocuments: []TestDoc{doc1},
	})
	addTest("operator DoesNotContain: single value", testCase{
		resultSelector: singleFilter(filter.RequestField{
			Name:     "stringField",
			Operator: filter.CompareOperatorDoesNotContain,
			Value:    secondPart,
		}),
		wantDocuments: []TestDoc{doc0, doc2},
	})
	addTest("operator BeginsWith: single value", testCase{
		resultSelector: singleFilter(filter.RequestField{
			Name:     "stringField",
			Operator: filter.CompareOperatorBeginsWith,
			Value:    firstPart,
		}),
		wantDocuments: []TestDoc{doc1},
	})
	addTest("operator DoesNotBeginWith: single value", testCase{
		resultSelector: singleFilter(filter.RequestField{
			Name:     "stringField",
			Operator: filter.CompareOperatorDoesNotBeginWith,
			Value:    firstPart,
		}),
		wantDocuments: []TestDoc{doc0, doc2},
	})
	addTest("operator IsStringCaseInsensitiveEqualTo: single value", testCase{
		resultSelector: singleFilter(filter.RequestField{
			Name:     "stringField",
			Operator: filter.CompareOperatorIsStringCaseInsensitiveEqualTo,
			Value:    strings.ToLower(value),
		}),
		wantDocuments: []TestDoc{doc1},
	})

	// string specific operators, invalid values (only string supported)
	valueTypes = map[string]any{
		"integerField":  doc1.Integer,
		"floatField":    doc1.Float,
		"booleanField":  doc1.Boolean,
		"dateTimeField": doc1.DateTime,
	}
	for fieldName, value := range valueTypes {
		addTest(fmt.Sprintf("operator BeginsWith: invalid value (%v)", fieldName), testCase{
			resultSelector: singleFilter(filter.RequestField{
				Name:     fieldName,
				Operator: filter.CompareOperatorBeginsWith,
				Value:    value,
			}),
			wantErr: true,
		})
		addTest(fmt.Sprintf("operator DoesNotBeginWith: invalid value (%v)", fieldName), testCase{
			resultSelector: singleFilter(filter.RequestField{
				Name:     fieldName,
				Operator: filter.CompareOperatorDoesNotBeginWith,
				Value:    value,
			}),
			wantErr: true,
		})
		addTest(fmt.Sprintf("operator Contains: invalid value (%v)", fieldName), testCase{
			resultSelector: singleFilter(filter.RequestField{
				Name:     fieldName,
				Operator: filter.CompareOperatorContains,
				Value:    value,
			}),
			wantErr: true,
		})
		addTest(fmt.Sprintf("operator DoesNotContain: invalid value (%v)", fieldName), testCase{
			resultSelector: singleFilter(filter.RequestField{
				Name:     fieldName,
				Operator: filter.CompareOperatorDoesNotContain,
				Value:    value,
			}),
			wantErr: true,
		})
		addTest(fmt.Sprintf("operator IsStringCaseInsensitiveEqualTo: invalid value (%v)", fieldName), testCase{
			resultSelector: singleFilter(filter.RequestField{
				Name:     "stringField",
				Operator: filter.CompareOperatorIsStringCaseInsensitiveEqualTo,
				Value:    value,
			}),
			wantErr: true,
		})
	}

	// string specific operators, multiple values
	values := []string{doc1.String, doc2.String}
	var firstParts, secondParts []any
	for _, value := range values {
		require.Greater(t, len(value), 1, "test requires value to be longer than 1 character for `contains` operator")
		firstPart := value[:len(value)/2]
		require.NotEqual(t, firstPart, strings.ToLower(firstPart),
			"test requires mixed case, to verify case insensitivity")
		firstParts = append(firstParts, strings.ToLower(firstPart))
		secondPart := value[len(value)/2:]
		require.NotEqual(t, secondPart, strings.ToLower(secondPart),
			"test requires mixed case, to verify case insensitivity")
		secondParts = append(secondParts, strings.ToLower(secondPart))
	}

	addTest("operator Contains: multiple values", testCase{
		resultSelector: singleFilter(filter.RequestField{
			Name:     "stringField",
			Operator: filter.CompareOperatorContains,
			Value:    secondParts,
		}),
		wantDocuments: []TestDoc{doc1, doc2},
	})
	addTest("operator DoesNotContain: multiple values", testCase{
		resultSelector: singleFilter(filter.RequestField{
			Name:     "stringField",
			Operator: filter.CompareOperatorDoesNotContain,
			Value:    secondParts,
		}),
		wantDocuments: []TestDoc{doc0},
	})
	addTest("operator BeginsWith: multiple values", testCase{
		resultSelector: singleFilter(filter.RequestField{
			Name:     "stringField",
			Operator: filter.CompareOperatorBeginsWith,
			Value:    firstParts,
		}),
		wantDocuments: []TestDoc{doc1, doc2},
	})
	addTest("operator DoesNotBeginWith: multiple values", testCase{
		resultSelector: singleFilter(filter.RequestField{
			Name:     "stringField",
			Operator: filter.CompareOperatorDoesNotBeginWith,
			Value:    firstParts,
		}),
		wantDocuments: []TestDoc{doc0},
	})
	addTest("operator IsStringCaseInsensitiveEqualTo: multiple values", testCase{
		resultSelector: singleFilter(filter.RequestField{
			Name:     "stringField",
			Operator: filter.CompareOperatorIsStringCaseInsensitiveEqualTo,
			Value:    []any{strings.ToLower(doc1.String), strings.ToLower(doc2.String)},
		}),
		wantDocuments: []TestDoc{doc1, doc2},
	})

	// date specific filters, single value
	dateValue := map[string]any{
		"date as string":    doc1.DateTime.Format(time.RFC3339Nano),
		"date as time.Time": doc1.DateTime,
	}
	for valueType, value := range dateValue {
		addTest(fmt.Sprintf("operator AfterDate: single value (%v)", valueType), testCase{
			resultSelector: singleFilter(filter.RequestField{
				Name:     "dateTimeField",
				Operator: filter.CompareOperatorAfterDate,
				Value:    value,
			}),
			wantDocuments: []TestDoc{doc2},
		})
		addTest(fmt.Sprintf("operator BeforDate: single value (%v)", valueType), testCase{
			resultSelector: singleFilter(filter.RequestField{
				Name:     "dateTimeField",
				Operator: filter.CompareOperatorBeforeDate,
				Value:    value,
			}),
			wantDocuments: []TestDoc{doc0},
		})
	}

	// date specific filters, multiple values
	dateValues := map[string][]any{
		"date as string": {
			doc1.DateTime.Format(time.RFC3339Nano),
			doc2.DateTime.Format(time.RFC3339Nano),
		},
		"date as time.Time": {doc1.DateTime, doc2.DateTime},
	}
	for valueType, values := range dateValues {
		addTest(fmt.Sprintf("operator AfterDate: multiple values (%v)", valueType), testCase{
			resultSelector: singleFilter(filter.RequestField{
				Name:     "dateTimeField",
				Operator: filter.CompareOperatorAfterDate,
				Value:    values,
			}),
			wantDocuments: []TestDoc{doc2},
		})
		addTest(fmt.Sprintf("operator BeforDate: multiple values (%v)", valueType), testCase{
			resultSelector: singleFilter(filter.RequestField{
				Name:     "dateTimeField",
				Operator: filter.CompareOperatorBeforeDate,
				Value:    values,
			}),
			wantDocuments: []TestDoc{doc0, doc1},
		})
	}

	// database setup only needed once, as test cases are only reading data
	db := pgtesting.NewDB(t, migrationsFS, migrationDir)
	repo := NewTestRepository(db)
	// make sure order by id (=tie-breaker) is not same as storage order,
	// as postgres will return tied rows in storage order (for a fresh db equal to insert order)
	slices.Reverse(allDocs)
	for _, doc := range allDocs {
		err := repo.CreateTestDoc(&doc)
		require.NoError(t, err, "failed to create test document")
	}
	slices.Reverse(allDocs) // restore original order

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// translate to query
			querySettings := Settings{
				FilterFieldMapping:      fieldMapping,
				SortingTieBreakerColumn: sortingTieBreakerColumn,
			}
			builder, err := NewPostgresQueryBuilder(querySettings)
			require.NoError(t, err, "failed to create Postgres query builder")
			conditionalQuery, args, err := builder.Build(tt.resultSelector)

			if tt.wantErr {
				assert.Error(t, err)
				return
			} else {
				require.NoError(t, err, "unexpected error building query")

				// build full query
				fullQuery := unfilteredListTestypesQuery + ` ` + conditionalQuery

				t.Logf("sending query to database: query: %v, args: %v (length: %d)", fullQuery, args, len(args))
				gotDocs, err := repo.ListTestDocs(fullQuery, args)
				require.NoError(t, err, "failed to list documents")

				for i := range gotDocs {
					// in go struct UTC is just location = `nil`, but when read from DB it is explicitly UTC
					gotDocs[i].DateTime = gotDocs[i].DateTime.UTC()
				}

				if len(tt.wantDocuments) == 0 && len(gotDocs) == 0 {
					// nothing more to check
					// here we don't care if result is nil or empty slice
					return
				}
				assert.Equal(t, tt.wantDocuments, gotDocs)
			}
		})
	}
}

func Test_NewPostgresQueryBuilder(t *testing.T) {
	tests := []struct {
		name     string
		settings Settings
		wantErr  bool
	}{
		{
			name: "success",
			settings: Settings{
				SortingTieBreakerColumn: "id",
			},
			wantErr: false,
		},
		{
			name:     "failure with missing mandatory setting",
			settings: Settings{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder, err := NewPostgresQueryBuilder(tt.settings)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, builder)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, builder)
			}
		})
	}
}
