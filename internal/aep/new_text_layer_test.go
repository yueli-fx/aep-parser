// internal/aep/new_text_layer_test.go
//
// Go-only coverage for NewTextLayer: template clone round-trips through
// WriteAEP + re-parse with the right type/name/text, and the create-time
// text-source wiring supports the length-preserving SetText without a Reopen.
package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func textLayerByName(p *aep.Project, name string) *aep.Layer {
	for _, c := range p.Compositions {
		for _, l := range c.Layers {
			if l.Name == name {
				return l
			}
		}
	}
	return nil
}

func TestNewTextLayer_RoundTrip(t *testing.T) {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := aep.NewTextLayer(comp, "T1")
	if err != nil {
		t.Fatalf("NewTextLayer: %v", err)
	}
	if l.Type != aep.LayerTypeText {
		t.Errorf("fresh layer Type = %v, want %v", l.Type, aep.LayerTypeText)
	}
	if l.TextSource == nil || l.TextSource.Text != "A" {
		t.Errorf("fresh layer TextSource = %+v, want template text %q", l.TextSource, "A")
	}

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	rp, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	rl := textLayerByName(rp, "T1")
	if rl == nil {
		t.Fatal("re-parsed project: layer T1 missing")
	}
	if rl.Type != aep.LayerTypeText {
		t.Errorf("re-parsed Type = %v, want %v", rl.Type, aep.LayerTypeText)
	}
	if rl.TextSource == nil || rl.TextSource.Text != "A" {
		t.Errorf("re-parsed TextSource = %+v, want text %q", rl.TextSource, "A")
	}
}

func TestNewTextLayer_SetTextBeforeWrite(t *testing.T) {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := aep.NewTextLayer(comp, "T2")
	if err != nil {
		t.Fatalf("NewTextLayer: %v", err)
	}
	if err := l.SetText("B"); err != nil {
		t.Fatalf("SetText on fresh layer: %v", err)
	}
	if l.TextSource.Text != "B" {
		t.Errorf("after SetText: Text = %q, want %q", l.TextSource.Text, "B")
	}

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	rp, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	rl := textLayerByName(rp, "T2")
	if rl == nil {
		t.Fatal("re-parsed project: layer T2 missing")
	}
	if rl.TextSource == nil || rl.TextSource.Text != "B" {
		t.Errorf("re-parsed TextSource = %+v, want text %q", rl.TextSource, "B")
	}
}

func TestNewTextLayer_TwoLayersDistinct(t *testing.T) {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	a, err := aep.NewTextLayer(comp, "TA")
	if err != nil {
		t.Fatalf("NewTextLayer TA: %v", err)
	}
	b, err := aep.NewTextLayer(comp, "TB")
	if err != nil {
		t.Fatalf("NewTextLayer TB: %v", err)
	}
	if a.ID == b.ID {
		t.Errorf("layer IDs collide: %d", a.ID)
	}
	// Templates must not share backing arrays: mutating one text leaves the
	// other untouched.
	if err := b.SetText("Z"); err != nil {
		t.Fatalf("SetText TB: %v", err)
	}
	if a.TextSource.Text != "A" {
		t.Errorf("TA text changed to %q after editing TB (shared template bytes)", a.TextSource.Text)
	}
}
