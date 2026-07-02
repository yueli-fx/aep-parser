package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/profile"
)

const explicitMatteSourceRuleID = "layer-explicit-matte-source"

type profileCapabilityFinding struct {
	Path string
	Rule VersionCapabilityRule
}

func capabilityBlockerEntries(source, target VersionLabel, prof *profile.Profile, ledger VersionCapabilityLedger) []Entry {
	var entries []Entry
	for _, finding := range profileCapabilityFindings(prof, ledger) {
		if !finding.Rule.Blocks(source, target) {
			continue
		}
		entries = append(entries, Entry{
			Path:          finding.Path,
			Class:         ClassBlocked,
			TargetVersion: target,
			Reason:        finding.Rule.BlockedReason,
			CapabilityKey: finding.Rule.CapabilityKey,
		})
	}
	return entries
}

func profileCapabilityFindings(prof *profile.Profile, ledger VersionCapabilityLedger) []profileCapabilityFinding {
	if prof == nil {
		return nil
	}
	var findings []profileCapabilityFinding
	if explicitMatteRule, ok := ledger.RuleByID(explicitMatteSourceRuleID); ok {
		findings = append(findings, explicitMatteCapabilityFindings(prof, explicitMatteRule)...)
	}
	return findings
}

func explicitMatteCapabilityFindings(prof *profile.Profile, rule VersionCapabilityRule) []profileCapabilityFinding {
	var findings []profileCapabilityFinding
	for _, comp := range prof.Comps {
		for _, layer := range comp.Layers {
			if !isExplicitMatteRef(layer) {
				continue
			}
			findings = append(findings, profileCapabilityFinding{
				Path: fmt.Sprintf("comps[%q].layers[%q].matte_ref", comp.Name, layer.Name),
				Rule: rule,
			})
		}
	}
	return findings
}

func (rule VersionCapabilityRule) SupportForTarget(target VersionLabel) CapabilitySupport {
	return rule.TargetSupport[target]
}

func (rule VersionCapabilityRule) Blocks(source, target VersionLabel) bool {
	return rule.AppliesToSource(source) && rule.SupportForTarget(target) == CapabilityBlocked
}
