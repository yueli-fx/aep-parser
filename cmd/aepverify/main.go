package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/capindex"
	"github.com/yueli-fx/aep-parser/internal/host"
	"github.com/yueli-fx/aep-parser/internal/recipe"
)

type RecipeProfileSummary struct {
	SchemaVersion        int                   `json:"schema_version"`
	Total                int                   `json:"total"`
	Passed               int                   `json:"passed"`
	Failed               int                   `json:"failed"`
	ValidateFailed       int                   `json:"validate_failed"`
	CompileFailed        int                   `json:"compile_failed"`
	InvalidReports       int                   `json:"invalid_reports"`
	ProfileCheckFailures int                   `json:"profile_check_failures"`
	CoveredProfilePaths  []string              `json:"covered_profile_paths"`
	Recipes              []RecipeProfileResult `json:"recipes"`
	WorkDir              string                `json:"work_dir,omitempty"`
	FailureReason        string                `json:"failure_reason,omitempty"`
	Details              string                `json:"details,omitempty"`
}

type RecipeProfileResult struct {
	Recipe                  string                `json:"recipe"`
	OK                      bool                  `json:"ok"`
	ValidateOK              bool                  `json:"validate_ok"`
	CompileOK               bool                  `json:"compile_ok"`
	ValidReport             bool                  `json:"valid_report"`
	CapabilityCount         int                   `json:"capability_count"`
	ProfileCheckCount       int                   `json:"profile_check_count"`
	ProfileCheckFailedCount int                   `json:"profile_check_failed_count"`
	ProfileCheckPaths       []string              `json:"profile_check_paths"`
	FailedChecks            []recipe.ProfileCheck `json:"failed_checks,omitempty"`
	OutputPath              string                `json:"output_path,omitempty"`
	ValidateError           string                `json:"validate_error,omitempty"`
	CompileError            string                `json:"compile_error,omitempty"`
}

func main() {
	os.Exit(runWithIO(os.Args[1:], os.Stdout, os.Stderr, host.ExecRunner{}))
}

func runWithIO(args []string, stdout, stderr io.Writer, runner host.Runner) int {
	if len(args) == 0 {
		usage(stderr)
		return 2
	}
	switch args[0] {
	case "cross-platform":
		return runCrossPlatform(args[1:], stdout, stderr, runner)
	case "recipe-profiles":
		return runRecipeProfiles(args[1:], stdout, stderr)
	default:
		usage(stderr)
		return 2
	}
}

func runCrossPlatform(args []string, stdout, stderr io.Writer, runner host.Runner) int {
	fs := flag.NewFlagSet("aepverify cross-platform", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if runner == nil {
		runner = host.ExecRunner{}
	}
	for _, target := range crossPlatformTargets() {
		fmt.Fprintf(stdout, "building GOOS=%s GOARCH=%s\n", target.GOOS, target.GOARCH)
		env := append(os.Environ(),
			"GOOS="+target.GOOS,
			"GOARCH="+target.GOARCH,
			"CGO_ENABLED=0",
		)
		cmdArgs := append([]string{"build"}, crossPlatformPackages()...)
		result := runner.Run(context.Background(), host.Command{
			Name:   "go",
			Args:   cmdArgs,
			Env:    env,
			Stdout: stdout,
			Stderr: stderr,
		})
		if result.ExitCode != 0 {
			return result.ExitCode
		}
	}
	return 0
}

type buildTarget struct {
	GOOS   string
	GOARCH string
}

func crossPlatformTargets() []buildTarget {
	return []buildTarget{
		{GOOS: "windows", GOARCH: "amd64"},
		{GOOS: "darwin", GOARCH: "arm64"},
		{GOOS: "linux", GOARCH: "amd64"},
	}
}

func crossPlatformPackages() []string {
	return []string{
		".",
		"./cmd/aep",
		"./cmd/aepserver",
		"./cmd/aepdiff",
		"./cmd/aeprecipe",
		"./cmd/aepsearch",
		"./cmd/aeptechnique",
		"./cmd/aeoracle",
		"./cmd/aepselfhost",
		"./cmd/aepverify",
		"./internal/aehost",
		"./internal/host",
		"./internal/profile",
		"./internal/profilediff",
		"./internal/projectindex",
		"./internal/recipe",
		"./internal/server",
		"./internal/technique",
		"./internal/toolkitcli",
	}
}

func runRecipeProfiles(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("aepverify recipe-profiles", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var recipePaths multiFlag
	fs.Var(&recipePaths, "recipe", "recipe JSON path; may be repeated")
	recipeDir := fs.String("recipe-dir", filepath.Join("examples", "recipes"), "recipe directory used when -recipe is not provided")
	filter := fs.String("filter", "*.json", "recipe file filter")
	outDir := fs.String("out", "", "directory for compiled .aep outputs")
	jsonOut := fs.Bool("json", false, "emit JSON summary")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	summary, exitCode := verifyRecipeProfiles(recipeProfileOptions{
		RecipePaths: recipePaths,
		RecipeDir:   *recipeDir,
		Filter:      *filter,
		OutDir:      *outDir,
	})
	if *jsonOut {
		enc := json.NewEncoder(stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(summary); err != nil {
			fmt.Fprintln(stderr, "json:", err)
			return 2
		}
		return exitCode
	}
	printRecipeProfileSummary(stdout, summary)
	return exitCode
}

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }

func (m *multiFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

type recipeProfileOptions struct {
	RecipePaths []string
	RecipeDir   string
	Filter      string
	OutDir      string
}

func verifyRecipeProfiles(opts recipeProfileOptions) (RecipeProfileSummary, int) {
	paths, err := resolveRecipePaths(opts)
	if err != nil {
		return emptyRecipeSummary("resolve_recipes_failed", err.Error()), 1
	}
	if len(paths) == 0 {
		return emptyRecipeSummary("no_recipes", fmt.Sprintf("No recipe files matched %q.", filepath.Join(opts.RecipeDir, opts.Filter))), 1
	}
	workDir := opts.OutDir
	if workDir == "" {
		dir, err := os.MkdirTemp("", "aep-parser-recipe-profile-verify-*")
		if err != nil {
			return emptyRecipeSummary("work_dir_failed", err.Error()), 1
		}
		workDir = dir
		defer os.RemoveAll(workDir)
	} else if err := os.MkdirAll(workDir, 0o755); err != nil {
		return emptyRecipeSummary("work_dir_failed", err.Error()), 1
	}
	caps := defaultCapabilities()
	var results []RecipeProfileResult
	coveredSet := map[string]bool{}
	for i, path := range paths {
		result := verifyOneRecipeProfile(i+1, path, workDir, opts.OutDir != "", caps)
		results = append(results, result)
		for _, p := range result.ProfileCheckPaths {
			coveredSet[p] = true
		}
	}
	covered := sortedKeys(coveredSet)
	summary := RecipeProfileSummary{
		SchemaVersion:       1,
		Total:               len(results),
		CoveredProfilePaths: covered,
		Recipes:             results,
	}
	if opts.OutDir != "" {
		summary.WorkDir = workDir
	}
	for _, result := range results {
		if result.OK {
			summary.Passed++
		} else {
			summary.Failed++
		}
		if !result.ValidateOK {
			summary.ValidateFailed++
		}
		if !result.CompileOK {
			summary.CompileFailed++
		}
		if !result.ValidReport {
			summary.InvalidReports++
		}
		summary.ProfileCheckFailures += result.ProfileCheckFailedCount
	}
	if summary.Failed > 0 {
		return summary, 1
	}
	return summary, 0
}

func resolveRecipePaths(opts recipeProfileOptions) ([]string, error) {
	if len(opts.RecipePaths) > 0 {
		out := append([]string(nil), opts.RecipePaths...)
		sort.Strings(out)
		return out, nil
	}
	pattern := filepath.Join(opts.RecipeDir, opts.Filter)
	paths, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	return paths, nil
}

func verifyOneRecipeProfile(index int, path, workDir string, keepOutput bool, caps recipe.CapabilityIndex) RecipeProfileResult {
	result := RecipeProfileResult{Recipe: filepath.Clean(path)}
	rec, err := readRecipe(path)
	if err != nil {
		result.ValidateError = err.Error()
		return result
	}
	validate := recipe.ValidateWithCapabilities(rec, caps)
	result.ValidateOK = validate.Valid
	if !validate.Valid {
		result.ValidateError = "recipe validation failed"
	}
	outputPath := filepath.Join(workDir, fmt.Sprintf("%03d-%s.aep", index, strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))))
	compile, err := recipe.CompileToFile(rec, outputPath, caps)
	if err != nil {
		result.CompileError = err.Error()
	} else {
		result.CompileOK = compile.Valid
		if !compile.Valid {
			result.CompileError = "recipe compile report invalid"
		}
	}
	if len(compile.Capabilities) > 0 {
		result.CapabilityCount = len(compile.Capabilities)
	} else {
		result.CapabilityCount = len(validate.Capabilities)
	}
	result.ValidReport = result.ValidateOK && result.CompileOK
	if keepOutput {
		result.OutputPath = outputPath
	}
	for _, check := range compile.ProfileChecks {
		if check.Path != "" {
			result.ProfileCheckPaths = append(result.ProfileCheckPaths, check.Path)
		}
		if !check.Passed {
			result.FailedChecks = append(result.FailedChecks, check)
		}
	}
	sort.Strings(result.ProfileCheckPaths)
	result.ProfileCheckPaths = compactStrings(result.ProfileCheckPaths)
	result.ProfileCheckCount = len(compile.ProfileChecks)
	result.ProfileCheckFailedCount = len(result.FailedChecks)
	result.OK = result.ValidateOK && result.CompileOK && result.ProfileCheckFailedCount == 0
	return result
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

func emptyRecipeSummary(reason, details string) RecipeProfileSummary {
	return RecipeProfileSummary{
		SchemaVersion: 1,
		Failed:        1,
		FailureReason: reason,
		Details:       details,
	}
}

func defaultCapabilities() recipe.CapabilityIndex {
	idx, err := capindex.Load(filepath.Join("docs", "capabilities.json"))
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

func printRecipeProfileSummary(w io.Writer, summary RecipeProfileSummary) {
	fmt.Fprintln(w, "recipe profile verification")
	fmt.Fprintf(w, "recipes: %d, passed: %d, failed: %d\n", summary.Total, summary.Passed, summary.Failed)
	fmt.Fprintf(w, "validate_failed: %d, compile_failed: %d, invalid_reports: %d, profile_check_failures: %d\n",
		summary.ValidateFailed, summary.CompileFailed, summary.InvalidReports, summary.ProfileCheckFailures)
	fmt.Fprintf(w, "covered_profile_paths: %d\n", len(summary.CoveredProfilePaths))
	if summary.FailureReason != "" {
		fmt.Fprintf(w, "failure_reason: %s\n", summary.FailureReason)
		if summary.Details != "" {
			fmt.Fprintln(w, summary.Details)
		}
		return
	}
	for _, result := range summary.Recipes {
		status := "fail"
		if result.OK {
			status = "ok"
		}
		fmt.Fprintf(w, "[%s] %s validate=%t compile=%t checks=%d failed=%d caps=%d\n",
			status, result.Recipe, result.ValidateOK, result.CompileOK, result.ProfileCheckCount, result.ProfileCheckFailedCount, result.CapabilityCount)
		for _, check := range result.FailedChecks {
			fmt.Fprintf(w, "  check failed: %s expected=%v actual=%v\n", check.Path, check.Expected, check.Actual)
		}
		if result.ValidateError != "" {
			fmt.Fprintf(w, "  validate_error: %s\n", result.ValidateError)
		}
		if result.CompileError != "" {
			fmt.Fprintf(w, "  compile_error: %s\n", result.CompileError)
		}
	}
}

func sortedKeys(values map[string]bool) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func compactStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := values[:1]
	for _, value := range values[1:] {
		if value != out[len(out)-1] {
			out = append(out, value)
		}
	}
	return out
}

func usage(stderr io.Writer) {
	fmt.Fprintln(stderr, "usage: aepverify <cross-platform|recipe-profiles> [flags]")
}
