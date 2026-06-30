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
	fieldPath := fs.String("field", "", "recipe field path to explain")
	jsonOut := fs.Bool("json", false, "print JSON report")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *recipePath != "" && *fieldPath != "" {
		fmt.Fprintln(os.Stderr, "usage: aeprecipe explain [-recipe recipe.json | -field recipe.path] [-json]")
		return 2
	}
	if *fieldPath != "" {
		return runExplainField(*fieldPath, *jsonOut)
	}
	if *recipePath == "" {
		fmt.Fprintln(os.Stderr, "usage: aeprecipe explain [-recipe recipe.json | -field recipe.path] [-json]")
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

type recipeIndexFile struct {
	Fields map[string]recipeIndexEntry `json:"fields"`
}

type recipeIndexEntry struct {
	Path           string   `json:"path"`
	JSONName       string   `json:"json_name"`
	SourceType     string   `json:"source_type"`
	SourceField    string   `json:"source_field"`
	Type           string   `json:"type"`
	GoType         string   `json:"go_type"`
	Requiredness   []string `json:"requiredness,omitempty"`
	Summary        string   `json:"summary,omitempty"`
	Validation     string   `json:"validation,omitempty"`
	Enum           []string `json:"enum,omitempty"`
	CapabilityKeys []string `json:"capability_keys,omitempty"`
	Example        string   `json:"example,omitempty"`
}

func runExplainField(fieldPath string, jsonOut bool) int {
	index, err := readRecipeIndex()
	if err != nil {
		fmt.Fprintln(os.Stderr, "recipe index:", err)
		return 2
	}
	entry, ok := index.Fields[fieldPath]
	if !ok {
		fmt.Fprintf(os.Stderr, "recipe field %q not found\n", fieldPath)
		return 1
	}
	if jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(entry); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		return 0
	}
	fmt.Printf("field %s\n", entry.Path)
	fmt.Printf("type %s\n", entry.Type)
	if entry.SourceType != "" || entry.SourceField != "" {
		fmt.Printf("source %s.%s\n", entry.SourceType, entry.SourceField)
	}
	if len(entry.Requiredness) > 0 {
		fmt.Printf("requiredness %s\n", strings.Join(entry.Requiredness, ","))
	}
	if entry.Summary != "" {
		fmt.Printf("summary %s\n", entry.Summary)
	}
	if entry.Validation != "" {
		fmt.Printf("validation %s\n", entry.Validation)
	}
	for _, value := range entry.Enum {
		fmt.Printf("enum %s\n", value)
	}
	for _, key := range entry.CapabilityKeys {
		fmt.Printf("capability %s\n", key)
	}
	if entry.Example != "" {
		fmt.Printf("example %s\n", entry.Example)
	}
	return 0
}

func readRecipeIndex() (recipeIndexFile, error) {
	root, err := repoRoot()
	if err != nil {
		return recipeIndexFile{}, err
	}
	data, err := os.ReadFile(filepath.Join(root, "docs", "recipe_index.json"))
	if err != nil {
		return recipeIndexFile{}, err
	}
	var index recipeIndexFile
	if err := json.Unmarshal(data, &index); err != nil {
		return recipeIndexFile{}, err
	}
	return index, nil
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
