package registry

import (
	"sort"
)

var DefaultAEVersionAxis = []string{"AE2020", "AE2021", "AE2022", "AE2023", "AE2024", "AE2025"}

type CoverageAxisReport struct {
	SchemaVersion int                       `json:"schema_version"`
	Status        string                    `json:"status"`
	VersionAxis   []string                  `json:"version_axis"`
	Filter        CoverageAxisFilter        `json:"filter"`
	Summary       CoverageAxisSummary       `json:"summary"`
	ByVersionPair []CoverageAxisPairSummary `json:"by_version_pair"`
	Rows          []CoverageAxisRow         `json:"rows"`
}

type CoverageAxisFilter struct {
	RecordID              string `json:"record_id,omitempty"`
	AtomID                string `json:"atom_id,omitempty"`
	Recipe                string `json:"recipe,omitempty"`
	CaseStatus            string `json:"case_status,omitempty"`
	WriterStatus          string `json:"writer_status,omitempty"`
	HostOpenEvidenceLevel string `json:"host_open_evidence_level,omitempty"`
	BoundaryStatus        string `json:"boundary_status,omitempty"`
}

type CoverageAxisSummary struct {
	AtomRows              int `json:"atom_rows"`
	WriterFullAxisRows    int `json:"writer_full_axis_rows"`
	WriterPartialRows     int `json:"writer_partial_rows"`
	BoundaryRows          int `json:"boundary_rows"`
	HostFullAxisRows      int `json:"host_full_axis_rows"`
	HostDirectOnlyRows    int `json:"host_direct_only_rows"`
	HostInferredRows      int `json:"host_inferred_rows"`
	HostBoundaryRows      int `json:"host_boundary_rows"`
	HostMissingRows       int `json:"host_missing_rows"`
	SourceTargetPairs     int `json:"source_target_pairs"`
	SourceTargetPairsFull int `json:"source_target_pairs_full"`
	Cells                 int `json:"cells"`
	Pass                  int `json:"pass"`
	Blocked               int `json:"blocked"`
	Failed                int `json:"failed"`
	Skipped               int `json:"skipped"`
}

type CoverageAxisRow struct {
	AtomID                string         `json:"atom_id"`
	RecordID              string         `json:"record_id"`
	Recipe                string         `json:"recipe,omitempty"`
	WriterStatus          string         `json:"writer_status,omitempty"`
	BoundaryStatus        string         `json:"boundary_status,omitempty"`
	BoundaryReason        string         `json:"boundary_reason,omitempty"`
	HostOpenEvidenceLevel string         `json:"host_open_evidence_level,omitempty"`
	WriterAxisStatus      string         `json:"writer_axis_status"`
	SourceVersions        []string       `json:"source_versions,omitempty"`
	TargetVersions        []string       `json:"target_versions,omitempty"`
	MissingSourceVersions []string       `json:"missing_source_versions,omitempty"`
	MissingTargetVersions []string       `json:"missing_target_versions,omitempty"`
	HostAxisStatus        string         `json:"host_axis_status"`
	DirectHostVersions    []string       `json:"direct_host_versions,omitempty"`
	InferredHostVersions  []string       `json:"inferred_host_versions,omitempty"`
	MissingHostVersions   []string       `json:"missing_host_versions,omitempty"`
	Totals                CoverageTotals `json:"totals"`
}

type CoverageAxisPairSummary struct {
	SourceVersion string `json:"source_version"`
	TargetVersion string `json:"target_version"`
	Total         int    `json:"total"`
	Pass          int    `json:"pass"`
	Blocked       int    `json:"blocked"`
	Failed        int    `json:"failed"`
	Skipped       int    `json:"skipped"`
}

func CoverageAxis(root, coveragePath string, versionAxis []string) (CoverageAxisReport, error) {
	return CoverageAxisWithFilter(root, coveragePath, versionAxis, CoverageAxisFilter{})
}

func CoverageAxisWithFilter(root, coveragePath string, versionAxis []string, filter CoverageAxisFilter) (CoverageAxisReport, error) {
	if len(versionAxis) == 0 {
		versionAxis = DefaultAEVersionAxis
	}
	versionAxis = sortedStrings(versionAxis)
	report, err := ValidateCoverage(root, coveragePath)
	if err != nil {
		return CoverageAxisReport{}, err
	}
	cells, err := CoverageCells(root, coveragePath, CoverageCellFilter{
		RecordID:              filter.RecordID,
		AtomID:                filter.AtomID,
		Recipe:                filter.Recipe,
		CaseStatus:            filter.CaseStatus,
		WriterStatus:          filter.WriterStatus,
		HostOpenEvidenceLevel: filter.HostOpenEvidenceLevel,
		BoundaryStatus:        filter.BoundaryStatus,
	})
	if err != nil {
		return CoverageAxisReport{}, err
	}
	rowKeysFromCells := map[string]bool{}
	if filter.CaseStatus != "" {
		for _, cell := range cells.Cells {
			rowKeysFromCells[coverageAxisIdentity(cell.AtomID, cell.RecordID, cell.Recipe)] = true
		}
	}
	axis := CoverageAxisReport{
		SchemaVersion: 1,
		Status:        report.Status,
		VersionAxis:   append([]string(nil), versionAxis...),
		Filter:        filter,
	}
	for _, row := range report.AtomRows {
		if !coverageAxisRowMatches(row, filter) {
			continue
		}
		if filter.CaseStatus != "" && !rowKeysFromCells[coverageAxisIdentity(row.AtomID, row.RecordID, row.Recipe)] {
			continue
		}
		axisRow := coverageAxisRow(row, versionAxis)
		axis.Rows = append(axis.Rows, axisRow)
		axis.Summary.AtomRows++
		switch axisRow.WriterAxisStatus {
		case "full_source_target_axis":
			axis.Summary.WriterFullAxisRows++
		case "partial_source_target_axis":
			axis.Summary.WriterPartialRows++
		case "boundary_source_contract":
			axis.Summary.BoundaryRows++
		}
		switch axisRow.HostAxisStatus {
		case "full_direct_or_inferred_axis", "full_direct_all_hosts_axis":
			axis.Summary.HostFullAxisRows++
		case "direct_hosts_only":
			axis.Summary.HostDirectOnlyRows++
		case "direct_and_inferred_hosts":
			axis.Summary.HostInferredRows++
		case "excluded_known_boundary":
			axis.Summary.HostBoundaryRows++
		default:
			axis.Summary.HostMissingRows++
		}
	}
	axis.ByVersionPair = coverageAxisPairSummaries(cells.Cells, versionAxis)
	axis.Summary.SourceTargetPairs = len(axis.ByVersionPair)
	for _, pair := range axis.ByVersionPair {
		if pair.Total > 0 {
			axis.Summary.SourceTargetPairsFull++
		}
		axis.Summary.Cells += pair.Total
		axis.Summary.Pass += pair.Pass
		axis.Summary.Blocked += pair.Blocked
		axis.Summary.Failed += pair.Failed
		axis.Summary.Skipped += pair.Skipped
	}
	sort.Slice(axis.Rows, func(i, j int) bool {
		if axis.Rows[i].AtomID != axis.Rows[j].AtomID {
			return axis.Rows[i].AtomID < axis.Rows[j].AtomID
		}
		if axis.Rows[i].RecordID != axis.Rows[j].RecordID {
			return axis.Rows[i].RecordID < axis.Rows[j].RecordID
		}
		return axis.Rows[i].Recipe < axis.Rows[j].Recipe
	})
	return axis, nil
}

func coverageAxisIdentity(atomID, recordID, recipe string) string {
	return atomID + "\x00" + recordID + "\x00" + recipe
}

func coverageAxisRowMatches(row AtomCoverageRow, filter CoverageAxisFilter) bool {
	if filter.RecordID != "" && row.RecordID != filter.RecordID {
		return false
	}
	if filter.AtomID != "" && row.AtomID != filter.AtomID {
		return false
	}
	if filter.Recipe != "" && row.Recipe != filter.Recipe {
		return false
	}
	if filter.WriterStatus != "" && row.WriterStatus != filter.WriterStatus {
		return false
	}
	if filter.HostOpenEvidenceLevel != "" && row.HostOpenEvidenceLevel != filter.HostOpenEvidenceLevel {
		return false
	}
	if filter.BoundaryStatus != "" && row.BoundaryStatus != filter.BoundaryStatus {
		return false
	}
	return true
}

func coverageAxisRow(row AtomCoverageRow, versionAxis []string) CoverageAxisRow {
	axisRow := CoverageAxisRow{
		AtomID:                row.AtomID,
		RecordID:              row.RecordID,
		Recipe:                row.Recipe,
		WriterStatus:          row.WriterStatus,
		BoundaryStatus:        row.BoundaryStatus,
		BoundaryReason:        row.BoundaryReason,
		HostOpenEvidenceLevel: row.HostOpenEvidenceLevel,
		SourceVersions:        sortedStrings(row.SourceVersions),
		TargetVersions:        sortedStrings(row.TargetVersions),
		DirectHostVersions:    sortedStrings(row.DirectHostVersions),
		InferredHostVersions:  sortedStrings(row.InferredHostVersions),
		Totals:                row.Totals,
	}
	axisRow.MissingSourceVersions = missingVersions(versionAxis, axisRow.SourceVersions)
	axisRow.MissingTargetVersions = missingVersions(versionAxis, axisRow.TargetVersions)
	if row.BoundaryStatus != "" {
		axisRow.WriterAxisStatus = "boundary_source_contract"
	} else if len(axisRow.MissingSourceVersions) == 0 && len(axisRow.MissingTargetVersions) == 0 {
		axisRow.WriterAxisStatus = "full_source_target_axis"
	} else {
		axisRow.WriterAxisStatus = "partial_source_target_axis"
	}
	switch {
	case row.HostOpenEvidenceLevel == "direct_all_hosts_representative" || row.HostOpenEvidenceLevel == "direct_all_hosts_record":
		axisRow.HostAxisStatus = "full_direct_all_hosts_axis"
	case row.HostOpenEvidenceLevel == "excluded_known_boundary":
		axisRow.HostAxisStatus = "excluded_known_boundary"
	default:
		hostVersions := append([]string{}, axisRow.DirectHostVersions...)
		hostVersions = append(hostVersions, axisRow.InferredHostVersions...)
		hostVersions = sortedUniqueStrings(hostVersions)
		axisRow.MissingHostVersions = missingVersions(versionAxis, hostVersions)
		switch {
		case len(hostVersions) > 0 && len(axisRow.MissingHostVersions) == 0:
			axisRow.HostAxisStatus = "full_direct_or_inferred_axis"
		case len(axisRow.DirectHostVersions) > 0 && len(axisRow.InferredHostVersions) > 0:
			axisRow.HostAxisStatus = "direct_and_inferred_hosts"
		case len(axisRow.DirectHostVersions) > 0:
			axisRow.HostAxisStatus = "direct_hosts_only"
		default:
			axisRow.HostAxisStatus = "missing_host_axis"
		}
	}
	return axisRow
}

func coverageAxisPairSummaries(cells []CoverageCell, versionAxis []string) []CoverageAxisPairSummary {
	pairs := map[string]*CoverageAxisPairSummary{}
	for _, source := range versionAxis {
		for _, target := range versionAxis {
			key := source + "\x00" + target
			pairs[key] = &CoverageAxisPairSummary{
				SourceVersion: source,
				TargetVersion: target,
			}
		}
	}
	for _, cell := range cells {
		if cell.SourceVersion == "" || cell.TargetVersion == "" {
			continue
		}
		key := cell.SourceVersion + "\x00" + cell.TargetVersion
		pair := pairs[key]
		if pair == nil {
			pair = &CoverageAxisPairSummary{
				SourceVersion: cell.SourceVersion,
				TargetVersion: cell.TargetVersion,
			}
			pairs[key] = pair
		}
		pair.Total++
		switch cell.Status {
		case "pass":
			pair.Pass++
		case "blocked":
			pair.Blocked++
		case "failed":
			pair.Failed++
		case "skipped":
			pair.Skipped++
		}
	}
	keys := make([]string, 0, len(pairs))
	for key := range pairs {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]CoverageAxisPairSummary, 0, len(keys))
	for _, key := range keys {
		out = append(out, *pairs[key])
	}
	return out
}

func missingVersions(axis, seen []string) []string {
	seenSet := stringSet(seen)
	var missing []string
	for _, version := range axis {
		if !seenSet[version] {
			missing = append(missing, version)
		}
	}
	return missing
}

func sortedUniqueStrings(values []string) []string {
	set := map[string]bool{}
	for _, value := range values {
		if value != "" {
			set[value] = true
		}
	}
	return sortedKeys(set)
}
