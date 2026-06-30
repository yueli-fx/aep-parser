package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return 2
	}
	switch args[0] {
	case "outcome":
		return runOutcome(args[1:], stdout, stderr)
	case "status":
		return runStatus(args[1:], stdout, stderr)
	case "watch":
		return runWatch(args[1:], stdout, stderr)
	case "start-watch":
		return runStartWatch(args[1:], stdout, stderr)
	default:
		usage(stderr)
		return 2
	}
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

func runStatus(args []string, stdout, stderr io.Writer) int {
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
	printStatus(stdout, status)
	return 0
}

func runWatch(args []string, stdout, stderr io.Writer) int {
	opts, ok := parseWatchOptions("aepselfhost watch", args, stderr)
	if !ok {
		return 2
	}
	if err := opts.validate(); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return watchSelfhost(opts, stdout, stderr)
}

func runStartWatch(args []string, stdout, stderr io.Writer) int {
	opts, ok := parseWatchOptions("aepselfhost start-watch", args, stderr)
	if !ok {
		return 2
	}
	if err := opts.validate(); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return startWatch(opts, stdout, stderr)
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

func watchSelfhost(opts watchOptions, stdout, stderr io.Writer) int {
	startedAt := utcNow()
	verifyArgs := verifyCommandArgs(opts, opts.OpenFirst)
	showArgs := []string{"outcome", "-out-root", opts.OutRoot}
	if opts.DryRun {
		fmt.Fprintln(stdout, "DRY RUN technique selfhost watch")
		fmt.Fprintf(stdout, "verify command: pwsh %s\n", strings.Join(verifyArgs, " "))
		fmt.Fprintf(stdout, "show command:   aepselfhost %s\n", strings.Join(showArgs, " "))
		status := watchFile{
			Mode:              "dry_run",
			StartedAtUTC:      startedAt,
			PlannedIterations: opts.Iterations,
			DurationMinutes:   opts.DurationMinutes,
			IntervalSeconds:   opts.IntervalSeconds,
			VerifyCommand:     "pwsh " + strings.Join(verifyArgs, " "),
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
		lastExit = runCommand(stdout, stderr, "pwsh", verifyCommandArgs(opts, opts.OpenFirst && iteration == 1)...)
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
			VerifyCommand:       "pwsh " + strings.Join(verifyCommandArgs(opts, opts.OpenFirst && iteration == 1), " "),
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

func startWatch(opts watchOptions, stdout, stderr io.Writer) int {
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
	cmd := exec.Command(exe, watchArgs...)
	cmd.Stdout = outFile
	cmd.Stderr = errFile
	cmd.Dir = mustGetwd()
	if err := cmd.Start(); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	process.Mode = "started"
	process.PID = cmd.Process.Pid
	process.WatchStatus = filepath.Join(opts.OutRoot, "watch_status.json")
	process.LatestOutcomeHTML = filepath.Join(opts.OutRoot, "latest_outcome.html")
	process.LatestOutcomeJSON = filepath.Join(opts.OutRoot, "latest_outcome.json")
	if err := writeProcessStatus(opts.OutRoot, process); err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	_ = cmd.Process.Release()
	fmt.Fprintf(stdout, "started technique selfhost watch pid=%d\n", process.PID)
	fmt.Fprintf(stdout, "stdout: %s\n", stdoutLog)
	fmt.Fprintf(stdout, "stderr: %s\n", stderrLog)
	fmt.Fprintf(stdout, "status: %s\n", filepath.Join(opts.OutRoot, "watch_process.json"))
	return 0
}

func verifyCommandArgs(opts watchOptions, open bool) []string {
	args := []string{"-NoProfile", "-File", filepath.Join("scripts", "verify_technique_selfhost.ps1"), "-OutRoot", opts.OutRoot}
	if opts.Limit > 0 {
		args = append(args, "-Limit", strconv.Itoa(opts.Limit))
	}
	if open {
		args = append(args, "-Open")
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

func runCommand(stdout, stderr io.Writer, name string, args ...string) int {
	cmd := exec.Command(name, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
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

func printStatus(w io.Writer, status statusView) {
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
		fmt.Fprintf(w, "- state: %s\n", processState(process))
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

func processState(process processFile) string {
	if process.Mode == "dry_run" {
		return "dry-run"
	}
	if process.PID <= 0 {
		return "n/a"
	}
	if processRunning(process.PID) {
		return "running"
	}
	return "not-running"
}

func processRunning(pid int) bool {
	if runtime.GOOS == "windows" {
		out, err := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(pid), "/FO", "CSV", "/NH").Output()
		if err != nil {
			return false
		}
		return strings.Contains(string(out), strconv.Itoa(pid))
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(os.Signal(nil)) == nil
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
	fmt.Fprintln(stderr, "usage: aepselfhost <outcome|status> -out-root tmp\\technique_selfhost_gate")
}
