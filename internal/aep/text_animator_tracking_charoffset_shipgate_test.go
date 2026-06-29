// internal/aep/text_animator_tracking_charoffset_shipgate_test.go
//
// AE ship gate for the last two Booyah comp ② Text Animator leaves — Tracking
// Amount and Character Offset. Both are 1D scalars on the established splice +
// Range Selector machinery (RE'd straight from the real Booyah project, see
// tmp_debug/probe_text_tracking + incidents/text-animator-create-re.md), so per
// delivery-contract red line 4 the gate renders THREE frames of the SAME
// from-scratch layer and asserts on ACTUAL PIXELS that each leaf's visible
// surface changes as the Range Selector sweeps the per-character effect off the
// glyphs (offset 0→100 over t=0..2):
//
//   - Tracking Amount  → glyph ink bbox WIDTH (selected chars spread apart →
//     wide at t0, normal at t2)
//   - Character Offset → frameDiff between frames (selected glyphs are remapped
//     to other letters at t0, original letters at t2 — the
//     rendered glyph shapes differ, while every frame keeps ink)
//
// Reuses buildNeighborDemo / runOneNeighborGate + verify_text_animator_neighbor.jsx
// from text_animator_neighbor_shipgate_test.go. Gated by AE_SHIP_GATE.
package aep_test

import (
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// TestTextTrackingCharOffset_RoundTrip (no AE): both animators survive
// WriteAEP → Open, so normal `go test ./...` guards against regression even
// without the AE ship gate.
func TestTextTrackingCharOffset_RoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name  string
		mn    string
		build func(*aep.Layer) error
	}{
		{"Tracking", "ADBE Text Tracking Amount", func(tl *aep.Layer) error {
			_, e := aep.AddTextTrackingAnimator(tl, 500, 0, 100, 0)
			return e
		}},
		{"CharOffset", "ADBE Text Character Offset", func(tl *aep.Layer) error {
			_, e := aep.AddTextCharacterOffsetAnimator(tl, 13, 0, 100, 0)
			return e
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := buildNeighborDemo(t, aep.TargetAE2020, "TXRT_"+tc.name, "ABCDEF", tc.build)
			path := filepath.Join(t.TempDir(), "tc.aep")
			f, err := os.Create(path)
			if err != nil {
				t.Fatal(err)
			}
			if err := p.WriteAEP(f); err != nil {
				f.Close()
				t.Fatalf("WriteAEP: %v", err)
			}
			f.Close()
			if _, err := aep.Open(path); err != nil {
				t.Fatalf("reopen: %v", err)
			}
			root := parseAEP(t, path)
			if !findShipChunkTdmnName(root, tc.mn) {
				t.Errorf("%s leaf %q dropped after round-trip", tc.name, tc.mn)
			}
		})
	}
}

// Sample box over the centred text region in the 1280×720 comp (same as neighbors).
const tcoX0, tcoY0, tcoX1, tcoY1, tcoStep = 120, 60, 1160, 680, 2
const tcoBgLum = 12

func trackingSpec() neighborSpec {
	return neighborSpec{
		key:  "Tracking",
		text: "ABCDEF",
		add:  func(tl *aep.Layer) error { _, e := aep.AddTextTrackingAnimator(tl, 500, 0, 100, 0); return e },
		args: `,"fontSize":120,"posX":220,"posY":420`,
		check: func(t *testing.T, ver string, im0, im1, im2 image.Image) {
			w0, _, n0 := inkBBox(im0, tcoX0, tcoY0, tcoX1, tcoY1, tcoStep, tcoBgLum, 40)
			w1, _, n1 := inkBBox(im1, tcoX0, tcoY0, tcoX1, tcoY1, tcoStep, tcoBgLum, 40)
			w2, _, n2 := inkBBox(im2, tcoX0, tcoY0, tcoX1, tcoY1, tcoStep, tcoBgLum, 40)
			t.Logf("%s Tracking ink-width: t0=%d(n=%d) t1=%d(n=%d) t2=%d(n=%d)", ver, w0, n0, w1, n1, w2, n2)
			if n0 < 100 || n1 < 100 || n2 < 100 {
				t.Errorf("%s glyphs not rendered on some frame (n0=%d n1=%d n2=%d)", ver, n0, n1, n2)
			}
			// Selected (t0) → wide spacing; deselected (t2) → normal spacing.
			if w0 <= w2 || w0 < int(float64(w2)*1.15) {
				t.Errorf("%s tracking not spreading characters (t0 width=%d should exceed t2=%d by 1.15x)", ver, w0, w2)
			}
			if w1 > w0 || w1 < w2 {
				t.Errorf("%s mid width not between ends (t0=%d t1=%d t2=%d)", ver, w0, w1, w2)
			}
		},
	}
}

func charOffsetSpec() neighborSpec {
	return neighborSpec{
		key:  "CharOffset",
		text: "ABCDEF",
		add:  func(tl *aep.Layer) error { _, e := aep.AddTextCharacterOffsetAnimator(tl, 13, 0, 100, 0); return e },
		args: `,"fontSize":120,"posX":360,"posY":420`,
		check: func(t *testing.T, ver string, im0, im1, im2 image.Image) {
			_, _, n0 := inkBBox(im0, tcoX0, tcoY0, tcoX1, tcoY1, tcoStep, tcoBgLum, 40)
			_, _, n2 := inkBBox(im2, tcoX0, tcoY0, tcoX1, tcoY1, tcoStep, tcoBgLum, 40)
			d01 := frameDiff(im0, im1, tcoX0, tcoY0, tcoX1, tcoY1, tcoStep, 40)
			d12 := frameDiff(im1, im2, tcoX0, tcoY0, tcoX1, tcoY1, tcoStep, 40)
			d02 := frameDiff(im0, im2, tcoX0, tcoY0, tcoX1, tcoY1, tcoStep, 40)
			t.Logf("%s CharOffset: ink n0=%d n2=%d ; frameDiff 0-1=%d 1-2=%d 0-2=%d", ver, n0, n2, d01, d12, d02)
			if n0 < 100 || n2 < 100 {
				t.Errorf("%s glyphs not rendered (n0=%d n2=%d) — layer dropped?", ver, n0, n2)
			}
			// Offset glyphs (t0) differ from original glyphs (t2): the rendered
			// letters are remapped, so a large fraction of pixels differ.
			if d02 < 200 {
				t.Errorf("%s character offset not remapping glyphs (frameDiff 0-2=%d, want > 200)", ver, d02)
			}
			// The mid frame (partial selection) differs from at least one end too.
			if d01 < 80 && d12 < 80 {
				t.Errorf("%s mid frame static (d01=%d d12=%d) — sweep not affecting glyphs", ver, d01, d12)
			}
		},
	}
}

func runTrackingCharOffset(t *testing.T, aeExe, ver string, target aep.AETarget, sp neighborSpec) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	runOneNeighborGate(t, aeExe, ver, target, sp)
}

func TestTextTrackingAnimator_AEShipGate_AE2020(t *testing.T) {
	runTrackingCharOffset(t, ae2020(), "AE2020", aep.TargetAE2020, trackingSpec())
}

func TestTextTrackingAnimator_AEShipGate_AE2025(t *testing.T) {
	runTrackingCharOffset(t, ae2025(), "AE2025", aep.TargetAE2025, trackingSpec())
}

func TestTextCharOffsetAnimator_AEShipGate_AE2020(t *testing.T) {
	runTrackingCharOffset(t, ae2020(), "AE2020", aep.TargetAE2020, charOffsetSpec())
}

func TestTextCharOffsetAnimator_AEShipGate_AE2025(t *testing.T) {
	runTrackingCharOffset(t, ae2025(), "AE2025", aep.TargetAE2025, charOffsetSpec())
}
