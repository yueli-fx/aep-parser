package registry

import "fmt"

type VersionBoundaryCheckReport struct {
	SchemaVersion int                         `json:"schema_version"`
	Status        string                      `json:"status"`
	CoveragePath  string                      `json:"coverage_path"`
	VersionAxis   []string                    `json:"version_axis"`
	Summary       VersionBoundaryCheckSummary `json:"summary"`
	Boundaries    []VersionBoundaryCheck      `json:"boundaries"`
	Issues        []VersionBoundaryCheckIssue `json:"issues"`
}

type VersionBoundaryCheckSummary struct {
	Boundaries       int `json:"boundaries"`
	Matched          int `json:"matched"`
	MissingRows      int `json:"missing_rows"`
	DuplicateRows    int `json:"duplicate_rows"`
	MismatchedTotals int `json:"mismatched_totals"`
	MismatchedStatus int `json:"mismatched_status"`
	CheckedCells     int `json:"checked_cells"`
	MissingCells     int `json:"missing_cells"`
	MismatchedCells  int `json:"mismatched_cells"`
	Errors           int `json:"errors"`
}

type VersionBoundaryCheck struct {
	BoundaryID             string                        `json:"boundary_id"`
	AtomID                 string                        `json:"atom_id"`
	Recipe                 string                        `json:"recipe"`
	Policy                 string                        `json:"policy"`
	ExpectedBoundaryStatus string                        `json:"expected_boundary_status,omitempty"`
	ActualBoundaryStatus   string                        `json:"actual_boundary_status,omitempty"`
	WriterAxisStatus       string                        `json:"writer_axis_status,omitempty"`
	ExpectedCells          CoverageTotals                `json:"expected_cells"`
	ActualCells            CoverageTotals                `json:"actual_cells"`
	CellPolicy             []VersionBoundaryCellPolicy   `json:"cell_policy,omitempty"`
	CheckedCells           int                           `json:"checked_cells"`
	CellMismatches         []VersionBoundaryCellMismatch `json:"cell_mismatches,omitempty"`
	Match                  bool                          `json:"match"`
}

type VersionBoundaryCellMismatch struct {
	SourceVersion  string `json:"source_version"`
	TargetVersion  string `json:"target_version"`
	ExpectedStatus string `json:"expected_status,omitempty"`
	ActualStatus   string `json:"actual_status,omitempty"`
	Code           string `json:"code"`
	Message        string `json:"message"`
}

type VersionBoundaryCheckIssue struct {
	Code       string `json:"code"`
	Severity   string `json:"severity"`
	BoundaryID string `json:"boundary_id,omitempty"`
	AtomID     string `json:"atom_id,omitempty"`
	Recipe     string `json:"recipe,omitempty"`
	Message    string `json:"message"`
}

func CheckVersionBoundaries(root, coveragePath string, versionAxis []string) (VersionBoundaryCheckReport, error) {
	reg, err := Load(root)
	if err != nil {
		return VersionBoundaryCheckReport{}, err
	}
	if len(versionAxis) == 0 {
		versionAxis = DefaultAEVersionAxis()
	}
	report := VersionBoundaryCheckReport{
		SchemaVersion: 1,
		Status:        StatusPass,
		CoveragePath:  coveragePath,
		VersionAxis:   append([]string(nil), sortedStrings(versionAxis)...),
		Summary: VersionBoundaryCheckSummary{
			Boundaries: len(reg.VersionBoundaries),
		},
	}
	axis, err := coverageAxisWithFilter(root, coveragePath, report.VersionAxis, CoverageAxisFilter{}, false)
	if err != nil {
		return VersionBoundaryCheckReport{}, err
	}
	rowsByBoundary := map[string][]CoverageAxisRow{}
	for _, row := range axis.Rows {
		key := versionBoundaryRowKey(row.AtomID, row.Recipe)
		rowsByBoundary[key] = append(rowsByBoundary[key], row)
	}
	cells, err := coverageCells(root, coveragePath, CoverageCellFilter{}, false)
	if err != nil {
		return VersionBoundaryCheckReport{}, err
	}
	cellsByBoundary := map[string]CoverageCell{}
	for _, cell := range cells.Cells {
		key := versionBoundaryCellKey(cell.AtomID, cell.Recipe, cell.SourceVersion, cell.TargetVersion)
		cellsByBoundary[key] = cell
	}
	for _, boundary := range reg.VersionBoundaries {
		check, issues := checkVersionBoundary(boundary, rowsByBoundary[versionBoundaryRowKey(boundary.AtomID, boundary.Recipe)], cellsByBoundary)
		report.Boundaries = append(report.Boundaries, check)
		report.Summary.CheckedCells += check.CheckedCells
		if check.Match {
			report.Summary.Matched++
		}
		for _, issue := range issues {
			report.addBoundaryCheckIssue(issue)
		}
	}
	if report.Summary.Errors > 0 {
		report.Status = StatusFail
	}
	return report, nil
}

func checkVersionBoundary(boundary VersionBoundary, rows []CoverageAxisRow, cells map[string]CoverageCell) (VersionBoundaryCheck, []VersionBoundaryCheckIssue) {
	check := VersionBoundaryCheck{
		BoundaryID:             boundary.ID,
		AtomID:                 boundary.AtomID,
		Recipe:                 boundary.Recipe,
		Policy:                 boundary.Policy,
		ExpectedBoundaryStatus: boundary.CoverageBoundaryStatus,
		ExpectedCells:          boundary.ExpectedCells,
		CellPolicy:             boundary.CellPolicy,
		Match:                  true,
	}
	var issues []VersionBoundaryCheckIssue
	if len(rows) == 0 {
		check.Match = false
		issues = append(issues, versionBoundaryCheckIssue("missing_boundary_coverage_row", boundary, "coverage axis has no row for boundary atom and recipe"))
		return check, issues
	}
	if len(rows) > 1 {
		check.Match = false
		issues = append(issues, versionBoundaryCheckIssue("duplicate_boundary_coverage_rows", boundary, fmt.Sprintf("coverage axis has %d rows for boundary atom and recipe", len(rows))))
		return check, issues
	}
	row := rows[0]
	check.ActualBoundaryStatus = row.BoundaryStatus
	check.WriterAxisStatus = row.WriterAxisStatus
	check.ActualCells = row.Totals
	if boundary.CoverageBoundaryStatus != "" && row.BoundaryStatus != boundary.CoverageBoundaryStatus {
		check.Match = false
		issues = append(issues, versionBoundaryCheckIssue("boundary_status_mismatch", boundary, fmt.Sprintf("coverage boundary status = %q, want %q", row.BoundaryStatus, boundary.CoverageBoundaryStatus)))
	}
	if row.WriterAxisStatus != "boundary_source_contract" {
		check.Match = false
		issues = append(issues, versionBoundaryCheckIssue("writer_axis_status_mismatch", boundary, fmt.Sprintf("writer axis status = %q, want boundary_source_contract", row.WriterAxisStatus)))
	}
	if !coverageTotalsEqual(row.Totals, boundary.ExpectedCells) {
		check.Match = false
		issues = append(issues, versionBoundaryCheckIssue("boundary_totals_mismatch", boundary, fmt.Sprintf("coverage totals = %+v, want %+v", row.Totals, boundary.ExpectedCells)))
	}
	cellIssues := checkVersionBoundaryCells(boundary, cells, &check)
	if len(cellIssues) > 0 {
		check.Match = false
		issues = append(issues, cellIssues...)
	}
	return check, issues
}

func versionBoundaryRowKey(atomID, recipe string) string {
	return atomID + "\x00" + recipe
}

func versionBoundaryCellKey(atomID, recipe, sourceVersion, targetVersion string) string {
	return atomID + "\x00" + recipe + "\x00" + sourceVersion + "\x00" + targetVersion
}

func checkVersionBoundaryCells(boundary VersionBoundary, cells map[string]CoverageCell, check *VersionBoundaryCheck) []VersionBoundaryCheckIssue {
	var issues []VersionBoundaryCheckIssue
	for _, policy := range boundary.CellPolicy {
		for _, source := range policy.SourceVersions {
			for _, target := range policy.TargetVersions {
				check.CheckedCells++
				key := versionBoundaryCellKey(boundary.AtomID, boundary.Recipe, source, target)
				cell, ok := cells[key]
				if !ok {
					mismatch := VersionBoundaryCellMismatch{
						SourceVersion:  source,
						TargetVersion:  target,
						ExpectedStatus: policy.Status,
						Code:           "missing_boundary_cell",
						Message:        "coverage cells have no source-target case for boundary policy",
					}
					check.CellMismatches = append(check.CellMismatches, mismatch)
					issues = append(issues, versionBoundaryCellIssue(boundary, mismatch))
					continue
				}
				if cell.Status != policy.Status {
					mismatch := VersionBoundaryCellMismatch{
						SourceVersion:  source,
						TargetVersion:  target,
						ExpectedStatus: policy.Status,
						ActualStatus:   cell.Status,
						Code:           "boundary_cell_status_mismatch",
						Message:        fmt.Sprintf("coverage cell status = %q, want %q", cell.Status, policy.Status),
					}
					check.CellMismatches = append(check.CellMismatches, mismatch)
					issues = append(issues, versionBoundaryCellIssue(boundary, mismatch))
				}
			}
		}
	}
	return issues
}

func (r *VersionBoundaryCheckReport) addBoundaryCheckIssue(issue VersionBoundaryCheckIssue) {
	r.Issues = append(r.Issues, issue)
	if issue.Severity != SeverityError {
		return
	}
	r.Summary.Errors++
	switch issue.Code {
	case "missing_boundary_coverage_row":
		r.Summary.MissingRows++
	case "duplicate_boundary_coverage_rows":
		r.Summary.DuplicateRows++
	case "boundary_status_mismatch", "writer_axis_status_mismatch":
		r.Summary.MismatchedStatus++
	case "boundary_totals_mismatch":
		r.Summary.MismatchedTotals++
	case "missing_boundary_cell":
		r.Summary.MissingCells++
	case "boundary_cell_status_mismatch":
		r.Summary.MismatchedCells++
	}
}

func versionBoundaryCheckIssue(code string, boundary VersionBoundary, message string) VersionBoundaryCheckIssue {
	return VersionBoundaryCheckIssue{
		Code:       code,
		Severity:   SeverityError,
		BoundaryID: boundary.ID,
		AtomID:     boundary.AtomID,
		Recipe:     boundary.Recipe,
		Message:    message,
	}
}

func versionBoundaryCellIssue(boundary VersionBoundary, mismatch VersionBoundaryCellMismatch) VersionBoundaryCheckIssue {
	return VersionBoundaryCheckIssue{
		Code:       mismatch.Code,
		Severity:   SeverityError,
		BoundaryID: boundary.ID,
		AtomID:     boundary.AtomID,
		Recipe:     boundary.Recipe,
		Message:    fmt.Sprintf("%s source=%s target=%s", mismatch.Message, mismatch.SourceVersion, mismatch.TargetVersion),
	}
}

func coverageTotalsEqual(a, b CoverageTotals) bool {
	return a.Total == b.Total &&
		a.Pass == b.Pass &&
		a.Blocked == b.Blocked &&
		a.Failed == b.Failed &&
		a.Skipped == b.Skipped
}
