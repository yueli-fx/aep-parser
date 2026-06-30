package selfhost

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type CompareOptions struct {
	BaseDir string
	NewDir  string
	OutDir  string
	Top     int
}

type CompareResult struct {
	SchemaVersion int          `json:"schema_version"`
	BaseDir       string       `json:"base_dir"`
	NewDir        string       `json:"new_dir"`
	ScalarDiffs   []ScalarDiff `json:"scalar_diffs"`
	CountDiffs    []CountDiff  `json:"count_diffs"`
}

type ScalarDiff struct {
	Name  string `json:"name"`
	Base  int    `json:"base"`
	New   int    `json:"new"`
	Delta int    `json:"delta"`
}

type CountDiff struct {
	Group string `json:"group"`
	Name  string `json:"name"`
	Base  int    `json:"base"`
	New   int    `json:"new"`
	Delta int    `json:"delta"`
}

type reportSummary struct {
	ProjectCount       int            `json:"project_count"`
	ErrorCount         int            `json:"error_count"`
	Totals             map[string]any `json:"totals"`
	ReadinessCounts    map[string]any `json:"readiness_counts"`
	PatternCounts      map[string]any `json:"pattern_counts"`
	ArchetypeCounts    map[string]any `json:"archetype_counts"`
	HintCounts         map[string]any `json:"hint_counts"`
	PluginEffectCounts map[string]any `json:"plugin_effect_counts"`
	EffectCounts       map[string]any `json:"effect_counts"`
	ShapeFamilies      map[string]any `json:"shape_families"`
	TextAnimators      map[string]any `json:"text_animators"`
	LayerRoles         map[string]any `json:"layer_roles"`
	GraphEdges         map[string]any `json:"graph_edges"`
}

type reportDigest struct {
	Patterns []any `json:"patterns"`
}

func CompareReports(opts CompareOptions) (CompareResult, error) {
	if opts.Top <= 0 {
		opts.Top = 20
	}
	if opts.OutDir == "" {
		opts.OutDir = filepath.Join(opts.NewDir, "compare")
	}
	baseSummary, err := readReportSummary(filepath.Join(opts.BaseDir, "summary.json"))
	if err != nil {
		return CompareResult{}, err
	}
	newSummary, err := readReportSummary(filepath.Join(opts.NewDir, "summary.json"))
	if err != nil {
		return CompareResult{}, err
	}
	baseDigest, err := readReportDigest(filepath.Join(opts.BaseDir, "digest.json"))
	if err != nil {
		return CompareResult{}, err
	}
	newDigest, err := readReportDigest(filepath.Join(opts.NewDir, "digest.json"))
	if err != nil {
		return CompareResult{}, err
	}

	result := CompareResult{
		SchemaVersion: 1,
		BaseDir:       opts.BaseDir,
		NewDir:        opts.NewDir,
		CountDiffs:    []CountDiff{},
		ScalarDiffs: []ScalarDiff{
			scalarDiff("projects", baseSummary.ProjectCount, newSummary.ProjectCount),
			scalarDiff("errors", baseSummary.ErrorCount, newSummary.ErrorCount),
			scalarDiff("comps", intFromMap(baseSummary.Totals, "comp_count"), intFromMap(newSummary.Totals, "comp_count")),
			scalarDiff("layers", intFromMap(baseSummary.Totals, "layer_count"), intFromMap(newSummary.Totals, "layer_count")),
			scalarDiff("effects", intFromMap(baseSummary.Totals, "effect_count"), intFromMap(newSummary.Totals, "effect_count")),
			scalarDiff("text_animators", intFromMap(baseSummary.Totals, "text_animator_count"), intFromMap(newSummary.Totals, "text_animator_count")),
			scalarDiff("shape_operators", intFromMap(baseSummary.Totals, "shape_operator_count"), intFromMap(newSummary.Totals, "shape_operator_count")),
			scalarDiff("dependency_edges", intFromMap(baseSummary.Totals, "dependency_count"), intFromMap(newSummary.Totals, "dependency_count")),
			scalarDiff("patterns", len(baseDigest.Patterns), len(newDigest.Patterns)),
		},
	}
	for _, group := range []struct {
		name string
		base map[string]any
		next map[string]any
	}{
		{name: "readiness", base: baseSummary.ReadinessCounts, next: newSummary.ReadinessCounts},
		{name: "patterns", base: baseSummary.PatternCounts, next: newSummary.PatternCounts},
		{name: "archetypes", base: baseSummary.ArchetypeCounts, next: newSummary.ArchetypeCounts},
		{name: "technique_hints", base: baseSummary.HintCounts, next: newSummary.HintCounts},
		{name: "plugin_effects", base: baseSummary.PluginEffectCounts, next: newSummary.PluginEffectCounts},
		{name: "effects", base: baseSummary.EffectCounts, next: newSummary.EffectCounts},
		{name: "shape_families", base: baseSummary.ShapeFamilies, next: newSummary.ShapeFamilies},
		{name: "text_animators", base: baseSummary.TextAnimators, next: newSummary.TextAnimators},
		{name: "layer_roles", base: baseSummary.LayerRoles, next: newSummary.LayerRoles},
		{name: "graph_edges", base: baseSummary.GraphEdges, next: newSummary.GraphEdges},
	} {
		result.CountDiffs = append(result.CountDiffs, compareCountMap(group.name, group.base, group.next, opts.Top)...)
	}
	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return CompareResult{}, err
	}
	if err := writeIndentedJSON(filepath.Join(opts.OutDir, "compare.json"), result); err != nil {
		return CompareResult{}, err
	}
	if err := os.WriteFile(filepath.Join(opts.OutDir, "compare.md"), []byte(formatCompareMarkdown(result)), 0o644); err != nil {
		return CompareResult{}, err
	}
	return result, nil
}

func readReportSummary(path string) (reportSummary, error) {
	var summary reportSummary
	if err := readIndentedJSON(path, &summary); err != nil {
		return reportSummary{}, fmt.Errorf("read %s: %w", path, err)
	}
	return summary, nil
}

func readReportDigest(path string) (reportDigest, error) {
	var digest reportDigest
	if err := readIndentedJSON(path, &digest); err != nil {
		return reportDigest{}, fmt.Errorf("read %s: %w", path, err)
	}
	return digest, nil
}

func readIndentedJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func writeIndentedJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func scalarDiff(name string, base, next int) ScalarDiff {
	return ScalarDiff{Name: name, Base: base, New: next, Delta: next - base}
}

func compareCountMap(group string, baseCounts, newCounts map[string]any, max int) []CountDiff {
	keys := map[string]struct{}{}
	for key := range baseCounts {
		keys[key] = struct{}{}
	}
	for key := range newCounts {
		keys[key] = struct{}{}
	}
	rows := make([]CountDiff, 0, len(keys))
	for key := range keys {
		base := intFromMap(baseCounts, key)
		next := intFromMap(newCounts, key)
		if base == next {
			continue
		}
		rows = append(rows, CountDiff{
			Group: group,
			Name:  key,
			Base:  base,
			New:   next,
			Delta: next - base,
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		left := math.Abs(float64(rows[i].Delta))
		right := math.Abs(float64(rows[j].Delta))
		if left != right {
			return left > right
		}
		return rows[i].Name < rows[j].Name
	})
	if max > 0 && len(rows) > max {
		rows = rows[:max]
	}
	return rows
}

func intFromMap(values map[string]any, key string) int {
	if values == nil {
		return 0
	}
	switch value := values[key].(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	case json.Number:
		parsed, _ := value.Int64()
		return int(parsed)
	default:
		return 0
	}
}

func formatCompareMarkdown(result CompareResult) string {
	var b strings.Builder
	fmt.Fprintln(&b, "# Technique Report Compare")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "- base: `%s`\n", result.BaseDir)
	fmt.Fprintf(&b, "- new: `%s`\n", result.NewDir)
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## Scalar Diffs")
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "| metric | base | new | delta |")
	fmt.Fprintln(&b, "| --- | ---: | ---: | ---: |")
	for _, row := range result.ScalarDiffs {
		fmt.Fprintf(&b, "| %s | %d | %d | %d |\n", row.Name, row.Base, row.New, row.Delta)
	}
	fmt.Fprintln(&b)
	fmt.Fprintln(&b, "## Count Diffs")
	fmt.Fprintln(&b)
	if len(result.CountDiffs) == 0 {
		fmt.Fprintln(&b, "_no count changes_")
		return b.String()
	}
	fmt.Fprintln(&b, "| group | name | base | new | delta |")
	fmt.Fprintln(&b, "| --- | --- | ---: | ---: | ---: |")
	for _, row := range result.CountDiffs {
		fmt.Fprintf(&b, "| %s | %s | %d | %d | %d |\n", row.Group, row.Name, row.Base, row.New, row.Delta)
	}
	return b.String()
}
