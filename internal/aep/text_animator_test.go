// internal/aep/text_animator_test.go
//
// Non-AE coverage for from-scratch Text Animators: AddTextOpacityAnimator +
// AnimateTextRangeOffset round-trip through WriteAEP → Open with no warnings,
// plus byte-structural assertions on the spliced chunk tree (the AE-acceptance
// surface is the double-version ship gate in text_animator_shipgate_test.go).
package aep_test

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/rifx"
)

func writeReopen(t *testing.T, p *aep.Project, name string) (*aep.Project, *rifx.Chunk) {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(f); err != nil {
		f.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	f.Close()
	rp, err := aep.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if len(rp.Warnings) != 0 {
		t.Fatalf("reopen warnings: %v", rp.Warnings)
	}
	return rp, parseAEP(t, path)
}

// followingList returns the LIST chunk (any FormType) directly after the first
// tdmn == matchName found anywhere under root, or nil.
func followingList(root *rifx.Chunk, matchName string) *rifx.Chunk {
	var found *rifx.Chunk
	var walk func(*rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		for i := 0; i+1 < len(c.Children); i++ {
			ch := c.Children[i]
			if ch.ID == rifx.IDTdmn && string(bytes.TrimRight(ch.Data, "\x00")) == matchName {
				if next := c.Children[i+1]; next.IsList() {
					found = next
					return
				}
			}
			if ch.IsList() {
				walk(ch)
				if found != nil {
					return
				}
			}
		}
	}
	walk(root)
	return found
}

// cdatF64 returns the BE f64 at cdat[0:8] of the tdbs following tdmn matchName.
func cdatF64(t *testing.T, root *rifx.Chunk, matchName string) float64 {
	t.Helper()
	tdbs := followingList(root, matchName)
	if tdbs == nil {
		t.Fatalf("%s: tdbs not found", matchName)
	}
	cdat := findShipChunk(tdbs, rifx.IDCdat)
	if cdat == nil || len(cdat.Data) < 8 {
		t.Fatalf("%s: cdat missing/short", matchName)
	}
	return math.Float64frombits(binary.BigEndian.Uint64(cdat.Data[0:8]))
}

// cdatVec3 returns the three BE f64 at cdat[0:24] of the tdbs following tdmn
// matchName (a spatial 3-component value such as Position 3D).
func cdatVec3(t *testing.T, root *rifx.Chunk, matchName string) [3]float64 {
	t.Helper()
	tdbs := followingList(root, matchName)
	if tdbs == nil {
		t.Fatalf("%s: tdbs not found", matchName)
	}
	cdat := findShipChunk(tdbs, rifx.IDCdat)
	if cdat == nil || len(cdat.Data) < 24 {
		t.Fatalf("%s: cdat missing/short (%d)", matchName, len(cdat.Data))
	}
	var v [3]float64
	for i := range v {
		v[i] = math.Float64frombits(binary.BigEndian.Uint64(cdat.Data[8*i : 8*i+8]))
	}
	return v
}

// cdatVec4 returns the four BE f64 at cdat[0:32] of the tdbs following tdmn
// matchName (a 4-channel colour value such as Fill Color, stored [A,R,G,B]×255).
func cdatVec4(t *testing.T, root *rifx.Chunk, matchName string) [4]float64 {
	t.Helper()
	tdbs := followingList(root, matchName)
	if tdbs == nil {
		t.Fatalf("%s: tdbs not found", matchName)
	}
	cdat := findShipChunk(tdbs, rifx.IDCdat)
	if cdat == nil || len(cdat.Data) < 32 {
		t.Fatalf("%s: cdat missing/short (%d)", matchName, len(cdat.Data))
	}
	var v [4]float64
	for i := range v {
		v[i] = math.Float64frombits(binary.BigEndian.Uint64(cdat.Data[8*i : 8*i+8]))
	}
	return v
}

func TestTextColorAnimator_Static_RoundTrip(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "T", 1280, 720, 24, 5)
	if err != nil {
		t.Fatal(err)
	}
	tl, err := aep.NewTextLayer(comp, "TXT")
	if err != nil {
		t.Fatal(err)
	}
	if err := tl.SetText("ABCDEF"); err != nil {
		t.Fatal(err)
	}
	// r,g,b,a = 0.2,0.4,0.8,1 → on disk [A,R,G,B]×255 = [255,51,102,204].
	if _, err := aep.AddTextColorAnimator(tl, 0.2, 0.4, 0.8, 1, 10, 60, 25); err != nil {
		t.Fatalf("AddTextColorAnimator: %v", err)
	}

	_, root := writeReopen(t, p, "txcolorstatic.aep")

	got := cdatVec4(t, root, "ADBE Text Fill Color")
	want := [4]float64{255, 0.2 * 255, 0.4 * 255, 0.8 * 255}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-3 {
			t.Errorf("Fill Color cdat[%d] = %g, want %g (full=%v)", i, got[i], want[i], got)
		}
	}
	if got := cdatF64(t, root, "ADBE Text Percent Start"); got != 10 {
		t.Errorf("Start cdat = %g, want 10", got)
	}
	if got := cdatF64(t, root, "ADBE Text Percent End"); got != 60 {
		t.Errorf("End cdat = %g, want 60", got)
	}
	if got := cdatF64(t, root, "ADBE Text Percent Offset"); got != 25 {
		t.Errorf("Offset cdat = %g, want 25", got)
	}
}

func TestTextColorAnimator_RevealSweep_RoundTrip(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "T", 1280, 720, 24, 5)
	if err != nil {
		t.Fatal(err)
	}
	tl, err := aep.NewTextLayer(comp, "TXT")
	if err != nil {
		t.Fatal(err)
	}
	if err := tl.SetText("ABCDEF"); err != nil {
		t.Fatal(err)
	}
	if _, err := aep.AddTextColorAnimator(tl, 1, 0, 0, 1, 0, 100, 0); err != nil {
		t.Fatalf("AddTextColorAnimator: %v", err)
	}
	if err := aep.AnimateTextRangeOffset(tl, 0, []aep.ScalarKeyframe{{Time: 0, Value: 0}, {Time: 2, Value: 100}}); err != nil {
		t.Fatalf("AnimateTextRangeOffset: %v", err)
	}

	_, root := writeReopen(t, p, "txcoloranim.aep")

	kfl := findShipList(root, "ADBE Text Percent Offset")
	if kfl == nil {
		t.Fatal("Offset list not found")
	}
	lhd3 := findShipChunk(kfl, rifx.IDLhd3)
	if lhd3 == nil {
		t.Fatal("Offset lhd3 missing (not animated)")
	}
	if n := binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]); n != 2 {
		t.Errorf("Offset numKf = %d, want 2", n)
	}
	got := cdatVec4(t, root, "ADBE Text Fill Color")
	want := [4]float64{255, 255, 0, 0} // [A,R,G,B]×255 of opaque red
	for i := range want {
		if math.Abs(got[i]-want[i]) > 1e-3 {
			t.Errorf("Fill Color cdat[%d] = %g, want %g (full=%v)", i, got[i], want[i], got)
		}
	}
}

func TestAddTextColorAnimator_RefusesNonText(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "T", 1280, 720, 24, 5)
	if err != nil {
		t.Fatal(err)
	}
	nl, err := aep.NewNullLayer(comp, "NULL")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aep.AddTextColorAnimator(nl, 1, 0, 0, 1, 0, 50, 0); err == nil {
		t.Fatal("expected refuse on non-text layer, got nil")
	}
}

func TestTextPositionAnimator_Static_RoundTrip(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "T", 1280, 720, 24, 5)
	if err != nil {
		t.Fatal(err)
	}
	tl, err := aep.NewTextLayer(comp, "TXT")
	if err != nil {
		t.Fatal(err)
	}
	if err := tl.SetText("ABCDEF"); err != nil {
		t.Fatal(err)
	}
	if _, err := aep.AddTextPositionAnimator(tl, 0, -100, 0, 10, 60, 25); err != nil {
		t.Fatalf("AddTextPositionAnimator: %v", err)
	}

	_, root := writeReopen(t, p, "txposstatic.aep")

	if got := cdatVec3(t, root, "ADBE Text Position 3D"); got != [3]float64{0, -100, 0} {
		t.Errorf("Position cdat = %v, want [0 -100 0]", got)
	}
	if got := cdatF64(t, root, "ADBE Text Percent Start"); got != 10 {
		t.Errorf("Start cdat = %g, want 10", got)
	}
	if got := cdatF64(t, root, "ADBE Text Percent End"); got != 60 {
		t.Errorf("End cdat = %g, want 60", got)
	}
	if got := cdatF64(t, root, "ADBE Text Percent Offset"); got != 25 {
		t.Errorf("Offset cdat = %g, want 25", got)
	}
}

func TestTextPositionAnimator_RevealSweep_RoundTrip(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "T", 1280, 720, 24, 5)
	if err != nil {
		t.Fatal(err)
	}
	tl, err := aep.NewTextLayer(comp, "TXT")
	if err != nil {
		t.Fatal(err)
	}
	if err := tl.SetText("ABCDEF"); err != nil {
		t.Fatal(err)
	}
	if _, err := aep.AddTextPositionAnimator(tl, 0, -120, 0, 0, 100, 0); err != nil {
		t.Fatalf("AddTextPositionAnimator: %v", err)
	}
	if err := aep.AnimateTextRangeOffset(tl, 0, []aep.ScalarKeyframe{{Time: 0, Value: 0}, {Time: 2, Value: 100}}); err != nil {
		t.Fatalf("AnimateTextRangeOffset: %v", err)
	}

	_, root := writeReopen(t, p, "txposanim.aep")

	// Offset became a 2-keyframe bpk-48 1D non-spatial container (the sweep).
	kfl := findShipList(root, "ADBE Text Percent Offset")
	if kfl == nil {
		t.Fatal("Offset list not found")
	}
	lhd3 := findShipChunk(kfl, rifx.IDLhd3)
	if lhd3 == nil {
		t.Fatal("Offset lhd3 missing (not animated)")
	}
	if n := binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]); n != 2 {
		t.Errorf("Offset numKf = %d, want 2", n)
	}
	// Position stays the static displacement.
	if got := cdatVec3(t, root, "ADBE Text Position 3D"); got != [3]float64{0, -120, 0} {
		t.Errorf("Position cdat = %v, want [0 -120 0]", got)
	}
}

func TestAddTextPositionAnimator_AppendsToExisting(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "T", 1280, 720, 24, 5)
	if err != nil {
		t.Fatal(err)
	}
	tl, err := aep.NewTextLayer(comp, "TXT")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aep.AddTextOpacityAnimator(tl, 0, 0, 50, 0); err != nil {
		t.Fatalf("opacity animator: %v", err)
	}
	if _, err := aep.AddTextPositionAnimator(tl, 0, -80, 0, 0, 50, 0); err != nil {
		t.Fatalf("position animator: %v", err)
	}
	_, root := writeReopen(t, p, "txmixed.aep")
	animators := followingList(root, "ADBE Text Animators")
	if animators == nil {
		t.Fatal("Animators group not found")
	}
	n := 0
	for _, ch := range animators.Children {
		if ch.ID == rifx.IDTdmn && string(bytes.TrimRight(ch.Data, "\x00")) == "ADBE Text Animator" {
			n++
		}
	}
	if n != 2 {
		t.Errorf("animator count = %d, want 2", n)
	}
	if got := cdatVec3(t, root, "ADBE Text Position 3D"); got != [3]float64{0, -80, 0} {
		t.Errorf("Position cdat = %v, want [0 -80 0]", got)
	}
	if got := cdatF64(t, root, "ADBE Text Opacity"); got != 0 {
		t.Errorf("Opacity cdat = %g, want 0", got)
	}
}

func TestAddTextPositionAnimator_RefusesNonText(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "T", 1280, 720, 24, 5)
	if err != nil {
		t.Fatal(err)
	}
	nl, err := aep.NewNullLayer(comp, "NULL")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aep.AddTextPositionAnimator(nl, 0, -100, 0, 0, 50, 0); err == nil {
		t.Fatal("expected refuse on non-text layer, got nil")
	}
}

func TestTextScaleAnimator_Static_RoundTrip(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "T", 1280, 720, 24, 5)
	if err != nil {
		t.Fatal(err)
	}
	tl, err := aep.NewTextLayer(comp, "TXT")
	if err != nil {
		t.Fatal(err)
	}
	if err := tl.SetText("ABCDEF"); err != nil {
		t.Fatal(err)
	}
	if _, err := aep.AddTextScaleAnimator(tl, 220, 180, 100, 10, 60, 25); err != nil {
		t.Fatalf("AddTextScaleAnimator: %v", err)
	}

	_, root := writeReopen(t, p, "txscalestatic.aep")

	if got := cdatVec3(t, root, "ADBE Text Scale 3D"); got != [3]float64{220, 180, 100} {
		t.Errorf("Scale cdat = %v, want [220 180 100]", got)
	}
	if got := cdatF64(t, root, "ADBE Text Percent Start"); got != 10 {
		t.Errorf("Start cdat = %g, want 10", got)
	}
	if got := cdatF64(t, root, "ADBE Text Percent End"); got != 60 {
		t.Errorf("End cdat = %g, want 60", got)
	}
	if got := cdatF64(t, root, "ADBE Text Percent Offset"); got != 25 {
		t.Errorf("Offset cdat = %g, want 25", got)
	}
}

func TestTextScaleAnimator_RevealSweep_RoundTrip(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "T", 1280, 720, 24, 5)
	if err != nil {
		t.Fatal(err)
	}
	tl, err := aep.NewTextLayer(comp, "TXT")
	if err != nil {
		t.Fatal(err)
	}
	if err := tl.SetText("ABCDEF"); err != nil {
		t.Fatal(err)
	}
	if _, err := aep.AddTextScaleAnimator(tl, 0, 0, 100, 0, 100, 0); err != nil {
		t.Fatalf("AddTextScaleAnimator: %v", err)
	}
	if err := aep.AnimateTextRangeOffset(tl, 0, []aep.ScalarKeyframe{{Time: 0, Value: 0}, {Time: 2, Value: 100}}); err != nil {
		t.Fatalf("AnimateTextRangeOffset: %v", err)
	}

	_, root := writeReopen(t, p, "txscaleanim.aep")

	kfl := findShipList(root, "ADBE Text Percent Offset")
	if kfl == nil {
		t.Fatal("Offset list not found")
	}
	lhd3 := findShipChunk(kfl, rifx.IDLhd3)
	if lhd3 == nil {
		t.Fatal("Offset lhd3 missing (not animated)")
	}
	if n := binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]); n != 2 {
		t.Errorf("Offset numKf = %d, want 2", n)
	}
	if got := cdatVec3(t, root, "ADBE Text Scale 3D"); got != [3]float64{0, 0, 100} {
		t.Errorf("Scale cdat = %v, want [0 0 100]", got)
	}
}

func TestAddTextScaleAnimator_RefusesNonText(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "T", 1280, 720, 24, 5)
	if err != nil {
		t.Fatal(err)
	}
	nl, err := aep.NewNullLayer(comp, "NULL")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aep.AddTextScaleAnimator(nl, 0, 0, 100, 0, 50, 0); err == nil {
		t.Fatal("expected refuse on non-text layer, got nil")
	}
}

func TestTextRotationAnimator_Static_RoundTrip(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "T", 1280, 720, 24, 5)
	if err != nil {
		t.Fatal(err)
	}
	tl, err := aep.NewTextLayer(comp, "TXT")
	if err != nil {
		t.Fatal(err)
	}
	if err := tl.SetText("ABCDEF"); err != nil {
		t.Fatal(err)
	}
	if _, err := aep.AddTextRotationAnimator(tl, 90, 10, 60, 25); err != nil {
		t.Fatalf("AddTextRotationAnimator: %v", err)
	}

	_, root := writeReopen(t, p, "txrotstatic.aep")

	if got := cdatF64(t, root, "ADBE Text Rotation"); got != 90 {
		t.Errorf("Rotation cdat = %g, want 90", got)
	}
	if got := cdatF64(t, root, "ADBE Text Percent Start"); got != 10 {
		t.Errorf("Start cdat = %g, want 10", got)
	}
	if got := cdatF64(t, root, "ADBE Text Percent End"); got != 60 {
		t.Errorf("End cdat = %g, want 60", got)
	}
	if got := cdatF64(t, root, "ADBE Text Percent Offset"); got != 25 {
		t.Errorf("Offset cdat = %g, want 25", got)
	}
}

func TestTextRotationAnimator_RevealSweep_RoundTrip(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "T", 1280, 720, 24, 5)
	if err != nil {
		t.Fatal(err)
	}
	tl, err := aep.NewTextLayer(comp, "TXT")
	if err != nil {
		t.Fatal(err)
	}
	if err := tl.SetText("ABCDEF"); err != nil {
		t.Fatal(err)
	}
	if _, err := aep.AddTextRotationAnimator(tl, 90, 0, 100, 0); err != nil {
		t.Fatalf("AddTextRotationAnimator: %v", err)
	}
	if err := aep.AnimateTextRangeOffset(tl, 0, []aep.ScalarKeyframe{{Time: 0, Value: 0}, {Time: 2, Value: 100}}); err != nil {
		t.Fatalf("AnimateTextRangeOffset: %v", err)
	}

	_, root := writeReopen(t, p, "txrotanim.aep")

	kfl := findShipList(root, "ADBE Text Percent Offset")
	if kfl == nil {
		t.Fatal("Offset list not found")
	}
	lhd3 := findShipChunk(kfl, rifx.IDLhd3)
	if lhd3 == nil {
		t.Fatal("Offset lhd3 missing (not animated)")
	}
	if n := binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]); n != 2 {
		t.Errorf("Offset numKf = %d, want 2", n)
	}
	if got := cdatF64(t, root, "ADBE Text Rotation"); got != 90 {
		t.Errorf("Rotation cdat = %g, want 90", got)
	}
}

func TestAddTextRotationAnimator_RefusesNonText(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "T", 1280, 720, 24, 5)
	if err != nil {
		t.Fatal(err)
	}
	nl, err := aep.NewNullLayer(comp, "NULL")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aep.AddTextRotationAnimator(nl, 90, 0, 50, 0); err == nil {
		t.Fatal("expected refuse on non-text layer, got nil")
	}
}

func TestTextOpacityAnimator_Static_RoundTrip(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "T", 1280, 720, 24, 5)
	if err != nil {
		t.Fatal(err)
	}
	tl, err := aep.NewTextLayer(comp, "TXT")
	if err != nil {
		t.Fatal(err)
	}
	if err := tl.SetText("ABCDEF"); err != nil {
		t.Fatal(err)
	}
	if _, err := aep.AddTextOpacityAnimator(tl, 0, 10, 60, 25); err != nil {
		t.Fatalf("AddTextOpacityAnimator: %v", err)
	}

	_, root := writeReopen(t, p, "txstatic.aep")

	if got := cdatF64(t, root, "ADBE Text Opacity"); got != 0 {
		t.Errorf("Opacity cdat = %g, want 0", got)
	}
	if got := cdatF64(t, root, "ADBE Text Percent Start"); got != 10 {
		t.Errorf("Start cdat = %g, want 10", got)
	}
	if got := cdatF64(t, root, "ADBE Text Percent End"); got != 60 {
		t.Errorf("End cdat = %g, want 60", got)
	}
	if got := cdatF64(t, root, "ADBE Text Percent Offset"); got != 25 {
		t.Errorf("Offset cdat = %g, want 25", got)
	}
}

func TestTextOpacityAnimator_AnimatedOffset_RoundTrip(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "T", 1280, 720, 24, 5)
	if err != nil {
		t.Fatal(err)
	}
	tl, err := aep.NewTextLayer(comp, "TXT")
	if err != nil {
		t.Fatal(err)
	}
	if err := tl.SetText("ABCDEF"); err != nil {
		t.Fatal(err)
	}
	if _, err := aep.AddTextOpacityAnimator(tl, 0, 0, 100, 0); err != nil {
		t.Fatalf("AddTextOpacityAnimator: %v", err)
	}
	if err := aep.AnimateTextRangeOffset(tl, 0, []aep.ScalarKeyframe{{Time: 0, Value: 0}, {Time: 2, Value: 100}}); err != nil {
		t.Fatalf("AnimateTextRangeOffset: %v", err)
	}

	_, root := writeReopen(t, p, "txanim.aep")

	// Offset became a 2-keyframe bpk-48 1D non-spatial container.
	kfl := findShipList(root, "ADBE Text Percent Offset")
	if kfl == nil {
		t.Fatal("Offset list not found")
	}
	lhd3 := findShipChunk(kfl, rifx.IDLhd3)
	if lhd3 == nil {
		t.Fatal("Offset lhd3 missing (not animated)")
	}
	if n := binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]); n != 2 {
		t.Errorf("Offset numKf = %d, want 2", n)
	}
	if bpk := binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]); bpk != 48 {
		t.Errorf("Offset bpk = %d, want 48", bpk)
	}
	// Opacity stays static 0.
	if got := cdatF64(t, root, "ADBE Text Opacity"); got != 0 {
		t.Errorf("Opacity cdat = %g, want 0", got)
	}
}

func TestAnimateTextOpacity_LeafBecomesKeyframed(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "T", 1280, 720, 24, 5)
	if err != nil {
		t.Fatal(err)
	}
	tl, err := aep.NewTextLayer(comp, "TXT")
	if err != nil {
		t.Fatal(err)
	}
	if err := tl.SetText("ABCDEF"); err != nil {
		t.Fatal(err)
	}
	// Static full-select animator; then animate the Opacity leaf itself 100→0.
	if _, err := aep.AddTextOpacityAnimator(tl, 100, 0, 100, 0); err != nil {
		t.Fatalf("AddTextOpacityAnimator: %v", err)
	}
	if err := aep.AnimateTextOpacity(tl, 0, []aep.ScalarKeyframe{{Time: 0, Value: 100}, {Time: 2, Value: 0}}); err != nil {
		t.Fatalf("AnimateTextOpacity: %v", err)
	}

	_, root := writeReopen(t, p, "txopleafanim.aep")

	// The Opacity leaf is now a 2-keyframe stream...
	opKfl := findShipList(root, "ADBE Text Opacity")
	if opKfl == nil {
		t.Fatal("Opacity list not found (leaf not animated)")
	}
	lhd3 := findShipChunk(opKfl, rifx.IDLhd3)
	if lhd3 == nil {
		t.Fatal("Opacity lhd3 missing (not animated)")
	}
	if n := binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]); n != 2 {
		t.Errorf("Opacity numKf = %d, want 2", n)
	}
	// ...while the Range Offset stays static (no keyframe container).
	if offKfl := findShipList(root, "ADBE Text Percent Offset"); offKfl != nil {
		t.Error("Range Offset unexpectedly animated; this path keyframes the leaf, not the selector")
	}
}

func TestAnimateTextRotation_LeafBecomesKeyframed(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "T", 1280, 720, 24, 5)
	if err != nil {
		t.Fatal(err)
	}
	tl, err := aep.NewTextLayer(comp, "TXT")
	if err != nil {
		t.Fatal(err)
	}
	if err := tl.SetText("L"); err != nil {
		t.Fatal(err)
	}
	if _, err := aep.AddTextRotationAnimator(tl, 0, 0, 100, 0); err != nil {
		t.Fatalf("AddTextRotationAnimator: %v", err)
	}
	if err := aep.AnimateTextRotation(tl, 0, []aep.ScalarKeyframe{{Time: 0, Value: 0}, {Time: 2, Value: 90}}); err != nil {
		t.Fatalf("AnimateTextRotation: %v", err)
	}

	_, root := writeReopen(t, p, "txrotleafanim.aep")

	rotKfl := findShipList(root, "ADBE Text Rotation")
	if rotKfl == nil {
		t.Fatal("Rotation list not found (leaf not animated)")
	}
	lhd3 := findShipChunk(rotKfl, rifx.IDLhd3)
	if lhd3 == nil || binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]) != 2 {
		t.Fatalf("Rotation numKf != 2")
	}
	if offKfl := findShipList(root, "ADBE Text Percent Offset"); offKfl != nil {
		t.Error("Range Offset unexpectedly animated")
	}
}

func TestAnimateTextOpacity_RefusesWithoutAnimator(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "T", 1280, 720, 24, 5)
	if err != nil {
		t.Fatal(err)
	}
	tl, err := aep.NewTextLayer(comp, "TXT")
	if err != nil {
		t.Fatal(err)
	}
	if err := aep.AnimateTextOpacity(tl, 0, []aep.ScalarKeyframe{{Time: 0, Value: 100}, {Time: 2, Value: 0}}); err == nil {
		t.Fatal("expected refuse when no animator present, got nil")
	}
}

func TestAddTextOpacityAnimator_RefusesNonText(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "T", 1280, 720, 24, 5)
	if err != nil {
		t.Fatal(err)
	}
	nl, err := aep.NewNullLayer(comp, "NULL")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aep.AddTextOpacityAnimator(nl, 0, 0, 50, 0); err == nil {
		t.Fatal("expected refuse on non-text layer, got nil")
	}
}

func TestAddTextOpacityAnimator_TwoAnimators(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "T", 1280, 720, 24, 5)
	if err != nil {
		t.Fatal(err)
	}
	tl, err := aep.NewTextLayer(comp, "TXT")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aep.AddTextOpacityAnimator(tl, 0, 0, 50, 0); err != nil {
		t.Fatalf("first animator: %v", err)
	}
	if _, err := aep.AddTextOpacityAnimator(tl, 0, 50, 100, 0); err != nil {
		t.Fatalf("second animator: %v", err)
	}
	_, root := writeReopen(t, p, "txtwo.aep")
	animators := followingList(root, "ADBE Text Animators")
	if animators == nil {
		t.Fatal("Animators group not found")
	}
	n := 0
	for _, ch := range animators.Children {
		if ch.ID == rifx.IDTdmn && string(bytes.TrimRight(ch.Data, "\x00")) == "ADBE Text Animator" {
			n++
		}
	}
	if n != 2 {
		t.Errorf("animator count = %d, want 2", n)
	}
}
