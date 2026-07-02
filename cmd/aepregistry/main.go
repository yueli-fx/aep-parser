package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
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
	case "coverage-batch":
		return runCoverageBatch(args[1:])
	case "coverage-md":
		return runCoverageMD(args[1:])
	case "coverage-update":
		return runCoverageUpdate(args[1:])
	case "checkpoint":
		return runCheckpoint(args[1:])
	case "current":
		return runCurrent(args[1:])
	case "gate":
		return runGate(args[1:])
	case "host-open-gaps":
		return runHostOpenGaps(args[1:])
	case "migration-summary":
		return runMigrationSummary(args[1:])
	case "cleanup":
		return runCleanup(args[1:])
	case "inventory":
		return runInventory(args[1:])
	case "layout":
		return runLayout(args[1:])
	case "ownership":
		return runOwnership(args[1:])
	case "recurring-matrix":
		return runRecurringMatrix(args[1:])
	default:
		usage()
		return 2
	}
}

func runCoverageMD(args []string) int {
	fs := flag.NewFlagSet("aepregistry coverage-md", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repository root")
	coveragePath := fs.String("coverage", "flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json", "coverage ledger JSON path")
	outPath := fs.String("out", "tmp/migration_coverage.md", "coverage markdown output path")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: aepregistry coverage-md [-root .] [-coverage flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json] [-out tmp/migration_coverage.md]")
		return 2
	}
	md, err := registry.RenderCoverageMarkdown(*root, *coveragePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "coverage-md:", err)
		return 2
	}
	out := resolveRootPath(*root, *outPath)
	if err := writeTextFile(out, md); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		return 2
	}
	fmt.Printf("rendered coverage markdown: %s\n", *outPath)
	return 0
}

func runCoverageUpdate(args []string) int {
	fs := flag.NewFlagSet("aepregistry coverage-update", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repository root")
	coveragePath := fs.String("coverage", "flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json", "coverage ledger JSON path")
	id := fs.String("id", "", "coverage record id")
	matrixPath := fs.String("matrix", "", "matrix JSON artifact path")
	domain := fs.String("domain", "", "domain for a new or updated coverage record")
	scope := fs.String("scope", "", "scope for a new or updated coverage record")
	writerStatus := fs.String("writer-status", "", "override writer status")
	hostOpenStatus := fs.String("host-open-status", "", "override host-open status")
	ledgerPath := fs.String("ledger", "", "optional ledger path; defaults to ledger.md next to the matrix")
	outPath := fs.String("out", "tmp/registry_coverage_update.json", "coverage update report JSON path")
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 || *id == "" || *matrixPath == "" {
		fmt.Fprintln(os.Stderr, "usage: aepregistry coverage-update [-root .] -id id -matrix tmp/matrix.json [-coverage path] [-domain name] [-scope text] [-writer-status status] [-host-open-status status] [-ledger path] [-out tmp/registry_coverage_update.json] [-json]")
		return 2
	}
	report, err := registry.UpdateCoverageFromMatrix(*root, registry.CoverageUpdateOptions{
		ID:             *id,
		CoveragePath:   *coveragePath,
		MatrixPath:     *matrixPath,
		Domain:         *domain,
		Scope:          *scope,
		WriterStatus:   *writerStatus,
		HostOpenStatus: *hostOpenStatus,
		LedgerPath:     *ledgerPath,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "coverage-update:", err)
		return 2
	}
	if err := writeJSONFile(resolveRootPath(*root, *outPath), report); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		return 2
	}
	if *jsonOut {
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, "stdout:", err)
			return 2
		}
	} else {
		fmt.Printf("updated coverage id %q from %s: total=%d pass=%d blocked=%d failed=%d skipped=%d\n",
			report.CoverageID, report.MatrixPath, report.Totals.Total, report.Totals.Pass, report.Totals.Blocked, report.Totals.Failed, report.Totals.Skipped)
	}
	return 0
}

func runCoverageBatch(args []string) int {
	fs := flag.NewFlagSet("aepregistry coverage-batch", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repository root")
	currentPath := fs.String("current", "flightdeck/work/versioned-aep-migration/versioned-aep-migration-current.json", "current state JSON path")
	coveragePath := fs.String("coverage", "flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json", "coverage ledger JSON path")
	outPath := fs.String("out", "tmp/registry_coverage_batch.json", "coverage batch report JSON path")
	batchID := fs.String("batch-id", "", "coverage batch id")
	list := fs.Bool("list", false, "list known coverage batches")
	skipRun := fs.Bool("skip-run", false, "validate existing batch artifacts without regenerating matrices or rewriting coverage JSON")
	runMatrices := fs.Bool("run-matrices", false, "rerun batch matrix artifacts with go run ./cmd/aepmigrate matrix, then sync and validate coverage")
	sync := fs.Bool("sync", false, "update the coverage JSON at -coverage from existing batch matrix artifacts before validating")
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: aepregistry coverage-batch [-root .] [-current path] [-coverage path] [-out tmp/registry_coverage_batch.json] [-list|-batch-id id (-skip-run [-sync]|-run-matrices)] [-json]")
		return 2
	}
	if *list {
		report, err := registry.ListCoverageBatches(*root, *currentPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "coverage-batch:", err)
			return 2
		}
		if err := writeJSONFile(resolveRootPath(*root, *outPath), report); err != nil {
			fmt.Fprintln(os.Stderr, "write:", err)
			return 2
		}
		if *jsonOut {
			if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
				fmt.Fprintln(os.Stderr, "stdout:", err)
				return 2
			}
		} else {
			fmt.Printf("coverage batches: %d batches, %d entries\n", report.Summary.Batches, report.Summary.Entries)
		}
		return 0
	}
	if *batchID == "" || (*skipRun == *runMatrices) {
		fmt.Fprintln(os.Stderr, "usage: aepregistry coverage-batch [-root .] [-current path] [-coverage path] [-out tmp/registry_coverage_batch.json] [-list|-batch-id id (-skip-run [-sync]|-run-matrices)] [-json]")
		return 2
	}
	var report registry.CoverageBatchReport
	var err error
	if *runMatrices {
		if err := runCoverageBatchMatrices(*root, *currentPath, *batchID); err != nil {
			fmt.Fprintln(os.Stderr, "coverage-batch:", err)
			return 2
		}
		report, err = registry.SyncCoverageBatchFromMatrices(*root, *currentPath, *coveragePath, *batchID)
	} else if *sync {
		report, err = registry.SyncCoverageBatchFromMatrices(*root, *currentPath, *coveragePath, *batchID)
	} else {
		report, err = registry.CheckCoverageBatch(*root, *currentPath, *coveragePath, *batchID)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "coverage-batch:", err)
		return 2
	}
	if err := writeJSONFile(resolveRootPath(*root, *outPath), report); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		return 2
	}
	if *jsonOut {
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, "stdout:", err)
			return 2
		}
	} else {
		fmt.Printf("coverage batch %s: %s (%d entries, %d errors)\n", report.BatchID, report.Status, report.Summary.Entries, report.Summary.Errors)
	}
	if report.Status == registry.StatusFail {
		return 1
	}
	return 0
}

var coverageBatchMatrixRunner = runCoverageBatchMatrixCommand

func runCoverageBatchMatrices(root, currentPath, batchID string) error {
	plan, err := registry.PlanCoverageBatchMatrices(root, currentPath, batchID)
	if err != nil {
		return err
	}
	if plan.Status == registry.StatusFail {
		return fmt.Errorf("matrix plan failed: %+v", plan.Issues)
	}
	for _, entry := range plan.Entries {
		exitCode, err := coverageBatchMatrixRunner(root, entry.Args)
		if err != nil && !allowedMatrixExit(exitCode, entry.AllowExitCodes) {
			return fmt.Errorf("matrix run failed for %q with exit code %d: %w", entry.CoverageID, exitCode, err)
		}
		if err == nil && exitCode != 0 && !allowedMatrixExit(exitCode, entry.AllowExitCodes) {
			return fmt.Errorf("matrix run failed for %q with exit code %d", entry.CoverageID, exitCode)
		}
	}
	return nil
}

func runCoverageBatchMatrixCommand(root string, args []string) (int, error) {
	cmd := exec.Command("go", args...)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err == nil {
		return 0, nil
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), err
	}
	return 1, err
}

func allowedMatrixExit(exitCode int, allowed []int) bool {
	if exitCode == 0 {
		return true
	}
	for _, value := range allowed {
		if value == exitCode {
			return true
		}
	}
	return false
}

type checkpointReport struct {
	SchemaVersion int                    `json:"schema_version"`
	Status        string                 `json:"status"`
	Root          string                 `json:"root"`
	CurrentPath   string                 `json:"current_path"`
	CoveragePath  string                 `json:"coverage_path"`
	SummaryPath   string                 `json:"summary_path"`
	Summary       checkpointSummary      `json:"summary"`
	Steps         []checkpointStepReport `json:"steps"`
}

type checkpointSummary struct {
	Steps  int `json:"steps"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
	Errors int `json:"errors"`
}

type checkpointStepReport struct {
	ID      string `json:"id"`
	Command string `json:"command"`
	Output  string `json:"output,omitempty"`
	Status  string `json:"status"`
	Errors  int    `json:"errors,omitempty"`
	Message string `json:"message,omitempty"`
}

type checkpointCurrentState struct {
	TruthSources struct {
		Coverage string `json:"coverage"`
	} `json:"truth_sources"`
	CurrentState struct {
		AEInstallRoot string `json:"ae_install_root"`
	} `json:"current_state"`
	CanonicalCoverageBatch string `json:"canonical_coverage_batch"`
}

type checkpointCoverageContractFile struct {
	ContractGates []checkpointCoverageContractGate `json:"contract_gates"`
}

type checkpointCoverageContractGate struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"`
	Command  string `json:"command"`
	Artifact string `json:"artifact"`
}

func runCheckpoint(args []string) int {
	fs := flag.NewFlagSet("aepregistry checkpoint", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repository root")
	currentPath := fs.String("current", "flightdeck/work/versioned-aep-migration/versioned-aep-migration-current.json", "current state JSON path")
	coveragePath := fs.String("coverage", "flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json", "coverage ledger JSON path")
	summaryPath := fs.String("summary", "tmp/migration_coverage_summary.json", "migration coverage summary JSON path")
	outPath := fs.String("out", "tmp/registry_checkpoint.json", "checkpoint report JSON path")
	coverageBatchID := fs.String("coverage-batch-id", "", "coverage batch id for -include-coverage-batch; defaults to current canonical_coverage_batch")
	includeCoverageBatch := fs.Bool("include-coverage-batch", false, "replay the coverage batch into a temporary coverage candidate using Go sync")
	runCoverageBatchMatricesFlag := fs.Bool("run-coverage-batch-matrices", false, "with -include-coverage-batch, rerun batch matrices before syncing the temporary coverage candidate")
	skipDiffCheck := fs.Bool("skip-diff-check", false, "skip git diff --check; intended for non-git test fixtures")
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: aepregistry checkpoint [-root .] [-current path] [-coverage path] [-summary tmp/migration_coverage_summary.json] [-out tmp/registry_checkpoint.json] [-include-coverage-batch] [-run-coverage-batch-matrices] [-coverage-batch-id id] [-skip-diff-check] [-json]")
		return 2
	}

	report := checkpointReport{
		SchemaVersion: 1,
		Status:        registry.StatusPass,
		Root:          filepath.ToSlash(*root),
		CurrentPath:   filepath.ToSlash(*currentPath),
		CoveragePath:  filepath.ToSlash(*coveragePath),
		SummaryPath:   filepath.ToSlash(*summaryPath),
	}
	current, err := readCheckpointCurrent(resolveRootPath(*root, *currentPath))
	if err != nil {
		fmt.Fprintln(os.Stderr, "checkpoint:", err)
		return 2
	}
	if *coveragePath == "" && current.TruthSources.Coverage != "" {
		*coveragePath = current.TruthSources.Coverage
		report.CoveragePath = filepath.ToSlash(*coveragePath)
	}
	if *coverageBatchID == "" {
		*coverageBatchID = current.CanonicalCoverageBatch
	}

	addStep := func(step checkpointStepReport) {
		report.Steps = append(report.Steps, step)
		report.Summary.Steps++
		if step.Status == registry.StatusFail || step.Errors > 0 {
			report.Status = registry.StatusFail
			report.Summary.Failed++
			report.Summary.Errors += max(1, step.Errors)
			return
		}
		report.Summary.Passed++
	}
	stopIfFailed := func() bool {
		return report.Status == registry.StatusFail
	}

	currentReport, err := registry.ValidateCurrent(*root, *currentPath)
	addStep(checkpointStep("current", "go run ./cmd/aepregistry current -root . -current "+*currentPath+" -out tmp/registry_current.json", "tmp/registry_current.json", currentReport.Status, currentReport.Summary.Errors, err))
	if err == nil {
		if writeErr := writeJSONFile(resolveRootPath(*root, "tmp/registry_current.json"), currentReport); writeErr != nil {
			addStep(checkpointStep("current_write", "write tmp/registry_current.json", "tmp/registry_current.json", registry.StatusFail, 1, writeErr))
		}
	}
	if !stopIfFailed() {
		for _, gate := range checkpointVersionBoundaryGates(*root, *coveragePath) {
			out := checkpointValueOr(gate.Artifact, "tmp/registry_version_boundaries.json")
			command := checkpointValueOr(gate.Command, "go run ./cmd/aepregistry boundaries -root . -coverage "+*coveragePath+" -out "+out)
			boundaries, err := registry.CheckVersionBoundaries(*root, *coveragePath, nil)
			addStep(checkpointStep("version_boundaries", command, out, boundaries.Status, boundaries.Summary.Errors, err))
			if err == nil {
				if writeErr := writeJSONFile(resolveRootPath(*root, out), boundaries); writeErr != nil {
					addStep(checkpointStep("version_boundaries_write", "write "+out, out, registry.StatusFail, 1, writeErr))
				}
			}
			if stopIfFailed() {
				break
			}
		}
	}
	if !stopIfFailed() {
		coverageReport, err := registry.ValidateCoverageWithOptions(*root, *coveragePath, registry.CoverageValidationOptions{RequireLedgers: true, RequireAllRecipes: true})
		addStep(checkpointStep("coverage", "go run ./cmd/aepregistry coverage -root . -coverage "+*coveragePath+" -out tmp/registry_coverage.json -require-ledgers", "tmp/registry_coverage.json", coverageReport.Status, coverageReport.Summary.Errors, err))
		if err == nil {
			if writeErr := writeJSONFile(resolveRootPath(*root, "tmp/registry_coverage.json"), coverageReport); writeErr != nil {
				addStep(checkpointStep("coverage_write", "write tmp/registry_coverage.json", "tmp/registry_coverage.json", registry.StatusFail, 1, writeErr))
			}
		}
	}
	if !stopIfFailed() {
		summary, err := registry.BuildMigrationCoverageSummary(*root, *coveragePath)
		addStep(checkpointStep("coverage_summary_render", "go run ./cmd/aepregistry migration-summary -root . -coverage "+*coveragePath+" -out "+*summaryPath, *summaryPath, registry.StatusPass, 0, err))
		if err == nil {
			if writeErr := writeJSONFile(resolveRootPath(*root, *summaryPath), summary); writeErr != nil {
				addStep(checkpointStep("coverage_summary_write", "write "+*summaryPath, *summaryPath, registry.StatusFail, 1, writeErr))
			}
		}
	}
	if !stopIfFailed() {
		summaryCheck, err := registry.CheckMigrationCoverageSummary(*root, *coveragePath, *summaryPath)
		addStep(checkpointStep("coverage_summary_check", "go run ./cmd/aepregistry migration-summary -root . -coverage "+*coveragePath+" -out "+*summaryPath+" -check", *summaryPath, summaryCheck.Status, summaryCheck.Errors, err))
	}
	if !stopIfFailed() && current.CurrentState.AEInstallRoot != "" {
		hostOpen, err := registry.PlanHostOpenGaps(*root, *coveragePath, registry.HostOpenGapPlanOptions{AERoot: current.CurrentState.AEInstallRoot, MaxAEOpenCases: 24})
		addStep(checkpointStep("host_open_gaps", "go run ./cmd/aepregistry host-open-gaps -root . -coverage "+*coveragePath+" -out tmp/host_open_gap_audit_go.json -ae-root "+current.CurrentState.AEInstallRoot, "tmp/host_open_gap_audit_go.json", hostOpen.Status, 0, err))
		if err == nil {
			if writeErr := writeJSONFile(resolveRootPath(*root, "tmp/host_open_gap_audit_go.json"), hostOpen); writeErr != nil {
				addStep(checkpointStep("host_open_gaps_write", "write tmp/host_open_gap_audit_go.json", "tmp/host_open_gap_audit_go.json", registry.StatusFail, 1, writeErr))
			}
		}
	}
	if !stopIfFailed() && *includeCoverageBatch {
		if *coverageBatchID == "" {
			addStep(checkpointStep("coverage_batch_replay", "go run ./cmd/aepregistry coverage-batch -skip-run -sync", "tmp/registry_coverage_batch.json", registry.StatusFail, 1, fmt.Errorf("coverage batch id is required")))
		} else {
			tempCoverage, cleanup, err := copyCoverageCandidate(*root, *coveragePath)
			if err != nil {
				addStep(checkpointStep("coverage_batch_replay", "copy temporary coverage candidate", "", registry.StatusFail, 1, err))
			} else {
				defer cleanup()
				if *runCoverageBatchMatricesFlag {
					err := runCoverageBatchMatrices(*root, *currentPath, *coverageBatchID)
					addStep(checkpointStep("coverage_batch_matrices", "go run ./cmd/aepregistry coverage-batch -root . -current "+*currentPath+" -coverage <temp coverage copy> -out tmp/registry_coverage_batch.json -batch-id "+*coverageBatchID+" -run-matrices", "", registry.StatusPass, 0, err))
				}
				if !stopIfFailed() {
					batch, err := registry.SyncCoverageBatchFromMatrices(*root, *currentPath, tempCoverage, *coverageBatchID)
					addStep(checkpointStep("coverage_batch_replay", "go run ./cmd/aepregistry coverage-batch -root . -current "+*currentPath+" -coverage <temp coverage copy> -out tmp/registry_coverage_batch.json -batch-id "+*coverageBatchID+" -skip-run -sync", "tmp/registry_coverage_batch.json", batch.Status, batch.Summary.Errors, err))
					if err == nil {
						if writeErr := writeJSONFile(resolveRootPath(*root, "tmp/registry_coverage_batch.json"), batch); writeErr != nil {
							addStep(checkpointStep("coverage_batch_replay_write", "write tmp/registry_coverage_batch.json", "tmp/registry_coverage_batch.json", registry.StatusFail, 1, writeErr))
						}
					}
				}
			}
		}
	}
	if !stopIfFailed() && !*skipDiffCheck {
		err := runGitDiffCheck(*root)
		addStep(checkpointStep("diff_check", "git diff --check", "", registry.StatusPass, 0, err))
	}

	if err := writeJSONFile(resolveRootPath(*root, *outPath), report); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		return 2
	}
	if *jsonOut {
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, "stdout:", err)
			return 2
		}
	} else {
		fmt.Printf("checkpoint: %s (%d steps, %d failed)\n", report.Status, report.Summary.Steps, report.Summary.Failed)
	}
	if report.Status == registry.StatusFail {
		return 1
	}
	return 0
}

func checkpointValueOr(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func checkpointVersionBoundaryGates(root, coveragePath string) []checkpointCoverageContractGate {
	var coverage checkpointCoverageContractFile
	if err := readJSONFile(resolveRootPath(root, coveragePath), &coverage); err != nil {
		return nil
	}
	var gates []checkpointCoverageContractGate
	for _, gate := range coverage.ContractGates {
		if gate.Kind == "version_boundaries" {
			gates = append(gates, gate)
		}
	}
	return gates
}

func checkpointStep(id, command, output, status string, errors int, err error) checkpointStepReport {
	if err != nil {
		return checkpointStepReport{ID: id, Command: command, Output: output, Status: registry.StatusFail, Errors: max(1, errors), Message: err.Error()}
	}
	if status == "" {
		status = registry.StatusPass
	}
	return checkpointStepReport{ID: id, Command: command, Output: output, Status: status, Errors: errors}
}

func readCheckpointCurrent(path string) (checkpointCurrentState, error) {
	var current checkpointCurrentState
	if err := readJSONFile(path, &current); err != nil {
		return checkpointCurrentState{}, err
	}
	return current, nil
}

func copyCoverageCandidate(root, coveragePath string) (string, func(), error) {
	data, err := os.ReadFile(resolveRootPath(root, coveragePath))
	if err != nil {
		return "", nil, err
	}
	file, err := os.CreateTemp("", "aep-coverage-candidate-*.json")
	if err != nil {
		return "", nil, err
	}
	path := file.Name()
	if _, err := file.Write(data); err != nil {
		file.Close()
		os.Remove(path)
		return "", nil, err
	}
	if err := file.Close(); err != nil {
		os.Remove(path)
		return "", nil, err
	}
	return path, func() { _ = os.Remove(path) }, nil
}

func runGitDiffCheck(root string) error {
	cmd := exec.Command("git", "diff", "--check")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func runCurrent(args []string) int {
	fs := flag.NewFlagSet("aepregistry current", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repository root")
	currentPath := fs.String("current", "flightdeck/work/versioned-aep-migration/versioned-aep-migration-current.json", "current state JSON path")
	outPath := fs.String("out", "tmp/registry_current.json", "current validation report JSON path")
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: aepregistry current [-root .] [-current flightdeck/work/versioned-aep-migration/versioned-aep-migration-current.json] [-out tmp/registry_current.json] [-json]")
		return 2
	}
	report, err := registry.ValidateCurrent(*root, *currentPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "current:", err)
		return 2
	}
	if err := writeJSONFile(resolveRootPath(*root, *outPath), report); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		return 2
	}
	if *jsonOut {
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, "stdout:", err)
			return 2
		}
	} else {
		fmt.Printf("current validation: %s (%d coverage batches, %d tooling entries, %d errors)\n", report.Status, report.Summary.CoverageBatches, report.Summary.Tooling, report.Summary.Errors)
	}
	if report.Status == registry.StatusFail {
		return 1
	}
	return 0
}

func runRecurringMatrix(args []string) int {
	fs := flag.NewFlagSet("aepregistry recurring-matrix", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repository root")
	outRoot := fs.String("out-root", "tmp/migration_matrix_verify", "recurring matrix artifact root")
	outPath := fs.String("out", "tmp/registry_recurring_matrix.json", "recurring matrix gate report JSON path")
	skipRun := fs.Bool("skip-run", false, "validate existing artifacts without regenerating matrices")
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: aepregistry recurring-matrix [-root .] [-out-root tmp/migration_matrix_verify] [-out tmp/registry_recurring_matrix.json] [-skip-run] [-json]")
		return 2
	}
	if !*skipRun {
		if err := runRecurringMatrixArtifacts(*root, *outRoot); err != nil {
			fmt.Fprintln(os.Stderr, "recurring-matrix:", err)
			return 2
		}
	}
	report, err := registry.CheckRecurringMatrixGates(*root, *outRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, "recurring-matrix:", err)
		return 2
	}
	reportPath := resolveRootPath(*root, *outPath)
	if err := writeJSONFile(reportPath, report); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		return 2
	}
	if *jsonOut {
		if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, "stdout:", err)
			return 2
		}
	} else {
		fmt.Printf("recurring matrix gates: %s (%d gates, %d errors)\n", report.Status, report.Summary.Gates, report.Summary.Errors)
	}
	if report.Status == registry.StatusFail {
		return 1
	}
	return 0
}

func runRecurringMatrixArtifacts(root, outRoot string) error {
	smokeOut := filepath.ToSlash(filepath.Join(outRoot, "smoke_all"))
	explicitMatteOut := filepath.ToSlash(filepath.Join(outRoot, "explicit_matte_ae2025"))
	if err := runGoCommand(root,
		"run", "./cmd/aepmigrate", "matrix",
		"-recipes", "examples/recipes",
		"-sources", "AE2020",
		"-targets", "all",
		"-out", smokeOut,
		"-ledger-out", filepath.ToSlash(filepath.Join(smokeOut, "ledger.md")),
	); err != nil {
		return err
	}
	verifyCase := filepath.ToSlash(filepath.Join(smokeOut, "minimal-adjustment-layer", "AE2020_to_AE2025"))
	if err := runGoCommand(root,
		"run", "./cmd/aepmigrate", "verify",
		"-source", filepath.ToSlash(filepath.Join(verifyCase, "source.aep")),
		"-target", filepath.ToSlash(filepath.Join(verifyCase, "target.aep")),
		"-target-version", "AE2025",
		"-report", filepath.ToSlash(filepath.Join(verifyCase, "convert_report.json")),
		"-out", filepath.ToSlash(filepath.Join(verifyCase, "verify_report.json")),
	); err != nil {
		return err
	}
	return runGoCommand(root,
		"run", "./cmd/aepmigrate", "matrix",
		"-recipe", "examples/recipes/minimal-layer-explicit-matte.json",
		"-sources", "AE2025",
		"-targets", "AE2025",
		"-out", explicitMatteOut,
		"-ledger-out", filepath.ToSlash(filepath.Join(explicitMatteOut, "ledger.md")),
	)
}

func runGoCommand(root string, args ...string) error {
	cmd := exec.Command("go", args...)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func resolveRootPath(root, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(root, filepath.FromSlash(path))
}

func runMigrationSummary(args []string) int {
	fs := flag.NewFlagSet("aepregistry migration-summary", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repository root")
	coveragePath := fs.String("coverage", "flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json", "coverage ledger JSON path")
	outPath := fs.String("out", "tmp/migration_coverage_summary.json", "migration coverage summary JSON path")
	check := fs.Bool("check", false, "validate the existing summary at -out instead of writing a new summary")
	totalsQuery := fs.Bool("totals", false, "print only migration coverage summary totals")
	domainQuery := fs.String("domain", "", "print migration coverage summary domain rollup")
	coverageIDQuery := fs.String("coverage-id", "", "print migration coverage summary record by coverage id")
	recipeQuery := fs.String("recipe", "", "print migration coverage summary recipe index entry")
	evidenceLevelQuery := fs.String("evidence-level", "", "print migration coverage summary recipes by host-open evidence level")
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: aepregistry migration-summary [-root .] [-coverage flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json] [-out tmp/migration_coverage_summary.json] [-check] [-totals|-domain name|-coverage-id id|-recipe name|-evidence-level level] [-json]")
		return 2
	}
	queryModes := 0
	for _, enabled := range []bool{*totalsQuery, *domainQuery != "", *coverageIDQuery != "", *recipeQuery != "", *evidenceLevelQuery != ""} {
		if enabled {
			queryModes++
		}
	}
	if queryModes > 1 {
		fmt.Fprintln(os.Stderr, "migration-summary: choose only one of -totals, -domain, -coverage-id, -recipe, or -evidence-level")
		return 2
	}
	if *check && queryModes > 0 {
		fmt.Fprintln(os.Stderr, "migration-summary: -check cannot be combined with query flags")
		return 2
	}
	if queryModes > 0 {
		summary, err := readOrBuildMigrationSummary(*root, *coveragePath, *outPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "migration-summary:", err)
			return 2
		}
		result, err := registry.QueryMigrationCoverageSummary(summary, registry.MigrationCoverageSummaryQuery{
			Totals:        *totalsQuery,
			Domain:        *domainQuery,
			CoverageID:    *coverageIDQuery,
			Recipe:        *recipeQuery,
			EvidenceLevel: *evidenceLevelQuery,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, "migration-summary:", err)
			return 1
		}
		if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
			fmt.Fprintln(os.Stderr, "stdout:", err)
			return 2
		}
		return 0
	}
	if *check {
		report, err := registry.CheckMigrationCoverageSummary(*root, *coveragePath, *outPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "migration-summary:", err)
			return 2
		}
		if *jsonOut {
			if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
				fmt.Fprintln(os.Stderr, "stdout:", err)
				return 2
			}
		} else {
			fmt.Printf("coverage summary check: %s (%d errors)\n", report.Status, report.Errors)
		}
		if report.Status == registry.StatusFail {
			return 1
		}
		return 0
	}
	summary, err := registry.BuildMigrationCoverageSummary(*root, *coveragePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "migration-summary:", err)
		return 2
	}
	if err := writeJSONFile(*outPath, summary); err != nil {
		fmt.Fprintln(os.Stderr, "write:", err)
		return 2
	}
	if *jsonOut {
		if err := json.NewEncoder(os.Stdout).Encode(summary); err != nil {
			fmt.Fprintln(os.Stderr, "stdout:", err)
			return 2
		}
	} else {
		fmt.Printf("rendered coverage summary json: %s\n", *outPath)
	}
	return 0
}

func readOrBuildMigrationSummary(root, coveragePath, outPath string) (registry.MigrationCoverageSummary, error) {
	var summary registry.MigrationCoverageSummary
	if err := readJSONFile(outPath, &summary); err == nil {
		return summary, nil
	}
	summary, err := registry.BuildMigrationCoverageSummary(root, coveragePath)
	if err != nil {
		return registry.MigrationCoverageSummary{}, err
	}
	if err := writeJSONFile(outPath, summary); err != nil {
		return registry.MigrationCoverageSummary{}, err
	}
	return summary, nil
}

func readJSONFile(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

func runHostOpenGaps(args []string) int {
	fs := flag.NewFlagSet("aepregistry host-open-gaps", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repository root")
	coveragePath := fs.String("coverage", "flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json", "coverage ledger JSON path")
	outPath := fs.String("out", "tmp/host_open_gap_audit.json", "host-open gap planner JSON path")
	aeRoot := fs.String("ae-root", "", "After Effects install root to include in generated matrix commands")
	maxAEOpenCases := fs.Int("max-ae-open-cases", 24, "maximum AE-open cases per generated matrix chunk")
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: aepregistry host-open-gaps [-root .] [-coverage flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json] [-out tmp/host_open_gap_audit.json] [-ae-root E:/adobe] [-max-ae-open-cases 24] [-json]")
		return 2
	}
	report, err := registry.PlanHostOpenGaps(*root, *coveragePath, registry.HostOpenGapPlanOptions{
		AERoot:         *aeRoot,
		MaxAEOpenCases: *maxAEOpenCases,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "host-open-gaps:", err)
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
		fmt.Printf("host-open gap audit: %s (%d groups, %d recipes, %d chunks)\n", report.Status, report.Summary.GapGroups, report.Summary.GapRecipes, report.Summary.Chunks)
	}
	return 0
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

type mainlineSpecCoverageAxisReport struct {
	SchemaVersion int                             `json:"schema_version"`
	Status        string                          `json:"status"`
	SpecPath      string                          `json:"spec_path"`
	Expected      map[string]int                  `json:"expected"`
	Actual        map[string]int                  `json:"actual"`
	Issues        []mainlineSpecCoverageAxisIssue `json:"issues,omitempty"`
}

type mainlineSpecCoverageAxisIssue struct {
	Field    string `json:"field"`
	Expected int    `json:"expected"`
	Actual   int    `json:"actual"`
	Message  string `json:"message"`
}

type mainlineSpecCoverageSummaryReport struct {
	SchemaVersion int                                `json:"schema_version"`
	Status        string                             `json:"status"`
	SpecPath      string                             `json:"spec_path"`
	Expected      map[string]int                     `json:"expected"`
	Actual        map[string]int                     `json:"actual"`
	Issues        []mainlineSpecCoverageSummaryIssue `json:"issues,omitempty"`
}

type mainlineSpecCoverageSummaryIssue struct {
	Field    string `json:"field"`
	Expected int    `json:"expected"`
	Actual   int    `json:"actual"`
	Message  string `json:"message"`
}

type layoutCleanupConsistencyReport struct {
	SchemaVersion int                             `json:"schema_version"`
	Status        string                          `json:"status"`
	Layout        map[string]int                  `json:"layout"`
	Cleanup       map[string]int                  `json:"cleanup"`
	Issues        []layoutCleanupConsistencyIssue `json:"issues,omitempty"`
}

type layoutCleanupConsistencyIssue struct {
	Field   string `json:"field"`
	Layout  int    `json:"layout"`
	Cleanup int    `json:"cleanup"`
	Message string `json:"message"`
}

func runGate(args []string) int {
	fs := flag.NewFlagSet("aepregistry gate", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repository root")
	coveragePath := fs.String("coverage", "flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json", "coverage ledger JSON path")
	outPath := fs.String("out", "tmp/registry_gate.json", "ordered registry gate report JSON path")
	scope := fs.String("scope", "version-matrix", "gate scope: version-matrix, asset-policy, or all")
	versions := fs.String("versions", "", "comma-separated AE versions; defaults to "+aeversion.SupportedRange())
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: aepregistry gate [-root .] [-coverage flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json] [-out tmp/registry_gate.json] [-scope version-matrix|asset-policy|all] [-versions AE2020,AE2021,...] [-json]")
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
	summaryStatus := summary.Status
	summaryErrors := summary.Errors + summary.Summary.RecipesWithoutAtomRows + summary.Summary.AtomRowsWithoutRecipes
	if summaryErrors > 0 {
		summaryStatus = registry.StatusFail
	}
	addStep(gateStepReport{ID: "registry_coverage_summary", Command: "go run ./cmd/aepregistry coverage -root . -out " + summaryOut + " -summary", Output: summaryOut, Status: summaryStatus, Errors: summaryErrors})

	specSummaryOut := "tmp/registry_mainline_spec_coverage_summary.json"
	specSummary, exists, err := checkMainlineSpecCoverageSummary(root, summary)
	if err != nil {
		return fmt.Errorf("mainline spec coverage summary: %w", err)
	}
	if exists {
		if err := writeJSONFile(filepath.Join(root, filepath.FromSlash(specSummaryOut)), specSummary); err != nil {
			return fmt.Errorf("write mainline spec coverage summary: %w", err)
		}
		addStep(gateStepReport{ID: "registry_mainline_spec_coverage_summary", Command: "go run ./cmd/aepregistry gate -root . -out tmp/registry_gate.json -scope version-matrix", Output: specSummaryOut, Status: specSummary.Status, Errors: len(specSummary.Issues)})
	}

	axisOut := "tmp/registry_coverage_axis.json"
	axisReport, err := registry.CoverageAxisWithFilter(root, coveragePath, axis, registry.CoverageAxisFilter{})
	if err != nil {
		return fmt.Errorf("coverage axis: %w", err)
	}
	if err := writeJSONFile(filepath.Join(root, filepath.FromSlash(axisOut)), axisReport); err != nil {
		return fmt.Errorf("write coverage axis: %w", err)
	}
	addStep(gateStepReport{ID: "registry_coverage_axis", Command: "go run ./cmd/aepregistry coverage -root . -out " + axisOut + " -axis", Output: axisOut, Status: axisReport.Status})

	specAxisOut := "tmp/registry_mainline_spec_coverage_axis.json"
	specAxis, exists, err := checkMainlineSpecCoverageAxis(root, axisReport)
	if err != nil {
		return fmt.Errorf("mainline spec coverage axis: %w", err)
	}
	if exists {
		if err := writeJSONFile(filepath.Join(root, filepath.FromSlash(specAxisOut)), specAxis); err != nil {
			return fmt.Errorf("write mainline spec coverage axis: %w", err)
		}
		addStep(gateStepReport{ID: "registry_mainline_spec_coverage_axis", Command: "go run ./cmd/aepregistry gate -root . -out tmp/registry_gate.json -scope version-matrix", Output: specAxisOut, Status: specAxis.Status, Errors: len(specAxis.Issues)})
	}

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

func checkMainlineSpecCoverageSummary(root string, summary registry.CoverageSummaryReport) (mainlineSpecCoverageSummaryReport, bool, error) {
	specPath := "flightdeck/work/versioned-aep-migration/mainline-spec.json"
	absPath := filepath.Join(root, filepath.FromSlash(specPath))
	if _, err := os.Stat(absPath); err != nil {
		if os.IsNotExist(err) {
			return mainlineSpecCoverageSummaryReport{}, false, nil
		}
		return mainlineSpecCoverageSummaryReport{}, false, err
	}
	var spec struct {
		CurrentExecution struct {
			LastCodeMaintenanceTarget struct {
				Result struct {
					CurrentTotals map[string]int `json:"current_totals"`
				} `json:"result"`
			} `json:"last_code_maintenance_target"`
		} `json:"current_execution"`
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return mainlineSpecCoverageSummaryReport{}, false, err
	}
	if err := json.Unmarshal(data, &spec); err != nil {
		return mainlineSpecCoverageSummaryReport{}, false, err
	}
	expected := spec.CurrentExecution.LastCodeMaintenanceTarget.Result.CurrentTotals
	if len(expected) == 0 {
		return mainlineSpecCoverageSummaryReport{}, false, nil
	}
	actual := coverageSummaryTotalsMap(summary.Summary)
	report := mainlineSpecCoverageSummaryReport{
		SchemaVersion: 1,
		Status:        registry.StatusPass,
		SpecPath:      specPath,
		Expected:      expected,
		Actual:        actual,
	}
	for _, field := range coverageSummaryTotalsFields() {
		if expected[field] != actual[field] {
			report.Issues = append(report.Issues, mainlineSpecCoverageSummaryIssue{
				Field:    field,
				Expected: expected[field],
				Actual:   actual[field],
				Message:  fmt.Sprintf("mainline spec current_totals %s=%d does not match coverage summary %d", field, expected[field], actual[field]),
			})
		}
	}
	if len(report.Issues) > 0 {
		report.Status = registry.StatusFail
	}
	return report, true, nil
}

func coverageSummaryTotalsFields() []string {
	return []string{
		"records",
		"artifacts",
		"contract_gates",
		"atoms",
		"atom_rows",
		"errors",
		"direct_host_atoms",
		"inferred_host_atoms",
		"declared_recipes",
		"observed_recipes",
		"atom_row_recipes",
		"recipes_without_atom_rows",
		"atom_rows_without_recipes",
		"observed_recipes_without_declaration",
		"records_with_undeclared_recipes",
	}
}

func coverageSummaryTotalsMap(summary registry.CoverageSummaryTotals) map[string]int {
	return map[string]int{
		"records":                              summary.Records,
		"artifacts":                            summary.Artifacts,
		"contract_gates":                       summary.ContractGates,
		"atoms":                                summary.Atoms,
		"atom_rows":                            summary.AtomRows,
		"errors":                               summary.Errors,
		"direct_host_atoms":                    summary.DirectHostAtoms,
		"inferred_host_atoms":                  summary.InferredHostAtoms,
		"declared_recipes":                     summary.DeclaredRecipes,
		"observed_recipes":                     summary.ObservedRecipes,
		"atom_row_recipes":                     summary.AtomRowRecipes,
		"recipes_without_atom_rows":            summary.RecipesWithoutAtomRows,
		"atom_rows_without_recipes":            summary.AtomRowsWithoutRecipes,
		"observed_recipes_without_declaration": summary.ObservedRecipesWithoutDeclaration,
		"records_with_undeclared_recipes":      summary.RecordsWithUndeclaredRecipes,
	}
}

func checkMainlineSpecCoverageAxis(root string, axisReport registry.CoverageAxisReport) (mainlineSpecCoverageAxisReport, bool, error) {
	specPath := "flightdeck/work/versioned-aep-migration/mainline-spec.json"
	absPath := filepath.Join(root, filepath.FromSlash(specPath))
	if _, err := os.Stat(absPath); err != nil {
		if os.IsNotExist(err) {
			return mainlineSpecCoverageAxisReport{}, false, nil
		}
		return mainlineSpecCoverageAxisReport{}, false, err
	}
	var spec struct {
		CurrentExecution struct {
			LastCompletedTarget struct {
				Result struct {
					Summary map[string]int `json:"summary"`
				} `json:"result"`
			} `json:"last_completed_target"`
		} `json:"current_execution"`
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return mainlineSpecCoverageAxisReport{}, false, err
	}
	if err := json.Unmarshal(data, &spec); err != nil {
		return mainlineSpecCoverageAxisReport{}, false, err
	}
	expected := spec.CurrentExecution.LastCompletedTarget.Result.Summary
	actual := coverageAxisSummaryMap(axisReport.Summary)
	report := mainlineSpecCoverageAxisReport{
		SchemaVersion: 1,
		Status:        registry.StatusPass,
		SpecPath:      specPath,
		Expected:      expected,
		Actual:        actual,
	}
	if len(expected) == 0 {
		report.Status = registry.StatusFail
		report.Issues = append(report.Issues, mainlineSpecCoverageAxisIssue{Field: "summary", Message: "current_execution.last_completed_target.result.summary is missing"})
		return report, true, nil
	}
	for _, field := range coverageAxisSummaryFields() {
		if expected[field] != actual[field] {
			report.Issues = append(report.Issues, mainlineSpecCoverageAxisIssue{
				Field:    field,
				Expected: expected[field],
				Actual:   actual[field],
				Message:  fmt.Sprintf("mainline spec summary %s=%d does not match coverage axis %d", field, expected[field], actual[field]),
			})
		}
	}
	if len(report.Issues) > 0 {
		report.Status = registry.StatusFail
	}
	return report, true, nil
}

func coverageAxisSummaryFields() []string {
	return []string{
		"atom_rows",
		"writer_full_axis_rows",
		"writer_partial_rows",
		"boundary_rows",
		"host_full_axis_rows",
		"host_missing_rows",
		"source_target_pairs",
		"cells",
		"pass",
		"blocked",
		"failed",
		"skipped",
	}
}

func coverageAxisSummaryMap(summary registry.CoverageAxisSummary) map[string]int {
	return map[string]int{
		"atom_rows":             summary.AtomRows,
		"writer_full_axis_rows": summary.WriterFullAxisRows,
		"writer_partial_rows":   summary.WriterPartialRows,
		"boundary_rows":         summary.BoundaryRows,
		"host_full_axis_rows":   summary.HostFullAxisRows,
		"host_missing_rows":     summary.HostMissingRows,
		"source_target_pairs":   summary.SourceTargetPairs,
		"cells":                 summary.Cells,
		"pass":                  summary.Pass,
		"blocked":               summary.Blocked,
		"failed":                summary.Failed,
		"skipped":               summary.Skipped,
	}
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
	layout, err := registry.LayoutRepository(root, registry.LayoutOptions{
		SampleLimit:         20,
		StateReferenceFiles: defaultCleanupStateReferenceFiles(),
	})
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

	consistencyOut := "tmp/registry_layout_cleanup_consistency.json"
	consistency := checkLayoutCleanupConsistency(layout, cleanup)
	if err := writeJSONFile(filepath.Join(root, filepath.FromSlash(consistencyOut)), consistency); err != nil {
		return fmt.Errorf("write layout cleanup consistency: %w", err)
	}
	addStep(gateStepReport{ID: "registry_layout_cleanup_consistency", Command: "go run ./cmd/aepregistry gate -root . -out tmp/registry_asset_gate.json -scope asset-policy", Output: consistencyOut, Status: consistency.Status, Errors: len(consistency.Issues)})

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

func checkLayoutCleanupConsistency(layout registry.LayoutReport, cleanup registry.GeneratedCleanupReport) layoutCleanupConsistencyReport {
	layoutValues := map[string]int{
		"cleanup_candidate_files":        layout.Summary.CleanupCandidateFiles,
		"direct_cleanup_candidate_files": layout.Summary.DirectCleanupCandidateFiles,
		"review_prunable_files":          layout.Summary.ReviewPrunableFiles,
		"registry_report_files":          layout.Summary.RegistryReportFiles,
	}
	cleanupValues := map[string]int{
		"cleanup_candidate_files":        cleanup.Summary.CleanupCandidateFiles,
		"direct_cleanup_candidate_files": cleanup.Summary.DirectCleanupCandidateFiles,
		"review_prunable_files":          cleanup.Summary.ReviewPrunableFiles,
		"registry_report_files":          cleanup.Summary.RegistryReportFiles,
	}
	report := layoutCleanupConsistencyReport{
		SchemaVersion: 1,
		Status:        registry.StatusPass,
		Layout:        layoutValues,
		Cleanup:       cleanupValues,
	}
	for _, field := range []string{"cleanup_candidate_files", "direct_cleanup_candidate_files", "review_prunable_files"} {
		if layoutValues[field] != cleanupValues[field] {
			report.Issues = append(report.Issues, layoutCleanupConsistencyIssue{
				Field:   field,
				Layout:  layoutValues[field],
				Cleanup: cleanupValues[field],
				Message: fmt.Sprintf("layout %s=%d does not match generated cleanup %d", field, layoutValues[field], cleanupValues[field]),
			})
		}
	}
	if len(report.Issues) > 0 {
		report.Status = registry.StatusFail
	}
	return report
}

func runBoundaries(args []string) int {
	fs := flag.NewFlagSet("aepregistry boundaries", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	root := fs.String("root", ".", "repository root")
	coveragePath := fs.String("coverage", "flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json", "coverage ledger JSON path")
	outPath := fs.String("out", "tmp/registry_version_boundaries.json", "version boundary check report JSON path")
	versions := fs.String("versions", "", "comma-separated AE versions; defaults to "+aeversion.SupportedRange())
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: aepregistry boundaries [-root .] [-coverage flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json] [-out tmp/registry_version_boundaries.json] [-versions AE2020,AE2021,...] [-json]")
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
	if *apply && *pruneReviewSiblings && len(splitCSV(*producers)) == 0 {
		fmt.Fprintln(os.Stderr, "cleanup: -apply with -prune-review-siblings requires -producer to scope destructive sibling pruning")
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
		"flightdeck/work/versioned-aep-migration/versioned-aep-migration-current.json",
		"flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json",
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
	coveragePath := fs.String("coverage", "flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json", "coverage ledger JSON path")
	outPath := fs.String("out", "tmp/registry_coverage.json", "coverage validation report JSON path")
	summaryOut := fs.Bool("summary", false, "write summary JSON instead of full coverage validation report")
	rowsOut := fs.Bool("rows", false, "write filtered atom rows JSON instead of full coverage validation report")
	cellsOut := fs.Bool("cells", false, "write filtered source-target matrix cells JSON instead of full coverage validation report")
	axisOut := fs.Bool("axis", false, "write AE version-axis coverage summary JSON instead of full coverage validation report")
	requireLedgers := fs.Bool("require-ledgers", false, "fail when declared coverage ledgers are missing")
	versions := fs.String("versions", "", "comma-separated AE versions for -axis; defaults to "+aeversion.SupportedRange())
	recordFilter := fs.String("record", "", "filter rows or cells by coverage record id")
	domainFilter := fs.String("domain", "", "filter rows, cells, or axis by atom domain")
	atomFilter := fs.String("atom", "", "filter rows or cells by atom id")
	recipeFilter := fs.String("recipe", "", "filter cells by recipe id")
	caseStatusFilter := fs.String("case-status", "", "filter cells by matrix case status")
	writerStatusFilter := fs.String("writer-status", "", "filter rows or cells by writer status")
	writerAxisStatusFilter := fs.String("writer-axis-status", "", "filter axis rows by writer axis status")
	hostLevelFilter := fs.String("host-level", "", "filter rows or cells by host-open evidence level")
	hostAxisStatusFilter := fs.String("host-axis-status", "", "filter axis rows by host axis status")
	missingSourceVersionFilter := fs.String("missing-source-version", "", "filter axis rows missing a source AE version")
	missingTargetVersionFilter := fs.String("missing-target-version", "", "filter axis rows missing a target AE version")
	missingHostVersionFilter := fs.String("missing-host-version", "", "filter axis rows missing host-open evidence for an AE version")
	boundaryStatusFilter := fs.String("boundary-status", "", "filter rows or cells by boundary status")
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: aepregistry coverage [-root .] [-coverage flightdeck/work/versioned-aep-migration/versioned-aep-migration-coverage.json] [-out tmp/registry_coverage.json] [-summary|-rows|-cells|-axis] [-require-ledgers] [-versions AE2020,AE2021,...] [-record id] [-domain name] [-atom id] [-recipe id] [-case-status status] [-writer-status status] [-writer-axis-status status] [-host-level level] [-host-axis-status status] [-missing-source-version AE2025] [-missing-target-version AE2025] [-missing-host-version AE2025] [-boundary-status status] [-json]")
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
			Domain:                *domainFilter,
			AtomID:                *atomFilter,
			Recipe:                *recipeFilter,
			CaseStatus:            *caseStatusFilter,
			WriterStatus:          *writerStatusFilter,
			WriterAxisStatus:      *writerAxisStatusFilter,
			HostOpenEvidenceLevel: *hostLevelFilter,
			HostAxisStatus:        *hostAxisStatusFilter,
			MissingSourceVersion:  *missingSourceVersionFilter,
			MissingTargetVersion:  *missingTargetVersionFilter,
			MissingHostVersion:    *missingHostVersionFilter,
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
			Domain:                *domainFilter,
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
		report, err = registry.ValidateCoverageWithOptions(*root, *coveragePath, registry.CoverageValidationOptions{
			RequireLedgers:    *requireLedgers,
			RequireAllRecipes: *requireLedgers,
		})
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
				Domain:                *domainFilter,
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
	stateRefs := fs.String("state-ref", strings.Join(defaultCleanupStateReferenceFiles(), ","), "comma-separated JSON state files whose tmp references protect generated cleanup groups")
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "usage: aepregistry layout [-root .] [-out tmp/registry_layout.json] [-state-ref json[,json...]] [-sample-limit 20] [-json]")
		return 2
	}

	report, err := registry.LayoutRepository(*root, registry.LayoutOptions{
		SampleLimit:         *sampleLimit,
		StateReferenceFiles: splitCSV(*stateRefs),
	})
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

func writeTextFile(path, value string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(value), 0o644)
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: aepregistry <audit|boundaries|checkpoint|cleanup|coverage|coverage-batch|coverage-md|coverage-update|current|gate|host-open-gaps|inventory|layout|migration-summary|ownership|recurring-matrix> [flags]")
}
