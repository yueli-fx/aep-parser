package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
	"github.com/yueli-fx/aep-parser/internal/technique"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("aeptechnique", flag.ContinueOnError)
	fs.SetOutput(stderr)
	input := fs.String("in", "", "input .aep file")
	_ = fs.Bool("json", true, "emit technique facts as JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *input == "" {
		fmt.Fprintln(stderr, "usage: aeptechnique -in file.aep")
		return 2
	}

	project, err := aep.Open(*input)
	if err != nil {
		fmt.Fprintf(stderr, "open %q: %v\n", *input, err)
		return 1
	}
	prof, err := profile.Build(project, profile.Options{Path: *input})
	if err != nil {
		fmt.Fprintf(stderr, "profile %q: %v\n", *input, err)
		return 1
	}
	facts, err := technique.Build(prof)
	if err != nil {
		fmt.Fprintf(stderr, "technique %q: %v\n", *input, err)
		return 1
	}

	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(facts); err != nil {
		fmt.Fprintf(stderr, "json: %v\n", err)
		return 1
	}
	return 0
}
