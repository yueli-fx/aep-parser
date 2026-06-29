// internal/aep/text_para_shipgate_test.go
//
// AE ship gate for batch-7b paragraph SetParagraph* setters. A from-scratch
// single-paragraph text layer surfaces paragraph attributes as whole-document
// textDocument properties. The indent/spacing DOM (firstLineIndent / leftMargin /
// rightMargin / spaceBefore / spaceAfter) is AE2022+; AE2020 falls back to Go
// resave-preservation. This gate's discovery run exposed (and this commit fixes)
// the SetParagraph{FirstLineIndent,StartIndent,EndIndent,SpaceBefore,SpaceAfter}
// FormatPSReal bug — a bare-int point value read as value/65536 (same class as
// the SetRunTsume bug).
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

func runTextParaGate(t *testing.T, aeExe, ver string, target aep.AETarget) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/generated/args/text_para_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/generators/verify_text_para.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "PARA", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	tl, err := aep.NewTextLayer(comp, "T")
	if err != nil {
		t.Fatalf("NewTextLayer: %v", err)
	}
	if err := tl.SetText("Para"); err != nil {
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
	must("SetParagraphFirstLineIndent", rl.SetParagraphFirstLineIndent(0, 20))
	must("SetParagraphStartIndent", rl.SetParagraphStartIndent(0, 15))
	must("SetParagraphEndIndent", rl.SetParagraphEndIndent(0, 10))
	must("SetParagraphSpaceBefore", rl.SetParagraphSpaceBefore(0, 25))
	must("SetParagraphSpaceAfter", rl.SetParagraphSpaceAfter(0, 30))

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "para_in.aep")
	resavedAEP := filepath.Join(tempDir, "para_resaved.aep")
	doneFile := filepath.Join(tempDir, "para.done")

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
	t.Logf("%s text para AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("%s text para ship gate FAIL:\n%s", ver, body)
	}

	// AE2020 has no paragraph-indent DOM — prove acceptance + preservation by
	// re-parsing AE's resave and confirming the paragraph values survived.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	rt := re.Compositions[0].LayerByName("T")
	if rt == nil || rt.TextSource == nil || len(rt.TextSource.Paragraphs) == 0 {
		t.Fatalf("%s resaved: text paragraphs missing", ver)
	}
	pp := rt.TextSource.Paragraphs[0]
	chkPara(t, ver, "FirstLineIndent", pp.FirstLineIndent, 20)
	chkPara(t, ver, "StartIndent", pp.StartIndent, 15)
	chkPara(t, ver, "EndIndent", pp.EndIndent, 10)
	chkPara(t, ver, "SpaceBefore", pp.SpaceBefore, 25)
	chkPara(t, ver, "SpaceAfter", pp.SpaceAfter, 30)
}

func chkPara(t *testing.T, ver, label string, got, want float64) {
	t.Helper()
	if got-want > 0.1 || want-got > 0.1 {
		t.Errorf("%s resaved: paragraph %s = %g, want %g", ver, label, got, want)
	}
}

func TestTextPara_AEShipGate_AE2020(t *testing.T) {
	runTextParaGate(t, ae2020(), "AE2020", aep.TargetAE2020)
}
func TestTextPara_AEShipGate_AE2025(t *testing.T) {
	runTextParaGate(t, ae2025(), "AE2025", aep.TargetAE2025)
}
