package aepmigrate

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func TestConvertWritesTargetVersionNoLayerProject(t *testing.T) {
	source := writeTempProjectWithOneComp(t, aep.TargetAE2020)
	outPath := filepath.Join(t.TempDir(), "converted.aep")

	report, err := Convert(ConvertOptions{
		InputPath:  source,
		OutputPath: outPath,
		Target:     VersionAE2025,
	})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if report.Summary.Status != StatusPass {
		t.Fatalf("status = %q, entries=%+v", report.Summary.Status, report.Entries)
	}
	if report.Summary.Preserved != 0 || report.Summary.Retargeted == 0 {
		t.Fatalf("summary preserved=%d retargeted=%d, want no preserved placeholder and retargeted entries", report.Summary.Preserved, report.Summary.Retargeted)
	}
	converted, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open converted: %v", err)
	}
	version := NormalizeSourceVersion((&aep.Application{Project: converted}).Version())
	if version.Label != VersionAE2025 {
		t.Fatalf("converted version = %+v, want AE2025", version)
	}
	prof, err := profile.Build(converted, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build converted: %v", err)
	}
	if len(prof.Comps) != 1 {
		t.Fatalf("converted comps = %d, want 1", len(prof.Comps))
	}
	comp := prof.Comps[0]
	if comp.Name != "Main" || comp.Width != 640 || comp.Height != 360 {
		t.Fatalf("converted comp identity = %+v", comp)
	}
	if math.Abs(comp.FrameRate-24) > 0.001 || math.Abs(comp.Duration-2.5) > 0.001 {
		t.Fatalf("converted timing frameRate=%f duration=%f, want 24/2.5", comp.FrameRate, comp.Duration)
	}
}

func TestConvertPreservesNoLayerCompSettings(t *testing.T) {
	source := writeTempProjectWithConfiguredNoLayerComp(t)
	outPath := filepath.Join(t.TempDir(), "converted.aep")

	report, err := Convert(ConvertOptions{
		InputPath:  source,
		OutputPath: outPath,
		Target:     VersionAE2025,
	})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if report.Summary.Status != StatusPass {
		t.Fatalf("status = %q, entries=%+v", report.Summary.Status, report.Entries)
	}
	converted, err := aep.Open(outPath)
	if err != nil {
		t.Fatalf("Open converted: %v", err)
	}
	prof, err := profile.Build(converted, profile.Options{Path: outPath})
	if err != nil {
		t.Fatalf("profile.Build converted: %v", err)
	}
	if len(prof.Comps) != 1 {
		t.Fatalf("converted comps = %d, want 1", len(prof.Comps))
	}
	comp := prof.Comps[0]
	if comp.BackgroundColor != [3]uint8{12, 34, 56} {
		t.Fatalf("background color = %#v, want [12 34 56]", comp.BackgroundColor)
	}
	if comp.ResolutionFactor != [2]uint16{3, 2} {
		t.Fatalf("resolution factor = %#v, want [3 2]", comp.ResolutionFactor)
	}
	if math.Abs(comp.PixelAspect-1.5) > 0.001 {
		t.Fatalf("pixel aspect = %f, want 1.5", comp.PixelAspect)
	}
	if math.Abs(comp.DisplayStartTime-1.25) > 0.001 {
		t.Fatalf("display start time = %f, want 1.25", comp.DisplayStartTime)
	}
	if math.Abs(comp.WorkArea.Start-0.5) > 0.001 || math.Abs(comp.WorkArea.End-4.5) > 0.001 {
		t.Fatalf("work area = %+v, want 0.5..4.5", comp.WorkArea)
	}
	if !comp.FrameBlending || !comp.HideShyLayers || !comp.PreserveNestedFrameRate || !comp.PreserveNestedResolution {
		t.Fatalf("flags not preserved: frame_blending=%t hide_shy=%t preserve_fps=%t preserve_res=%t",
			comp.FrameBlending, comp.HideShyLayers, comp.PreserveNestedFrameRate, comp.PreserveNestedResolution)
	}
	if !comp.MotionBlur.Enabled ||
		comp.MotionBlur.ShutterAngle != 270 ||
		comp.MotionBlur.ShutterPhase != -45 ||
		comp.MotionBlur.AdaptiveSampleLimit != 192 ||
		comp.MotionBlur.SamplesPerFrame != 24 {
		t.Fatalf("motion blur = %+v, want enabled angle=270 phase=-45 adaptive=192 samples=24", comp.MotionBlur)
	}
}

func TestConvertRefusesLayerProjects(t *testing.T) {
	source := writeTempProjectWithOneSolidLayer(t)
	outPath := filepath.Join(t.TempDir(), "converted.aep")

	report, err := Convert(ConvertOptions{
		InputPath:  source,
		OutputPath: outPath,
		Target:     VersionAE2025,
	})
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}
	if report.Summary.Status != StatusBlocked {
		t.Fatalf("status = %q, want blocked; entries=%+v", report.Summary.Status, report.Entries)
	}
	if _, err := os.Stat(outPath); !os.IsNotExist(err) {
		t.Fatalf("output exists or stat failed unexpectedly: %v", err)
	}
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
