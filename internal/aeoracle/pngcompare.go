package aeoracle

import (
	"fmt"
	"image"
	"image/png"
	"os"
)

type CompareOptions struct {
	ChannelThreshold uint8 `json:"channel_threshold"`
}

type CompareReport struct {
	SchemaVersion    int     `json:"schema_version"`
	ExpectedPath     string  `json:"expected_path"`
	ActualPath       string  `json:"actual_path"`
	Width            int     `json:"width"`
	Height           int     `json:"height"`
	TotalPixels      int     `json:"total_pixels"`
	DifferentPixels  int     `json:"different_pixels"`
	DifferentPercent float64 `json:"different_percent"`
	MaxChannelDelta  uint8   `json:"max_channel_delta"`
	ChannelThreshold uint8   `json:"channel_threshold"`
}

func ComparePNG(expectedPath, actualPath string, opts CompareOptions) (CompareReport, error) {
	expected, err := readPNG(expectedPath)
	if err != nil {
		return CompareReport{}, fmt.Errorf("aeoracle: read expected png: %w", err)
	}
	actual, err := readPNG(actualPath)
	if err != nil {
		return CompareReport{}, fmt.Errorf("aeoracle: read actual png: %w", err)
	}
	eb, ab := expected.Bounds(), actual.Bounds()
	if eb.Dx() != ab.Dx() || eb.Dy() != ab.Dy() {
		return CompareReport{}, fmt.Errorf("aeoracle: png size mismatch expected=%dx%d actual=%dx%d", eb.Dx(), eb.Dy(), ab.Dx(), ab.Dy())
	}

	report := CompareReport{
		SchemaVersion:    SchemaVersion,
		ExpectedPath:     expectedPath,
		ActualPath:       actualPath,
		Width:            eb.Dx(),
		Height:           eb.Dy(),
		TotalPixels:      eb.Dx() * eb.Dy(),
		ChannelThreshold: opts.ChannelThreshold,
	}
	for y := 0; y < eb.Dy(); y++ {
		for x := 0; x < eb.Dx(); x++ {
			er, eg, ebv, ea := expected.At(eb.Min.X+x, eb.Min.Y+y).RGBA()
			ar, ag, abv, aa := actual.At(ab.Min.X+x, ab.Min.Y+y).RGBA()
			deltas := []uint8{
				delta8(er, ar),
				delta8(eg, ag),
				delta8(ebv, abv),
				delta8(ea, aa),
			}
			pixelDiff := false
			for _, d := range deltas {
				if d > report.MaxChannelDelta {
					report.MaxChannelDelta = d
				}
				if d > opts.ChannelThreshold {
					pixelDiff = true
				}
			}
			if pixelDiff {
				report.DifferentPixels++
			}
		}
	}
	if report.TotalPixels > 0 {
		report.DifferentPercent = float64(report.DifferentPixels) * 100 / float64(report.TotalPixels)
	}
	return report, nil
}

func readPNG(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return png.Decode(f)
}

func delta8(a, b uint32) uint8 {
	av := uint8(a >> 8)
	bv := uint8(b >> 8)
	if av > bv {
		return av - bv
	}
	return bv - av
}
