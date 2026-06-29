// internal/aep/shape_geom_shipgate_test.go
//
// AE ship gate for batch-8b shape geometry scalar setters: Rect Roundness, Star
// Inner/Outer Roundness, Repeater Offset, Trim Offset/Start. Each writes a 1D
// shape-property cdat. The gate builds a from-scratch shape layer carrying all of
// them, has AE open + resave (acceptance), then re-parses AE's resave and asserts
// each value survived its engine (acceptance-preservation, value-checked by
// match-name — same machinery as shape_enums). ae-accept tier.
//
// Batch-23 extension — the 3 transform-group anchor/scale residuals that lower
// with template-confirmed match-names but had no AE gate: Repeater Transform
// Anchor (ADBE Vector Repeater Anchor), Wiggle Transform Anchor/Scale (ADBE
// Vector Wiggler Anchor / Scale). Same resave-preservation 2D value check. The
// sibling Position/Rotation are render-pixel gated elsewhere; these are not
// pixel-gatable cleanly (anchor shifts pivot; wiggler amplitudes need noise), so
// DOM/byte value readback is the correct ceiling.
//
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func runShapeGeomGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/shape_geom_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_shape_geom.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "GEO", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := aep.NewShapeLayer(comp, "S")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	rg := l.RootGroup()
	rect, err := rg.AddRect()
	if err != nil {
		t.Fatalf("AddRect: %v", err)
	}
	must := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	must("rect.SetSize", rect.SetSize([2]float64{300, 200}))
	must("rect.SetRoundness", rect.SetRoundness(25))
	must("rect.SetPosition", rect.SetPosition([2]float64{120, 80}))

	star, err := rg.AddStar()
	if err != nil {
		t.Fatalf("AddStar: %v", err)
	}
	must("star.SetInnerRoundness", star.SetInnerRoundness(40))
	must("star.SetOuterRoundness", star.SetOuterRoundness(60))
	must("star.SetPosition", star.SetPosition([2]float64{200, 150}))

	_, _ = rg.AddFill() // give the shapes something to render (harmless)

	rep, err := rg.AddRepeater()
	if err != nil {
		t.Fatalf("AddRepeater: %v", err)
	}
	must("rep.SetOffset", rep.SetOffset(2))
	must("rep.Transform().SetRotation", rep.Transform().SetRotation(30))
	must("rep.Transform().SetAnchor", rep.Transform().SetAnchor([2]float64{15, 25}))

	trim, err := rg.AddTrim()
	if err != nil {
		t.Fatalf("AddTrim: %v", err)
	}
	must("trim.SetStart", trim.SetStart(20))
	must("trim.SetOffset", trim.SetOffset(15))

	// Wiggle Transform: the only carrier for the Wiggler transform-group
	// anchor/scale residuals. Position/Rotation are render-pixel gated elsewhere
	// (TestMGWiggleTransform); Anchor/Scale are value-readback here.
	wt, err := rg.AddWiggleTransform()
	if err != nil {
		t.Fatalf("AddWiggleTransform: %v", err)
	}
	must("wt.Transform().SetAnchor", wt.Transform().SetAnchor([2]float64{30, 40}))
	must("wt.Transform().SetScale", wt.Transform().SetScale([2]float64{120, 80}))

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "geom_in.aep")
	resavedAEP := filepath.Join(tempDir, "geom_resaved.aep")
	doneFile := filepath.Join(tempDir, "geom.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := writeGeneratedArgs(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	t.Logf("%s shape geom AE readback: %s", ver, strings.TrimSpace(body))
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("%s shape geom ship gate FAIL:\n%s", ver, body)
	}

	root := parseAEP(t, resavedAEP)
	rd := func(b []byte, off int) float64 { return math.Float64frombits(binary.BigEndian.Uint64(b[off : off+8])) }
	// 1D scalars.
	for _, c := range []struct {
		name string
		want float64
	}{
		{"ADBE Vector Rect Roundness", 25},
		{"ADBE Vector Star Inner Roundess", 40},
		{"ADBE Vector Star Outer Roundess", 60},
		{"ADBE Vector Repeater Offset", 2},
		{"ADBE Vector Repeater Rotation", 30},
		{"ADBE Vector Trim Start", 20},
		{"ADBE Vector Trim Offset", 15},
	} {
		cdat := streamCdat(root, c.name)
		if cdat == nil {
			t.Errorf("%s resaved %q cdat missing — AE dropped the slot or wrong match-name", ver, c.name)
			continue
		}
		if got := rd(cdat, 0); math.Abs(got-c.want) > 0.01 {
			t.Errorf("%s resaved %q = %.4g, want %.4g", ver, c.name, got, c.want)
		}
	}
	// 2D points: cdat carries x at off 0, y at off 8.
	for _, c := range []struct {
		name string
		x, y float64
	}{
		{"ADBE Vector Rect Position", 120, 80},
		{"ADBE Vector Star Position", 200, 150},
		{"ADBE Vector Repeater Anchor", 15, 25},
		{"ADBE Vector Wiggler Anchor", 30, 40},
		{"ADBE Vector Wiggler Scale", 120, 80},
	} {
		cdat := streamCdat(root, c.name)
		if cdat == nil {
			t.Errorf("%s resaved %q cdat missing — AE dropped the slot or wrong match-name", ver, c.name)
			continue
		}
		if gx, gy := rd(cdat, 0), rd(cdat, 8); math.Abs(gx-c.x) > 0.01 || math.Abs(gy-c.y) > 0.01 {
			t.Errorf("%s resaved %q = [%.4g %.4g], want [%.4g %.4g]", ver, c.name, gx, gy, c.x, c.y)
		}
	}
}

func TestShapeGeom_AEShipGate_AE2020(t *testing.T) {
	runShapeGeomGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestShapeGeom_AEShipGate_AE2025(t *testing.T) {
	runShapeGeomGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
