// internal/aep/mg_set_matte_shipgate_test.go
//
// Go round-trip + AE render gate for Set Matte — a LAYER-REFERENCE effect
// (SetEffectLayerParam). AE stores the matte source as the target layer's ID in
// the effect param's tdpi (RE: re_set_matte.aep). On a 100% Go-built file:
//
//	SETMATTE comp: MATTE shape (white rect covering the LEFT half, right half
//	transparent) at the bottom + a full-frame RED solid FX on top carrying Set
//	Matte (ADBE Set Matte3) whose "Take Matte From Layer" (-0001) points at
//	MATTE. The matte gates FX by MATTE's channel: FX red shows only on the LEFT
//	(matte on), the RIGHT goes transparent → black. Without the layer ref FX
//	would be full-frame red, so sampling right=black proves the reference took.
//
// Gated by AE_SHIP_GATE. Uses test_data/generators/verify_set_matte.jsx.
package aep_test

import (
	"encoding/binary"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/rifx"
)

func buildSetMatteDemo(t *testing.T, target aep.AETarget) (*aep.Project, uint32) {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "SETMATTE", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	matte, err := aep.NewShapeLayer(comp, "MATTE")
	if err != nil {
		t.Fatalf("NewShapeLayer MATTE: %v", err)
	}
	rect, err := matte.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect: %v", err)
	}
	if err := rect.SetSize([2]float64{960, 1080}); err != nil {
		t.Fatalf("SetSize: %v", err)
	}
	fill, err := matte.RootGroup().AddFill()
	if err != nil {
		t.Fatalf("AddFill: %v", err)
	}
	if err := fill.SetColor([4]float64{1, 1, 1, 1}); err != nil {
		t.Fatalf("SetColor: %v", err)
	}
	if err := matte.Position().SetStaticValue([2]float64{480, 540}); err != nil { // left half
		t.Fatalf("Position: %v", err)
	}
	if _, err := aep.NewSolidLayer(comp, "FX", 1920, 1080, [3]float64{1, 0, 0}); err != nil {
		t.Fatalf("NewSolidLayer FX: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	c := rp.Compositions[0]
	ml := c.LayerByName("MATTE")
	fl := c.LayerByName("FX")
	if ml == nil || fl == nil {
		t.Fatal("reopened: MATTE/FX missing")
	}
	if err := aep.MoveToEnd(ml); err != nil { // matte source below FX
		t.Fatalf("MoveToEnd MATTE: %v", err)
	}
	sm, err := aep.AddEffect(fl, aep.EffectSetMatte)
	if err != nil {
		t.Fatalf("AddEffect(SetMatte): %v", err)
	}
	if err := aep.SetEffectLayerParam(fl, sm, "ADBE Set Matte3-0001", ml); err != nil {
		t.Fatalf("SetEffectLayerParam: %v", err)
	}
	return rp, ml.ID
}

// findParamTdpi returns the tdpi bytes of the first param matching name.
func findParamTdpi(root *rifx.Chunk, name string) []byte {
	var found []byte
	var walk func(c *rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		for i := 0; i+1 < len(c.Children); i++ {
			ch := c.Children[i]
			if ch.ID == rifx.IDTdmn && trimShipNUL(string(ch.Data)) == name {
				tdbs := c.Children[i+1]
				if tdbs.IsList() && tdbs.FormType == rifx.IDTdbs {
					for _, t := range tdbs.Children {
						if t.ID == rifx.IDTdpi {
							found = t.Data
							return
						}
					}
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

func TestSetMatte_GoRoundTrip(t *testing.T) {
	rp, matteID := buildSetMatteDemo(t, aep.TargetAE2020)
	dir := t.TempDir()
	fpath := filepath.Join(dir, "setmatte_rt.aep")
	out, err := os.Create(fpath)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	root := parseAEP(t, fpath)
	tdpi := findParamTdpi(root, "ADBE Set Matte3-0001")
	if len(tdpi) < 4 {
		t.Fatal("Set Matte -0001 tdpi not found in written file")
	}
	got := binary.BigEndian.Uint32(tdpi)
	if got != matteID {
		t.Errorf("Set Matte -0001 tdpi = %d, want MATTE layer ID %d", got, matteID)
	}
}

func runSetMatteGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/set_matte_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_set_matte.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	rp, _ := buildSetMatteDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "set_matte_in.aep")
	resavedAEP := filepath.Join(tempDir, "set_matte_resaved.aep")
	doneFile := filepath.Join(tempDir, "set_matte.done")
	framePNG := filepath.Join(tempDir, "set_matte_frame.png")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
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
	t.Logf("set matte %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("set matte %s ship gate FAIL:\n%s", ver, body)
	}

	f, err := os.Open(framePNG)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode frame: %v", ver, err)
	}
	const win = 12
	lr, lg, lb := avgRGB(img, 480, 540, win)
	rr, rg, rb := avgRGB(img, 1440, 540, win)
	t.Logf("%s SETMATTE left=(%d,%d,%d) right=(%d,%d,%d)", ver, lr, lg, lb, rr, rg, rb)
	if !(lr > 150 && lg < 100 && lb < 100) {
		t.Errorf("%s left=(%d,%d,%d), want RED (matte on) — Set Matte source not applied", ver, lr, lg, lb)
	}
	if !(rr < 80 && rg < 80 && rb < 80) {
		t.Errorf("%s right=(%d,%d,%d), want BLACK (matte off) — layer ref not honored (FX would be full-frame red)", ver, rr, rg, rb)
	}

	// Resave proof: the Set Matte effect survives AE's re-encode on FX.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	fl := re.Compositions[0].LayerByName("FX")
	if fl == nil {
		t.Fatal("resaved: FX missing")
	}
	hasSM := false
	for _, e := range fl.Effects {
		if e.MatchName == aep.EffectSetMatte {
			hasSM = true
		}
	}
	if !hasSM {
		t.Error("resaved: FX lost its Set Matte effect")
	}
}

func TestSetMatte_AEShipGate_AE2020(t *testing.T) {
	runSetMatteGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestSetMatte_AEShipGate_AE2025(t *testing.T) {
	runSetMatteGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
