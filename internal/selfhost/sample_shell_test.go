package selfhost

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/profile"
	"github.com/yueli-fx/aep-parser/internal/rifx"
)

func TestDiscoverSampleShellInputsReturnsSortedLimitedAEPs(t *testing.T) {
	root := t.TempDir()
	mustWriteFile(t, filepath.Join(root, "b", "two.aep"))
	mustWriteFile(t, filepath.Join(root, "a", "one.aep"))
	mustWriteFile(t, filepath.Join(root, "a", "ignore.txt"))
	mustWriteFile(t, filepath.Join(root, "c", "three.AEP"))

	inputs, err := DiscoverSampleShellInputs(root, 2)
	if err != nil {
		t.Fatalf("DiscoverSampleShellInputs returned error: %v", err)
	}

	want := []string{
		filepath.Join(root, "a", "one.aep"),
		filepath.Join(root, "b", "two.aep"),
	}
	if len(inputs) != len(want) {
		t.Fatalf("inputs = %v, want %v", inputs, want)
	}
	for i := range want {
		if inputs[i] != want[i] {
			t.Fatalf("inputs[%d] = %q, want %q", i, inputs[i], want[i])
		}
	}
}

func TestRunSampleShellEffectTemplateExtractionWritesCandidates(t *testing.T) {
	root := t.TempDir()
	samplePath := filepath.Join(root, "samples", "one.aep")
	writeSampleShellTestAEP(t, samplePath, "ADBE Missing FX")
	summaryPath := filepath.Join(root, "batch_summary.json")
	mustWriteFileText(t, summaryPath, `{
  "schema_version": 1,
  "root": "ignored",
  "effect_work_items": [
    {"match_name":"ADBE Missing FX","count":2,"class":"native_template_gap","action":"add_effect_template"},
    {"match_name":"ADBE Absent FX","count":1,"class":"native_template_gap","action":"add_effect_template"},
    {"match_name":"tc Particular","count":9,"class":"third_party","action":"ignore"}
  ]
}`)
	outDir := filepath.Join(root, "out")

	result, err := RunSampleShellEffectTemplateExtraction(SampleShellEffectTemplateExtractionOptions{
		SummaryPath: summaryPath,
		Root:        filepath.Join(root, "samples"),
		OutDir:      outDir,
	})
	if err != nil {
		t.Fatalf("RunSampleShellEffectTemplateExtraction returned error: %v", err)
	}

	if len(result.Requested) != 2 || result.Requested[0] != "ADBE Absent FX" || result.Requested[1] != "ADBE Missing FX" {
		t.Fatalf("requested = %+v", result.Requested)
	}
	if len(result.Hits) != 1 || result.Hits[0].MatchName != "ADBE Missing FX" || result.Hits[0].Bytes == 0 {
		t.Fatalf("hits = %+v", result.Hits)
	}
	if len(result.Missing) != 1 || result.Missing[0] != "ADBE Absent FX" {
		t.Fatalf("missing = %+v", result.Missing)
	}
	if _, err := os.Stat(result.ResultPath); err != nil {
		t.Fatalf("result summary missing: %v", err)
	}
	data, err := os.ReadFile(result.Hits[0].OutPath)
	if err != nil {
		t.Fatalf("ReadFile candidate: %v", err)
	}
	wrapper, err := rifx.ReadChunk(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("ReadChunk candidate: %v", err)
	}
	if len(wrapper.Children) != 2 || wrapper.Children[0].ID != rifx.IDTdmn || sampleShellTrimNUL(wrapper.Children[0].Data) != "ADBE Missing FX" {
		t.Fatalf("candidate wrapper = %+v", wrapper.Children)
	}
}

func TestBuildSampleShellBatchSummaryAggregatesRunsAndFailures(t *testing.T) {
	results := []SampleShellResult{{
		InputPath:        filepath.Join("samples", "one.aep"),
		OutDir:           filepath.Join("out", "001_one"),
		SemanticDiffPath: filepath.Join("out", "001_one", "semantic_diff.json"),
		SemanticDiff: SampleShellSemanticDiff{
			Summary: SampleShellSummary{
				CompCount:           CountPair{Original: 2, Generated: 2},
				LayerCount:          CountPair{Original: 4, Generated: 4},
				ShapeLayerCount:     CountPair{Original: 3, Generated: 3},
				EffectCount:         CountPair{Original: 1, Generated: 0},
				FootageCount:        CountPair{Original: 2, Generated: 1},
				RawProfileDiffCount: RawDiffSummary{Shell: 50},
			},
			MissingEffectNames: []CountRow{{Name: "ADBE Box Blur2", Count: 1}},
		},
	}, {
		InputPath:        filepath.Join("samples", "two.aep"),
		OutDir:           filepath.Join("out", "002_two"),
		SemanticDiffPath: filepath.Join("out", "002_two", "semantic_diff.json"),
		SemanticDiff: SampleShellSemanticDiff{
			Summary: SampleShellSummary{
				CompCount:           CountPair{Original: 1, Generated: 1},
				LayerCount:          CountPair{Original: 2, Generated: 1},
				ShapeLayerCount:     CountPair{Original: 1, Generated: 1},
				EffectCount:         CountPair{Original: 2, Generated: 0},
				FootageCount:        CountPair{Original: 0, Generated: 0},
				RawProfileDiffCount: RawDiffSummary{Shell: 10},
			},
			MissingEffectNames: []CountRow{{Name: "Pseudo/YanKFB2", Count: 1}, {Name: "tc Particular", Count: 1}},
		},
	}}
	failures := []SampleShellBatchFailure{{
		InputPath: filepath.Join("samples", "bad.aep"),
		OutDir:    filepath.Join("out", "003_bad"),
		Error:     "open failed",
	}}

	summary := BuildSampleShellBatchSummary("samples", "out", "AE2020", results, failures)

	if summary.Summary.Total != 3 || summary.Summary.Succeeded != 2 || summary.Summary.Failed != 1 {
		t.Fatalf("summary counts = %+v", summary.Summary)
	}
	if summary.Summary.CompCount.Original != 3 || summary.Summary.CompCount.Generated != 3 {
		t.Fatalf("comp aggregate = %+v", summary.Summary.CompCount)
	}
	if summary.Summary.LayerCount.Original != 6 || summary.Summary.LayerCount.Generated != 5 {
		t.Fatalf("layer aggregate = %+v", summary.Summary.LayerCount)
	}
	if summary.Summary.RawProfileDiffCount != 60 {
		t.Fatalf("raw diff aggregate = %d, want 60", summary.Summary.RawProfileDiffCount)
	}
	if len(summary.Runs) != 2 || len(summary.Failures) != 1 {
		t.Fatalf("runs/failures = %d/%d, want 2/1", len(summary.Runs), len(summary.Failures))
	}
	if len(summary.MissingEffectNames) != 3 ||
		summary.MissingEffectNames[0] != (CountRow{Name: "ADBE Box Blur2", Count: 1}) ||
		summary.MissingEffectNames[1] != (CountRow{Name: "Pseudo/YanKFB2", Count: 1}) ||
		summary.MissingEffectNames[2] != (CountRow{Name: "tc Particular", Count: 1}) {
		t.Fatalf("missing effect aggregate = %+v", summary.MissingEffectNames)
	}
	if len(summary.UnsupportedEffectClasses) != 3 ||
		summary.UnsupportedEffectClasses[0] != (CountRow{Name: "pseudo", Count: 1}) ||
		summary.UnsupportedEffectClasses[1] != (CountRow{Name: "supported_native_pending", Count: 1}) ||
		summary.UnsupportedEffectClasses[2] != (CountRow{Name: "third_party", Count: 1}) {
		t.Fatalf("unsupported effect classes = %+v", summary.UnsupportedEffectClasses)
	}
	if len(summary.EffectWorkItems) != 1 ||
		summary.EffectWorkItems[0] != (SampleShellEffectWorkItem{
			MatchName: "ADBE Box Blur2",
			Count:     1,
			Class:     "supported_native_pending",
			Action:    "enable_sample_shell_copy",
		}) {
		t.Fatalf("effect work items = %+v", summary.EffectWorkItems)
	}
}

func TestClassifySampleShellUnsupportedEffect(t *testing.T) {
	cases := map[string]string{
		"ADBE Box Blur2":                  "supported_native_pending",
		"ADBE Fill":                       "supported_native_pending",
		"Pseudo/YanKFB2":                  "pseudo",
		"SC_Strdst_cntrls_particle_1":     "third_party",
		"tc Particular":                   "third_party",
		"GG AEPixelSorter2":               "third_party",
		"Mettle SkyBox Chromatic Aberrat": "third_party",
		"VISINF Grain Implant":            "third_party",
		"VIDEOCOPILOT VIBRANCE":           "third_party",
		"Universe_Glow_Glow":              "third_party",
		"RS Motion Blur Pro A 3.x":        "third_party",
		"KNSW Unmult":                     "third_party",
		"PixelSort":                       "third_party",
		"CINEMA 4D Effect":                "supported_native_pending",
		"APC Colorama":                    "supported_native_pending",
		"mochaAECC":                       "supported_native_pending",
		"Keylight 906":                    "supported_native_pending",
		"PEDG":                            "third_party",
		"PEDX":                            "third_party",
		"PEQCAGL":                         "third_party",
		"PECMBS":                          "third_party",
		"GUTS SEPRGB":                     "third_party",
		"PE QCA":                          "third_party",
	}
	for matchName, want := range cases {
		if got := ClassifySampleShellUnsupportedEffect(matchName); got != want {
			t.Fatalf("ClassifySampleShellUnsupportedEffect(%q) = %q, want %q", matchName, got, want)
		}
	}
}

func TestSampleShellEffectSupportedIncludesNativeTemplateGapAlreadySupported(t *testing.T) {
	for _, matchName := range []string{
		"ADBE Tritone", "CC Cylinder", "ADBE Exposure2", "CC Particle Systems II",
		"ADBE Box Blur2", "ADBE Set Matte3", "CC Vector Blur", "CS Vignette",
		"ADBE AIF Perlin Noise 3D", "ADBE Camera Lens Blur", "ADBE Cartoonify",
		"ADBE Linear Color Key2", "ADBE Paint", "ADBE PhotoFilterPS",
		"ADBE PS Arbitrary Map", "ADBE Scatter", "APC Colorama",
		"APC Radio Waves", "APC Shatter", "APC Vegas", "CINEMA 4D Effect",
		"Keylight 906", "mochaAECC",
	} {
		if !sampleShellEffectSupported(matchName) {
			t.Fatalf("sampleShellEffectSupported(%q) = false, want true", matchName)
		}
	}
}

func TestSampleShellEffectAliasesUseCanonicalRecipeMatchNames(t *testing.T) {
	prof := sampleShellProfile()
	prof.Comps[0].Layers[0].Effects = []profile.Effect{{MatchName: "ADBE Box Blur2"}}

	rec := BuildSampleShellRecipe(prof, "AE2020")

	effects := rec.Comps[0].Layers[0].Effects
	if len(effects) != 1 || effects[0].MatchName != "ADBE Box Blur" {
		t.Fatalf("effects = %+v, want canonical ADBE Box Blur", effects)
	}
}

func TestBuildSampleShellSemanticDiffTreatsEffectAliasesAsCovered(t *testing.T) {
	original := sampleShellProfile()
	generated := sampleShellProfile()
	original.Comps[0].Layers[0].Effects = []profile.Effect{{MatchName: "ADBE Box Blur2"}, {MatchName: "ADBE Gaussian Blur"}}
	generated.Comps[0].Layers[0].Effects = []profile.Effect{{MatchName: "ADBE Box Blur"}, {MatchName: "ADBE Gaussian Blur 2"}}

	diff := BuildSampleShellSemanticDiff(original, generated, "source.aep", "shell.aep", 0)

	if len(diff.MissingEffectNames) != 0 {
		t.Fatalf("missing effects = %+v, want aliases covered", diff.MissingEffectNames)
	}
}

func TestBuildSampleShellRecipeCopiesCompAndLayerShell(t *testing.T) {
	prof := sampleShellProfile()

	rec := BuildSampleShellRecipe(prof, "AE2025")

	if rec.Project.TargetVersion != "AE2025" {
		t.Fatalf("target version = %q, want AE2025", rec.Project.TargetVersion)
	}
	if len(rec.Comps) != 2 {
		t.Fatalf("comps = %d, want 2", len(rec.Comps))
	}
	if rec.Comps[0].Name != "Source" || rec.Comps[0].Width != 640 || rec.Comps[0].Height != 360 {
		t.Fatalf("source comp = %+v", rec.Comps[0])
	}
	if len(rec.Comps[0].Layers) != 1 || rec.Comps[0].Layers[0].Type != "shape" {
		t.Fatalf("source layers = %+v", rec.Comps[0].Layers)
	}
	if rec.Comps[0].Layers[0].Shape == nil || rec.Comps[0].Layers[0].Shape.Kind != "rect" {
		t.Fatalf("shape = %+v, want path downgraded to rect placeholder", rec.Comps[0].Layers[0].Shape)
	}
	if got := rec.Comps[0].Layers[0].Shape.FillColor; len(got) != 4 || got[0] != 255 || got[3] != 128 {
		t.Fatalf("fill color = %v, want RGBA copied", got)
	}
	if len(rec.Comps[1].Layers) != 1 || rec.Comps[1].Layers[0].Type != "precomp" || rec.Comps[1].Layers[0].Source != "Source" {
		t.Fatalf("main layers = %+v, want precomp source", rec.Comps[1].Layers)
	}
	if rec.ExpectedProfile.CompCount == nil || *rec.ExpectedProfile.CompCount != 2 {
		t.Fatalf("expected comp count = %+v, want 2", rec.ExpectedProfile.CompCount)
	}
	if rec.ExpectedProfile.LayerCount == nil || *rec.ExpectedProfile.LayerCount != 2 {
		t.Fatalf("expected layer count = %+v, want 2", rec.ExpectedProfile.LayerCount)
	}
	if rec.ExpectedProfile.ShapeLayerCount == nil || *rec.ExpectedProfile.ShapeLayerCount != 1 {
		t.Fatalf("expected shape layer count = %+v, want 1", rec.ExpectedProfile.ShapeLayerCount)
	}
}

func TestBuildSampleShellRecipeCopiesSafeEffectShells(t *testing.T) {
	prof := sampleShellProfile()
	prof.Comps[0].Layers[0].Effects = []profile.Effect{
		{MatchName: "ADBE Fill"},
		{
			MatchName: "ADBE Slider Control",
			Params: []profile.Property{{
				MatchName:   "ADBE Slider Control-0001",
				StaticValue: 42.5,
			}, {
				MatchName:   "ADBE Slider Control-9999",
				StaticValue: 1.0,
			}, {
				MatchName:   "ADBE Slider Control-0001",
				StaticValue: 13.0,
				Expression:  "time",
			}},
		},
		{MatchName: "ADBE Drop Shadow"},
		{MatchName: "Pseudo/Unsupported"},
	}
	prof.Comps[1].Layers = append(prof.Comps[1].Layers, profile.Layer{
		ID:      12,
		Index:   1,
		Name:    "Camera",
		Type:    "camera",
		Effects: []profile.Effect{{MatchName: "ADBE Fill"}},
	})

	rec := BuildSampleShellRecipe(prof, "AE2020")

	effects := rec.Comps[0].Layers[0].Effects
	if len(effects) != 3 {
		t.Fatalf("effects = %+v, want 3 safe copied effects", effects)
	}
	if effects[0].MatchName != "ADBE Fill" || effects[1].MatchName != "ADBE Slider Control" || effects[2].MatchName != "ADBE Drop Shadow" {
		t.Fatalf("effects = %+v, want Fill, Slider Control, Drop Shadow", effects)
	}
	if len(effects[1].Params) != 1 ||
		effects[1].Params[0].MatchName != "ADBE Slider Control-0001" ||
		effects[1].Params[0].Value != 42.5 {
		t.Fatalf("slider params = %+v, want one copied static param", effects[1].Params)
	}
	if got := rec.Comps[1].Layers[1].Effects; len(got) != 0 {
		t.Fatalf("camera effects = %+v, want skipped", got)
	}
}

func TestBuildSampleShellSemanticDiffReportsMaterializedAndMissingDomains(t *testing.T) {
	original := sampleShellProfile()
	generated := sampleShellProfile()
	generated.Fingerprint.FootageCount = 0
	original.Comps[0].Layers[0].Effects = []profile.Effect{{MatchName: "ADBE Fill"}, {MatchName: "ADBE Fill"}, {MatchName: "ADBE Slider Control"}}
	generated.Comps[0].Layers[0].Effects = []profile.Effect{{MatchName: "ADBE Fill"}}

	diff := BuildSampleShellSemanticDiff(original, generated, "source.aep", "shell.aep", 42)

	if diff.Summary.CompCount.Original != 2 || diff.Summary.CompCount.Generated != 2 || diff.Summary.CompCount.MatchedByName != 2 {
		t.Fatalf("comp summary = %+v", diff.Summary.CompCount)
	}
	if diff.Summary.LayerCount.Original != 2 || diff.Summary.LayerCount.Generated != 2 {
		t.Fatalf("layer summary = %+v", diff.Summary.LayerCount)
	}
	if diff.Summary.EffectCount.Original != 3 || diff.Summary.EffectCount.Generated != 1 {
		t.Fatalf("effect summary = %+v", diff.Summary.EffectCount)
	}
	if len(diff.MissingEffectNames) != 2 ||
		diff.MissingEffectNames[0] != (CountRow{Name: "ADBE Fill", Count: 1}) ||
		diff.MissingEffectNames[1] != (CountRow{Name: "ADBE Slider Control", Count: 1}) {
		t.Fatalf("missing effects = %+v", diff.MissingEffectNames)
	}
	if diff.Summary.RawProfileDiffCount.Shell != 42 {
		t.Fatalf("raw diff count = %+v, want 42", diff.Summary.RawProfileDiffCount)
	}
}

func TestBuildSampleShellRecipeSanitizesNamesShapesAndColors(t *testing.T) {
	prof := &profile.Profile{
		SchemaVersion: 1,
		Meta:          profile.Meta{Path: "dupes.aep"},
		Comps: []profile.Composition{{
			ID:          1,
			Name:        "Box",
			Width:       1920,
			Height:      1080,
			FrameRate:   30,
			Duration:    1,
			PixelAspect: 1,
			Layers: []profile.Layer{{
				ID:    10,
				Index: 0,
				Name:  "empty shape",
				Type:  "shape",
			}},
		}, {
			ID:          2,
			Name:        "Box",
			Width:       1920,
			Height:      1080,
			FrameRate:   math.NaN(),
			Duration:    1,
			PixelAspect: math.NaN(),
			Layers: []profile.Layer{{
				ID:        11,
				Index:     0,
				Name:      "self",
				Type:      "av",
				SourceRef: &profile.ItemRef{ID: 2, Kind: "composition", Name: "Box"},
			}, {
				ID:        12,
				Index:     1,
				Name:      "other",
				Type:      "av",
				SourceRef: &profile.ItemRef{ID: 1, Kind: "composition", Name: "Box"},
			}, {
				ID:     13,
				Index:  2,
				Name:   "color",
				Type:   "shape",
				Shapes: []profile.Shape{{Kind: "ellipse"}},
				Properties: []profile.Property{{
					MatchName:   "ADBE Vector Fill Color",
					StaticValue: []float64{-10, 300, math.NaN(), math.Inf(1)},
				}},
			}},
		}},
	}

	rec := BuildSampleShellRecipe(prof, "AE2020")

	if rec.Comps[0].Name != "Box" || rec.Comps[1].Name != "Box__2" {
		t.Fatalf("comp names = %q/%q, want unique names", rec.Comps[0].Name, rec.Comps[1].Name)
	}
	if rec.Comps[0].Layers[0].Shape == nil || rec.Comps[0].Layers[0].Shape.Kind != "rect" {
		t.Fatalf("empty shape layer spec = %+v, want rect placeholder", rec.Comps[0].Layers[0].Shape)
	}
	if rec.Comps[1].FrameRate != 30 || rec.Comps[1].PixelAspect == nil || *rec.Comps[1].PixelAspect != 1 {
		t.Fatalf("sanitized comp timing = frameRate %v pixelAspect %+v", rec.Comps[1].FrameRate, rec.Comps[1].PixelAspect)
	}
	if rec.Comps[1].Layers[0].Type != "solid" || rec.Comps[1].Layers[0].Source != "" {
		t.Fatalf("self precomp layer = %+v, want solid placeholder", rec.Comps[1].Layers[0])
	}
	if rec.Comps[1].Layers[1].Type != "precomp" || rec.Comps[1].Layers[1].Source != "Box" {
		t.Fatalf("other precomp layer = %+v, want source Box", rec.Comps[1].Layers[1])
	}
	fill := rec.Comps[1].Layers[2].Shape.FillColor
	if len(fill) != 4 || fill[0] != 0 || fill[1] != 255 || fill[2] != 0 || fill[3] != 255 {
		t.Fatalf("sanitized fill = %v, want clamped finite RGBA", fill)
	}
}

func TestSampleShellJSONSafeValueReplacesNonFiniteFloats(t *testing.T) {
	input := map[string]any{
		"nan": math.NaN(),
		"pos": math.Inf(1),
		"neg": math.Inf(-1),
		"ok":  12.5,
		"nested": []any{
			math.NaN(),
			map[string]any{"value": math.Inf(1)},
		},
	}

	got := sampleShellJSONSafeValue(input).(map[string]any)

	if got["nan"] != float64(0) || got["pos"] != float64(0) || got["neg"] != float64(0) || got["ok"] != 12.5 {
		t.Fatalf("safe scalar values = %+v", got)
	}
	nested := got["nested"].([]any)
	if nested[0] != float64(0) || nested[1].(map[string]any)["value"] != float64(0) {
		t.Fatalf("safe nested values = %+v", nested)
	}
}

func sampleShellProfile() *profile.Profile {
	return &profile.Profile{
		SchemaVersion: 1,
		Meta:          profile.Meta{Path: "demo.aep"},
		Fingerprint:   profile.Fingerprint{CompCount: 2, LayerCount: 2, FootageCount: 1},
		Comps: []profile.Composition{{
			ID:               1,
			Name:             "Source",
			Width:            640,
			Height:           360,
			FrameRate:        24,
			Duration:         2,
			BackgroundColor:  [3]uint8{1, 2, 3},
			ResolutionFactor: [2]uint16{1, 1},
			PixelAspect:      1,
			WorkArea:         profile.WorkArea{Start: 0, End: 2},
			MotionBlur:       profile.MotionBlurSettings{ShutterAngle: 180, AdaptiveSampleLimit: 128, SamplesPerFrame: 16},
			Layers: []profile.Layer{{
				ID:    10,
				Index: 0,
				Name:  "Shape",
				Type:  "shape",
				Properties: []profile.Property{{
					MatchName:   "ADBE Vector Fill Color",
					StaticValue: []float64{255, 10, 20, 128},
				}, {
					MatchName:   "ADBE Vector Stroke Width",
					StaticValue: float64(3),
				}},
				Effects: []profile.Effect{{MatchName: "ADBE Fill"}},
				Shapes:  []profile.Shape{{Kind: "path"}},
			}},
		}, {
			ID:               2,
			Name:             "Main",
			Width:            640,
			Height:           360,
			FrameRate:        24,
			Duration:         2,
			BackgroundColor:  [3]uint8{0, 0, 0},
			ResolutionFactor: [2]uint16{1, 1},
			PixelAspect:      1,
			WorkArea:         profile.WorkArea{Start: 0, End: 2},
			MotionBlur:       profile.MotionBlurSettings{ShutterAngle: 180, AdaptiveSampleLimit: 128, SamplesPerFrame: 16},
			Layers: []profile.Layer{{
				ID:        11,
				Index:     0,
				Name:      "",
				Type:      "av",
				SourceRef: &profile.ItemRef{ID: 1, Kind: "composition", Name: "Source"},
			}},
		}},
	}
}

func mustWriteFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func mustWriteFileText(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func writeSampleShellTestAEP(t *testing.T, path, effectMatchName string) {
	t.Helper()
	root := &rifx.Chunk{ID: rifx.IDRifx, FormType: rifx.IDEgg, Children: []*rifx.Chunk{
		{
			ID:       rifx.IDList,
			FormType: rifx.IDTdgp,
			Children: []*rifx.Chunk{
				{ID: rifx.IDTdmn, Data: []byte("ADBE Effect Parade")},
				{
					ID:       rifx.IDList,
					FormType: rifx.IDTdgp,
					Children: []*rifx.Chunk{
						{ID: rifx.IDTdmn, Data: []byte(effectMatchName)},
						{ID: rifx.IDList, FormType: rifx.IDSspc, Children: []*rifx.Chunk{
							{ID: rifx.IDFnam, Data: []byte("candidate")},
						}},
						{ID: rifx.IDTdmn, Data: []byte("ADBE Group End")},
					},
				},
				{ID: rifx.IDTdmn, Data: []byte("ADBE Group End")},
			},
		},
	}}
	var buf bytes.Buffer
	if err := root.Write(&buf); err != nil {
		t.Fatalf("write rifx: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write aep: %v", err)
	}
}
