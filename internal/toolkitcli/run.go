package toolkitcli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"

	publicaep "github.com/yueli-fx/aep-parser"
	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/aepmigrate"
	"github.com/yueli-fx/aep-parser/internal/profile"
	"github.com/yueli-fx/aep-parser/internal/profilediff"
)

const schemaVersion = 1

type envelope struct {
	SchemaVersion int    `json:"schema_version"`
	Command       string `json:"command"`
	Data          any    `json:"data"`
}

type errorEnvelope struct {
	SchemaVersion int         `json:"schema_version"`
	Error         errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Run executes the unified headless toolkit CLI.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return writeFailure(stderr, "usage", "usage: aep <inspect|profile|diff|migrate|capabilities> [flags]", 2)
	}
	switch args[0] {
	case "inspect":
		return runInspect(args[1:], stdout, stderr)
	case "profile":
		return runProfile(args[1:], stdout, stderr)
	case "diff":
		return runDiff(args[1:], stdout, stderr)
	case "migrate":
		return runMigrate(args[1:], stdout, stderr)
	case "capabilities":
		return runCapabilities(args[1:], stdout, stderr)
	case "help", "-h", "--help":
		return writeResult(stdout, "help", map[string]any{
			"usage":    "aep <inspect|profile|diff|migrate|capabilities> [flags]",
			"commands": []string{"inspect", "profile", "diff", "migrate", "capabilities"},
		})
	default:
		return writeFailure(stderr, "unknown_command", fmt.Sprintf("unknown command %q", args[0]), 2)
	}
}

func runInspect(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("aep inspect")
	input := fs.String("in", "", "AEP input path")
	if err := fs.Parse(args); err != nil {
		return writeFailure(stderr, "usage", err.Error(), 2)
	}
	if *input == "" {
		return writeFailure(stderr, "usage", "usage: aep inspect -in project.aep", 2)
	}
	document, err := publicaep.Open(*input)
	if err != nil {
		return writeFailure(stderr, "open_failed", err.Error(), 1)
	}
	return writeResult(stdout, "inspect", document.Inspect())
}

func runProfile(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("aep profile")
	input := fs.String("in", "", "AEP input path")
	if err := fs.Parse(args); err != nil {
		return writeFailure(stderr, "usage", err.Error(), 2)
	}
	if *input == "" {
		return writeFailure(stderr, "usage", "usage: aep profile -in project.aep", 2)
	}
	document, err := publicaep.Open(*input)
	if err != nil {
		return writeFailure(stderr, "open_failed", err.Error(), 1)
	}
	data, err := document.ProfileJSON()
	if err != nil {
		return writeFailure(stderr, "profile_failed", err.Error(), 1)
	}
	return writeResult(stdout, "profile", json.RawMessage(data))
}

func runDiff(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("aep diff")
	expectedPath := fs.String("expected", "", "expected AEP path")
	actualPath := fs.String("actual", "", "actual AEP path")
	ignorePath := fs.String("ignore", "", "optional diff ignore-rules JSON")
	if err := fs.Parse(args); err != nil {
		return writeFailure(stderr, "usage", err.Error(), 2)
	}
	if *expectedPath == "" || *actualPath == "" {
		return writeFailure(stderr, "usage", "usage: aep diff -expected expected.aep -actual actual.aep [-ignore rules.json]", 2)
	}
	expected, err := buildProfile(*expectedPath)
	if err != nil {
		return writeFailure(stderr, "expected_profile_failed", err.Error(), 1)
	}
	actual, err := buildProfile(*actualPath)
	if err != nil {
		return writeFailure(stderr, "actual_profile_failed", err.Error(), 1)
	}
	var ignores *profilediff.IgnoreRules
	if *ignorePath != "" {
		ignores, err = profilediff.LoadIgnoreRules(*ignorePath)
		if err != nil {
			return writeFailure(stderr, "invalid_ignore_rules", err.Error(), 2)
		}
	}
	report, err := profilediff.Compare(expected, actual, profilediff.Options{IgnoreRules: ignores})
	if err != nil {
		return writeFailure(stderr, "diff_failed", err.Error(), 1)
	}
	if code := writeResult(stdout, "diff", report); code != 0 {
		return code
	}
	if report.DiffCount > 0 {
		return 1
	}
	return 0
}

func runMigrate(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("aep migrate")
	input := fs.String("in", "", "source AEP path")
	output := fs.String("out", "", "target AEP path")
	targetRaw := fs.String("target", "", aepmigrate.TargetVersionHelp())
	if err := fs.Parse(args); err != nil {
		return writeFailure(stderr, "usage", err.Error(), 2)
	}
	if *input == "" || *output == "" || *targetRaw == "" {
		return writeFailure(stderr, "usage", "usage: aep migrate -in source.aep -target AE2025 -out migrated.aep", 2)
	}
	target, err := aepmigrate.ParseVersionLabel(*targetRaw)
	if err != nil {
		return writeFailure(stderr, "invalid_target", err.Error(), 2)
	}
	report, err := aepmigrate.Convert(aepmigrate.ConvertOptions{InputPath: *input, OutputPath: *output, Target: target})
	if err != nil {
		return writeFailure(stderr, "migration_failed", err.Error(), 1)
	}
	if code := writeResult(stdout, "migrate", report); code != 0 {
		return code
	}
	if report.Summary.Status == aepmigrate.StatusBlocked || report.Summary.Status == aepmigrate.StatusError {
		return 1
	}
	return 0
}

func runCapabilities(args []string, stdout, stderr io.Writer) int {
	fs := newFlagSet("aep capabilities")
	if err := fs.Parse(args); err != nil {
		return writeFailure(stderr, "usage", err.Error(), 2)
	}
	if fs.NArg() != 0 {
		return writeFailure(stderr, "usage", "usage: aep capabilities", 2)
	}
	return writeResult(stdout, "capabilities", map[string]any{
		"commands":          []string{"inspect", "profile", "diff", "migrate", "capabilities"},
		"migration_targets": aepmigrate.SupportedVersionStrings(),
		"requires_ae":       false,
		"profile_schema":    profile.SchemaVersion,
	})
}

func buildProfile(path string) (*profile.Profile, error) {
	project, err := aep.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	return profile.Build(project, profile.Options{Path: path})
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func writeResult(w io.Writer, command string, data any) int {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(envelope{SchemaVersion: schemaVersion, Command: command, Data: data}); err != nil {
		return 1
	}
	return 0
}

func writeFailure(w io.Writer, code, message string, exitCode int) int {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(errorEnvelope{
		SchemaVersion: schemaVersion,
		Error:         errorDetail{Code: code, Message: message},
	})
	return exitCode
}
