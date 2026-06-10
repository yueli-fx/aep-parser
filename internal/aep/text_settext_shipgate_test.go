// internal/aep/text_settext_shipgate_test.go
//
// AE ship gate for length-variable SetText. Builds a 100% Go-built project
// with two text layers whose strings were replaced with DIFFERENT lengths
// than the template's 1-char "A" — "Hello AEP parser" (ASCII, 1→17 units)
// and "你好世界" (CJK, 1→5 units) — WriteAEP, and has AE open it: proving AE
// accepts the spliced string + paragraph/run character counters with the
// stale layout cache, reads back both strings exactly, and keeps them across
// its own resave.
//
// Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"

	aep "github.com/example/aep-parser/internal/aep"
)

// jsStringEscape renders s as a double-quoted ExtendScript string literal with
// all non-ASCII escaped to \uXXXX (surrogate pairs for astral runes), so the
// args file stays pure ASCII and no File-encoding negotiation is needed.
func jsStringEscape(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, u := range utf16.Encode([]rune(s)) {
		if u < 0x80 && u != '"' && u != '\\' {
			b.WriteByte(byte(u))
			continue
		}
		fmt.Fprintf(&b, "\\u%04x", u)
	}
	b.WriteByte('"')
	return b.String()
}

func runSetTextGate(t *testing.T, target aep.AETarget, aeExe, ver string) {
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	const argsPath = `e:/projects/tools/aep-parser/test_data/set_text_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_set_text.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	expect := map[string]string{
		"T1": "Hello AEP parser",
		"T2": "你好世界",
	}

	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	for _, name := range []string{"T1", "T2"} {
		l, err := aep.NewTextLayer(comp, name)
		if err != nil {
			t.Fatalf("NewTextLayer %s: %v", name, err)
		}
		if err := l.SetText(expect[name]); err != nil {
			t.Fatalf("SetText %s: %v", name, err)
		}
	}

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
	for _, name := range []string{"T1", "T2"} {
		expectParts = append(expectParts, fmt.Sprintf("%q:%s", name, jsStringEscape(expect[name])))
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

	// Preservation proof: AE's resave kept the replaced strings.
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
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

func TestSetTextVariable_AEShipGate_AE2020(t *testing.T) {
	runSetTextGate(t, aep.TargetAE2020, ae2020(), "AE2020")
}

func TestSetTextVariable_AEShipGate_AE2025(t *testing.T) {
	runSetTextGate(t, aep.TargetAE2025, ae2025(), "AE2025")
}
