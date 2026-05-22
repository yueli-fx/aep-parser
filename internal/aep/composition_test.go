package aep_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// buildAEPWithCustomCdta wraps a single keyframed Opacity property into a
// comp with custom cdta tick-rate bytes. Returns the bytes ready for
// aep.FromReader.
func buildAEPWithCustomCdta(cdta []byte, kfTime float64, kfTickRate float64) []byte {
	rb := &rifxBuilder{}
	// Build a 1D Opacity keyframe at the requested time, where the time
	// will be encoded as ticks at kfTickRate.
	kf := make([]byte, 48)
	binary.BigEndian.PutUint32(kf[0:], uint32(math.Round(kfTime*kfTickRate)))
	binary.BigEndian.PutUint64(kf[0x08:], math.Float64bits(0.5))
	op := rb.leafKeyframed("ADBE Opacity", 0x01, 48, [][]byte{kf})
	var tdgpBody []byte
	tdgpBody = append(tdgpBody, op...)
	tdgpBody = append(tdgpBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	tdgpList := rb.listChunk("LIST", "tdgp", tdgpBody)

	ldtaData := make([]byte, 0x2C)
	binary.BigEndian.PutUint32(ldtaData[0x28:], 1)
	ldta := rb.chunk("ldta", ldtaData)
	var layerBody []byte
	layerBody = append(layerBody, ldta...)
	layerBody = append(layerBody, tdgpList...)
	layrList := rb.listChunk("LIST", "Layr", layerBody)

	compName := rb.chunk("Utf8", []byte("Test"))
	compIdta := rb.chunk("idta", buildIdta(0x04, 1))
	compCdta := rb.chunk("cdta", cdta)
	var compBody []byte
	compBody = append(compBody, compName...)
	compBody = append(compBody, compIdta...)
	compBody = append(compBody, compCdta...)
	compBody = append(compBody, layrList...)
	compItem := rb.listChunk("LIST", "Item", compBody)
	foldList := rb.listChunk("LIST", "Fold", compItem)

	var root bytes.Buffer
	root.WriteString("RIFX")
	_ = binary.Write(&root, binary.BigEndian, uint32(4+len(foldList)))
	root.WriteString("Egg!")
	root.Write(foldList)
	return root.Bytes()
}

func TestPerCompTickRate(t *testing.T) {
	cases := []struct {
		name        string
		fpsWhole    uint16
		fpsFrac     uint16
		cdtaTick    uint32 // cdta_0x08
		legacyScale uint32 // cdta_0xA8
		wantRate    float64
	}{
		{"modern_30fps", 30, 0, 30720, 1, 30720},
		{"modern_24fps", 24, 0, 24576, 1, 24576},
		{"modern_29_97", 29, 63570, 23976, 1, 23976},
		{"legacy_29_97", 29, 63570, 23976, 2997, 8000},
		{"missing_scale_falls_back", 30, 0, 8000, 0, 8000},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cdta := buildCdtaWithRate(1920, 1080, tc.fpsWhole, tc.fpsFrac, 100,
				tc.cdtaTick, tc.legacyScale)
			// Keyframe at user-intent t=2.0s, encoded at the comp's tick rate
			// so the parser MUST derive the rate correctly to read t=2.0s
			// back out.
			data := buildAEPWithCustomCdta(cdta, 2.0, tc.wantRate)
			proj, err := aep.FromReader(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("FromReader: %v", err)
			}
			if len(proj.Compositions) != 1 {
				t.Fatalf("got %d comps", len(proj.Compositions))
			}
			comp := proj.Compositions[0]
			if comp.TickRate != tc.wantRate {
				t.Errorf("TickRate = %v, want %v", comp.TickRate, tc.wantRate)
			}
			if len(comp.Layers) != 1 || comp.Layers[0].Opacity() == nil {
				t.Fatalf("layer/opacity missing")
			}
			kf := comp.Layers[0].Opacity().Keyframes[0]
			if math.Abs(kf.Time-2.0) > 1e-3 {
				t.Errorf("kf.Time = %.4f, want ~2.0 (rate=%v)", kf.Time, tc.wantRate)
			}

			// Roundtrip: SetTime to 3.5s, write, reparse, verify.
			if err := kf.SetTime(3.5); err != nil {
				t.Fatalf("SetTime: %v", err)
			}
			var out bytes.Buffer
			if err := proj.WriteAEP(&out); err != nil {
				t.Fatalf("WriteAEP: %v", err)
			}
			proj2, err := aep.FromReader(bytes.NewReader(out.Bytes()))
			if err != nil {
				t.Fatalf("re-parse: %v", err)
			}
			kf2 := proj2.Compositions[0].Layers[0].Opacity().Keyframes[0]
			if math.Abs(kf2.Time-3.5) > 1e-3 {
				t.Errorf("after SetTime, kf2.Time = %.4f, want 3.5", kf2.Time)
			}
		})
	}
}

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
		BGColor:               [3]uint8{200, 100, 50},
		WorkAreaStartDivisor:  600,
		WorkAreaEndDividend:   0xFFFFFFFF,
		WorkAreaEndDivisor:    600,
		ShutterAngle:          180,
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
				WorkAreaStartDivisor:  600,
				WorkAreaEndDivisor:    600,
				ShutterAngle:          180,
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
				BGColor:               [3]uint8{12, 34, 56},
				WorkAreaStartDividend: 1200, // 1200/600 = 2.0s
				WorkAreaStartDivisor:  600,
				WorkAreaEndDividend:   3600, // 3600/600 = 6.0s
				WorkAreaEndDivisor:    600,
				ShutterAngle:          270,
				ShutterPhase:          -90, // signed; matches fixture comp #0
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
				WorkAreaStartDivisor: 600,
				WorkAreaEndDividend:  0xFFFFFFFF, // sentinel
				WorkAreaEndDivisor:   600,
				ShutterAngle:         180,
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

func TestCompositionLayerByID(t *testing.T) {
	comp := &aep.Composition{
		Layers: []*aep.Layer{
			{ID: 10, Name: "first"},
			{ID: 20, Name: "second"},
			{ID: 30, Name: "third"},
		},
	}
	if got := comp.LayerByID(20); got == nil || got.Name != "second" {
		t.Errorf("LayerByID(20) = %v, want layer named \"second\"", got)
	}
	if got := comp.LayerByID(999); got != nil {
		t.Errorf("LayerByID(999) = %v, want nil", got)
	}
	// Edge case: ID 0 is a valid "no parent" sentinel — should NOT match
	// any layer (real AE projects start IDs at 1).
	if got := comp.LayerByID(0); got != nil {
		t.Errorf("LayerByID(0) = %v, want nil", got)
	}
}

// TestCompositionSetters round-trips every Composition.SetX(): change
// the value, WriteAEP, re-parse, confirm persistence.
func TestCompositionSetters(t *testing.T) {
	data := buildMinimalAEP()
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	comp := proj.Compositions[0]

	mustNoErr := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	mustNoErr("SetBGColor", comp.SetBGColor([3]uint8{0x33, 0x66, 0x99}))
	mustNoErr("SetSize", comp.SetSize(3840, 2160))
	mustNoErr("SetShutterAngle", comp.SetShutterAngle(360))
	mustNoErr("SetShutterPhase", comp.SetShutterPhase(-180))
	mustNoErr("SetMotionBlurAdaptiveSampleLimit", comp.SetMotionBlurAdaptiveSampleLimit(256))
	mustNoErr("SetMotionBlurSamplesPerFrame", comp.SetMotionBlurSamplesPerFrame(32))
	mustNoErr("SetWorkArea", comp.SetWorkArea(1.5, 4.5))

	if comp.BGColor != [3]uint8{0x33, 0x66, 0x99} {
		t.Errorf("BGColor in-mem = %v", comp.BGColor)
	}
	if comp.Width != 3840 || comp.Height != 2160 {
		t.Errorf("Size in-mem = %dx%d, want 3840x2160", comp.Width, comp.Height)
	}
	if comp.ShutterAngle != 360 {
		t.Errorf("ShutterAngle in-mem = %d", comp.ShutterAngle)
	}
	if comp.ShutterPhase != -180 {
		t.Errorf("ShutterPhase in-mem = %d", comp.ShutterPhase)
	}
	if comp.MotionBlurAdaptiveSampleLimit != 256 {
		t.Errorf("MotionBlurAdaptiveSampleLimit in-mem = %d", comp.MotionBlurAdaptiveSampleLimit)
	}
	if comp.MotionBlurSamplesPerFrame != 32 {
		t.Errorf("MotionBlurSamplesPerFrame in-mem = %d", comp.MotionBlurSamplesPerFrame)
	}
	if math.Abs(comp.WorkAreaStart-1.5) > 1e-6 || math.Abs(comp.WorkAreaEnd-4.5) > 1e-6 {
		t.Errorf("WorkArea in-mem = (%g, %g), want (1.5, 4.5)", comp.WorkAreaStart, comp.WorkAreaEnd)
	}

	// Round-trip
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	c2 := proj2.Compositions[0]
	if c2.BGColor != [3]uint8{0x33, 0x66, 0x99} {
		t.Errorf("after roundtrip: BGColor = %v", c2.BGColor)
	}
	if c2.Width != 3840 || c2.Height != 2160 {
		t.Errorf("after roundtrip: Size = %dx%d, want 3840x2160", c2.Width, c2.Height)
	}
	if c2.ShutterAngle != 360 {
		t.Errorf("after roundtrip: ShutterAngle = %d", c2.ShutterAngle)
	}
	if c2.ShutterPhase != -180 {
		t.Errorf("after roundtrip: ShutterPhase = %d", c2.ShutterPhase)
	}
	if c2.MotionBlurAdaptiveSampleLimit != 256 {
		t.Errorf("after roundtrip: MotionBlurAdaptiveSampleLimit = %d", c2.MotionBlurAdaptiveSampleLimit)
	}
	if c2.MotionBlurSamplesPerFrame != 32 {
		t.Errorf("after roundtrip: MotionBlurSamplesPerFrame = %d", c2.MotionBlurSamplesPerFrame)
	}
	if math.Abs(c2.WorkAreaStart-1.5) > 1e-6 || math.Abs(c2.WorkAreaEnd-4.5) > 1e-6 {
		t.Errorf("after roundtrip: WorkArea = (%g, %g)", c2.WorkAreaStart, c2.WorkAreaEnd)
	}
}

// TestCompositionNameAndTiming covers SetName (length-variable Utf8)
// + SetFrameRate + SetDuration (length-preserving cdta writes).
func TestCompositionNameAndTiming(t *testing.T) {
	data := buildMinimalAEP()
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	comp := proj.Compositions[0]

	mustNoErr := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	mustNoErr("SetName", comp.SetName("Renamed Comp"))
	mustNoErr("SetFrameRate", comp.SetFrameRate(60.0))
	mustNoErr("SetDuration", comp.SetDuration(7.5))

	if comp.Name != "Renamed Comp" {
		t.Errorf("in-mem Name = %q", comp.Name)
	}
	if math.Abs(comp.FrameRate-60.0) > 1e-6 {
		t.Errorf("in-mem FrameRate = %g", comp.FrameRate)
	}
	if math.Abs(comp.Duration-7.5) > 1e-3 {
		t.Errorf("in-mem Duration = %g", comp.Duration)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	c2 := proj2.Compositions[0]
	if c2.Name != "Renamed Comp" {
		t.Errorf("roundtrip Name = %q", c2.Name)
	}
	if math.Abs(c2.FrameRate-60.0) > 1e-6 {
		t.Errorf("roundtrip FrameRate = %g", c2.FrameRate)
	}
	if math.Abs(c2.Duration-7.5) > 1e-3 {
		t.Errorf("roundtrip Duration = %g", c2.Duration)
	}
}

func TestCompositionSettersRejectMissingCdta(t *testing.T) {
	c := &aep.Composition{Name: "standalone"}
	if err := c.SetBGColor([3]uint8{1, 2, 3}); err == nil {
		t.Error("SetBGColor on comp without cdta: expected error")
	}
	if err := c.SetSize(1920, 1080); err == nil {
		t.Error("SetSize on comp without cdta: expected error")
	}
	if err := c.SetShutterAngle(180); err == nil {
		t.Error("SetShutterAngle on comp without cdta: expected error")
	}
	if err := c.SetWorkArea(0, 1); err == nil {
		t.Error("SetWorkArea on comp without cdta: expected error")
	}
}

// TestCompositionExtraFlagSetters covers SetFrameBlending /
// SetPreserveNestedResolution / SetPixelAspect against re_batch3.aep.
func TestCompositionExtraFlagSetters(t *testing.T) {
	proj, err := aep.FromReader(bytes.NewReader(buildMinimalAEP()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	comp := proj.Compositions[0]

	mustNoErr := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	mustNoErr("SetFrameBlending", comp.SetFrameBlending(true))
	mustNoErr("SetPreserveNestedResolution", comp.SetPreserveNestedResolution(true))
	mustNoErr("SetPixelAspect", comp.SetPixelAspect(2.0))

	if math.Abs(comp.PixelAspect-2.0) > 1e-6 {
		t.Errorf("in-mem PixelAspect = %g", comp.PixelAspect)
	}

	// Round-trip.
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if math.Abs(proj2.Compositions[0].PixelAspect-2.0) > 1e-6 {
		t.Errorf("roundtrip PixelAspect = %g", proj2.Compositions[0].PixelAspect)
	}
}

// TestPixelAspectReadReal verifies the cdta @0x90/@0x94 PAR read
// against re_batch3.aep's RE_B3_D_pixelAspect comp (PAR=2.0).
func TestPixelAspectReadReal(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_batch3.aep")
	if err != nil {
		t.Skipf("re_batch3.aep not present")
	}
	for _, c := range proj.Compositions {
		if c.Name == "RE_B3_D_pixelAspect" {
			if math.Abs(c.PixelAspect-2.0) > 1e-6 {
				t.Errorf("RE_B3_D_pixelAspect: PixelAspect = %g, want 2.0", c.PixelAspect)
			}
			return
		}
	}
	t.Skip("RE_B3_D_pixelAspect comp not found")
}

// TestResolutionFactorReal verifies cdta @0x00 / @0x02 read against
// re_cdta_probe.aep, where comps were explicitly set to varying
// resolutionFactor pairs.
func TestResolutionFactorReal(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_cdta_probe.aep")
	if err != nil {
		t.Skipf("re_cdta_probe.aep not present; run test_data/re_cdta_probe.jsx in AE")
	}
	want := map[string][2]uint16{
		"A_baseline":           {1, 1},
		"B_resolution_half":    {2, 2},
		"C_resolution_quarter": {4, 4},
		"D_resolution_3x4":     {3, 4},
	}
	got := map[string][2]uint16{}
	for _, c := range proj.Compositions {
		if _, ok := want[c.Name]; ok {
			got[c.Name] = c.ResolutionFactor
		}
	}
	for name, w := range want {
		g, ok := got[name]
		if !ok {
			t.Errorf("comp %q not found in fixture", name)
			continue
		}
		if g != w {
			t.Errorf("comp %q: ResolutionFactor = %v, want %v", name, g, w)
		}
	}
}

// TestSetResolutionFactorRoundtrip writes a new resolution factor pair,
// re-serializes, re-parses, and confirms persistence + JSON output.
func TestSetResolutionFactorRoundtrip(t *testing.T) {
	data := buildMinimalAEP()
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	c := proj.Compositions[0]
	if err := c.SetResolutionFactor(3, 5); err != nil {
		t.Fatalf("SetResolutionFactor: %v", err)
	}
	if c.ResolutionFactor != [2]uint16{3, 5} {
		t.Errorf("in-mem after set: %v, want [3 5]", c.ResolutionFactor)
	}
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if got := re.Compositions[0].ResolutionFactor; got != [2]uint16{3, 5} {
		t.Errorf("after roundtrip: %v, want [3 5]", got)
	}
	// Reject 0 factors.
	if err := c.SetResolutionFactor(0, 1); err == nil {
		t.Error("SetResolutionFactor(0, 1) should error")
	}
	if err := c.SetResolutionFactor(1, 0); err == nil {
		t.Error("SetResolutionFactor(1, 0) should error")
	}
	// Reject on missing cdta.
	standalone := &aep.Composition{Name: "standalone"}
	if err := standalone.SetResolutionFactor(2, 2); err == nil {
		t.Error("SetResolutionFactor on missing cdta should error")
	}
}

// TestCompositionRendererReal verifies Composition.Renderer decode
// from PRin LIST → prin chunk @offset 4 (ASCII NUL-padded match-name)
// against re_renderer.aep, which has 3 comps with different renderers.
func TestCompositionRendererReal(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_renderer.aep")
	if err != nil {
		t.Skipf("re_renderer.aep not present; run test_data/re_renderer.jsx in AE")
	}
	// In AE 2025, "ADBE Standard 3d" (Classic) and "ADBE Picasso" (Advanced)
	// both end up as "ADBE Escher" internally; "ADBE Ernst" stays Cinema 4D.
	want := map[string]string{
		"RDR_classic":  "ADBE Escher",
		"RDR_advanced": "ADBE Escher",
		"RDR_cinema4d": "ADBE Ernst",
	}
	for _, c := range proj.Compositions {
		if expected, ok := want[c.Name]; ok {
			if c.Renderer != expected {
				t.Errorf("comp %q: Renderer = %q, want %q", c.Name, c.Renderer, expected)
			}
		}
	}
}

// TestCompositionActiveCamera verifies ActiveCamera() picks the
// topmost enabled camera layer using re_cameralight.aep.
func TestCompositionActiveCamera(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present")
	}
	for _, c := range proj.Compositions {
		if c.Name != "RE_CL" {
			continue
		}
		cam := c.ActiveCamera()
		if cam == nil {
			t.Fatal("ActiveCamera() = nil, expected a camera layer")
		}
		if cam.Type != aep.LayerTypeCamera {
			t.Errorf("ActiveCamera type = %s, want camera", cam.Type)
		}
		if cam.Name != "MyCamera" {
			t.Errorf("ActiveCamera name = %q, want MyCamera", cam.Name)
		}
		return
	}
	t.Skip("RE_CL comp not found")
}

// TestCompositionFlagSetters covers the 4 cdta flag bit setters
// (Draft3D, HideShyLayers, CompMotionBlur, PreserveNestedFrameRate).
func TestCompositionFlagSetters(t *testing.T) {
	proj, err := aep.FromReader(bytes.NewReader(buildMinimalAEP()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	comp := proj.Compositions[0]

	mustNoErr := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	mustNoErr("SetHideShyLayers", comp.SetHideShyLayers(true))
	mustNoErr("SetCompMotionBlur", comp.SetCompMotionBlur(true))
	mustNoErr("SetPreserveNestedFrameRate", comp.SetPreserveNestedFrameRate(true))
	mustNoErr("SetDraft3D", comp.SetDraft3D(true))

	// Each flag occupies a separate bit — confirm they coexist by
	// reading back the cdta bytes.
	// (No public getter for these yet; just confirm WriteAEP + reparse
	// doesn't blow up.)
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	_, err = aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}

	// Flip back off; should be idempotent.
	mustNoErr("SetHideShyLayers(false)", comp.SetHideShyLayers(false))
	mustNoErr("SetCompMotionBlur(false)", comp.SetCompMotionBlur(false))
}

func TestDisplayStartTimeReal(t *testing.T) {
	proj := openWave2AE24(t)
	var baseline, dsf120 *aep.Composition
	for _, c := range proj.Compositions {
		switch c.Name {
		case "RE_CDTA_DSF":
			if baseline == nil || c.ID > baseline.ID {
				baseline = c
			}
		case "RE_CDTA_DSF_120":
			if dsf120 == nil || c.ID > dsf120.ID {
				dsf120 = c
			}
		}
	}
	if baseline == nil || dsf120 == nil {
		t.Fatalf("baseline / dsf120 missing")
	}
	if baseline.DisplayStartTime != 0 {
		t.Errorf("baseline DisplayStartTime = %g, want 0", baseline.DisplayStartTime)
	}
	// dsf120 was set to displayStartFrame=120 at 29.97 fps ≈ 4 sec.
	// AE 25's encoding produces ~3.98 sec (rounding quirks).
	if dsf120.DisplayStartTime < 3.9 || dsf120.DisplayStartTime > 4.1 {
		t.Errorf("dsf120 DisplayStartTime = %g, want ~4.0", dsf120.DisplayStartTime)
	}
}

func TestSetDisplayStartTimeRoundtrip(t *testing.T) {
	proj := openWave2AE24(t)
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "RE_CDTA_DSF" && (comp == nil || c.ID > comp.ID) {
			comp = c
		}
	}
	if comp == nil {
		t.Fatalf("RE_CDTA_DSF missing")
	}
	compID := comp.ID
	// Set to 2.5 seconds
	if err := comp.SetDisplayStartTime(2.5); err != nil {
		t.Fatalf("SetDisplayStartTime: %v", err)
	}
	if math.Abs(comp.DisplayStartTime-2.5) > 0.01 {
		t.Errorf("after Set: DisplayStartTime = %g, want ~2.5", comp.DisplayStartTime)
	}

	// Roundtrip
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	var found *aep.Composition
	for _, c := range proj2.Compositions {
		if c.ID == compID {
			found = c
			break
		}
	}
	if found == nil {
		t.Fatalf("roundtrip: comp (id=%d) missing", compID)
	}
	if math.Abs(found.DisplayStartTime-2.5) > 0.01 {
		t.Errorf("roundtrip DisplayStartTime = %g, want ~2.5", found.DisplayStartTime)
	}

	// Clear back to 0
	if err := found.SetDisplayStartTime(0); err != nil {
		t.Fatalf("clear: %v", err)
	}
	if found.DisplayStartTime != 0 {
		t.Errorf("after clear: DisplayStartTime = %g, want 0", found.DisplayStartTime)
	}

	// Frame helper
	if err := found.SetDisplayStartFrame(60); err != nil {
		t.Fatalf("SetDisplayStartFrame: %v", err)
	}
	wantSec := 60.0 / found.FrameRate
	if math.Abs(found.DisplayStartTime-wantSec) > 0.01 {
		t.Errorf("SetDisplayStartFrame(60): DisplayStartTime = %g, want ~%g", found.DisplayStartTime, wantSec)
	}

	// Reject negative
	if err := found.SetDisplayStartTime(-1); err == nil {
		t.Error("negative: expected error")
	}
}
