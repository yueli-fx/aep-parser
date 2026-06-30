package aepmigrate

const SchemaVersion = 1

type VersionLabel string

const (
	VersionUnknown VersionLabel = "unknown"
	VersionAE2020  VersionLabel = "AE2020"
	VersionAE2022  VersionLabel = "AE2022"
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
	ProfileDiffStatus string `json:"profile_diff_status"`
	AEOpenStatus      string `json:"ae_open_status"`
	RenderStatus      string `json:"render_status"`
}

type Report struct {
	SchemaVersion int          `json:"schema_version"`
	Source        SourceInfo   `json:"source"`
	Target        TargetInfo   `json:"target"`
	Summary       Summary      `json:"summary"`
	Entries       []Entry      `json:"entries,omitempty"`
	Verification  Verification `json:"verification"`
}
