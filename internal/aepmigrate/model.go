package aepmigrate

const SchemaVersion = 1

type VersionLabel string

const (
	VersionUnknown VersionLabel = "unknown"
	VersionAE2020  VersionLabel = "AE2020"
	VersionAE2021  VersionLabel = "AE2021"
	VersionAE2022  VersionLabel = "AE2022"
	VersionAE2023  VersionLabel = "AE2023"
	VersionAE2024  VersionLabel = "AE2024"
	VersionAE2025  VersionLabel = "AE2025"
)

type Status string

const (
	StatusPass    Status = "pass"
	StatusWarn    Status = "warn"
	StatusBlocked Status = "blocked"
	StatusError   Status = "error"
)

type MigrationClass string

const (
	ClassPreserved    MigrationClass = "preserved"
	ClassRetargeted   MigrationClass = "retargeted"
	ClassTranslated   MigrationClass = "translated"
	ClassApproximated MigrationClass = "approximated"
	ClassDropped      MigrationClass = "dropped"
	ClassBlocked      MigrationClass = "blocked"
	ClassUnknown      MigrationClass = "unknown"
)

type SourceInfo struct {
	Path       string       `json:"path"`
	Version    VersionLabel `json:"version_label"`
	VersionRaw string       `json:"version_raw,omitempty"`
}

type TargetInfo struct {
	Version VersionLabel `json:"version_label"`
	Path    string       `json:"path,omitempty"`
}

type Summary struct {
	Status       Status `json:"status"`
	Preserved    int    `json:"preserved"`
	Retargeted   int    `json:"retargeted"`
	Translated   int    `json:"translated"`
	Approximated int    `json:"approximated"`
	Dropped      int    `json:"dropped"`
	Blocked      int    `json:"blocked"`
	Unknown      int    `json:"unknown"`
}

type Entry struct {
	Path          string         `json:"path"`
	Class         MigrationClass `json:"class"`
	TargetVersion VersionLabel   `json:"target_version"`
	Reason        string         `json:"reason"`
	CapabilityKey string         `json:"capability_key,omitempty"`
}

type Verification struct {
	ProfileDiffStatus       string             `json:"profile_diff_status"`
	ProfileDiffCount        int                `json:"profile_diff_count"`
	ProfileDiffIgnoredCount int                `json:"profile_diff_ignored_count,omitempty"`
	ProfileDiffs            []VerificationDiff `json:"profile_diffs,omitempty"`
	AEOpenStatus            string             `json:"ae_open_status"`
	AEOpenExitCode          *int               `json:"ae_open_exit_code,omitempty"`
	AEOpenLog               string             `json:"ae_open_log,omitempty"`
	RenderStatus            string             `json:"render_status"`
}

type VerificationDiff struct {
	Path       string `json:"path"`
	Kind       string `json:"kind"`
	Severity   string `json:"severity"`
	ActionType string `json:"action_type"`
	Expected   any    `json:"expected,omitempty"`
	Actual     any    `json:"actual,omitempty"`
}

type Report struct {
	SchemaVersion int          `json:"schema_version"`
	Source        SourceInfo   `json:"source"`
	Target        TargetInfo   `json:"target"`
	Summary       Summary      `json:"summary"`
	Entries       []Entry      `json:"entries,omitempty"`
	Verification  Verification `json:"verification"`
}
