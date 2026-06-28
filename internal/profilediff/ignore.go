package profilediff

import (
	"encoding/json"
	"fmt"
	"os"
)

type IgnoreRules struct {
	SchemaVersion int          `json:"schema_version"`
	Rules         []IgnoreRule `json:"rules,omitempty"`
}

type IgnoreRule struct {
	Path        string `json:"path"`
	Kind        Kind   `json:"kind"`
	Condition   string `json:"condition,omitempty"`
	Reason      string `json:"reason"`
	ExpiresWhen string `json:"expires_when,omitempty"`
}

func LoadIgnoreRules(path string) (*IgnoreRules, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("profilediff: read ignore rules: %w", err)
	}
	var rules IgnoreRules
	if err := json.Unmarshal(data, &rules); err != nil {
		return nil, fmt.Errorf("profilediff: parse ignore rules: %w", err)
	}
	if err := rules.Validate(); err != nil {
		return nil, err
	}
	return &rules, nil
}

func (r *IgnoreRules) Validate() error {
	if r == nil {
		return nil
	}
	if r.SchemaVersion != 1 {
		return fmt.Errorf("profilediff: unsupported ignore schema_version %d", r.SchemaVersion)
	}
	for i, rule := range r.Rules {
		if rule.Path == "" {
			return fmt.Errorf("profilediff: ignore rule %d has empty path", i)
		}
		if rule.Kind == "" {
			return fmt.Errorf("profilediff: ignore rule %d has empty kind", i)
		}
		if rule.Reason == "" {
			return fmt.Errorf("profilediff: ignore rule %d has empty reason", i)
		}
	}
	return nil
}

func (r *IgnoreRules) Matches(diff Diff) bool {
	if r == nil {
		return false
	}
	for _, rule := range r.Rules {
		if rule.Path == diff.Path && rule.Kind == diff.Kind {
			return true
		}
	}
	return false
}
