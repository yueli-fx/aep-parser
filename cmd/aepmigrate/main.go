package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aepmigrate"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "usage: aepmigrate assess -in source.aep -target AE2020 [-out assess.json]")
		return 2
	}
	switch args[0] {
	case "assess":
		return runAssess(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		return 2
	}
}

func runAssess(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("aepmigrate assess", flag.ContinueOnError)
	fs.SetOutput(stderr)
	input := fs.String("in", "", "source .aep path")
	targetRaw := fs.String("target", "", "target AE version: AE2020, AE2022, or AE2025")
	outPath := fs.String("out", "", "optional JSON output path")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *input == "" || *targetRaw == "" {
		fmt.Fprintln(stderr, "usage: aepmigrate assess -in source.aep -target AE2020 [-out assess.json]")
		return 2
	}
	target, err := aepmigrate.ParseVersionLabel(*targetRaw)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	report, err := aepmigrate.Assess(aepmigrate.Options{InputPath: *input, Target: target})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	data = append(data, '\n')
	if *outPath != "" {
		if err := os.WriteFile(*outPath, data, 0o644); err != nil {
			fmt.Fprintln(stderr, err)
			return 1
		}
		fmt.Fprintf(stdout, "migration assess: %s\n", *outPath)
		return statusCode(report.Summary.Status)
	}
	fmt.Fprint(stdout, string(data))
	return statusCode(report.Summary.Status)
}

func statusCode(status aepmigrate.Status) int {
	if status == aepmigrate.StatusBlocked || status == aepmigrate.StatusError {
		return 1
	}
	return 0
}
