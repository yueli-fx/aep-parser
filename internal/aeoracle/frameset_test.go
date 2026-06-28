package aeoracle_test

import (
	"encoding/json"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	"github.com/example/aep-parser/internal/aeoracle"
)

func TestCompareFrameSetsReportsAllOKForIdenticalPNGs(t *testing.T) {
	dir := t.TempDir()
	expectedPNG := filepath.Join(dir, "expected.png")
	actualPNG := filepath.Join(dir, "actual.png")
	writeTestPNG(t, expectedPNG, color.RGBA{R: 10, G: 20, B: 30, A: 255})
	writeTestPNG(t, actualPNG, color.RGBA{R: 10, G: 20, B: 30, A: 255})
	expectedMeta := writeRenderMetadata(t, dir, "expected_meta.json", "expected.aep", []aeoracle.RenderedFrameRecord{
		{Frame: 0, Seconds: 0, Tag: "f000000", Reason: "comp_start", OutputPath: expectedPNG, Status: "rendered"},
	})
	actualMeta := writeRenderMetadata(t, dir, "actual_meta.json", "actual.aep", []aeoracle.RenderedFrameRecord{
		{Frame: 0, Seconds: 0, Tag: "f000000", Reason: "comp_start", OutputPath: actualPNG, Status: "rendered"},
	})

	report, err := aeoracle.CompareFrameSets(expectedMeta, actualMeta, aeoracle.CompareOptions{})
	if err != nil {
		t.Fatalf("CompareFrameSets: %v", err)
	}
	if report.Summary.TotalFrames != 1 || report.Summary.OKFrames != 1 || report.Summary.DifferentFrames != 0 {
		t.Fatalf("Summary = %+v", report.Summary)
	}
	if len(report.Frames) != 1 || report.Frames[0].Status != "ok" || report.Frames[0].Compare == nil {
		t.Fatalf("Frames = %+v", report.Frames)
	}
}

func TestCompareFrameSetsReportsDifferentPixels(t *testing.T) {
	dir := t.TempDir()
	expectedPNG := filepath.Join(dir, "expected.png")
	actualPNG := filepath.Join(dir, "actual.png")
	writeTestPNG(t, expectedPNG, color.RGBA{R: 10, G: 20, B: 30, A: 255})
	writeOnePixelDeltaPNG(t, actualPNG, color.RGBA{R: 10, G: 20, B: 30, A: 255}, 1)
	expectedMeta := writeRenderMetadata(t, dir, "expected_meta.json", "expected.aep", []aeoracle.RenderedFrameRecord{
		{Frame: 24, Seconds: 1, Tag: "f000024", Reason: "keyframe", OutputPath: expectedPNG, Status: "rendered"},
	})
	actualMeta := writeRenderMetadata(t, dir, "actual_meta.json", "actual.aep", []aeoracle.RenderedFrameRecord{
		{Frame: 24, Seconds: 1, Tag: "f000024", Reason: "keyframe", OutputPath: actualPNG, Status: "rendered"},
	})

	report, err := aeoracle.CompareFrameSets(expectedMeta, actualMeta, aeoracle.CompareOptions{})
	if err != nil {
		t.Fatalf("CompareFrameSets: %v", err)
	}
	if report.Summary.DifferentFrames != 1 {
		t.Fatalf("DifferentFrames = %d, want 1; summary=%+v", report.Summary.DifferentFrames, report.Summary)
	}
	if report.Frames[0].Status != "different" || report.Frames[0].Compare.DifferentPixels != 1 {
		t.Fatalf("frame = %+v", report.Frames[0])
	}
}

func TestCompareFrameSetsReportsMissingActualFrame(t *testing.T) {
	dir := t.TempDir()
	expectedPNG := filepath.Join(dir, "expected.png")
	writeTestPNG(t, expectedPNG, color.RGBA{R: 10, G: 20, B: 30, A: 255})
	expectedMeta := writeRenderMetadata(t, dir, "expected_meta.json", "expected.aep", []aeoracle.RenderedFrameRecord{
		{Frame: 24, Seconds: 1, Tag: "f000024", Reason: "keyframe", OutputPath: expectedPNG, Status: "rendered"},
	})
	actualMeta := writeRenderMetadata(t, dir, "actual_meta.json", "actual.aep", nil)

	report, err := aeoracle.CompareFrameSets(expectedMeta, actualMeta, aeoracle.CompareOptions{})
	if err != nil {
		t.Fatalf("CompareFrameSets: %v", err)
	}
	if report.Summary.MissingActualFrames != 1 {
		t.Fatalf("MissingActualFrames = %d, want 1; summary=%+v", report.Summary.MissingActualFrames, report.Summary)
	}
	if report.Frames[0].Status != "missing_actual" || report.Frames[0].Compare != nil {
		t.Fatalf("frame = %+v", report.Frames[0])
	}
}

func TestCompareFrameSetsRejectsDuplicateTags(t *testing.T) {
	dir := t.TempDir()
	expectedMeta := writeRenderMetadata(t, dir, "expected_meta.json", "expected.aep", []aeoracle.RenderedFrameRecord{
		{Frame: 0, Seconds: 0, Tag: "f000000", Reason: "comp_start", OutputPath: "a.png", Status: "rendered"},
		{Frame: 0, Seconds: 0, Tag: "f000000", Reason: "comp_start", OutputPath: "b.png", Status: "rendered"},
	})
	actualMeta := writeRenderMetadata(t, dir, "actual_meta.json", "actual.aep", nil)

	_, err := aeoracle.CompareFrameSets(expectedMeta, actualMeta, aeoracle.CompareOptions{})
	if err == nil {
		t.Fatal("CompareFrameSets error = nil, want duplicate tag error")
	}
}

func writeRenderMetadata(t *testing.T, dir, name, aepPath string, frames []aeoracle.RenderedFrameRecord) string {
	t.Helper()
	path := filepath.Join(dir, name)
	meta := aeoracle.RenderMetadata{
		SchemaVersion: aeoracle.SchemaVersion,
		AEPPath:       aepPath,
		CompName:      "Main",
		OutputDir:     dir,
		Status:        "ok",
		Frames:        frames,
	}
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		t.Fatalf("MarshalIndent: %v", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}
