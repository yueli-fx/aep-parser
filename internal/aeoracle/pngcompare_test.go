package aeoracle_test

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aeoracle"
)

func TestComparePNGIdenticalImages(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.png")
	b := filepath.Join(dir, "b.png")
	writeTestPNG(t, a, color.RGBA{R: 10, G: 20, B: 30, A: 255})
	writeTestPNG(t, b, color.RGBA{R: 10, G: 20, B: 30, A: 255})

	report, err := aeoracle.ComparePNG(a, b, aeoracle.CompareOptions{})
	if err != nil {
		t.Fatalf("ComparePNG: %v", err)
	}
	if report.DifferentPixels != 0 || report.TotalPixels != 4 {
		t.Fatalf("report = %+v", report)
	}
}

func TestComparePNGCountsDifferentPixels(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.png")
	b := filepath.Join(dir, "b.png")
	writeTestPNG(t, a, color.RGBA{R: 10, G: 20, B: 30, A: 255})
	writeOnePixelDeltaPNG(t, b, color.RGBA{R: 10, G: 20, B: 30, A: 255}, 1)

	report, err := aeoracle.ComparePNG(a, b, aeoracle.CompareOptions{})
	if err != nil {
		t.Fatalf("ComparePNG: %v", err)
	}
	if report.DifferentPixels != 1 {
		t.Fatalf("DifferentPixels = %d, want 1; report=%+v", report.DifferentPixels, report)
	}
}

func TestComparePNGToleranceIgnoresSmallDeltas(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.png")
	b := filepath.Join(dir, "b.png")
	writeTestPNG(t, a, color.RGBA{R: 10, G: 20, B: 30, A: 255})
	writeOnePixelDeltaPNG(t, b, color.RGBA{R: 10, G: 20, B: 30, A: 255}, 1)

	report, err := aeoracle.ComparePNG(a, b, aeoracle.CompareOptions{ChannelThreshold: 1})
	if err != nil {
		t.Fatalf("ComparePNG: %v", err)
	}
	if report.DifferentPixels != 0 {
		t.Fatalf("DifferentPixels = %d, want 0; report=%+v", report.DifferentPixels, report)
	}
}

func TestComparePNGRejectsSizeMismatch(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.png")
	b := filepath.Join(dir, "b.png")
	writePNG(t, a, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	writePNG(t, b, image.NewRGBA(image.Rect(0, 0, 3, 2)))

	_, err := aeoracle.ComparePNG(a, b, aeoracle.CompareOptions{})
	if err == nil {
		t.Fatal("ComparePNG error = nil, want size mismatch")
	}
}

func writeTestPNG(t *testing.T, path string, pixel color.RGBA) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.SetRGBA(x, y, pixel)
		}
	}
	writePNG(t, path, img)
}

func writeOnePixelDeltaPNG(t *testing.T, path string, pixel color.RGBA, delta uint8) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			img.SetRGBA(x, y, pixel)
		}
	}
	img.SetRGBA(1, 1, color.RGBA{R: pixel.R, G: pixel.G, B: pixel.B + delta, A: pixel.A})
	writePNG(t, path, img)
}

func writePNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("Encode: %v", err)
	}
}
