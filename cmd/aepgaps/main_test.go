package main

import (
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/example/aep-parser/internal/gapledger"
)

func TestRunRenderReturnsZeroForMatchingPNGs(t *testing.T) {
	dir := t.TempDir()
	expected := filepath.Join(dir, "expected.png")
	actual := filepath.Join(dir, "actual.png")
	writeTestPNG(t, expected, color.RGBA{R: 10, G: 20, B: 30, A: 255})
	writeTestPNG(t, actual, color.RGBA{R: 10, G: 20, B: 30, A: 255})

	code := run([]string{"render", "-expected", expected, "-actual", actual})
	if code != 0 {
		t.Fatalf("run(render matching) = %d, want 0", code)
	}
}

func TestRunRenderReturnsOneAndWritesGapReportForDifferentPNGs(t *testing.T) {
	dir := t.TempDir()
	expected := filepath.Join(dir, "expected.png")
	actual := filepath.Join(dir, "actual.png")
	out := filepath.Join(dir, "gaps.json")
	writeTestPNG(t, expected, color.RGBA{R: 10, G: 20, B: 30, A: 255})
	writeTestPNG(t, actual, color.RGBA{R: 11, G: 20, B: 30, A: 255})

	code := run([]string{"render", "-expected", expected, "-actual", actual, "-out", out})
	if code != 1 {
		t.Fatalf("run(render different) = %d, want 1", code)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var report gapledger.Report
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.GapCount != 1 || len(report.Gaps) != 1 {
		t.Fatalf("gap report count = %d/%d, want 1/1", report.GapCount, len(report.Gaps))
	}
	if report.Gaps[0].Type != gapledger.TypeRenderGap {
		t.Fatalf("gap type = %q, want %q", report.Gaps[0].Type, gapledger.TypeRenderGap)
	}
}

func writeTestPNG(t *testing.T, path string, c color.RGBA) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.SetRGBA(x, y, c)
		}
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := png.Encode(file, img); err != nil {
		t.Fatal(err)
	}
}
