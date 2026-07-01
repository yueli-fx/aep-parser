package aepmigrate

import (
	"encoding/json"

	_ "embed"
)

//go:embed version_capability_ledger.json
var defaultCapabilityLedgerJSON []byte

type CapabilitySupport string

const (
	CapabilityPreserved CapabilitySupport = "preserved"
	CapabilityBlocked   CapabilitySupport = "blocked"
)

type VersionCapabilityLedger struct {
	SchemaVersion int                     `json:"schema_version"`
	Rules         []VersionCapabilityRule `json:"rules"`
}

type VersionCapabilityRule struct {
	ID              string                             `json:"id"`
	Path            string                             `json:"path"`
	Kind            string                             `json:"kind"`
	CapabilityKey   string                             `json:"capability_key"`
	SourceVersions  []VersionLabel                     `json:"source_versions"`
	TargetSupport   map[VersionLabel]CapabilitySupport `json:"target_support"`
	BlockedReason   string                             `json:"blocked_reason"`
	PreservedReason string                             `json:"preserved_reason"`
}

var defaultCapabilityLedger = mustLoadCapabilityLedger(defaultCapabilityLedgerJSON)

func DefaultCapabilityLedger() VersionCapabilityLedger {
	return defaultCapabilityLedger
}

func (ledger VersionCapabilityLedger) RuleByID(id string) (VersionCapabilityRule, bool) {
	for _, rule := range ledger.Rules {
		if rule.ID == id {
			return rule, true
		}
	}
	return VersionCapabilityRule{}, false
}

func (rule VersionCapabilityRule) AppliesToSource(source VersionLabel) bool {
	for _, candidate := range rule.SourceVersions {
		if candidate == source {
			return true
		}
	}
	return false
}

func mustLoadCapabilityLedger(data []byte) VersionCapabilityLedger {
	var ledger VersionCapabilityLedger
	if err := json.Unmarshal(data, &ledger); err != nil {
		panic(err)
	}
	return ledger
}
