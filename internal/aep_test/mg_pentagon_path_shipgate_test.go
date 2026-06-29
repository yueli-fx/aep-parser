// internal/aep/mg_pentagon_path_shipgate_test.go
//
// AE ship gate for >4-vertex path geometry (priority-1 leftover:
// "path 几何 lhd3 >4 顶点分页"). All prior path/mask gates used ≤4 vertices
// (rect 4 / triangle 3); add-mask-create-re.md hypothesized AE's geometry lhd3
// capacity fields (@0x0C / @0x1C) follow nextPow2(n) — which would make a
// 5-vertex path (cap between pow2 4 and 8) emit the wrong capacity.
//
// Empirically DISPROVEN (this gate): a Go-built n=5 pentagon — both as a MASK
// (encodeBezier + mask-strictness patch, @0x14=4/@0x1C=4·n=20) and as a SHAPE
// path ("ADBE Vector Shape", encodeBezier raw, @0x14=n=5/@0x1C=16 const) — is
// accepted by AE 2020 + AE 2025, reads back 5 vertices, renders the pentagon
// silhouette, and AE resaves the geometry lhd3 BYTE-IDENTICAL (verified during
// RE). AE uses cap=n, not nextPow2(n); shape and mask genuinely differ at
// @0x14/@0x1C but our encoder matches each. So >4-vertex paths need no fix —
// this gate pins the finding against regression.
//
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"image"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func pentagonVerts(radius float64) [][2]float64 {
	var v [][2]float64
	for k := range 5 {
		rad := (-90.0 + float64(k)*72.0) * math.Pi / 180.0
		v = append(v, [2]float64{radius * math.Cos(rad), radius * math.Sin(rad)})
	}
	return v
}

func buildMGPentagonDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "PENTA", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	bg, _ := aep.NewShapeLayer(comp, "BG")
	bgRect, _ := bg.RootGroup().AddRect()
	bgRect.SetSize([2]float64{2200, 1300})
	bgFill, _ := bg.RootGroup().AddFill()
	bgFill.SetColor([4]float64{0.05, 0.05, 0.08, 1})
	bg.Position().SetStaticValue([2]float64{960, 540})

	// CARD: white rect masked by a 5-vertex pentagon (encodeBezier + mask patch).
	card, _ := aep.NewShapeLayer(comp, "CARD")
	r, _ := card.RootGroup().AddRect()
	r.SetSize([2]float64{600, 600})
	cf, _ := card.RootGroup().AddFill()
	cf.SetColor([4]float64{1, 1, 1, 1})
	card.Position().SetStaticValue([2]float64{960, 540})

	// PENTASHAPE: a 5-vertex SHAPE path (encodeBezier raw, no mask patch).
	sh, _ := aep.NewShapeLayer(comp, "PENTASHAPE")
	pp, _ := sh.RootGroup().AddPath()
	if err := pp.SetVertices(pentagonVerts(220)); err != nil {
		t.Fatalf("SetVertices: %v", err)
	}
	pp.SetClosed(true)
	shFill, _ := sh.RootGroup().AddFill()
	shFill.SetColor([4]float64{1, 0.4, 0.1, 1})
	sh.Position().SetStaticValue([2]float64{1500, 540})

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	if err := aep.MoveToEnd(rp.Compositions[0].LayerByName("BG")); err != nil {
		t.Fatalf("MoveToEnd: %v", err)
	}
	l := rp.Compositions[0].LayerByName("CARD")
	m, err := aep.AddMask(l, "M", aep.BezierPath{Vertices: pentagonVerts(280), Closed: true})
	if err != nil {
		t.Fatalf("AddMask pentagon: %v", err)
	}
	if len(m.Vertices) != 5 {
		t.Fatalf("scene mask verts=%d, want 5", len(m.Vertices))
	}
	return rp
}

func runMGPentagonGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/mg_pentagon_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_mg_pentagon.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGPentagonDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "penta_in.aep")
	resavedAEP := filepath.Join(tempDir, "penta_resaved.aep")
	doneFile := filepath.Join(tempDir, "penta.done")
	framePNG := filepath.Join(tempDir, "penta.png")

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

	body, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("PENTA %s AE readback:\n%s", ver, string(body))
	if lines := strings.SplitN(string(body), "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("PENTA %s ship gate FAIL:\n%s", ver, string(body))
	}

	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, derr := image.Decode(f)
	f.Close()
	if derr != nil {
		t.Fatalf("%s decode frame: %v", ver, derr)
	}
	const win = 8
	// Pentagon mask centre (960,540) revealed white; a rect corner (690,270) is
	// inside the old 600×600 rect but outside the pentagon → masked dark.
	maskCentreWhite := whiteNear(img, 960, 540, win)
	maskCornerDark := !whiteNear(img, 690, 270, win)
	t.Logf("%s pentagon: mask centre white=%v corner dark=%v", ver, maskCentreWhite, maskCornerDark)
	if !maskCentreWhite {
		t.Errorf("%s pentagon mask centre not white — mask dropped/wrong", ver)
	}
	if !maskCornerDark {
		t.Errorf("%s pentagon mask corner not dark — mask not a pentagon (5-vertex geometry wrong)", ver)
	}

	// Resave proof: both 5-vertex paths survive AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	rl := re.Compositions[0].LayerByName("CARD")
	if rl == nil || len(rl.Masks) != 1 {
		t.Fatal("resaved: CARD mask missing")
	}
	if got := len(rl.Masks[0].Vertices); got != 5 {
		t.Errorf("resaved mask vertices = %d, want 5 (pentagon)", got)
	}
}

func TestMGPentagonPath_AEShipGate_AE2020(t *testing.T) {
	runMGPentagonGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGPentagonPath_AEShipGate_AE2025(t *testing.T) {
	runMGPentagonGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
