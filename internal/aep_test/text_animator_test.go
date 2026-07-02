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

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
	"github.com/yueli-fx/aep-parser/internal/rifx"
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

func TestAnimateTextFillOpacity_LeafBecomesKeyframed(t *testing.T) {
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
	if _, err := aep.AddTextFillOpacityAnimator(tl, 100, 0, 100, 0); err != nil {
		t.Fatalf("AddTextFillOpacityAnimator: %v", err)
	}
	if err := aep.AnimateTextFillOpacity(tl, 0, []aep.ScalarKeyframe{{Time: 0, Value: 100}, {Time: 2, Value: 25}}); err != nil {
		t.Fatalf("AnimateTextFillOpacity: %v", err)
	}

	reopened, root := writeReopen(t, p, "txfillopleafanim.aep")

	fillOpacityKfl := findShipList(root, "ADBE Text Fill Opacity")
	if fillOpacityKfl == nil {
		t.Fatal("Fill Opacity list not found (leaf not animated)")
	}
	lhd3 := findShipChunk(fillOpacityKfl, rifx.IDLhd3)
	if lhd3 == nil {
		t.Fatal("Fill Opacity lhd3 missing (not animated)")
	}
	if n := binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]); n != 2 {
		t.Errorf("Fill Opacity numKf = %d, want 2", n)
	}
	if offKfl := findShipList(root, "ADBE Text Percent Offset"); offKfl != nil {
		t.Error("Range Offset unexpectedly animated; this path keyframes the leaf, not the selector")
	}
	prof, err := profile.Build(reopened, profile.Options{Path: "txfillopleafanim.aep"})
	if err != nil {
		t.Fatalf("profile.Build: %v", err)
	}
	if prop := profileLayerPropertyWithKeyframes(prof, "ADBE Text Fill Opacity"); prop == nil || len(prop.Keyframes) != 2 {
		t.Fatalf("profile Fill Opacity keyframes = %+v, want two keyframes", prop)
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

func profileLayerPropertyWithKeyframes(prof *profile.Profile, matchName string) *profile.Property {
	if prof == nil {
		return nil
	}
	for ci := range prof.Comps {
		for li := range prof.Comps[ci].Layers {
			for pi := range prof.Comps[ci].Layers[li].Properties {
				prop := &prof.Comps[ci].Layers[li].Properties[pi]
				if prop.MatchName == matchName && len(prop.Keyframes) > 0 {
					return prop
				}
			}
		}
	}
	return nil
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

// kflKeyframeValues returns the dim BE f64 values of keyframe kfIndex from a
// keyframe container (LIST list/kfl), reading at the spatial value offset 0x38
// within the keyframe's bpk-sized record.
func kflKeyframeValues(t *testing.T, kfl *rifx.Chunk, kfIndex, dim int) []float64 {
	return kflKeyframeValuesAt(t, kfl, kfIndex, 0x38, dim)
}

// kflKeyframeValuesAt reads dim BE f64 of keyframe kfIndex at valueOff within the
// keyframe's bpk-sized record (0x38 for the spatial block, 0x08 for non-spatial).
func kflKeyframeValuesAt(t *testing.T, kfl *rifx.Chunk, kfIndex, valueOff, dim int) []float64 {
	t.Helper()
	lhd3 := findShipChunk(kfl, rifx.IDLhd3)
	ldat := findShipChunk(kfl, rifx.IDLdat)
	if lhd3 == nil || ldat == nil {
		t.Fatal("kfl missing lhd3/ldat")
	}
	bpk := int(binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]))
	base := kfIndex*bpk + valueOff
	out := make([]float64, dim)
	for i := range out {
		off := base + i*8
		if off+8 > len(ldat.Data) {
			t.Fatalf("ldat too short for kf %d comp %d (off %d, len %d)", kfIndex, i, off, len(ldat.Data))
		}
		out[i] = math.Float64frombits(binary.BigEndian.Uint64(ldat.Data[off : off+8]))
	}
	return out
}

func kflBpk(t *testing.T, kfl *rifx.Chunk) int {
	t.Helper()
	lhd3 := findShipChunk(kfl, rifx.IDLhd3)
	if lhd3 == nil {
		t.Fatal("kfl missing lhd3")
	}
	return int(binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]))
}

func TestAnimateTextPosition_LeafBecomesKeyframed(t *testing.T) {
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
	if _, err := aep.AddTextPositionAnimator(tl, 0, 0, 0, 0, 100, 0); err != nil {
		t.Fatalf("AddTextPositionAnimator: %v", err)
	}
	if err := aep.AnimateTextPosition(tl, 0, []aep.VectorKeyframe{
		{Time: 0, Value: []float64{0, 0, 0}},
		{Time: 2, Value: []float64{100, 50, 0}},
	}); err != nil {
		t.Fatalf("AnimateTextPosition: %v", err)
	}

	_, root := writeReopen(t, p, "txposleafanim.aep")

	kfl := findShipList(root, "ADBE Text Position 3D")
	if kfl == nil {
		t.Fatal("Position list not found (leaf not animated)")
	}
	// AE ground truth (re_text_animator_animatedvec): spatial 3D block, bpk 128,
	// value @ 0x38.
	if bpk := kflBpk(t, kfl); bpk != 128 {
		t.Errorf("Position bpk = %d, want 128", bpk)
	}
	if got := kflKeyframeValues(t, kfl, 1, 3); got[0] != 100 || got[1] != 50 || got[2] != 0 {
		t.Errorf("Position kf1 value = %v, want [100 50 0]", got)
	}
	if findShipList(root, "ADBE Text Percent Offset") != nil {
		t.Error("Range Offset unexpectedly animated")
	}
}

func TestAnimateTextColor_LeafBecomesKeyframed(t *testing.T) {
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
	// red → blue
	if err := aep.AnimateTextColor(tl, 0, []aep.VectorKeyframe{
		{Time: 0, Value: []float64{1, 0, 0, 1}},
		{Time: 2, Value: []float64{0, 0, 1, 1}},
	}); err != nil {
		t.Fatalf("AnimateTextColor: %v", err)
	}

	_, root := writeReopen(t, p, "txcolorleafanim.aep")

	kfl := findShipList(root, "ADBE Text Fill Color")
	if kfl == nil {
		t.Fatal("Fill Color list not found (leaf not animated)")
	}
	// AE ground truth: 4-channel color block, bpk 152, value @ 0x38 as [A,R,G,B]×255.
	if bpk := kflBpk(t, kfl); bpk != 152 {
		t.Errorf("Color bpk = %d, want 152", bpk)
	}
	// kf0 = opaque red → [255,255,0,0]; kf1 = opaque blue → [255,0,0,255].
	if got := kflKeyframeValues(t, kfl, 0, 4); got[0] != 255 || got[1] != 255 || got[2] != 0 || got[3] != 0 {
		t.Errorf("Color kf0 = %v, want [255 255 0 0]", got)
	}
	if got := kflKeyframeValues(t, kfl, 1, 4); got[0] != 255 || got[1] != 0 || got[2] != 0 || got[3] != 255 {
		t.Errorf("Color kf1 = %v, want [255 0 0 255]", got)
	}
}

func TestAnimateTextScale_LeafBecomesKeyframed(t *testing.T) {
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
	if _, err := aep.AddTextScaleAnimator(tl, 100, 100, 100, 0, 100, 0); err != nil {
		t.Fatalf("AddTextScaleAnimator: %v", err)
	}
	if err := aep.AnimateTextScale(tl, 0, []aep.VectorKeyframe{
		{Time: 0, Value: []float64{100, 100, 100}},
		{Time: 2, Value: []float64{200, 200, 100}},
	}); err != nil {
		t.Fatalf("AnimateTextScale: %v", err)
	}

	_, root := writeReopen(t, p, "txscaleleafanim.aep")

	kfl := findShipList(root, "ADBE Text Scale 3D")
	if kfl == nil {
		t.Fatal("Scale list not found (leaf not animated)")
	}
	// AE ground truth (re_text_animator_animatedvec): NON-spatial 3D block, bpk
	// 128, value @ 0x08 (NOT the spatial 0x38).
	if bpk := kflBpk(t, kfl); bpk != 128 {
		t.Errorf("Scale bpk = %d, want 128", bpk)
	}
	if got := kflKeyframeValuesAt(t, kfl, 0, 0x08, 3); got[0] != 100 || got[1] != 100 || got[2] != 100 {
		t.Errorf("Scale kf0 value@0x08 = %v, want [100 100 100]", got)
	}
	if got := kflKeyframeValuesAt(t, kfl, 1, 0x08, 3); got[0] != 200 || got[1] != 200 || got[2] != 100 {
		t.Errorf("Scale kf1 value@0x08 = %v, want [200 200 100]", got)
	}
	// The spatial value slot 0x38 must stay zero (proves non-spatial layout).
	if got := kflKeyframeValuesAt(t, kfl, 0, 0x38, 3); got[0] != 0 || got[1] != 0 || got[2] != 0 {
		t.Errorf("Scale value unexpectedly at spatial 0x38 = %v (should be non-spatial @0x08)", got)
	}
	if findShipList(root, "ADBE Text Percent Offset") != nil {
		t.Error("Range Offset unexpectedly animated")
	}
}

func TestAnimateTextColor_RefusesBadChannelCount(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "T", 1280, 720, 24, 5)
	if err != nil {
		t.Fatal(err)
	}
	tl, err := aep.NewTextLayer(comp, "TXT")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aep.AddTextColorAnimator(tl, 1, 0, 0, 1, 0, 100, 0); err != nil {
		t.Fatal(err)
	}
	if err := aep.AnimateTextColor(tl, 0, []aep.VectorKeyframe{
		{Time: 0, Value: []float64{1, 0, 0}},
		{Time: 2, Value: []float64{0, 0, 1}},
	}); err == nil {
		t.Fatal("expected refuse on 3-channel color keyframe, got nil")
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
