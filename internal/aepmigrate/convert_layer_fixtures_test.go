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
