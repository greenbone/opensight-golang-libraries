// SPDX-FileCopyrightText: 2026 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package discovery

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/greenbone-hive/opensight-golang-libraries/pkg/events"
)

func scan(target string, version int64, coverage CoverageStatus, resources ...ObservedResource) ScanCompleted {
	p := Provenance{
		SourceID: "src-1", SourceRevision: 2,
		TargetScopeID: target, RunID: "10", ScopeRunID: "77", Provider: "aws",
	}

	return ScanCompleted{
		Meta: events.Meta{
			ID: "evt", Type: SubjectScanCompleted, Source: "discovery",
			EntityID: p.PartitionKey(), Version: version,
		},
		Provenance: p,
		Account:    target,
		Coverage:   coverage,
		Resources:  resources,
	}
}

// Two target partitions under one source have independent ordering keys:
// neither event's (entity_id, version) guard can drop the other, in any
// delivery order.
func TestTargetPartitionsOrderIndependently(t *testing.T) {
	a := scan("111111111111", 5, CoverageComplete)
	b := scan("222222222222", 3, CoverageComplete)

	if a.EntityID == b.EntityID {
		t.Fatalf("partitions share entity id %q; one target would gate the other", a.EntityID)
	}
	if err := a.Validate(); err != nil {
		t.Fatalf("a invalid: %v", err)
	}
	if err := b.Validate(); err != nil {
		t.Fatalf("b invalid: %v", err)
	}
	// Reverse-order delivery: b (version 3) after a (version 5) is still a fresh
	// key for b's own stream.
	if a.Key() == b.Key() {
		t.Fatalf("keys collide: %q", a.Key())
	}
}

// Duplicate delivery of one partition event yields the identical idempotency
// key, so a consumer's high-water mark drops it exactly once.
func TestDuplicateDeliverySameKey(t *testing.T) {
	a := scan("111111111111", 5, CoverageComplete)
	dup := scan("111111111111", 5, CoverageComplete)
	if a.Key() != dup.Key() {
		t.Fatalf("duplicate delivery changed key: %q vs %q", a.Key(), dup.Key())
	}
}

// Every inline resource must carry its identity; an empty scan is legal
// (a partition can genuinely hold zero resources).
func TestScanResourceIdentityRequired(t *testing.T) {
	empty := scan("111111111111", 6, CoverageComplete)
	if err := empty.Validate(); err != nil {
		t.Fatalf("empty scan rejected: %v", err)
	}

	e := scan("111111111111", 6, CoverageComplete,
		ObservedResource{Type: "AWS::EC2::VPC", ProviderResourceID: "vpc-1"},
		ObservedResource{Type: "", ProviderResourceID: "vpc-2"})
	err := e.Validate()
	if err == nil || !strings.Contains(err.Error(), "missing type or provider_resource_id") {
		t.Fatalf("identity-less resource validated: %v", err)
	}
}

// The scan entity id must be the partition key, and versions must be
// positive per-partition sequences.
func TestScanOrderingContract(t *testing.T) {
	e := scan("111111111111", 6, CoverageComplete)
	e.EntityID = "src-1" // source-wide stream: two targets would share it
	if err := e.Validate(); err == nil {
		t.Fatal("non-partition entity id validated")
	}

	e = scan("111111111111", 0, CoverageComplete)
	if err := e.Validate(); err == nil {
		t.Fatal("zero version validated")
	}

	for _, missing := range []func(*ScanCompleted){
		func(e *ScanCompleted) { e.SourceID = "" },
		func(e *ScanCompleted) { e.TargetScopeID = "" },
		func(e *ScanCompleted) { e.Provider = "" },
		func(e *ScanCompleted) { e.ScopeRunID = "" },
	} {
		bad := scan("111111111111", 6, CoverageComplete)
		missing(&bad)
		bad.EntityID = bad.PartitionKey()
		if err := bad.Validate(); err == nil {
			t.Fatalf("scan with missing provenance validated: %+v", bad.Provenance)
		}
	}
}

// Lifecycle reasons: only resource_deleted asserts provider deletion; only the
// four control-plane reasons retire claims; authorization loss does neither.
func TestLifecycleReasonSemantics(t *testing.T) {
	if !ReasonResourceDeleted.IsProviderDeletion() {
		t.Fatal("resource_deleted must assert provider deletion")
	}
	for _, r := range []LifecycleReason{
		ReasonScopeExcluded, ReasonSourceDeleted,
		ReasonAuthorizationLost, ReasonTargetMoved, ReasonTargetClosed,
	} {
		if r.IsProviderDeletion() {
			t.Fatalf("%s must not assert provider deletion", r)
		}
	}
	for _, r := range []LifecycleReason{
		ReasonScopeExcluded, ReasonSourceDeleted,
		ReasonTargetMoved, ReasonTargetClosed,
	} {
		if !r.RetiresClaim() {
			t.Fatalf("%s must retire claims", r)
		}
	}
	for _, r := range []LifecycleReason{ReasonResourceDeleted, ReasonAuthorizationLost} {
		if r.RetiresClaim() {
			t.Fatalf("%s must not retire claims", r)
		}
	}
}

// A scope retirement only carries claim-retiring reasons: provider deletion
// travels per-resource in scans, and authorization loss keeps the claim.
func TestScopeRetirementReasonGate(t *testing.T) {
	p := Provenance{SourceID: "src-1", TargetScopeID: "111111111111", Provider: "aws"}

	ok := ScopeRetired{
		Meta:       events.Meta{EntityID: p.PartitionKey(), Version: 9},
		Provenance: p,
		Reason:     ReasonScopeExcluded,
	}
	if err := ok.Validate(); err != nil {
		t.Fatalf("scope_excluded retirement rejected: %v", err)
	}

	for _, r := range []LifecycleReason{ReasonResourceDeleted, ReasonAuthorizationLost} {
		bad := ok
		bad.Reason = r
		if err := bad.Validate(); err == nil {
			t.Fatalf("retirement with reason %s validated", r)
		}
	}
}

// The full provenance survives a JSON round trip on the scan, and the
// asset event carries reason and scope provenance downstream.
func TestProvenanceJSONRoundTrip(t *testing.T) {
	in := scan("111111111111", 6, CoverageComplete,
		ObservedResource{
			Type: "AWS::EC2::VPC", ProviderResourceID: "vpc-1",
			Body: json.RawMessage(`{"name":"main"}`),
		})
	in.Collectors = []CollectorCoverage{
		{Collector: "aws.ec2.network_core", RegionOrGlobal: "eu-central-1", Status: CoverageComplete},
		{Collector: "aws.route53", RegionOrGlobal: "global", Status: CoverageComplete},
	}

	data, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var out ScanCompleted
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Provenance != in.Provenance {
		t.Fatalf("provenance not round-tripped: %+v", out.Provenance)
	}
	if len(out.Collectors) != 2 || out.Collectors[1].RegionOrGlobal != "global" {
		t.Fatalf("collector coverage not round-tripped: %+v", out.Collectors)
	}
	if err := out.Validate(); err != nil {
		t.Fatalf("round-tripped scan invalid: %v", err)
	}
}
