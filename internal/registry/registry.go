package registry

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	StatusPass = "pass"
	StatusFail = "fail"

	SeverityError   = "error"
	SeverityWarning = "warning"
)

type Registry struct {
	Locations         []Location
	Workflows         []Workflow
	CapabilityAtoms   []CapabilityAtom
	EvidenceSets      []EvidenceSet
	VersionBoundaries []VersionBoundary
}

type Location struct {
	ID        string `json:"id"`
	Path      string `json:"path"`
	Class     string `json:"class"`
	Tracked   bool   `json:"tracked"`
	Required  bool   `json:"required"`
	Lifecycle string `json:"lifecycle"`
}

type Workflow struct {
	ID            string `json:"id"`
	Summary       string `json:"summary"`
	CrossPlatform string `json:"cross_platform"`
}

type CapabilityAtom struct {
	ID           string       `json:"id"`
	Domain       string       `json:"domain"`
	Tier         string       `json:"tier"`
	Status       string       `json:"status"`
	Platform     Platform     `json:"platform"`
	VersionAxis  VersionAxis  `json:"version_axis"`
	Workflows    []string     `json:"workflows"`
	Dependencies []Dependency `json:"dependencies"`
	Boundaries   []string     `json:"boundaries,omitempty"`
}

type Platform struct {
	HostRequired bool     `json:"host_required"`
	OS           []string `json:"os"`
}

type VersionAxis struct {
	MinSupported    string   `json:"min_supported"`
	KnownSupported  []string `json:"known_supported"`
	ExpansionPolicy string   `json:"expansion_policy"`
}

type Dependency struct {
	Kind     string `json:"kind"`
	Path     string `json:"path"`
	Required bool   `json:"required"`
}

type EvidenceSet struct {
	ID           string   `json:"id"`
	Class        string   `json:"class"`
	ArtifactPath string   `json:"artifact_path"`
	Required     bool     `json:"required"`
	Workflows    []string `json:"workflows"`
}

type VersionBoundary struct {
	ID                     string                        `json:"id"`
	AtomID                 string                        `json:"atom_id"`
	Recipe                 string                        `json:"recipe"`
	Feature                string                        `json:"feature"`
	Policy                 string                        `json:"policy"`
	CoverageBoundaryStatus string                        `json:"coverage_boundary_status,omitempty"`
	SourceContract         VersionBoundarySourceContract `json:"source_contract"`
	TargetContract         VersionBoundaryTargetContract `json:"target_contract"`
	ExpectedCells          CoverageTotals                `json:"expected_cells"`
	CellPolicy             []VersionBoundaryCellPolicy   `json:"cell_policy,omitempty"`
	Evidence               []Dependency                  `json:"evidence"`
	CleanupPolicy          string                        `json:"cleanup_policy,omitempty"`
	ExpansionPolicy        string                        `json:"expansion_policy,omitempty"`
	Notes                  []string                      `json:"notes,omitempty"`
}

type VersionBoundarySourceContract struct {
	MinSourceVersion          string   `json:"min_source_version"`
	AvailableSourceVersions   []string `json:"available_source_versions"`
	UnavailableSourceVersions []string `json:"unavailable_source_versions"`
}

type VersionBoundaryTargetContract struct {
	SupportedTargets        []string `json:"supported_targets"`
	BlockedDowngradeTargets []string `json:"blocked_downgrade_targets"`
}

type VersionBoundaryCellPolicy struct {
	SourceVersions []string `json:"source_versions"`
	TargetVersions []string `json:"target_versions"`
	Status         string   `json:"status"`
	Reason         string   `json:"reason,omitempty"`
}

type AuditReport struct {
	SchemaVersion int          `json:"schema_version"`
	Status        string       `json:"status"`
	Summary       AuditSummary `json:"summary"`
	Issues        []AuditIssue `json:"issues"`
}

type AuditSummary struct {
	Locations         int `json:"locations"`
	Workflows         int `json:"workflows"`
	CapabilityAtoms   int `json:"capability_atoms"`
	EvidenceSets      int `json:"evidence_sets"`
	VersionBoundaries int `json:"version_boundaries"`
	Errors            int `json:"errors"`
	Warnings          int `json:"warnings"`
}

type AuditIssue struct {
	Code       string `json:"code"`
	Severity   string `json:"severity"`
	Path       string `json:"path,omitempty"`
	AtomID     string `json:"atom_id,omitempty"`
	BoundaryID string `json:"boundary_id,omitempty"`
	Message    string `json:"message"`
}

type locationsFile struct {
	SchemaVersion int        `json:"schema_version"`
	Locations     []Location `json:"locations"`
}

type workflowsFile struct {
	SchemaVersion int        `json:"schema_version"`
	Workflows     []Workflow `json:"workflows"`
}

type atomsFile struct {
	SchemaVersion   int              `json:"schema_version"`
	CapabilityAtoms []CapabilityAtom `json:"capability_atoms"`
}

type evidenceFile struct {
	SchemaVersion int           `json:"schema_version"`
	EvidenceSets  []EvidenceSet `json:"evidence_sets"`
}

type versionBoundariesFile struct {
	SchemaVersion     int               `json:"schema_version"`
	VersionBoundaries []VersionBoundary `json:"version_boundaries"`
}

func AuditRepository(root string) (AuditReport, error) {
	reg, err := Load(root)
	if err != nil {
		return AuditReport{}, err
	}
	return Audit(root, reg), nil
}

func Load(root string) (Registry, error) {
	var (
		locations  locationsFile
		workflows  workflowsFile
		atoms      atomsFile
		evidence   evidenceFile
		boundaries versionBoundariesFile
	)
	if err := readJSON(root, "registry/locations.json", &locations); err != nil {
		return Registry{}, err
	}
	if err := readJSON(root, "registry/workflows.json", &workflows); err != nil {
		return Registry{}, err
	}
	if err := readJSON(root, "registry/capability_atoms.json", &atoms); err != nil {
		return Registry{}, err
	}
	if err := readJSON(root, "registry/evidence.json", &evidence); err != nil {
		return Registry{}, err
	}
	if err := readOptionalJSON(root, "registry/version_boundaries.json", &boundaries); err != nil {
		return Registry{}, err
	}
	return Registry{
		Locations:         locations.Locations,
		Workflows:         workflows.Workflows,
		CapabilityAtoms:   atoms.CapabilityAtoms,
		EvidenceSets:      evidence.EvidenceSets,
		VersionBoundaries: boundaries.VersionBoundaries,
	}, nil
}

func Audit(root string, reg Registry) AuditReport {
	report := AuditReport{
		SchemaVersion: 1,
		Status:        StatusPass,
		Summary: AuditSummary{
			Locations:         len(reg.Locations),
			Workflows:         len(reg.Workflows),
			CapabilityAtoms:   len(reg.CapabilityAtoms),
			EvidenceSets:      len(reg.EvidenceSets),
			VersionBoundaries: len(reg.VersionBoundaries),
		},
	}

	workflowIDs := map[string]bool{}
	for _, workflow := range reg.Workflows {
		if workflow.ID == "" {
			report.addIssue("missing_workflow_id", SeverityError, "", "", "workflow id is required")
			continue
		}
		if workflowIDs[workflow.ID] {
			report.addIssue("duplicate_workflow", SeverityError, workflow.ID, "", "workflow id must be unique")
			continue
		}
		workflowIDs[workflow.ID] = true
	}

	locationIDs := map[string]bool{}
	for _, location := range reg.Locations {
		if location.ID == "" {
			report.addIssue("missing_location_id", SeverityError, location.Path, "", "location id is required")
		} else if locationIDs[location.ID] {
			report.addIssue("duplicate_location", SeverityError, location.Path, "", "location id must be unique")
		}
		locationIDs[location.ID] = true
		if location.Path == "" {
			report.addIssue("missing_location_path", SeverityError, location.ID, "", "location path is required")
			continue
		}
		if !relPathOK(location.Path) {
			report.addIssue("invalid_location_path", SeverityError, location.Path, "", "location path must be repository relative")
			continue
		}
		if !exists(root, location.Path) {
			severity := SeverityWarning
			code := "missing_optional_location"
			if location.Required {
				severity = SeverityError
				code = "missing_location"
			}
			report.addIssue(code, severity, location.Path, "", "registered location does not exist")
		}
	}

	coverage := newLocationCoverage(reg.Locations)

	atomIDs := map[string]bool{}
	for _, atom := range reg.CapabilityAtoms {
		if atom.ID == "" {
			report.addIssue("missing_atom_id", SeverityError, "", "", "capability atom id is required")
			continue
		}
		if atomIDs[atom.ID] {
			report.addIssue("duplicate_atom", SeverityError, "", atom.ID, "capability atom id must be unique")
		}
		atomIDs[atom.ID] = true
		auditAtom(root, atom, workflowIDs, coverage, &report)
	}

	for _, evidence := range reg.EvidenceSets {
		if evidence.ID == "" {
			report.addIssue("missing_evidence_id", SeverityError, evidence.ArtifactPath, "", "evidence id is required")
		}
		for _, workflow := range evidence.Workflows {
			if !workflowIDs[workflow] {
				report.addIssue("unknown_workflow", SeverityError, workflow, "", "evidence references an unknown workflow")
			}
		}
		if evidence.ArtifactPath == "" {
			report.addIssue("missing_evidence_path", SeverityError, evidence.ID, "", "evidence artifact_path is required")
			continue
		}
		if !relPathOK(evidence.ArtifactPath) {
			report.addIssue("invalid_evidence_path", SeverityError, evidence.ArtifactPath, "", "evidence artifact_path must be repository relative")
			continue
		}
		if !coverage.covers(evidence.ArtifactPath) {
			report.addIssue("unregistered_evidence_location", SeverityError, evidence.ArtifactPath, "", "evidence artifact_path is not covered by any registered location")
		}
		if !exists(root, evidence.ArtifactPath) {
			severity := SeverityWarning
			code := "missing_optional_evidence"
			if evidence.Required {
				severity = SeverityError
				code = "missing_evidence"
			}
			report.addIssue(code, severity, evidence.ArtifactPath, "", "registered evidence artifact does not exist")
		}
	}

	boundaryIDs := map[string]bool{}
	for _, boundary := range reg.VersionBoundaries {
		auditVersionBoundary(root, boundary, boundaryIDs, atomIDs, coverage, &report)
	}

	report.finish()
	return report
}

func auditAtom(root string, atom CapabilityAtom, workflowIDs map[string]bool, coverage locationCoverage, report *AuditReport) {
	if atom.Domain == "" {
		report.addIssue("missing_atom_domain", SeverityError, "", atom.ID, "capability atom domain is required")
	}
	if atom.Tier == "" {
		report.addIssue("missing_atom_tier", SeverityError, "", atom.ID, "capability atom tier is required")
	}
	if atom.Status == "" {
		report.addIssue("missing_atom_status", SeverityError, "", atom.ID, "capability atom status is required")
	}
	if len(atom.Platform.OS) == 0 {
		report.addIssue("missing_platform_os", SeverityError, "", atom.ID, "capability atom platform.os is required")
	}
	if atom.VersionAxis.MinSupported == "" {
		report.addIssue("missing_min_supported_version", SeverityError, "", atom.ID, "capability atom version_axis.min_supported is required")
	}
	if len(atom.VersionAxis.KnownSupported) == 0 {
		report.addIssue("missing_known_supported_versions", SeverityError, "", atom.ID, "capability atom version_axis.known_supported is required")
	}
	if atom.VersionAxis.ExpansionPolicy == "" {
		report.addIssue("missing_version_expansion_policy", SeverityError, "", atom.ID, "capability atom version_axis.expansion_policy is required")
	}
	for _, workflow := range atom.Workflows {
		if !workflowIDs[workflow] {
			report.addIssue("unknown_workflow", SeverityError, workflow, atom.ID, "capability atom references an unknown workflow")
		}
	}
	for _, dep := range atom.Dependencies {
		if dep.Path == "" {
			report.addIssue("missing_dependency_path", SeverityError, "", atom.ID, "dependency path is required")
			continue
		}
		if !relPathOK(dep.Path) {
			report.addIssue("invalid_dependency_path", SeverityError, dep.Path, atom.ID, "dependency path must be repository relative")
			continue
		}
		if !coverage.covers(dep.Path) {
			report.addIssue("unregistered_dependency_location", SeverityError, dep.Path, atom.ID, "dependency path is not covered by any registered location")
		}
		if dependencyExists(root, dep) {
			continue
		}
		severity := SeverityWarning
		code := "missing_optional_dependency"
		if dep.Required {
			severity = SeverityError
			code = "missing_dependency"
		}
		report.addIssue(code, severity, dep.Path, atom.ID, fmt.Sprintf("%s dependency does not exist", valueOr(dep.Kind, "registered")))
	}
}

func auditVersionBoundary(root string, boundary VersionBoundary, boundaryIDs, atomIDs map[string]bool, coverage locationCoverage, report *AuditReport) {
	if boundary.ID == "" {
		report.addBoundaryIssue("missing_version_boundary_id", SeverityError, "", boundary.AtomID, "", "version boundary id is required")
		return
	}
	if boundaryIDs[boundary.ID] {
		report.addBoundaryIssue("duplicate_version_boundary", SeverityError, "", boundary.AtomID, boundary.ID, "version boundary id must be unique")
	}
	boundaryIDs[boundary.ID] = true
	if boundary.AtomID == "" {
		report.addBoundaryIssue("missing_version_boundary_atom", SeverityError, "", "", boundary.ID, "version boundary atom_id is required")
	} else if !atomIDs[boundary.AtomID] {
		report.addBoundaryIssue("unknown_version_boundary_atom", SeverityError, "", boundary.AtomID, boundary.ID, "version boundary references an unknown capability atom")
	}
	if boundary.Recipe == "" {
		report.addBoundaryIssue("missing_version_boundary_recipe", SeverityError, "", boundary.AtomID, boundary.ID, "version boundary recipe is required")
	}
	if boundary.Feature == "" {
		report.addBoundaryIssue("missing_version_boundary_feature", SeverityError, "", boundary.AtomID, boundary.ID, "version boundary feature is required")
	}
	if boundary.Policy == "" {
		report.addBoundaryIssue("missing_version_boundary_policy", SeverityError, "", boundary.AtomID, boundary.ID, "version boundary policy is required")
	}
	if boundary.SourceContract.MinSourceVersion == "" {
		report.addBoundaryIssue("missing_version_boundary_min_source", SeverityError, "", boundary.AtomID, boundary.ID, "version boundary source_contract.min_source_version is required")
	}
	if len(boundary.SourceContract.AvailableSourceVersions) == 0 {
		report.addBoundaryIssue("missing_version_boundary_available_sources", SeverityError, "", boundary.AtomID, boundary.ID, "version boundary source_contract.available_source_versions is required")
	}
	if len(boundary.TargetContract.SupportedTargets) == 0 {
		report.addBoundaryIssue("missing_version_boundary_supported_targets", SeverityError, "", boundary.AtomID, boundary.ID, "version boundary target_contract.supported_targets is required")
	}
	for _, dep := range boundary.Evidence {
		auditBoundaryDependency(root, boundary, dep, coverage, report)
	}
	for _, policy := range boundary.CellPolicy {
		auditVersionBoundaryCellPolicy(boundary, policy, report)
	}
}

func auditVersionBoundaryCellPolicy(boundary VersionBoundary, policy VersionBoundaryCellPolicy, report *AuditReport) {
	if len(policy.SourceVersions) == 0 {
		report.addBoundaryIssue("missing_version_boundary_cell_policy_sources", SeverityError, "", boundary.AtomID, boundary.ID, "version boundary cell_policy.source_versions is required")
	}
	if len(policy.TargetVersions) == 0 {
		report.addBoundaryIssue("missing_version_boundary_cell_policy_targets", SeverityError, "", boundary.AtomID, boundary.ID, "version boundary cell_policy.target_versions is required")
	}
	if policy.Status == "" {
		report.addBoundaryIssue("missing_version_boundary_cell_policy_status", SeverityError, "", boundary.AtomID, boundary.ID, "version boundary cell_policy.status is required")
	}
}

func auditBoundaryDependency(root string, boundary VersionBoundary, dep Dependency, coverage locationCoverage, report *AuditReport) {
	if dep.Path == "" {
		report.addBoundaryIssue("missing_version_boundary_evidence_path", SeverityError, "", boundary.AtomID, boundary.ID, "version boundary evidence path is required")
		return
	}
	if !relPathOK(dep.Path) {
		report.addBoundaryIssue("invalid_version_boundary_evidence_path", SeverityError, dep.Path, boundary.AtomID, boundary.ID, "version boundary evidence path must be repository relative")
		return
	}
	if !coverage.covers(dep.Path) {
		report.addBoundaryIssue("unregistered_version_boundary_evidence_location", SeverityError, dep.Path, boundary.AtomID, boundary.ID, "version boundary evidence path is not covered by any registered location")
	}
	if dependencyExists(root, dep) {
		return
	}
	if dep.Required {
		report.addBoundaryIssue("missing_version_boundary_evidence", SeverityError, dep.Path, boundary.AtomID, boundary.ID, fmt.Sprintf("%s evidence does not exist", valueOr(dep.Kind, "registered")))
	}
}

func (r *AuditReport) addIssue(code, severity, path, atomID, message string) {
	r.Issues = append(r.Issues, AuditIssue{
		Code:     code,
		Severity: severity,
		Path:     path,
		AtomID:   atomID,
		Message:  message,
	})
	if severity == SeverityError {
		r.Summary.Errors++
		return
	}
	if severity == SeverityWarning {
		r.Summary.Warnings++
	}
}

func (r *AuditReport) addBoundaryIssue(code, severity, path, atomID, boundaryID, message string) {
	r.Issues = append(r.Issues, AuditIssue{
		Code:       code,
		Severity:   severity,
		Path:       path,
		AtomID:     atomID,
		BoundaryID: boundaryID,
		Message:    message,
	})
	if severity == SeverityError {
		r.Summary.Errors++
		return
	}
	if severity == SeverityWarning {
		r.Summary.Warnings++
	}
}

func (r *AuditReport) finish() {
	if r.Summary.Errors > 0 {
		r.Status = StatusFail
		return
	}
	r.Status = StatusPass
}

func readJSON(root, rel string, v any) error {
	path := filepath.Join(root, filepath.FromSlash(rel))
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", rel, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("decode %s: %w", rel, err)
	}
	return nil
}

func readOptionalJSON(root, rel string, v any) error {
	path := filepath.Join(root, filepath.FromSlash(rel))
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read %s: %w", rel, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("decode %s: %w", rel, err)
	}
	return nil
}

func exists(root, rel string) bool {
	path := filepath.Join(root, filepath.FromSlash(rel))
	_, err := os.Stat(path)
	return err == nil
}

func dependencyExists(root string, dep Dependency) bool {
	if dep.Kind != "glob" {
		return exists(root, dep.Path)
	}
	matches, err := dependencyGlobMatches(root, dep.Path)
	return err == nil && len(matches) > 0
}

func dependencyGlobMatches(root, pattern string) ([]string, error) {
	fullPattern := filepath.Join(root, filepath.FromSlash(pattern))
	matches, err := filepath.Glob(fullPattern)
	if err != nil {
		return nil, err
	}
	var relMatches []string
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil || info.IsDir() {
			continue
		}
		rel, err := filepath.Rel(root, match)
		if err != nil {
			continue
		}
		relMatches = append(relMatches, cleanRel(rel))
	}
	return relMatches, nil
}

func relPathOK(path string) bool {
	return path != "" && !filepath.IsAbs(path) && filepath.Clean(path) != ".." && !startsWithDotDot(filepath.Clean(path))
}

func startsWithDotDot(path string) bool {
	return len(path) >= 3 && path[:3] == ".."+string(filepath.Separator)
}

func valueOr(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

type locationCoverage []string

func newLocationCoverage(locations []Location) locationCoverage {
	coverage := make(locationCoverage, 0, len(locations))
	for _, location := range locations {
		if relPathOK(location.Path) {
			coverage = append(coverage, cleanRel(location.Path))
		}
	}
	return coverage
}

func (c locationCoverage) covers(path string) bool {
	clean := cleanRel(path)
	for _, root := range c {
		if clean == root || hasPathPrefix(clean, root) {
			return true
		}
	}
	return false
}

func cleanRel(path string) string {
	return filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
}

func hasPathPrefix(path, prefix string) bool {
	return len(path) > len(prefix) && path[:len(prefix)] == prefix && path[len(prefix)] == '/'
}
