// internal/aep/mg_precomp_shipgate_test.go
//
// AE ship gate for precomp / nested-composition layers from scratch (MG roadmap
// S4, specs/2026-06-12-from-scratch-mg-roadmap.md). Verified at the capability's
// surface per delivery-contract red line 4: the parent comp's ONLY layer is a
// precomp of the child comp, and the gate renders the parent frame to assert the
// child comp's content (a green square) shows through.
//
// Coincidence-proof (gate-fixture-ID-coincidence scar): the parent comp is
// created BEFORE the child, so the child's item ID is NOT the embedded
// template's stale SourceID (1). If SetSource silently no-op'd, the precomp
// would point at the wrong item (or self-reference) and no green would render.
//
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func buildMGPrecompDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)

	// Parent created FIRST → low item ID; child created after → its ID differs
	// from the precomp template's stale SourceID, making the gate coincidence-proof.
	parent, err := aep.NewComposition(p, "MGPREC_Parent", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition parent: %v", err)
	}
	child, err := aep.NewComposition(p, "MGPREC_Child", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition child: %v", err)
	}

	// Child content: dark-blue full-frame BG + a centred green square. Only the
	// child comp contains green, so green in the parent render proves the precomp
	// resolved to the child.
	cbg, err := aep.NewShapeLayer(child, "ChildBG")
	if err != nil {
		t.Fatalf("NewShapeLayer ChildBG: %v", err)
	}
	cbgRect, err := cbg.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("ChildBG AddRect: %v", err)
	}
	if err := cbgRect.SetSize([2]float64{2200, 1300}); err != nil {
		t.Fatalf("ChildBG SetSize: %v", err)
	}
	cbgFill, err := cbg.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("ChildBG AddFill: %v", err)
	}
	if err := cbgFill.SetColor([4]float64{0.04, 0.06, 0.18, 1}); err != nil {
		t.Fatalf("ChildBG SetColor: %v", err)
	}
	if err := cbg.Position().SetStaticValue([2]float64{960, 540}); err != nil {
		t.Fatalf("ChildBG Position: %v", err)
	}

	green, err := aep.NewShapeLayer(child, "ChildGreen")
	if err != nil {
		t.Fatalf("NewShapeLayer ChildGreen: %v", err)
	}
	gRect, err := green.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("ChildGreen AddRect: %v", err)
	}
	if err := gRect.SetSize([2]float64{600, 600}); err != nil {
		t.Fatalf("ChildGreen SetSize: %v", err)
	}
	gFill, err := green.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("ChildGreen AddFill: %v", err)
	}
	if err := gFill.SetColor([4]float64{0.1, 0.9, 0.2, 1}); err != nil {
		t.Fatalf("ChildGreen SetColor: %v", err)
	}
	if err := green.Position().SetStaticValue([2]float64{960, 540}); err != nil {
		t.Fatalf("ChildGreen Position: %v", err)
	}

	// Reopen so the comps are fully parsed (NewPrecompLayer needs itemList
	// back-refs), then nest child into parent.
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	var parentR, childR *aep.Composition
	for _, c := range rp.Compositions {
		switch c.Name {
		case "MGPREC_Parent":
			parentR = c
		case "MGPREC_Child":
			childR = c
		}
	}
	if parentR == nil || childR == nil {
		t.Fatalf("reopened comps missing (parent=%v child=%v)", parentR, childR)
	}
	if childR.ID <= 1 {
		t.Fatalf("child comp ID=%d not coincidence-proof (== template stale SourceID)", childR.ID)
	}
	// Push the child BG to the bottom so the green square renders on top
	// (first-added layer is otherwise at the top of the stack).
	if err := aep.MoveToEnd(childR.LayerByName("ChildBG")); err != nil {
		t.Fatalf("MoveToEnd ChildBG: %v", err)
	}
	if _, err := aep.NewPrecompLayer(parentR, childR, "NESTED"); err != nil {
		t.Fatalf("NewPrecompLayer: %v", err)
	}
	_ = parent
	_ = child
	return rp
}

// solidColorAt samples a small window around (px,py) and reports whether the
// average pixel is within tol of want (per-channel, 0..255).
func solidColorAt(img image.Image, px, py int, want [3]uint8, tol int) bool {
	b := img.Bounds()
	if px < b.Min.X || px >= b.Max.X || py < b.Min.Y || py >= b.Max.Y {
		return false
	}
	r, g, bb, _ := img.At(px, py).RGBA()
	got := [3]int{int(r >> 8), int(g >> 8), int(bb >> 8)}
	for i := range 3 {
		d := got[i] - int(want[i])
		if d < -tol || d > tol {
			return false
		}
	}
	return true
}

func runMGPrecompGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/mg_precomp_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_mg_precomp.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGPrecompDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_precomp_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_precomp_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_precomp.done")
	framePNG := filepath.Join(tempDir, "mg_precomp_frame.png")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"png":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), toFwd(framePNG))
	if err := writeGeneratedArgs(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 240)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	t.Logf("mg precomp %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg precomp %s ship gate FAIL:\n%s", ver, body)
	}

	// Render-pixel proof at frame 0: the child comp's green square (centre) and
	// its dark-blue BG (corner) must show through the precomp layer.
	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode rendered frame: %v", ver, err)
	}
	const tol = 45
	greenCentre := solidColorAt(img, 960, 540, [3]uint8{26, 230, 51}, tol)
	blueCorner := solidColorAt(img, 200, 200, [3]uint8{10, 15, 46}, tol)
	t.Logf("%s pixels: green-centre=%v blue-corner=%v", ver, greenCentre, blueCorner)
	if !greenCentre {
		t.Errorf("%s child green square not visible at parent centre — precomp did not render the child comp", ver)
	}
	if !blueCorner {
		t.Errorf("%s child dark-blue BG not visible at parent corner — precomp not rendering full child frame", ver)
	}

	// Resave proof: the precomp layer's source still resolves to the child comp.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	var parentRe *aep.Composition
	for _, c := range re.Compositions {
		if c.Name == "MGPREC_Parent" {
			parentRe = c
		}
	}
	if parentRe == nil || len(parentRe.Layers) != 1 {
		t.Fatalf("resaved: parent missing or wrong layer count (%v)", parentRe)
	}
	src := parentRe.Layers[0].SourceComposition()
	if src == nil || src.Name != "MGPREC_Child" {
		t.Errorf("resaved: precomp source = %v, want MGPREC_Child", src)
	}
}

func TestMGPrecomp_AEShipGate_AE2020(t *testing.T) {
	runMGPrecompGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGPrecomp_AEShipGate_AE2025(t *testing.T) {
	runMGPrecompGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
