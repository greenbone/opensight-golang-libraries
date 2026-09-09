// SPDX-FileCopyrightText: 2026 Greenbone AG
//
// SPDX-License-Identifier: AGPL-3.0-or-later

// Package discovery is the discovery -> asset service event contract: the two
// payloads discovery pushes and the provenance that says which observation
// produced them.
package discovery

import (
	"encoding/json"

	"github.com/greenbone-hive/opensight-golang-libraries/pkg/events"
)

// ScanCompleted carries the partition's COMPLETE live resource set inline.
// Discovery computes no diff and keeps no resource state; the consumer
// reconciles against its own previous state.
//
// A resource missing from Resources is a deletion assertion ONLY when Coverage
// is complete. A partial or failed scan says nothing about absence, so
// consumers apply it upsert-only and never reap on it. Scope exclusion,
// source deletion and target moves are not absences: they travel as
// ScopeRetired.
type ScanCompleted struct {
	events.Meta
	Provenance
	Account       string              `json:"account"`
	TriggerSource string              `json:"trigger_source"`
	Coverage      CoverageStatus      `json:"coverage"`
	Collectors    []CollectorCoverage `json:"collectors,omitempty"`
	Resources     []ObservedResource  `json:"resources"`
}

// ObservedResource carries the resource in Body as marshaled JSON: name,
// resourceName, assetType, computed, tags and identifiers.
type ObservedResource struct {
	Type               string          `json:"type"`
	ProviderResourceID string          `json:"provider_resource_id"`
	Body               json.RawMessage `json:"body"`
}

// ScopeRetired never asserts the partition's resources were deleted at the
// provider; the reason says which control-plane change removed it.
type ScopeRetired struct {
	events.Meta
	Provenance
	Reason LifecycleReason `json:"reason"`
}
