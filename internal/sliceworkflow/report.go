// Package sliceworkflow builds reusable reports for AEP slice replication runs.
package sliceworkflow

import (
	"fmt"
	"sort"

	"github.com/example/aep-parser/internal/aeoracle"
	"github.com/example/aep-parser/internal/gapledger"
	"github.com/example/aep-parser/internal/profile"
	"github.com/example/aep-parser/internal/profilediff"
)

const (
	SchemaVersion    = 1
	DefaultMaxSlices = 3
)

type Classification string

const (
	ClassificationProcedural      Classification = "procedural"
	ClassificationFootageAssembly Classification = "footage-assembly"
	ClassificationPluginDependent Classification = "plugin-dependent"
	ClassificationMixed           Classification = "mixed"
)

type Options struct {
	SourceProject string
	ObservedIn    string
	MaxSlices     int
}

type Report struct {
	SchemaVersion int                `json:"schema_version"`
	SourceProject string             `json:"source_project,omitempty"`
	ObservedIn    string             `json:"observed_in,omitempty"`
	Mode          string             `json:"mode"`
	Summary       Summary            `json:"summary"`
	Slices        []Slice            `json:"slices,omitempty"`
	GapReports    []gapledger.Report `json:"gap_reports,omitempty"`
	Commands      []Command          `json:"commands,omitempty"`
}

type Summary struct {
	CompCount          int `json:"comp_count"`
	LayerCount         int `json:"layer_count"`
	FootageCount       int `json:"footage_count"`
	SelectedSliceCount int `json:"selected_slice_count"`
	GapCount           int `json:"gap_count,omitempty"`
}

type Slice struct {
	CompID                uint32         `json:"comp_id"`
	Name                  string         `json:"name"`
	ProfilePath           string         `json:"profile_path,omitempty"`
	Classification        Classification `json:"classification"`
	Score                 int            `json:"score"`
	Reasons               []string       `json:"reasons,omitempty"`
	LayerCount            int            `json:"layer_count"`
	TextLayerCount        int            `json:"text_layer_count,omitempty"`
	ShapeLayerCount       int            `json:"shape_layer_count,omitempty"`
	FootageLayerCount     int            `json:"footage_layer_count,omitempty"`
	EffectCount           int            `json:"effect_count,omitempty"`
	PluginDependencyCount int            `json:"plugin_dependency_count,omitempty"`
}

type Command struct {
	Name    string `json:"name"`
	Command string `json:"command"`
	Purpose string `json:"purpose,omitempty"`
}

func BuildReport(prof *profile.Profile, opts Options) (Report, error) {
	if prof == nil {
		return Report{}, fmt.Errorf("sliceworkflow: nil profile")
	}
	source := valueOr(opts.SourceProject, prof.Meta.Path)
	report := Report{
		SchemaVersion: SchemaVersion,
		SourceProject: source,
		Mode:          "plan",
		Summary: Summary{
			CompCount:    prof.Fingerprint.CompCount,
			LayerCount:   prof.Fingerprint.LayerCount,
			FootageCount: prof.Fingerprint.FootageCount,
		},
	}
	for _, comp := range prof.Comps {
		report.Slices = append(report.Slices, buildSlice(comp))
	}
	sort.Slice(report.Slices, func(i, j int) bool {
		left, right := report.Slices[i], report.Slices[j]
		if left.Score != right.Score {
			return left.Score > right.Score
		}
		if left.CompID != right.CompID {
			return left.CompID < right.CompID
		}
		return left.Name < right.Name
	})
	maxSlices := opts.MaxSlices
	if maxSlices <= 0 {
		maxSlices = DefaultMaxSlices
	}
	if len(report.Slices) > maxSlices {
		report.Slices = report.Slices[:maxSlices]
	}
	report.Summary.SelectedSliceCount = len(report.Slices)
	report.Commands = planCommands(source)
	return report, nil
}

func BuildDiagnoseReport(
	source *profile.Profile,
	observed *profile.Profile,
	diffReport *profilediff.Report,
	renderReports []aeoracle.CompareReport,
	opts Options,
) (Report, error) {
	if observed == nil {
		return Report{}, fmt.Errorf("sliceworkflow: nil observed profile")
	}
	report, err := BuildReport(source, opts)
	if err != nil {
		return Report{}, err
	}
	report.Mode = "diagnose"
	report.ObservedIn = valueOr(opts.ObservedIn, observed.Meta.Path)
	report.Commands = diagnoseCommands(report.SourceProject, report.ObservedIn)

	ctx := gapledger.Context{SourceProject: report.SourceProject, ObservedIn: report.ObservedIn}
	if diffGapReport := gapledger.FromDiffReport(diffReport, ctx); diffGapReport.GapCount > 0 {
		report.Summary.GapCount += diffGapReport.GapCount
		report.GapReports = append(report.GapReports, diffGapReport)
	}
	for _, renderReport := range renderReports {
		renderGapReport := gapledger.FromRenderCompare(renderReport, ctx)
		if renderGapReport.GapCount == 0 {
			continue
		}
		report.Summary.GapCount += renderGapReport.GapCount
		report.GapReports = append(report.GapReports, renderGapReport)
	}
	return report, nil
}

func buildSlice(comp profile.Composition) Slice {
	slice := Slice{
		CompID:      comp.ID,
		Name:        comp.Name,
		ProfilePath: comp.Path.Path,
		LayerCount:  len(comp.Layers),
	}
	for _, layer := range comp.Layers {
		if layer.Text != nil {
			slice.TextLayerCount++
		}
		if len(layer.Shapes) > 0 {
			slice.ShapeLayerCount++
		}
		if layer.SourceRef != nil && layer.SourceRef.Kind == "footage" {
			slice.FootageLayerCount++
		}
		for _, effect := range layer.Effects {
			slice.EffectCount++
			if effect.DependencyClass == "third_party" {
				slice.PluginDependencyCount++
			}
		}
	}
	slice.Classification = classifySlice(slice)
	slice.Score = scoreSlice(slice)
	slice.Reasons = sliceReasons(slice)
	return slice
}

func classifySlice(slice Slice) Classification {
	if slice.PluginDependencyCount > 0 {
		return ClassificationPluginDependent
	}
	if slice.FootageLayerCount > 0 {
		return ClassificationFootageAssembly
	}
	return ClassificationProcedural
}

func scoreSlice(slice Slice) int {
	score := slice.LayerCount
	switch slice.Classification {
	case ClassificationPluginDependent:
		score += 100
	case ClassificationFootageAssembly:
		score += 50
	case ClassificationMixed:
		score += 75
	}
	score += slice.EffectCount * 10
	score += slice.TextLayerCount * 5
	score += slice.ShapeLayerCount * 5
	return score
}

func sliceReasons(slice Slice) []string {
	var reasons []string
	if slice.PluginDependencyCount > 0 {
		reasons = append(reasons, "plugin-dependency")
	}
	if slice.FootageLayerCount > 0 {
		reasons = append(reasons, "footage-layer")
	}
	if slice.TextLayerCount > 0 {
		reasons = append(reasons, "text-layer")
	}
	if slice.ShapeLayerCount > 0 {
		reasons = append(reasons, "shape-layer")
	}
	if slice.EffectCount > 0 {
		reasons = append(reasons, "effect")
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "composition")
	}
	return reasons
}

func planCommands(source string) []Command {
	return []Command{
		{
			Name:    "profile",
			Command: fmt.Sprintf("go run ./cmd/aepdissect -json %s", source),
			Purpose: "inspect the normalized profile before choosing a clone slice",
		},
		{
			Name:    "render-plan",
			Command: fmt.Sprintf("go run ./cmd/aeoracle plan -aep %s -out tmp_debug/aeoracle/slice -json", source),
			Purpose: "create sentinel render frames for the selected slice",
		},
	}
}

func diagnoseCommands(source, observed string) []Command {
	commands := planCommands(source)
	commands = append(commands,
		Command{
			Name:    "diff-gaps",
			Command: fmt.Sprintf("go run ./cmd/aepgaps diff %s %s", source, observed),
			Purpose: "convert structural profile differences into gap entries",
		},
		Command{
			Name:    "render-gaps",
			Command: "go run ./cmd/aepgaps render -expected expected.png -actual actual.png",
			Purpose: "convert selected render differences into gap entries",
		},
	)
	return commands
}

func valueOr(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
