package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/example/aep-parser/internal/recipe"
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
		fmt.Println("valid")
		return nil
	}
	fmt.Printf("invalid refusals=%d\n", len(report.Refusals))
	for _, refusal := range report.Refusals {
		fmt.Printf("%s %s %s\n", refusal.Code, refusal.Path, refusal.Message)
	}
	return nil
}

func defaultCapabilities() recipe.StaticCapabilities {
	return recipe.StaticCapabilities{
		"Third Party Magic": recipe.CapabilityUnsupported,
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: aeprecipe <validate|compile|explain> [flags]")
}
