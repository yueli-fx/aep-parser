// internal/aep/layer_3d_shadow_shipgate_test.go
//
// AE ship gate for from-scratch 3D SHADOWS (roadmap priority 2, the 阴影 half of
// the remaining Material-Options item — the biggest single piece). A from-scratch
// 3D WALL catcher + a from-scratch 3D CASTER + a from-scratch POINT light, all
// Go-built. The caster CASTS a shadow only because its Material "Casts Shadows"
// leaf is synthesized via aep.SetMaterialOption (a from-scratch shape emits an
// EMPTY Material Options group — see mutate_layer_material.go); the light's own
// Casts Shadows (already present on a from-scratch light) is turned on; the wall
// accepts shadows by default (empty material group).
//
// Per delivery-contract red line 4 the gate renders the frame and asserts on
// pixels that a shadow darkens the wall: a POINT light placed above + in front of
// the caster casts a hard shadow DOWNWARD onto the wall, a large dark region
// directly below the caster. With no ambient/fill the shadow is ~pure black while
// the surrounding wall is lit gray. A scene without the synthesized Casts Shadows
// would leave that central region LIT (≥ the off-centre wall), so a dark centre
// flanked by a lit wall is the proof the synthesis works and AE renders it.
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

func runLayer3DShadowGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/3d_shadow_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_3d_shadow.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "SHADOW3D", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	wall, _ := aep.NewShapeLayer(comp, "WALL")
	wr, _ := wall.RootGroup().AddRect()
	_ = wr.SetSize([2]float64{2400, 1500})
	wf, _ := wall.RootGroup().AddFill()
	_ = wf.SetColor([4]float64{0.6, 0.6, 0.6, 1})
	_ = wall.Position().SetStaticValue([2]float64{960, 540})

	caster, _ := aep.NewShapeLayer(comp, "CASTER")
	cr, _ := caster.RootGroup().AddRect()
	_ = cr.SetSize([2]float64{300, 300})
	cf, _ := caster.RootGroup().AddFill()
	_ = cf.SetColor([4]float64{1, 1, 1, 1})
	_ = caster.Position().SetStaticValue([2]float64{960, 400})

	if _, err := aep.NewLightLayer(comp, "Lamp"); err != nil {
		t.Fatalf("NewLightLayer: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	c := rp.Compositions[0]

	w := c.LayerByName("WALL")
	if err := w.SetIs3D(true); err != nil {
		t.Fatalf("WALL SetIs3D: %v", err)
	}
	if err := w.SetPosition([]float64{960, 540, 600}); err != nil { // behind
		t.Fatalf("WALL SetPosition: %v", err)
	}
	cs := c.LayerByName("CASTER")
	if err := cs.SetIs3D(true); err != nil {
		t.Fatalf("CASTER SetIs3D: %v", err)
	}
	if err := cs.SetPosition([]float64{960, 400, 0}); err != nil { // in front
		t.Fatalf("CASTER SetPosition: %v", err)
	}
	// The crux: synthesize the Material Casts Shadows leaf (empty from scratch).
	if _, err := aep.SetMaterialOption(cs, "ADBE Casts Shadows", float64(aep.MaterialCastsOn)); err != nil {
		t.Fatalf("SetMaterialOption(Casts Shadows): %v", err)
	}
	lamp := c.LayerByName("Lamp")
	if err := lamp.SetLightKind(aep.LightKindPoint); err != nil {
		t.Fatalf("SetLightKind: %v", err)
	}
	if err := lamp.SetPosition([]float64{960, -200, -400}); err != nil { // above + front
		t.Fatalf("Lamp SetPosition: %v", err)
	}
	if err := lamp.SetLightCastsShadows(true); err != nil {
		t.Fatalf("SetLightCastsShadows: %v", err)
	}
	if lamp.LightIntensity() != nil {
		_ = lamp.SetLightIntensity(160)
	}
	if err := aep.MoveToEnd(c.LayerByName("WALL")); err != nil {
		t.Fatalf("MoveToEnd WALL: %v", err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "shadow_in.aep")
	resavedAEP := filepath.Join(tempDir, "shadow_resaved.aep")
	doneFile := filepath.Join(tempDir, "shadow.done")
	png := filepath.Join(tempDir, "shadow.png")

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
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), toFwd(png))
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
	t.Logf("3d shadow %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("3d shadow %s ship gate FAIL:\n%s", ver, body)
	}

	f, err := os.Open(png)
	if err != nil {
		t.Fatalf("%s rendered frame missing: %v", ver, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatalf("%s decode: %v", ver, err)
	}
	// Shadow falls directly below the caster (centre-x), large + ~black; the wall
	// at the same height but off-centre stays lit (symmetric L/R, see probe).
	shadow := avgLuma(img, 800, 880, 1120, 1040)
	litLeft := avgLuma(img, 200, 880, 500, 1040)
	litRight := avgLuma(img, 1420, 880, 1720, 1040)
	t.Logf("%s shadow=%.1f litLeft=%.1f litRight=%.1f", ver, shadow, litLeft, litRight)
	if shadow > 40 {
		t.Errorf("%s no shadow: centre-below-caster luma=%.1f (expected dark) — Material Casts Shadows not rendering", ver, shadow)
	}
	if litLeft < 80 || litRight < 80 {
		t.Errorf("%s wall not lit (litLeft=%.1f litRight=%.1f) — scene/light wrong", ver, litLeft, litRight)
	}
	if litLeft-shadow < 60 {
		t.Errorf("%s insufficient shadow contrast (lit=%.1f shadow=%.1f) — shadow not distinct from lit wall", ver, litLeft, shadow)
	}

	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	rc := re.Compositions[0]
	if rc.LayerByName("CASTER") == nil || rc.LayerByName("Lamp") == nil {
		t.Fatalf("resaved CASTER/Lamp missing")
	}
}

func TestLayer3DShadow_AEShipGate_AE2020(t *testing.T) {
	runLayer3DShadowGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestLayer3DShadow_AEShipGate_AE2025(t *testing.T) {
	runLayer3DShadowGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
