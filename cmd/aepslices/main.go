package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/example/aep-parser/internal/aeoracle"
	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/profile"
	"github.com/example/aep-parser/internal/profilediff"
	"github.com/example/aep-parser/internal/sliceworkflow"
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
	case "plan":
		return runPlan(args[1:])
	case "diagnose":
		return runDiagnose(args[1:])
	default:
		usage()
		return 2
	}
}

func runPlan(args []string) int {
	fs := flag.NewFlagSet("aepslices plan", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	aepPath := fs.String("aep", "", "source AEP path")
	maxSlices := fs.Int("max", sliceworkflow.DefaultMaxSlices, "maximum slices to select")
	dictPath := fs.String("dict", defaultDictPath, "effect dictionary json; empty to disable")
	jsonOut := fs.Bool("json", false, "emit JSON workflow report")
	outPath := fs.String("out", "", "optional output path")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *aepPath == "" {
		fmt.Fprintln(os.Stderr, "usage: aepslices plan -aep project.aep [-max 3] [-json] [-out report.json]")
		return 2
	}
	dict, err := loadDictionary(*dictPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warn: no effect dict (%v) - falling back to raw matchName output\n", err)
	}
	prof, err := buildProfile(*aepPath, dict)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	report, err := sliceworkflow.BuildReport(prof, sliceworkflow.Options{
		SourceProject: *aepPath,
		MaxSlices:     *maxSlices,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	return emitReport(report, *jsonOut, *outPath)
}

func runDiagnose(args []string) int {
	fs := flag.NewFlagSet("aepslices diagnose", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	expectedPath := fs.String("expected", "", "expected/source AEP path")
	actualPath := fs.String("actual", "", "actual/observed AEP path")
	maxSlices := fs.Int("max", sliceworkflow.DefaultMaxSlices, "maximum slices to select")
	dictPath := fs.String("dict", defaultDictPath, "effect dictionary json; empty to disable")
	ignorePath := fs.String("ignore", "", "JSON ignore rules file")
	expectedPNG := fs.String("expected-png", "", "expected PNG path for optional render gap")
	actualPNG := fs.String("actual-png", "", "actual PNG path for optional render gap")
	threshold := fs.Int("threshold", 0, "per-channel render threshold 0..255")
	jsonOut := fs.Bool("json", false, "emit JSON workflow report")
	outPath := fs.String("out", "", "optional output path")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *expectedPath == "" || *actualPath == "" || *threshold < 0 || *threshold > 255 {
		fmt.Fprintln(os.Stderr, "usage: aepslices diagnose -expected original.aep -actual clone.aep [-expected-png a.png -actual-png b.png] [-json] [-out report.json]")
		return 2
	}
	if (*expectedPNG == "") != (*actualPNG == "") {
		fmt.Fprintln(os.Stderr, "expected-png and actual-png must be provided together")
		return 2
	}

	dict, err := loadDictionary(*dictPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warn: no effect dict (%v) - falling back to raw matchName output\n", err)
	}
	expected, err := buildProfile(*expectedPath, dict)
	if err != nil {
		fmt.Fprintln(os.Stderr, "expected:", err)
		return 2
	}
	actual, err := buildProfile(*actualPath, dict)
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

	var renderReports []aeoracle.CompareReport
	if *expectedPNG != "" {
		renderReport, err := aeoracle.ComparePNG(*expectedPNG, *actualPNG, aeoracle.CompareOptions{ChannelThreshold: uint8(*threshold)})
		if err != nil {
			fmt.Fprintln(os.Stderr, "render compare:", err)
			return 2
		}
		renderReports = append(renderReports, renderReport)
	}

	report, err := sliceworkflow.BuildDiagnoseReport(expected, actual, diffReport, renderReports, sliceworkflow.Options{
		SourceProject: *expectedPath,
		ObservedIn:    *actualPath,
		MaxSlices:     *maxSlices,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	code := emitReport(report, *jsonOut, *outPath)
	if code != 0 {
		return code
	}
	if report.Summary.GapCount > 0 {
		return 1
	}
	return 0
}

func loadDictionary(path string) (*profile.EffectDictionary, error) {
	if path == "" {
		return nil, nil
	}
	return profile.LoadEffectDictionary(path)
}

func buildProfile(path string, dict *profile.EffectDictionary) (*profile.Profile, error) {
	project, err := aep.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", path, err)
	}
	return profile.Build(project, profile.Options{Path: path, Dict: dict})
}

func emitReport(report sliceworkflow.Report, jsonOut bool, outPath string) int {
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
	return 0
}

func printText(report sliceworkflow.Report) {
	fmt.Printf("%s slices: %d", report.Mode, report.Summary.SelectedSliceCount)
	if report.Summary.GapCount > 0 {
		fmt.Printf(" gaps: %d", report.Summary.GapCount)
	}
	fmt.Println()
	for _, slice := range report.Slices {
		fmt.Printf("%s %d %d %s\n", slice.Classification, slice.Score, slice.CompID, slice.Name)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: aepslices <plan|diagnose> [flags]")
}
