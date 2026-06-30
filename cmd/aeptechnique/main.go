package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

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
	mode := fs.String("mode", "facts", "output mode: facts or portrait")
	portraitMode := fs.Bool("portrait", false, "emit project portrait JSON")
	corpusMode := fs.Bool("corpus", false, "emit one JSONL record per discovered project")
	recursive := fs.Bool("recursive", false, "discover .aep files recursively when -in is a directory")
	limit := fs.Int("limit", 0, "maximum number of discovered projects to process; 0 means no limit")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *input == "" {
		fmt.Fprintln(stderr, "usage: aeptechnique -in file.aep")
		return 2
	}
	if *portraitMode {
		*mode = "portrait"
	}
	if *mode != "facts" && *mode != "portrait" {
		fmt.Fprintf(stderr, "invalid -mode %q; want facts or portrait\n", *mode)
		return 2
	}
	if *corpusMode {
		return runCorpus(*input, *mode, *recursive, *limit, stdout, stderr)
	}

	output, code := buildOutput(*input, *mode, stderr)
	if code != 0 {
		return code
	}
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(output); err != nil {
		fmt.Fprintf(stderr, "json: %v\n", err)
		return 1
	}
	return 0
}

type corpusRecord struct {
	Path     string              `json:"path"`
	Mode     string              `json:"mode"`
	Facts    *technique.FactSet  `json:"facts,omitempty"`
	Portrait *technique.Portrait `json:"portrait,omitempty"`
	Error    string              `json:"error,omitempty"`
}

func buildOutput(input, mode string, stderr io.Writer) (any, int) {
	project, err := aep.Open(input)
	if err != nil {
		fmt.Fprintf(stderr, "open %q: %v\n", input, err)
		return nil, 1
	}
	prof, err := profile.Build(project, profile.Options{Path: input})
	if err != nil {
		fmt.Fprintf(stderr, "profile %q: %v\n", input, err)
		return nil, 1
	}
	facts, err := technique.Build(prof)
	if err != nil {
		fmt.Fprintf(stderr, "technique %q: %v\n", input, err)
		return nil, 1
	}
	switch mode {
	case "facts":
		return facts, 0
	case "portrait":
		portrait, err := technique.BuildPortrait(facts)
		if err != nil {
			fmt.Fprintf(stderr, "portrait %q: %v\n", input, err)
			return nil, 1
		}
		return portrait, 0
	default:
		fmt.Fprintf(stderr, "invalid -mode %q; want facts or portrait\n", mode)
		return nil, 2
	}
}

func runCorpus(input, mode string, recursive bool, limit int, stdout, stderr io.Writer) int {
	paths, err := discoverAEPs(input, recursive)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if limit > 0 && len(paths) > limit {
		paths = paths[:limit]
	}
	enc := json.NewEncoder(stdout)
	hadError := false
	for _, path := range paths {
		record := corpusRecord{Path: path, Mode: mode}
		output, code := buildOutput(path, mode, stderr)
		if code != 0 {
			hadError = true
			record.Error = fmt.Sprintf("build failed with exit code %d", code)
		} else {
			switch value := output.(type) {
			case *technique.FactSet:
				record.Facts = value
			case *technique.Portrait:
				record.Portrait = value
			}
		}
		if err := enc.Encode(record); err != nil {
			fmt.Fprintf(stderr, "jsonl: %v\n", err)
			return 1
		}
	}
	if hadError {
		return 1
	}
	return 0
}

func discoverAEPs(input string, recursive bool) ([]string, error) {
	info, err := os.Stat(input)
	if err != nil {
		return nil, fmt.Errorf("stat %q: %w", input, err)
	}
	if !info.IsDir() {
		return []string{input}, nil
	}
	if !recursive {
		return nil, fmt.Errorf("directory input %q requires -recursive in corpus mode", input)
	}
	var paths []string
	err = filepath.WalkDir(input, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(filepath.Ext(path), ".aep") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	return paths, nil
}
