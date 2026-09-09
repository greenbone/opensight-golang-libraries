// SPDX-FileCopyrightText: 2026 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package provider

import "testing"

func TestValid(t *testing.T) {
	for _, p := range All {
		if !Valid(p) {
			t.Errorf("Valid(%q) = false, want true (it is in All)", p)
		}
	}
	for _, p := range []Provider{"", "AWS", "gcpx", "oracle"} {
		if Valid(p) {
			t.Errorf("Valid(%q) = true, want false", p)
		}
	}
}

func TestAllHasNoDuplicates(t *testing.T) {
	seen := map[Provider]struct{}{}
	for _, p := range All {
		if _, dup := seen[p]; dup {
			t.Errorf("All lists %q more than once", p)
		}
		seen[p] = struct{}{}
	}
	if len(All) != 5 {
		t.Fatal("All providers should be in list", len(All))
	}
}
