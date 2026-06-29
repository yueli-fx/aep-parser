package recipe_test

import (
	"math"
	"path/filepath"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/profile"
	"github.com/example/aep-parser/internal/recipe"
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
	rec.ExpectedProfile = recipe.ExpectedProfile{
		TextStyles: []recipe.ExpectedTextStyle{{
			LayerName:      "Title",
			RunIndex:       0,
			ParagraphIndex: 0,
			FontSize:       ptr(96),
			FillColor:      []float64{64.0 / 255.0, 128.0 / 255.0, 1, 1},
			Tracking:       ptr(120),
			FauxBold:       boolPtr(true),
			FauxItalic:     boolPtr(true),
			ApplyStroke:    boolPtr(true),
			StrokeColor:    []float64{1, 32.0 / 255.0, 64.0 / 255.0, 1},
			StrokeWidth:    ptr(8),
			Justification:  "center",
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
	assertProfileCheck(t, report, "expected_profile.text_styles[0].tracking", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].faux_bold", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].faux_italic", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].apply_stroke", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].stroke_color", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].stroke_width", true)
	assertProfileCheck(t, report, "expected_profile.text_styles[0].justification", true)
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
