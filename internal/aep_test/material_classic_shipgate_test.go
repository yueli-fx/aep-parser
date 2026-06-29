// internal/aep/material_classic_shipgate_test.go
//
// AE ship gate for batch-5a classic Material Options setters. Opens the
// AE2020-authored carrier re_material_classic_2020.aep (a 3D solid whose classic
// material props are materialized = non-elided), Go-mutates the 8 classic
// coefficients to distinctive values DIFFERENT from the builder's, WriteAEP,
// and has AE read them back from materialOption + resave so Go confirms survival.
//
// Carrier is AE2020-authored on purpose: the only other material fixture
// (re_material_options.aep) is AE25-saved and AE2020 REFUSES to open it, which
// is what blocks a double-version gate there. AE2020's classic renderer exposes
// + permits only these 8 (+ CastsShadows, already render-pixel gated); ShadowColor
// is absent and the 7 ray-traced coefficients are present-but-disabled in classic
// — those stay roundtrip (see scene_layer_property_access.go boundaries).
//
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func runMaterialClassicGate(t *testing.T, aeExe, ver string) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/material_classic_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_material_classic.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p, err := aep.Open("../../test_data/fixtures/re_material_classic_2020.aep")
	if err != nil {
		t.Skipf("re_material_classic_2020.aep not present: %v", err)
	}
	comp := p.Compositions[0]
	var solid *aep.Layer
	for _, l := range comp.Layers {
		if l.MaterialSpecular() != nil {
			solid = l
			break
		}
	}
	if solid == nil {
		t.Fatal("3D solid with material not found in carrier")
	}

	must := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	// Distinctive values, all DIFFERENT from the builder's, so each readback
	// proves the Go write took (not the builder's residue).
	must("SetMaterialLightTransmission", solid.SetMaterialLightTransmission(0.55))
	must("SetMaterialAcceptsShadows", solid.SetMaterialAcceptsShadows(true))
	must("SetMaterialAcceptsLights", solid.SetMaterialAcceptsLights(true))
	must("SetMaterialAmbient", solid.SetMaterialAmbient(0.66))
	must("SetMaterialDiffuse", solid.SetMaterialDiffuse(0.44))
	must("SetMaterialSpecular", solid.SetMaterialSpecular(0.77))
	must("SetMaterialShininess", solid.SetMaterialShininess(60))
	must("SetMaterialMetal", solid.SetMaterialMetal(0.88))

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mat_in.aep")
	resavedAEP := filepath.Join(tempDir, "mat_resaved.aep")
	doneFile := filepath.Join(tempDir, "mat.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
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
	t.Logf("%s material classic AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("%s material classic ship gate FAIL:\n%s", ver, body)
	}

	// Preservation proof: AE's resave kept the Go-written coefficients.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	var reSolid *aep.Layer
	for _, l := range re.Compositions[0].Layers {
		if l.MaterialSpecular() != nil {
			reSolid = l
			break
		}
	}
	if reSolid == nil {
		t.Fatalf("%s resaved: material solid missing", ver)
	}
	chkMat(t, ver, "LightTransmission", reSolid.MaterialLightTransmission(), 0.55)
	// AcceptsShadows/AcceptsLights were Go-set to 1 (= AE's default for both),
	// so AE legitimately ELIDES them on resave and Go re-parse finds them absent.
	// That is correct AE behaviour, not a write failure — the JSX DOM readback
	// (=1, differing from the carrier's 0) is the authoritative proof the Go
	// write took. So we don't assert Go-parse survival for these two bools.
	chkMat(t, ver, "Ambient", reSolid.MaterialAmbient(), 0.66)
	chkMat(t, ver, "Diffuse", reSolid.MaterialDiffuse(), 0.44)
	chkMat(t, ver, "Specular", reSolid.MaterialSpecular(), 0.77)
	chkMat(t, ver, "Shininess", reSolid.MaterialShininess(), 60)
	chkMat(t, ver, "Metal", reSolid.MaterialMetal(), 0.88)
}

func chkMat(t *testing.T, ver, label string, pr *aep.Property, want float64) {
	t.Helper()
	if pr == nil {
		t.Errorf("%s resaved: %s property nil", ver, label)
		return
	}
	got, ok := pr.StaticValue.(float64)
	if !ok || abs(got-want) > 1e-3 {
		t.Errorf("%s resaved: %s = %v, want %g", ver, label, pr.StaticValue, want)
	}
}

func TestMaterialClassic_AEShipGate_AE2020(t *testing.T) {
	runMaterialClassicGate(t, ae2020(), "AE2020")
}

func TestMaterialClassic_AEShipGate_AE2025(t *testing.T) {
	runMaterialClassicGate(t, ae2025(), "AE2025")
}
