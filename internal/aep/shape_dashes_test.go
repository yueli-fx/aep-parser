// internal/aep/shape_dashes_test.go
//
// Stroke Dashes setter: model defaults + WriteAEP->FromReader roundtrip.
// Dash/Gap are OneD float64-BE at cdat[0:8] inside the Dashes LIST(tdgp),
// emitted only when enabled (the serializer swaps to a dashed stroke-body
// template). V2.2 models one Dash + Gap pair; Offset and Dash/Gap 2/3 are
// deferred. RE: incident-reports/stroke-line-cap-join-miter-re.md (Dashes).
package aep_test

import (
	"bytes"
	"math"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestStrokeDashes_Defaults(t *testing.T) {
	s := aep.NewStrokeNode()
	d := s.Dashes()
	if d == nil {
		t.Fatal("Dashes() must be non-nil on a fresh StrokeNode")
	}
	if d.Enabled() {
		t.Error("dashes must be disabled by default (solid stroke)")
	}
	if d.Dash() != 10 || d.Gap() != 10 {
		t.Errorf("dash/gap defaults = %g/%g, want 10/10", d.Dash(), d.Gap())
	}
	// SetDash / SetGap auto-enable.
	_ = d.SetDash(5)
	if !d.Enabled() {
		t.Error("SetDash must enable dashing")
	}
	if err := d.SetGap(-1); err == nil {
		t.Error("SetGap(-1) must reject negative")
	}
}

func TestV2_2_StrokeDashes_Roundtrip(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := aep.NewShapeLayer(comp, "Dashed")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	rect, _ := l.RootGroup().AddRect()
	_ = rect.SetSize([2]float64{200, 100})
	stroke, _ := l.RootGroup().AddStroke()
	_ = stroke.SetWidth(12)
	// Values distinct from the embedded dashed template (Dash 20 / Gap 10) so the
	// roundtrip proves the cdat overwrite, not template passthrough.
	d := stroke.Dashes()
	_ = d.SetDash(18)
	_ = d.SetGap(7)

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	ss := recoverStrokeNode(t, re.Compositions[0], "Dashed")
	rd := ss.Dashes()
	if !rd.Enabled() {
		t.Fatal("dashes must round-trip as enabled")
	}
	if math.Abs(rd.Dash()-18) > 1e-9 {
		t.Errorf("dash = %g, want 18", rd.Dash())
	}
	if math.Abs(rd.Gap()-7) > 1e-9 {
		t.Errorf("gap = %g, want 7", rd.Gap())
	}
}

// TestV2_2_StrokeDashes_DisabledStaysSolid proves a default stroke (dashes off)
// round-trips with dashing disabled — i.e. the solid template carries no
// Dash/Gap slots and hydrate does not spuriously flag enabled.
func TestV2_2_StrokeDashes_DisabledStaysSolid(t *testing.T) {
	p := aep.NewProject()
	comp, _ := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	l, _ := aep.NewShapeLayer(comp, "Solid")
	rect, _ := l.RootGroup().AddRect()
	_ = rect.SetSize([2]float64{200, 100})
	stroke, _ := l.RootGroup().AddStroke()
	_ = stroke.SetWidth(12) // dashes left disabled

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	ss := recoverStrokeNode(t, re.Compositions[0], "Solid")
	if ss.Dashes().Enabled() {
		t.Error("a stroke with dashes off must round-trip as disabled")
	}
}

func recoverStrokeNode(t *testing.T, comp *aep.Composition, layerName string) *aep.StrokeNode {
	t.Helper()
	sl := aep.WrapShapeLayer(findLayerByName(comp, layerName))
	for _, ch := range sl.RootGroup().Children {
		if n, ok := ch.(*aep.StrokeNode); ok {
			return n
		}
	}
	t.Fatalf("stroke node not recovered after roundtrip (layer %q)", layerName)
	return nil
}
