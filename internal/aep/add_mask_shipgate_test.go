// internal/aep/add_mask_shipgate_test.go
//
// AE ship gate for AddMask on the Mask Parade (verify_mask.jsx). Two scenarios
// per AE version:
//
//   - baseline: AE-2020-native fixture layer (3 effects, no masks) + Go AddMask
//     — proves AE accepts the from-scratch atom triple spliced next to real AE
//     sibling groups AND the auto-created Mask Parade ordered before the
//     existing Effect Parade. AE reads back name / mode / inverted / closed /
//     vertices (layer px, eps 0.5) and the untouched effect count.
//   - autogo: 100% Go-built file (NewProject → NewComposition → NewShapeLayer →
//     Reopen → AddMask ×2) — closed rect + open polyline, indexes 1 / 2, both
//     read back and kept across AE's own resave.
//
// Gated by AE_SHIP_GATE. Baseline built by test_data/generators/re_property_struct.jsx.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

type gateMaskExpect struct {
	name     string
	closed   bool
	inverted bool
	vertices [][2]float64
}

func maskArgsJSON(input, done, resaved string, effectCount int, masks []gateMaskExpect) string {
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }
	var ms []string
	for _, m := range masks {
		var vs []string
		for _, v := range m.vertices {
			vs = append(vs, fmt.Sprintf("[%g,%g]", v[0], v[1]))
		}
		ms = append(ms, fmt.Sprintf(`{"name":%q,"closed":%v,"inverted":%v,"vertices":[%s]}`,
			m.name, m.closed, m.inverted, strings.Join(vs, ",")))
	}
	return fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"effectCount":%d,"masks":[%s]}`,
		toFwd(input), toFwd(done), toFwd(resaved), effectCount, strings.Join(ms, ","))
}

// runMaskShipGate writes proj to a temp .aep, has AE verify it against expect
// via verify_mask.jsx, then re-opens AE's resave and asserts the masks
// survived AE's own re-encoding (count + names + closed flags).
func runMaskShipGate(t *testing.T, aeExe, label string, proj *aep.Project, layerID uint32, effectCount int, expect []gateMaskExpect) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/mask_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_mask.jsx`

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, label+"_in.aep")
	resavedAEP := filepath.Join(tempDir, label+"_resaved.aep")
	doneFile := filepath.Join(tempDir, label+".done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := proj.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	if err := writeGeneratedArgs(argsPath, []byte(maskArgsJSON(inputAEP, doneFile, resavedAEP, effectCount, expect)), 0644); err != nil {
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
	t.Logf("%s AE readback:\n%s", label, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("%s ship gate FAIL:\n%s", label, body)
	}

	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	rl := maskParadeLayer(re, layerID)
	if rl == nil {
		t.Fatal("resaved: layer not found")
	}
	if len(rl.Masks) != len(expect) {
		t.Fatalf("resaved masks = %d, want %d (AE dropped masks)", len(rl.Masks), len(expect))
	}
	for i, m := range rl.Masks {
		if m.Name != expect[i].name {
			t.Errorf("resaved mask %d name = %q, want %q", i, m.Name, expect[i].name)
		}
		if m.Closed != expect[i].closed {
			t.Errorf("resaved mask %d closed = %v, want %v", i, m.Closed, expect[i].closed)
		}
		if len(m.Vertices) != len(expect[i].vertices) {
			t.Errorf("resaved mask %d vertices = %d, want %d", i, len(m.Vertices), len(expect[i].vertices))
		}
	}
}

func runMaskBaselineGate(t *testing.T, aeExe, ver string) {
	proj, err := aep.Open("../../test_data/generated/fixtures/re_property_struct_baseline.aep")
	if err != nil {
		t.Skipf("re_property_struct_baseline.aep not present: %v", err)
	}
	l := layerWithEffects(proj)
	if l == nil {
		t.Fatal("no layer with effects in baseline")
	}
	if _, err := aep.AddMask(l, "Gate Mask", rectPath()); err != nil {
		t.Fatalf("AddMask: %v", err)
	}
	expect := []gateMaskExpect{{
		name:     "Gate Mask",
		closed:   true,
		vertices: [][2]float64{{10, 10}, {190, 10}, {190, 190}, {10, 190}},
	}}
	runMaskShipGate(t, aeExe, "maskbaseline-"+ver, proj, l.ID, 3, expect)
}

func runMaskAutoGoGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aep.NewShapeLayer(comp, "S"); err != nil {
		t.Fatal(err)
	}
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	var l *aep.Layer
	for _, c := range rp.Compositions {
		for _, cl := range c.Layers {
			if cl.Name == "S" {
				l = cl
			}
		}
	}
	if l == nil {
		t.Fatal("reopened project: layer S not found")
	}
	if l.MaskParade() != nil {
		t.Fatal("fresh shape layer should have no Mask Parade before AddMask")
	}
	if _, err := aep.AddMask(l, "Region", rectPath()); err != nil {
		t.Fatalf("AddMask #1: %v", err)
	}
	openPath := aep.BezierPath{Vertices: [][2]float64{{100, 100}, {960, 540}, {1820, 100}}, Closed: false}
	m2, err := aep.AddMask(l, "Polyline", openPath)
	if err != nil {
		t.Fatalf("AddMask #2: %v", err)
	}
	if m2.Index != 2 {
		t.Fatalf("second mask index = %d, want 2", m2.Index)
	}
	expect := []gateMaskExpect{
		{name: "Region", closed: true, vertices: rectPath().Vertices},
		{name: "Polyline", closed: false, vertices: openPath.Vertices},
	}
	runMaskShipGate(t, aeExe, "maskautogo-"+ver, rp, l.ID, 0, expect)
}

func TestAddMask_AEShipGate_AE2020(t *testing.T) { runMaskBaselineGate(t, ae2020(), "AE2020") }
func TestAddMask_AEShipGate_AE2025(t *testing.T) { runMaskBaselineGate(t, ae2025(), "AE2025") }

func TestAddMaskAutoGo_AEShipGate_AE2020(t *testing.T) {
	runMaskAutoGoGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestAddMaskAutoGo_AEShipGate_AE2025(t *testing.T) {
	runMaskAutoGoGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
