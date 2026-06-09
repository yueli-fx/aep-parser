package aep_test

import (
	"bytes"
	"fmt"
	"math"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// TestCompositionBGColorRegression locks in the bug-fix: prior to this
// change, Composition.BGColor was declared and JSON-emitted but never
// populated by the parser, so every parsed project reported "#000000"
// regardless of AE's actual BG color setting. The synthetic cdta below
// writes (200, 100, 50) at the verified offsets 0x34/0x35/0x36; a parser
// that forgets to decode bg_color returns [0,0,0] and fails this test.
// TestMarkerNmHdDecode exercises duration + label decode through synthetic
// NmHd payloads. Duration is stored in NmHd @0x08 as a uint32 of 600ths-
// of-a-second; label color at NmHd @0x10 as a raw uint8 index (0..16).
func TestCompositionBGColorRegression(t *testing.T) {
	cdta := buildCdtaWithOpts(1920, 1080, 30, 0, 300, cdtaOpts{
		BGColor:                   [3]uint8{200, 100, 50},
		WorkAreaStartDivisor:      600,
		WorkAreaEndDividend:       0xFFFFFFFF,
		WorkAreaEndDivisor:        600,
		ShutterAngle:              180,
		MotionBlurAdaptiveLimit:   128,
		MotionBlurSamplesPerFrame: 16,
	})
	data := buildAEPWithCustomCdta(cdta, 0, 8000)
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	got := proj.Compositions[0].BGColor
	want := [3]uint8{200, 100, 50}
	if got != want {
		t.Errorf("BGColor = %v, want %v (regression: was always {0,0,0})", got, want)
	}
}

func TestCompositionCDTAFields(t *testing.T) {
	// Comp duration: 300 frames @ 30fps = 10.0s. Pick that explicitly so
	// the work-area "extend to Duration" sentinel test below has a known
	// substituted value.
	const (
		fps        = 30
		durFrames  = 300
		wantDurSec = 10.0
	)
	cases := []struct {
		name string
		opts cdtaOpts

		wantBG                        [3]uint8
		wantWorkAreaStart             float64
		wantWorkAreaEnd               float64
		wantShutterAngle              uint16
		wantShutterPhase              int32
		wantMotionBlurAdaptiveLimit   int32
		wantMotionBlurSamplesPerFrame int32
	}{
		{
			name: "ae defaults (zeroes everywhere)",
			opts: cdtaOpts{
				WorkAreaStartDivisor:      600,
				WorkAreaEndDivisor:        600,
				ShutterAngle:              180,
				MotionBlurAdaptiveLimit:   128,
				MotionBlurSamplesPerFrame: 16,
			},
			wantBG:                        [3]uint8{0, 0, 0},
			wantWorkAreaStart:             0,
			wantWorkAreaEnd:               0,
			wantShutterAngle:              180,
			wantShutterPhase:              0,
			wantMotionBlurAdaptiveLimit:   128,
			wantMotionBlurSamplesPerFrame: 16,
		},
		{
			name: "non-default values",
			opts: cdtaOpts{
				BGColor:                   [3]uint8{12, 34, 56},
				WorkAreaStartDividend:     1200, // 1200/600 = 2.0s
				WorkAreaStartDivisor:      600,
				WorkAreaEndDividend:       3600, // 3600/600 = 6.0s
				WorkAreaEndDivisor:        600,
				ShutterAngle:              270,
				ShutterPhase:              -90, // signed; matches fixture comp #0
				MotionBlurAdaptiveLimit:   200,
				MotionBlurSamplesPerFrame: 32,
			},
			wantBG:                        [3]uint8{12, 34, 56},
			wantWorkAreaStart:             2.0,
			wantWorkAreaEnd:               6.0,
			wantShutterAngle:              270,
			wantShutterPhase:              -90,
			wantMotionBlurAdaptiveLimit:   200,
			wantMotionBlurSamplesPerFrame: 32,
		},
		{
			name: "work area end sentinel 0xFFFFFFFF substitutes Duration",
			opts: cdtaOpts{
				WorkAreaStartDivisor:      600,
				WorkAreaEndDividend:       0xFFFFFFFF, // sentinel
				WorkAreaEndDivisor:        600,
				ShutterAngle:              180,
				MotionBlurAdaptiveLimit:   128,
				MotionBlurSamplesPerFrame: 16,
			},
			wantBG:                        [3]uint8{0, 0, 0},
			wantWorkAreaStart:             0,
			wantWorkAreaEnd:               wantDurSec,
			wantShutterAngle:              180,
			wantShutterPhase:              0,
			wantMotionBlurAdaptiveLimit:   128,
			wantMotionBlurSamplesPerFrame: 16,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cdta := buildCdtaWithOpts(1920, 1080, fps, 0, durFrames, c.opts)
			data := buildAEPWithCustomCdta(cdta, 0, 8000)
			proj, err := aep.FromReader(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("FromReader: %v", err)
			}
			comp := proj.Compositions[0]
			if comp.BGColor != c.wantBG {
				t.Errorf("BGColor = %v, want %v", comp.BGColor, c.wantBG)
			}
			if math.Abs(comp.WorkAreaStart-c.wantWorkAreaStart) > 1e-6 {
				t.Errorf("WorkAreaStart = %v, want %v", comp.WorkAreaStart, c.wantWorkAreaStart)
			}
			if math.Abs(comp.WorkAreaEnd-c.wantWorkAreaEnd) > 1e-6 {
				t.Errorf("WorkAreaEnd = %v, want %v", comp.WorkAreaEnd, c.wantWorkAreaEnd)
			}
			if comp.ShutterAngle != c.wantShutterAngle {
				t.Errorf("ShutterAngle = %v, want %v", comp.ShutterAngle, c.wantShutterAngle)
			}
			if comp.ShutterPhase != c.wantShutterPhase {
				t.Errorf("ShutterPhase = %v, want %v", comp.ShutterPhase, c.wantShutterPhase)
			}
			if comp.MotionBlurAdaptiveSampleLimit != c.wantMotionBlurAdaptiveLimit {
				t.Errorf("MotionBlurAdaptiveSampleLimit = %v, want %v",
					comp.MotionBlurAdaptiveSampleLimit, c.wantMotionBlurAdaptiveLimit)
			}
			if comp.MotionBlurSamplesPerFrame != c.wantMotionBlurSamplesPerFrame {
				t.Errorf("MotionBlurSamplesPerFrame = %v, want %v",
					comp.MotionBlurSamplesPerFrame, c.wantMotionBlurSamplesPerFrame)
			}
		})
	}
}

// TestCompositionCDTARealFixture parses the real test_data/re_tickrate.aep
// and asserts the CDTA tail values match what we hex-dumped from the file
// (shutter angle 180°, motion blur defaults). All 3 comps were exported
// from AE with default render settings, so their CDTA values are AE's
// own defaults — a strong indicator the parser is honoring real-world
// bytes, not just our synthetic ones.
func TestCompositionCDTARealFixture(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_tickrate.aep")
	if err != nil {
		t.Fatalf("Open re_tickrate.aep: %v", err)
	}
	if len(proj.Compositions) < 1 {
		t.Fatalf("expected at least 1 composition, got %d", len(proj.Compositions))
	}
	for i, comp := range proj.Compositions {
		t.Run(fmt.Sprintf("comp_%d_%s", i, comp.Name), func(t *testing.T) {
			if comp.ShutterAngle != 180 {
				t.Errorf("ShutterAngle = %d, want 180 (AE default)", comp.ShutterAngle)
			}
			if comp.MotionBlurAdaptiveSampleLimit != 128 {
				t.Errorf("MotionBlurAdaptiveSampleLimit = %d, want 128 (AE default)",
					comp.MotionBlurAdaptiveSampleLimit)
			}
			if comp.MotionBlurSamplesPerFrame != 16 {
				t.Errorf("MotionBlurSamplesPerFrame = %d, want 16 (AE default)",
					comp.MotionBlurSamplesPerFrame)
			}
			// Fixture has work_area_start=0 and end_dividend=0xFFFFFFFF
			// → WorkAreaEnd should equal Duration (sentinel substitution).
			if comp.WorkAreaStart != 0 {
				t.Errorf("WorkAreaStart = %v, want 0", comp.WorkAreaStart)
			}
			if math.Abs(comp.WorkAreaEnd-comp.Duration) > 1e-3 {
				t.Errorf("WorkAreaEnd = %v, want Duration=%v (sentinel substitution)",
					comp.WorkAreaEnd, comp.Duration)
			}
		})
	}
}
