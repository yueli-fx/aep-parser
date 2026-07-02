package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

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
	case "coverage":
		return runCoverage(args[1:])
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
		fmt.Fprintln(os.Stderr, "usage: aepregistry coverage [-root .] [-coverage flightdeck/work/aep-understanding-generation/versioned-aep-migration-coverage.json] [-out tmp/registry_coverage.json] [-summary|-rows|-cells] [-record id] [-atom id] [-recipe id] [-case-status status] [-writer-status status] [-host-level level] [-boundary-status status] [-json]")
		return 2
	}
	modes := 0
	for _, enabled := range []bool{*summaryOut, *rowsOut, *cellsOut} {
		if enabled {
			modes++
		}
	}
	if modes > 1 {
		fmt.Fprintln(os.Stderr, "coverage: choose only one of -summary, -rows, or -cells")
		return 2
	}

	var status string
	var report registry.CoverageReport
	var output any
	if *cellsOut {
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
	fmt.Fprintln(os.Stderr, "usage: aepregistry <audit|coverage|inventory|layout|ownership> [flags]")
}
