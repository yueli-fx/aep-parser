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
	summaryMode := fs.Bool("summary", false, "emit a single aggregate JSON summary in corpus mode")
	outPath := fs.String("out", "", "optional output file")
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
		return runCorpus(*input, *mode, *recursive, *limit, *summaryMode, *outPath, stdout, stderr)
	}

	output, code := buildOutput(*input, *mode, stderr)
	if code != 0 {
		return code
	}
	out, closeOut, err := outputWriter(*outPath, stdout)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	defer closeOut()
	enc := json.NewEncoder(out)
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

type corpusSummary struct {
	SchemaVersion int                          `json:"schema_version"`
	Mode          string                       `json:"mode"`
	ProjectCount  int                          `json:"project_count"`
	ErrorCount    int                          `json:"error_count,omitempty"`
	Totals        technique.FingerprintSummary `json:"totals"`
	HintCounts    map[string]int               `json:"hint_counts"`
	EffectCounts  map[string]int               `json:"effect_counts"`
	ShapeFamilies map[string]int               `json:"shape_families"`
	TextAnimators map[string]int               `json:"text_animators"`
	LayerRoles    map[string]int               `json:"layer_roles"`
	GraphEdges    map[string]int               `json:"graph_edges"`
}

func newCorpusSummary(mode string) *corpusSummary {
	return &corpusSummary{
		SchemaVersion: technique.SchemaVersion,
		Mode:          mode,
		Totals: technique.FingerprintSummary{
			LayerRoleCounts: map[string]int{},
		},
		HintCounts:    map[string]int{},
		EffectCounts:  map[string]int{},
		ShapeFamilies: map[string]int{},
		TextAnimators: map[string]int{},
		LayerRoles:    map[string]int{},
		GraphEdges:    map[string]int{},
	}
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

func runCorpus(input, mode string, recursive bool, limit int, summaryMode bool, outPath string, stdout, stderr io.Writer) int {
	paths, err := discoverAEPs(input, recursive)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if limit > 0 && len(paths) > limit {
		paths = paths[:limit]
	}
	out, closeOut, err := outputWriter(outPath, stdout)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	defer closeOut()
	enc := json.NewEncoder(out)
	summary := newCorpusSummary(mode)
	hadError := false
	for _, path := range paths {
		record := corpusRecord{Path: path, Mode: mode}
		output, code := buildOutput(path, mode, stderr)
		if code != 0 {
			hadError = true
			summary.ErrorCount++
			record.Error = fmt.Sprintf("build failed with exit code %d", code)
		} else {
			switch value := output.(type) {
			case *technique.FactSet:
				record.Facts = value
				addFactsToSummary(summary, value)
			case *technique.Portrait:
				record.Portrait = value
				addPortraitToSummary(summary, value)
			}
		}
		if !summaryMode {
			if err := enc.Encode(record); err != nil {
				fmt.Fprintf(stderr, "jsonl: %v\n", err)
				return 1
			}
		}
	}
	if summaryMode {
		enc.SetIndent("", "  ")
		if err := enc.Encode(summary); err != nil {
			fmt.Fprintf(stderr, "json: %v\n", err)
			return 1
		}
	}
	if hadError {
		return 1
	}
	return 0
}

func outputWriter(path string, stdout io.Writer) (io.Writer, func(), error) {
	if path == "" {
		return stdout, func() {}, nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, nil, err
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}
	return file, func() { _ = file.Close() }, nil
}

func addFactsToSummary(summary *corpusSummary, facts *technique.FactSet) {
	if facts == nil {
		return
	}
	summary.ProjectCount++
	summary.Totals.CompCount += facts.Summary.CompCount
	summary.Totals.LayerCount += facts.Summary.LayerCount
	summary.Totals.EffectCount += facts.Summary.EffectCount
	summary.Totals.TextLayerCount += facts.Summary.TextLayerCount
	summary.Totals.ShapeLayerCount += facts.Summary.ShapeLayerCount
	summary.Totals.TextAnimatorCount += len(facts.TextAnimators)
	summary.Totals.ShapeOperatorCount += len(facts.ShapeOperators)
	summary.Totals.DependencyCount += len(facts.Dependencies)
	summary.Totals.UnknownCount += len(facts.Unknowns)
	for _, layer := range facts.Layers {
		if layer.Role != "" {
			summary.LayerRoles[layer.Role]++
			summary.Totals.LayerRoleCounts[layer.Role]++
		}
	}
	for _, effect := range facts.Effects {
		summary.EffectCounts[effect.MatchName]++
	}
	for _, operator := range facts.ShapeOperators {
		summary.ShapeFamilies[operator.Family]++
	}
	for _, animator := range facts.TextAnimators {
		summary.TextAnimators[animator.PropertyKind]++
	}
	for _, dep := range facts.Dependencies {
		summary.GraphEdges[dep.Relation]++
	}
}

func addPortraitToSummary(summary *corpusSummary, portrait *technique.Portrait) {
	if portrait == nil {
		return
	}
	summary.ProjectCount++
	summary.Totals.CompCount += portrait.Fingerprint.CompCount
	summary.Totals.LayerCount += portrait.Fingerprint.LayerCount
	summary.Totals.EffectCount += portrait.Fingerprint.EffectCount
	summary.Totals.TextLayerCount += portrait.Fingerprint.TextLayerCount
	summary.Totals.ShapeLayerCount += portrait.Fingerprint.ShapeLayerCount
	summary.Totals.TextAnimatorCount += portrait.Fingerprint.TextAnimatorCount
	summary.Totals.ShapeOperatorCount += portrait.Fingerprint.ShapeOperatorCount
	summary.Totals.DependencyCount += portrait.Fingerprint.DependencyCount
	summary.Totals.UnknownCount += portrait.Fingerprint.UnknownCount
	addCounts(summary.LayerRoles, portrait.Fingerprint.LayerRoleCounts)
	addCounts(summary.Totals.LayerRoleCounts, portrait.Fingerprint.LayerRoleCounts)
	addCounts(summary.EffectCounts, portrait.Mechanisms.EffectMatchCounts)
	addCounts(summary.ShapeFamilies, portrait.Mechanisms.ShapeFamilyCounts)
	addCounts(summary.TextAnimators, portrait.Mechanisms.TextAnimatorKindCounts)
	addCounts(summary.GraphEdges, portrait.Graph.RelationCounts)
	for _, hint := range portrait.TechniqueHints {
		summary.HintCounts[hint.ID]++
	}
}

func addCounts(dst, src map[string]int) {
	for key, value := range src {
		dst[key] += value
	}
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
