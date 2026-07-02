package aepmigrate

import (
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

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
	writeProjectFile(t, project, path)
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
