// internal/aep/text_settext_shipgate_test.go
//
// AE ship gate for length-variable SetText. Two gates:
//
//   runSetTextGate — 100% Go-built project whose text layers' strings were
//   replaced with different shapes than the template's 1-char "A": ASCII
//   1→17 units, CJK 1→5, a three-paragraph block (paragraph-array rebuild),
//   and an empty string (single "\r" paragraph).
//
//   runSetTextMultiRunGate — loads re_text_multirun.aep (a two-style-run doc
//   built by AE) and replaces its text, exercising the run-array collapse
//   (2 runs → 1) that a from-scratch single-run template can't reach.
//
// Both prove AE accepts the spliced string + rebuilt entry arrays with the
// stale layout cache, reads back every string exactly, and keeps them across
// its own resave. Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// jsStringEscape renders s as a double-quoted ExtendScript string literal with
// every non-printable-ASCII unit escaped to \uXXXX (surrogate pairs for astral
// runes; carriage-return / newline paragraph breaks included), so the args
// file stays pure printable ASCII and eval() sees no raw line terminators.
func jsStringEscape(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, u := range utf16.Encode([]rune(s)) {
		if u >= 0x20 && u < 0x80 && u != '"' && u != '\\' {
			b.WriteByte(byte(u))
			continue
		}
		fmt.Fprintf(&b, "\\u%04x", u)
	}
	b.WriteByte('"')
	return b.String()
}

// textVerifyInAE writes p to a temp .aep, has AE open it via verify_set_text.jsx
// asserting each expect[name] reads back, then returns the AE-resaved project
// reopened for caller-specific assertions. expect values use the Go-decoded
// form (\n breaks, no trailing terminator); the JSX side compares against the
// \n→\r mapping AE's sourceText.value.text uses.
func textVerifyInAE(t *testing.T, aeExe, ver string, p *aep.Project, expect map[string]string) *aep.Project {
	t.Helper()
	const argsPath = `e:/projects/tools/aep-parser/test_data/set_text_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_set_text.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "settext_in.aep")
	resavedAEP := filepath.Join(tempDir, "settext_resaved.aep")
	doneFile := filepath.Join(tempDir, "settext.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := p.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	var expectParts []string
	for name, text := range expect {
		jsExpect := strings.ReplaceAll(text, "\n", "\r")
		expectParts = append(expectParts, fmt.Sprintf("%q:%s", name, jsStringEscape(jsExpect)))
	}
	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q,"expect":{%s}}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP), strings.Join(expectParts, ","))
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
	t.Logf("%s AE readback:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Errorf("%s set-text ship gate FAIL:\n%s", ver, body)
	}

	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	return re
}

func runSetTextGate(t *testing.T, target aep.AETarget, aeExe, ver string) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	names := []string{"T1", "T2", "T3", "T4"}
	expect := map[string]string{
		"T1": "Hello AEP parser",
		"T2": "你好世界",
		"T3": "Line one\nLine two\nLine three",
		"T4": "",
	}

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	for _, name := range names {
		l, err := aep.NewTextLayer(comp, name)
		if err != nil {
			t.Fatalf("NewTextLayer %s: %v", name, err)
		}
		if err := l.SetText(expect[name]); err != nil {
			t.Fatalf("SetText %s: %v", name, err)
		}
	}

	re := textVerifyInAE(t, aeExe, ver, p, expect)
	for name, text := range expect {
		rl := textLayerByName(re, name)
		if rl == nil {
			t.Errorf("%s resaved: text layer %q missing", ver, name)
			continue
		}
		if rl.TextSource == nil || rl.TextSource.Text != text {
			t.Errorf("%s resaved: %q TextSource = %+v, want text %q", ver, name, rl.TextSource, text)
		}
	}
}

// runSetTextMultiRunGate loads the AE-built two-run fixture, replaces its text
// (collapsing the run array 2→1), and proves AE accepts the result, reads the
// new text, and resaves a single-run document.
//
// AE 2024 + AE 2025: the fixture needs characterRange (AE 24+) to hold two
// runs, so it is AE-24-stamped — openable by AE 2024 (native) and AE 2025
// (backward-compat). AE 2020 refuses it outright (forward-incompat, independent
// of our bytes), and AE 2020 has no scripting path to a multi-run doc anyway,
// so a 2020-openable variant can't be automated; buildEntryArray is separately
// double-version-proven via the paragraph case in runSetTextGate (T3 passes on
// both AE 2020 and AE 2025).
func runSetTextMultiRunGate(t *testing.T, aeExe, ver string) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const fixture = `e:/projects/tools/aep-parser/test_data/re_text_multirun.aep`
	proj, err := aep.Open(fixture)
	if err != nil {
		t.Skipf("re_text_multirun.aep not present; run test_data/re_text_multirun.jsx in AE 2025")
	}
	src := textLayerByName(proj, "multirun_src")
	if src == nil || src.TextSource == nil {
		t.Fatal("multirun_src layer not found")
	}
	if len(src.TextSource.Runs) != 2 {
		t.Fatalf("fixture multirun_src should carry 2 style runs, got %d", len(src.TextSource.Runs))
	}
	const newText = "collapsed to one run"
	if err := src.SetText(newText); err != nil {
		t.Fatalf("SetText on multi-run doc: %v", err)
	}

	re := textVerifyInAE(t, aeExe, ver, proj, map[string]string{"multirun_src": newText})
	rl := textLayerByName(re, "multirun_src")
	if rl == nil || rl.TextSource == nil {
		t.Fatalf("%s resaved: multirun_src missing", ver)
	}
	if rl.TextSource.Text != newText {
		t.Errorf("%s resaved: text = %q, want %q", ver, rl.TextSource.Text, newText)
	}
	if len(rl.TextSource.Runs) != 1 {
		t.Errorf("%s resaved: run count = %d, want 1 (collapsed)", ver, len(rl.TextSource.Runs))
	}
}

// runSetTextKerningGate loads the AE-built manual-kerning fixture and replaces
// its text, exercising the kerning-table drop. Like the multi-run gate it is
// AE 2024 + AE 2025: td.kerning needs AE 24+, so the fixture is AE-24-stamped
// (AE 2020 forward-incompat, and 2020 has no manual-kerning scripting path).
func runSetTextKerningGate(t *testing.T, aeExe, ver string) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const fixture = `e:/projects/tools/aep-parser/test_data/re_text_kern_resize.aep`
	proj, err := aep.Open(fixture)
	if err != nil {
		t.Skipf("re_text_kern_resize.aep not present; run test_data/re_text_kern_resize.jsx in AE 2024")
	}
	src := textLayerByName(proj, "kern_src")
	if src == nil || src.TextSource == nil {
		t.Fatal("kern_src layer not found")
	}
	if len(src.TextSource.ManualKerning) == 0 {
		t.Fatalf("fixture kern_src should carry a manual-kerning table, got none")
	}
	const newText = "kerning dropped here"
	if err := src.SetText(newText); err != nil {
		t.Fatalf("SetText on kerned doc: %v", err)
	}

	re := textVerifyInAE(t, aeExe, ver, proj, map[string]string{"kern_src": newText})
	rl := textLayerByName(re, "kern_src")
	if rl == nil || rl.TextSource == nil {
		t.Fatalf("%s resaved: kern_src missing", ver)
	}
	if rl.TextSource.Text != newText {
		t.Errorf("%s resaved: text = %q, want %q", ver, rl.TextSource.Text, newText)
	}
	if len(rl.TextSource.ManualKerning) != 0 {
		t.Errorf("%s resaved: ManualKerning = %v, want empty (dropped)", ver, rl.TextSource.ManualKerning)
	}
}

func TestSetTextVariable_AEShipGate_AE2020(t *testing.T) {
	runSetTextGate(t, aep.TargetAE2020, ae2020(), "AE2020")
}

func TestSetTextKerning_AEShipGate_AE2024(t *testing.T) {
	runSetTextKerningGate(t, ae2024(), "AE2024")
}

func TestSetTextKerning_AEShipGate_AE2025(t *testing.T) {
	runSetTextKerningGate(t, ae2025(), "AE2025")
}

func TestSetTextVariable_AEShipGate_AE2025(t *testing.T) {
	runSetTextGate(t, aep.TargetAE2025, ae2025(), "AE2025")
}

func TestSetTextMultiRun_AEShipGate_AE2024(t *testing.T) {
	runSetTextMultiRunGate(t, ae2024(), "AE2024")
}

func TestSetTextMultiRun_AEShipGate_AE2025(t *testing.T) {
	runSetTextMultiRunGate(t, ae2025(), "AE2025")
}
