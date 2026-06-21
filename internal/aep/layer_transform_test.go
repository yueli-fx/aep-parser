package aep_test

import (
	"bytes"
	"math"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// TestSetLayerTransform_TemplatedLayer_Animated verifies the from-scratch
// animated-transform path for a TEMPLATED layer (text): a fresh NewTextLayer
// elides its default Position/Opacity/Anchor channels, so SetLayerTransform must
// materialize + animate them. Round-trips and asserts the keyframes survive.
func TestSetLayerTransform_TemplatedLayer_Animated(t *testing.T) {
	p := aep.NewProject(aep.TargetAE2020)
	c, err := aep.NewComposition(p, "T", 1920, 1080, 30, 6)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aep.NewTextLayer(c, "GLITCH"); err != nil {
		t.Fatal(err)
	}

	// Round-trip so the templated layer's chunks are parsed (back-refs present).
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	layer := rp.Compositions[0].Layers[0]

	// Sanity: the fresh template elides Position/Opacity (the gap this closes).
	if layer.Position() != nil {
		t.Logf("note: fresh text layer already exposes Position (template changed?)")
	}

	tr := aep.NewLayerTransform()
	if err := tr.AnchorPoint().SetStaticValue([2]float64{9, -436}); err != nil {
		t.Fatal(err)
	}
	// Position: 3 linear keyframes (pixels).
	for _, kf := range []struct {
		t    float64
		x, y float64
	}{{0, 960, 488}, {0.3, 960, 616}, {0.6, 960, 465}} {
		if err := tr.Position().AddKeyframeLinear(kf.t, [2]float64{kf.x, kf.y}); err != nil {
			t.Fatal(err)
		}
	}
	// Opacity: 3 linear keyframes (percent 0-100).
	for _, kf := range []struct {
		t, v float64
	}{{0, 0}, {0.3, 100}, {0.6, 3}} {
		if err := tr.Opacity().AddKeyframeLinear(kf.t, kf.v); err != nil {
			t.Fatal(err)
		}
	}
	if err := aep.SetLayerTransform(layer, tr); err != nil {
		t.Fatalf("SetLayerTransform: %v", err)
	}

	// Write + reopen; assert the materialized animated channels survived.
	var buf bytes.Buffer
	if err := rp.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	rp2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	l2 := rp2.Compositions[0].Layers[0]

	pos := l2.Position()
	if pos == nil {
		t.Fatal("Position() nil after SetLayerTransform — channel not materialized")
	}
	if len(pos.Keyframes) != 3 {
		t.Fatalf("Position keyframes = %d, want 3", len(pos.Keyframes))
	}
	if got := pos.Keyframes[1]; math.Abs(got.Time-0.3) > 1e-3 {
		t.Errorf("Position kf[1].Time = %.4f, want 0.3", got.Time)
	}

	op := l2.Opacity()
	if op == nil {
		t.Fatal("Opacity() nil after SetLayerTransform — channel not materialized")
	}
	if len(op.Keyframes) != 3 {
		t.Fatalf("Opacity keyframes = %d, want 3", len(op.Keyframes))
	}

	anchor := l2.AnchorPoint()
	if anchor == nil {
		t.Fatal("AnchorPoint() nil after SetLayerTransform")
	}
	t.Logf("OK: pos kf=%d op kf=%d anchor static=%v pos[1].val=%v op[1].val=%v",
		len(pos.Keyframes), len(op.Keyframes), anchor.StaticValue, pos.Keyframes[1].Value, op.Keyframes[1].Value)
}
