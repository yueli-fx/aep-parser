// internal/aep/hydrate_shape_test.go
//
// hydrateShapeNodes minimum test.
// Round-trips a single RectNode through WriteAEP → FromReader and asserts
// the hydrated runtime tree carries Size=[200,100].
package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestHydrate_RectNode_FromChunks(t *testing.T) {
	p := aep.NewProject()
	c, err := p.NewComposition("M", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	s, err := c.NewShapeLayer("S1")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	r, err := s.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect: %v", err)
	}
	if err := r.SetSize([2]float64{200, 100}); err != nil {
		t.Fatalf("SetSize: %v", err)
	}

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}

	if len(re.Compositions) == 0 {
		t.Fatal("re has no compositions")
	}
	reComp := re.Compositions[0]
	if len(reComp.Layers) != 1 {
		t.Fatalf("layers = %d, want 1", len(reComp.Layers))
	}
	reShape := aep.WrapShapeLayer(reComp.Layers[0])
	if reShape.RootGroup() == nil {
		t.Fatal("nil RootGroup post-hydrate")
	}
	if got := len(reShape.RootGroup().Children); got != 1 {
		t.Fatalf("RootGroup children = %d, want 1 (rect)", got)
	}
	reRect, ok := reShape.RootGroup().Children[0].(*aep.RectNode)
	if !ok {
		t.Fatalf("RootGroup.Children[0] type = %T, want *RectNode", reShape.RootGroup().Children[0])
	}
	v, isStatic := reRect.Size().StaticValue()
	if !isStatic {
		t.Fatal("Size not static post-hydrate")
	}
	if v != [2]float64{200, 100} {
		t.Fatalf("hydrated Size = %v, want [200,100]", v)
	}
}
