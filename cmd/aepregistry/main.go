package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/aeversion"
	"github.com/yueli-fx/aep-parser/internal/registry"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	switch args[0] {
	case "audit":
		return runAudit(args[1:])
	case "boundaries":
		return runBoundaries(args[1:])
	case "coverage":
		return runCoverage(args[1:])
	case "gate":
		return runGate(args[1:])
	case "cleanup":
		return runCleanup(args[1:])
	case "inventory":
		return runInventory(args[1:])
	case "layout":
		return runLayout(args[1:])
	case "ownership":
		return runOwnership(args[1:])
	default:
		usage()
		return 2
	}
}

type gateReport struct {
	SchemaVersion int              `json:"schema_version"`
	Status        string           `json:"status"`
	Root          string           `json:"root"`
	Scope         string           `json:"scope"`
	CoveragePath  string           `json:"coverage_path"`
	VersionAxis   []string         `json:"version_axis,omitempty"`
	Summary       gateSummary      `json:"summary"`
	Steps         []gateStepReport `json:"steps"`
}

type gateSummary struct {
	Steps  int `json:"steps"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
	Errors int `json:"errors"`
}

type gateStepReport struct {
	ID      string `json:"id"`
	Command string `json:"command"`
	Output  string `json:"output"`
	Status  string `json:"status"`
	Errors  int    `json:"errors,omitempty"`
}

func runGate(args []string) int {
	fs := flag.NewFlagSet("aepregistry gate", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repository root")
	coveragePath := fs.String("coverage", "flightdeck/work/aep-understanding-generation/versioned-aep-migration-coverage.json", "coverage ledger JSON path")
	outPath := fs.String("out", "tmp/registry_gate.json", "ordered registry gate report JSON path")
	scope := fs.String("scope", "version-matrix", "gate scope: version-matrix, asset-policy, or all")
	versions := fs.String("versions", "", "comma-separated AE versions; defaults to "+aeversion.SupportedRange())
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: aepregistry gate [-root .] [-coverage flightdeck/work/aep-understanding-generation/versioned-aep-migration-coverage.json] [-out tmp/registry_gate.json] [-scope version-matrix|asset-policy|all] [-versions AE2020,AE2021,...] [-json]")
		return 2
	}
	if *scope != "version-matrix" && *scope != "asset-policy" && *scope != "all" {
		fmt.Fprintln(os.Stderr, "gate: -scope must be version-matrix, asset-policy, or all")
		return 2
	}

	axis := splitCSV(*versions)
	if len(axis) == 0 && (*scope == "version-matrix" || *scope == "all") {
		axis = registry.DefaultAEVersionAxis()
	}
	report := gateReport{
		SchemaVersion: 1,
		Status:        registry.StatusPass,
		Root:          filepath.ToSlash(*root),
		Scope:         *scope,
		CoveragePath:  filepath.ToSlash(*coveragePath),
		VersionAxis:   axis,
	}
	addStep := func(step gateStepReport) {
		report.Steps = append(report.Steps, step)
		report.Summary.Steps++
		if step.Status == registry.StatusFail || step.Errors > 0 {
			report.Summary.Failed++
			report.Summary.Errors += max(1, step.Errors)
			report.Status = registry.StatusFail
			return
		}
		report.Summary.Passed++
	}

	auditOut := "tmp/registry_audit.json"
	audit, err := registry.AuditRepository(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "gate audit:", err)
		return 2
	}
	if err := writeJSONFile(filepath.Join(*root, filepath.FromSlash(auditOut)), audit); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		return 2
	}
	addStep(gateStepReport{ID: "registry_audit", Command: "go run ./cmd/aepregistry audit -root . -out " + auditOut, Output: auditOut, Status: audit.Status, Errors: audit.Summary.Errors})

	if *scope == "version-matrix" || *scope == "all" {
		if err := runVersionMatrixGate(*root, *coveragePath, axis, addStep); err != nil {
			fmt.Fprintln(os.Stderr, "gate version-matrix:", err)
			return 2
		}
	}
	if *scope == "asset-policy" || *scope == "all" {
		if err := runAssetPolicyGate(*root, addStep); err != nil {
			fmt.Fprintln(os.Stderr, "gate asset-policy:", err)
			return 2
		}
	}

	if err := writeJSONFile(*outPath, report); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		return 2
	}
	if *jsonOut {
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, "stdout:", err)
			return 2
		}
	} else {
		fmt.Printf("registry gate: %s (%d steps, %d failed)\n", report.Status, report.Summary.Steps, report.Summary.Failed)
	}
	if report.Status == registry.StatusFail {
		return 1
	}
	return 0
}

func runVersionMatrixGate(root, coveragePath string, axis []string, addStep func(gateStepReport)) error {
	boundariesOut := "tmp/registry_version_boundaries.json"
	boundaries, err := registry.CheckVersionBoundaries(root, coveragePath, axis)
	if err != nil {
		return fmt.Errorf("boundaries: %w", err)
	}
	if err := writeJSONFile(filepath.Join(root, filepath.FromSlash(boundariesOut)), boundaries); err != nil {
		return fmt.Errorf("write boundaries: %w", err)
	}
	addStep(gateStepReport{ID: "registry_version_boundaries", Command: "go run ./cmd/aepregistry boundaries -root . -out " + boundariesOut, Output: boundariesOut, Status: boundaries.Status, Errors: boundaries.Summary.Errors})

	coverageOut := "tmp/registry_coverage.json"
	coverage, err := registry.ValidateCoverage(root, coveragePath)
	if err != nil {
		return fmt.Errorf("coverage: %w", err)
	}
	if err := writeJSONFile(filepath.Join(root, filepath.FromSlash(coverageOut)), coverage); err != nil {
		return fmt.Errorf("write coverage: %w", err)
	}
	addStep(gateStepReport{ID: "registry_coverage", Command: "go run ./cmd/aepregistry coverage -root . -out " + coverageOut, Output: coverageOut, Status: coverage.Status, Errors: coverage.Summary.Errors})

	summaryOut := "tmp/registry_coverage_summary.json"
	summary := registry.SummarizeCoverage(coverage)
	if err := writeJSONFile(filepath.Join(root, filepath.FromSlash(summaryOut)), summary); err != nil {
		return fmt.Errorf("write coverage summary: %w", err)
	}
	addStep(gateStepReport{ID: "registry_coverage_summary", Command: "go run ./cmd/aepregistry coverage -root . -out " + summaryOut + " -summary", Output: summaryOut, Status: summary.Status, Errors: summary.Errors})

	axisOut := "tmp/registry_coverage_axis.json"
	axisReport, err := registry.CoverageAxisWithFilter(root, coveragePath, axis, registry.CoverageAxisFilter{})
	if err != nil {
		return fmt.Errorf("coverage axis: %w", err)
	}
	if err := writeJSONFile(filepath.Join(root, filepath.FromSlash(axisOut)), axisReport); err != nil {
		return fmt.Errorf("write coverage axis: %w", err)
	}
	addStep(gateStepReport{ID: "registry_coverage_axis", Command: "go run ./cmd/aepregistry coverage -root . -out " + axisOut + " -axis", Output: axisOut, Status: axisReport.Status})

	boundaryAxisOut := "tmp/registry_coverage_axis_boundary.json"
	boundaryAxis, err := registry.CoverageAxisWithFilter(root, coveragePath, axis, registry.CoverageAxisFilter{AtomID: "layer.track_matte.explicit_source"})
	if err != nil {
		return fmt.Errorf("boundary axis: %w", err)
	}
	if err := writeJSONFile(filepath.Join(root, filepath.FromSlash(boundaryAxisOut)), boundaryAxis); err != nil {
		return fmt.Errorf("write boundary axis: %w", err)
	}
	addStep(gateStepReport{ID: "registry_coverage_axis_boundary", Command: "go run ./cmd/aepregistry coverage -root . -out " + boundaryAxisOut + " -axis -atom layer.track_matte.explicit_source", Output: boundaryAxisOut, Status: boundaryAxis.Status})
	return nil
}

func runAssetPolicyGate(root string, addStep func(gateStepReport)) error {
	ownershipOut := "tmp/registry_ownership.json"
	ownership, err := registry.OwnershipRepository(root, registry.OwnershipOptions{SampleLimit: 20})
	if err != nil {
		return fmt.Errorf("ownership: %w", err)
	}
	if err := writeJSONFile(filepath.Join(root, filepath.FromSlash(ownershipOut)), ownership); err != nil {
		return fmt.Errorf("write ownership: %w", err)
	}
	addStep(gateStepReport{ID: "registry_ownership", Command: "go run ./cmd/aepregistry ownership -root . -out " + ownershipOut + " -sample-limit 20", Output: ownershipOut, Status: registry.StatusPass})

	layoutOut := "tmp/registry_layout.json"
	layout, err := registry.LayoutRepository(root, registry.LayoutOptions{SampleLimit: 20})
	if err != nil {
		return fmt.Errorf("layout: %w", err)
	}
	if err := writeJSONFile(filepath.Join(root, filepath.FromSlash(layoutOut)), layout); err != nil {
		return fmt.Errorf("write layout: %w", err)
	}
	layoutErrors := layout.Summary.BlockedUnownedFiles + layout.Summary.MissingRequired
	layoutStatus := registry.StatusPass
	if layoutErrors > 0 {
		layoutStatus = registry.StatusFail
	}
	addStep(gateStepReport{ID: "registry_layout", Command: "go run ./cmd/aepregistry layout -root . -out " + layoutOut + " -sample-limit 20", Output: layoutOut, Status: layoutStatus, Errors: layoutErrors})

	cleanupOut := "tmp/registry_generated_cleanup.json"
	cleanup, err := registry.GeneratedCleanupRepository(root, registry.GeneratedCleanupOptions{
		SampleLimit:         3,
		StateReferenceFiles: defaultCleanupStateReferenceFiles(),
	})
	if err != nil {
		return fmt.Errorf("cleanup: %w", err)
	}
	if err := writeJSONFile(filepath.Join(root, filepath.FromSlash(cleanupOut)), cleanup); err != nil {
		return fmt.Errorf("write cleanup: %w", err)
	}
	addStep(gateStepReport{ID: "registry_generated_cleanup", Command: "go run ./cmd/aepregistry cleanup -root . -out " + cleanupOut + " -sample-limit 3", Output: cleanupOut, Status: registry.StatusPass})

	execOut := "tmp/registry_generated_cleanup_execution.json"
	exec := registry.ExecuteGeneratedCleanup(root, cleanup, registry.GeneratedCleanupExecutionOptions{ExcludeProducers: []string{"registry_report"}})
	if err := writeJSONFile(filepath.Join(root, filepath.FromSlash(execOut)), exec); err != nil {
		return fmt.Errorf("write cleanup execution: %w", err)
	}
	execStatus := registry.StatusPass
	if exec.Summary.Errors > 0 {
		execStatus = registry.StatusFail
	}
	addStep(gateStepReport{ID: "registry_generated_cleanup_execution_dry_run", Command: "go run ./cmd/aepregistry cleanup -root . -out " + cleanupOut + " -exec-out " + execOut + " -exclude-producer registry_report -sample-limit 3", Output: execOut, Status: execStatus, Errors: exec.Summary.Errors})

	pruneExecOut := "tmp/registry_generated_cleanup_execution_prune_review.json"
	pruneExec := registry.ExecuteGeneratedCleanup(root, cleanup, registry.GeneratedCleanupExecutionOptions{ExcludeProducers: []string{"registry_report"}, PruneReviewSiblings: true})
	if err := writeJSONFile(filepath.Join(root, filepath.FromSlash(pruneExecOut)), pruneExec); err != nil {
		return fmt.Errorf("write prune-review cleanup execution: %w", err)
	}
	pruneExecStatus := registry.StatusPass
	if pruneExec.Summary.Errors > 0 {
		pruneExecStatus = registry.StatusFail
	}
	addStep(gateStepReport{ID: "registry_generated_cleanup_execution_prune_review_dry_run", Command: "go run ./cmd/aepregistry cleanup -root . -out " + cleanupOut + " -exec-out " + pruneExecOut + " -prune-review-siblings -exclude-producer registry_report -sample-limit 3", Output: pruneExecOut, Status: pruneExecStatus, Errors: pruneExec.Summary.Errors})
	return nil
}

func runBoundaries(args []string) int {
	fs := flag.NewFlagSet("aepregistry boundaries", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repository root")
	coveragePath := fs.String("coverage", "flightdeck/work/aep-understanding-generation/versioned-aep-migration-coverage.json", "coverage ledger JSON path")
	outPath := fs.String("out", "tmp/registry_version_boundaries.json", "version boundary check report JSON path")
	versions := fs.String("versions", "", "comma-separated AE versions; defaults to "+aeversion.SupportedRange())
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: aepregistry boundaries [-root .] [-coverage flightdeck/work/aep-understanding-generation/versioned-aep-migration-coverage.json] [-out tmp/registry_version_boundaries.json] [-versions AE2020,AE2021,...] [-json]")
		return 2
	}

	report, err := registry.CheckVersionBoundaries(*root, *coveragePath, splitCSV(*versions))
	if err != nil {
		fmt.Fprintln(os.Stderr, "boundaries:", err)
		return 2
	}
	if err := writeJSONFile(*outPath, report); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		return 2
	}
	if *jsonOut {
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, "stdout:", err)
			return 2
		}
	} else {
		fmt.Printf("registry boundaries: %s (%d boundaries, %d matched, %d errors)\n",
			report.Status, report.Summary.Boundaries, report.Summary.Matched, report.Summary.Errors)
	}
	if report.Status == registry.StatusFail {
		return 1
	}
	return 0
}

func runCleanup(args []string) int {
	fs := flag.NewFlagSet("aepregistry cleanup", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repository root")
	outPath := fs.String("out", "tmp/registry_generated_cleanup.json", "generated cleanup classification JSON path")
	execOutPath := fs.String("exec-out", "", "optional generated cleanup execution report JSON path")
	sampleLimit := fs.Int("sample-limit", 5, "maximum file samples per generated group")
	apply := fs.Bool("apply", false, "apply deletion for cleanup_candidate groups; default is dry-run only")
	pruneReviewSiblings := fs.Bool("prune-review-siblings", false, "allow review groups to delete unreferenced siblings while preserving registered and state-referenced paths")
	producers := fs.String("producer", "", "comma-separated producer categories to include in execution report")
	excludeProducers := fs.String("exclude-producer", "registry_report", "comma-separated producer categories to exclude from execution report")
	stateRefs := fs.String("state-ref", strings.Join(defaultCleanupStateReferenceFiles(), ","), "comma-separated JSON state files whose tmp references protect generated cleanup groups")
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: aepregistry cleanup [-root .] [-out tmp/registry_generated_cleanup.json] [-exec-out tmp/registry_generated_cleanup_execution.json] [-apply] [-prune-review-siblings] [-producer categories] [-exclude-producer categories] [-state-ref json[,json...]] [-sample-limit 5] [-json]")
		return 2
	}
	if *apply && *execOutPath == "" {
		*execOutPath = "tmp/registry_generated_cleanup_execution.json"
	}

	report, err := registry.GeneratedCleanupRepository(*root, registry.GeneratedCleanupOptions{
		SampleLimit:         *sampleLimit,
		StateReferenceFiles: splitCSV(*stateRefs),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "cleanup:", err)
		return 2
	}
	if err := writeJSONFile(*outPath, report); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		return 2
	}
	if *execOutPath != "" {
		exec := registry.ExecuteGeneratedCleanup(*root, report, registry.GeneratedCleanupExecutionOptions{
			Apply:               *apply,
			PruneReviewSiblings: *pruneReviewSiblings,
			IncludeProducers:    splitCSV(*producers),
			ExcludeProducers:    splitCSV(*excludeProducers),
		})
		if err := writeJSONFile(*execOutPath, exec); err != nil {
			fmt.Fprintln(os.Stderr, "write:", err)
			return 2
		}
		if exec.Summary.Errors > 0 {
			fmt.Fprintf(os.Stderr, "cleanup execution: %d errors\n", exec.Summary.Errors)
			return 1
		}
	}
	if *jsonOut {
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, "stdout:", err)
			return 2
		}
	} else {
		fmt.Printf("registry cleanup: %d locations, %d groups, %d cleanup candidates\n",
			report.Summary.Locations, report.Summary.Groups, report.Summary.CleanupCandidateFiles)
	}
	return 0
}

func splitCSV(value string) []string {
	var out []string
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func defaultCleanupStateReferenceFiles() []string {
	return []string{
		"flightdeck/work/aep-understanding-generation/versioned-aep-migration-current.json",
		"flightdeck/work/aep-understanding-generation/versioned-aep-migration-coverage.json",
	}
}

func runAudit(args []string) int {
	fs := flag.NewFlagSet("aepregistry audit", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repository root")
	outPath := fs.String("out", "tmp/registry_audit.json", "audit report JSON path")
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: aepregistry audit [-root .] [-out tmp/registry_audit.json] [-json]")
		return 2
	}

	report, err := registry.AuditRepository(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "audit:", err)
		return 2
	}
	if err := writeJSONFile(*outPath, report); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		return 2
	}
	if *jsonOut {
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, "stdout:", err)
			return 2
		}
	} else {
		fmt.Printf("registry audit: %s (%d errors, %d warnings)\n", report.Status, report.Summary.Errors, report.Summary.Warnings)
	}
	if report.Status == registry.StatusFail {
		return 1
	}
	return 0
}

func runCoverage(args []string) int {
	fs := flag.NewFlagSet("aepregistry coverage", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repository root")
	coveragePath := fs.String("coverage", "flightdeck/work/aep-understanding-generation/versioned-aep-migration-coverage.json", "coverage ledger JSON path")
	outPath := fs.String("out", "tmp/registry_coverage.json", "coverage validation report JSON path")
	summaryOut := fs.Bool("summary", false, "write summary JSON instead of full coverage validation report")
	rowsOut := fs.Bool("rows", false, "write filtered atom rows JSON instead of full coverage validation report")
	cellsOut := fs.Bool("cells", false, "write filtered source-target matrix cells JSON instead of full coverage validation report")
	axisOut := fs.Bool("axis", false, "write AE version-axis coverage summary JSON instead of full coverage validation report")
	versions := fs.String("versions", "", "comma-separated AE versions for -axis; defaults to "+aeversion.SupportedRange())
	recordFilter := fs.String("record", "", "filter rows or cells by coverage record id")
	atomFilter := fs.String("atom", "", "filter rows or cells by atom id")
	recipeFilter := fs.String("recipe", "", "filter cells by recipe id")
	caseStatusFilter := fs.String("case-status", "", "filter cells by matrix case status")
	writerStatusFilter := fs.String("writer-status", "", "filter rows or cells by writer status")
	hostLevelFilter := fs.String("host-level", "", "filter rows or cells by host-open evidence level")
	boundaryStatusFilter := fs.String("boundary-status", "", "filter rows or cells by boundary status")
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: aepregistry coverage [-root .] [-coverage flightdeck/work/aep-understanding-generation/versioned-aep-migration-coverage.json] [-out tmp/registry_coverage.json] [-summary|-rows|-cells|-axis] [-versions AE2020,AE2021,...] [-record id] [-atom id] [-recipe id] [-case-status status] [-writer-status status] [-host-level level] [-boundary-status status] [-json]")
		return 2
	}
	modes := 0
	for _, enabled := range []bool{*summaryOut, *rowsOut, *cellsOut, *axisOut} {
		if enabled {
			modes++
		}
	}
	if modes > 1 {
		fmt.Fprintln(os.Stderr, "coverage: choose only one of -summary, -rows, -cells, or -axis")
		return 2
	}

	var status string
	var report registry.CoverageReport
	var output any
	if *axisOut {
		axis, err := registry.CoverageAxisWithFilter(*root, *coveragePath, splitCSV(*versions), registry.CoverageAxisFilter{
			RecordID:              *recordFilter,
			AtomID:                *atomFilter,
			Recipe:                *recipeFilter,
			CaseStatus:            *caseStatusFilter,
			WriterStatus:          *writerStatusFilter,
			HostOpenEvidenceLevel: *hostLevelFilter,
			BoundaryStatus:        *boundaryStatusFilter,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, "coverage:", err)
			return 2
		}
		output = axis
		status = axis.Status
	} else if *cellsOut {
		cells, err := registry.CoverageCells(*root, *coveragePath, registry.CoverageCellFilter{
			RecordID:              *recordFilter,
			AtomID:                *atomFilter,
			Recipe:                *recipeFilter,
			CaseStatus:            *caseStatusFilter,
			WriterStatus:          *writerStatusFilter,
			HostOpenEvidenceLevel: *hostLevelFilter,
			BoundaryStatus:        *boundaryStatusFilter,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, "coverage:", err)
			return 2
		}
		output = cells
		status = cells.Status
	} else {
		var err error
		report, err = registry.ValidateCoverage(*root, *coveragePath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "coverage:", err)
			return 2
		}
		output = report
		status = report.Status
		if *summaryOut {
			output = registry.SummarizeCoverage(report)
		} else if *rowsOut {
			output = registry.CoverageRows(report, registry.CoverageRowFilter{
				RecordID:              *recordFilter,
				AtomID:                *atomFilter,
				WriterStatus:          *writerStatusFilter,
				HostOpenEvidenceLevel: *hostLevelFilter,
				BoundaryStatus:        *boundaryStatusFilter,
			})
		}
	}
	if err := writeJSONFile(*outPath, output); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		return 2
	}
	if *jsonOut {
		if err := json.NewEncoder(os.Stdout).Encode(output); err != nil {
			fmt.Fprintln(os.Stderr, "stdout:", err)
			return 2
		}
	} else {
		if *summaryOut {
			summary := output.(registry.CoverageSummaryReport)
			fmt.Printf("registry coverage summary: %s (%d atom rows, %d errors)\n", summary.Status, summary.AtomRows, summary.Errors)
		} else if *rowsOut {
			rows := output.(registry.CoverageRowsReport)
			fmt.Printf("registry coverage rows: %s (%d rows)\n", rows.Status, rows.Count)
		} else if *cellsOut {
			cells := output.(registry.CoverageCellsReport)
			fmt.Printf("registry coverage cells: %s (%d cells)\n", cells.Status, cells.Count)
		} else if *axisOut {
			axis := output.(registry.CoverageAxisReport)
			fmt.Printf("registry coverage axis: %s (%d atom rows, %d cells)\n", axis.Status, axis.Summary.AtomRows, axis.Summary.Cells)
		} else {
			fmt.Printf("registry coverage: %s (%d artifacts, %d errors)\n", report.Status, report.Summary.Artifacts, report.Summary.Errors)
		}
	}
	if status == registry.StatusFail {
		return 1
	}
	return 0
}

func runInventory(args []string) int {
	fs := flag.NewFlagSet("aepregistry inventory", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repository root")
	outPath := fs.String("out", "tmp/registry_inventory.json", "inventory report JSON path")
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: aepregistry inventory [-root .] [-out tmp/registry_inventory.json] [-json]")
		return 2
	}

	report, err := registry.InventoryRepository(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "inventory:", err)
		return 2
	}
	if err := writeJSONFile(*outPath, report); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		return 2
	}
	if *jsonOut {
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, "stdout:", err)
			return 2
		}
	} else {
		fmt.Printf("registry inventory: %d locations, %d files, %d bytes\n", report.Summary.Locations, report.Summary.Files, report.Summary.Bytes)
	}
	return 0
}

func runLayout(args []string) int {
	fs := flag.NewFlagSet("aepregistry layout", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repository root")
	outPath := fs.String("out", "tmp/registry_layout.json", "layout cleanup guardrail JSON path")
	sampleLimit := fs.Int("sample-limit", 20, "maximum unowned samples per location")
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: aepregistry layout [-root .] [-out tmp/registry_layout.json] [-sample-limit 20] [-json]")
		return 2
	}

	report, err := registry.LayoutRepository(*root, registry.LayoutOptions{SampleLimit: *sampleLimit})
	if err != nil {
		fmt.Fprintln(os.Stderr, "layout:", err)
		return 2
	}
	if err := writeJSONFile(*outPath, report); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		return 2
	}
	if *jsonOut {
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, "stdout:", err)
			return 2
		}
	} else {
		fmt.Printf("registry layout: %d locations, %d cleanup candidates, %d blocked unowned\n",
			report.Summary.Locations, report.Summary.CleanupCandidateFiles, report.Summary.BlockedUnownedFiles)
	}
	return 0
}

func runOwnership(args []string) int {
	fs := flag.NewFlagSet("aepregistry ownership", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repository root")
	outPath := fs.String("out", "tmp/registry_ownership.json", "ownership report JSON path")
	sampleLimit := fs.Int("sample-limit", 20, "maximum unowned file samples per location")
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: aepregistry ownership [-root .] [-out tmp/registry_ownership.json] [-sample-limit 20] [-json]")
		return 2
	}

	report, err := registry.OwnershipRepository(*root, registry.OwnershipOptions{SampleLimit: *sampleLimit})
	if err != nil {
		fmt.Fprintln(os.Stderr, "ownership:", err)
		return 2
	}
	if err := writeJSONFile(*outPath, report); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		return 2
	}
	if *jsonOut {
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, "stdout:", err)
			return 2
		}
	} else {
		fmt.Printf("registry ownership: %d locations, %d files, %d owned, %d unowned\n",
			report.Summary.Locations,
			report.Summary.Files,
			report.Summary.OwnedFiles,
			report.Summary.UnownedFiles,
		)
	}
	return 0
}

func writeJSONFile(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: aepregistry <audit|boundaries|cleanup|coverage|gate|inventory|layout|ownership> [flags]")
}
