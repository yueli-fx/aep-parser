// internal/aep/marker_shipgate_test.go
//
// AE ship gate for comp marker structural add/remove (P3 §3G). Opens the
// AE-2020-native re_compmarker.aep (RE_CM has 2 comp markers, openable by both
// AE 2020 and 2025), removes the first marker and appends a new one at t=4.0
// via Marker.Remove + Composition.AddMarker, WriteAEP, and has AE open the
// mutated file: proves AE ACCEPTS the spliced ldat / lhd3 count / mrky Nmrd
// (no data-loss / corrupt) and reads back 2 markers with the expected content.
// Resaves so the Go side confirms the markers survived a full AE round-trip.
//
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

func runMarkerShipGate(t *testing.T, aeExe string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	const base = "../../test_data/re_compmarker.aep"
	const argsPath = `e:/projects/tools/aep-parser/test_data/marker_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_marker.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	proj, err := aep.Open(base)
	if err != nil {
		t.Skipf("%s not present: %v", base, err)
	}
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "RE_CM" {
			comp = c
			break
		}
	}
	if comp == nil {
		t.Fatal("RE_CM comp missing from fixture")
	}
	if len(comp.Markers) != 2 {
		t.Fatalf("expected 2 comp markers; got %d", len(comp.Markers))
	}

	// Exercise both ops in one gate: remove the first marker (survivor is
	// "second marker" @ 2.5), then add a new marker at t=4.0 (after the
	// survivor, so AE's time ordering needs no resort — isolates "does AE
	// accept the splice" from the sort question).
	if err := comp.Markers[0].Remove(); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	nm, err := comp.AddMarker(4.0)
	if err != nil {
		t.Fatalf("AddMarker: %v", err)
	}
	if err := nm.SetComment("shipgate added"); err != nil {
		t.Fatalf("SetComment: %v", err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "marker_in.aep")
	resavedAEP := filepath.Join(tempDir, "marker_resaved.aep")
	doneFile := filepath.Join(tempDir, "marker.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := proj.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"compName":"RE_CM","expectMarkers":2,"addedComment":"shipgate added","addedTime":4.0,"survComment":"second marker"}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
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
	t.Logf("marker ship gate AE readback:\n%s", body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("marker ship gate FAIL:\n%s", body)
	}

	// Acceptance proof: AE-resaved file still carries 2 RE_CM comp markers
	// with our content.
	proj2, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("reopen resaved: %v", err)
	}
	var comp2 *aep.Composition
	for _, c := range proj2.Compositions {
		if c.Name == "RE_CM" {
			comp2 = c
			break
		}
	}
	if comp2 == nil {
		t.Fatal("resaved: RE_CM comp missing")
	}
	if len(comp2.Markers) != 2 {
		t.Fatalf("resaved RE_CM markers = %d, want 2", len(comp2.Markers))
	}
	var added, surv bool
	for _, m := range comp2.Markers {
		if m.Comment == "shipgate added" {
			added = true
		}
		if m.Comment == "second marker" {
			surv = true
		}
	}
	if !added {
		t.Error("resaved: added marker 'shipgate added' missing")
	}
	if !surv {
		t.Error("resaved: survivor 'second marker' missing")
	}
}

func TestMarker_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	runMarkerShipGate(t, aeExe)
}

func TestMarker_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	runMarkerShipGate(t, aeExe)
}
