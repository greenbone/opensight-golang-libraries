package query

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetQuotedName(t *testing.T) {
	tests := map[string]struct {
		fieldName               string
		expectedQuotedFieldName string
		errMessage              string
	}{
		"shouldReturnEmptyValue": {fieldName: "", expectedQuotedFieldName: `""`},
		"shouldReturnQuotedFields": {
			fieldName:               "table.field",
			expectedQuotedFieldName: `"table"."field"`,
		},
		"shouldReturnAnErrorWhenTableNameIsEmpty": {
			fieldName:               ".field",
			expectedQuotedFieldName: ``, errMessage: `table name can not be empty`,
		},
		"shouldReturnAnErrorWhenFieldNameIsEmpty": {
			fieldName:  "table.",
			errMessage: `field name can not be empty`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			quotedName, err := getQuotedName(tc.fieldName)
			if tc.errMessage != "" {
				require.ErrorContains(t, err, tc.errMessage)
			}
			assert.Equal(t, tc.expectedQuotedFieldName, quotedName)
		})
	}
}
