// SPDX-FileCopyrightText: 2026 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package discovery

import (
	"errors"
	"fmt"
)

// CoverageStatus is how completely a scan partition was read. Only
// CoverageComplete authorizes removals; partial and failed can only upsert.
type CoverageStatus string

const (
	CoverageComplete CoverageStatus = "complete"
	CoveragePartial  CoverageStatus = "partial"
	CoverageFailed   CoverageStatus = "failed"
)

// LifecycleReason says WHY something left a consumer's view. Only
// ReasonResourceDeleted asserts the resource is gone at the provider; every
// other reason is a control-plane or authorization change.
type LifecycleReason string

const (
	ReasonResourceDeleted LifecycleReason = "resource_deleted"
	ReasonScopeExcluded   LifecycleReason = "scope_excluded"
	ReasonSourceDeleted   LifecycleReason = "source_deleted"
	// ReasonAuthorizationLost keeps the claim: the producer lost read access, so
	// coverage goes stale rather than the claim being retired.
	ReasonAuthorizationLost LifecycleReason = "authorization_lost"
	ReasonTargetMoved       LifecycleReason = "target_moved"
	ReasonTargetClosed      LifecycleReason = "target_closed"
)

// RetiresClaim is false for resource deletion: that travels per-resource in a
// scan, not as a retirement.
func (r LifecycleReason) RetiresClaim() bool {
	switch r {
	case ReasonScopeExcluded, ReasonSourceDeleted, ReasonTargetMoved, ReasonTargetClosed:
		return true
	case ReasonResourceDeleted, ReasonAuthorizationLost:
		return false
	}

	return false
}

func (r LifecycleReason) IsProviderDeletion() bool { return r == ReasonResourceDeleted }

// Provenance travels intact through discovery -> asset management -> exposure.
// SourceID is provenance, NOT identity: canonical identity is (provider,
// canonicalResourceId), so two sources observing one resource yield two
// source claims and one asset.
type Provenance struct {
	SourceID       string `json:"source_id"`
	SourceRevision int64  `json:"source_revision"`
	TargetScopeID  string `json:"target_scope_id"`
	RunID          string `json:"run_id"`
	ScopeRunID     string `json:"scope_run_id"`
	Provider       string `json:"provider"`
}

// PartitionKey is the event-stream identity of one target partition. Producers
// must set events.Meta.EntityID to it: partitions are ordered independently and
// must never gate each other.
func (p *Provenance) PartitionKey() string {
	return p.SourceID + "|" + p.TargetScopeID
}

func (p *Provenance) validate() error {
	switch {
	case p.SourceID == "":
		return errors.New("discovery: provenance missing source_id")
	case p.TargetScopeID == "":
		return errors.New("discovery: provenance missing target_scope_id")
	case p.Provider == "":
		return errors.New("discovery: provenance missing provider")
	}

	return nil
}

// CollectorCoverage carries "global" as RegionOrGlobal for collectors that are
// not regional.
type CollectorCoverage struct {
	Collector      string         `json:"collector"`
	RegionOrGlobal string         `json:"region_or_global"`
	Status         CoverageStatus `json:"status"`
	Error          string         `json:"error,omitempty"`
}

// Validate enforces what the producer can be held to. The absence rule is not
// among it: absence is implicit, so consumers must gate their reap on
// Coverage == CoverageComplete themselves.
func (e *ScanCompleted) Validate() error {
	if err := e.validate(); err != nil {
		return err
	}
	if e.RunID == "" || e.ScopeRunID == "" {
		return errors.New("discovery: scan missing run/scope-run id")
	}
	if e.EntityID != e.PartitionKey() {
		return fmt.Errorf("discovery: scan entity_id %q must be the partition key %q", e.EntityID, e.PartitionKey())
	}
	if e.Version <= 0 {
		return errors.New("discovery: scan needs a positive per-partition version")
	}
	for i := range e.Resources {
		if e.Resources[i].Type == "" || e.Resources[i].ProviderResourceID == "" {
			return fmt.Errorf("discovery: scan resource %d missing type or provider_resource_id", i)
		}
	}

	return nil
}

func (e *ScopeRetired) Validate() error {
	if err := e.validate(); err != nil {
		return err
	}
	if !e.Reason.RetiresClaim() {
		return fmt.Errorf("discovery: scope retirement reason %q does not retire claims", e.Reason)
	}
	if e.EntityID != e.PartitionKey() {
		return fmt.Errorf("discovery: scope retirement entity_id %q must be the partition key %q", e.EntityID, e.PartitionKey())
	}

	return nil
}
