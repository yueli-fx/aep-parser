package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/example/aep-parser/internal/aeoracle"
	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/gapledger"
	"github.com/example/aep-parser/internal/profile"
	"github.com/example/aep-parser/internal/profilediff"
)

const defaultDictPath = "data/effects-dict/effects_en_US_25.1x68.json"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	switch args[0] {
	case "diff":
		return runDiff(args[1:])
	case "render":
		return runRender(args[1:])
	default:
		usage()
		return 2
	}
}

func runDiff(args []string) int {
	fs := flag.NewFlagSet("aepgaps diff", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	jsonOut := fs.Bool("json", false, "emit JSON gap report")
	outPath := fs.String("out", "", "optional output path")
	ignorePath := fs.String("ignore", "", "JSON ignore rules file")
	dictPath := fs.String("dict", defaultDictPath, "effect dictionary json; empty to disable")
	source := fs.String("source", "", "source project label; defaults to expected AEP path")
	observed := fs.String("observed", "", "observed target label; defaults to actual AEP path")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	paths := fs.Args()
	if len(paths) != 2 {
		fmt.Fprintln(os.Stderr, "usage: aepgaps diff [-json] [-out gaps.json] [-ignore rules.json] expected.aep actual.aep")
		return 2
	}
	ctx := gapledger.Context{SourceProject: valueOr(*source, paths[0]), ObservedIn: valueOr(*observed, paths[1])}

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
	diffReport, err := profilediff.Compare(expected, actual, profilediff.Options{IgnoreRules: ignores})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	report := gapledger.FromDiffReport(diffReport, ctx)
	return emit(report, *jsonOut, *outPath)
}

func runRender(args []string) int {
	fs := flag.NewFlagSet("aepgaps render", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	jsonOut := fs.Bool("json", false, "emit JSON gap report")
	outPath := fs.String("out", "", "optional output path")
	expected := fs.String("expected", "", "expected PNG path")
	actual := fs.String("actual", "", "actual PNG path")
	threshold := fs.Int("threshold", 0, "per-channel threshold 0..255")
	source := fs.String("source", "", "source project label; defaults to expected PNG path")
	observed := fs.String("observed", "", "observed target label; defaults to actual PNG path")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *expected == "" || *actual == "" || *threshold < 0 || *threshold > 255 {
		fmt.Fprintln(os.Stderr, "usage: aepgaps render -expected original.png -actual clone.png [-threshold 0..255] [-json] [-out gaps.json]")
		return 2
	}
	compare, err := aeoracle.ComparePNG(*expected, *actual, aeoracle.CompareOptions{ChannelThreshold: uint8(*threshold)})
	if err != nil {
		fmt.Fprintln(os.Stderr, "compare:", err)
		return 2
	}
	report := gapledger.FromRenderCompare(compare, gapledger.Context{
		SourceProject: valueOr(*source, *expected),
		ObservedIn:    valueOr(*observed, *actual),
	})
	return emit(report, *jsonOut, *outPath)
}

func buildProfile(path string, dict *profile.EffectDictionary) (*profile.Profile, error) {
	project, err := aep.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	return profile.Build(project, profile.Options{Path: path, Dict: dict})
}

func emit(report gapledger.Report, jsonOut bool, outPath string) int {
	var data []byte
	var err error
	if jsonOut || outPath != "" {
		data, err = json.MarshalIndent(report, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, "json:", err)
			return 2
		}
		data = append(data, '\n')
	}
	if outPath != "" {
		if err := os.WriteFile(outPath, data, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "write:", err)
			return 2
		}
	}
	if jsonOut {
		if _, err := os.Stdout.Write(data); err != nil {
			fmt.Fprintln(os.Stderr, "stdout:", err)
			return 2
		}
	} else {
		printText(report)
	}
	if report.GapCount > 0 {
		return 1
	}
	return 0
}

func printText(report gapledger.Report) {
	fmt.Printf("gaps: %d\n", report.GapCount)
	const limit = 50
	for i, gap := range report.Gaps {
		if i == limit {
			fmt.Printf("... %d more\n", len(report.Gaps)-limit)
			return
		}
		fmt.Printf("%s %s %s %s %s\n", gap.ID, gap.Type, gap.Severity, gap.ActionType, gap.ProfilePath)
	}
}

func valueOr(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: aepgaps <diff|render> [flags]")
}
