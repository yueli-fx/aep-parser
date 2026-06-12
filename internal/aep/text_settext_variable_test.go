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
	"reflect"
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

// btdkParaCounts returns the per-paragraph UTF-16 counts from the paragraph
// array at /1/1/0/0/5/0 (each entry's /1).
func btdkParaCounts(t *testing.T, raw []byte) []float64 {
	t.Helper()
	body, _, err := codec.ExtractBtdkBody(raw)
	if err != nil {
		t.Fatalf("ExtractBtdkBody: %v", err)
	}
	root := codec.ParsePSDict(body)
	if root == nil {
		t.Fatal("btdk dict empty")
	}
	arr := codec.PsPath(root, "/1/1/0/0/5/0")
	if arr == nil || arr.Kind != codec.PsArr {
		t.Fatal("no paragraph array at /1/1/0/0/5/0")
	}
	out := make([]float64, len(arr.Arr))
	for i, entry := range arr.Arr {
		out[i] = codec.PsPath(entry, "/1").AsNum()
	}
	return out
}

func TestSetTextVariable_MultiParagraphAndEmpty(t *testing.T) {
	p, l := freshTextLayer(t, "T")

	// Grow "A" into three paragraphs: paragraph array gains one entry per line,
	// each counting its text + trailing \r; the single run carries the total.
	if err := l.SetText("L1\nL2\nL3"); err != nil {
		t.Fatalf("multi-paragraph SetText: %v", err)
	}
	if l.TextSource.Text != "L1\nL2\nL3" {
		t.Errorf("Text = %q, want %q", l.TextSource.Text, "L1\nL2\nL3")
	}
	if got := btdkParaCounts(t, l.TextSourceRaw); !reflect.DeepEqual(got, []float64{3, 3, 3}) {
		t.Errorf("paragraph counts = %v, want [3 3 3]", got)
	}
	if got := btdkCounter(t, l.TextSourceRaw, "/1/1/0/0/6/0/0/1"); got != 9 {
		t.Errorf("run count = %v, want 9", got)
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
	if rl := textLayerByName(rp, "T"); rl == nil || rl.TextSource == nil || rl.TextSource.Text != "L1\nL2\nL3" {
		t.Errorf("re-parsed text = %+v, want %q", rl, "L1\nL2\nL3")
	}

	// Shrink to empty: a single empty paragraph "\r" (count 1).
	if err := l.SetText(""); err != nil {
		t.Fatalf("empty SetText: %v", err)
	}
	if l.TextSource.Text != "" {
		t.Errorf("Text = %q, want empty", l.TextSource.Text)
	}
	if got := btdkParaCounts(t, l.TextSourceRaw); !reflect.DeepEqual(got, []float64{1}) {
		t.Errorf("empty paragraph counts = %v, want [1]", got)
	}
	if got := btdkCounter(t, l.TextSourceRaw, "/1/1/0/0/6/0/0/1"); got != 1 {
		t.Errorf("empty run count = %v, want 1", got)
	}
}

func TestSetTextVariable_MultiRunCollapse(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_text_multirun.aep")
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

	// Length-changing replacement on a multi-run doc: collapse to one run.
	if err := src.SetText("collapsed run"); err != nil {
		t.Fatalf("SetText on multi-run doc: %v", err)
	}
	if got := btdkCounter(t, src.TextSourceRaw, "/1/1/0/0/6/0/0/1"); got != 14 {
		t.Errorf("collapsed run count = %v, want 14", got) // "collapsed run" = 13 + \r
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	rp, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	if len(rp.Warnings) != 0 {
		t.Errorf("re-parse produced warnings: %v", rp.Warnings)
	}
	rl := textLayerByName(rp, "multirun_src")
	if rl == nil || rl.TextSource == nil {
		t.Fatal("re-parsed multirun_src missing")
	}
	if rl.TextSource.Text != "collapsed run" {
		t.Errorf("text = %q, want %q", rl.TextSource.Text, "collapsed run")
	}
	if len(rl.TextSource.Runs) != 1 {
		t.Errorf("run count after collapse = %d, want 1", len(rl.TextSource.Runs))
	}
}

func TestSetTextVariable_DropsManualKerning(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_text_kern_resize.aep")
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

	// A length-changing replacement drops the manual-kerning table (AE behavior).
	if err := src.SetText("Hello world"); err != nil {
		t.Fatalf("SetText on kerned doc: %v", err)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	rp, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	if len(rp.Warnings) != 0 {
		t.Errorf("re-parse produced warnings: %v", rp.Warnings)
	}
	rl := textLayerByName(rp, "kern_src")
	if rl == nil || rl.TextSource == nil {
		t.Fatal("re-parsed kern_src missing")
	}
	if rl.TextSource.Text != "Hello world" {
		t.Errorf("text = %q, want %q", rl.TextSource.Text, "Hello world")
	}
	if len(rl.TextSource.ManualKerning) != 0 {
		t.Errorf("ManualKerning after text change = %v, want empty (dropped)", rl.TextSource.ManualKerning)
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

	// Multi-paragraph parsed document collapsing to a single paragraph: the
	// paragraph array is rebuilt down to one entry, run count set to the total.
	if err := twoLines.SetText("XYZ"); err != nil {
		t.Fatalf("multi-paragraph → single SetText: %v", err)
	}
	if got := btdkParaCounts(t, twoLines.TextSourceRaw); !reflect.DeepEqual(got, []float64{4}) {
		t.Errorf("paragraph counts after collapse = %v, want [4]", got)
	}
	if twoLines.TextSource.Text != "XYZ" {
		t.Errorf("two_lines text = %q, want %q", twoLines.TextSource.Text, "XYZ")
	}
}
