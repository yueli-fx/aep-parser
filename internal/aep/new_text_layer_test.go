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

// TestNewLayer_CrossCompIDsDistinct guards the cross-composition layer-ID
// regression: AE allocates items AND layers from one global head counter, so a
// layer created in comp B must not reuse a layer ID already handed out in comp A.
// The per-comp maxLayerIDInItemList+1 formula gave every comp's first user layer
// the same ID (13, just past the service layers) — the file rendered but AE's UI
// delete/edit (ID-keyed) then hit "unexpected match name searched for in group".
func TestNewLayer_CrossCompIDsDistinct(t *testing.T) {
	p := aep.NewProject(aep.TargetAE2020)
	cA, err := aep.NewComposition(p, "A", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition A: %v", err)
	}
	cB, err := aep.NewComposition(p, "B", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition B: %v", err)
	}
	sa, err := aep.NewShapeLayer(cA, "ShapeA")
	if err != nil {
		t.Fatalf("NewShapeLayer A: %v", err)
	}
	tb, err := aep.NewTextLayer(cB, "TextB")
	if err != nil {
		t.Fatalf("NewTextLayer B: %v", err)
	}
	if sa.ID == tb.ID {
		t.Errorf("cross-comp layer IDs collide: ShapeA and TextB both %d", sa.ID)
	}
	// Every entity ID in the project must be unique (comps + layers).
	seen := map[uint32]string{}
	for _, c := range p.Compositions {
		if prev, ok := seen[c.ID]; ok {
			t.Errorf("ID %d reused: comp %q and %s", c.ID, c.Name, prev)
		}
		seen[c.ID] = "comp " + c.Name
		for _, l := range c.Layers {
			if prev, ok := seen[l.ID]; ok {
				t.Errorf("ID %d reused: layer %q/%q and %s", l.ID, c.Name, l.Name, prev)
			}
			seen[l.ID] = "layer " + l.Name
		}
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
