package registry

type LayoutOptions struct {
	SampleLimit         int
	StateReferenceFiles []string
}

type LayoutReport struct {
	SchemaVersion int              `json:"schema_version"`
	Summary       LayoutSummary    `json:"summary"`
	Locations     []LayoutLocation `json:"locations"`
}

type LayoutSummary struct {
	Locations                   int `json:"locations"`
	Files                       int `json:"files"`
	OwnedFiles                  int `json:"owned_files"`
	UnownedFiles                int `json:"unowned_files"`
	CleanupCandidateFiles       int `json:"cleanup_candidate_files"`
	DirectCleanupCandidateFiles int `json:"direct_cleanup_candidate_files"`
	ReviewPrunableFiles         int `json:"review_prunable_files"`
	RegistryReportFiles         int `json:"registry_report_files"`
	BlockedUnownedFiles         int `json:"blocked_unowned_files"`
	LocalOnlyFiles              int `json:"local_only_files"`
	MissingRequired             int `json:"missing_required"`
	SampleLimit                 int `json:"sample_limit"`
}

type LayoutLocation struct {
	ID                          string         `json:"id"`
	Path                        string         `json:"path"`
	Class                       string         `json:"class"`
	Lifecycle                   string         `json:"lifecycle"`
	Tracked                     bool           `json:"tracked"`
	Required                    bool           `json:"required"`
	Exists                      bool           `json:"exists"`
	Files                       int            `json:"files"`
	Bytes                       int64          `json:"bytes"`
	OwnedFiles                  int            `json:"owned_files"`
	UnownedFiles                int            `json:"unowned_files"`
	CleanupCandidateFiles       int            `json:"cleanup_candidate_files,omitempty"`
	DirectCleanupCandidateFiles int            `json:"direct_cleanup_candidate_files,omitempty"`
	ReviewPrunableFiles         int            `json:"review_prunable_files,omitempty"`
	RegistryReportFiles         int            `json:"registry_report_files,omitempty"`
	BlockedUnownedFiles         int            `json:"blocked_unowned_files,omitempty"`
	LocalOnlyFiles              int            `json:"local_only_files,omitempty"`
	Action                      string         `json:"action"`
	CleanupSafety               string         `json:"cleanup_safety"`
	Reason                      string         `json:"reason"`
	UnownedSamples              []string       `json:"unowned_samples,omitempty"`
	UnownedGroups               []UnownedGroup `json:"unowned_groups,omitempty"`
}

func LayoutRepository(root string, opts LayoutOptions) (LayoutReport, error) {
	reg, err := Load(root)
	if err != nil {
		return LayoutReport{}, err
	}
	return Layout(root, reg, opts)
}

func Layout(root string, reg Registry, opts LayoutOptions) (LayoutReport, error) {
	if opts.SampleLimit < 0 {
		opts.SampleLimit = 0
	}
	inventory, err := Inventory(root, reg)
	if err != nil {
		return LayoutReport{}, err
	}
	ownership, err := Ownership(root, reg, OwnershipOptions{SampleLimit: opts.SampleLimit})
	if err != nil {
		return LayoutReport{}, err
	}
	ownershipByID := map[string]LocationOwnership{}
	for _, location := range ownership.Locations {
		ownershipByID[location.ID] = location
	}
	generatedCleanup, err := GeneratedCleanup(root, reg, GeneratedCleanupOptions{
		SampleLimit:         opts.SampleLimit,
		StateReferenceFiles: opts.StateReferenceFiles,
	})
	if err != nil {
		return LayoutReport{}, err
	}
	generatedCleanupByID := map[string]GeneratedCleanupLocation{}
	for _, location := range generatedCleanup.Locations {
		generatedCleanupByID[location.ID] = location
	}

	report := LayoutReport{
		SchemaVersion: 1,
		Summary: LayoutSummary{
			Locations:   len(inventory.Locations),
			SampleLimit: opts.SampleLimit,
		},
		Locations: make([]LayoutLocation, 0, len(inventory.Locations)),
	}
	for _, inv := range inventory.Locations {
		own := ownershipByID[inv.ID]
		location := layoutLocation(inv, own, generatedCleanupByID[inv.ID])
		report.Locations = append(report.Locations, location)
		report.Summary.Files += location.Files
		report.Summary.OwnedFiles += location.OwnedFiles
		report.Summary.UnownedFiles += location.UnownedFiles
		report.Summary.CleanupCandidateFiles += location.CleanupCandidateFiles
		report.Summary.DirectCleanupCandidateFiles += location.DirectCleanupCandidateFiles
		report.Summary.ReviewPrunableFiles += location.ReviewPrunableFiles
		report.Summary.RegistryReportFiles += location.RegistryReportFiles
		report.Summary.BlockedUnownedFiles += location.BlockedUnownedFiles
		report.Summary.LocalOnlyFiles += location.LocalOnlyFiles
		if location.Action == "restore_missing_required" {
			report.Summary.MissingRequired++
		}
	}
	return report, nil
}

func layoutLocation(inv LocationInventory, own LocationOwnership, generated GeneratedCleanupLocation) LayoutLocation {
	location := LayoutLocation{
		ID:             inv.ID,
		Path:           inv.Path,
		Class:          inv.Class,
		Lifecycle:      inv.Lifecycle,
		Tracked:        inv.Tracked,
		Required:       inv.Required,
		Exists:         inv.Exists,
		Files:          inv.Files,
		Bytes:          inv.Bytes,
		OwnedFiles:     own.OwnedFiles,
		UnownedFiles:   own.UnownedFiles,
		UnownedSamples: own.UnownedSamples,
		UnownedGroups:  own.UnownedGroups,
	}
	location.Action, location.CleanupSafety, location.Reason = layoutDecision(location)
	switch location.Action {
	case "review_generated_cleanup":
		applyGeneratedCleanupLayout(&location, generated)
	case "atomize_or_register_unowned":
		location.BlockedUnownedFiles = location.UnownedFiles
	case "exclude_local_only":
		location.LocalOnlyFiles = location.Files
	}
	return location
}

func applyGeneratedCleanupLayout(location *LayoutLocation, generated GeneratedCleanupLocation) {
	for _, group := range generated.Groups {
		switch group.Action {
		case "cleanup_candidate":
			location.CleanupCandidateFiles += group.UnreferencedFiles
			location.DirectCleanupCandidateFiles += group.UnreferencedFiles
		case "review_mixed_registered_generated", "review_state_referenced_generated":
			location.CleanupCandidateFiles += group.UnreferencedFiles
			location.ReviewPrunableFiles += group.UnreferencedFiles
		case "retain_registry_report_outputs":
			location.RegistryReportFiles += group.Files
		}
	}
}

func layoutDecision(location LayoutLocation) (action, safety, reason string) {
	if !location.Exists {
		if location.Required {
			return "restore_missing_required", "blocked_missing_required", "required registered location is missing"
		}
		return "keep_registered", "missing_optional", "optional registered location is absent"
	}
	if location.Lifecycle == "local_only" || location.Class == "local_corpus" {
		return "exclude_local_only", "not_project_cleanup", "local-only corpus should stay outside tracked cleanup and atom ledgers"
	}
	if !location.Tracked && location.Class == "generated_evidence" {
		return "review_generated_cleanup", "generated_untracked_review", "unowned generated files may be deleted only after the producing workflow is known"
	}
	if location.UnownedFiles > 0 && (location.Tracked || location.Required) {
		return "atomize_or_register_unowned", "blocked_until_owned", "tracked or required files need atom, evidence, or location ownership before cleanup"
	}
	return "keep_registered", "no_cleanup_action", "registered location is owned or intentionally retained"
}
