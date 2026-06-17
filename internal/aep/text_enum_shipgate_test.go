// internal/aep/text_enum_shipgate_test.go
//
// AE ship gate for batch-7c text enum setters (AE24+ run/paragraph enums:
// SetRunAutoKernType/BaselineOption/NoBreak/LineJoinType/DigitSet +
// SetParagraphAutoHyphenate/LeadingType/HangingRoman/Direction). These are
// "AE24+ ScriptingAPI writeable" — the key question this gate answers is whether
// AE (esp. AE2020) ACCEPTS and PRESERVES the Go-written btdk enum bytes through a
// resave. Verification is Go re-parse preservation on BOTH versions (the values
// survive AE's engine); AE2025 textDocument DOM is logged best-effort.
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

func runTextEnumGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/text_enum_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_text_enum.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "ENUM", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	tl, err := aep.NewTextLayer(comp, "T")
	if err != nil {
		t.Fatalf("NewTextLayer: %v", err)
	}
	if err := tl.SetText("Eg"); err != nil {
		t.Fatalf("SetText: %v", err)
	}

	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	rl := rp.Compositions[0].LayerByName("T")
	if rl == nil {
		t.Fatal("reopened text layer missing")
	}
	must := func(label string, err error) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", label, err)
		}
	}
	// Run enums (non-default values).
	must("SetRunAutoKernType", rl.SetRunAutoKernType(0, aep.TextAutoKernOptical))
	must("SetRunBaselineOption", rl.SetRunBaselineOption(0, aep.TextBaselineSuperscript))
	must("SetRunNoBreak", rl.SetRunNoBreak(0, true))
	must("SetRunLineJoinType", rl.SetRunLineJoinType(0, aep.TextLineJoinBevel))
	must("SetRunDigitSet", rl.SetRunDigitSet(0, aep.TextDigitSetHindi))
	// Paragraph enums/bools (non-default values).
	must("SetParagraphAutoHyphenate", rl.SetParagraphAutoHyphenate(0, false))
	must("SetParagraphLeadingType", rl.SetParagraphLeadingType(0, aep.TextLeadingJapanese))
	must("SetParagraphHangingRoman", rl.SetParagraphHangingRoman(0, true))
	must("SetParagraphDirection", rl.SetParagraphDirection(0, aep.TextDirectionRightToLeft))

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "enum_in.aep")
	resavedAEP := filepath.Join(tempDir, "enum_resaved.aep")
	doneFile := filepath.Join(tempDir, "enum.done")

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
	t.Logf("%s text enum AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("%s text enum ship gate FAIL:\n%s", ver, body)
	}

	// Authoritative: AE accepted + preserved the enum bytes through its resave.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	rt := re.Compositions[0].LayerByName("T")
	if rt == nil || rt.TextSource == nil || len(rt.TextSource.Runs) == 0 || len(rt.TextSource.Paragraphs) == 0 {
		t.Fatalf("%s resaved: text runs/paragraphs missing", ver)
	}
	run0 := rt.TextSource.Runs[0]
	para0 := rt.TextSource.Paragraphs[0]
	chkEnum(t, ver, "AutoKernType", int(run0.AutoKernType), int(aep.TextAutoKernOptical))
	chkEnum(t, ver, "BaselineOption", int(run0.BaselineOption), int(aep.TextBaselineSuperscript))
	chkEnumB(t, ver, "NoBreak", run0.NoBreak, true)
	chkEnum(t, ver, "LineJoinType", int(run0.LineJoinType), int(aep.TextLineJoinBevel))
	chkEnum(t, ver, "DigitSet", int(run0.DigitSet), int(aep.TextDigitSetHindi))
	chkEnumB(t, ver, "AutoHyphenate", para0.AutoHyphenate, false)
	chkEnum(t, ver, "LeadingType", int(para0.LeadingType), int(aep.TextLeadingJapanese))
	chkEnumB(t, ver, "HangingRoman", para0.HangingRoman, true)
	chkEnum(t, ver, "Direction", int(para0.Direction), int(aep.TextDirectionRightToLeft))
}

func chkEnum(t *testing.T, ver, label string, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("%s resaved: %s = %d, want %d", ver, label, got, want)
	}
}

func chkEnumB(t *testing.T, ver, label string, got, want bool) {
	t.Helper()
	if got != want {
		t.Errorf("%s resaved: %s = %v, want %v", ver, label, got, want)
	}
}

func TestTextEnum_AEShipGate_AE2020(t *testing.T) {
	runTextEnumGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestTextEnum_AEShipGate_AE2025(t *testing.T) {
	runTextEnumGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
