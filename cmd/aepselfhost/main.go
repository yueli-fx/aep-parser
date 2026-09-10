package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/yueli-fx/aep-parser/internal/host"
	"github.com/yueli-fx/aep-parser/internal/selfhost"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, host.DefaultPlatform()))
}

func run(args []string, stdout, stderr io.Writer, platform host.Platform) int {
	if len(args) == 0 {
		usage(stderr)
		return 2
	}
	switch args[0] {
	case "verify":
		return runVerify(args[1:], stdout, stderr, platform)
	case "compare-reports":
		return runCompareReports(args[1:], stdout, stderr)
	case "recipe-smoke":
		return runRecipeSmoke(args[1:], stdout, stderr, platform)
	case "sample-shell":
		return runSampleShell(args[1:], stdout, stderr)
	case "sample-shell-batch":
		return runSampleShellBatch(args[1:], stdout, stderr)
	case "extract-effect-templates":
		return runExtractEffectTemplates(args[1:], stdout, stderr)
	case "audit-effect-templates":
		return runAuditEffectTemplates(args[1:], stdout, stderr)
	case "effect-field-inventory":
		return runEffectFieldInventory(args[1:], stdout, stderr)
	case "effect-field-understanding":
		return runEffectFieldUnderstanding(args[1:], stdout, stderr)
	case "pseudo-controller-rebuild-proof":
		return runPseudoControllerRebuildProof(args[1:], stdout, stderr)
	case "pseudo-behavior-wiring-plan":
		return runPseudoBehaviorWiringPlan(args[1:], stdout, stderr)
	case "pseudo-behavior-payload-extract":
		return runPseudoBehaviorPayloadExtract(args[1:], stdout, stderr)
	case "pseudo-behavior-application":
		return runPseudoBehaviorApplication(args[1:], stdout, stderr)
	case "verify-report":
		return runVerifyReport(args[1:], stdout, stderr)
	case "technique-report":
		return runTechniqueReport(args[1:], stdout, stderr, platform)
	case "finalize-run":
		return runFinalizeRun(args[1:], stdout, stderr)
	case "outcome":
		return runOutcome(args[1:], stdout, stderr)
	case "status":
		return runStatus(args[1:], stdout, stderr, platform)
	case "watch":
		return runWatch(args[1:], stdout, stderr, platform)
	case "start-watch":
		return runStartWatch(args[1:], stdout, stderr, platform)
	default:
		usage(stderr)
		return 2
	}
}

func runSampleShell(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("aepselfhost sample-shell", flag.ContinueOnError)
	fs.SetOutput(stderr)
	inputPath := fs.String("input", "", "input .aep sample path")
	outDir := fs.String("out", filepath.Join("tmp", "sample_shell"), "output directory")
	targetVersion := fs.String("target", "AE2020", "recipe target version")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *inputPath == "" {
		fmt.Fprintln(stderr, "usage: aepselfhost sample-shell -input sample.aep [-out dir] [-target AE2020]")
		return 2
	}
	result, err := selfhost.RunSampleShell(selfhost.SampleShellOptions{
		InputPath:     *inputPath,
		OutDir:        *outDir,
		TargetVersion: *targetVersion,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "recipe:        %s\n", result.RecipePath)
	fmt.Fprintf(stdout, "aep:           %s\n", result.CompiledAEPPath)
	fmt.Fprintf(stdout, "semantic diff: %s\n", result.SemanticDiffPath)
	fmt.Fprintf(stdout, "raw diff:      %s\n", result.RawDiffPath)
	fmt.Fprintf(stdout, "counts: comps %d/%d layers %d/%d shape_layers %d/%d effects %d/%d effect_params %d/%d\n",
		result.SemanticDiff.Summary.CompCount.Generated,
		result.SemanticDiff.Summary.CompCount.Original,
		result.SemanticDiff.Summary.LayerCount.Generated,
		result.SemanticDiff.Summary.LayerCount.Original,
		result.SemanticDiff.Summary.ShapeLayerCount.Generated,
		result.SemanticDiff.Summary.ShapeLayerCount.Original,
		result.SemanticDiff.Summary.EffectCount.Generated,
		result.SemanticDiff.Summary.EffectCount.Original,
		result.SemanticDiff.Summary.EffectParamCount.Generated,
		result.SemanticDiff.Summary.EffectParamCount.Original,
	)
	return 0
}

func runSampleShellBatch(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("aepselfhost sample-shell-batch", flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := fs.String("root", filepath.Join("data", "samples"), "root directory containing .aep samples")
	outDir := fs.String("out", filepath.Join("tmp", "sample_shell_batch"), "output directory")
	targetVersion := fs.String("target", "AE2020", "recipe target version")
	limit := fs.Int("limit", 0, "maximum number of .aep samples to process; 0 means all")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	result, err := selfhost.RunSampleShellBatch(selfhost.SampleShellBatchOptions{
		Root:          *root,
		OutDir:        *outDir,
		TargetVersion: *targetVersion,
		Limit:         *limit,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "batch summary: %s\n", result.SummaryPath)
	fmt.Fprintf(stdout, "samples: total %d succeeded %d failed %d\n", result.Summary.Total, result.Summary.Succeeded, result.Summary.Failed)
	fmt.Fprintf(stdout, "counts: comps %d/%d layers %d/%d shape_layers %d/%d effects %d/%d effect_params %d/%d footage %d/%d raw_diff %d\n",
		result.Summary.CompCount.Generated,
		result.Summary.CompCount.Original,
		result.Summary.LayerCount.Generated,
		result.Summary.LayerCount.Original,
		result.Summary.ShapeLayerCount.Generated,
		result.Summary.ShapeLayerCount.Original,
		result.Summary.EffectCount.Generated,
		result.Summary.EffectCount.Original,
		result.Summary.EffectParamCount.Generated,
		result.Summary.EffectParamCount.Original,
		result.Summary.FootageCount.Generated,
		result.Summary.FootageCount.Original,
		result.Summary.RawProfileDiffCount,
	)
	return 0
}

func runExtractEffectTemplates(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("aepselfhost extract-effect-templates", flag.ContinueOnError)
	fs.SetOutput(stderr)
	summaryPath := fs.String("summary", filepath.Join("tmp", "sample_shell_batch_full", "batch_summary.json"), "sample-shell batch_summary.json")
	root := fs.String("root", "", "root directory containing .aep samples; defaults to summary.root")
	outDir := fs.String("out", filepath.Join("tmp", "effect_template_candidates"), "candidate template output directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	result, err := selfhost.RunSampleShellEffectTemplateExtraction(selfhost.SampleShellEffectTemplateExtractionOptions{
		SummaryPath: *summaryPath,
		Root:        *root,
		OutDir:      *outDir,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "extraction summary: %s\n", result.ResultPath)
	fmt.Fprintf(stdout, "requested %d hit %d missing %d\n", len(result.Requested), len(result.Hits), len(result.Missing))
	return 0
}

func runAuditEffectTemplates(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("aepselfhost audit-effect-templates", flag.ContinueOnError)
	fs.SetOutput(stderr)
	candidatesDir := fs.String("candidates", filepath.Join("tmp", "effect_template_candidates"), "candidate template directory")
	outPath := fs.String("out", "", "audit output path; defaults to <candidates>/candidate_audit.json")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	result, err := selfhost.AuditSampleShellEffectTemplateCandidates(selfhost.SampleShellEffectTemplateAuditOptions{
		CandidatesDir: *candidatesDir,
		OutPath:       *outPath,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "candidate audit: %s\n", result.ResultPath)
	fmt.Fprintf(stdout, "total %d candidate %d review %d reject %d parse_error %d\n",
		result.Summary.Total,
		result.Summary.Candidate,
		result.Summary.Review,
		result.Summary.Reject,
		result.Summary.ParseError,
	)
	return 0
}

func runEffectFieldInventory(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("aepselfhost effect-field-inventory", flag.ContinueOnError)
	fs.SetOutput(stderr)
	root := fs.String("root", filepath.Join("data", "samples"), "root directory containing .aep samples")
	outPath := fs.String("out", filepath.Join("tmp", "effect_field_inventory", "inventory.json"), "effect field inventory JSON path")
	limit := fs.Int("limit", 0, "maximum number of .aep samples to process; 0 means all")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	result, err := selfhost.RunEffectFieldInventory(selfhost.EffectFieldInventoryOptions{
		Root:    *root,
		OutPath: *outPath,
		Limit:   *limit,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "effect field inventory: %s\n", result.OutputPath)
	fmt.Fprintf(stdout, "projects: %d failed %d effects %d/%d params %d/%d\n",
		result.Summary.ProjectCount,
		result.Summary.FailedProjects,
		result.Summary.EffectKinds,
		result.Summary.EffectOccurrences,
		result.Summary.ParamKinds,
		result.Summary.ParamOccurrences,
	)
	return 0
}

func runEffectFieldUnderstanding(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("aepselfhost effect-field-understanding", flag.ContinueOnError)
	fs.SetOutput(stderr)
	inventoryPath := fs.String("inventory", filepath.Join("tmp", "effect_field_inventory", "inventory.json"), "effect field inventory JSON path")
	outPath := fs.String("out", filepath.Join("tmp", "effect_field_understanding", "understanding.json"), "effect field understanding JSON path")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	result, err := selfhost.RunEffectFieldUnderstanding(selfhost.EffectFieldUnderstandingOptions{
		InventoryPath: *inventoryPath,
		OutPath:       *outPath,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "effect field understanding: %s\n", result.OutputPath)
	fmt.Fprintf(stdout, "effects %d/%d params %d/%d\n",
		result.Summary.EffectKinds,
		result.Summary.EffectOccurrences,
		result.Summary.ParamKinds,
		result.Summary.ParamOccurrences,
	)
	return 0
}

func runPseudoControllerRebuildProof(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("aepselfhost pseudo-controller-rebuild-proof", flag.ContinueOnError)
	fs.SetOutput(stderr)
	inventoryPath := fs.String("inventory", filepath.Join("tmp", "effect_field_inventory", "inventory.json"), "effect field inventory JSON path")
	understandingPath := fs.String("understanding", filepath.Join("tmp", "effect_field_understanding", "understanding.json"), "effect field understanding JSON path")
	outDir := fs.String("out", filepath.Join("tmp", "pseudo_controller_rebuild"), "pseudo controller rebuild proof output directory")
	maxFamilies := fs.Int("max", 10, "maximum pseudo families to plan")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	result, err := selfhost.RunPseudoControllerRebuildProof(selfhost.PseudoControllerRebuildOptions{
		InventoryPath:     *inventoryPath,
		UnderstandingPath: *understandingPath,
		OutDir:            *outDir,
		MaxFamilies:       *maxFamilies,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "pseudo controller rebuild proof: %s\n", result.OutputPath)
	fmt.Fprintf(stdout, "families: pseudo %d selected %d generated %d controls %d unsupported %d\n",
		result.Summary.PseudoFamilies,
		result.Summary.SelectedFamilies,
		result.Summary.GeneratedFamilies,
		result.Summary.GeneratedControls,
		result.Summary.UnsupportedControlCount,
	)
	if len(result.Families) > 0 && result.Families[0].GeneratedAEP != "" {
		fmt.Fprintf(stdout, "generated aep: %s\n", result.Families[0].GeneratedAEP)
	}
	return 0
}

func runPseudoBehaviorWiringPlan(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("aepselfhost pseudo-behavior-wiring-plan", flag.ContinueOnError)
	fs.SetOutput(stderr)
	proofPath := fs.String("proof", filepath.Join("tmp", "pseudo_controller_rebuild", "proof.json"), "pseudo controller rebuild proof JSON path")
	outDir := fs.String("out", filepath.Join("tmp", "pseudo_behavior_wiring"), "pseudo behavior wiring output directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	result, err := selfhost.RunPseudoBehaviorWiringPlan(selfhost.PseudoBehaviorWiringOptions{
		ProofPath: *proofPath,
		OutDir:    *outDir,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "pseudo behavior wiring plan: %s\n", result.OutputPath)
	fmt.Fprintf(stdout, "families: %d controls %d static %d keyframe %d expression %d conflicts %d\n",
		result.Summary.Families,
		result.Summary.Controls,
		result.Summary.StaticControls,
		result.Summary.KeyframeTasks,
		result.Summary.ExpressionTasks,
		result.Summary.ConflictTasks,
	)
	return 0
}

func runPseudoBehaviorPayloadExtract(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("aepselfhost pseudo-behavior-payload-extract", flag.ContinueOnError)
	fs.SetOutput(stderr)
	planPath := fs.String("plan", filepath.Join("tmp", "pseudo_behavior_wiring", "plan.json"), "pseudo behavior wiring plan JSON path")
	root := fs.String("root", filepath.Join("data", "samples"), "root directory containing .aep samples")
	outDir := fs.String("out", filepath.Join("tmp", "pseudo_behavior_payloads"), "pseudo behavior payload output directory")
	limit := fs.Int("limit", 0, "maximum number of .aep samples to process; 0 means all")
	maxExamples := fs.Int("max-examples", 5, "maximum payload examples per control")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	result, err := selfhost.RunPseudoBehaviorPayloadExtraction(selfhost.PseudoBehaviorPayloadOptions{
		PlanPath:    *planPath,
		Root:        *root,
		OutDir:      *outDir,
		Limit:       *limit,
		MaxExamples: *maxExamples,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "pseudo behavior payloads: %s\n", result.OutputPath)
	fmt.Fprintf(stdout, "families: %d controls %d behavior %d payloads %d missing %d keyframe %d expression %d projects %d failed %d\n",
		result.Summary.PlannedFamilies,
		result.Summary.PlannedControls,
		result.Summary.BehaviorTasks,
		result.Summary.ControlsWithPayloads,
		result.Summary.ControlsMissingPayloads,
		result.Summary.KeyframePayloads,
		result.Summary.ExpressionPayloads,
		result.Summary.ScannedProjects,
		result.Summary.FailedProjects,
	)
	return 0
}

func runPseudoBehaviorApplication(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("aepselfhost pseudo-behavior-application", flag.ContinueOnError)
	fs.SetOutput(stderr)
	proofPath := fs.String("proof", filepath.Join("tmp", "pseudo_controller_rebuild", "proof.json"), "pseudo controller rebuild proof JSON path")
	payloadPath := fs.String("payloads", filepath.Join("tmp", "pseudo_behavior_payloads", "payloads.json"), "pseudo behavior payload JSON path")
	outDir := fs.String("out", filepath.Join("tmp", "pseudo_behavior_application"), "pseudo behavior application output directory")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	result, err := selfhost.RunPseudoBehaviorApplication(selfhost.PseudoBehaviorApplicationOptions{
		ProofPath:   *proofPath,
		PayloadPath: *payloadPath,
		OutDir:      *outDir,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "pseudo behavior application: %s\n", result.OutputPath)
	fmt.Fprintf(stdout, "families: %d generated %d skipped %d aeps %d applied %d verified %d deferred_expression %d errors %d\n",
		result.Summary.Families,
		result.Summary.GeneratedFamilies,
		result.Summary.SkippedFamilies,
		result.Summary.GeneratedAEPs,
		result.Summary.AppliedControls,
		result.Summary.VerifiedKeyframedControls,
		result.Summary.ExpressionDeferredControls,
		result.Summary.ErrorControls,
	)
	return 0
}

func runCompareReports(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("aepselfhost compare-reports", flag.ContinueOnError)
	fs.SetOutput(stderr)
	baseDir := fs.String("base", "", "base technique report directory")
	newDir := fs.String("new", "", "new technique report directory")
	outDir := fs.String("out", "", "output directory; defaults to <new>/compare")
	top := fs.Int("top", 20, "maximum count diffs per group")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *baseDir == "" || *newDir == "" {
		fmt.Fprintln(stderr, "usage: aepselfhost compare-reports -base old_report -new new_report [-out dir] [-top n]")
		return 2
	}
	result, err := selfhost.CompareReports(selfhost.CompareOptions{
		BaseDir: *baseDir,
		NewDir:  *newDir,
		OutDir:  *outDir,
		Top:     *top,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	reportOutDir := *outDir
	if reportOutDir == "" {
		reportOutDir = filepath.Join(*newDir, "compare")
	}
	_ = result
	fmt.Fprintf(stdout, "compare json: %s\n", filepath.Join(reportOutDir, "compare.json"))
	fmt.Fprintf(stdout, "compare md:   %s\n", filepath.Join(reportOutDir, "compare.md"))
	return 0
}

func runTechniqueReport(args []string, stdout, stderr io.Writer, platform host.Platform) int {
	fs := flag.NewFlagSet("aepselfhost technique-report", flag.ContinueOnError)
	fs.SetOutput(stderr)
	inputPath := fs.String("input", filepath.Join("data", "samples"), "input project file or corpus directory")
	outDir := fs.String("out", filepath.Join("tmp", "technique_showcase_report"), "report output directory")
	limit := fs.Int("limit", 0, "optional corpus limit")
	verify := fs.Bool("verify", false, "verify generated report artifacts")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if err := generateTechniqueReport(context.Background(), platform.Runner, *inputPath, *outDir, *limit, *verify, stdout, stderr); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "summary: %s\n", filepath.Join(*outDir, "summary.json"))
	fmt.Fprintf(stdout, "corpus:  %s\n", filepath.Join(*outDir, "corpus.jsonl"))
	fmt.Fprintf(stdout, "manifest: %s\n", filepath.Join(*outDir, "manifest.json"))
	fmt.Fprintf(stdout, "html:    %s\n", filepath.Join(*outDir, "report.html"))
	return 0
}

func runFinalizeRun(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("aepselfhost finalize-run", flag.ContinueOnError)
	fs.SetOutput(stderr)
	outRoot := fs.String("out-root", filepath.Join("tmp", "technique_selfhost_gate"), "selfhost output root")
	runRoot := fs.String("run-root", "", "specific selfhost run root")
	runID := fs.String("run-id", "", "run identifier; defaults to run root base name")
	inputPath := fs.String("input", "", "source corpus input path")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *runRoot == "" {
		fmt.Fprintln(stderr, "usage: aepselfhost finalize-run -out-root root -run-root root/run-id [-input path]")
		return 2
	}
	result, err := selfhost.FinalizeSelfhostRun(context.Background(), selfhost.FinalizeOptions{
		OutRoot:   *outRoot,
		RunRoot:   *runRoot,
		RunID:     *runID,
		InputPath: *inputPath,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "latest index: %s\n", result.LatestIndex)
	fmt.Fprintf(stdout, "outcome: %s\n", result.OutcomeStatus)
	return 0
}

func runVerifyReport(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("aepselfhost verify-report", flag.ContinueOnError)
	fs.SetOutput(stderr)
	outDir := fs.String("out-dir", filepath.Join("tmp", "technique_showcase_report"), "technique report output directory")
	minProjects := fs.Int("min-projects", 1, "minimum parsed project count")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	result, err := selfhost.VerifyTechniqueReport(selfhost.ReportVerifyOptions{
		OutDir:      *outDir,
		MinProjects: *minProjects,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "ok: %s\n", *outDir)
	fmt.Fprintf(stdout, "projects: %d\n", result.ProjectCount)
	fmt.Fprintf(stdout, "patterns: %d\n", result.PatternCount)
	return 0
}

func runRecipeSmoke(args []string, stdout, stderr io.Writer, platform host.Platform) int {
	fs := flag.NewFlagSet("aepselfhost recipe-smoke", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fullReportDir := fs.String("full-report", "", "technique full_report directory containing recipe_drafts.jsonl")
	runRoot := fs.String("run-root", "", "selfhost run root; defaults to parent of -full-report")
	batchLimit := fs.Int("batch-limit", 3, "number of recipe draft rows to smoke test")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *fullReportDir == "" {
		fmt.Fprintln(stderr, "usage: aepselfhost recipe-smoke -full-report run/full_report [-run-root run] [-batch-limit n]")
		return 2
	}
	root := *runRoot
	if root == "" {
		root = filepath.Dir(*fullReportDir)
	}
	result, err := selfhost.RunRecipeDraftSmoke(context.Background(), selfhost.RecipeDraftSmokeOptions{
		DraftsPath: filepath.Join(*fullReportDir, "recipe_drafts.jsonl"),
		CompileDir: filepath.Join(root, "recipe_draft_compile"),
		ReparseDir: filepath.Join(root, "recipe_draft_reparse"),
		BatchDir:   filepath.Join(root, "recipe_draft_batch"),
		BatchLimit: *batchLimit,
		Runner:     platform.Runner,
		WorkingDir: ".",
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	fmt.Fprintf(stdout, "recipe draft compile: %s\n", result.Compile.Output)
	fmt.Fprintf(stdout, "recipe draft reparse summary: %s\n", filepath.Join(root, "recipe_draft_reparse", "reparse_summary.json"))
	fmt.Fprintf(stdout, "recipe draft batch summary: %s\n", filepath.Join(root, "recipe_draft_batch", "summary.json"))
	return 0
}

func runOutcome(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("aepselfhost outcome", flag.ContinueOnError)
	fs.SetOutput(stderr)
	outRoot := fs.String("out-root", filepath.Join("tmp", "technique_selfhost_gate"), "selfhost output root")
	jsonPath := fs.String("json-path", "", "explicit selfhost outcome JSON path")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	outcome, path, err := readOutcome(*outRoot, *jsonPath)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	printOutcome(stdout, outcome, path)
	return 0
}

func runStatus(args []string, stdout, stderr io.Writer, platform host.Platform) int {
	fs := flag.NewFlagSet("aepselfhost status", flag.ContinueOnError)
	fs.SetOutput(stderr)
	outRoot := fs.String("out-root", filepath.Join("tmp", "technique_selfhost_gate"), "selfhost output root")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	status, err := buildStatus(*outRoot)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	printStatus(stdout, status, platform.ProcessInspector)
	return 0
}

func runVerify(args []string, stdout, stderr io.Writer, platform host.Platform) int {
	opts, ok := parseVerifyOptions("aepselfhost verify", args, stderr)
	if !ok {
		return 2
	}
	return verifySelfhost(opts, stdout, stderr, platform)
}

func runWatch(args []string, stdout, stderr io.Writer, platform host.Platform) int {
	opts, ok := parseWatchOptions("aepselfhost watch", args, stderr)
	if !ok {
		return 2
	}
	if err := opts.validate(); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return watchSelfhost(opts, stdout, stderr, platform)
}

func runStartWatch(args []string, stdout, stderr io.Writer, platform host.Platform) int {
	opts, ok := parseWatchOptions("aepselfhost start-watch", args, stderr)
	if !ok {
		return 2
	}
	if err := opts.validate(); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return startWatch(opts, stdout, stderr, platform)
}

type verifyOptions struct {
	InputPath string
	OutRoot   string
	Limit     int
	Open      bool
	DryRun    bool
}

func parseVerifyOptions(name string, args []string, stderr io.Writer) (verifyOptions, bool) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	opts := verifyOptions{}
	fs.StringVar(&opts.InputPath, "input", filepath.Join("data", "samples"), "input corpus path")
	fs.StringVar(&opts.OutRoot, "out-root", filepath.Join("tmp", "technique_selfhost_gate"), "selfhost output root")
	fs.IntVar(&opts.Limit, "limit", 0, "optional sample limit")
	fs.BoolVar(&opts.Open, "open", false, "open the generated outcome page")
	fs.BoolVar(&opts.DryRun, "dry-run", false, "print the underlying gate command without running")
	if err := fs.Parse(args); err != nil {
		return verifyOptions{}, false
	}
	return opts, true
}

func verifySelfhost(opts verifyOptions, stdout, stderr io.Writer, platform host.Platform) int {
	if opts.DryRun {
		fmt.Fprint(stdout, selfhost.FormatVerifyDryRun(selfhost.VerifyOptions{
			OutRoot:   opts.OutRoot,
			InputPath: opts.InputPath,
			Limit:     opts.Limit,
			Open:      opts.Open,
		}))
		return 0
	}
	runID := time.Now().UTC().Format("20060102T150405Z")
	runRoot := filepath.Join(opts.OutRoot, runID)
	fullReportDir := filepath.Join(runRoot, "full_report")
	partialInputDir := filepath.Join(runRoot, "partial_input")
	partialReportDir := filepath.Join(runRoot, "partial_report")
	compareSelfDir := filepath.Join(runRoot, "compare_self")
	comparePartialDir := filepath.Join(runRoot, "compare_partial_to_full")
	if err := os.MkdirAll(runRoot, 0o755); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}

	runStep := func(name string, fn func() error) bool {
		start := time.Now()
		err := fn()
		seconds := time.Since(start).Seconds()
		if err != nil {
			fmt.Fprintf(stderr, "%s failed after %.2fs: %v\n", name, seconds, err)
			return false
		}
		fmt.Fprintf(stdout, "%s: exit=0 seconds=%.2f\n", name, seconds)
		return true
	}
	runExternal := func(name string, args ...string) bool {
		return runStep(name, func() error {
			result := platform.Runner.Run(context.Background(), host.Command{
				Name:   args[0],
				Args:   args[1:],
				Stdout: stdout,
				Stderr: stderr,
			})
			if result.ExitCode != 0 {
				return fmt.Errorf("exit code %d", result.ExitCode)
			}
			return nil
		})
	}

	if !runExternal("go technique tests", "go", "test", "./cmd/aeptechnique", "./internal/technique", "-count=1") {
		return 1
	}
	inputPath := resolveRepoPath(opts.InputPath)
	if !runStep("full technique report", func() error {
		return generateTechniqueReport(context.Background(), platform.Runner, inputPath, fullReportDir, opts.Limit, true, stdout, stderr)
	}) {
		return 1
	}
	if err := os.MkdirAll(partialInputDir, 0o755); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := copyFile(resolveRepoPath(filepath.Join("test_data", "fixtures", "showcase", "text.aep")), filepath.Join(partialInputDir, "good.aep")); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if err := os.WriteFile(filepath.Join(partialInputDir, "bad.aep"), []byte("not an aep"), 0o644); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if !runStep("partial-error technique report", func() error {
		return generateTechniqueReport(context.Background(), platform.Runner, partialInputDir, partialReportDir, 0, true, stdout, stderr)
	}) {
		return 1
	}
	if !runStep("self compare", func() error {
		_, err := selfhost.CompareReports(selfhost.CompareOptions{BaseDir: fullReportDir, NewDir: fullReportDir, OutDir: compareSelfDir})
		return err
	}) {
		return 1
	}
	if !runStep("partial-to-full compare", func() error {
		_, err := selfhost.CompareReports(selfhost.CompareOptions{BaseDir: partialReportDir, NewDir: fullReportDir, OutDir: comparePartialDir, Top: 5})
		return err
	}) {
		return 1
	}
	if !runStep("recipe draft smoke", func() error {
		_, err := selfhost.RunRecipeDraftSmoke(context.Background(), selfhost.RecipeDraftSmokeOptions{
			DraftsPath: filepath.Join(fullReportDir, "recipe_drafts.jsonl"),
			CompileDir: filepath.Join(runRoot, "recipe_draft_compile"),
			ReparseDir: filepath.Join(runRoot, "recipe_draft_reparse"),
			BatchDir:   filepath.Join(runRoot, "recipe_draft_batch"),
			BatchLimit: 3,
			Runner:     platform.Runner,
			WorkingDir: ".",
		})
		return err
	}) {
		return 1
	}
	if !runStep("finalize selfhost run", func() error {
		_, err := selfhost.FinalizeSelfhostRun(context.Background(), selfhost.FinalizeOptions{
			OutRoot:   opts.OutRoot,
			RunRoot:   runRoot,
			RunID:     runID,
			InputPath: inputPath,
		})
		return err
	}) {
		return 1
	}
	fmt.Fprintf(stdout, "latest index:    %s\n", filepath.Join(opts.OutRoot, "latest_index.html"))
	fmt.Fprintf(stdout, "latest outcome:  %s\n", filepath.Join(opts.OutRoot, "latest_outcome.md"))
	if opts.Open {
		if err := platform.BrowserOpener.Open(context.Background(), filepath.Join(opts.OutRoot, "latest_outcome.html")); err != nil {
			fmt.Fprintf(stderr, "open outcome: %v\n", err)
		}
	}
	return 0
}

type watchOptions struct {
	OutRoot         string
	DurationMinutes int
	IntervalSeconds int
	Iterations      int
	Limit           int
	OpenFirst       bool
	DryRun          bool
}

func parseWatchOptions(name string, args []string, stderr io.Writer) (watchOptions, bool) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	opts := watchOptions{}
	fs.StringVar(&opts.OutRoot, "out-root", filepath.Join("tmp", "technique_selfhost_gate"), "selfhost output root")
	fs.IntVar(&opts.DurationMinutes, "duration-minutes", 60, "watch duration in minutes")
	fs.IntVar(&opts.IntervalSeconds, "interval-seconds", 300, "seconds between watch iterations")
	fs.IntVar(&opts.Iterations, "iterations", 0, "fixed number of watch iterations")
	fs.IntVar(&opts.Limit, "limit", 0, "optional sample limit for each verification pass")
	fs.BoolVar(&opts.OpenFirst, "open-first", false, "open the first verification outcome")
	fs.BoolVar(&opts.DryRun, "dry-run", false, "print commands and write status without running")
	if err := fs.Parse(args); err != nil {
		return watchOptions{}, false
	}
	return opts, true
}

func (opts watchOptions) validate() error {
	if opts.DurationMinutes < 0 {
		return fmt.Errorf("DurationMinutes must be >= 0")
	}
	if opts.IntervalSeconds < 1 {
		return fmt.Errorf("IntervalSeconds must be >= 1")
	}
	if opts.Iterations < 0 {
		return fmt.Errorf("Iterations must be >= 0")
	}
	if !opts.DryRun && opts.Iterations == 0 && opts.DurationMinutes == 0 {
		return fmt.Errorf("Iterations and DurationMinutes cannot both be 0")
	}
	return nil
}

func watchSelfhost(opts watchOptions, stdout, stderr io.Writer, platform host.Platform) int {
	startedAt := utcNow()
	verifyArgs := verifyCLIArgs(opts, opts.OpenFirst)
	showArgs := []string{"outcome", "-out-root", opts.OutRoot}
	if opts.DryRun {
		fmt.Fprintln(stdout, "DRY RUN technique selfhost watch")
		fmt.Fprintf(stdout, "verify command: aepselfhost %s\n", strings.Join(verifyArgs, " "))
		fmt.Fprintf(stdout, "show command:   aepselfhost %s\n", strings.Join(showArgs, " "))
		status := watchFile{
			Mode:              "dry_run",
			StartedAtUTC:      startedAt,
			PlannedIterations: opts.Iterations,
			DurationMinutes:   opts.DurationMinutes,
			IntervalSeconds:   opts.IntervalSeconds,
			VerifyCommand:     "aepselfhost " + strings.Join(verifyArgs, " "),
			ShowCommand:       "aepselfhost " + strings.Join(showArgs, " "),
		}
		if err := writeWatchStatus(opts.OutRoot, status); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	}

	deadline := time.Now().Add(time.Duration(opts.DurationMinutes) * time.Minute)
	iteration := 0
	lastExit := 0
	for {
		if opts.Iterations > 0 && iteration >= opts.Iterations {
			break
		}
		if opts.Iterations == 0 && opts.DurationMinutes > 0 && time.Now().After(deadline) {
			break
		}

		iteration++
		iterationStartedAt := utcNow()
		fmt.Fprintf(stdout, "watch iteration %d started at %s\n", iteration, iterationStartedAt)
		lastExit = verifySelfhost(verifyOptions{
			OutRoot: opts.OutRoot,
			Limit:   opts.Limit,
			Open:    opts.OpenFirst && iteration == 1,
		}, stdout, stderr, platform)
		if lastExit == 0 {
			lastExit = runOutcome([]string{"-out-root", opts.OutRoot}, stdout, stderr)
		}
		status := watchFile{
			Mode:                "watch",
			StartedAtUTC:        startedAt,
			LastIterationAtUTC:  iterationStartedAt,
			PlannedIterations:   opts.Iterations,
			DurationMinutes:     opts.DurationMinutes,
			IntervalSeconds:     opts.IntervalSeconds,
			CompletedIterations: ptrInt(iteration),
			LastExitCode:        ptrInt(lastExit),
			VerifyCommand:       "aepselfhost " + strings.Join(verifyCLIArgs(opts, opts.OpenFirst && iteration == 1), " "),
			ShowCommand:         "aepselfhost outcome -out-root " + opts.OutRoot,
			LatestOutcomeJSON:   filepath.Join(opts.OutRoot, "latest_outcome.json"),
			LatestOutcomeHTML:   filepath.Join(opts.OutRoot, "latest_outcome.html"),
		}
		if err := writeWatchStatus(opts.OutRoot, status); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		if lastExit != 0 {
			return lastExit
		}
		if opts.Iterations > 0 && iteration >= opts.Iterations {
			break
		}
		if opts.Iterations == 0 && opts.DurationMinutes > 0 && time.Now().Add(time.Duration(opts.IntervalSeconds)*time.Second).After(deadline) {
			break
		}
		time.Sleep(time.Duration(opts.IntervalSeconds) * time.Second)
	}
	return lastExit
}

func startWatch(opts watchOptions, stdout, stderr io.Writer, platform host.Platform) int {
	if err := os.MkdirAll(opts.OutRoot, 0o755); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	logDir := filepath.Join(opts.OutRoot, "watch_logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	stamp := time.Now().UTC().Format("20060102T150405Z")
	stdoutLog := filepath.Join(logDir, "watch_"+stamp+".out.log")
	stderrLog := filepath.Join(logDir, "watch_"+stamp+".err.log")
	watchArgs := watchCommandArgs(opts)
	commandLine := "aepselfhost " + strings.Join(watchArgs, " ")
	startedAt := utcNow()
	process := processFile{
		Mode:         "dry_run",
		StartedAtUTC: startedAt,
		Command:      commandLine,
		StdoutLog:    stdoutLog,
		StderrLog:    stderrLog,
		OutRoot:      opts.OutRoot,
	}
	if opts.DryRun {
		fmt.Fprintln(stdout, "DRY RUN technique selfhost background start")
		fmt.Fprintf(stdout, "Start aepselfhost %s\n", strings.Join(watchArgs, " "))
		fmt.Fprintf(stdout, "stdout: %s\n", stdoutLog)
		fmt.Fprintf(stdout, "stderr: %s\n", stderrLog)
		if err := writeProcessStatus(opts.OutRoot, process); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		return 0
	}

	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	outFile, err := os.Create(stdoutLog)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	defer outFile.Close()
	errFile, err := os.Create(stderrLog)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	defer errFile.Close()
	started, err := platform.Runner.Start(context.Background(), host.Command{
		Name:   exe,
		Args:   watchArgs,
		Dir:    mustGetwd(),
		Stdout: outFile,
		Stderr: errFile,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	process.Mode = "started"
	process.PID = started.PID()
	process.WatchStatus = filepath.Join(opts.OutRoot, "watch_status.json")
	process.LatestOutcomeHTML = filepath.Join(opts.OutRoot, "latest_outcome.html")
	process.LatestOutcomeJSON = filepath.Join(opts.OutRoot, "latest_outcome.json")
	if err := writeProcessStatus(opts.OutRoot, process); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	_ = started.Release()
	fmt.Fprintf(stdout, "started technique selfhost watch pid=%d\n", process.PID)
	fmt.Fprintf(stdout, "stdout: %s\n", stdoutLog)
	fmt.Fprintf(stdout, "stderr: %s\n", stderrLog)
	fmt.Fprintf(stdout, "status: %s\n", filepath.Join(opts.OutRoot, "watch_process.json"))
	return 0
}

func verifyCLIArgs(opts watchOptions, open bool) []string {
	args := []string{"verify", "-out-root", opts.OutRoot}
	if opts.Limit > 0 {
		args = append(args, "-limit", strconv.Itoa(opts.Limit))
	}
	if open {
		args = append(args, "-open")
	}
	return args
}

func watchCommandArgs(opts watchOptions) []string {
	args := []string{
		"watch",
		"-out-root", opts.OutRoot,
		"-duration-minutes", strconv.Itoa(opts.DurationMinutes),
		"-interval-seconds", strconv.Itoa(opts.IntervalSeconds),
	}
	if opts.Iterations > 0 {
		args = append(args, "-iterations", strconv.Itoa(opts.Iterations))
	}
	if opts.Limit > 0 {
		args = append(args, "-limit", strconv.Itoa(opts.Limit))
	}
	if opts.OpenFirst {
		args = append(args, "-open-first")
	}
	return args
}

func runCommand(stdout, stderr io.Writer, platform host.Platform, name string, args ...string) int {
	result := platform.Runner.Run(context.Background(), host.Command{
		Name:   name,
		Args:   args,
		Stdout: stdout,
		Stderr: stderr,
	})
	return result.ExitCode
}

func writeWatchStatus(outRoot string, status watchFile) error {
	return writeJSON(filepath.Join(outRoot, "watch_status.json"), status)
}

func writeProcessStatus(outRoot string, status processFile) error {
	return writeJSON(filepath.Join(outRoot, "watch_process.json"), status)
}

func writeJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func utcNow() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}

func ptrInt(v int) *int {
	return &v
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o644)
}

func resolveRepoPath(path string) string {
	if _, err := os.Stat(path); err == nil {
		return path
	}
	parentPath := filepath.Join("..", "..", path)
	if _, err := os.Stat(parentPath); err == nil {
		return parentPath
	}
	return path
}

func generateTechniqueReport(ctx context.Context, runner host.Runner, inputPath, outDir string, limit int, verify bool, stdout, stderr io.Writer) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	corpusPath := filepath.Join(outDir, "corpus.jsonl")
	summaryPath := filepath.Join(outDir, "summary.json")
	args := []string{
		"run", "./cmd/aeptechnique",
		"-in", inputPath,
		"-mode", "explain",
		"-corpus",
		"-recursive",
		"-out", corpusPath,
		"-summary-out", summaryPath,
	}
	if limit > 0 {
		args = append(args, "-limit", strconv.Itoa(limit))
	}
	result := runner.Run(ctx, host.Command{
		Name:   "go",
		Args:   args,
		Stdout: stdout,
		Stderr: stderr,
	})
	if result.ExitCode != 0 {
		var summary struct {
			ErrorCount int `json:"error_count"`
		}
		if err := readJSON(summaryPath, &summary); err != nil || result.ExitCode != 1 || summary.ErrorCount == 0 {
			return fmt.Errorf("aeptechnique corpus failed with exit code %d", result.ExitCode)
		}
	}
	if err := selfhost.RenderTechniqueReportArtifacts(selfhost.TechniqueReportRenderOptions{
		OutDir:    outDir,
		InputPath: inputPath,
		Limit:     limit,
	}); err != nil {
		return err
	}
	if verify {
		_, err := selfhost.VerifyTechniqueReport(selfhost.ReportVerifyOptions{OutDir: outDir, MinProjects: 1})
		return err
	}
	return nil
}

type outcomeFile struct {
	OutcomeStatus struct {
		Status string `json:"status"`
	} `json:"outcome_status"`
	OutcomeSummary struct {
		Headline string `json:"headline"`
	} `json:"outcome_summary"`
	Corpus struct {
		ParsedProjects   int `json:"parsed_projects"`
		TechniquePattern int `json:"technique_patterns"`
		ParseErrors      int `json:"parse_errors"`
	} `json:"corpus"`
	ClosedLoop struct {
		BatchPassed    int `json:"batch_passed"`
		BatchAttempted int `json:"batch_attempted"`
	} `json:"closed_loop"`
	StableOutputs struct {
		OpenTarget string `json:"open_target"`
	} `json:"stable_outputs"`
	ActionPlan struct {
		NextActions []nextAction `json:"next_actions"`
	} `json:"action_plan"`
}

type nextAction struct {
	Priority int    `json:"priority"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
}

type processFile struct {
	Mode              string `json:"mode"`
	StartedAtUTC      string `json:"started_at_utc"`
	PID               int    `json:"pid"`
	Command           string `json:"command"`
	StdoutLog         string `json:"stdout_log"`
	StderrLog         string `json:"stderr_log"`
	OutRoot           string `json:"out_root"`
	WatchStatus       string `json:"watch_status"`
	LatestOutcomeHTML string `json:"latest_outcome_html"`
	LatestOutcomeJSON string `json:"latest_outcome_json"`
}

type watchFile struct {
	Mode                string `json:"mode"`
	StartedAtUTC        string `json:"started_at_utc"`
	LastIterationAtUTC  string `json:"last_iteration_at_utc"`
	PlannedIterations   int    `json:"planned_iterations"`
	DurationMinutes     int    `json:"duration_minutes"`
	IntervalSeconds     int    `json:"interval_seconds"`
	CompletedIterations *int   `json:"completed_iterations"`
	LastExitCode        *int   `json:"last_exit_code"`
	VerifyCommand       string `json:"verify_command"`
	ShowCommand         string `json:"show_command"`
	LatestOutcomeJSON   string `json:"latest_outcome_json"`
	LatestOutcomeHTML   string `json:"latest_outcome_html"`
}

type jsonSource[T any] struct {
	Path  string
	Value T
	OK    bool
}

type statusView struct {
	OutRoot string
	Process jsonSource[processFile]
	Watch   jsonSource[watchFile]
	Outcome jsonSource[outcomeFile]
}

func readOutcome(outRoot, jsonPath string) (outcomeFile, string, error) {
	path := jsonPath
	if path == "" {
		path = filepath.Join(outRoot, "latest_outcome.json")
	}
	var outcome outcomeFile
	if err := readJSON(path, &outcome); err != nil {
		return outcomeFile{}, path, fmt.Errorf("read %s: %w", path, err)
	}
	return outcome, path, nil
}

func buildStatus(outRoot string) (statusView, error) {
	status := statusView{OutRoot: outRoot}
	status.Process = readFirstJSON[processFile](
		filepath.Join(outRoot, "watch_process.json"),
		filepath.Join(outRoot, "start_dry_run", "watch_process.json"),
	)
	status.Watch = readFirstJSON[watchFile](
		filepath.Join(outRoot, "watch_status.json"),
		filepath.Join(outRoot, "watch_dry_run", "watch_status.json"),
	)
	status.Outcome = readFirstJSON[outcomeFile](
		filepath.Join(outRoot, "latest_outcome.json"),
	)
	return status, nil
}

func readFirstJSON[T any](paths ...string) jsonSource[T] {
	for _, path := range paths {
		var value T
		if err := readJSON(path, &value); err == nil {
			return jsonSource[T]{Path: path, Value: value, OK: true}
		}
	}
	return jsonSource[T]{}
}

func readJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func printOutcome(w io.Writer, outcome outcomeFile, path string) {
	_ = path
	fmt.Fprintln(w, "Self-Hosted Outcome")
	fmt.Fprintf(w, "Headline: %s\n", value(outcome.OutcomeSummary.Headline))
	fmt.Fprintf(w, "Status:   %s\n", value(outcome.OutcomeStatus.Status))
	fmt.Fprintf(w, "Corpus:   %d projects, %d patterns, %d parse errors\n", outcome.Corpus.ParsedProjects, outcome.Corpus.TechniquePattern, outcome.Corpus.ParseErrors)
	fmt.Fprintf(w, "Smoke:    %d/%d\n", outcome.ClosedLoop.BatchPassed, outcome.ClosedLoop.BatchAttempted)
	fmt.Fprintf(w, "Open:     %s\n", value(outcome.StableOutputs.OpenTarget))
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Next Actions")
	if len(outcome.ActionPlan.NextActions) == 0 {
		fmt.Fprintln(w, "- none")
		return
	}
	for _, action := range outcome.ActionPlan.NextActions {
		fmt.Fprintf(w, "- P%d: %s - %s\n", action.Priority, value(action.Title), value(action.Detail))
	}
}

func printStatus(w io.Writer, status statusView, inspector host.ProcessInspector) {
	fmt.Fprintln(w, "Self-Hosted Status")
	fmt.Fprintf(w, "OutRoot: %s\n\n", status.OutRoot)
	fmt.Fprintln(w, "Watch Process")
	if !status.Process.OK {
		fmt.Fprintln(w, "- process status: missing")
	} else {
		process := status.Process.Value
		fmt.Fprintf(w, "- file: %s\n", status.Process.Path)
		fmt.Fprintf(w, "- mode: %s\n", value(process.Mode))
		fmt.Fprintf(w, "- pid: %s\n", intValue(process.PID))
		fmt.Fprintf(w, "- state: %s\n", processState(process, inspector))
		fmt.Fprintf(w, "- command: %s\n", value(process.Command))
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Watch Status")
	if !status.Watch.OK {
		fmt.Fprintln(w, "- watch status: missing")
	} else {
		watch := status.Watch.Value
		fmt.Fprintf(w, "- file: %s\n", status.Watch.Path)
		fmt.Fprintf(w, "- mode: %s\n", value(watch.Mode))
		fmt.Fprintf(w, "- completed iterations: %s\n", ptrIntValue(watch.CompletedIterations))
		fmt.Fprintf(w, "- last exit code: %s\n", ptrIntValue(watch.LastExitCode))
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Latest Outcome")
	if !status.Outcome.OK {
		fmt.Fprintln(w, "- latest outcome: missing")
	} else {
		outcome := status.Outcome.Value
		fmt.Fprintf(w, "- file: %s\n", status.Outcome.Path)
		fmt.Fprintf(w, "- headline: %s\n", value(outcome.OutcomeSummary.Headline))
		fmt.Fprintf(w, "- status: %s\n", value(outcome.OutcomeStatus.Status))
		fmt.Fprintf(w, "- open: %s\n", value(outcome.StableOutputs.OpenTarget))
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Logs")
	if !status.Process.OK {
		fmt.Fprintln(w, "- stdout: n/a")
		fmt.Fprintln(w, "- stderr: n/a")
		return
	}
	fmt.Fprintf(w, "- stdout: %s\n", value(status.Process.Value.StdoutLog))
	fmt.Fprintf(w, "- stderr: %s\n", value(status.Process.Value.StderrLog))
}

func processState(process processFile, inspector host.ProcessInspector) string {
	if process.Mode == "dry_run" {
		return "dry-run"
	}
	if process.PID <= 0 {
		return "n/a"
	}
	if inspector != nil && inspector.IsRunning(process.PID) {
		return "running"
	}
	return "not-running"
}

func value(s string) string {
	if s == "" {
		return "n/a"
	}
	return s
}

func intValue(v int) string {
	if v == 0 {
		return "n/a"
	}
	return strconv.Itoa(v)
}

func ptrIntValue(v *int) string {
	if v == nil {
		return "n/a"
	}
	return strconv.Itoa(*v)
}

func usage(stderr io.Writer) {
	fmt.Fprintln(stderr, "usage: aepselfhost <verify|compare-reports|verify-report|technique-report|recipe-smoke|sample-shell|sample-shell-batch|extract-effect-templates|audit-effect-templates|effect-field-inventory|effect-field-understanding|pseudo-controller-rebuild-proof|pseudo-behavior-wiring-plan|pseudo-behavior-payload-extract|pseudo-behavior-application|finalize-run|outcome|status|watch|start-watch> -out-root tmp\\technique_selfhost_gate")
}
