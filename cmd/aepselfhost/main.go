package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
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
	Mode      string `json:"mode"`
	PID       int    `json:"pid"`
	Command   string `json:"command"`
	StdoutLog string `json:"stdout_log"`
	StderrLog string `json:"stderr_log"`
}

type watchFile struct {
	Mode                string `json:"mode"`
	CompletedIterations *int   `json:"completed_iterations"`
	LastExitCode        *int   `json:"last_exit_code"`
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
