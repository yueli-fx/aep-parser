// internal/aep/text_settext_variable_test.go
//
// Length-variable SetText (single-paragraph / single-run / kerning-free):
// string splice + paragraph & run character-counter sync + btdk size-header
// fix, layout cache left for AE to recompute. Uses NewTextLayer's embedded
// AE-native template as the always-available document, plus re_text.aep
// fixtures for parsed-layer and refuse-path coverage.
package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/codec"
)

// btdkCounter reads the UTF-16 character counter at the given btdk path from
// a layer's raw text source.
func btdkCounter(t *testing.T, raw []byte, path string) float64 {
	t.Helper()
	body, _, err := codec.ExtractBtdkBody(raw)
	if err != nil {
		t.Fatalf("ExtractBtdkBody: %v", err)
	}
	root := codec.ParsePSDict(body)
	if root == nil {
		t.Fatal("btdk dict empty")
	}
	v := codec.PsPath(root, path)
	if v == nil || v.Kind != codec.PsNum {
		t.Fatalf("no number at %s", path)
	}
	return v.Num
}

func freshTextLayer(t *testing.T, name string) (*aep.Project, *aep.Layer) {
	t.Helper()
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := aep.NewTextLayer(comp, name)
	if err != nil {
		t.Fatalf("NewTextLayer: %v", err)
	}
	return p, l
}

func TestSetTextVariable_GrowAndCounters(t *testing.T) {
	p, l := freshTextLayer(t, "T")
	const text = "Hello AEP 你好"
	if err := l.SetText(text); err != nil {
		t.Fatalf("SetText variable: %v", err)
	}
	if l.TextSource.Text != text {
		t.Errorf("TextSource.Text = %q, want %q", l.TextSource.Text, text)
	}
	wantCount := float64(len([]rune(text)) + 1) // BMP only: runes == UTF-16 units; +1 trailing \r
	for _, path := range []string{"/1/1/0/0/5/0/0/1", "/1/1/0/0/6/0/0/1"} {
		if got := btdkCounter(t, l.TextSourceRaw, path); got != wantCount {
			t.Errorf("counter %s = %v, want %v", path, got, wantCount)
		}
	}

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	rp, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	if len(rp.Warnings) != 0 {
		t.Errorf("re-parse produced warnings: %v", rp.Warnings)
	}
	rl := textLayerByName(rp, "T")
	if rl == nil {
		t.Fatal("re-parsed: layer T missing")
	}
	if rl.TextSource == nil || rl.TextSource.Text != text {
		t.Errorf("re-parsed TextSource = %+v, want text %q", rl.TextSource, text)
	}
}

func TestSetTextVariable_ShrinkAndSurrogate(t *testing.T) {
	p, l := freshTextLayer(t, "T")
	if err := l.SetText("a longer interim string"); err != nil {
		t.Fatalf("grow: %v", err)
	}
	if err := l.SetText("X"); err != nil {
		t.Fatalf("shrink: %v", err)
	}
	if got := btdkCounter(t, l.TextSourceRaw, "/1/1/0/0/5/0/0/1"); got != 2 {
		t.Errorf("paragraph counter after shrink = %v, want 2", got)
	}

	// Astral rune: counters are UTF-16 code units, so 𝄞 counts as 2.
	if err := l.SetText("𝄞"); err != nil {
		t.Fatalf("surrogate: %v", err)
	}
	if got := btdkCounter(t, l.TextSourceRaw, "/1/1/0/0/6/0/0/1"); got != 3 {
		t.Errorf("run counter for surrogate text = %v, want 3 (2 units + \\r)", got)
	}

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	rp, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	rl := textLayerByName(rp, "T")
	if rl == nil || rl.TextSource == nil || rl.TextSource.Text != "𝄞" {
		t.Fatalf("re-parsed text = %+v, want 𝄞", rl)
	}
}

func TestSetTextVariable_Refusals(t *testing.T) {
	_, l := freshTextLayer(t, "T")
	if err := l.SetText(""); err == nil {
		t.Error("empty text: expected refuse, got nil")
	}
	if err := l.SetText("a\nb"); err == nil {
		t.Error("line break: expected refuse, got nil")
	}
	if l.TextSource.Text != "A" {
		t.Errorf("after refusals: Text = %q, want untouched %q", l.TextSource.Text, "A")
	}
}

func TestSetTextVariable_ParsedFixture(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_text.aep")
	if err != nil {
		t.Skipf("re_text.aep not present; run test_data/re_text.jsx in AE")
	}
	// re_text.aep carries historic duplicate RE_TEXT comps (the generator JSX
	// predates the fresh_project guard — see incidents/jsx-state-leak.md), so
	// pin the FIRST match in parse order for both mutation and verification.
	hello := textLayerByName(proj, "text_hello")
	twoLines := textLayerByName(proj, "text_two_lines")
	if hello == nil || twoLines == nil {
		t.Fatal("fixture layers text_hello / text_two_lines not found")
	}

	// Parsed single-paragraph layer: arbitrary-length replacement.
	const text = "Replaced with a longer string"
	if err := hello.SetText(text); err != nil {
		t.Fatalf("SetText on parsed layer: %v", err)
	}
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	rp, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	rl := textLayerByName(rp, "text_hello")
	if rl == nil || rl.TextSource == nil || rl.TextSource.Text != text {
		t.Errorf("re-parsed text_hello = %+v, want %q", rl.TextSource, text)
	}

	// Multi-paragraph document: length-variable refused, equal-length in-place
	// still works ("A\rB" → "X\rY").
	if err := twoLines.SetText("XYZ"); err == nil {
		t.Error("multi-paragraph variable SetText: expected refuse, got nil")
	}
	if err := twoLines.SetText("X\nY"); err != nil {
		t.Errorf("multi-paragraph equal-length SetText: %v", err)
	} else if twoLines.TextSource.Text != "X\nY" {
		t.Errorf("two_lines text = %q, want %q", twoLines.TextSource.Text, "X\nY")
	}
}
