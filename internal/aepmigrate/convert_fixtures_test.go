package aepmigrate

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

func writeTempProjectWithRendererNoLayerComp(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Renderer", 800, 450, 24, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	mustSetCompSetting(t, "SetRenderer", aep.SetRenderer(comp, "ADBE Calder"))
	mustSetCompSetting(t, "SetMotionGraphicsTemplateName", comp.SetMotionGraphicsTemplateName("Migration Template"))
	path := filepath.Join(t.TempDir(), "renderer-source.aep")
	writeProjectFile(t, project, path)
	return path
}

func writeTempProjectWithMetadataNoLayerComp(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	basePath := filepath.Join(dir, "metadata-base.aep")
	project := aep.NewProject(aep.TargetAE2020)
	if _, err := aep.NewComposition(project, "Metadata", 800, 450, 24, 5); err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	writeProjectFile(t, project, basePath)

	reopened, err := aep.Open(basePath)
	if err != nil {
		t.Fatalf("Open base: %v", err)
	}
	comp := reopened.CompositionByName("Metadata")
	if comp == nil {
		t.Fatal("comp Metadata not found")
	}
	mustSetCompSetting(t, "SetLabel", comp.SetLabel(11))
	mustSetCompSetting(t, "SetComment", comp.SetComment("migration note\nsecond line"))

	path := filepath.Join(dir, "metadata-source.aep")
	writeProjectFile(t, reopened, path)
	return path
}

func writeTempProjectWithConfiguredNoLayerComp(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Configured", 800, 450, 24, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	mustSetCompSetting(t, "SetBGColor", comp.SetBGColor([3]uint8{12, 34, 56}))
	mustSetCompSetting(t, "SetResolutionFactor", comp.SetResolutionFactor(3, 2))
	mustSetCompSetting(t, "SetPixelAspect", comp.SetPixelAspect(1.5))
	mustSetCompSetting(t, "SetDisplayStartTime", comp.SetDisplayStartTime(1.25))
	mustSetCompSetting(t, "SetFrameBlending", comp.SetFrameBlending(true))
	mustSetCompSetting(t, "SetHideShyLayers", comp.SetHideShyLayers(true))
	mustSetCompSetting(t, "SetPreserveNestedFrameRate", comp.SetPreserveNestedFrameRate(true))
	mustSetCompSetting(t, "SetPreserveNestedResolution", comp.SetPreserveNestedResolution(true))
	mustSetCompSetting(t, "SetCompMotionBlur", comp.SetCompMotionBlur(true))
	mustSetCompSetting(t, "SetShutterAngle", comp.SetShutterAngle(270))
	mustSetCompSetting(t, "SetShutterPhase", comp.SetShutterPhase(-45))
	mustSetCompSetting(t, "SetMotionBlurAdaptiveSampleLimit", comp.SetMotionBlurAdaptiveSampleLimit(192))
	mustSetCompSetting(t, "SetMotionBlurSamplesPerFrame", comp.SetMotionBlurSamplesPerFrame(24))
	mustSetCompSetting(t, "SetWorkArea", comp.SetWorkArea(0.5, 4.5))
	path := filepath.Join(t.TempDir(), "configured-comp.aep")
	writeProjectFile(t, project, path)
	return path
}

func writeProjectFile(t *testing.T, project *aep.Project, path string) {
	t.Helper()
	out, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer out.Close()
	if err := project.WriteAEP(out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
}

func mustSetCompSetting(t *testing.T, name string, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
}

func writeTempProjectWithOneComp(t *testing.T, target aep.AETarget) string {
	t.Helper()
	project := aep.NewProject(target)
	if _, err := aep.NewComposition(project, "Main", 640, 360, 24, 2.5); err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	path := filepath.Join(t.TempDir(), "one-comp.aep")
	out, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer out.Close()
	if err := project.WriteAEP(out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	return path
}

func writeTempProjectWithOneSolidLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewSolidLayer(comp, "Solid", 640, 360, [3]float64{1, 0, 0}); err != nil {
		t.Fatalf("NewSolidLayer: %v", err)
	}
	path := filepath.Join(t.TempDir(), "one-layer.aep")
	out, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer out.Close()
	if err := project.WriteAEP(out); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	return path
}

func writeTempProjectWithOneTextLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewTextLayer(comp, "Title"); err != nil {
		t.Fatalf("NewTextLayer: %v", err)
	}
	path := filepath.Join(t.TempDir(), "one-text-layer.aep")
	writeProjectFile(t, project, path)
	return path
}

func writeTempProjectWithDefaultTextLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	layer, err := aep.NewTextLayer(comp, "Title")
	if err != nil {
		t.Fatalf("NewTextLayer: %v", err)
	}
	if err := layer.SetText("Hello"); err != nil {
		t.Fatalf("SetText: %v", err)
	}
	path := filepath.Join(t.TempDir(), "default-text-layer.aep")
	writeProjectFile(t, project, path)
	return path
}

func writeTempProjectWithCommentedTextLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	layer, err := aep.NewTextLayer(comp, "Title")
	if err != nil {
		t.Fatalf("NewTextLayer: %v", err)
	}
	if err := layer.SetComment("unsupported text layer comment"); err != nil {
		t.Fatalf("SetComment: %v", err)
	}
	path := filepath.Join(t.TempDir(), "commented-text-layer.aep")
	writeProjectFile(t, project, path)
	return path
}

func writeTempProjectWithDefaultShapeLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewShapeLayer(comp, "Shape"); err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	path := filepath.Join(t.TempDir(), "default-shape-layer.aep")
	writeProjectFile(t, project, path)
	return path
}

func writeTempProjectWithRectFillShapeLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	shape, err := aep.NewShapeLayer(comp, "Card")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	rect, err := shape.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect: %v", err)
	}
	if err := rect.SetSize([2]float64{320, 180}); err != nil {
		t.Fatalf("Rect.SetSize: %v", err)
	}
	fill, err := shape.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill: %v", err)
	}
	if err := fill.SetColor([4]float64{1, 0.25, 0.5, 1}); err != nil {
		t.Fatalf("Fill.SetColor: %v", err)
	}
	path := filepath.Join(t.TempDir(), "rect-fill-shape-layer.aep")
	writeProjectFile(t, project, path)
	return path
}

func writeTempProjectWithOneAdjustmentLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewAdjustmentLayer(comp, "Grade"); err != nil {
		t.Fatalf("NewAdjustmentLayer: %v", err)
	}
	path := filepath.Join(t.TempDir(), "one-adjustment-layer.aep")
	writeProjectFile(t, project, path)
	return path
}

func writeTempProjectWithOneCameraLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewCameraLayer(comp, "Camera"); err != nil {
		t.Fatalf("NewCameraLayer: %v", err)
	}
	path := filepath.Join(t.TempDir(), "one-camera-layer.aep")
	writeProjectFile(t, project, path)
	return path
}

func writeTempProjectWithOneLightLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewLightLayer(comp, "Light"); err != nil {
		t.Fatalf("NewLightLayer: %v", err)
	}
	path := filepath.Join(t.TempDir(), "one-light-layer.aep")
	writeProjectFile(t, project, path)
	return path
}

func writeTempProjectWithOnePrecompLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	source, err := aep.NewComposition(project, "Source", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition source: %v", err)
	}
	main, err := aep.NewComposition(project, "Main", 1920, 1080, 30, 3)
	if err != nil {
		t.Fatalf("NewComposition main: %v", err)
	}
	if _, err := aep.NewPrecompLayer(main, source, "Nested Source"); err != nil {
		t.Fatalf("NewPrecompLayer: %v", err)
	}
	path := filepath.Join(t.TempDir(), "one-precomp-layer.aep")
	writeProjectFile(t, project, path)
	return path
}

func writeTempProjectWithDefaultNullLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewNullLayer(comp, "Controller"); err != nil {
		t.Fatalf("NewNullLayer: %v", err)
	}
	path := filepath.Join(t.TempDir(), "one-null-layer.aep")
	writeProjectFile(t, project, path)
	return path
}

func writeTempProjectWithMovedNullLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	layer, err := aep.NewNullLayer(comp, "Controller")
	if err != nil {
		t.Fatalf("NewNullLayer: %v", err)
	}
	transform := aep.NewLayerTransform()
	if err := transform.Position().SetStaticValue([2]float64{320, 180}); err != nil {
		t.Fatalf("Position.SetStaticValue: %v", err)
	}
	if err := aep.SetLayerTransform(layer, transform); err != nil {
		t.Fatalf("SetLayerTransform: %v", err)
	}
	path := filepath.Join(t.TempDir(), "moved-null-layer.aep")
	writeProjectFile(t, project, path)
	return path
}

func writeTempProjectWithKeyframedNullLayer(t *testing.T) string {
	t.Helper()
	project := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(project, "Main", 640, 360, 24, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	layer, err := aep.NewNullLayer(comp, "Controller")
	if err != nil {
		t.Fatalf("NewNullLayer: %v", err)
	}
	transform := aep.NewLayerTransform()
	if err := transform.Position().AddKeyframeLinear(0, [2]float64{0, 0}); err != nil {
		t.Fatalf("Position.AddKeyframeLinear 0: %v", err)
	}
	if err := transform.Position().AddKeyframeLinear(1, [2]float64{320, 180}); err != nil {
		t.Fatalf("Position.AddKeyframeLinear 1: %v", err)
	}
	if err := aep.SetLayerTransform(layer, transform); err != nil {
		t.Fatalf("SetLayerTransform: %v", err)
	}
	path := filepath.Join(t.TempDir(), "keyframed-null-layer.aep")
	writeProjectFile(t, project, path)
	return path
}

func writeTempRecipeDefaultNullLayer(t *testing.T) string {
	t.Helper()
	rec := recipe.Recipe{
		SchemaVersion: 1,
		Project:       recipe.ProjectSpec{TargetVersion: "AE2020", Name: "default-null"},
		Comps: []recipe.CompSpec{{
			Name:      "Main",
			Width:     640,
			Height:    360,
			FrameRate: 24,
			Duration:  2,
			Layers: []recipe.Layer{{
				Type: "null",
				Name: "Controller",
			}},
		}},
	}
	path := filepath.Join(t.TempDir(), "recipe-default-null.aep")
	report, err := recipe.CompileToFile(rec, path, recipe.StaticCapabilities{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report.Valid = false, refusals=%+v", report.Refusals)
	}
	return path
}

func writeTempRecipe(t *testing.T, recipePath string) string {
	t.Helper()
	raw, err := os.ReadFile(recipePath)
	if err != nil {
		t.Fatalf("ReadFile recipe: %v", err)
	}
	var rec recipe.Recipe
	if err := json.Unmarshal(raw, &rec); err != nil {
		t.Fatalf("Unmarshal recipe: %v", err)
	}
	path := filepath.Join(t.TempDir(), "recipe-source.aep")
	report, err := recipe.CompileToFile(rec, path, recipe.StaticCapabilities{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report.Valid = false, refusals=%+v", report.Refusals)
	}
	return path
}

func writeTempRecipeWithTarget(t *testing.T, recipePath, targetVersion string) string {
	t.Helper()
	raw, err := os.ReadFile(recipePath)
	if err != nil {
		t.Fatalf("ReadFile recipe: %v", err)
	}
	var rec recipe.Recipe
	if err := json.Unmarshal(raw, &rec); err != nil {
		t.Fatalf("Unmarshal recipe: %v", err)
	}
	rec.Project.TargetVersion = targetVersion
	path := filepath.Join(t.TempDir(), "recipe-source-"+targetVersion+".aep")
	report, err := recipe.CompileToFile(rec, path, recipe.StaticCapabilities{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report.Valid = false, refusals=%+v", report.Refusals)
	}
	return path
}

func assertConvertedLayerMaskProfile(t *testing.T, outPath string) {
	t.Helper()
	converted, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open converted: %v", err)
	}
	prof, err := profile.Build(converted, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("Build converted profile: %v", err)
	}
	layer := profileLayerByName(prof, "Plate")
	if layer == nil || len(layer.Masks) != 1 {
		t.Fatalf("converted masks = %+v, want one mask on Plate", layer)
	}
	mask := layer.Masks[0]
	if mask.Name != "Window" || mask.Mode != "subtract" || !mask.Inverted || !mask.Locked {
		t.Fatalf("mask identity/options = %+v", mask)
	}
	if !profileFloatSlicesEqual(mask.Color, []float64{255, 128, 0}) ||
		mask.MotionBlur != "off" ||
		mask.FeatherFalloff != "linear" ||
		math.Abs(mask.Opacity-0.5) > 1e-9 ||
		!profileFloatSlicesEqual(mask.Feather, []float64{12, 8}) ||
		math.Abs(mask.Expansion-(-4)) > 1e-9 ||
		!mask.Closed ||
		len(mask.Vertices) != 4 ||
		len(mask.PathKeyframes) != 2 {
		t.Fatalf("mask profile = %+v, want preserved options and path keyframes", mask)
	}
}

func profileLayerByName(prof *profile.Profile, name string) *profile.Layer {
	if prof == nil {
		return nil
	}
	for ci := range prof.Comps {
		for li := range prof.Comps[ci].Layers {
			if prof.Comps[ci].Layers[li].Name == name {
				return &prof.Comps[ci].Layers[li]
			}
		}
	}
	return nil
}

func profileFloatSlicesEqual(got, want []float64) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-9 {
			return false
		}
	}
	return true
}
