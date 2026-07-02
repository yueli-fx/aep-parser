package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func runBoundaries(args []string) int {
	fs := flag.NewFlagSet("aepregistry boundaries", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repository root")
	coveragePath := fs.String("coverage", "flightdeck/work/aep-understanding-generation/versioned-aep-migration-coverage.json", "coverage ledger JSON path")
	outPath := fs.String("out", "tmp/registry_version_boundaries.json", "version boundary check report JSON path")
	versions := fs.String("versions", "", "comma-separated AE versions; defaults to AE2020-AE2025")
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
	producers := fs.String("producer", "", "comma-separated producer categories to include in execution report")
	excludeProducers := fs.String("exclude-producer", "registry_report", "comma-separated producer categories to exclude from execution report")
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: aepregistry cleanup [-root .] [-out tmp/registry_generated_cleanup.json] [-exec-out tmp/registry_generated_cleanup_execution.json] [-apply] [-producer categories] [-exclude-producer categories] [-sample-limit 5] [-json]")
		return 2
	}
	if *apply && *execOutPath == "" {
		*execOutPath = "tmp/registry_generated_cleanup_execution.json"
	}

	report, err := registry.GeneratedCleanupRepository(*root, registry.GeneratedCleanupOptions{SampleLimit: *sampleLimit})
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
			Apply:            *apply,
			IncludeProducers: splitCSV(*producers),
			ExcludeProducers: splitCSV(*excludeProducers),
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
	versions := fs.String("versions", "", "comma-separated AE versions for -axis; defaults to AE2020-AE2025")
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
	fmt.Fprintln(os.Stderr, "usage: aepregistry <audit|boundaries|cleanup|coverage|inventory|layout|ownership> [flags]")
}
