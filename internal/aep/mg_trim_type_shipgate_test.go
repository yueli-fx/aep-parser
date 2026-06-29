// internal/aep/mg_trim_type_shipgate_test.go
//
// AE ship gate for Trim Paths Trim Type (priority-3 shape remaining,
// specs/2026-06-14-remaining-capability-roadmap.md). The `ADBE Vector Trim Type`
// enum (Simultaneously=1 default / Individually=2) is AE-default-elided (no slot
// in the trim template); the serializer materializes it via synthesis-insert
// (splice the leaf into the trim body via spliceShapeLeafBeforeGroupEnd, mirroring
// Offset Copies / SetMaterialOption) when SetType selects Individually.
//
// Verified at the capability's surface per delivery-contract red line 4: a CARD
// with TWO ellipses + a white stroke + Trim End=50, Type=Individually. Individually
// trims the paths in sequence, so at End=50% the first (left) ellipse is fully
// drawn and the second (right) is empty — distinct from the default Simultaneously,
// which trims both paths in lockstep to 50% (both half-arcs). The gate asserts the
// left ellipse is a FULL ring (its LEFT edge is white — dark under Simultaneously)
// AND the right ellipse is EMPTY (its right edge is dark — white under
// Simultaneously) — proving the spliced enum actually changed AE's render, not
// just that the value round-tripped.
//
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"encoding/binary"
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

func buildMGTrimTypeDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "TRIMTYPE", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}

	bg, err := aep.NewShapeLayer(comp, "BG")
	if err != nil {
		t.Fatalf("NewShapeLayer BG: %v", err)
	}
	bgRect, err := bg.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect BG: %v", err)
	}
	if err := bgRect.SetSize([2]float64{2200, 1300}); err != nil {
		t.Fatalf("BG rect SetSize: %v", err)
	}
	bgFill, err := bg.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill BG: %v", err)
	}
	if err := bgFill.SetColor([4]float64{0.05, 0.05, 0.08, 1}); err != nil {
		t.Fatalf("BG SetColor: %v", err)
	}
	if err := bg.Position().SetStaticValue([2]float64{960, 540}); err != nil {
		t.Fatalf("BG Position: %v", err)
	}

	// Two ellipses (local ±220 → screen centres 740 / 1180) + white stroke +
	// Trim End=50, Type=Individually. Sequential trim → left ellipse full, right
	// empty (vs Simultaneously: both half).
	card, err := aep.NewShapeLayer(comp, "CARD")
	if err != nil {
		t.Fatalf("NewShapeLayer CARD: %v", err)
	}
	e1, err := card.RootGroup().AddEllipse()
	if err != nil {
		t.Fatalf("AddEllipse 1: %v", err)
	}
	if err := e1.SetSize([2]float64{200, 200}); err != nil {
		t.Fatalf("e1 SetSize: %v", err)
	}
	if err := e1.SetPosition([2]float64{-220, 0}); err != nil {
		t.Fatalf("e1 SetPosition: %v", err)
	}
	e2, err := card.RootGroup().AddEllipse()
	if err != nil {
		t.Fatalf("AddEllipse 2: %v", err)
	}
	if err := e2.SetSize([2]float64{200, 200}); err != nil {
		t.Fatalf("e2 SetSize: %v", err)
	}
	if err := e2.SetPosition([2]float64{220, 0}); err != nil {
		t.Fatalf("e2 SetPosition: %v", err)
	}
	stroke, err := card.RootGroup().AddStroke()
	if err != nil {
		t.Fatalf("AddStroke: %v", err)
	}
	if err := stroke.SetColor([4]float64{1, 1, 1, 1}); err != nil {
		t.Fatalf("stroke SetColor: %v", err)
	}
	if err := stroke.SetWidth(20); err != nil {
		t.Fatalf("stroke SetWidth: %v", err)
	}
	trim, err := card.RootGroup().AddTrim()
	if err != nil {
		t.Fatalf("AddTrim: %v", err)
	}
	if err := trim.SetEnd(50); err != nil {
		t.Fatalf("SetEnd: %v", err)
	}
	if err := trim.SetType(aep.TrimTypeIndividually); err != nil {
		t.Fatalf("SetType: %v", err)
	}
	if err := card.Position().SetStaticValue([2]float64{960, 540}); err != nil {
		t.Fatalf("CARD Position: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	if err := aep.MoveToEnd(rp.Compositions[0].LayerByName("BG")); err != nil {
		t.Fatalf("MoveToEnd BG: %v", err)
	}
	return rp
}

func runMGTrimTypeGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/mg_trim_type_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_mg_trim_type.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGTrimTypeDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_trim_type_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_trim_type_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_trim_type.done")
	framePNG := filepath.Join(tempDir, "mg_trim_type_frame.png")

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
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
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
	t.Logf("mg trim type %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg trim type %s ship gate FAIL:\n%s", ver, body)
	}

	// Render-pixel proof at frame 0. Left ellipse centre (740,540) r=100; right
	// ellipse centre (1180,540). Individually → left FULL, right EMPTY.
	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode rendered frame: %v", ver, err)
	}
	const win = 8
	// Left ellipse is a FULL ring: all 4 cardinal ring points white. The LEFT
	// edge (640) being white is the Individually discriminator — Simultaneously
	// trims the left ellipse to a right-half arc, leaving its left edge dark.
	leftRingPts := [][2]int{{640, 540}, {840, 540}, {740, 440}, {740, 640}}
	// Right ellipse is EMPTY: all 4 cardinal ring points dark. The RIGHT edge
	// (1280) being dark is the discriminator — Simultaneously draws the right
	// ellipse's right-half arc (right edge white).
	rightRingPts := [][2]int{{1080, 540}, {1280, 540}, {1180, 440}, {1180, 640}}

	leftWhite := 0
	for _, pt := range leftRingPts {
		if whiteNear(img, pt[0], pt[1], win) {
			leftWhite++
		}
	}
	rightDark := 0
	for _, pt := range rightRingPts {
		if !whiteNear(img, pt[0], pt[1], win) {
			rightDark++
		}
	}
	t.Logf("%s trim type: left-ring white=%d/4 right-ellipse dark=%d/4", ver, leftWhite, rightDark)
	if leftWhite != 4 {
		t.Errorf("%s only %d/4 left-ring points white — left ellipse not fully drawn (Individually not active?)", ver, leftWhite)
	}
	if rightDark != 4 {
		t.Errorf("%s only %d/4 right-ellipse points dark — right ellipse drawn (Simultaneously, not Individually)", ver, rightDark)
	}

	// Resave proof: the spliced Trim Type leaf survives AE's re-encode at value 2.
	root := parseAEP(t, resavedAEP)
	cdat := streamCdat(root, "ADBE Vector Trim Type")
	if len(cdat) < 8 {
		t.Fatal("resaved: Trim Type cdat missing — spliced leaf dropped")
	}
	if v := math.Float64frombits(binary.BigEndian.Uint64(cdat[:8])); math.Abs(v-2) > 0.01 {
		t.Errorf("resaved Trim Type = %g, want 2 (Individually)", v)
	}
}

func TestMGTrimType_AEShipGate_AE2020(t *testing.T) {
	runMGTrimTypeGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGTrimType_AEShipGate_AE2025(t *testing.T) {
	runMGTrimTypeGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
