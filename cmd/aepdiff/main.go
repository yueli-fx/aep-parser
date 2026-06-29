package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
	"github.com/yueli-fx/aep-parser/internal/profilediff"
)

const defaultDictPath = "data/effects-dict/effects_en_US_25.1x68.json"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("aepdiff", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	jsonOut := fs.Bool("json", false, "emit the full diff report as JSON")
	ignorePath := fs.String("ignore", "", "JSON ignore rules file")
	dictPath := fs.String("dict", defaultDictPath, "effect dictionary json; empty to disable")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	paths := fs.Args()
	if len(paths) != 2 {
		fmt.Fprintln(os.Stderr, "usage: aepdiff [-json] [-ignore rules.json] [-dict effects.json] expected.aep actual.aep")
		return 2
	}

	var dict *profile.EffectDictionary
	if *dictPath != "" {
		var err error
		dict, err = profile.LoadEffectDictionary(*dictPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: no effect dict (%v) - falling back to raw matchName output\n", err)
		}
	}

	expected, err := buildProfile(paths[0], dict)
	if err != nil {
		fmt.Fprintln(os.Stderr, "expected:", err)
		return 2
	}
	actual, err := buildProfile(paths[1], dict)
	if err != nil {
		fmt.Fprintln(os.Stderr, "actual:", err)
		return 2
	}

	var ignores *profilediff.IgnoreRules
	if *ignorePath != "" {
		ignores, err = profilediff.LoadIgnoreRules(*ignorePath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
	}

	report, err := profilediff.Compare(expected, actual, profilediff.Options{IgnoreRules: ignores})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			fmt.Fprintln(os.Stderr, "json:", err)
			return 2
		}
	} else {
		printText(report)
	}
	if report.DiffCount > 0 {
		return 1
	}
	return 0
}

func buildProfile(path string, dict *profile.EffectDictionary) (*profile.Profile, error) {
	project, err := aep.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	return profile.Build(project, profile.Options{Path: path, Dict: dict})
}

func printText(report *profilediff.Report) {
	fmt.Printf("diffs: %d", report.DiffCount)
	if report.IgnoredCount > 0 {
		fmt.Printf(" ignored: %d", report.IgnoredCount)
	}
	fmt.Println()
	const limit = 50
	for i, diff := range report.Diffs {
		if i == limit {
			fmt.Printf("... %d more\n", len(report.Diffs)-limit)
			return
		}
		fmt.Printf("%s %s %s %s\n", diff.Kind, diff.Severity, diff.ActionType, diff.Path)
	}
}
