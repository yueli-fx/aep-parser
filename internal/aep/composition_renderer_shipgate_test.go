// internal/aep/composition_renderer_shipgate_test.go
//
// AE ship gate for Composition.SetRenderer. Opens a renderer fixture (which
// carries a PRin LIST), switches the 3D engine via SetRenderer, WriteAEP, and
// has AE open the mutated file: proves AE ACCEPTS the rewritten prin + replaced
// prda (no data-loss / corrupt) and reports the renderer readback. Resaves so
// the Go side confirms the PRin LIST survived (prin/prda still present).
//
// The match-name -> engine mapping is AE-version-dependent, so the readback
// string is logged, not hard-asserted; the structural gate is acceptance.
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/rifx"
)

// rendererShipTargets is the set of engines each version run switches to.
// runRendererShipGate switches baseFixture's renderer to each target via
// SetRenderer and confirms AE opens the result (no corrupt/data-loss) with the
// PRin LIST intact. baseFixture must be openable by the AE version under test
// (AE refuses files saved by a NEWER AE), and targets must be engines that
// version exposes — see the two callers.
func runRendererShipGate(t *testing.T, aeExe, baseFixture string, targets []string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}

	srcFixture := baseFixture
	const argsPath = `e:/projects/tools/aep-parser/test_data/renderer_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_renderer.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	for _, target := range targets {
		t.Run(target, func(t *testing.T) {
			proj, err := aep.Open(srcFixture)
			if err != nil {
				t.Skipf("%s not present: %v", srcFixture, err)
			}
			comp := proj.Compositions[0]
			if err := comp.SetRenderer(target); err != nil {
				t.Fatalf("SetRenderer(%q): %v", target, err)
			}

			tempDir := t.TempDir()
			inputAEP := filepath.Join(tempDir, "renderer_in.aep")
			resavedAEP := filepath.Join(tempDir, "renderer_resaved.aep")
			doneFile := filepath.Join(tempDir, "renderer.done")

			out, err := os.Create(inputAEP)
			if err != nil {
				t.Fatal(err)
			}
			if err := proj.WriteAEP(out); err != nil {
				out.Close()
				t.Fatalf("WriteAEP: %v", err)
			}
			out.Close()

			argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"expected":%q}`,
				toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), target)
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
			t.Logf("SetRenderer(%q) AE readback:\n%s", target, body)
			if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
				t.Errorf("renderer ship gate FAIL for %q:\n%s", target, body)
			}

			// Acceptance proof: resaved file still carries prin + prda.
			root := parseAEP(t, resavedAEP)
			if findShipChunk(root, rifx.IDPrin) == nil {
				t.Errorf("resaved %q: prin chunk missing — AE dropped the PRin LIST?", target)
			}
			if findShipChunk(root, rifx.IDPrda) == nil {
				t.Errorf("resaved %q: prda chunk missing — AE dropped the PRin LIST?", target)
			}
		})
	}
}

func TestSetRenderer_AEShipGate_AE2025(t *testing.T) {
	aeExe := os.Getenv("AE2025_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2025/Support Files/AfterFX.exe`
	}
	// AE 2025 base (saved by AE 2025) + all four binary engines. AE 2025
	// auto-promotes legacy Escher/Picasso to Advanced 3D on load (still PASS).
	runRendererShipGate(t, aeExe,
		"../../test_data/renderer_classic_3d.aep",
		[]string{"ADBE Calder", "ADBE Ernst", "ADBE Picasso", "ADBE Escher"})
}

func TestSetRenderer_AEShipGate_AE2020(t *testing.T) {
	aeExe := os.Getenv("AE2020_EXE")
	if aeExe == "" {
		aeExe = `E:/adobe/Adobe After Effects 2020/Support Files/AfterFX.exe`
	}
	// AE 2020 can't open AE-2025-saved files, so use an AE-2020-native base
	// (built by test_data/build_renderer_ae2020.jsx). AE 2020 exposes Escher
	// (Advanced 3D) + Ernst (Cinema 4D); Calder/Picasso aren't AE 2020 engines.
	runRendererShipGate(t, aeExe,
		"../../test_data/renderer_ae2020_r0.aep",
		[]string{"ADBE Ernst", "ADBE Escher"})
}
