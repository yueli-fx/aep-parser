package selfhost

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
	"github.com/yueli-fx/aep-parser/internal/profilediff"
	"github.com/yueli-fx/aep-parser/internal/recipe"
	"github.com/yueli-fx/aep-parser/internal/rifx"
)

type SampleShellOptions struct {
	InputPath     string
	OutDir        string
	TargetVersion string
}

type SampleShellResult struct {
	SchemaVersion        int                     `json:"schema_version"`
	InputPath            string                  `json:"input_path"`
	OutDir               string                  `json:"out_dir"`
	RecipePath           string                  `json:"recipe_path"`
	CompiledAEPPath      string                  `json:"compiled_aep_path"`
	OriginalProfilePath  string                  `json:"original_profile_path"`
	GeneratedProfilePath string                  `json:"generated_profile_path"`
	CompileReportPath    string                  `json:"compile_report_path"`
	RawDiffPath          string                  `json:"raw_diff_path"`
	SemanticDiffPath     string                  `json:"semantic_diff_path"`
	CompileReport        recipe.Report           `json:"compile_report"`
	RawDiff              profilediff.Report      `json:"raw_diff"`
	SemanticDiff         SampleShellSemanticDiff `json:"semantic_diff"`
}

type SampleShellBatchOptions struct {
	Root          string
	OutDir        string
	TargetVersion string
	Limit         int
}

type SampleShellBatchResult struct {
	SchemaVersion            int                         `json:"schema_version"`
	Root                     string                      `json:"root"`
	OutDir                   string                      `json:"out_dir"`
	TargetVersion            string                      `json:"target_version"`
	SummaryPath              string                      `json:"summary_path,omitempty"`
	Summary                  SampleShellBatchSummary     `json:"summary"`
	MissingEffectNames       []CountRow                  `json:"missing_effect_match_names"`
	UnsupportedEffectClasses []CountRow                  `json:"unsupported_effect_classes"`
	EffectWorkItems          []SampleShellEffectWorkItem `json:"effect_work_items"`
	Runs                     []SampleShellBatchRun       `json:"runs"`
	Failures                 []SampleShellBatchFailure   `json:"failures"`
}

type SampleShellBatchSummary struct {
	Total               int       `json:"total"`
	Succeeded           int       `json:"succeeded"`
	Failed              int       `json:"failed"`
	CompCount           CountPair `json:"comp_count"`
	LayerCount          CountPair `json:"layer_count"`
	ShapeLayerCount     CountPair `json:"shape_layer_count"`
	EffectCount         CountPair `json:"effect_count"`
	EffectParamCount    CountPair `json:"effect_param_count"`
	FootageCount        CountPair `json:"footage_count"`
	RawProfileDiffCount int       `json:"raw_profile_diff_count"`
}

type SampleShellBatchRun struct {
	InputPath        string             `json:"input_path"`
	OutDir           string             `json:"out_dir"`
	SemanticDiffPath string             `json:"semantic_diff_path"`
	Summary          SampleShellSummary `json:"summary"`
}

type SampleShellBatchFailure struct {
	InputPath string `json:"input_path"`
	OutDir    string `json:"out_dir"`
	Error     string `json:"error"`
}

type SampleShellSemanticDiff struct {
	SchemaVersion        int                  `json:"schema_version"`
	Sample               string               `json:"sample"`
	GeneratedAEP         string               `json:"generated_aep"`
	Level                string               `json:"level"`
	Summary              SampleShellSummary   `json:"summary"`
	CompLayerAlignment   []CompLayerAlignment `json:"comp_layer_alignment"`
	OriginalLayerTypes   []CountRow           `json:"original_layer_types"`
	GeneratedLayerTypes  []CountRow           `json:"generated_layer_types"`
	OriginalShapeKinds   []CountRow           `json:"original_shape_kinds"`
	GeneratedShapeKinds  []CountRow           `json:"generated_shape_kinds"`
	MissingEffectNames   []CountRow           `json:"missing_effect_match_names"`
	MaterializedBoundary []string             `json:"materialized_boundaries"`
	RemainingGaps        []string             `json:"remaining_gaps"`
}

type SampleShellSummary struct {
	CompCount           CountPair      `json:"comp_count"`
	LayerCount          CountPair      `json:"layer_count"`
	ShapeLayerCount     CountPair      `json:"shape_layer_count"`
	EffectCount         CountPair      `json:"effect_count"`
	EffectParamCount    CountPair      `json:"effect_param_count"`
	FootageCount        CountPair      `json:"footage_count"`
	RawProfileDiffCount RawDiffSummary `json:"raw_profile_diff_count"`
}

type CountPair struct {
	Original      int `json:"original"`
	Generated     int `json:"generated"`
	MatchedByName int `json:"matched_by_name,omitempty"`
}

type RawDiffSummary struct {
	Shell int    `json:"shell"`
	Note  string `json:"note"`
}

type CompLayerAlignment struct {
	Name            string `json:"name"`
	Matched         bool   `json:"matched"`
	OriginalLayers  int    `json:"original_layers"`
	GeneratedLayers int    `json:"generated_layers"`
	TypeMismatches  int    `json:"type_mismatches,omitempty"`
	NameMismatches  int    `json:"name_mismatches,omitempty"`
}

type CountRow struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type SampleShellEffectWorkItem struct {
	MatchName string `json:"match_name"`
	Count     int    `json:"count"`
	Class     string `json:"class"`
	Action    string `json:"action"`
}

type SampleShellEffectTemplateExtractionOptions struct {
	SummaryPath string
	Root        string
	OutDir      string
}

type SampleShellEffectTemplateExtractionResult struct {
	SchemaVersion int                            `json:"schema_version"`
	SummaryPath   string                         `json:"summary_path"`
	Root          string                         `json:"root"`
	OutDir        string                         `json:"out_dir"`
	ResultPath    string                         `json:"result_path,omitempty"`
	Requested     []string                       `json:"requested"`
	Hits          []SampleShellEffectTemplateHit `json:"hits"`
	Missing       []string                       `json:"missing"`
}

type SampleShellEffectTemplateHit struct {
	MatchName  string `json:"match_name"`
	SourcePath string `json:"source_path"`
	OutPath    string `json:"out_path"`
	Bytes      int    `json:"bytes"`
}

func RunSampleShellBatch(opts SampleShellBatchOptions) (SampleShellBatchResult, error) {
	if opts.Root == "" {
		return SampleShellBatchResult{}, fmt.Errorf("sample shell batch: root is required")
	}
	if opts.OutDir == "" {
		opts.OutDir = filepath.Join("tmp", "sample_shell_batch")
	}
	if opts.TargetVersion == "" {
		opts.TargetVersion = "AE2020"
	}
	inputs, err := DiscoverSampleShellInputs(opts.Root, opts.Limit)
	if err != nil {
		return SampleShellBatchResult{}, err
	}
	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return SampleShellBatchResult{}, err
	}
	results := make([]SampleShellResult, 0, len(inputs))
	failures := make([]SampleShellBatchFailure, 0)
	for i, input := range inputs {
		runOutDir := filepath.Join(opts.OutDir, sampleShellBatchDirName(i+1, input))
		result, err := RunSampleShell(SampleShellOptions{
			InputPath:     input,
			OutDir:        runOutDir,
			TargetVersion: opts.TargetVersion,
		})
		if err != nil {
			failures = append(failures, SampleShellBatchFailure{
				InputPath: input,
				OutDir:    runOutDir,
				Error:     err.Error(),
			})
			continue
		}
		results = append(results, result)
	}
	batch := BuildSampleShellBatchSummary(opts.Root, opts.OutDir, opts.TargetVersion, results, failures)
	summaryPath := filepath.Join(opts.OutDir, "batch_summary.json")
	batch.SummaryPath = summaryPath
	if err := writeIndentedJSON(summaryPath, batch); err != nil {
		return SampleShellBatchResult{}, err
	}
	return batch, nil
}

func DiscoverSampleShellInputs(root string, limit int) ([]string, error) {
	var inputs []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".aep" {
			inputs = append(inputs, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(inputs)
	if limit > 0 && len(inputs) > limit {
		inputs = inputs[:limit]
	}
	return inputs, nil
}

func RunSampleShellEffectTemplateExtraction(opts SampleShellEffectTemplateExtractionOptions) (SampleShellEffectTemplateExtractionResult, error) {
	if opts.SummaryPath == "" {
		return SampleShellEffectTemplateExtractionResult{}, fmt.Errorf("effect template extraction: summary path is required")
	}
	if opts.OutDir == "" {
		opts.OutDir = filepath.Join("tmp", "effect_template_candidates")
	}
	var batch SampleShellBatchResult
	data, err := os.ReadFile(opts.SummaryPath)
	if err != nil {
		return SampleShellEffectTemplateExtractionResult{}, err
	}
	if err := json.Unmarshal(data, &batch); err != nil {
		return SampleShellEffectTemplateExtractionResult{}, err
	}
	if opts.Root == "" {
		opts.Root = batch.Root
	}
	if opts.Root == "" {
		return SampleShellEffectTemplateExtractionResult{}, fmt.Errorf("effect template extraction: root is required")
	}
	requested := sampleShellEffectTemplateRequests(batch.EffectWorkItems)
	result := SampleShellEffectTemplateExtractionResult{
		SchemaVersion: 1,
		SummaryPath:   opts.SummaryPath,
		Root:          opts.Root,
		OutDir:        opts.OutDir,
		Requested:     requested,
		Hits:          []SampleShellEffectTemplateHit{},
		Missing:       []string{},
	}
	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return SampleShellEffectTemplateExtractionResult{}, err
	}
	inputs, err := DiscoverSampleShellInputs(opts.Root, 0)
	if err != nil {
		return SampleShellEffectTemplateExtractionResult{}, err
	}
	remaining := map[string]struct{}{}
	for _, matchName := range requested {
		remaining[matchName] = struct{}{}
	}
	for _, input := range inputs {
		if len(remaining) == 0 {
			break
		}
		root, err := sampleShellReadRIFX(input)
		if err != nil {
			continue
		}
		hits := sampleShellFindEffectTemplateCandidates(root, remaining)
		for _, matchName := range requested {
			chunk, ok := hits[matchName]
			if !ok {
				continue
			}
			outPath := filepath.Join(opts.OutDir, sampleShellEffectTemplateCandidateFileName(matchName))
			data, err := sampleShellChunkBytes(chunk)
			if err != nil {
				return SampleShellEffectTemplateExtractionResult{}, err
			}
			if err := os.WriteFile(outPath, data, 0o644); err != nil {
				return SampleShellEffectTemplateExtractionResult{}, err
			}
			result.Hits = append(result.Hits, SampleShellEffectTemplateHit{
				MatchName:  matchName,
				SourcePath: input,
				OutPath:    outPath,
				Bytes:      len(data),
			})
			delete(remaining, matchName)
		}
	}
	for _, matchName := range requested {
		if _, ok := remaining[matchName]; ok {
			result.Missing = append(result.Missing, matchName)
		}
	}
	result.ResultPath = filepath.Join(opts.OutDir, "extraction_summary.json")
	if err := writeIndentedJSON(result.ResultPath, result); err != nil {
		return SampleShellEffectTemplateExtractionResult{}, err
	}
	return result, nil
}

func BuildSampleShellBatchSummary(root, outDir, targetVersion string, results []SampleShellResult, failures []SampleShellBatchFailure) SampleShellBatchResult {
	if targetVersion == "" {
		targetVersion = "AE2020"
	}
	batch := SampleShellBatchResult{
		SchemaVersion:            1,
		Root:                     root,
		OutDir:                   outDir,
		TargetVersion:            targetVersion,
		MissingEffectNames:       []CountRow{},
		UnsupportedEffectClasses: []CountRow{},
		EffectWorkItems:          []SampleShellEffectWorkItem{},
		Runs:                     make([]SampleShellBatchRun, 0, len(results)),
		Failures:                 append([]SampleShellBatchFailure(nil), failures...),
	}
	missingEffects := map[string]int{}
	unsupportedClasses := map[string]int{}
	batch.Summary.Total = len(results) + len(failures)
	batch.Summary.Succeeded = len(results)
	batch.Summary.Failed = len(failures)
	for _, result := range results {
		summary := result.SemanticDiff.Summary
		batch.Runs = append(batch.Runs, SampleShellBatchRun{
			InputPath:        result.InputPath,
			OutDir:           result.OutDir,
			SemanticDiffPath: result.SemanticDiffPath,
			Summary:          summary,
		})
		addCountPair(&batch.Summary.CompCount, summary.CompCount)
		addCountPair(&batch.Summary.LayerCount, summary.LayerCount)
		addCountPair(&batch.Summary.ShapeLayerCount, summary.ShapeLayerCount)
		addCountPair(&batch.Summary.EffectCount, summary.EffectCount)
		addCountPair(&batch.Summary.EffectParamCount, summary.EffectParamCount)
		addCountPair(&batch.Summary.FootageCount, summary.FootageCount)
		batch.Summary.RawProfileDiffCount += summary.RawProfileDiffCount.Shell
		for _, row := range result.SemanticDiff.MissingEffectNames {
			missingEffects[row.Name] += row.Count
			unsupportedClasses[ClassifySampleShellUnsupportedEffect(row.Name)] += row.Count
		}
	}
	batch.MissingEffectNames = countMapRows(missingEffects)
	batch.UnsupportedEffectClasses = countMapRows(unsupportedClasses)
	batch.EffectWorkItems = sampleShellEffectWorkItems(missingEffects)
	return batch
}

func RunSampleShell(opts SampleShellOptions) (SampleShellResult, error) {
	if opts.InputPath == "" {
		return SampleShellResult{}, fmt.Errorf("sample shell: input path is required")
	}
	if opts.OutDir == "" {
		opts.OutDir = filepath.Join("tmp", "sample_shell")
	}
	if opts.TargetVersion == "" {
		opts.TargetVersion = "AE2020"
	}
	if err := os.MkdirAll(opts.OutDir, 0o755); err != nil {
		return SampleShellResult{}, err
	}

	original, err := openProfile(opts.InputPath)
	if err != nil {
		return SampleShellResult{}, err
	}
	recipeDoc := BuildSampleShellRecipe(original, opts.TargetVersion)

	recipePath := filepath.Join(opts.OutDir, "shell_recipe.json")
	compiledPath := filepath.Join(opts.OutDir, "shell.aep")
	originalProfilePath := filepath.Join(opts.OutDir, "original_profile.json")
	generatedProfilePath := filepath.Join(opts.OutDir, "generated_profile.json")
	compileReportPath := filepath.Join(opts.OutDir, "compile_report.json")
	rawDiffPath := filepath.Join(opts.OutDir, "raw_profile_diff.json")
	semanticDiffPath := filepath.Join(opts.OutDir, "semantic_diff.json")

	if err := writeSampleShellJSON(recipePath, recipeDoc); err != nil {
		return SampleShellResult{}, err
	}
	if err := writeSampleShellJSON(originalProfilePath, original); err != nil {
		return SampleShellResult{}, err
	}
	compileReport, err := recipe.CompileToFile(recipeDoc, compiledPath, recipe.StaticCapabilities{})
	if err != nil {
		return SampleShellResult{}, err
	}
	if err := writeSampleShellJSON(compileReportPath, compileReport); err != nil {
		return SampleShellResult{}, err
	}
	if !compileReport.Valid {
		return SampleShellResult{}, fmt.Errorf("sample shell: recipe compile report is invalid: %s", compileReportPath)
	}

	generated, err := openProfile(compiledPath)
	if err != nil {
		return SampleShellResult{}, err
	}
	if err := writeSampleShellJSON(generatedProfilePath, generated); err != nil {
		return SampleShellResult{}, err
	}
	rawDiff, err := profilediff.Compare(original, generated, profilediff.Options{})
	if err != nil {
		return SampleShellResult{}, err
	}
	if err := writeSampleShellJSON(rawDiffPath, rawDiff); err != nil {
		return SampleShellResult{}, err
	}
	semantic := BuildSampleShellSemanticDiff(original, generated, opts.InputPath, compiledPath, rawDiff.DiffCount)
	semantic.Summary.EffectParamCount = CountPair{
		Original:  countSampleShellCopyableEffectParams(original),
		Generated: countRecipeEffectParams(recipeDoc),
	}
	if err := writeSampleShellJSON(semanticDiffPath, semantic); err != nil {
		return SampleShellResult{}, err
	}

	result := SampleShellResult{
		SchemaVersion:        1,
		InputPath:            opts.InputPath,
		OutDir:               opts.OutDir,
		RecipePath:           recipePath,
		CompiledAEPPath:      compiledPath,
		OriginalProfilePath:  originalProfilePath,
		GeneratedProfilePath: generatedProfilePath,
		CompileReportPath:    compileReportPath,
		RawDiffPath:          rawDiffPath,
		SemanticDiffPath:     semanticDiffPath,
		CompileReport:        compileReport,
		RawDiff:              *rawDiff,
		SemanticDiff:         semantic,
	}
	if err := writeSampleShellJSON(filepath.Join(opts.OutDir, "summary.json"), result); err != nil {
		return SampleShellResult{}, err
	}
	return result, nil
}

func BuildSampleShellRecipe(prof *profile.Profile, targetVersion string) recipe.Recipe {
	if targetVersion == "" {
		targetVersion = "AE2020"
	}
	rec := recipe.Recipe{
		SchemaVersion: recipe.SchemaVersion,
		Project: recipe.ProjectSpec{
			Name:          projectNameFromProfile(prof),
			TargetVersion: targetVersion,
		},
	}
	compNamesByID := sampleShellCompNames(prof)
	var layerCount, shapeLayerCount int
	for i, comp := range prof.Comps {
		duration := finiteFloat(comp.Duration, 1)
		compSpec := recipe.CompSpec{
			Name:                       compNamesByID[sampleShellCompKey(i, comp)],
			Width:                      int(comp.Width),
			Height:                     int(comp.Height),
			FrameRate:                  finiteFloat(comp.FrameRate, 30),
			Duration:                   duration,
			BackgroundColor:            []float64{float64(comp.BackgroundColor[0]), float64(comp.BackgroundColor[1]), float64(comp.BackgroundColor[2])},
			MotionGraphicsTemplateName: comp.MotionGraphicsTemplateName,
			Renderer:                   comp.Renderer,
			ResolutionFactor:           []float64{float64(comp.ResolutionFactor[0]), float64(comp.ResolutionFactor[1])},
			PixelAspect:                floatPtr(finiteFloat(comp.PixelAspect, 1)),
			DisplayStartTime:           floatPtr(finiteFloat(comp.DisplayStartTime, 0)),
			FrameBlending:              boolPtr(comp.FrameBlending),
			HideShyLayers:              boolPtr(comp.HideShyLayers),
			PreserveNestedFrameRate:    boolPtr(comp.PreserveNestedFrameRate),
			PreserveNestedResolution:   boolPtr(comp.PreserveNestedResolution),
			MotionBlur: &recipe.CompMotionBlurSpec{
				Enabled:             boolPtr(comp.MotionBlur.Enabled),
				ShutterAngle:        floatPtr(float64(comp.MotionBlur.ShutterAngle)),
				ShutterPhase:        floatPtr(float64(comp.MotionBlur.ShutterPhase)),
				AdaptiveSampleLimit: floatPtr(float64(comp.MotionBlur.AdaptiveSampleLimit)),
				SamplesPerFrame:     floatPtr(float64(comp.MotionBlur.SamplesPerFrame)),
			},
			WorkArea: &recipe.CompWorkAreaSpec{
				Start: floatPtr(finiteFloat(comp.WorkArea.Start, 0)),
				End:   floatPtr(finiteFloat(comp.WorkArea.End, duration)),
			},
		}
		if comp.Label != 0 {
			compSpec.Label = floatPtr(float64(comp.Label))
		}
		for _, layer := range comp.Layers {
			layerCount++
			layerSpec := recipe.Layer{
				Type: shellRecipeLayerType(layer, comp.ID),
				Name: shellRecipeLayerName(layer),
			}
			if layerSpec.Type == "precomp" && layer.SourceRef != nil {
				layerSpec.Source = sampleShellSourceName(compNamesByID, layer.SourceRef)
			}
			if layerSpec.Type == "shape" {
				shapeLayerCount++
				layerSpec.Shape = shellShapeSpec(layer)
			}
			layerSpec.Effects = shellRecipeEffects(layer, layerSpec.Type)
			compSpec.Layers = append(compSpec.Layers, layerSpec)
		}
		rec.Comps = append(rec.Comps, compSpec)
	}
	rec.ExpectedProfile = recipe.ExpectedProfile{
		CompCount:       intPtr(len(prof.Comps)),
		LayerCount:      intPtr(layerCount),
		ShapeLayerCount: intPtr(shapeLayerCount),
	}
	return rec
}

func sampleShellCompNames(prof *profile.Profile) map[string]string {
	names := make(map[string]string, len(prof.Comps))
	used := map[string]int{}
	for i, comp := range prof.Comps {
		base := comp.Name
		if base == "" {
			if comp.ID != 0 {
				base = fmt.Sprintf("comp_%d", comp.ID)
			} else {
				base = fmt.Sprintf("comp_index_%d", i)
			}
		}
		used[base]++
		name := base
		if used[base] > 1 {
			name = fmt.Sprintf("%s__%d", base, used[base])
		}
		names[sampleShellCompKey(i, comp)] = name
		if comp.ID != 0 {
			names[sampleShellCompIDKey(comp.ID)] = name
		}
	}
	return names
}

func sampleShellCompKey(index int, comp profile.Composition) string {
	if comp.ID != 0 {
		return sampleShellCompIDKey(comp.ID)
	}
	return fmt.Sprintf("index:%d", index)
}

func sampleShellCompIDKey(id uint32) string {
	return fmt.Sprintf("id:%d", id)
}

func sampleShellSourceName(names map[string]string, ref *profile.ItemRef) string {
	if ref == nil {
		return ""
	}
	if ref.ID != 0 {
		if name := names[sampleShellCompIDKey(ref.ID)]; name != "" {
			return name
		}
	}
	return ref.Name
}

func BuildSampleShellSemanticDiff(original, generated *profile.Profile, samplePath, generatedPath string, rawDiffCount int) SampleShellSemanticDiff {
	originalLayers := allLayers(original)
	generatedLayers := allLayers(generated)
	originalEffects := allEffects(original)
	generatedEffects := allEffects(generated)
	originalShapes := allShapes(original)
	generatedShapes := allShapes(generated)
	alignments := compLayerAlignments(original, generated)
	matchedComps := 0
	for _, row := range alignments {
		if row.Matched {
			matchedComps++
		}
	}
	return SampleShellSemanticDiff{
		SchemaVersion: 1,
		Sample:        samplePath,
		GeneratedAEP:  generatedPath,
		Level:         "semantic-shell-poc",
		Summary: SampleShellSummary{
			CompCount: CountPair{
				Original:      len(original.Comps),
				Generated:     len(generated.Comps),
				MatchedByName: matchedComps,
			},
			LayerCount:      CountPair{Original: len(originalLayers), Generated: len(generatedLayers)},
			ShapeLayerCount: CountPair{Original: countShapeLayers(originalLayers), Generated: countShapeLayers(generatedLayers)},
			EffectCount:     CountPair{Original: len(originalEffects), Generated: len(generatedEffects)},
			FootageCount:    CountPair{Original: original.Fingerprint.FootageCount, Generated: generated.Fingerprint.FootageCount},
			RawProfileDiffCount: RawDiffSummary{
				Shell: rawDiffCount,
				Note:  "raw diff is ID/default-property sensitive; semantic counts are more useful for generated shells",
			},
		},
		CompLayerAlignment:   alignments,
		OriginalLayerTypes:   countStrings(layerTypes(originalLayers)),
		GeneratedLayerTypes:  countStrings(layerTypes(generatedLayers)),
		OriginalShapeKinds:   countStrings(shapeKinds(originalShapes)),
		GeneratedShapeKinds:  countStrings(shapeKinds(generatedShapes)),
		MissingEffectNames:   missingEffectNames(originalEffects, generatedEffects),
		MaterializedBoundary: []string{"comp shell count and names", "layer shell count and broad types", "basic shape layer placeholders", "solid/precomp/null layer placeholders"},
		RemainingGaps: []string{
			"internal IDs are regenerated, so raw profilediff reports missing/extra paths",
			"layer timing/label/flags/transform are intentionally not materialized in this shell stage",
			"shape path geometry/keyframes are not translated into recipe fields",
			"effects are detected but not materialized in this shell stage",
			"footage/null backing items are approximated, not cloned 1:1",
		},
	}
}

func openProfile(path string) (*profile.Profile, error) {
	project, err := aep.Open(path)
	if err != nil {
		return nil, err
	}
	return profile.Build(project, profile.Options{Path: path})
}

func projectNameFromProfile(prof *profile.Profile) string {
	if prof == nil || prof.Meta.Path == "" {
		return "sample-shell"
	}
	base := filepath.Base(prof.Meta.Path)
	ext := filepath.Ext(base)
	return base[:len(base)-len(ext)]
}

func shellRecipeLayerType(layer profile.Layer, ownerCompID uint32) string {
	switch layer.Type {
	case "shape", "null", "text", "camera", "light", "adjustment":
		return layer.Type
	case "av":
		if layer.SourceRef != nil && layer.SourceRef.Kind == "composition" {
			if ownerCompID != 0 && layer.SourceRef.ID == ownerCompID {
				return "solid"
			}
			return "precomp"
		}
		return "solid"
	default:
		return "solid"
	}
}

func shellRecipeLayerName(layer profile.Layer) string {
	if layer.Name != "" {
		return layer.Name
	}
	if layer.SourceRef != nil && layer.SourceRef.Name != "" {
		return layer.SourceRef.Name
	}
	if layer.ID != 0 {
		return fmt.Sprintf("layer_%d", layer.ID)
	}
	return fmt.Sprintf("layer_index_%d", layer.Index)
}

func shellRecipeEffects(layer profile.Layer, recipeLayerType string) []recipe.Effect {
	if recipeLayerType == "camera" || recipeLayerType == "light" {
		return nil
	}
	out := make([]recipe.Effect, 0, len(layer.Effects))
	for _, effect := range layer.Effects {
		if !sampleShellEffectSupported(effect.MatchName) {
			continue
		}
		out = append(out, recipe.Effect{MatchName: sampleShellEffectRecipeMatchName(effect.MatchName), Params: shellRecipeEffectParams(effect)})
	}
	return out
}

func shellRecipeEffectParams(effect profile.Effect) []recipe.EffectParam {
	out := make([]recipe.EffectParam, 0, len(effect.Params))
	for _, param := range effect.Params {
		value, ok := sampleShellEffectParamValue(param)
		if !ok {
			continue
		}
		out = append(out, recipe.EffectParam{MatchName: param.MatchName, Value: value})
	}
	return out
}

func sampleShellEffectParamValue(param profile.Property) (any, bool) {
	if !sampleShellEffectParamSupported(param.MatchName) {
		return nil, false
	}
	if param.LayerRef != nil || len(param.Keyframes) != 0 || param.Expression != "" || param.ExpressionEnabled != nil {
		return nil, false
	}
	return sampleShellStaticParamValue(param.StaticValue)
}

func sampleShellEffectParamSupported(matchName string) bool {
	for _, supported := range aep.SupportedEffectParams() {
		if supported == matchName {
			return true
		}
	}
	return false
}

func sampleShellStaticParamValue(value any) (any, bool) {
	switch v := value.(type) {
	case nil:
		return nil, false
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, false
		}
		return v, true
	case float32:
		n := float64(v)
		if math.IsNaN(n) || math.IsInf(n, 0) {
			return nil, false
		}
		return n, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case bool:
		if v {
			return 1.0, true
		}
		return 0.0, true
	case []float64:
		out := make([]float64, 0, len(v))
		for _, n := range v {
			if math.IsNaN(n) || math.IsInf(n, 0) {
				return nil, false
			}
			out = append(out, n)
		}
		return out, true
	case []any:
		out := make([]float64, 0, len(v))
		for _, item := range v {
			n, ok := sampleShellNumber(item)
			if !ok {
				return nil, false
			}
			out = append(out, n)
		}
		return out, true
	default:
		return nil, false
	}
}

func sampleShellNumber(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return 0, false
		}
		return v, true
	case float32:
		n := float64(v)
		if math.IsNaN(n) || math.IsInf(n, 0) {
			return 0, false
		}
		return n, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

func sampleShellEffectSupported(matchName string) bool {
	recipeMatchName := sampleShellEffectRecipeMatchName(matchName)
	if _, ok := sampleShellEffectAllowlist()[recipeMatchName]; !ok {
		return false
	}
	for _, supported := range aep.SupportedEffects() {
		if supported == recipeMatchName {
			return true
		}
	}
	return false
}

func sampleShellEffectRecipeMatchName(matchName string) string {
	if alias, ok := sampleShellEffectAliases()[matchName]; ok {
		return alias
	}
	return matchName
}

func sampleShellEffectAliases() map[string]string {
	return map[string]string{
		"ADBE Box Blur2":     "ADBE Box Blur",
		"ADBE Gaussian Blur": "ADBE Gaussian Blur 2",
		"ADBE Noise2":        "ADBE Noise",
		"ADBE WRPMESH":       "ADBE MESH WARP",
	}
}

func ClassifySampleShellUnsupportedEffect(matchName string) string {
	if sampleShellEffectSupported(matchName) {
		return "supported_native_pending"
	}
	if strings.HasPrefix(matchName, "Pseudo/") {
		return "pseudo"
	}
	if sampleShellThirdPartyEffect(matchName) {
		return "third_party"
	}
	if sampleShellNativeEffectName(matchName) {
		return "native_template_gap"
	}
	return "unknown_or_alias"
}

func sampleShellEffectWorkItems(missingEffects map[string]int) []SampleShellEffectWorkItem {
	items := make([]SampleShellEffectWorkItem, 0, len(missingEffects))
	for matchName, count := range missingEffects {
		class := ClassifySampleShellUnsupportedEffect(matchName)
		action := sampleShellEffectAction(class)
		if action == "" {
			continue
		}
		items = append(items, SampleShellEffectWorkItem{
			MatchName: matchName,
			Count:     count,
			Class:     class,
			Action:    action,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count != items[j].Count {
			return items[i].Count > items[j].Count
		}
		return items[i].MatchName < items[j].MatchName
	})
	return items
}

func sampleShellEffectAction(class string) string {
	switch class {
	case "supported_native_pending":
		return "enable_sample_shell_copy"
	case "native_template_gap":
		return "add_effect_template"
	case "unknown_or_alias":
		return "classify_or_alias"
	default:
		return ""
	}
}

func sampleShellEffectTemplateRequests(items []SampleShellEffectWorkItem) []string {
	seen := map[string]struct{}{}
	var requested []string
	for _, item := range items {
		if item.Action != "add_effect_template" || item.MatchName == "" {
			continue
		}
		if _, ok := seen[item.MatchName]; ok {
			continue
		}
		seen[item.MatchName] = struct{}{}
		requested = append(requested, item.MatchName)
	}
	sort.Strings(requested)
	return requested
}

func sampleShellReadRIFX(path string) (*rifx.Chunk, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return rifx.Parse(f)
}

func sampleShellFindEffectTemplateCandidates(root *rifx.Chunk, wanted map[string]struct{}) map[string]*rifx.Chunk {
	hits := map[string]*rifx.Chunk{}
	var walk func(*rifx.Chunk)
	walk = func(chunk *rifx.Chunk) {
		if chunk == nil || len(hits) == len(wanted) {
			return
		}
		if chunk.IsList() {
			sampleShellScanEffectParadeReferences(chunk, wanted, hits)
			for _, child := range chunk.Children {
				walk(child)
			}
		}
	}
	walk(root)
	return hits
}

func sampleShellScanEffectParadeReferences(chunk *rifx.Chunk, wanted map[string]struct{}, hits map[string]*rifx.Chunk) {
	children := chunk.Children
	for i := 0; i < len(children)-1; i++ {
		if children[i].ID != rifx.IDTdmn || sampleShellTrimNUL(children[i].Data) != "ADBE Effect Parade" {
			continue
		}
		parade := children[i+1]
		if !parade.IsList() || parade.FormType != rifx.IDTdgp {
			continue
		}
		sampleShellScanEffectEntries(parade, wanted, hits)
	}
}

func sampleShellScanEffectEntries(parade *rifx.Chunk, wanted map[string]struct{}, hits map[string]*rifx.Chunk) {
	children := parade.Children
	for i := 0; i < len(children)-1; i++ {
		nameChunk := children[i]
		if nameChunk.ID != rifx.IDTdmn {
			continue
		}
		matchName := sampleShellTrimNUL(nameChunk.Data)
		if matchName == "ADBE Group End" {
			return
		}
		if _, ok := wanted[matchName]; !ok {
			continue
		}
		payload := children[i+1]
		if !payload.IsList() {
			continue
		}
		hits[matchName] = &rifx.Chunk{
			ID:       rifx.IDList,
			FormType: rifx.IDTdgp,
			Children: []*rifx.Chunk{nameChunk, payload},
		}
		i++
	}
}

func sampleShellChunkBytes(chunk *rifx.Chunk) ([]byte, error) {
	var buf bytes.Buffer
	if err := chunk.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func sampleShellEffectTemplateCandidateFileName(matchName string) string {
	var b strings.Builder
	b.WriteString("effect_")
	lastUnderscore := false
	for _, r := range strings.ToLower(matchName) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}
	name := strings.Trim(b.String(), "_")
	if name == "effect" {
		name = "effect_candidate"
	}
	return name + ".bin"
}

func sampleShellTrimNUL(data []byte) string {
	return string(bytes.TrimRight(data, "\x00"))
}

func sampleShellNativeEffectName(matchName string) bool {
	return strings.HasPrefix(matchName, "ADBE ") ||
		strings.HasPrefix(matchName, "CC ") ||
		strings.HasPrefix(matchName, "CINEMA 4D") ||
		strings.HasPrefix(matchName, "APC ") ||
		matchName == "Keylight 906" ||
		matchName == "mochaAECC"
}

func sampleShellThirdPartyEffect(matchName string) bool {
	exact := map[string]struct{}{
		"PEDG":        {},
		"PEDX":        {},
		"PEQCAGL":     {},
		"PECMBS":      {},
		"GUTS SEPRGB": {},
		"PE QCA":      {},
	}
	if _, ok := exact[matchName]; ok {
		return true
	}
	prefixes := []string{
		"SC_",
		"SC ",
		"tc ",
		"GG ",
		"Mettle ",
		"Universe_",
		"VIDEOCOPILOT ",
		"Videocopilot ",
		"VISINF ",
		"RS ",
		"KNSW ",
		"PixelSort",
		"Chromatic Aberration by ",
		"PEDG_",
		"Particular",
		"Form",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(matchName, prefix) {
			return true
		}
	}
	return false
}

func sampleShellEffectAllowlist() map[string]struct{} {
	return map[string]struct{}{
		"ADBE Fill":                    {},
		"ADBE Slider Control":          {},
		"ADBE Mosaic":                  {},
		"ADBE Easy Levels2":            {},
		"ADBE Turbulent Displace":      {},
		"ADBE Fractal Noise":           {},
		"ADBE Geometry2":               {},
		"ADBE Wave Warp":               {},
		"ADBE Layer Control":           {},
		"ADBE Glo2":                    {},
		"ADBE Posterize Time":          {},
		"ADBE Tint":                    {},
		"ADBE Color Control":           {},
		"ADBE Displacement Map":        {},
		"ADBE Venetian Blinds":         {},
		"ADBE Roughen Edges":           {},
		"ADBE Echo":                    {},
		"CC Composite":                 {},
		"ADBE Linear Wipe":             {},
		"ADBE Ramp":                    {},
		"ADBE Optics Compensation":     {},
		"ADBE Box Blur":                {},
		"CC RepeTile":                  {},
		"ADBE Drop Shadow":             {},
		"ADBE CurvesCustom":            {},
		"ADBE Emboss":                  {},
		"ADBE Invert":                  {},
		"ADBE Gaussian Blur 2":         {},
		"ADBE Angle Control":           {},
		"ADBE Simple Choker":           {},
		"ADBE Brightness & Contrast 2": {},
		"ADBE HUE SATURATION":          {},
		"ADBE Minimax":                 {},
		"ADBE Cell Pattern":            {},
		"ADBE Motion Blur":             {},
		"ADBE Find Edges":              {},
		"ADBE Noise":                   {},
		"ADBE Tritone":                 {},
		"ADBE Sharpen":                 {},
		"ADBE Tile":                    {},
		"ADBE Point Control":           {},
		"ADBE Checkbox Control":        {},
		"ADBE 4ColorGradient":          {},
		"ADBE Grid":                    {},
		"ADBE Stroke":                  {},
		"ADBE Polar Coordinates":       {},
		"ADBE Matte Choker":            {},
		"ADBE Radial Wipe":             {},
		"ADBE Lumetri":                 {},
		"CC Cylinder":                  {},
		"CC Sphere":                    {},
		"CC Glass":                     {},
		"ADBE Channel Combiner":        {},
		"CC Jaws":                      {},
		"CC Mr. Mercury":               {},
		"CC Particle World":            {},
		"ADBE Exposure2":               {},
		"ADBE Checkerboard":            {},
		"ADBE Threshold2":              {},
		"ADBE Color Emboss":            {},
		"ADBE Shift Channels":          {},
		"ADBE Solid Composite":         {},
		"ADBE Paint Bucket":            {},
		"ADBE Change To Color":         {},
		"ADBE Extract":                 {},
		"ADBE GROW BOUNDS":             {},
		"ADBE Luma Key":                {},
		"ADBE Time Displacement":       {},
		"CC Ball Action":               {},
		"CC Particle Systems II":       {},
		"ADBE MESH WARP":               {},
		"ADBE Set Matte3":              {},
		"ADBE Radial Blur":             {},
		"ADBE Twirl":                   {},
		"ADBE Posterize":               {},
		"ADBE Black&White":             {},
		"ADBE Channel Blur":            {},
		"ADBE Bulge":                   {},
		"ADBE AudWave":                 {},
		"ADBE Remove Color Matting":    {},
		"ADBE Noise Alpha2":            {},
		"ADBE Noise HLS2":              {},
		"CC Radial Fast Blur":          {},
		"CC Radial Blur":               {},
		"CC Bend It":                   {},
		"CC Flo Motion":                {},
		"CC Lens":                      {},
		"CC Power Pin":                 {},
		"CC Light Sweep":               {},
		"CC Mr. Smoothie":              {},
		"CC Plastic":                   {},
		"CC Threshold RGB":             {},
		"CC Wide Time":                 {},
		"ADBE Compound Blur":           {},
		"CC Vector Blur":               {},
		"ADBE Color Key":               {},
		"CC Scale Wipe":                {},
		"ADBE Set Channels":            {},
		"CS Vignette":                  {},
		"CS BlockLoad":                 {},
		"CS HexTile":                   {},
		"ADBE AIF Perlin Noise 3D":     {},
		"ADBE Camera Lens Blur":        {},
		"ADBE Cartoonify":              {},
		"ADBE Linear Color Key2":       {},
		"ADBE Paint":                   {},
		"ADBE PhotoFilterPS":           {},
		"ADBE PS Arbitrary Map":        {},
		"ADBE Scatter":                 {},
		"APC Colorama":                 {},
		"APC Radio Waves":              {},
		"APC Shatter":                  {},
		"APC Vegas":                    {},
		"CINEMA 4D Effect":             {},
		"Keylight 906":                 {},
		"mochaAECC":                    {},
	}
}

func shellShapeSpec(layer profile.Layer) *recipe.ShapeSpec {
	if len(layer.Shapes) == 0 {
		return &recipe.ShapeSpec{Kind: "rect"}
	}
	kind := layer.Shapes[0].Kind
	switch kind {
	case "rectangle":
		kind = "rect"
	case "path":
		kind = "rect"
	case "rect", "ellipse", "star", "polygon":
	default:
		kind = "ellipse"
	}
	spec := &recipe.ShapeSpec{Kind: kind}
	if fill := staticNumericSlice(layer.Properties, "ADBE Vector Fill Color"); len(fill) >= 3 {
		spec.FillColor = color4(fill)
	}
	strokeColor := staticNumericSlice(layer.Properties, "ADBE Vector Stroke Color")
	strokeWidth, hasStrokeWidth := staticNumeric(layer.Properties, "ADBE Vector Stroke Width")
	if len(strokeColor) >= 3 || hasStrokeWidth {
		spec.Stroke = &recipe.StrokeSpec{}
		if len(strokeColor) >= 3 {
			spec.Stroke.Color = color4(strokeColor)
		}
		if hasStrokeWidth {
			spec.Stroke.Width = floatPtr(strokeWidth)
		}
	}
	return spec
}

func staticNumeric(properties []profile.Property, matchName string) (float64, bool) {
	for _, prop := range properties {
		if prop.MatchName != matchName {
			continue
		}
		switch v := prop.StaticValue.(type) {
		case float64:
			return v, true
		case int:
			return float64(v), true
		}
	}
	return 0, false
}

func staticNumericSlice(properties []profile.Property, matchName string) []float64 {
	for _, prop := range properties {
		if prop.MatchName != matchName {
			continue
		}
		return numericSlice(prop.StaticValue)
	}
	return nil
}

func numericSlice(value any) []float64 {
	switch v := value.(type) {
	case []float64:
		return append([]float64(nil), v...)
	case [2]float64:
		return []float64{v[0], v[1]}
	case [3]float64:
		return []float64{v[0], v[1], v[2]}
	case [4]float64:
		return []float64{v[0], v[1], v[2], v[3]}
	case []any:
		out := make([]float64, 0, len(v))
		for _, item := range v {
			n, ok := item.(float64)
			if !ok {
				return nil
			}
			out = append(out, n)
		}
		return out
	case string:
		fields := strings.Fields(v)
		out := make([]float64, 0, len(fields))
		for _, field := range fields {
			n, err := strconv.ParseFloat(field, 64)
			if err != nil {
				return nil
			}
			out = append(out, n)
		}
		return out
	default:
		return nil
	}
}

func color4(values []float64) []float64 {
	out := []float64{colorChannel(values[0]), colorChannel(values[1]), colorChannel(values[2]), 255}
	if len(values) >= 4 {
		out[3] = colorChannel(values[3])
	}
	return out
}

func compLayerAlignments(original, generated *profile.Profile) []CompLayerAlignment {
	byName := map[string]profile.Composition{}
	for _, comp := range generated.Comps {
		if _, exists := byName[comp.Name]; !exists {
			byName[comp.Name] = comp
		}
	}
	rows := make([]CompLayerAlignment, 0, len(original.Comps))
	for _, oc := range original.Comps {
		gc, ok := byName[oc.Name]
		row := CompLayerAlignment{
			Name:           oc.Name,
			Matched:        ok,
			OriginalLayers: len(oc.Layers),
		}
		if !ok {
			rows = append(rows, row)
			continue
		}
		row.GeneratedLayers = len(gc.Layers)
		n := len(oc.Layers)
		if len(gc.Layers) < n {
			n = len(gc.Layers)
		}
		origLayers := append([]profile.Layer(nil), oc.Layers...)
		genLayers := append([]profile.Layer(nil), gc.Layers...)
		sort.Slice(origLayers, func(i, j int) bool { return origLayers[i].Index < origLayers[j].Index })
		sort.Slice(genLayers, func(i, j int) bool { return genLayers[i].Index < genLayers[j].Index })
		for i := 0; i < n; i++ {
			if semanticLayerType(origLayers[i]) != semanticLayerType(genLayers[i]) {
				row.TypeMismatches++
			}
			if semanticLayerName(origLayers[i]) != semanticLayerName(genLayers[i]) {
				row.NameMismatches++
			}
		}
		rows = append(rows, row)
	}
	return rows
}

func semanticLayerType(layer profile.Layer) string {
	return layer.Type
}

func semanticLayerName(layer profile.Layer) string {
	return layer.Name
}

func allLayers(prof *profile.Profile) []profile.Layer {
	var out []profile.Layer
	for _, comp := range prof.Comps {
		out = append(out, comp.Layers...)
	}
	return out
}

func allEffects(prof *profile.Profile) []profile.Effect {
	var out []profile.Effect
	for _, layer := range allLayers(prof) {
		out = append(out, layer.Effects...)
	}
	return out
}

func allShapes(prof *profile.Profile) []profile.Shape {
	var out []profile.Shape
	for _, layer := range allLayers(prof) {
		out = append(out, layer.Shapes...)
	}
	return out
}

func countShapeLayers(layers []profile.Layer) int {
	var count int
	for _, layer := range layers {
		if layer.Type == "shape" {
			count++
		}
	}
	return count
}

func layerTypes(layers []profile.Layer) []string {
	out := make([]string, 0, len(layers))
	for _, layer := range layers {
		out = append(out, layer.Type)
	}
	return out
}

func shapeKinds(shapes []profile.Shape) []string {
	out := make([]string, 0, len(shapes))
	for _, shape := range shapes {
		out = append(out, shape.Kind)
	}
	return out
}

func effectNames(effects []profile.Effect) []string {
	out := make([]string, 0, len(effects))
	for _, effect := range effects {
		out = append(out, effect.MatchName)
	}
	return out
}

func missingEffectNames(original, generated []profile.Effect) []CountRow {
	counts := map[string]int{}
	for _, effect := range original {
		counts[sampleShellEffectRecipeMatchName(effect.MatchName)]++
	}
	for _, effect := range generated {
		name := sampleShellEffectRecipeMatchName(effect.MatchName)
		if counts[name] > 0 {
			counts[name]--
		}
	}
	for name, count := range counts {
		if count == 0 {
			delete(counts, name)
		}
	}
	return countMapRows(counts)
}

func countSampleShellCopyableEffectParams(prof *profile.Profile) int {
	var count int
	for _, comp := range prof.Comps {
		for _, layer := range comp.Layers {
			if !sampleShellRecipeLayerCanHaveEffects(shellRecipeLayerType(layer, comp.ID)) {
				continue
			}
			for _, effect := range layer.Effects {
				if !sampleShellEffectSupported(effect.MatchName) {
					continue
				}
				for _, param := range effect.Params {
					if _, ok := sampleShellEffectParamValue(param); ok {
						count++
					}
				}
			}
		}
	}
	return count
}

func countRecipeEffectParams(rec recipe.Recipe) int {
	var count int
	for _, comp := range rec.Comps {
		for _, layer := range comp.Layers {
			for _, effect := range layer.Effects {
				count += len(effect.Params)
			}
		}
	}
	return count
}

func sampleShellRecipeLayerCanHaveEffects(recipeLayerType string) bool {
	return recipeLayerType != "camera" && recipeLayerType != "light"
}

func countStrings(values []string) []CountRow {
	counts := map[string]int{}
	for _, value := range values {
		counts[value]++
	}
	return countMapRows(counts)
}

func countMapRows(counts map[string]int) []CountRow {
	rows := make([]CountRow, 0, len(counts))
	for name, count := range counts {
		rows = append(rows, CountRow{Name: name, Count: count})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Count != rows[j].Count {
			return rows[i].Count > rows[j].Count
		}
		return rows[i].Name < rows[j].Name
	})
	return rows
}

func writeSampleShellJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		data, err = json.MarshalIndent(sampleShellJSONSafeValue(value), "", "  ")
	}
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func sampleShellJSONSafeValue(value any) any {
	return sampleShellJSONSafeReflect(reflect.ValueOf(value))
}

func sampleShellJSONSafeReflect(value reflect.Value) any {
	if !value.IsValid() {
		return nil
	}
	switch value.Kind() {
	case reflect.Interface, reflect.Pointer:
		if value.IsNil() {
			return nil
		}
		return sampleShellJSONSafeReflect(value.Elem())
	case reflect.Float32, reflect.Float64:
		n := value.Float()
		if math.IsNaN(n) || math.IsInf(n, 0) {
			return float64(0)
		}
		return n
	case reflect.Bool:
		return value.Bool()
	case reflect.String:
		return value.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint()
	case reflect.Slice, reflect.Array:
		if value.Kind() == reflect.Slice && value.IsNil() {
			return nil
		}
		out := make([]any, 0, value.Len())
		for i := 0; i < value.Len(); i++ {
			out = append(out, sampleShellJSONSafeReflect(value.Index(i)))
		}
		return out
	case reflect.Map:
		if value.IsNil() {
			return nil
		}
		out := make(map[string]any, value.Len())
		iter := value.MapRange()
		for iter.Next() {
			out[fmt.Sprint(iter.Key().Interface())] = sampleShellJSONSafeReflect(iter.Value())
		}
		return out
	case reflect.Struct:
		out := map[string]any{}
		t := value.Type()
		for i := 0; i < value.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath != "" {
				continue
			}
			name := sampleShellJSONFieldName(field)
			if name == "" {
				continue
			}
			out[name] = sampleShellJSONSafeReflect(value.Field(i))
		}
		return out
	default:
		if value.CanInterface() {
			return value.Interface()
		}
		return nil
	}
}

func sampleShellJSONFieldName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "-" {
		return ""
	}
	if tag != "" {
		name := strings.Split(tag, ",")[0]
		if name != "" {
			return name
		}
	}
	return field.Name
}

func addCountPair(total *CountPair, value CountPair) {
	total.Original += value.Original
	total.Generated += value.Generated
	total.MatchedByName += value.MatchedByName
}

func sampleShellBatchDirName(index int, input string) string {
	base := filepath.Base(input)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]
	return fmt.Sprintf("%03d_%s", index, sampleShellSlug(name))
}

func sampleShellSlug(value string) string {
	value = strings.ToLower(value)
	var b strings.Builder
	lastUnderscore := false
	for _, r := range value {
		ok := r >= 'a' && r <= 'z' || r >= '0' && r <= '9'
		if ok {
			b.WriteRune(r)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			b.WriteByte('_')
			lastUnderscore = true
		}
	}
	out := strings.Trim(b.String(), "_")
	if out == "" {
		return "sample"
	}
	return out
}

func nonZeroFloat(value, fallback float64) float64 {
	if value == 0 {
		return fallback
	}
	return value
}

func finiteFloat(value, fallback float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) || value == 0 {
		return fallback
	}
	return value
}

func colorChannel(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		if math.IsInf(value, 1) {
			return 255
		}
		return 0
	}
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}
	return value
}

func intPtr(v int) *int { return &v }

func floatPtr(v float64) *float64 { return &v }

func boolPtr(v bool) *bool { return &v }
