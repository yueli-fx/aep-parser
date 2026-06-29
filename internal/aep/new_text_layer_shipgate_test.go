// internal/aep/new_text_layer_shipgate_test.go
//
// AE ship gate for NewTextLayer. Builds a fresh project with one comp holding
// two Go-created text layers — "T1" with the template text "A" untouched, and
// "T2" after a length-preserving SetText("B") — WriteAEP, and has AE open it:
// proving AE ACCEPTS the cloned text Layr (no silent-drop / corrupt), types
// both as TextLayer, and reads back both strings. T2 doubles as the
// stale-layout-cache probe: its btdk still carries "A"'s glyph metrics, so a
// PASS certifies AE recomputes layout on load (groundwork for a future
// length-variable text write). Resaves so the Go side confirms preservation.
//
// Gated by AE_SHIP_GATE. Template extracted from test_data/fixtures/re_text.aep.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func runTextLayerGate(t *testing.T, target aep.AETarget, aeExe, ver string) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/text_layer_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_text_layer.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewTextLayer(comp, "T1"); err != nil {
		t.Fatalf("NewTextLayer T1: %v", err)
	}
	l2, err := aep.NewTextLayer(comp, "T2")
	if err != nil {
		t.Fatalf("NewTextLayer T2: %v", err)
	}
	if err := l2.SetText("B"); err != nil {
		t.Fatalf("SetText T2: %v", err)
	}

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "textlayer_in.aep")
	resavedAEP := filepath.Join(tempDir, "textlayer_resaved.aep")
	doneFile := filepath.Join(tempDir, "textlayer.done")

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
		t.Errorf("%s text-layer ship gate FAIL:\n%s", ver, body)
	}

	// Preservation proof: AE's resave kept both text layers with their strings.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	for _, want := range []struct{ name, text string }{{"T1", "A"}, {"T2", "B"}} {
		rl := textLayerByName(re, want.name)
		if rl == nil {
			t.Errorf("%s resaved: text layer %q missing", ver, want.name)
			continue
		}
		if rl.Type != aep.LayerTypeText {
			t.Errorf("%s resaved: %q Type = %v, want text", ver, want.name, rl.Type)
		}
		if rl.TextSource == nil || rl.TextSource.Text != want.text {
			t.Errorf("%s resaved: %q TextSource = %+v, want text %q", ver, want.name, rl.TextSource, want.text)
		}
	}
}

func TestNewTextLayer_AEShipGate_AE2020(t *testing.T) {
	runTextLayerGate(t, aep.TargetAE2020, ae2020(), "AE2020")
}

func TestNewTextLayer_AEShipGate_AE2025(t *testing.T) {
	runTextLayerGate(t, aep.TargetAE2025, ae2025(), "AE2025")
}
