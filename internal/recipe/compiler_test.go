package recipe_test

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
	"github.com/yueli-fx/aep-parser/internal/recipe"
)

func TestCompileMinimalTextShapeRecipeBuildsProfile(t *testing.T) {
	rec := minimalRecipe()
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, recipe.StaticCapabilities{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid || report.OutputPath != outPath {
		t.Fatalf("report = %+v", report)
	}

	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	if prof.Fingerprint.CompCount != 1 {
		t.Fatalf("CompCount = %d, want 1", prof.Fingerprint.CompCount)
	}
	if prof.Fingerprint.LayerCount != 2 {
		t.Fatalf("LayerCount = %d, want 2", prof.Fingerprint.LayerCount)
	}
	var textLayers, shapeLayers int
	for _, layer := range prof.Comps[0].Layers {
		if layer.Text != nil {
			textLayers++
		}
		if len(layer.Shapes) > 0 {
			shapeLayers++
		}
	}
	if textLayers != 1 || shapeLayers != 1 {
		t.Fatalf("textLayers=%d shapeLayers=%d, want 1/1; layers=%+v", textLayers, shapeLayers, prof.Comps[0].Layers)
	}
}

func TestCompileToFileSupportsMultipleCompsAndPrecompLayer(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps = []recipe.CompSpec{
		{
			Name:      "Source",
			Width:     640,
			Height:    360,
			FrameRate: 24,
			Duration:  2,
			Layers: []recipe.Layer{{
				Type: "text",
				Name: "Source Title",
				Text: "Nested source",
			}},
		},
		{
			Name:      "Main",
			Width:     1280,
			Height:    720,
			FrameRate: 24,
			Duration:  3,
			Layers: []recipe.Layer{{
				Type:   "precomp",
				Name:   "Nested",
				Source: "Source",
			}},
		},
	}
	rec.ExpectedProfile = recipe.ExpectedProfile{
		CompCount:  intPtr(2),
		LayerCount: intPtr(2),
		Comps: []recipe.ExpectedComp{{
			Name:      "Source",
			Width:     ptr(640),
			Height:    ptr(360),
			FrameRate: ptr(24),
			Duration:  ptr(2),
		}, {
			Name:      "Main",
			Width:     ptr(1280),
			Height:    ptr(720),
			FrameRate: ptr(24),
			Duration:  ptr(3),
		}},
		Layers: []recipe.ExpectedLayer{{
			Name:       "Nested",
			Type:       "av",
			Source:     "Source",
			SourceKind: "composition",
		}},
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report.Valid = false, refusals=%+v checks=%+v", report.Refusals, report.ProfileChecks)
	}

	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	if prof.Fingerprint.CompCount != 2 {
		t.Fatalf("CompCount = %d, want 2", prof.Fingerprint.CompCount)
	}
	assertProfileCheck(t, report, "expected_profile.comps[0].width", true)
	assertProfileCheck(t, report, "expected_profile.comps[1].duration", true)
	layer := findProfileLayer(t, prof, "Nested")
	if layer.SourceRef == nil || layer.SourceRef.Name != "Source" || layer.SourceRef.Kind != "composition" {
		t.Fatalf("SourceRef = %+v, want Source composition", layer.SourceRef)
	}
}

func TestRecipeExamplesExpectedProfilesAreNotCountOnly(t *testing.T) {
	recipePaths, err := filepath.Glob(filepath.Join("..", "..", "examples", "recipes", "*.json"))
	if err != nil {
		t.Fatalf("Glob recipe examples: %v", err)
	}
	if len(recipePaths) == 0 {
		t.Fatal("no recipe examples found")
	}
	countKeys := map[string]bool{
		"comp_count":        true,
		"layer_count":       true,
		"text_layer_count":  true,
		"shape_layer_count": true,
	}
	for _, recipePath := range recipePaths {
		t.Run(filepath.Base(recipePath), func(t *testing.T) {
			raw, err := os.ReadFile(recipePath)
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			var doc map[string]any
			if err := json.Unmarshal(raw, &doc); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			expectedProfile, ok := doc["expected_profile"].(map[string]any)
			if !ok {
				t.Fatal("expected_profile is required")
			}
			for key := range expectedProfile {
				if !countKeys[key] {
					return
				}
			}
			t.Fatal("expected_profile must include at least one non-count check")
		})
	}
}

func TestRecipeExamplesWithLayersAssertLayerProfiles(t *testing.T) {
	recipePaths, err := filepath.Glob(filepath.Join("..", "..", "examples", "recipes", "*.json"))
	if err != nil {
		t.Fatalf("Glob recipe examples: %v", err)
	}
	if len(recipePaths) == 0 {
		t.Fatal("no recipe examples found")
	}
	for _, recipePath := range recipePaths {
		t.Run(filepath.Base(recipePath), func(t *testing.T) {
			raw, err := os.ReadFile(recipePath)
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			var doc map[string]any
			if err := json.Unmarshal(raw, &doc); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			authoredLayerCount := recipeExampleAuthoredLayerCount(doc)
			if authoredLayerCount == 0 {
				return
			}
			expectedProfile, ok := doc["expected_profile"].(map[string]any)
			if !ok {
				t.Fatal("expected_profile is required")
			}
			expectedLayers, ok := expectedProfile["layers"].([]any)
			if !ok {
				t.Fatal("expected_profile.layers is required when recipe authors layers")
			}
			if len(expectedLayers) < authoredLayerCount {
				t.Fatalf("expected_profile.layers has %d entries, want at least %d", len(expectedLayers), authoredLayerCount)
			}
		})
	}
}

func TestRecipeExamplesWithTransformKeyframesAssertKeyframeProfiles(t *testing.T) {
	recipePaths, err := filepath.Glob(filepath.Join("..", "..", "examples", "recipes", "*.json"))
	if err != nil {
		t.Fatalf("Glob recipe examples: %v", err)
	}
	if len(recipePaths) == 0 {
		t.Fatal("no recipe examples found")
	}
	for _, recipePath := range recipePaths {
		t.Run(filepath.Base(recipePath), func(t *testing.T) {
			raw, err := os.ReadFile(recipePath)
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			var doc map[string]any
			if err := json.Unmarshal(raw, &doc); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			authoredKeyframeStreams := recipeExampleAuthoredTransformKeyframeStreams(doc)
			if authoredKeyframeStreams == 0 {
				return
			}
			expectedProfile, ok := doc["expected_profile"].(map[string]any)
			if !ok {
				t.Fatal("expected_profile is required")
			}
			expectedKeyframes, ok := expectedProfile["keyframes"].([]any)
			if !ok {
				t.Fatal("expected_profile.keyframes is required when recipe authors transform keyframes")
			}
			if len(expectedKeyframes) < authoredKeyframeStreams {
				t.Fatalf("expected_profile.keyframes has %d entries, want at least %d", len(expectedKeyframes), authoredKeyframeStreams)
			}
		})
	}
}

func TestRecipeExamplesWithNestedContentAssertProfileFamilies(t *testing.T) {
	recipePaths, err := filepath.Glob(filepath.Join("..", "..", "examples", "recipes", "*.json"))
	if err != nil {
		t.Fatalf("Glob recipe examples: %v", err)
	}
	if len(recipePaths) == 0 {
		t.Fatal("no recipe examples found")
	}
	for _, recipePath := range recipePaths {
		t.Run(filepath.Base(recipePath), func(t *testing.T) {
			raw, err := os.ReadFile(recipePath)
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			var doc map[string]any
			if err := json.Unmarshal(raw, &doc); err != nil {
				t.Fatalf("Unmarshal: %v", err)
			}
			requirements := recipeExampleNestedProfileRequirements(doc)
			if len(requirements) == 0 {
				return
			}
			expectedProfile, ok := doc["expected_profile"].(map[string]any)
			if !ok {
				t.Fatal("expected_profile is required")
			}
			for _, key := range requirements {
				if expectedProfileListLen(expectedProfile, key) == 0 {
					t.Fatalf("expected_profile.%s is required for nested recipe content", key)
				}
			}
		})
	}
}

func recipeExampleAuthoredLayerCount(doc map[string]any) int {
	comps, ok := doc["comps"].([]any)
	if !ok {
		return 0
	}
	var count int
	for _, rawComp := range comps {
		comp, ok := rawComp.(map[string]any)
		if !ok {
			continue
		}
		layers, ok := comp["layers"].([]any)
		if ok {
			count += len(layers)
		}
	}
	return count
}

func recipeExampleNestedProfileRequirements(doc map[string]any) []string {
	comps, ok := doc["comps"].([]any)
	if !ok {
		return nil
	}
	required := map[string]bool{}
	for _, rawComp := range comps {
		comp, ok := rawComp.(map[string]any)
		if !ok {
			continue
		}
		layers, ok := comp["layers"].([]any)
		if !ok {
			continue
		}
		for _, rawLayer := range layers {
			layer, ok := rawLayer.(map[string]any)
			if !ok {
				continue
			}
			if textStyle, ok := layer["text_style"].(map[string]any); ok && len(textStyle) > 0 {
				required["text_styles"] = true
			}
			if textAnimators, ok := layer["text_animators"].([]any); ok && len(textAnimators) > 0 {
				for _, rawAnimator := range textAnimators {
					animator, ok := rawAnimator.(map[string]any)
					if !ok {
						required["properties"] = true
						continue
					}
					if _, ok := animator["value_keyframes"]; ok {
						required["keyframes"] = true
					} else {
						required["properties"] = true
					}
					if _, ok := animator["range_offset_keyframes"]; ok {
						required["keyframes"] = true
					}
				}
			}
			if effects, ok := layer["effects"].([]any); ok && len(effects) > 0 {
				required["effects"] = true
			}
			if shape, ok := layer["shape"].(map[string]any); ok && shapeRequiresPropertyProfile(layer, shape) {
				required["properties"] = true
			}
			if masks, ok := layer["masks"].([]any); ok && len(masks) > 0 {
				required["masks"] = true
			}
			if camera, ok := layer["camera"].(map[string]any); ok && len(camera) > 0 {
				required["properties"] = true
			}
			if light, ok := layer["light"].(map[string]any); ok && lightRequiresPropertyProfile(light) {
				required["properties"] = true
			}
		}
	}
	keys := make([]string, 0, len(required))
	for key := range required {
		keys = append(keys, key)
	}
	return keys
}

func shapeRequiresPropertyProfile(layer map[string]any, shape map[string]any) bool {
	if len(shape) == 0 {
		return false
	}
	if layer["type"] != "solid" {
		return true
	}
	for key := range shape {
		if key != "fill_color" {
			return true
		}
	}
	return false
}

func lightRequiresPropertyProfile(light map[string]any) bool {
	for key := range light {
		if key != "kind" && key != "source_layer" {
			return true
		}
	}
	return false
}

func expectedProfileListLen(expectedProfile map[string]any, key string) int {
	values, ok := expectedProfile[key].([]any)
	if !ok {
		return 0
	}
	return len(values)
}

func recipeExampleAuthoredTransformKeyframeStreams(doc map[string]any) int {
	comps, ok := doc["comps"].([]any)
	if !ok {
		return 0
	}
	keyframeFields := map[string]bool{
		"position_keyframes":     true,
		"anchor_point_keyframes": true,
		"scale_keyframes":        true,
		"rotation_keyframes":     true,
		"opacity_keyframes":      true,
	}
	var count int
	for _, rawComp := range comps {
		comp, ok := rawComp.(map[string]any)
		if !ok {
			continue
		}
		layers, ok := comp["layers"].([]any)
		if !ok {
			continue
		}
		for _, rawLayer := range layers {
			layer, ok := rawLayer.(map[string]any)
			if !ok {
				continue
			}
			transform, ok := layer["transform"].(map[string]any)
			if !ok {
				continue
			}
			for field := range keyframeFields {
				values, ok := transform[field].([]any)
				if ok && len(values) > 0 {
					count++
				}
			}
		}
	}
	return count
}

func TestCompileToFileSetsCompBackgroundColor(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].BackgroundColor = []float64{12, 34, 56}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	if got, want := project.Compositions[0].BGColor, ([3]uint8{12, 34, 56}); got != want {
		t.Fatalf("BGColor = %v, want %v", got, want)
	}
}

func TestCompileToFileSetsCompLabel(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Label = ptr(12)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	if got := project.Compositions[0].Label; got != 12 {
		t.Fatalf("label = %d, want 12", got)
	}
}

func TestCompileToFileSetsCompComment(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Comment = "reviewed recipe comp"
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	if got := project.Compositions[0].Comment; got != "reviewed recipe comp" {
		t.Fatalf("comment = %q, want reviewed recipe comp", got)
	}
}

func TestCompileToFileSetsMotionGraphicsTemplateName(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "EG template"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"motion_graphics_template_name": "Lower Third Pack"
		}],
		"expected_profile": {
			"motion_graphics_template_name": "Lower Third Pack"
		}
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	if got := project.Compositions[0].MotionGraphicsTemplateName; got != "Lower Third Pack" {
		t.Fatalf("MotionGraphicsTemplateName = %q, want Lower Third Pack", got)
	}
	assertProfileCheck(t, report, "expected_profile.motion_graphics_template_name", true)
}

func TestCompileToFileSetsCompDraft3D(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Draft 3D"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"draft_3d": true,
			"layers": [{
				"type": "text",
				"name": "Title",
				"text": "Draft 3D",
				"transform": {"position": [960, 540]}
			}]
		}],
		"expected_profile": {
			"draft_3d": true
		}
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.draft_3d", true)
}

func TestCompileToFileChecksCompDisplayProfile(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Comp display profile"},
		"comps": [{
			"name": "Main",
			"width": 1280,
			"height": 720,
			"frame_rate": 24,
			"duration": 3,
			"background_color": [12, 34, 56],
			"resolution_factor": [2, 2],
			"pixel_aspect": 2,
			"display_start_time": 0.5
		}],
		"expected_profile": {
			"comp_count": 1,
			"layer_count": 0,
			"background_color": [12, 34, 56],
			"resolution_factor": [2, 2],
			"pixel_aspect": 2,
			"display_start_time": 0.5
		}
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.background_color", true)
	assertProfileCheck(t, report, "expected_profile.resolution_factor", true)
	assertProfileCheck(t, report, "expected_profile.pixel_aspect", true)
	assertProfileCheck(t, report, "expected_profile.display_start_time", true)
}

func TestCompileToFileChecksCompFlagProfile(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Comp flag profile"},
		"comps": [{
			"name": "Main",
			"width": 1280,
			"height": 720,
			"frame_rate": 24,
			"duration": 3,
			"background_color": [0, 0, 0],
			"frame_blending": true,
			"hide_shy_layers": true,
			"preserve_nested_frame_rate": true,
			"preserve_nested_resolution": true,
			"motion_blur": {
				"enabled": true
			}
		}],
		"expected_profile": {
			"comp_count": 1,
			"layer_count": 0,
			"frame_blending": true,
			"hide_shy_layers": true,
			"preserve_nested_frame_rate": true,
			"preserve_nested_resolution": true,
			"motion_blur": {
				"enabled": true
			}
		}
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.frame_blending", true)
	assertProfileCheck(t, report, "expected_profile.hide_shy_layers", true)
	assertProfileCheck(t, report, "expected_profile.preserve_nested_frame_rate", true)
	assertProfileCheck(t, report, "expected_profile.preserve_nested_resolution", true)
	assertProfileCheck(t, report, "expected_profile.motion_blur.enabled", true)
}

func TestCompileToFileChecksCompMetadataProfile(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Comp metadata profile"},
		"comps": [{
			"name": "Main",
			"width": 1280,
			"height": 720,
			"frame_rate": 24,
			"duration": 3,
			"background_color": [0, 0, 0],
			"label": 12,
			"comment": "reviewed recipe comp"
		}],
		"expected_profile": {
			"comp_count": 1,
			"layer_count": 0,
			"label": 12,
			"comment": "reviewed recipe comp"
		}
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.label", true)
	assertProfileCheck(t, report, "expected_profile.comment", true)
}

func TestCompileToFileChecksCompObjectProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-comp-object-profile.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	for _, path := range []string{
		"expected_profile.name",
		"expected_profile.width",
		"expected_profile.height",
		"expected_profile.frame_rate",
		"expected_profile.duration",
		"expected_profile.background_color",
		"expected_profile.resolution_factor",
		"expected_profile.pixel_aspect",
		"expected_profile.display_start_time",
		"expected_profile.renderer",
		"expected_profile.frame_blending",
		"expected_profile.hide_shy_layers",
		"expected_profile.preserve_nested_frame_rate",
		"expected_profile.preserve_nested_resolution",
		"expected_profile.motion_blur.enabled",
		"expected_profile.motion_blur.shutter_angle",
		"expected_profile.motion_blur.shutter_phase",
		"expected_profile.motion_blur.adaptive_sample_limit",
		"expected_profile.motion_blur.samples_per_frame",
		"expected_profile.work_area.start",
		"expected_profile.work_area.end",
		"expected_profile.label",
		"expected_profile.comment",
	} {
		assertProfileCheck(t, report, path, true)
	}
}

func TestCompileToFileChecksStandaloneCompProfileExamples(t *testing.T) {
	cases := []struct {
		recipe string
		paths  []string
	}{
		{
			recipe: "minimal-comp-background-color.json",
			paths:  []string{"expected_profile.background_color"},
		},
		{
			recipe: "minimal-comp-label.json",
			paths:  []string{"expected_profile.label"},
		},
		{
			recipe: "minimal-comp-comment.json",
			paths:  []string{"expected_profile.comment"},
		},
		{
			recipe: "minimal-comp-resolution-factor.json",
			paths:  []string{"expected_profile.resolution_factor"},
		},
		{
			recipe: "minimal-comp-pixel-aspect.json",
			paths:  []string{"expected_profile.pixel_aspect"},
		},
		{
			recipe: "minimal-comp-display-start-time.json",
			paths:  []string{"expected_profile.display_start_time"},
		},
		{
			recipe: "minimal-comp-frame-blending.json",
			paths:  []string{"expected_profile.frame_blending"},
		},
		{
			recipe: "minimal-comp-hide-shy-layers.json",
			paths:  []string{"expected_profile.hide_shy_layers"},
		},
		{
			recipe: "minimal-comp-preserve-nested-frame-rate.json",
			paths:  []string{"expected_profile.preserve_nested_frame_rate"},
		},
		{
			recipe: "minimal-comp-preserve-nested-resolution.json",
			paths:  []string{"expected_profile.preserve_nested_resolution"},
		},
		{
			recipe: "minimal-comp-motion-blur-enabled.json",
			paths:  []string{"expected_profile.motion_blur.enabled"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.recipe, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", tc.recipe))
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			rec := mustUnmarshalRecipe(t, string(raw))
			outPath := filepath.Join(t.TempDir(), "recipe.aep")

			report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
			if err != nil {
				t.Fatalf("CompileToFile: %v", err)
			}
			if !report.Valid {
				t.Fatalf("report = %+v, want valid", report)
			}
			for _, path := range tc.paths {
				assertProfileCheck(t, report, path, true)
			}
		})
	}
}

func TestCompileToFileChecksTransformKeyframeProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-transform-keyframes.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	for _, path := range []string{
		"expected_profile.layers[0].name",
		"expected_profile.layers[0].type",
		"expected_profile.keyframes[0].keyframes[0]",
		"expected_profile.keyframes[0].keyframes[1]",
		"expected_profile.keyframes[0].keyframes[2]",
		"expected_profile.keyframes[1].keyframes[0]",
		"expected_profile.keyframes[1].keyframes[1]",
		"expected_profile.keyframes[1].keyframes[2]",
		"expected_profile.keyframes[2].keyframes[0]",
		"expected_profile.keyframes[2].keyframes[1]",
		"expected_profile.keyframes[2].keyframes[2]",
		"expected_profile.keyframes[3].keyframes[0]",
		"expected_profile.keyframes[3].keyframes[1]",
		"expected_profile.keyframes[3].keyframes[2]",
		"expected_profile.keyframes[4].keyframes[0]",
		"expected_profile.keyframes[4].keyframes[1]",
		"expected_profile.keyframes[4].keyframes[2]",
	} {
		assertProfileCheck(t, report, path, true)
	}
}

func TestCompileToFileChecksTextStyleProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-text-style.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	for _, path := range []string{
		"expected_profile.layers[0].name",
		"expected_profile.layers[0].type",
		"expected_profile.text_styles[0].font_size",
		"expected_profile.text_styles[0].fill_color",
		"expected_profile.text_styles[0].tracking",
		"expected_profile.text_styles[0].faux_bold",
		"expected_profile.text_styles[0].faux_italic",
		"expected_profile.text_styles[0].apply_stroke",
		"expected_profile.text_styles[0].stroke_color",
		"expected_profile.text_styles[0].stroke_width",
		"expected_profile.text_styles[0].justification",
	} {
		assertProfileCheck(t, report, path, true)
	}
}

func TestCompileToFileChecksTextAnimatorOpacityProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-text-animator-opacity.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.layers[0].name", true)
	assertProfileCheck(t, report, "expected_profile.layers[0].type", true)
	assertProfileCheck(t, report, "expected_profile.properties[0]", true)
}

func TestCompileToFileChecksTextAnimatorPositionProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-text-animator-position.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.layers[0].name", true)
	assertProfileCheck(t, report, "expected_profile.layers[0].type", true)
	assertProfileCheck(t, report, "expected_profile.properties[0]", true)
}

func TestCompileToFileChecksTextAnimatorScaleProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-text-animator-scale.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.layers[0].name", true)
	assertProfileCheck(t, report, "expected_profile.layers[0].type", true)
	assertProfileCheck(t, report, "expected_profile.properties[0]", true)
}

func TestCompileToFileChecksTextAnimatorRotationProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-text-animator-rotation.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.layers[0].name", true)
	assertProfileCheck(t, report, "expected_profile.layers[0].type", true)
	assertProfileCheck(t, report, "expected_profile.properties[0]", true)
}

func TestCompileToFileChecksTextAnimatorColorProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-text-animator-color.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.layers[0].name", true)
	assertProfileCheck(t, report, "expected_profile.layers[0].type", true)
	assertProfileCheck(t, report, "expected_profile.properties[0]", true)
}

func TestCompileToFileChecksTextAnimatorTrackingProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-text-animator-tracking.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.layers[0].name", true)
	assertProfileCheck(t, report, "expected_profile.layers[0].type", true)
	assertProfileCheck(t, report, "expected_profile.properties[0]", true)
}

func TestCompileToFileChecksTextAnimatorCharacterOffsetProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-text-animator-character-offset.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.layers[0].name", true)
	assertProfileCheck(t, report, "expected_profile.layers[0].type", true)
	assertProfileCheck(t, report, "expected_profile.properties[0]", true)
}

func TestCompileToFileChecksTextAnimatorFillOpacityProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-text-animator-fill-opacity.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.layers[0].name", true)
	assertProfileCheck(t, report, "expected_profile.layers[0].type", true)
	assertProfileCheck(t, report, "expected_profile.properties[0]", true)
}

func TestCompileToFileChecksTextAnimatorStrokeOpacityProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-text-animator-stroke-opacity.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.layers[0].name", true)
	assertProfileCheck(t, report, "expected_profile.layers[0].type", true)
	assertProfileCheck(t, report, "expected_profile.properties[0]", true)
}

func TestCompileToFileChecksTextAnimatorStrokeWidthProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-text-animator-stroke-width.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.layers[0].name", true)
	assertProfileCheck(t, report, "expected_profile.layers[0].type", true)
	assertProfileCheck(t, report, "expected_profile.properties[0]", true)
}

func TestCompileToFileChecksTextAnimatorSkewProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-text-animator-skew.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.layers[0].name", true)
	assertProfileCheck(t, report, "expected_profile.layers[0].type", true)
	assertProfileCheck(t, report, "expected_profile.properties[0]", true)
}

func TestCompileToFileChecksTextAnimatorRotationXProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-text-animator-rotation-x.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.layers[0].name", true)
	assertProfileCheck(t, report, "expected_profile.layers[0].type", true)
	assertProfileCheck(t, report, "expected_profile.properties[0]", true)
}

func TestCompileToFileChecksTextAnimatorRotationYProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-text-animator-rotation-y.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.layers[0].name", true)
	assertProfileCheck(t, report, "expected_profile.layers[0].type", true)
	assertProfileCheck(t, report, "expected_profile.properties[0]", true)
}

func TestCompileToFileChecksTextAnimatorStrokeColorProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-text-animator-stroke-color.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.layers[0].name", true)
	assertProfileCheck(t, report, "expected_profile.layers[0].type", true)
	assertProfileCheck(t, report, "expected_profile.properties[0]", true)
}

func TestCompileToFileChecksTextAnimatorRangeOffsetKeyframesExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-text-animator-range-offset.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.layers[0].name", true)
	assertProfileCheck(t, report, "expected_profile.layers[0].type", true)
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[0]", true)
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[1]", true)
}

func TestCompileToFileChecksTextAnimatorOpacityValueKeyframesExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-text-animator-opacity-value-keyframes.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.layers[0].name", true)
	assertProfileCheck(t, report, "expected_profile.layers[0].type", true)
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[0]", true)
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[1]", true)
}

func TestCompileToFileChecksTextAnimatorScalarValueKeyframesExamples(t *testing.T) {
	for _, name := range []string{
		"minimal-text-animator-rotation-value-keyframes.json",
		"minimal-text-animator-tracking-value-keyframes.json",
		"minimal-text-animator-character-offset-value-keyframes.json",
		"minimal-text-animator-position-value-keyframes.json",
		"minimal-text-animator-scale-value-keyframes.json",
		"minimal-text-animator-color-value-keyframes.json",
	} {
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", name))
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			rec := mustUnmarshalRecipe(t, string(raw))
			outPath := filepath.Join(t.TempDir(), "recipe.aep")

			report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
			if err != nil {
				t.Fatalf("CompileToFile: %v", err)
			}
			if !report.Valid {
				t.Fatalf("report = %+v, want valid", report)
			}
			assertProfileCheck(t, report, "expected_profile.layers[0].name", true)
			assertProfileCheck(t, report, "expected_profile.layers[0].type", true)
			assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[0]", true)
			assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[1]", true)
		})
	}
}

func TestCompileToFileChecksShapeFilterProfileExamples(t *testing.T) {
	cases := []struct {
		recipe string
		paths  []string
	}{
		{
			recipe: "minimal-shape-trim.json",
			paths: []string{
				"expected_profile.properties[0]",
				"expected_profile.properties[1]",
				"expected_profile.properties[2]",
			},
		},
		{
			recipe: "minimal-shape-round-corners.json",
			paths:  []string{"expected_profile.properties[0]"},
		},
		{
			recipe: "minimal-shape-offset-paths.json",
			paths: []string{
				"expected_profile.properties[0]",
				"expected_profile.properties[1]",
				"expected_profile.properties[2]",
				"expected_profile.properties[3]",
				"expected_profile.properties[4]",
			},
		},
		{
			recipe: "minimal-shape-zigzag.json",
			paths: []string{
				"expected_profile.properties[0]",
				"expected_profile.properties[1]",
				"expected_profile.properties[2]",
			},
		},
		{
			recipe: "minimal-shape-pucker-bloat.json",
			paths:  []string{"expected_profile.properties[0]"},
		},
		{
			recipe: "minimal-shape-twist.json",
			paths: []string{
				"expected_profile.properties[0]",
				"expected_profile.properties[1]",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.recipe, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", tc.recipe))
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			rec := mustUnmarshalRecipe(t, string(raw))
			outPath := filepath.Join(t.TempDir(), "recipe.aep")

			report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
			if err != nil {
				t.Fatalf("CompileToFile: %v", err)
			}
			if !report.Valid {
				t.Fatalf("report = %+v, want valid", report)
			}
			for _, path := range tc.paths {
				assertProfileCheck(t, report, path, true)
			}
		})
	}
}

func TestCompileToFileChecksExplicitMatteProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-layer-explicit-matte.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.layers[0].track_matte", true)
	assertProfileCheck(t, report, "expected_profile.layers[0].matte", true)
}

func TestCompileToFileChecksLayerObjectProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-layer-object-profile.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	for _, path := range []string{
		"expected_profile.layers[0].name",
		"expected_profile.layers[0].type",
		"expected_profile.layers[0].quality",
		"expected_profile.layers[0].blending_mode",
		"expected_profile.layers[0].auto_orient",
		"expected_profile.layers[0].label",
		"expected_profile.layers[0].comment",
		"expected_profile.layers[0].timing.start_time",
		"expected_profile.layers[0].timing.in_point",
		"expected_profile.layers[0].timing.out_point",
		"expected_profile.layers[0].timing.duration",
		"expected_profile.layers[0].timing.stretch",
		"expected_profile.layers[0].flags.visible",
		"expected_profile.layers[0].flags.solo",
		"expected_profile.layers[0].flags.locked",
		"expected_profile.layers[0].flags.shy",
		"expected_profile.layers[0].flags.motion_blur",
		"expected_profile.layers[0].flags.effects_enabled",
		"expected_profile.layers[0].flags.audio_enabled",
		"expected_profile.layers[0].flags.frame_blend_enabled",
		"expected_profile.layers[0].flags.markers_locked",
		"expected_profile.layers[0].flags.frame_blend_pixel_motion",
		"expected_profile.layers[0].flags.collapse_transform",
		"expected_profile.layers[0].flags.is_3d",
		"expected_profile.layers[0].flags.is_adjustment",
		"expected_profile.layers[0].flags.is_guide",
		"expected_profile.layers[0].flags.sampling_bicubic",
		"expected_profile.layers[0].flags.preserve_transparency",
	} {
		assertProfileCheck(t, report, path, true)
	}
}

func TestCompileToFileChecksCameraAndLightLayerProfileExamples(t *testing.T) {
	cases := []struct {
		recipe string
		paths  []string
	}{
		{
			recipe: "minimal-camera-layer.json",
			paths: []string{
				"expected_profile.layers[0].name",
				"expected_profile.layers[0].type",
				"expected_profile.layers[1].name",
				"expected_profile.layers[1].type",
			},
		},
		{
			recipe: "minimal-light-layer.json",
			paths: []string{
				"expected_profile.layers[0].name",
				"expected_profile.layers[0].type",
				"expected_profile.layers[1].name",
				"expected_profile.layers[1].type",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.recipe, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", tc.recipe))
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			rec := mustUnmarshalRecipe(t, string(raw))
			outPath := filepath.Join(t.TempDir(), "recipe.aep")

			report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
			if err != nil {
				t.Fatalf("CompileToFile: %v", err)
			}
			if !report.Valid {
				t.Fatalf("report = %+v, want valid", report)
			}
			for _, path := range tc.paths {
				assertProfileCheck(t, report, path, true)
			}
		})
	}
}

func TestCompileToFileChecksNullAndAdjustmentLayerProfileExamples(t *testing.T) {
	cases := []struct {
		recipe string
		paths  []string
	}{
		{
			recipe: "minimal-null-layer.json",
			paths: []string{
				"expected_profile.layers[0].name",
				"expected_profile.layers[0].type",
				"expected_profile.layers[0].flags.is_null",
				"expected_profile.layers[1].parent",
			},
		},
		{
			recipe: "minimal-adjustment-layer.json",
			paths: []string{
				"expected_profile.layers[0].name",
				"expected_profile.layers[0].type",
				"expected_profile.layers[0].flags.is_adjustment",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.recipe, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", tc.recipe))
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			rec := mustUnmarshalRecipe(t, string(raw))
			outPath := filepath.Join(t.TempDir(), "recipe.aep")

			report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
			if err != nil {
				t.Fatalf("CompileToFile: %v", err)
			}
			if !report.Valid {
				t.Fatalf("report = %+v, want valid", report)
			}
			for _, path := range tc.paths {
				assertProfileCheck(t, report, path, true)
			}
		})
	}
}

func TestCompileToFileChecksStandaloneLayerProfileExamples(t *testing.T) {
	cases := []struct {
		recipe string
		paths  []string
	}{
		{
			recipe: "minimal-layer-label.json",
			paths: []string{
				"expected_profile.layers[0].name",
				"expected_profile.layers[0].type",
				"expected_profile.layers[0].label",
			},
		},
		{
			recipe: "minimal-layer-comment.json",
			paths: []string{
				"expected_profile.layers[0].name",
				"expected_profile.layers[0].type",
				"expected_profile.layers[0].comment",
			},
		},
		{
			recipe: "minimal-layer-timing.json",
			paths: []string{
				"expected_profile.layers[0].name",
				"expected_profile.layers[0].type",
				"expected_profile.layers[0].timing.start_time",
				"expected_profile.layers[0].timing.in_point",
				"expected_profile.layers[0].timing.out_point",
				"expected_profile.layers[0].timing.duration",
				"expected_profile.layers[0].timing.stretch",
			},
		},
		{
			recipe: "minimal-layer-common-switches.json",
			paths: []string{
				"expected_profile.layers[0].name",
				"expected_profile.layers[0].type",
				"expected_profile.layers[0].flags.visible",
				"expected_profile.layers[0].flags.solo",
				"expected_profile.layers[0].flags.locked",
				"expected_profile.layers[0].flags.effects_enabled",
				"expected_profile.layers[0].flags.audio_enabled",
				"expected_profile.layers[0].flags.frame_blend_enabled",
			},
		},
		{
			recipe: "minimal-layer-motion-blur.json",
			paths: []string{
				"expected_profile.layers[0].name",
				"expected_profile.layers[0].type",
				"expected_profile.layers[0].flags.motion_blur",
			},
		},
		{
			recipe: "minimal-layer-shy.json",
			paths: []string{
				"expected_profile.layers[0].name",
				"expected_profile.layers[0].type",
				"expected_profile.layers[0].flags.shy",
			},
		},
		{
			recipe: "minimal-layer-advanced-switches.json",
			paths: []string{
				"expected_profile.layers[0].name",
				"expected_profile.layers[0].type",
				"expected_profile.layers[0].flags.collapse_transform",
				"expected_profile.layers[0].flags.is_3d",
				"expected_profile.layers[0].flags.is_adjustment",
				"expected_profile.layers[0].flags.is_guide",
				"expected_profile.layers[0].flags.frame_blend_enabled",
				"expected_profile.layers[0].flags.frame_blend_pixel_motion",
				"expected_profile.layers[0].flags.sampling_bicubic",
				"expected_profile.layers[0].flags.preserve_transparency",
			},
		},
		{
			recipe: "minimal-layer-quality-blending.json",
			paths: []string{
				"expected_profile.layers[0].name",
				"expected_profile.layers[0].type",
				"expected_profile.layers[0].quality",
				"expected_profile.layers[0].blending_mode",
			},
		},
		{
			recipe: "minimal-layer-auto-orient.json",
			paths: []string{
				"expected_profile.layers[0].name",
				"expected_profile.layers[0].type",
				"expected_profile.layers[0].auto_orient",
			},
		},
		{
			recipe: "minimal-layer-null-flag.json",
			paths: []string{
				"expected_profile.layers[0].name",
				"expected_profile.layers[0].type",
				"expected_profile.layers[0].flags.visible",
				"expected_profile.layers[0].flags.is_null",
				"expected_profile.layers[1].name",
				"expected_profile.layers[1].type",
				"expected_profile.layers[1].parent",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.recipe, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", tc.recipe))
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			rec := mustUnmarshalRecipe(t, string(raw))
			outPath := filepath.Join(t.TempDir(), "recipe.aep")

			report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
			if err != nil {
				t.Fatalf("CompileToFile: %v", err)
			}
			if !report.Valid {
				t.Fatalf("report = %+v, want valid", report)
			}
			for _, path := range tc.paths {
				assertProfileCheck(t, report, path, true)
			}
		})
	}
}

func TestCompileToFileChecksCameraObjectProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-camera-object-profile.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	for _, path := range []string{
		"expected_profile.layers[0].name",
		"expected_profile.layers[0].type",
		"expected_profile.properties[0]",
		"expected_profile.properties[1]",
		"expected_profile.properties[2]",
		"expected_profile.properties[3]",
		"expected_profile.properties[4]",
		"expected_profile.properties[5]",
		"expected_profile.properties[6]",
		"expected_profile.properties[7]",
		"expected_profile.properties[8]",
		"expected_profile.properties[9]",
		"expected_profile.properties[10]",
		"expected_profile.properties[11]",
		"expected_profile.properties[12]",
	} {
		assertProfileCheck(t, report, path, true)
	}
}

func TestCompileToFileChecksLightObjectProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-light-object-profile.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	for _, path := range []string{
		"expected_profile.layers[0].name",
		"expected_profile.layers[0].type",
		"expected_profile.layers[0].light_kind",
		"expected_profile.properties[0]",
		"expected_profile.properties[1]",
		"expected_profile.properties[2]",
		"expected_profile.properties[3]",
		"expected_profile.properties[4]",
		"expected_profile.properties[5]",
		"expected_profile.properties[6]",
		"expected_profile.properties[7]",
		"expected_profile.properties[8]",
		"expected_profile.properties[9]",
	} {
		assertProfileCheck(t, report, path, true)
	}
}

func TestCompileToFileChecksLightKindProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-light-kind.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	for _, path := range []string{
		"expected_profile.layers[0].name",
		"expected_profile.layers[0].type",
		"expected_profile.layers[0].light_kind",
		"expected_profile.layers[1].name",
		"expected_profile.layers[1].type",
	} {
		assertProfileCheck(t, report, path, true)
	}
}

func TestCompileToFileChecksCameraAndLightOptionProfileExamples(t *testing.T) {
	cases := []string{
		"minimal-camera-aperture.json",
		"minimal-camera-blur-level.json",
		"minimal-camera-depth-of-field.json",
		"minimal-camera-focus-distance.json",
		"minimal-camera-iris-aspect-ratio.json",
		"minimal-camera-iris-diffraction-fringe.json",
		"minimal-camera-iris-highlight-gain.json",
		"minimal-camera-iris-highlight-saturation.json",
		"minimal-camera-iris-highlight-threshold.json",
		"minimal-camera-iris-rotation.json",
		"minimal-camera-iris-roundness.json",
		"minimal-camera-iris-shape.json",
		"minimal-camera-zoom.json",
		"minimal-light-casts-shadows.json",
		"minimal-light-color.json",
		"minimal-light-cone-angle.json",
		"minimal-light-cone-feather.json",
		"minimal-light-falloff-distance.json",
		"minimal-light-falloff-start.json",
		"minimal-light-falloff-type.json",
		"minimal-light-intensity.json",
		"minimal-light-shadow-darkness.json",
		"minimal-light-shadow-diffusion.json",
	}
	for _, recipeName := range cases {
		t.Run(recipeName, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", recipeName))
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			rec := mustUnmarshalRecipe(t, string(raw))
			outPath := filepath.Join(t.TempDir(), "recipe.aep")

			report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
			if err != nil {
				t.Fatalf("CompileToFile: %v", err)
			}
			if !report.Valid {
				t.Fatalf("report = %+v, want valid", report)
			}
			for _, path := range []string{
				"expected_profile.layers[0].name",
				"expected_profile.layers[0].type",
				"expected_profile.layers[1].name",
				"expected_profile.layers[1].type",
			} {
				assertProfileCheck(t, report, path, true)
			}
		})
	}
}

func TestCompileToFileChecksLightSourceProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-light-source.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	for _, path := range []string{
		"expected_profile.layers[0].name",
		"expected_profile.layers[0].type",
		"expected_profile.layers[0].light_kind",
		"expected_profile.layers[0].light_source",
		"expected_profile.layers[1].name",
		"expected_profile.layers[1].type",
	} {
		assertProfileCheck(t, report, path, true)
	}
}

func TestCompileToFileChecksLayerSourceProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-layer-source-ref.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	for _, path := range []string{
		"expected_profile.layers[0].name",
		"expected_profile.layers[0].type",
		"expected_profile.layers[0].source",
		"expected_profile.layers[0].source_kind",
	} {
		assertProfileCheck(t, report, path, true)
	}
}

func TestCompileToFileSetsLayerLabel(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Label = ptr(10)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	if got := project.Compositions[0].Layers[0].Label; got != 10 {
		t.Fatalf("layer label = %d, want 10", got)
	}
}

func TestCompileToFileSetsLayerComment(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Comment = "reviewed recipe layer"
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	if got := project.Compositions[0].Layers[0].Comment; got != "reviewed recipe layer" {
		t.Fatalf("layer comment = %q, want %q", got, "reviewed recipe layer")
	}
}

func TestCompileToFileSetsLayerMotionBlur(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].MotionBlur = boolPtr(true)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	if got := project.Compositions[0].Layers[0].MotionBlur; !got {
		t.Fatal("layer motion_blur = false, want true")
	}
}

func TestCompileToFileSetsLayerShy(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Shy = boolPtr(true)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	if got := project.Compositions[0].Layers[0].Shy; !got {
		t.Fatal("layer shy = false, want true")
	}
}

func TestCompileToFileSetsLayerCommonSwitches(t *testing.T) {
	rec := minimalRecipe()
	layer := &rec.Comps[0].Layers[0]
	layer.Visible = boolPtr(false)
	layer.Solo = boolPtr(true)
	layer.Locked = boolPtr(true)
	layer.EffectsEnabled = boolPtr(false)
	layer.AudioEnabled = boolPtr(false)
	layer.FrameBlendEnabled = boolPtr(true)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0]
	if got.Visible {
		t.Fatal("layer visible = true, want false")
	}
	if !got.Solo {
		t.Fatal("layer solo = false, want true")
	}
	if !got.Locked {
		t.Fatal("layer locked = false, want true")
	}
	if got.EffectsEnabled {
		t.Fatal("layer effects_enabled = true, want false")
	}
	if got.AudioEnabled {
		t.Fatal("layer audio_enabled = true, want false")
	}
	if !got.FrameBlendEnabled {
		t.Fatal("layer frame_blend_enabled = false, want true")
	}
}

func TestCompileToFileSetsLayerAdvancedSwitches(t *testing.T) {
	rec := minimalRecipe()
	layer := &rec.Comps[0].Layers[0]
	layer.CollapseTransform = boolPtr(true)
	layer.Is3D = boolPtr(true)
	layer.IsAdjust = boolPtr(true)
	layer.IsNull = boolPtr(true)
	layer.IsGuide = boolPtr(true)
	layer.MarkersLocked = boolPtr(true)
	layer.SamplingBicubic = boolPtr(true)
	layer.FrameBlendPixelMotion = boolPtr(true)
	layer.PreserveTransparency = boolPtr(true)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0]
	if !got.CollapseTransform {
		t.Fatal("layer collapse_transform = false, want true")
	}
	if !got.Is3D {
		t.Fatal("layer is_3d = false, want true")
	}
	if !got.IsAdjust {
		t.Fatal("layer is_adjust = false, want true")
	}
	if !got.IsNull {
		t.Fatal("layer is_null = false, want true")
	}
	if !got.IsGuide {
		t.Fatal("layer is_guide = false, want true")
	}
	if !got.MarkersLocked {
		t.Fatal("layer markers_locked = false, want true")
	}
	if !got.SamplingBicubic {
		t.Fatal("layer sampling_bicubic = false, want true")
	}
	if !got.FrameBlendPixelMotion {
		t.Fatal("layer frame_blend_pixel_motion = false, want true")
	}
	if !got.PreserveTransparency {
		t.Fatal("layer preserve_transparency = false, want true")
	}
}

func TestCompileToFileSetsLayerQualityAndBlendingMode(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Quality = "draft"
	rec.Comps[0].Layers[0].BlendingMode = "multiply"
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0]
	if got.Quality != aep.LayerQualityDraft {
		t.Fatalf("layer quality = %v, want draft", got.Quality)
	}
	if got.BlendingMode != aep.BlendingModeMultiply {
		t.Fatalf("layer blending_mode = %v, want multiply", got.BlendingMode)
	}
}

func TestCompileToFileSetsLayerAutoOrient(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].AutoOrient = "along_path"
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0]
	if got.AutoOrient != aep.AutoOrientAlongPath {
		t.Fatalf("layer auto_orient = %v, want along_path", got.AutoOrient)
	}
}

func TestCompileToFileSetsLayerTiming(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].StartTime = ptr(0.25)
	rec.Comps[0].Layers[0].InPoint = ptr(0.1)
	rec.Comps[0].Layers[0].OutPoint = ptr(0.9)
	rec.Comps[0].Layers[0].Stretch = ptr(0.5)
	rec.ExpectedProfile = recipe.ExpectedProfile{
		Layers: []recipe.ExpectedLayer{{
			Name: "Title",
			Timing: &recipe.ExpectedLayerTiming{
				StartTime: ptr(0.25),
				InPoint:   ptr(0.1),
				OutPoint:  ptr(0.9),
				Stretch:   ptr(0.5),
			},
		}},
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0]
	if math.Abs(got.StartTime-0.25) > 1e-6 {
		t.Fatalf("layer start_time = %g, want 0.25", got.StartTime)
	}
	if math.Abs(got.InPoint()-0.1) > 1e-6 {
		t.Fatalf("layer in_point = %g, want 0.1", got.InPoint())
	}
	if math.Abs(got.OutPoint()-0.9) > 1e-6 {
		t.Fatalf("layer out_point = %g, want 0.9", got.OutPoint())
	}
	if math.Abs(got.Stretch-0.5) > 1e-6 {
		t.Fatalf("layer stretch = %g, want 0.5", got.Stretch)
	}
	assertProfileCheck(t, report, "expected_profile.layers[0].timing.stretch", true)
}

func TestCompileToFileSetsLayerParent(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers = append([]recipe.Layer{
		{
			Type: "text",
			Name: "Parent",
			Text: "Parent",
			Transform: recipe.Transform{
				Position: []float64{960, 540},
			},
		},
	}, rec.Comps[0].Layers...)
	rec.Comps[0].Layers[1].Parent = "Parent"
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	gotParent := project.Compositions[0].Layers[0]
	gotChild := project.Compositions[0].Layers[1]
	if gotChild.ParentID != gotParent.ID {
		t.Fatalf("child ParentID = %d, want parent ID %d", gotChild.ParentID, gotParent.ID)
	}
}

func TestCompileToFileLayerParentDuplicateNamesKeepsLastMatchSemantics(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers = []recipe.Layer{
		{
			Type: "text",
			Name: "Parent",
			Text: "First",
			Transform: recipe.Transform{
				Position: []float64{960, 480},
			},
		},
		{
			Type: "text",
			Name: "Parent",
			Text: "Second",
			Transform: recipe.Transform{
				Position: []float64{960, 520},
			},
		},
		{
			Type:   "text",
			Name:   "Child",
			Text:   "Child",
			Parent: "Parent",
			Transform: recipe.Transform{
				Position: []float64{960, 560},
			},
		},
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	gotFirstParent := project.Compositions[0].Layers[0]
	gotSecondParent := project.Compositions[0].Layers[1]
	gotChild := project.Compositions[0].Layers[2]
	if gotChild.ParentID == gotFirstParent.ID {
		t.Fatalf("child ParentID = first duplicate parent ID %d, want last duplicate parent ID %d", gotFirstParent.ID, gotSecondParent.ID)
	}
	if gotChild.ParentID != gotSecondParent.ID {
		t.Fatalf("child ParentID = %d, want last duplicate parent ID %d", gotChild.ParentID, gotSecondParent.ID)
	}
}

func TestCompileToFileChecksLayerParentProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-layer-parent.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.layers[0].name", true)
	assertProfileCheck(t, report, "expected_profile.layers[1].name", true)
	assertProfileCheck(t, report, "expected_profile.layers[1].parent", true)
}

func TestCompileToFileChecksLayerTrackMatteProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-layer-track-matte.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.layers[1].track_matte", true)
	assertProfileCheck(t, report, "expected_profile.layers[1].matte", true)
}

func TestCompileToFileCreatesNullLayer(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0] = recipe.Layer{
		Type: "null",
		Name: "Controller",
		Transform: recipe.Transform{
			Position: []float64{960, 540},
		},
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0]
	if !got.IsNull {
		t.Fatal("layer is_null = false, want true")
	}
}

func TestCompileToFileCreatesAdjustmentLayer(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0] = recipe.Layer{
		Type: "adjustment",
		Name: "Grade",
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0]
	if !got.IsAdjust {
		t.Fatal("layer is_adjust = false, want true")
	}
}

func TestCompileToFileCreatesCameraLayer(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0] = recipe.Layer{
		Type: "camera",
		Name: "Camera",
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0]
	if got.Type != aep.LayerTypeCamera {
		t.Fatalf("layer type = %v, want camera", got.Type)
	}
}

func TestCompileToFileCreatesLightLayer(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0] = recipe.Layer{
		Type: "light",
		Name: "Light",
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0]
	if got.Type != aep.LayerTypeLight {
		t.Fatalf("layer type = %v, want light", got.Type)
	}
}

func TestCompileToFileSetsLightIntensity(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Light intensity"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "light", "name": "Light", "light": {"intensity": 140}},
				{"type": "text", "name": "Title", "text": "Light intensity", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0].LightIntensity()
	if got == nil || got.StaticValue != 140.0 {
		t.Fatalf("light intensity = %+v, want 140", got)
	}
}

func TestCompileToFileSetsLightColor(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Light color"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "light", "name": "Light", "light": {"color": [255, 51, 102, 204]}},
				{"type": "text", "name": "Title", "text": "Light color", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0].LightColor()
	if got == nil {
		t.Fatal("light color property = nil")
	}
	values, ok := got.StaticValue.([]float64)
	if !ok {
		t.Fatalf("light color StaticValue = %v (%T), want []float64", got.StaticValue, got.StaticValue)
	}
	assertFloatArray(t, "LightColor", values, []float64{255, 51, 102, 204})
}

func TestCompileToFileSetsLightCastsShadows(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Light casts shadows"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "light", "name": "Light", "light": {"casts_shadows": true}},
				{"type": "text", "name": "Title", "text": "Light casts shadows", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0].LightCastsShadows()
	if got == nil || got.StaticValue != 1.0 {
		t.Fatalf("light casts_shadows = %+v, want 1", got)
	}
}

func TestCompileToFileSetsLightShadowDarkness(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Light shadow darkness"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "light", "name": "Light", "light": {"shadow_darkness": 75}},
				{"type": "text", "name": "Title", "text": "Light shadow darkness", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0].LightShadowDarkness()
	if got == nil || got.StaticValue != 75.0 {
		t.Fatalf("light shadow_darkness = %+v, want 75", got)
	}
}

func TestCompileToFileSetsLightShadowDiffusion(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Light shadow diffusion"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "light", "name": "Light", "light": {"shadow_diffusion": 18}},
				{"type": "text", "name": "Title", "text": "Light shadow diffusion", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0].LightShadowDiffusion()
	if got == nil || got.StaticValue != 18.0 {
		t.Fatalf("light shadow_diffusion = %+v, want 18", got)
	}
}

func TestCompileToFileSetsLightFalloffType(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Light falloff type"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "light", "name": "Light", "light": {"falloff_type": 2}},
				{"type": "text", "name": "Title", "text": "Light falloff type", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0].LightFalloffType()
	if got == nil || got.StaticValue != 2.0 {
		t.Fatalf("light falloff_type = %+v, want 2", got)
	}
}

func TestCompileToFileSetsLightFalloffStart(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Light falloff start"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "light", "name": "Light", "light": {"falloff_type": 2, "falloff_start": 100}},
				{"type": "text", "name": "Title", "text": "Light falloff start", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0].LightFalloffStart()
	if got == nil || got.StaticValue != 100.0 {
		t.Fatalf("light falloff_start = %+v, want 100", got)
	}
}

func TestCompileToFileSetsLightFalloffDistance(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Light falloff distance"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "light", "name": "Light", "light": {"falloff_type": 2, "falloff_distance": 750}},
				{"type": "text", "name": "Title", "text": "Light falloff distance", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0].LightFalloffDistance()
	if got == nil || got.StaticValue != 750.0 {
		t.Fatalf("light falloff_distance = %+v, want 750", got)
	}
}

func TestCompileToFileSetsLightConeAngle(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Light cone angle"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "light", "name": "Light", "light": {"cone_angle": 72}},
				{"type": "text", "name": "Title", "text": "Light cone angle", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0].LightConeAngle()
	if got == nil || got.StaticValue != 72.0 {
		t.Fatalf("light cone_angle = %+v, want 72", got)
	}
}

func TestCompileToFileSetsLightConeFeather(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Light cone feather"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "light", "name": "Light", "light": {"cone_feather": 35}},
				{"type": "text", "name": "Title", "text": "Light cone feather", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0].LightConeFeather()
	if got == nil || got.StaticValue != 35.0 {
		t.Fatalf("light cone_feather = %+v, want 35", got)
	}
}

func TestCompileToFileSetsLightKind(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Light kind"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "light", "name": "Light", "light": {"kind": "point"}},
				{"type": "text", "name": "Title", "text": "Light kind", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	if got := project.Compositions[0].Layers[0].LightKind; got != aep.LightKindPoint {
		t.Fatalf("light kind = %v, want point", got)
	}
}

func TestCompileToFileSetsLightSource(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Light source"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "light", "name": "Light", "light": {"kind": "ambient", "source_layer": "Source"}},
				{"type": "solid", "name": "Source", "shape": {"fill_color": [64, 128, 255]}},
				{"type": "text", "name": "Title", "text": "Light source", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0].LightSource()
	if got == nil || got.Name != "Source" {
		t.Fatalf("light source = %+v, want Source", got)
	}
}

func TestCompileToFileSetsCameraZoom(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Camera zoom"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "camera", "name": "Camera", "camera": {"zoom": 850}},
				{"type": "text", "name": "Title", "text": "Camera zoom", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0].CameraZoom()
	if got == nil || got.StaticValue != 850.0 {
		t.Fatalf("camera zoom = %+v, want 850", got)
	}
}

func TestCompileToFileSetsCameraDepthOfField(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Camera depth of field"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "camera", "name": "Camera", "camera": {"depth_of_field": false}},
				{"type": "text", "name": "Title", "text": "Camera DOF", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0].CameraDepthOfField()
	if got == nil || got.StaticValue != 0.0 {
		t.Fatalf("camera depth_of_field = %+v, want 0", got)
	}
}

func TestCompileToFileSetsCameraFocusDistance(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Camera focus distance"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "camera", "name": "Camera", "camera": {"focus_distance": 1200}},
				{"type": "text", "name": "Title", "text": "Camera focus", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0].CameraFocusDistance()
	if got == nil || got.StaticValue != 1200.0 {
		t.Fatalf("camera focus_distance = %+v, want 1200", got)
	}
}

func TestCompileToFileSetsCameraAperture(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Camera aperture"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "camera", "name": "Camera", "camera": {"aperture": 180}},
				{"type": "text", "name": "Title", "text": "Camera aperture", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0].CameraAperture()
	if got == nil || got.StaticValue != 180.0 {
		t.Fatalf("camera aperture = %+v, want 180", got)
	}
}

func TestCompileToFileSetsCameraBlurLevel(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Camera blur level"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "camera", "name": "Camera", "camera": {"blur_level": 120}},
				{"type": "text", "name": "Title", "text": "Camera blur", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0].CameraBlurLevel()
	if got == nil || got.StaticValue != 120.0 {
		t.Fatalf("camera blur_level = %+v, want 120", got)
	}
}

func TestCompileToFileSetsCameraIrisShape(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Camera iris shape"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "camera", "name": "Camera", "camera": {"iris_shape": 4}},
				{"type": "text", "name": "Title", "text": "Camera iris", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0].IrisShape()
	if got == nil || got.StaticValue != 4.0 {
		t.Fatalf("camera iris_shape = %+v, want 4", got)
	}
}

func TestCompileToFileSetsCameraIrisRotation(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Camera iris rotation"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "camera", "name": "Camera", "camera": {"iris_rotation": 25}},
				{"type": "text", "name": "Title", "text": "Camera iris", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0].IrisRotation()
	if got == nil || got.StaticValue != 25.0 {
		t.Fatalf("camera iris_rotation = %+v, want 25", got)
	}
}

func TestCompileToFileSetsCameraIrisRoundness(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Camera iris roundness"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "camera", "name": "Camera", "camera": {"iris_roundness": 60}},
				{"type": "text", "name": "Title", "text": "Camera iris", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0].IrisRoundness()
	if got == nil || got.StaticValue != 60.0 {
		t.Fatalf("camera iris_roundness = %+v, want 60", got)
	}
}

func TestCompileToFileSetsCameraIrisAspectRatio(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Camera iris aspect ratio"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "camera", "name": "Camera", "camera": {"iris_aspect_ratio": 1.8}},
				{"type": "text", "name": "Title", "text": "Camera iris", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0].IrisAspectRatio()
	if got == nil || got.StaticValue != 1.8 {
		t.Fatalf("camera iris_aspect_ratio = %+v, want 1.8", got)
	}
}

func TestCompileToFileSetsCameraIrisDiffractionFringe(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Camera iris diffraction fringe"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "camera", "name": "Camera", "camera": {"iris_diffraction_fringe": 30}},
				{"type": "text", "name": "Title", "text": "Camera iris", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0].IrisDiffractionFringe()
	if got == nil || got.StaticValue != 30.0 {
		t.Fatalf("camera iris_diffraction_fringe = %+v, want 30", got)
	}
}

func TestCompileToFileSetsCameraIrisHighlightGain(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Camera iris highlight gain"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "camera", "name": "Camera", "camera": {"iris_highlight_gain": 40}},
				{"type": "text", "name": "Title", "text": "Camera iris", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0].IrisHighlightGain()
	if got == nil || got.StaticValue != 40.0 {
		t.Fatalf("camera iris_highlight_gain = %+v, want 40", got)
	}
}

func TestCompileToFileSetsCameraIrisHighlightThreshold(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Camera iris highlight threshold"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "camera", "name": "Camera", "camera": {"iris_highlight_threshold": 0.7}},
				{"type": "text", "name": "Title", "text": "Camera iris", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0].IrisHighlightThreshold()
	if got == nil || got.StaticValue != 0.7 {
		t.Fatalf("camera iris_highlight_threshold = %+v, want 0.7", got)
	}
}

func TestCompileToFileSetsCameraIrisHighlightSaturation(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Camera iris highlight saturation"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [
				{"type": "camera", "name": "Camera", "camera": {"iris_highlight_saturation": 50}},
				{"type": "text", "name": "Title", "text": "Camera iris", "transform": {"position": [960, 540]}}
			]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	got := project.Compositions[0].Layers[0].IrisHighlightSaturation()
	if got == nil || got.StaticValue != 50.0 {
		t.Fatalf("camera iris_highlight_saturation = %+v, want 50", got)
	}
}

func TestCompileToFileSetsCompMotionBlur(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].MotionBlur = &recipe.CompMotionBlurSpec{
		ShutterAngle:        ptr(360),
		ShutterPhase:        ptr(-90),
		AdaptiveSampleLimit: ptr(256),
		SamplesPerFrame:     ptr(32),
	}
	rec.ExpectedProfile.MotionBlur = &recipe.ExpectedMotionBlurSpec{
		ShutterAngle:        ptr(360),
		ShutterPhase:        ptr(-90),
		AdaptiveSampleLimit: ptr(256),
		SamplesPerFrame:     ptr(32),
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.motion_blur.shutter_angle", true)
	assertProfileCheck(t, report, "expected_profile.motion_blur.shutter_phase", true)
	assertProfileCheck(t, report, "expected_profile.motion_blur.adaptive_sample_limit", true)
	assertProfileCheck(t, report, "expected_profile.motion_blur.samples_per_frame", true)

	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	comp := project.Compositions[0]
	if comp.ShutterAngle != 360 || comp.ShutterPhase != -90 || comp.MotionBlurAdaptiveSampleLimit != 256 || comp.MotionBlurSamplesPerFrame != 32 {
		t.Fatalf("motion blur = angle %d phase %d adaptive %d samples %d, want 360 -90 256 32",
			comp.ShutterAngle, comp.ShutterPhase, comp.MotionBlurAdaptiveSampleLimit, comp.MotionBlurSamplesPerFrame)
	}
}

func TestCompileToFileSetsCompMotionBlurEnabled(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].MotionBlur = &recipe.CompMotionBlurSpec{
		Enabled: boolPtr(true),
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	cdta := project.Compositions[0].CdtaRawBytes()
	if len(cdta) <= 0x8B {
		t.Fatalf("cdta len = %d, want > 0x8B", len(cdta))
	}
	if cdta[0x8B]&0x08 == 0 {
		t.Fatalf("cdta[0x8B] = 0x%02x, want comp motion blur bit 0x08 set", cdta[0x8B])
	}
}

func TestCompileToFileSetsCompWorkArea(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].WorkArea = &recipe.CompWorkAreaSpec{
		Start: ptr(1.5),
		End:   ptr(3.5),
	}
	rec.ExpectedProfile.WorkArea = &recipe.ExpectedWorkAreaSpec{
		Start: ptr(1.5),
		End:   ptr(3.5),
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.work_area.start", true)
	assertProfileCheck(t, report, "expected_profile.work_area.end", true)
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	comp := project.Compositions[0]
	if math.Abs(comp.WorkAreaStart-1.5) > 1e-6 || math.Abs(comp.WorkAreaEnd-3.5) > 1e-6 {
		t.Fatalf("work area = (%g, %g), want (1.5, 3.5)", comp.WorkAreaStart, comp.WorkAreaEnd)
	}
}

func TestCompileToFileSetsCompRenderer(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Renderer = "ADBE Escher"
	rec.ExpectedProfile.Renderer = "ADBE Escher"
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.renderer", true)
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	if got := project.Compositions[0].Renderer; got != "ADBE Escher" {
		t.Fatalf("renderer = %q, want ADBE Escher", got)
	}
}

func TestCompileToFileSetsCompResolutionFactor(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].ResolutionFactor = []float64{2, 2}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	if got := project.Compositions[0].ResolutionFactor; got != [2]uint16{2, 2} {
		t.Fatalf("resolution factor = %v, want [2 2]", got)
	}
}

func TestCompileToFileSetsCompPixelAspect(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].PixelAspect = ptr(2.0)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	if got := project.Compositions[0].PixelAspect; math.Abs(got-2.0) > 1e-6 {
		t.Fatalf("pixel aspect = %g, want 2", got)
	}
}

func TestCompileToFileSetsCompDisplayStartTime(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].DisplayStartTime = ptr(0.5)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	if got := project.Compositions[0].DisplayStartTime; math.Abs(got-0.5) > 1e-6 {
		t.Fatalf("display start time = %g, want 0.5", got)
	}
}

func TestCompileToFileSetsCompFrameBlending(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].FrameBlending = boolPtr(true)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	cdta := project.Compositions[0].CdtaRawBytes()
	if len(cdta) <= 0x8B {
		t.Fatalf("cdta len = %d, want > 0x8B", len(cdta))
	}
	if cdta[0x8B]&0x10 == 0 {
		t.Fatalf("cdta[0x8B] = 0x%02x, want frame blending bit 0x10 set", cdta[0x8B])
	}
}

func TestCompileToFileSetsCompHideShyLayers(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].HideShyLayers = boolPtr(true)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	cdta := project.Compositions[0].CdtaRawBytes()
	if len(cdta) <= 0x8B {
		t.Fatalf("cdta len = %d, want > 0x8B", len(cdta))
	}
	if cdta[0x8B]&0x01 == 0 {
		t.Fatalf("cdta[0x8B] = 0x%02x, want hide shy layers bit 0x01 set", cdta[0x8B])
	}
}

func TestCompileToFileSetsCompPreserveNestedFrameRate(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].PreserveNestedFrameRate = boolPtr(true)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	cdta := project.Compositions[0].CdtaRawBytes()
	if len(cdta) <= 0x8B {
		t.Fatalf("cdta len = %d, want > 0x8B", len(cdta))
	}
	if cdta[0x8B]&0x20 == 0 {
		t.Fatalf("cdta[0x8B] = 0x%02x, want preserve nested frame rate bit 0x20 set", cdta[0x8B])
	}
}

func TestCompileToFileSetsCompPreserveNestedResolution(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].PreserveNestedResolution = boolPtr(true)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	cdta := project.Compositions[0].CdtaRawBytes()
	if len(cdta) <= 0x8B {
		t.Fatalf("cdta len = %d, want > 0x8B", len(cdta))
	}
	if cdta[0x8B]&0x80 == 0 {
		t.Fatalf("cdta[0x8B] = 0x%02x, want preserve nested resolution bit 0x80 set", cdta[0x8B])
	}
}

func TestCompileToFileSetsShapeStroke(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Stroke = &recipe.StrokeSpec{
		Color:      []float64{255, 0, 0, 255},
		Width:      ptr(6),
		Opacity:    ptr(80),
		LineCap:    "projecting",
		LineJoin:   "bevel",
		MiterLimit: ptr(10),
		Taper: &recipe.StrokeTaperSpec{
			StartLength: ptr(30),
			EndLength:   ptr(40),
			StartWidth:  ptr(55),
			EndWidth:    ptr(65),
			StartEase:   ptr(35),
			EndEase:     ptr(50),
		},
		Wave: &recipe.StrokeWaveSpec{
			Amount:     ptr(20),
			Wavelength: ptr(70),
			Phase:      ptr(45),
		},
		Dashes: &recipe.StrokeDashesSpec{
			Dash: ptr(18),
			Gap:  ptr(7),
		},
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	layer := findProfileLayer(t, prof, "Underline")
	assertLayerPropertyValue(t, layer, "ADBE Vector Stroke Color", []float64{255, 255, 0, 0})
	assertLayerPropertyValue(t, layer, "ADBE Vector Stroke Width", 6.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Stroke Opacity", 80.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Stroke Line Cap", 3.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Stroke Line Join", 3.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Stroke Miter Limit", 10.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Taper Start Length", 30.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Taper End Length", 40.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Taper Start Width", 55.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Taper End Width", 65.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Taper Start Ease", 35.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Taper End Ease", 50.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Taper Wave Amount", 20.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Taper Wavelength", 70.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Taper Wave Phase", 45.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Stroke Dash 1", 18.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Stroke Gap 1", 7.0)
}

func TestCompileToFileSetsShapeStrokeCompositeOrder(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.FillColor = nil
	rec.Comps[0].Layers[1].Shape.Stroke = &recipe.StrokeSpec{
		Color:          []float64{255, 255, 255},
		Width:          ptr(18),
		CompositeOrder: "below_previous",
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	layer := findProfileLayer(t, prof, "Underline")
	assertLayerPropertyValue(t, layer, "ADBE Vector Composite Order", 2.0)
}

func TestCompileToFileSetsShapeDetail(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Position = []float64{12, -6}
	rec.Comps[0].Layers[1].Shape.Roundness = ptr(18)
	rec.Comps[0].Layers[1].Shape.FillOpacity = ptr(45)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	layer := findProfileLayer(t, prof, "Underline")
	assertLayerPropertyValue(t, layer, "ADBE Vector Rect Position", []float64{12, -6})
	assertLayerPropertyValue(t, layer, "ADBE Vector Rect Roundness", 18.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Fill Opacity", 45.0)
}

func TestCompileToFileSetsShapeFillCompositeOrder(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.FillCompositeOrder = "below_previous"
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	layer := findProfileLayer(t, prof, "Underline")
	assertLayerPropertyValue(t, layer, "ADBE Vector Composite Order", 2.0)
}

func TestCompileToFileSetsShapeFillBlendMode(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.FillBlendMode = ptr(3)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	layer := findProfileLayer(t, prof, "Underline")
	assertLayerPropertyValue(t, layer, "ADBE Vector Blend Mode", 3.0)
}

func TestCompileToFileSetsShapeFillRule(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.FillRule = "even_odd"
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	layer := findProfileLayer(t, prof, "Underline")
	assertLayerPropertyValue(t, layer, "ADBE Vector Fill Rule", 2.0)
}

func TestCompileToFileSetsShapeGradientFill(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.FillColor = nil
	rec.Comps[0].Layers[1].Shape.GradientFill = &recipe.GradientFillSpec{
		Type:       "radial",
		StartPoint: []float64{0, 0},
		EndPoint:   []float64{220, 0},
		ColorStops: []recipe.GradientColorStopSpec{
			{Offset: 0, Midpoint: ptr(0.5), Color: []float64{255, 0, 0}},
			{Offset: 1, Midpoint: ptr(0.5), Color: []float64{0, 0, 255}},
		},
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	layer := findProfileLayer(t, prof, "Underline")
	assertLayerPropertyValue(t, layer, "ADBE Vector Grad Type", 2.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Grad Start Pt", []float64{0, 0})
	assertLayerPropertyValue(t, layer, "ADBE Vector Grad End Pt", []float64{220, 0})
}

func TestCompileToFileSetsShapeGradientFillHighlight(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.FillColor = nil
	rec.Comps[0].Layers[1].Shape.GradientFill = &recipe.GradientFillSpec{
		Type:            "radial",
		StartPoint:      []float64{0, 0},
		EndPoint:        []float64{220, 0},
		HighlightLength: ptr(70),
		HighlightAngle:  ptr(35),
		ColorStops: []recipe.GradientColorStopSpec{
			{Offset: 0, Midpoint: ptr(0.5), Color: []float64{255, 0, 0}},
			{Offset: 1, Midpoint: ptr(0.5), Color: []float64{0, 0, 255}},
		},
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	layer := findProfileLayer(t, prof, "Underline")
	assertLayerPropertyValue(t, layer, "ADBE Vector Grad HiLite Length", 70.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Grad HiLite Angle", 35.0)
}

func TestCompileToFileSetsShapeGradientFillAlphaStops(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.FillColor = nil
	rec.Comps[0].Layers[1].Shape.GradientFill = &recipe.GradientFillSpec{
		Type:       "linear",
		StartPoint: []float64{-220, 0},
		EndPoint:   []float64{220, 0},
		ColorStops: []recipe.GradientColorStopSpec{
			{Offset: 0, Midpoint: ptr(0.5), Color: []float64{255, 0, 0}},
			{Offset: 1, Midpoint: ptr(0.5), Color: []float64{0, 0, 255}},
		},
		AlphaStops: []recipe.GradientAlphaStopSpec{
			{Offset: 0, Midpoint: ptr(0.5), Alpha: 1},
			{Offset: 1, Midpoint: ptr(0.5), Alpha: 0.35},
		},
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	fill := findGradientFillNode(t, project.Compositions[0].Layers[1])
	stops := fill.Gradient().AlphaStops
	if len(stops) != 2 {
		t.Fatalf("alpha stops = %d, want 2", len(stops))
	}
	if math.Abs(stops[1].Alpha-0.35) > 1e-9 {
		t.Fatalf("alpha stop 1 alpha = %g, want 0.35", stops[1].Alpha)
	}
}

func TestCompileToFileSetsShapeGradientStroke(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.FillColor = nil
	rec.Comps[0].Layers[1].Shape.GradientStroke = &recipe.GradientStrokeSpec{
		Type:       "radial",
		StartPoint: []float64{0, 0},
		EndPoint:   []float64{240, 0},
		Width:      ptr(18),
		ColorStops: []recipe.GradientColorStopSpec{
			{Offset: 0, Midpoint: ptr(0.5), Color: []float64{255, 0, 0}},
			{Offset: 1, Midpoint: ptr(0.5), Color: []float64{0, 0, 255}},
		},
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	layer := findProfileLayer(t, prof, "Underline")
	assertLayerPropertyValue(t, layer, "ADBE Vector Grad Type", 2.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Grad Start Pt", []float64{0, 0})
	assertLayerPropertyValue(t, layer, "ADBE Vector Grad End Pt", []float64{240, 0})
	assertLayerPropertyValue(t, layer, "ADBE Vector Stroke Width", 18.0)
}

func TestCompileToFileSetsShapeGradientStrokeHighlight(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.FillColor = nil
	rec.Comps[0].Layers[1].Shape.GradientStroke = &recipe.GradientStrokeSpec{
		Type:            "radial",
		StartPoint:      []float64{0, 0},
		EndPoint:        []float64{240, 0},
		HighlightLength: ptr(65),
		HighlightAngle:  ptr(40),
		Width:           ptr(18),
		ColorStops: []recipe.GradientColorStopSpec{
			{Offset: 0, Midpoint: ptr(0.5), Color: []float64{255, 0, 0}},
			{Offset: 1, Midpoint: ptr(0.5), Color: []float64{0, 0, 255}},
		},
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	layer := findProfileLayer(t, prof, "Underline")
	assertLayerPropertyValue(t, layer, "ADBE Vector Grad HiLite Length", 65.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Grad HiLite Angle", 40.0)
}

func TestCompileToFileSetsShapeGradientStrokeStyle(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.FillColor = nil
	rec.Comps[0].Layers[1].Shape.GradientStroke = &recipe.GradientStrokeSpec{
		Type:       "linear",
		StartPoint: []float64{-240, 0},
		EndPoint:   []float64{240, 0},
		Width:      ptr(18),
		LineCap:    "projecting",
		LineJoin:   "bevel",
		MiterLimit: ptr(9),
		ColorStops: []recipe.GradientColorStopSpec{
			{Offset: 0, Midpoint: ptr(0.5), Color: []float64{255, 0, 0}},
			{Offset: 1, Midpoint: ptr(0.5), Color: []float64{0, 0, 255}},
		},
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	layer := findProfileLayer(t, prof, "Underline")
	assertLayerPropertyValue(t, layer, "ADBE Vector Stroke Line Cap", 3.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Stroke Line Join", 3.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Stroke Miter Limit", 9.0)
}

func TestCompileToFileSetsShapeGradientStrokeAlphaStops(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.FillColor = nil
	rec.Comps[0].Layers[1].Shape.GradientStroke = &recipe.GradientStrokeSpec{
		Type:       "linear",
		StartPoint: []float64{-240, 0},
		EndPoint:   []float64{240, 0},
		Width:      ptr(20),
		ColorStops: []recipe.GradientColorStopSpec{
			{Offset: 0, Midpoint: ptr(0.5), Color: []float64{255, 0, 0}},
			{Offset: 1, Midpoint: ptr(0.5), Color: []float64{0, 0, 255}},
		},
		AlphaStops: []recipe.GradientAlphaStopSpec{
			{Offset: 0, Midpoint: ptr(0.5), Alpha: 1},
			{Offset: 1, Midpoint: ptr(0.5), Alpha: 0.25},
		},
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	stroke := findGradientStrokeNode(t, project.Compositions[0].Layers[1])
	stops := stroke.Gradient().AlphaStops
	if len(stops) != 2 {
		t.Fatalf("alpha stops = %d, want 2", len(stops))
	}
	if math.Abs(stops[1].Alpha-0.25) > 1e-9 {
		t.Fatalf("alpha stop 1 alpha = %g, want 0.25", stops[1].Alpha)
	}
}

func TestCompileToFileSetsShapeStar(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Kind = "polygon"
	rec.Comps[0].Layers[1].Shape.Points = ptr(6)
	rec.Comps[0].Layers[1].Shape.Position = []float64{20, -10}
	rec.Comps[0].Layers[1].Shape.Rotation = ptr(30)
	rec.Comps[0].Layers[1].Shape.InnerRadius = ptr(45)
	rec.Comps[0].Layers[1].Shape.OuterRadius = ptr(120)
	rec.Comps[0].Layers[1].Shape.InnerRoundness = ptr(10)
	rec.Comps[0].Layers[1].Shape.OuterRoundness = ptr(20)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	layer := findProfileLayer(t, prof, "Underline")
	assertLayerPropertyValue(t, layer, "ADBE Vector Star Type", 2.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Star Points", 6.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Star Position", []float64{20, -10})
	assertLayerPropertyValue(t, layer, "ADBE Vector Star Rotation", 30.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Star Inner Radius", 45.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Star Outer Radius", 120.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Star Inner Roundess", 10.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Star Outer Roundess", 20.0)
}

func TestCompileToFileSetsShapeTrim(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Trim = &recipe.TrimSpec{
		Start:  ptr(10),
		End:    ptr(85),
		Offset: ptr(15),
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	layer := findProfileLayer(t, prof, "Underline")
	assertLayerPropertyValue(t, layer, "ADBE Vector Trim Start", 10.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Trim End", 85.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Trim Offset", 15.0)
}

func TestCompileToFileSetsShapeRoundCorners(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.RoundCorners = &recipe.RoundCornersSpec{
		Radius: ptr(18),
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	layer := findProfileLayer(t, prof, "Underline")
	assertLayerPropertyValue(t, layer, "ADBE Vector RoundCorner Radius", 18.0)
}

func TestCompileToFileSetsShapeOffsetPaths(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.OffsetPaths = &recipe.OffsetPathsSpec{
		Amount:     ptr(24),
		LineJoin:   "bevel",
		MiterLimit: ptr(2),
		Copies:     ptr(3),
		CopyOffset: ptr(1.5),
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	layer := findProfileLayer(t, prof, "Underline")
	assertLayerPropertyValue(t, layer, "ADBE Vector Offset Amount", 24.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Offset Line Join", 3.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Offset Miter Limit", 2.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Offset Copies", 3.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Offset Copy Offset", 1.5)
}

func TestCompileToFileSetsShapeRepeater(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Repeater = &recipe.RepeaterSpec{
		Copies:       ptr(5),
		Offset:       ptr(2),
		Order:        "above",
		Anchor:       []float64{15, 25},
		Position:     []float64{120, 0},
		Scale:        []float64{80, 90},
		Rotation:     ptr(30),
		StartOpacity: ptr(100),
		EndOpacity:   ptr(25),
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	layer := findProfileLayer(t, prof, "Underline")
	assertLayerPropertyValue(t, layer, "ADBE Vector Repeater Copies", 5.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Repeater Offset", 2.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Repeater Order", 2.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Repeater Anchor", []float64{15, 25})
	assertLayerPropertyValue(t, layer, "ADBE Vector Repeater Position", []float64{120, 0})
	assertLayerPropertyValue(t, layer, "ADBE Vector Repeater Scale", []float64{80, 90})
	assertLayerPropertyValue(t, layer, "ADBE Vector Repeater Rotation", 30.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Repeater Opacity 1", 100.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Repeater Opacity 2", 25.0)
}

func TestCompileToFileSetsShapeMergePaths(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.MergePaths = &recipe.MergePathsSpec{
		Type: "subtract",
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	layer := findProfileLayer(t, prof, "Underline")
	assertLayerPropertyValue(t, layer, "ADBE Vector Merge Type", 3.0)
}

func TestCompileToFileSetsShapeZigZag(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.ZigZag = &recipe.ZigZagSpec{
		Size:   ptr(40),
		Detail: ptr(8),
		Points: "smooth",
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	layer := findProfileLayer(t, prof, "Underline")
	assertLayerPropertyValue(t, layer, "ADBE Vector Zigzag Size", 40.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Zigzag Detail", 8.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Zigzag Points", 2.0)
}

func TestCompileToFileSetsShapePuckerBloat(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.PuckerBloat = &recipe.PuckerBloatSpec{
		Amount: ptr(100),
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	layer := findProfileLayer(t, prof, "Underline")
	assertLayerPropertyValue(t, layer, "ADBE Vector PuckerBloat Amount", 100.0)
}

func TestCompileToFileSetsShapeTwist(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Twist = &recipe.TwistSpec{
		Angle:  ptr(150),
		Center: []float64{24, -12},
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	layer := findProfileLayer(t, prof, "Underline")
	assertLayerPropertyValue(t, layer, "ADBE Vector Twist Angle", 150.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Twist Center", []float64{24, -12})
}

func TestCompileToFileSetsShapeWigglePaths(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.WigglePaths = &recipe.WigglePathsSpec{
		Size:             ptr(60),
		Detail:           ptr(30),
		WigglesPerSecond: ptr(4),
		RandomSeed:       ptr(9),
		Points:           "smooth",
		Correlation:      ptr(80),
		TemporalPhase:    ptr(45),
		SpatialPhase:     ptr(20),
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	layer := findProfileLayer(t, prof, "Underline")
	assertLayerPropertyValue(t, layer, "ADBE Vector Roughen Size", 60.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Roughen Detail", 30.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Temporal Freq", 4.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Random Seed", 9.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Roughen Points", 2.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Correlation", 80.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Temporal Phase", 45.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Spatial Phase", 20.0)
}

func TestCompileToFileSetsShapeWiggleTransform(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.WiggleTransform = &recipe.WiggleTransformSpec{
		Anchor:           []float64{10, 12},
		Position:         []float64{80, 60},
		Scale:            []float64{20, 30},
		Rotation:         ptr(25),
		WigglesPerSecond: ptr(4),
		RandomSeed:       ptr(9),
		Correlation:      ptr(80),
		TemporalPhase:    ptr(45),
		SpatialPhase:     ptr(20),
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	layer := findProfileLayer(t, prof, "Underline")
	assertLayerPropertyValue(t, layer, "ADBE Vector Wiggler Anchor", []float64{10, 12})
	assertLayerPropertyValue(t, layer, "ADBE Vector Wiggler Position", []float64{80, 60})
	assertLayerPropertyValue(t, layer, "ADBE Vector Wiggler Scale", []float64{20, 30})
	assertLayerPropertyValue(t, layer, "ADBE Vector Wiggler Rotation", 25.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Xform Temporal Freq", 4.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Random Seed", 9.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Correlation", 80.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Temporal Phase", 45.0)
	assertLayerPropertyValue(t, layer, "ADBE Vector Spatial Phase", 20.0)
}

func TestCompileToFileSetsTextStyle(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].TextStyle = &recipe.TextStyleSpec{
		FontSize:      ptr(96),
		FillColor:     []float64{64, 128, 255, 255},
		Tracking:      ptr(120),
		FauxBold:      boolPtr(true),
		FauxItalic:    boolPtr(true),
		ApplyStroke:   boolPtr(true),
		StrokeColor:   []float64{255, 32, 64, 255},
		StrokeWidth:   ptr(8),
		Justification: "center",
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	layer := findProfileLayer(t, prof, "Title")
	if layer.Text == nil || len(layer.Text.Runs) == 0 || len(layer.Text.Paragraphs) == 0 {
		t.Fatalf("text profile missing runs/paragraphs: %+v", layer.Text)
	}
	if got := layer.Text.Runs[0].FontSize; got != 96 {
		t.Fatalf("FontSize = %v, want 96", got)
	}
	assertFloatArray(t, "FillColor", layer.Text.Runs[0].FillColor[:], []float64{64.0 / 255.0, 128.0 / 255.0, 1, 1})
	if got := layer.Text.Runs[0].Tracking; got != 120 {
		t.Fatalf("Tracking = %v, want 120", got)
	}
	if !layer.Text.Runs[0].FauxBold {
		t.Fatal("FauxBold = false, want true")
	}
	if !layer.Text.Runs[0].FauxItalic {
		t.Fatal("FauxItalic = false, want true")
	}
	if !layer.Text.Runs[0].ApplyStroke {
		t.Fatal("ApplyStroke = false, want true")
	}
	assertFloatArray(t, "StrokeColor", layer.Text.Runs[0].StrokeColor[:], []float64{1, 32.0 / 255.0, 64.0 / 255.0, 1})
	if got := layer.Text.Runs[0].StrokeWidth; got != 8 {
		t.Fatalf("StrokeWidth = %v, want 8", got)
	}
	if got := layer.Text.Paragraphs[0].Justification; got != "Center" {
		t.Fatalf("Justification = %q, want Center", got)
	}
}

func TestCompileToFileCreatesParentDirectory(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "nested", "recipe.aep")

	report, err := recipe.CompileToFile(minimalRecipe(), outPath, recipe.StaticCapabilities{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v", report)
	}
	if _, err := aep.Open(outPath); err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
}

func TestCompileToFileMaterializesSupportedEffects(t *testing.T) {
	effects := aep.SupportedEffects()
	if len(effects) == 0 {
		t.Fatal("SupportedEffects is empty")
	}
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Effects = []recipe.Effect{{MatchName: effects[0]}}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	if got := prof.Comps[0].Layers[0].Effects; len(got) != 1 || got[0].MatchName != effects[0] {
		t.Fatalf("Effects = %+v, want %q", got, effects[0])
	}
}

func TestCompileToFileSetsEffectParams(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Effects = []recipe.Effect{{
		MatchName: "ADBE Gaussian Blur 2",
		Params: []recipe.EffectParam{
			{MatchName: "ADBE Gaussian Blur 2-0001", Value: 25.0},
			{MatchName: "ADBE Gaussian Blur 2-0002", Value: 2.0},
			{MatchName: "ADBE Gaussian Blur 2-0003", Value: true},
		},
	}}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	params := prof.Comps[0].Layers[0].Effects[0].Params
	assertParamValue(t, params, "ADBE Gaussian Blur 2-0001", 25.0)
	assertParamValue(t, params, "ADBE Gaussian Blur 2-0002", 2.0)
	assertParamValue(t, params, "ADBE Gaussian Blur 2-0003", 1.0)
}

func TestCompileToFileSetsEffectParamExpression(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Effect param expression"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 2,
			"background_color": [0, 0, 0],
			"layers": [{
				"type": "text",
				"name": "Title",
				"text": "Expr FX",
				"transform": {"position": [960, 540]},
				"effects": [{
					"match_name": "ADBE Gaussian Blur 2",
					"params": [{
						"match_name": "ADBE Gaussian Blur 2-0001",
						"value": 0,
						"expression": {"source": "time * 40", "enabled": false}
					}]
				}]
			}]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	prof, err := profile.Build(project, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	params := prof.Comps[0].Layers[0].Effects[0].Params
	assertParamExpression(t, params, "ADBE Gaussian Blur 2-0001", "time * 40")

	layer := project.Compositions[0].LayerByName("Title")
	if layer == nil || len(layer.Effects) == 0 {
		t.Fatalf("Title effect missing: %+v", layer)
	}
	param := findEffectParam(t, layer.Effects[0], "ADBE Gaussian Blur 2-0001")
	if param == nil {
		t.Fatal("effect param ADBE Gaussian Blur 2-0001 = nil")
	}
	if param.ExpressionEnabled {
		t.Fatal("effect param expression enabled = true, want false")
	}
}

func TestCompileToFileExposesEffectParamAsEssentialGraphicsController(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "EG controller"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [{
				"type": "text",
				"name": "Title",
				"text": "EG",
				"transform": {"position": [960, 540]},
				"effects": [{
					"match_name": "ADBE Slider Control",
					"params": [{
						"match_name": "ADBE Slider Control-0001",
						"value": 42,
						"essential_graphics": {"name": "Amount"}
					}]
				}]
			}]
		}],
		"expected_profile": {
			"essential_graphics": [{
				"name": "Amount",
				"type": "slider"
			}]
		}
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	controllers := project.Compositions[0].EssentialGraphicsControllers
	if len(controllers) != 1 {
		t.Fatalf("EssentialGraphicsControllers = %d, want 1", len(controllers))
	}
	if got := controllers[0].Name; got != "Amount" {
		t.Fatalf("EG controller name = %q, want Amount", got)
	}
	if got := controllers[0].Type.String(); got != "slider" {
		t.Fatalf("EG controller type = %q, want slider", got)
	}
	assertProfileCheck(t, report, "expected_profile.essential_graphics[0].name", true)
	assertProfileCheck(t, report, "expected_profile.essential_graphics[0].type", true)
}

func TestCompileToFileChecksEffectParamKeyframesProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-effect-param-keyframes.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[0]", true)
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[1]", true)
}

func TestCompileToFileChecksEffectParamVectorKeyframesProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-effect-param-vector-keyframes.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[0]", true)
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[1]", true)
}

func TestCompileToFileChecksEffectLayerParamProfileExample(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "examples", "recipes", "minimal-effect-layer-param.json"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	rec := mustUnmarshalRecipe(t, string(raw))
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.effects[0].params[0].target_layer", true)
}

func TestCompileToFileChecksExpectedProfile(t *testing.T) {
	rec := minimalRecipe()
	rec.ExpectedProfile = recipe.ExpectedProfile{
		CompCount:       intPtr(1),
		LayerCount:      intPtr(2),
		TextLayerCount:  intPtr(1),
		ShapeLayerCount: intPtr(1),
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.layer_count", true)
}

func TestCompileToFileChecksProjectBitsPerChannelProfile(t *testing.T) {
	rec := minimalRecipe()
	rec.Project.BitsPerChannel = "32"
	rec.ExpectedProfile = recipe.ExpectedProfile{
		BitsPerChannel: "32",
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.bits_per_channel", true)
}

func TestCompileToFileChecksProjectLinearColorProfile(t *testing.T) {
	rec := minimalRecipe()
	enabled := true
	rec.Project.LinearBlending = &enabled
	rec.Project.LinearizeWorkingSpace = &enabled
	rec.ExpectedProfile = recipe.ExpectedProfile{
		CompCount:             intPtr(1),
		LinearBlending:        &enabled,
		LinearizeWorkingSpace: &enabled,
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.linear_blending", true)
	assertProfileCheck(t, report, "expected_profile.linearize_working_space", true)
}

func TestCompileToFileChecksProjectDisplayProfile(t *testing.T) {
	rec := minimalRecipe()
	enabled := true
	rec.Project.TimeDisplayType = "frames"
	rec.Project.FramesCountType = "start_1"
	rec.Project.FramesUseFeetFrames = &enabled
	rec.Project.FeetFramesFilmType = "16mm"
	rec.Project.FootageTimecodeDisplayStartType = "source_media"
	rec.ExpectedProfile = recipe.ExpectedProfile{
		CompCount:                       intPtr(1),
		TimeDisplayType:                 "frames",
		FramesCountType:                 "start_1",
		FramesUseFeetFrames:             &enabled,
		FeetFramesFilmType:              "16mm",
		FootageTimecodeDisplayStartType: "source_media",
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.time_display_type", true)
	assertProfileCheck(t, report, "expected_profile.frames_count_type", true)
	assertProfileCheck(t, report, "expected_profile.frames_use_feet_frames", true)
	assertProfileCheck(t, report, "expected_profile.feet_frames_film_type", true)
	assertProfileCheck(t, report, "expected_profile.footage_timecode_display_start_type", true)
}

func TestCompileToFileChecksProjectPreferenceProfile(t *testing.T) {
	rec := minimalRecipe()
	disabled := false
	rec.Project.ExpressionEngine = "javascript-1.0"
	rec.Project.AudioSampleRate = floatPtr(44100)
	rec.Project.WorkingGamma = floatPtr(2.4)
	rec.Project.CompensateForSceneReferredProfiles = &disabled
	rec.ExpectedProfile = recipe.ExpectedProfile{
		CompCount:                          intPtr(1),
		ExpressionEngine:                   "javascript-1.0",
		AudioSampleRate:                    floatPtr(44100),
		WorkingGamma:                       floatPtr(2.4),
		CompensateForSceneReferredProfiles: &disabled,
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.expression_engine", true)
	assertProfileCheck(t, report, "expected_profile.audio_sample_rate", true)
	assertProfileCheck(t, report, "expected_profile.working_gamma", true)
	assertProfileCheck(t, report, "expected_profile.compensate_for_scene_referred_profiles", true)
}

func TestCompileToFileChecksProjectDisplayScalarProfile(t *testing.T) {
	rec := minimalRecipe()
	enabled := true
	rec.Project.TimecodeDefaultBase = intPtr(24)
	rec.Project.TransparencyGridThumbnails = &enabled
	rec.ExpectedProfile = recipe.ExpectedProfile{
		CompCount:                  intPtr(1),
		TimecodeDefaultBase:        intPtr(24),
		TransparencyGridThumbnails: &enabled,
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.timecode_default_base", true)
	assertProfileCheck(t, report, "expected_profile.transparency_grid_thumbnails", true)
}

func TestCompileToFileRefusesExpectedProfileMismatch(t *testing.T) {
	rec := minimalRecipe()
	rec.ExpectedProfile = recipe.ExpectedProfile{
		LayerCount: intPtr(3),
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if report.Valid {
		t.Fatalf("report = %+v, want invalid", report)
	}
	assertRefusal(t, report, "profile_contract_mismatch")
	assertProfileCheck(t, report, "expected_profile.layer_count", false)
	if _, err := aep.Open(outPath); err == nil {
		t.Fatal("AEP was written despite expected-profile mismatch")
	}
}

func TestCompileToFileChecksExpectedEffectParamProfile(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Effects = []recipe.Effect{{
		MatchName: "ADBE Gaussian Blur 2",
		Params: []recipe.EffectParam{
			{MatchName: "ADBE Gaussian Blur 2-0001", Value: 25.0},
		},
	}}
	rec.ExpectedProfile = recipe.ExpectedProfile{
		Effects: []recipe.ExpectedEffect{{
			LayerName: "Title",
			MatchName: "ADBE Gaussian Blur 2",
			Params: []recipe.ExpectedEffectParam{{
				MatchName: "ADBE Gaussian Blur 2-0001",
				Value:     25.0,
			}},
		}},
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.effects[0].params[0]", true)
}

func TestCompileToFileChecksExpectedEffectParamExpressionProfile(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Expected effect param expression"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 2,
			"background_color": [0, 0, 0],
			"layers": [{
				"type": "text",
				"name": "Title",
				"text": "Expr FX",
				"transform": {"position": [960, 540]},
				"effects": [{
					"match_name": "ADBE Gaussian Blur 2",
					"params": [{
						"match_name": "ADBE Gaussian Blur 2-0001",
						"value": 0,
						"expression": {"source": "time * 40", "enabled": false}
					}]
				}]
			}]
		}],
		"expected_profile": {
			"effects": [{
				"layer_name": "Title",
				"match_name": "ADBE Gaussian Blur 2",
				"params": [{
					"match_name": "ADBE Gaussian Blur 2-0001",
					"expression": "time * 40",
					"expression_enabled": false
				}]
			}]
		}
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.effects[0].params[0].expression", true)
	assertProfileCheck(t, report, "expected_profile.effects[0].params[0].expression_enabled", true)
}

func TestCompileToFileChecksExpectedLayerPropertyProfile(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[1].Shape.Stroke = &recipe.StrokeSpec{
		Color: []float64{255, 0, 0, 255},
		Width: ptr(6),
	}
	rec.ExpectedProfile = recipe.ExpectedProfile{
		Properties: []recipe.ExpectedProperty{
			{LayerName: "Underline", MatchName: "ADBE Vector Stroke Color", Value: []float64{255, 255, 0, 0}},
			{LayerName: "Underline", MatchName: "ADBE Vector Stroke Width", Value: 6.0},
		},
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.properties[0]", true)
	assertProfileCheck(t, report, "expected_profile.properties[1]", true)
}

func TestCompileToFileChecksExpectedTextStyleProfile(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].TextStyle = &recipe.TextStyleSpec{
		FontSize:           ptr(96),
		FillColor:          []float64{64, 128, 255, 255},
		AutoLeading:        boolPtr(false),
		Leading:            ptr(110),
		Tracking:           ptr(120),
		BaselineShift:      ptr(12),
		HorizontalScale:    ptr(80),
		VerticalScale:      ptr(120),
		Tsume:              ptr(50),
		CapsOption:         "all_caps",
		BaselineOption:     "superscript",
		AutoKernType:       "optical",
		LineJoinType:       "round",
		DigitSet:           "hindi",
		StrokeOverFill:     boolPtr(false),
		NoBreak:            boolPtr(true),
		FauxBold:           boolPtr(true),
		FauxItalic:         boolPtr(true),
		ApplyStroke:        boolPtr(true),
		StrokeColor:        []float64{255, 32, 64, 255},
		StrokeWidth:        ptr(8),
		Justification:      "center",
		FirstLineIndent:    ptr(12),
		StartIndent:        ptr(24),
		EndIndent:          ptr(6),
		SpaceBefore:        ptr(8),
		SpaceAfter:         ptr(10),
		AutoHyphenate:      boolPtr(false),
		LeadingType:        "japanese",
		HangingRoman:       boolPtr(true),
		ParagraphDirection: "rtl",
	}
	rec.ExpectedProfile = recipe.ExpectedProfile{
		TextStyles: []recipe.ExpectedTextStyle{{
			LayerName:          "Title",
			RunIndex:           0,
			ParagraphIndex:     0,
			FontSize:           ptr(96),
			FillColor:          []float64{64.0 / 255.0, 128.0 / 255.0, 1, 1},
			AutoLeading:        boolPtr(false),
			Leading:            ptr(110),
			Tracking:           ptr(120),
			BaselineShift:      ptr(12),
			HorizontalScale:    ptr(80),
			VerticalScale:      ptr(120),
			Tsume:              ptr(50),
			CapsOption:         "all_caps",
			BaselineOption:     "superscript",
			AutoKernType:       "optical",
			LineJoinType:       "round",
			DigitSet:           "hindi",
			StrokeOverFill:     boolPtr(false),
			NoBreak:            boolPtr(true),
			FauxBold:           boolPtr(true),
			FauxItalic:         boolPtr(true),
			ApplyStroke:        boolPtr(true),
			StrokeColor:        []float64{1, 32.0 / 255.0, 64.0 / 255.0, 1},
			StrokeWidth:        ptr(8),
			Justification:      "center",
			FirstLineIndent:    ptr(12),
			StartIndent:        ptr(24),
			EndIndent:          ptr(6),
			SpaceBefore:        ptr(8),
			SpaceAfter:         ptr(10),
			AutoHyphenate:      boolPtr(false),
			LeadingType:        "japanese",
			HangingRoman:       boolPtr(true),
			ParagraphDirection: "rtl",
		}},
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.text_styles[0].font_size", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].fill_color", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].auto_leading", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].leading", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].tracking", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].baseline_shift", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].horizontal_scale", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].vertical_scale", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].tsume", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].caps_option", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].baseline_option", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].auto_kern_type", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].line_join_type", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].digit_set", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].stroke_over_fill", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].no_break", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].faux_bold", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].faux_italic", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].apply_stroke", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].stroke_color", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].stroke_width", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].justification", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].first_line_indent", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].start_indent", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].end_indent", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].space_before", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].space_after", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].auto_hyphenate", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].leading_type", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].hanging_roman", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].paragraph_direction", true)
}

func TestCompileToFileChecksExpectedKeyframeProfile(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.Position = nil
	rec.Comps[0].Layers[0].Transform.PositionKeyframes = []recipe.VectorKeyframe{
		{Time: 0, Value: []float64{900, 540}},
		{Time: 1, Value: []float64{1020, 540}},
	}
	rec.ExpectedProfile = recipe.ExpectedProfile{
		Keyframes: []recipe.ExpectedKeyframedProperty{{
			LayerName: "Title",
			MatchName: "ADBE Position",
			Keyframes: []recipe.ExpectedKeyframe{
				{Time: 0, Value: []float64{900, 540, 0}},
				{Time: 1, Value: []float64{1020, 540, 0}},
			},
		}},
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[0]", true)
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[1]", true)
}

func TestCompileToFileChecksExpectedMaskProfile(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Mask profile"},
		"comps": [{
			"name": "Main",
			"width": 200,
			"height": 200,
			"frame_rate": 30,
			"duration": 1,
			"background_color": [0, 0, 0],
			"layers": [{
				"type": "shape",
				"name": "Masked Shape",
				"shape": {
					"kind": "rect",
					"size": [160, 160],
					"fill_color": [255, 255, 255]
				},
				"masks": [{
					"name": "Window",
					"mode": "subtract",
					"inverted": true,
					"locked": true,
					"color": [255, 128, 0],
					"motion_blur": "off",
					"feather_falloff": "linear",
					"opacity": 0.5,
					"feather": [12, 8],
					"expansion": -4,
					"closed": true,
					"vertices": [[10, 10], [190, 10], [190, 190], [10, 190]],
					"path_keyframes": [
						{"time": 0, "vertices": [[10, 10], [190, 10], [190, 190], [10, 190]]},
						{"time": 1, "vertices": [[40, 40], [160, 20], [180, 160], [20, 180]]}
					]
				}]
			}]
		}],
		"expected_profile": {
			"comp_count": 1,
			"layer_count": 1,
			"shape_layer_count": 1,
			"layers": [
				{"name": "Masked Shape", "type": "shape"}
			],
			"masks": [{
				"layer_name": "Masked Shape",
				"name": "Window",
				"mode": "subtract",
				"inverted": true,
				"locked": true,
				"color": [255, 128, 0],
				"motion_blur": "off",
				"feather_falloff": "linear",
				"opacity": 0.5,
				"feather": [12, 8],
				"expansion": -4,
				"closed": true,
				"vertex_count": 4,
				"path_keyframes": [
					{"time": 0, "vertex_count": 4},
					{"time": 1, "vertex_count": 4}
				]
			}]
		}
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.masks[0]", true)
	assertProfileCheck(t, report, "expected_profile.masks[0].mode", true)
	assertProfileCheck(t, report, "expected_profile.masks[0].inverted", true)
	assertProfileCheck(t, report, "expected_profile.masks[0].locked", true)
	assertProfileCheck(t, report, "expected_profile.masks[0].color", true)
	assertProfileCheck(t, report, "expected_profile.masks[0].motion_blur", true)
	assertProfileCheck(t, report, "expected_profile.masks[0].feather_falloff", true)
	assertProfileCheck(t, report, "expected_profile.masks[0].opacity", true)
	assertProfileCheck(t, report, "expected_profile.masks[0].feather", true)
	assertProfileCheck(t, report, "expected_profile.masks[0].expansion", true)
	assertProfileCheck(t, report, "expected_profile.masks[0].vertex_count", true)
	assertProfileCheck(t, report, "expected_profile.masks[0].path_keyframes.count", true)
	assertProfileCheck(t, report, "expected_profile.masks[0].path_keyframes[0]", true)
	assertProfileCheck(t, report, "expected_profile.masks[0].path_keyframes[1]", true)
}

func TestCompileToFileSetsTransformKeyframeEase(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.Position = nil
	rec.Comps[0].Layers[0].Transform.Opacity = nil
	rec.Comps[0].Layers[0].Transform.PositionKeyframes = []recipe.VectorKeyframe{
		{
			Time:    0,
			Value:   []float64{900, 540},
			OutEase: &recipe.TemporalEase{Speed: 0, Influence: 0.8},
		},
		{
			Time:   1,
			Value:  []float64{1020, 540},
			InEase: &recipe.TemporalEase{Speed: 0, Influence: 0.35},
		},
	}
	rec.Comps[0].Layers[0].Transform.OpacityKeyframes = []recipe.ScalarKeyframe{
		{
			Time:    0,
			Value:   20,
			OutEase: &recipe.TemporalEase{Speed: 2.5, Influence: 0.6},
		},
		{
			Time:   1,
			Value:  100,
			InEase: &recipe.TemporalEase{Speed: 1.5, Influence: 0.25},
		},
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	layer := project.Compositions[0].Layers[0]
	position := layer.Position()
	if position == nil || len(position.Keyframes) != 2 {
		t.Fatalf("Position keyframes = %+v, want 2", position)
	}
	if position.Keyframes[0].OutInterp != aep.InterpBezier {
		t.Fatalf("position kf0 OutInterp = %s, want bezier", position.Keyframes[0].OutInterp)
	}
	if len(position.Keyframes[0].OutTemporalEase) != 1 {
		t.Fatalf("position kf0 OutTemporalEase length = %d, want 1", len(position.Keyframes[0].OutTemporalEase))
	}
	if got := position.Keyframes[0].OutTemporalEase[0].Influence; math.Abs(got-0.8) > 1e-9 {
		t.Fatalf("position kf0 out influence = %g, want 0.8", got)
	}
	if position.Keyframes[1].InInterp != aep.InterpBezier {
		t.Fatalf("position kf1 InInterp = %s, want bezier", position.Keyframes[1].InInterp)
	}
	if len(position.Keyframes[1].InTemporalEase) != 1 {
		t.Fatalf("position kf1 InTemporalEase length = %d, want 1", len(position.Keyframes[1].InTemporalEase))
	}
	if got := position.Keyframes[1].InTemporalEase[0].Influence; math.Abs(got-0.35) > 1e-9 {
		t.Fatalf("position kf1 in influence = %g, want 0.35", got)
	}

	opacity := layer.Opacity()
	if opacity == nil || len(opacity.Keyframes) != 2 {
		t.Fatalf("Opacity keyframes = %+v, want 2", opacity)
	}
	if opacity.Keyframes[0].OutInterp != aep.InterpBezier {
		t.Fatalf("opacity kf0 OutInterp = %s, want bezier", opacity.Keyframes[0].OutInterp)
	}
	if len(opacity.Keyframes[0].OutTemporalEase) != 1 {
		t.Fatalf("opacity kf0 OutTemporalEase length = %d, want 1", len(opacity.Keyframes[0].OutTemporalEase))
	}
	if got := opacity.Keyframes[0].OutTemporalEase[0]; math.Abs(got.Speed-2.5) > 1e-9 || math.Abs(got.Influence-0.6) > 1e-9 {
		t.Fatalf("opacity kf0 out ease = %+v, want speed=2.5 influence=0.6", got)
	}
	if opacity.Keyframes[1].InInterp != aep.InterpBezier {
		t.Fatalf("opacity kf1 InInterp = %s, want bezier", opacity.Keyframes[1].InInterp)
	}
	if len(opacity.Keyframes[1].InTemporalEase) != 1 {
		t.Fatalf("opacity kf1 InTemporalEase length = %d, want 1", len(opacity.Keyframes[1].InTemporalEase))
	}
	if got := opacity.Keyframes[1].InTemporalEase[0]; math.Abs(got.Speed-1.5) > 1e-9 || math.Abs(got.Influence-0.25) > 1e-9 {
		t.Fatalf("opacity kf1 in ease = %+v, want speed=1.5 influence=0.25", got)
	}
}

func TestCompileToFileSetsTransformExpressions(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Transform expressions"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 2,
			"background_color": [0, 0, 0],
			"layers": [{
				"type": "text",
				"name": "Title",
				"text": "Expr",
				"transform": {
					"position": [960, 540],
					"opacity": 100,
					"expressions": {
						"position": {"source": "[value[0] + time * 10, value[1], value[2]]"},
						"opacity": {"source": "time * 50", "enabled": false}
					}
				}
			}]
		}]
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	project, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open compiled AEP: %v", err)
	}
	layer := project.Compositions[0].LayerByName("Title")
	if layer == nil {
		t.Fatal("layer Title not found")
	}
	position := layer.Position()
	if position == nil {
		t.Fatal("Position property = nil")
	}
	if got := position.Expression; got != "[value[0] + time * 10, value[1], value[2]]" {
		t.Fatalf("position expression = %q, want source", got)
	}
	if !position.ExpressionEnabled {
		t.Fatal("position expression enabled = false, want true default")
	}
	opacity := layer.Opacity()
	if opacity == nil {
		t.Fatal("Opacity property = nil")
	}
	if got := opacity.Expression; got != "time * 50" {
		t.Fatalf("opacity expression = %q, want time * 50", got)
	}
	if opacity.ExpressionEnabled {
		t.Fatal("opacity expression enabled = true, want false")
	}
}

func TestCompileToFileChecksExpectedPropertyExpressionProfile(t *testing.T) {
	rec := mustUnmarshalRecipe(t, `{
		"schema_version": 1,
		"project": {"name": "Expected expression"},
		"comps": [{
			"name": "Main",
			"width": 1920,
			"height": 1080,
			"frame_rate": 30,
			"duration": 2,
			"background_color": [0, 0, 0],
			"layers": [{
				"type": "text",
				"name": "Title",
				"text": "Expr",
				"transform": {
					"position": [960, 540],
					"expressions": {
						"position": {"source": "[value[0] + time * 10, value[1], value[2]]"}
					}
				}
			}]
		}],
		"expected_profile": {
			"properties": [{
				"layer_name": "Title",
				"match_name": "ADBE Position",
				"expression": "[value[0] + time * 10, value[1], value[2]]",
				"expression_enabled": true
			}]
		}
	}`)
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.properties[0].expression", true)
	assertProfileCheck(t, report, "expected_profile.properties[0].expression_enabled", true)
}

func TestCompileToFileChecksExpectedOpacityKeyframeProfile(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.Opacity = nil
	rec.Comps[0].Layers[0].Transform.OpacityKeyframes = []recipe.ScalarKeyframe{
		{Time: 0, Value: 20},
		{Time: 1, Value: 100},
		{Time: 2, Value: 40},
	}
	rec.ExpectedProfile = recipe.ExpectedProfile{
		Keyframes: []recipe.ExpectedKeyframedProperty{{
			LayerName: "Title",
			MatchName: "ADBE Opacity",
			Keyframes: []recipe.ExpectedKeyframe{
				{Time: 0, Value: 0.2},
				{Time: 1, Value: 1.0},
				{Time: 2, Value: 0.4},
			},
		}},
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[0]", true)
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[1]", true)
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[2]", true)
}

func TestCompileToFileChecksExpectedScaleKeyframeProfile(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.Scale = nil
	rec.Comps[0].Layers[0].Transform.ScaleKeyframes = []recipe.VectorKeyframe{
		{Time: 0, Value: []float64{80, 80}},
		{Time: 1, Value: []float64{100, 120}},
		{Time: 2, Value: []float64{130, 90}},
	}
	rec.ExpectedProfile = recipe.ExpectedProfile{
		Keyframes: []recipe.ExpectedKeyframedProperty{{
			LayerName: "Title",
			MatchName: "ADBE Scale",
			Keyframes: []recipe.ExpectedKeyframe{
				{Time: 0, Value: []float64{0.8, 0.8, 1}},
				{Time: 1, Value: []float64{1, 1.2, 1}},
				{Time: 2, Value: []float64{1.3, 0.9, 1}},
			},
		}},
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[0]", true)
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[1]", true)
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[2]", true)
}

func TestCompileToFileChecksExpectedRotationKeyframeProfile(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.Rotation = nil
	rec.Comps[0].Layers[0].Transform.RotationKeyframes = []recipe.ScalarKeyframe{
		{Time: 0, Value: 0},
		{Time: 1, Value: 45},
		{Time: 2, Value: -30},
	}
	rec.ExpectedProfile = recipe.ExpectedProfile{
		Keyframes: []recipe.ExpectedKeyframedProperty{{
			LayerName: "Title",
			MatchName: "ADBE Rotate Z",
			Keyframes: []recipe.ExpectedKeyframe{
				{Time: 0, Value: 0.0},
				{Time: 1, Value: 45.0},
				{Time: 2, Value: -30.0},
			},
		}},
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[0]", true)
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[1]", true)
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[2]", true)
}

func TestCompileToFileChecksExpectedAnchorPointKeyframeProfile(t *testing.T) {
	rec := minimalRecipe()
	rec.Comps[0].Layers[0].Transform.AnchorPoint = nil
	rec.Comps[0].Layers[0].Transform.AnchorPointKeyframes = []recipe.VectorKeyframe{
		{Time: 0, Value: []float64{0, 0}},
		{Time: 1, Value: []float64{120, -40}},
		{Time: 2, Value: []float64{-60, 30}},
	}
	rec.ExpectedProfile = recipe.ExpectedProfile{
		Keyframes: []recipe.ExpectedKeyframedProperty{{
			LayerName: "Title",
			MatchName: "ADBE Anchor Point",
			Keyframes: []recipe.ExpectedKeyframe{
				{Time: 0, Value: []float64{0, 0, 0}},
				{Time: 1, Value: []float64{120, -40, 0}},
				{Time: 2, Value: []float64{-60, 30, 0}},
			},
		}},
	}
	outPath := filepath.Join(t.TempDir(), "recipe.aep")

	report, err := recipe.CompileToFile(rec, outPath, stableCapabilityIndex{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report = %+v, want valid", report)
	}
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[0]", true)
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[1]", true)
	assertProfileCheck(t, report, "expected_profile.keyframes[0].keyframes[2]", true)
}

func assertParamValue(t *testing.T, params []profile.Property, matchName string, want float64) {
	t.Helper()
	for _, param := range params {
		if param.MatchName != matchName {
			continue
		}
		got, ok := param.StaticValue.(float64)
		if !ok || got != want {
			t.Fatalf("%s StaticValue = %v, want %v", matchName, param.StaticValue, want)
		}
		return
	}
	t.Fatalf("param %q not found in %+v", matchName, params)
}

func assertParamExpression(t *testing.T, params []profile.Property, matchName string, want string) {
	t.Helper()
	for _, param := range params {
		if param.MatchName != matchName {
			continue
		}
		if param.Expression != want {
			t.Fatalf("%s Expression = %q, want %q", matchName, param.Expression, want)
		}
		return
	}
	t.Fatalf("param %q not found in %+v", matchName, params)
}

func findEffectParam(t *testing.T, effect *aep.Effect, matchName string) *aep.Property {
	t.Helper()
	for _, param := range effect.Parameters {
		if param.MatchName == matchName {
			return param
		}
	}
	t.Fatalf("effect param %q not found in %+v", matchName, effect.Parameters)
	return nil
}

func findProfileLayer(t *testing.T, prof *profile.Profile, name string) profile.Layer {
	t.Helper()
	for _, comp := range prof.Comps {
		for _, layer := range comp.Layers {
			if layer.Name == name {
				return layer
			}
		}
	}
	t.Fatalf("layer %q not found in %+v", name, prof.Comps)
	return profile.Layer{}
}

func findGradientStrokeNode(t *testing.T, layer *aep.Layer) *aep.GradientStrokeNode {
	t.Helper()
	shapeLayer := aep.WrapShapeLayer(layer)
	for _, child := range shapeLayer.RootGroup().Children {
		if stroke, ok := child.(*aep.GradientStrokeNode); ok {
			return stroke
		}
	}
	t.Fatalf("gradient stroke node not found on layer %q", layer.Name)
	return nil
}

func findGradientFillNode(t *testing.T, layer *aep.Layer) *aep.GradientFillNode {
	t.Helper()
	shapeLayer := aep.WrapShapeLayer(layer)
	for _, child := range shapeLayer.RootGroup().Children {
		if fill, ok := child.(*aep.GradientFillNode); ok {
			return fill
		}
	}
	t.Fatalf("gradient fill node not found on layer %q", layer.Name)
	return nil
}

func assertLayerPropertyValue(t *testing.T, layer profile.Layer, matchName string, want any) {
	t.Helper()
	for _, prop := range layer.Properties {
		if prop.MatchName == matchName {
			assertProfileValue(t, matchName, prop.StaticValue, want)
			return
		}
	}
	for _, shape := range layer.Shapes {
		for _, prop := range shape.Properties {
			if prop.MatchName == matchName {
				assertProfileValue(t, matchName, prop.StaticValue, want)
				return
			}
		}
	}
	t.Fatalf("property %q not found on layer %+v", matchName, layer)
}

func assertProfileValue(t *testing.T, label string, got, want any) {
	t.Helper()
	switch want := want.(type) {
	case float64:
		gotFloat, ok := got.(float64)
		if !ok || gotFloat != want {
			t.Fatalf("%s StaticValue = %v, want %v", label, got, want)
		}
	case []float64:
		gotSlice, ok := got.([]float64)
		if !ok || len(gotSlice) != len(want) {
			t.Fatalf("%s StaticValue = %v, want %v", label, got, want)
		}
		for i := range want {
			if gotSlice[i] != want[i] {
				t.Fatalf("%s StaticValue = %v, want %v", label, got, want)
			}
		}
	default:
		t.Fatalf("unsupported want type %T", want)
	}
}

func assertFloatArray(t *testing.T, label string, got, want []float64) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s = %v, want %v", label, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s = %v, want %v", label, got, want)
		}
	}
}

func assertProfileCheck(t *testing.T, report recipe.Report, path string, passed bool) {
	t.Helper()
	for _, check := range report.ProfileChecks {
		if check.Path == path && check.Passed == passed {
			return
		}
	}
	t.Fatalf("profile check %q passed=%v not found in %+v", path, passed, report.ProfileChecks)
}

func intPtr(v int) *int {
	return &v
}

func floatPtr(v float64) *float64 {
	return &v
}
