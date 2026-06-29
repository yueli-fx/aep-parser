// internal/aep/new_solid_null_shipgate_test.go
//
// AE ship gate for NewSolidLayer / NewNullLayer / NewAdjustmentLayer. Builds a
// fresh all-Go project (no fixture IDs anywhere — guards against the
// ID-coincidence trap) with one comp containing a Go-created solid + null +
// adjustment layer, WriteAEP, and has AE open it — proving AE ACCEPTS the
// imported solid footage Items + layers (no silent-drop / corrupt) and reads
// back the right flags, color, and dimensions. Resaves so the Go side confirms
// AE kept everything.
//
// Gated by AE_SHIP_GATE. Template embedded from test_data/fixtures/re_solidnull.aep.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func runSolidNullGate(t *testing.T, target aep.AETarget, aeExe, ver string) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/solid_null_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_solid_null.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "Main", 1280, 720, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewSolidLayer(comp, "Red BG", 1280, 720, [3]float64{0.75, 0.25, 0.5}); err != nil {
		t.Fatalf("NewSolidLayer: %v", err)
	}
	if _, err := aep.NewNullLayer(comp, "Controller"); err != nil {
		t.Fatalf("NewNullLayer: %v", err)
	}
	if _, err := aep.NewAdjustmentLayer(comp, "Grade"); err != nil {
		t.Fatalf("NewAdjustmentLayer: %v", err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "solidnull_in.aep")
	resavedAEP := filepath.Join(tempDir, "solidnull_resaved.aep")
	doneFile := filepath.Join(tempDir, "solidnull.done")

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
	t.Logf("%s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("%s solid/null/adjustment ship gate FAIL:\n%s", ver, body)
	}

	// Preservation proof: AE's resave kept all three layers with flags +
	// backing solid footage intact.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	var rc *aep.Composition
	for _, c := range re.Compositions {
		if c.Name == "Main" {
			rc = c
		}
	}
	if rc == nil {
		t.Fatalf("%s resaved: comp Main missing", ver)
	}
	var solid, null, adj *aep.Layer
	for _, l := range rc.Layers {
		switch l.Name {
		case "Red BG":
			solid = l
		case "Controller":
			null = l
		case "Grade":
			adj = l
		}
	}
	if solid == nil {
		t.Errorf("%s resaved: solid 'Red BG' missing", ver)
	}
	if null == nil || !null.IsNull {
		t.Errorf("%s resaved: null 'Controller' missing or IsNull unset (%+v)", ver, null)
	}
	if adj == nil || !adj.IsAdjust {
		t.Errorf("%s resaved: adjustment 'Grade' missing or IsAdjust unset (%+v)", ver, adj)
	}
	if solid != nil {
		var f *aep.Footage
		for _, ff := range re.Footage {
			if ff.ID == solid.SourceID {
				f = ff
			}
		}
		if f == nil || !f.IsSolid {
			t.Errorf("%s resaved: solid backing footage missing/not solid", ver)
		} else if f.Width != 1280 || f.Height != 720 {
			t.Errorf("%s resaved: solid footage dims %dx%d != 1280x720", ver, f.Width, f.Height)
		}
	}
}

func TestNewSolidNull_AEShipGate_AE2020(t *testing.T) {
	runSolidNullGate(t, aep.TargetAE2020, ae2020(), "AE2020")
}

func TestNewSolidNull_AEShipGate_AE2025(t *testing.T) {
	runSolidNullGate(t, aep.TargetAE2025, ae2025(), "AE2025")
}
