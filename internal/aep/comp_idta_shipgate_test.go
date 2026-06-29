// internal/aep/comp_idta_shipgate_test.go
//
// Automated AE ship gate for the item-level / comp-flag setters that needed RE:
//   - Composition.SetComment — item cmta + idta @0x39 has-comment flag (RE
//     2026-06-17 against re_comp_idta.aep: AE silently drops a cmta unless the
//     idta @0x39 flag is set; item-level analog of layer ldta @0x3C).
//   - Composition.SetLabel   — idta @0x3A label byte (RE-confirmed correct).
//
// SetDraft3D was probed here too and dropped: its cdta @0x8A bit0 byte is
// byte-IDENTICAL to AE-native (RE re_comp_idta.aep DRAFT), yet AE reopen reads
// comp.draft3d=false — a derived/runtime state AE does not reflect from the bit
// alone (false-green family with lnrp / SetTimeRemapEnabled). Stays roundtrip.
//
// Carrier: two from-scratch comps, Reopen'd so the idta/cmta back-refs wire up
// (the item-level setters require parsed back-refs), each carrying one field.
// AE DOM readback: comp.comment / comp.label. Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

const compIdtaComment = "HELLO_RE_COMMENT"
const compIdtaLabel = 9

func buildCompIdtaDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)

	for _, n := range []string{"CMT", "LBL"} {
		if _, err := aep.NewComposition(p, n, 640, 360, 30, 2); err != nil {
			t.Fatalf("NewComposition %s: %v", n, err)
		}
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	get := func(n string) *aep.Composition {
		c := rp.CompositionByName(n)
		if c == nil {
			t.Fatalf("reopened comp %s missing", n)
		}
		return c
	}
	if err := get("CMT").SetComment(compIdtaComment); err != nil {
		t.Fatalf("SetComment: %v", err)
	}
	if err := get("LBL").SetLabel(compIdtaLabel); err != nil {
		t.Fatalf("SetLabel: %v", err)
	}
	return rp
}

func runCompIdtaGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/comp_idta_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_comp_idta.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildCompIdtaDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "comp_idta_in.aep")
	doneFile := filepath.Join(tempDir, "comp_idta.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"comment":%q,"label":%d}`,
		toFwd(inputAEP), toFwd(doneFile), compIdtaComment, compIdtaLabel)
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
	t.Logf("comp_idta %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("comp_idta %s ship gate FAIL:\n%s", ver, body)
	}
}

func TestCompIdta_AEShipGate_AE2020(t *testing.T) {
	runCompIdtaGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestCompIdta_AEShipGate_AE2025(t *testing.T) {
	runCompIdtaGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
