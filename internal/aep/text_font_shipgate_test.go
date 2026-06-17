// internal/aep/text_font_shipgate_test.go
//
// AE ship gate for the text font-switch pair AddFont + SetRunFontIndex
// (verify=roundtrip). A from-scratch single-run text layer (NewTextLayer +
// SetText) makes Runs[0] a whole-doc textDocument; AddFont("ArialMT") extends
// the btdk Fonts table and SetRunFontIndex(0, idx) repoints the run, so AE reads
// back textDocument.font == "ArialMT" on both versions. A second layer "DEF"
// (no switch) logs the default font for contrast.
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

func runTextFontGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/text_font_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_text_font.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "TXTF", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	for _, n := range []string{"DEF", "SW"} {
		tl, err := aep.NewTextLayer(comp, n)
		if err != nil {
			t.Fatalf("NewTextLayer %s: %v", n, err)
		}
		if err := tl.SetText("Ag"); err != nil {
			t.Fatalf("SetText %s: %v", n, err)
		}
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	sw := rp.Compositions[0].LayerByName("SW")
	if sw == nil {
		t.Fatal("reopened SW missing")
	}
	idx, err := sw.AddFont("ArialMT")
	if err != nil {
		t.Fatalf("AddFont: %v", err)
	}
	t.Logf("%s: AddFont(ArialMT) → index %d", ver, idx)
	if err := sw.SetRunFontIndex(0, idx); err != nil {
		t.Fatalf("SetRunFontIndex: %v", err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "text_font_in.aep")
	resavedAEP := filepath.Join(tempDir, "text_font_resaved.aep")
	doneFile := filepath.Join(tempDir, "text_font.done")

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
	t.Logf("%s text font AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("%s text font ship gate FAIL:\n%s", ver, body)
	}

	// Resave proof: the switched font survives AE's re-encode + Go reparse.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	if re.Compositions[0].LayerByName("SW") == nil {
		t.Fatalf("%s resaved: text layer SW missing", ver)
	}
}

func TestTextFont_AEShipGate_AE2020(t *testing.T) {
	runTextFontGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestTextFont_AEShipGate_AE2025(t *testing.T) {
	runTextFontGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
