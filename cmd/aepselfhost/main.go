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
	OutRoot string
	Limit   int
	Open    bool
	DryRun  bool
}

func parseVerifyOptions(name string, args []string, stderr io.Writer) (verifyOptions, bool) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	opts := verifyOptions{}
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
	args := psVerifyArgs(opts)
	if opts.DryRun {
		fmt.Fprint(stdout, selfhost.FormatVerifyDryRun(selfhost.VerifyOptions{
			OutRoot: opts.OutRoot,
			Limit:   opts.Limit,
			Open:    opts.Open,
		}))
		return 0
	}
	return runCommand(stdout, stderr, platform, "pwsh", args...)
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

func psVerifyArgs(opts verifyOptions) []string {
	args := []string{"-NoProfile", "-File", filepath.Join("scripts", "verify_technique_selfhost.ps1"), "-OutRoot", opts.OutRoot}
	if opts.Limit > 0 {
		args = append(args, "-Limit", strconv.Itoa(opts.Limit))
	}
	if opts.Open {
		args = append(args, "-Open")
	}
	return args
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
	fmt.Fprintln(stderr, "usage: aepselfhost <verify|compare-reports|recipe-smoke|outcome|status|watch|start-watch> -out-root tmp\\technique_selfhost_gate")
}
