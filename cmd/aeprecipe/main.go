package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/capindex"
	"github.com/yueli-fx/aep-parser/internal/recipe"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	switch args[0] {
	case "validate":
		return runValidate(args[1:])
	case "compile":
		return runCompile(args[1:])
	case "explain":
		return runExplain(args[1:])
	default:
		usage()
		return 2
	}
}

func runValidate(args []string) int {
	fs := flag.NewFlagSet("aeprecipe validate", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	recipePath := fs.String("recipe", "", "recipe JSON path")
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *recipePath == "" {
		fmt.Fprintln(os.Stderr, "usage: aeprecipe validate -recipe recipe.json [-json]")
		return 2
	}
	rec, err := readRecipe(*recipePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "recipe:", err)
		return 2
	}
	report := recipe.ValidateWithCapabilities(rec, defaultCapabilities())
	if err := emitReport(report, *jsonOut); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if !report.Valid {
		return 1
	}
	return 0
}

func runCompile(args []string) int {
	fs := flag.NewFlagSet("aeprecipe compile", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	recipePath := fs.String("recipe", "", "recipe JSON path")
	outPath := fs.String("out", "", "output AEP path")
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *recipePath == "" || *outPath == "" {
		fmt.Fprintln(os.Stderr, "usage: aeprecipe compile -recipe recipe.json -out out.aep [-json]")
		return 2
	}
	rec, err := readRecipe(*recipePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "recipe:", err)
		return 2
	}
	report, err := recipe.CompileToFile(rec, *outPath, defaultCapabilities())
	if err != nil {
		fmt.Fprintln(os.Stderr, "compile:", err)
		return 2
	}
	if err := emitReport(report, *jsonOut); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if !report.Valid {
		return 1
	}
	return 0
}

func runExplain(args []string) int {
	fs := flag.NewFlagSet("aeprecipe explain", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	recipePath := fs.String("recipe", "", "recipe JSON path")
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *recipePath == "" {
		fmt.Fprintln(os.Stderr, "usage: aeprecipe explain -recipe recipe.json [-json]")
		return 2
	}
	rec, err := readRecipe(*recipePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "recipe:", err)
		return 2
	}
	report := recipe.ValidateWithCapabilities(rec, defaultCapabilities())
	if err := emitReport(report, *jsonOut); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	if !report.Valid {
		return 1
	}
	return 0
}

func readRecipe(path string) (recipe.Recipe, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return recipe.Recipe{}, err
	}
	var rec recipe.Recipe
	if err := json.Unmarshal(data, &rec); err != nil {
		return recipe.Recipe{}, err
	}
	return rec, nil
}

func emitReport(report recipe.Report, jsonOut bool) error {
	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(report)
	}
	if report.Valid {
		fmt.Printf("valid capabilities=%d downgrades=%d\n", len(report.Capabilities), len(report.Downgrades))
		return nil
	}
	fmt.Printf("invalid capabilities=%d downgrades=%d refusals=%d\n", len(report.Capabilities), len(report.Downgrades), len(report.Refusals))
	for _, downgrade := range report.Downgrades {
		fmt.Printf("downgrade %s %s %s\n", downgrade.Code, downgrade.Path, downgrade.Message)
	}
	for _, refusal := range report.Refusals {
		fmt.Printf("%s %s %s\n", refusal.Code, refusal.Path, refusal.Message)
	}
	return nil
}

func defaultCapabilities() recipe.CapabilityIndex {
	root, err := repoRoot()
	if err != nil {
		return recipe.StaticCapabilities{}
	}
	idx, err := capindex.Load(filepath.Join(root, "docs", "capabilities.json"))
	if err != nil {
		return recipe.StaticCapabilities{}
	}
	return capindexRecipeAdapter{idx: idx}
}

type capindexRecipeAdapter struct {
	idx *capindex.Index
}

func (a capindexRecipeAdapter) Lookup(query string) recipe.CapabilityLookup {
	got := a.idx.Lookup(query)
	out := recipe.CapabilityLookup{
		Query:  got.Query,
		Status: recipe.CapabilityStatus(got.Status),
	}
	if got.Entry.Symbol == "" {
		return out
	}
	out.Symbol = got.Entry.Symbol
	if got.Entry.Recv != "" {
		out.Symbol = strings.TrimPrefix(got.Entry.Recv, "*") + "." + got.Entry.Symbol
	}
	out.Domain = got.Entry.Cap.Domain
	out.Tier = got.Entry.Cap.Tier
	out.Verify = got.Entry.Cap.Verify
	out.MinVer = got.Entry.Cap.MinVer
	out.Boundary = got.Entry.Cap.Boundary
	out.Gate = got.Entry.Cap.Gate
	return out
}

func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found above %s", dir)
		}
		dir = parent
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: aeprecipe <validate|compile|explain> [flags]")
}
