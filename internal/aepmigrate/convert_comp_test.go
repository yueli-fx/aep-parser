package aepmigrate

import (
	"math"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

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

func TestConvertPreservesNoLayerCompMetadata(t *testing.T) {
	source := writeTempProjectWithMetadataNoLayerComp(t)
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
	if comp.Label != 11 || comp.Comment != "migration note\nsecond line" {
		t.Fatalf("metadata label=%d comment=%q, want label=11 comment preserved", comp.Label, comp.Comment)
	}
}

func TestConvertPreservesNoLayerCompRendererAndTemplateName(t *testing.T) {
	source := writeTempProjectWithRendererNoLayerComp(t)
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
	if comp.Renderer != "ADBE Calder" || comp.MotionGraphicsTemplateName != "Migration Template" {
		t.Fatalf("renderer=%q template=%q, want renderer ADBE Calder and template preserved", comp.Renderer, comp.MotionGraphicsTemplateName)
	}
}
