// internal/aep/footage_idta_shipgate_test.go
//
// Automated AE ship gate for the FOOTAGE item-level Comment / Label setters:
//   - Footage.SetComment — item cmta + idta @0x39 has-comment flag.
//   - Footage.SetLabel   — idta @0x3A label byte.
//
// These share the exact serializer helpers (setItemComment / setItemLabel,
// write_item.go) that the comp path already double-version gated (comp_idta,
// 2026-06-17 RE re_comp_idta.aep: AE silently drops a cmta unless idta @0x39 is
// set, and a tail-appended cmta makes AE fail to open). The comp gate proved the
// mechanism + cmta position; this gate proves the FOOTAGE carrier specifically
// (the incident warned the Item-LIST path was unverified for footage).
//
// Carrier: one from-scratch solid footage (NewSolidLayer → "Soli" opti + Item
// LIST), Reopen'd so the idta/cmta/itemList back-refs wire up. A solid's
// Item-level Utf8 is empty (its display name lives in the opti chunk), so the
// JSX finds the single FootageItem by instanceof — NOT by name — to avoid a
// from-scratch-name false-green. AE DOM readback: footage.comment / footage.label.
// No resave (would pop AE's Save dialog, unhandled by the OCR dispatcher).
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

const footageIdtaComment = "HELLO_FOOTAGE_COMMENT"
const footageIdtaLabel = 9

func buildFootageIdtaDemo(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)

	c, err := aep.NewComposition(p, "MAIN", 640, 360, 30, 2)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewSolidLayer(c, "FTGSOLID", 200, 200, [3]float64{1, 0, 0}); err != nil {
		t.Fatalf("NewSolidLayer: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	f := rp.FootageByName("FTGSOLID")
	if f == nil {
		t.Fatalf("reopened solid footage FTGSOLID missing")
	}
	if err := f.SetComment(footageIdtaComment); err != nil {
		t.Fatalf("SetComment: %v", err)
	}
	if err := f.SetLabel(footageIdtaLabel); err != nil {
		t.Fatalf("SetLabel: %v", err)
	}
	return rp
}

func runFootageIdtaGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/footage_idta_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_footage_idta.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := buildFootageIdtaDemo(t, target)

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "footage_idta_in.aep")
	doneFile := filepath.Join(tempDir, "footage_idta.done")

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
		toFwd(inputAEP), toFwd(doneFile), footageIdtaComment, footageIdtaLabel)
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
	t.Logf("footage_idta %s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("footage_idta %s ship gate FAIL:\n%s", ver, body)
	}
}

func TestFootageIdta_AEShipGate_AE2020(t *testing.T) {
	runFootageIdtaGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestFootageIdta_AEShipGate_AE2025(t *testing.T) {
	runFootageIdtaGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
