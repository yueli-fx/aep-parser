package registry

type AssetPolicyReport struct {
	SchemaVersion int                `json:"schema_version"`
	Status        string             `json:"status"`
	Summary       AssetPolicySummary `json:"summary"`
	Issues        []AssetPolicyIssue `json:"issues,omitempty"`
}

type AssetPolicySummary struct {
	Rules                    int `json:"rules"`
	RequiredLocations        int `json:"required_locations"`
	CoveredRequiredLocations int `json:"covered_required_locations"`
	Errors                   int `json:"errors"`
}

type AssetPolicyIssue struct {
	Code       string `json:"code"`
	RuleID     string `json:"rule_id,omitempty"`
	LocationID string `json:"location_id,omitempty"`
	Message    string `json:"message"`
}

func ValidateAssetPolicyRepository(root string) (AssetPolicyReport, error) {
	report := AssetPolicyReport{
		SchemaVersion: 1,
		Status:        StatusPass,
	}
	if !exists(root, "registry/asset_policy.json") {
		report.addIssue("missing_asset_policy", "", "", "registry/asset_policy.json is required for the asset-policy gate")
		return report, nil
	}
	reg, err := Load(root)
	if err != nil {
		return AssetPolicyReport{}, err
	}
	return ValidateAssetPolicy(reg), nil
}

func ValidateAssetPolicy(reg Registry) AssetPolicyReport {
	report := AssetPolicyReport{
		SchemaVersion: 1,
		Status:        StatusPass,
		Summary: AssetPolicySummary{
			Rules: len(reg.AssetPolicyRules),
		},
	}
	locationIDs := map[string]bool{}
	for _, location := range reg.Locations {
		locationIDs[location.ID] = true
		if assetPolicyRequiredForLocation(location) {
			report.Summary.RequiredLocations++
		}
	}
	ruleIDs := map[string]bool{}
	rulesByLocation := map[string]bool{}
	for _, rule := range reg.AssetPolicyRules {
		if rule.ID == "" {
			report.addIssue("missing_asset_policy_rule_id", "", rule.LocationID, "asset policy rule id is required")
		} else if ruleIDs[rule.ID] {
			report.addIssue("duplicate_asset_policy_rule", rule.ID, "", "asset policy rule id must be unique")
		}
		ruleIDs[rule.ID] = true
		if rule.LocationID == "" {
			report.addIssue("missing_asset_policy_location", rule.ID, "", "asset policy rule location_id is required")
		} else if !locationIDs[rule.LocationID] {
			report.addIssue("unknown_asset_policy_location", rule.ID, rule.LocationID, "asset policy rule references an unknown location")
		} else if !rulesByLocation[rule.LocationID] {
			rulesByLocation[rule.LocationID] = true
		}
		if rule.Action == "" {
			report.addIssue("missing_asset_policy_action", rule.ID, "", "asset policy rule action is required")
		}
		if rule.CleanupSafety == "" {
			report.addIssue("missing_asset_policy_cleanup_safety", rule.ID, "", "asset policy rule cleanup_safety is required")
		}
	}
	for _, location := range reg.Locations {
		if !assetPolicyRequiredForLocation(location) {
			continue
		}
		if rulesByLocation[location.ID] {
			report.Summary.CoveredRequiredLocations++
			continue
		}
		report.addIssue("missing_asset_policy_location_rule", "", location.ID, "generated or local-only locations must have an asset policy rule")
	}
	return report
}

func (r *AssetPolicyReport) addIssue(code, ruleID, locationID, message string) {
	r.Issues = append(r.Issues, AssetPolicyIssue{
		Code:       code,
		RuleID:     ruleID,
		LocationID: locationID,
		Message:    message,
	})
	r.Summary.Errors++
	r.Status = StatusFail
}
