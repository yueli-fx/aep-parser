// internal/aep/marker_fields_shipgate_test.go
//
// AE ship gate for batch-9 Marker field setters (SetChapter / SetCuePointName /
// SetURL / SetFrameTarget / SetDuration / SetFrameDuration / SetLabel / SetTime /
// SetFrameTime). Builds a from-scratch comp with 4 comp-markers, sets the fields,
// WriteAEP, and has AE read each back from its MarkerValue DOM (matched by
// comment) + resave. These symbols are domain-tagged "comp" but are Marker
// methods; the existing marker_shipgate only covered SetComment.
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

func runMarkerFieldsGate(t *testing.T, aeExe, ver string) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/marker_fields_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_marker_fields.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	// AddMarker needs an existing marker to clone (no from-scratch seed), so open
	// the marker fixture (RE_CM, 2 markers, AE2020-openable) and mutate it.
	p, err := aep.Open("../../test_data/re_compmarker.aep")
	if err != nil {
		t.Skipf("re_compmarker.aep not present: %v", err)
	}
	var comp *aep.Composition
	for _, c := range p.Compositions {
		if c.Name == "RE_CM" {
			comp = c
		}
	}
	if comp == nil || len(comp.Markers) < 2 {
		t.Fatalf("RE_CM fixture markers missing")
	}
	fps := comp.FrameRate
	must := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	// M1 = existing marker[0] — string/scalar fields.
	m1 := comp.Markers[0]
	must("M1.SetComment", m1.SetComment("M1"))
	must("M1.SetChapter", m1.SetChapter("ch1"))
	must("M1.SetCuePointName", m1.SetCuePointName("cue1"))
	must("M1.SetURL", m1.SetURL("http://e.x"))
	must("M1.SetFrameTarget", m1.SetFrameTarget("ft1"))
	must("M1.SetDuration", m1.SetDuration(1.5))
	must("M1.SetLabel", m1.SetLabel(4))
	// M2 = existing marker[1] then SetTime(3.5).
	m2 := comp.Markers[1]
	must("M2.SetComment", m2.SetComment("M2"))
	must("M2.SetTime", m2.SetTime(3.5))
	// M3 = new marker then SetFrameTime → 4.0s.
	m3, err := aep.AddMarker(comp, 0.5)
	if err != nil {
		t.Fatalf("AddMarker M3: %v", err)
	}
	must("M3.SetComment", m3.SetComment("M3"))
	must("M3.SetFrameTime", m3.SetFrameTime(int(4.0*fps+0.5)))
	// M4 = new marker then SetFrameDuration → 1.5s.
	m4, err := aep.AddMarker(comp, 6.0)
	if err != nil {
		t.Fatalf("AddMarker M4: %v", err)
	}
	must("M4.SetComment", m4.SetComment("M4"))
	must("M4.SetFrameDuration", m4.SetFrameDuration(int(1.5*fps+0.5)))

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "mk_in.aep")
	resavedAEP := filepath.Join(tempDir, "mk_resaved.aep")
	doneFile := filepath.Join(tempDir, "mk.done")

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
	t.Logf("%s marker fields AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("%s marker fields ship gate FAIL:\n%s", ver, body)
	}

	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	var reComp *aep.Composition
	for _, c := range re.Compositions {
		if c.Name == "RE_CM" {
			reComp = c
		}
	}
	if reComp == nil || len(reComp.Markers) != 4 {
		n := -1
		if reComp != nil {
			n = len(reComp.Markers)
		}
		t.Errorf("%s resaved: RE_CM markers = %d, want 4", ver, n)
	}
}

func TestMarkerFields_AEShipGate_AE2020(t *testing.T) {
	runMarkerFieldsGate(t, ae2020(), "AE2020")
}
func TestMarkerFields_AEShipGate_AE2025(t *testing.T) {
	runMarkerFieldsGate(t, ae2025(), "AE2025")
}
