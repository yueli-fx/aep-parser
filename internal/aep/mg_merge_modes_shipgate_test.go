// internal/aep/mg_merge_modes_shipgate_test.go
//
// AE ship gate for the Merge Paths boolean modes Add / Subtract / Intersect /
// Exclude (priority-3 shape remaining,
// specs/2026-06-14-remaining-capability-roadmap.md). The enum write path already
// shipped (SetType); only Subtract was render-gated (mg_merge_shipgate_test.go).
// This gate render-verifies the remaining three modes — plus Subtract — in one
// frame.
//
// Per delivery-contract red line 4, verified at the capability's surface: a 2×2
// grid of CARDs, each = two horizontally overlapping ellipses (a Venn pair) +
// Merge(mode) + Fill on top. Each boolean mode paints a distinct pattern across
// three regions (left-only / overlap / right-only), so one rendered frame
// distinguishes all four:
//
//	             left-only  overlap  right-only
//	Add (2)        white     white    white      (union — peanut)
//	Subtract (3)   white     dark     dark       (L−R — left crescent)
//	Intersect (4)  dark      white    dark       (overlap lens)
//	Exclude (5)    white     dark     white       (XOR — two crescents)
//
// Asserting the full 3-region signature per card proves AE actually performed
// each boolean combine, not just that the enum round-tripped.
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

	aep "github.com/example/aep-parser/internal/aep"
)

type mergeModeCell struct {
	name             string
	cx, cy           int
	mode             aep.MergeType
	left, mid, right bool // expected white at left-only / overlap / right-only
}

var mergeModeCells = []mergeModeCell{
	{"ADD", 560, 320, aep.MergeTypeAdd, true, true, true},
	{"SUB", 1360, 320, aep.MergeTypeSubtract, true, false, false},
	{"INT", 560, 760, aep.MergeTypeIntersect, false, true, false},
	{"EXC", 1360, 760, aep.MergeTypeExclude, true, false, true},
}

func buildMGMergeModesDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "MERGEMODES", 1920, 1080, 30, 5)
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

	// Each cell: two overlapping ellipses (r=80, local ±50 → overlap x∈[-30,30],
	// left-only x≈-80, right-only x≈+80) + Merge(mode) + Fill on top (the
	// combine-type fill-above-merge quirk — see trim-paths-vector-filter-re.md).
	for _, c := range mergeModeCells {
		card, err := aep.NewShapeLayer(comp, c.name)
		if err != nil {
			t.Fatalf("NewShapeLayer %s: %v", c.name, err)
		}
		e1, err := card.RootGroup().AddEllipse()
		if err != nil {
			t.Fatalf("%s AddEllipse 1: %v", c.name, err)
		}
		if err := e1.SetSize([2]float64{160, 160}); err != nil {
			t.Fatalf("%s e1 size: %v", c.name, err)
		}
		if err := e1.SetPosition([2]float64{-50, 0}); err != nil {
			t.Fatalf("%s e1 pos: %v", c.name, err)
		}
		e2, err := card.RootGroup().AddEllipse()
		if err != nil {
			t.Fatalf("%s AddEllipse 2: %v", c.name, err)
		}
		if err := e2.SetSize([2]float64{160, 160}); err != nil {
			t.Fatalf("%s e2 size: %v", c.name, err)
		}
		if err := e2.SetPosition([2]float64{50, 0}); err != nil {
			t.Fatalf("%s e2 pos: %v", c.name, err)
		}
		mg, err := card.RootGroup().AddMergePaths()
		if err != nil {
			t.Fatalf("%s AddMergePaths: %v", c.name, err)
		}
		if err := mg.SetType(c.mode); err != nil {
			t.Fatalf("%s SetType: %v", c.name, err)
		}
		fill, err := card.RootGroup().AddFill()
		if err != nil {
			t.Fatalf("%s AddFill: %v", c.name, err)
		}
		if err := fill.SetColor([4]float64{1, 1, 1, 1}); err != nil {
			t.Fatalf("%s fill color: %v", c.name, err)
		}
		if err := card.Position().SetStaticValue([2]float64{float64(c.cx), float64(c.cy)}); err != nil {
			t.Fatalf("%s position: %v", c.name, err)
		}
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

func runMGMergeModesGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/mg_merge_modes_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_mg_merge_modes.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildMGMergeModesDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mg_merge_modes_in.aep")
	resavedAEP := filepath.Join(tempDir, "mg_merge_modes_resaved.aep")
	doneFile := filepath.Join(tempDir, "mg_merge_modes.done")
	framePNG := filepath.Join(tempDir, "mg_merge_modes_frame.png")

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
	t.Logf("mg merge modes %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("mg merge modes %s ship gate FAIL:\n%s", ver, body)
	}

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
	const off = 80 // left-only / right-only sample offset from card centre
	for _, c := range mergeModeCells {
		gotLeft := whiteNear(img, c.cx-off, c.cy, win)
		gotMid := whiteNear(img, c.cx, c.cy, win)
		gotRight := whiteNear(img, c.cx+off, c.cy, win)
		t.Logf("%s %s: left=%v overlap=%v right=%v (want %v/%v/%v)",
			ver, c.name, gotLeft, gotMid, gotRight, c.left, c.mid, c.right)
		if gotLeft != c.left || gotMid != c.mid || gotRight != c.right {
			t.Errorf("%s %s (Type=%d) signature [%v,%v,%v] != want [%v,%v,%v] — boolean combine wrong",
				ver, c.name, c.mode, gotLeft, gotMid, gotRight, c.left, c.mid, c.right)
		}
	}

	// Resave proof: every CARD (with its Merge filter) survives AE's re-encode.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	for _, c := range mergeModeCells {
		if re.Compositions[0].LayerByName(c.name) == nil {
			t.Errorf("resaved: %s layer missing", c.name)
		}
	}
}

func TestMGMergeModes_AEShipGate_AE2020(t *testing.T) {
	runMGMergeModesGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestMGMergeModes_AEShipGate_AE2025(t *testing.T) {
	runMGMergeModesGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
