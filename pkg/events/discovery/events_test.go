// SPDX-FileCopyrightText: 2026 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package discovery

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/greenbone-hive/opensight-golang-libraries/pkg/events"
)

func TestEventJSONRoundTrip(t *testing.T) {
	in := ScopeRetired{
		Meta: events.Meta{
			ID:       "evt-1",
			Type:     SubjectScopeRetired,
			Source:   "discovery",
			Time:     time.Unix(1700000000, 0).UTC(),
			EntityID: "asset-1",
			Version:  3,
		},
		Provenance: Provenance{
			SourceID: "src-1", SourceRevision: 2,
			TargetScopeID: "111111111111", RunID: "10", ScopeRunID: "77",
			Provider: "aws",
		},
		Reason: ReasonScopeExcluded,
	}

	data, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var out ScopeRetired
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.EntityID != "asset-1" || out.Version != 3 {
		t.Errorf("meta not round-tripped: %+v", out.Meta)
	}
	if out.Reason != ReasonScopeExcluded || out.Provenance != in.Provenance {
		t.Errorf("payload not round-tripped: %+v", out)
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	for _, nested := range []string{"Meta", "Provenance"} {
		if _, ok := raw[nested]; ok {
			t.Errorf("%s should be embedded/promoted, not nested under its own key", nested)
		}
	}
	if raw["entity_id"] != "asset-1" {
		t.Errorf("entity_id should be a top-level field, got %v", raw["entity_id"])
	}
}
