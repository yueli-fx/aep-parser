package aepmigrate

import (
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
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

func mustSetCompSetting(t *testing.T, name string, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
}
