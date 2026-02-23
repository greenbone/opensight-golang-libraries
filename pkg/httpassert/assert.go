package httpassert

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NormalizeJSON parses JSON and re-marshals it with stable key ordering and indentation.
// - Uses Decoder.UseNumber() to preserve numeric fidelity (avoid float64 surprises).
func NormalizeJSON(t *testing.T, s string) string {
	t.Helper()

	dec := json.NewDecoder(bytes.NewReader([]byte(s)))
	dec.UseNumber()

	var v any
	require.NoError(t, dec.Decode(&v), "invalid JSON")

	b, err := json.MarshalIndent(v, "", "  ")
	require.NoError(t, err, "failed to marshal normalized JSON")

	return string(b)
}

// AssertJSONCanonicalEq compares two JSON strings by normalizing both first.
// On mismatch, it prints a readable diff of the normalized forms.
func AssertJSONCanonicalEq(t *testing.T, expected, actual string) bool {
	t.Helper()

	expNorm := NormalizeJSON(t, expected)
	actNorm := NormalizeJSON(t, actual)

	if expNorm != actNorm {
		diff := difflib.UnifiedDiff{
			A:        difflib.SplitLines(expNorm),
			B:        difflib.SplitLines(actNorm),
			FromFile: "Expected",
			ToFile:   "Actual",
			Context:  3,
		}

		text, _ := difflib.GetUnifiedDiffString(diff)

		return assert.Fail(t, "JSON mismatch:\nDiff:\n"+text)
	}
	return true
}
