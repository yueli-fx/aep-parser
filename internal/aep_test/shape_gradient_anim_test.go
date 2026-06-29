package aep_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/rifx"
)

// gradAnim builds a project with one shape layer whose gradient fill has
// animated color stops: kf0@0s = R/B/G, kf1@1s = G/R/B.
func buildGradAnim(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "GRADANIM", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := aep.NewShapeLayer(comp, "GRAD")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	rect, err := l.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect: %v", err)
	}
	_ = rect.SetSize([2]float64{1000, 1000})
	_ = l.Position().SetStaticValue([2]float64{960, 540})
	gf, err := l.RootGroup().AddGradientFill()
	if err != nil {
		t.Fatalf("AddGradientFill: %v", err)
	}
	_ = gf.SetStartPoint([2]float64{-500, 0})
	_ = gf.SetEndPoint([2]float64{500, 0})

	rbg := &aep.Gradient{
		Version: "4",
		ColorStops: []aep.GradientColorStop{
			{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0, 0}},
			{Offset: 0.5, Midpoint: 0.5, Color: [3]float64{0, 0, 1}},
			{Offset: 1, Midpoint: 0.5, Color: [3]float64{0, 1, 0}},
		},
		AlphaStops: []aep.GradientAlphaStop{
			{Offset: 0, Midpoint: 0.5, Alpha: 1},
			{Offset: 1, Midpoint: 0.5, Alpha: 1},
		},
	}
	grb := &aep.Gradient{
		Version: "4",
		ColorStops: []aep.GradientColorStop{
			{Offset: 0, Midpoint: 0.5, Color: [3]float64{0, 1, 0}},
			{Offset: 0.5, Midpoint: 0.5, Color: [3]float64{1, 0, 0}},
			{Offset: 1, Midpoint: 0.5, Color: [3]float64{0, 0, 1}},
		},
		AlphaStops: []aep.GradientAlphaStop{
			{Offset: 0, Midpoint: 0.5, Alpha: 1},
			{Offset: 1, Midpoint: 0.5, Alpha: 1},
		},
	}
	if err := gf.AddGradientKeyframe(0, rbg); err != nil {
		t.Fatalf("AddGradientKeyframe 0: %v", err)
	}
	if err := gf.AddGradientKeyframe(1, grb); err != nil {
		t.Fatalf("AddGradientKeyframe 1: %v", err)
	}
	return p
}

// buildGradStrokeAnim builds a shape layer with a rect outlined by a gradient
// STROKE whose color stops are animated: kf0@0s = R/B/G, kf1@1s = G/R/B.
func buildGradStrokeAnim(t *testing.T, target aep.AETarget) *aep.Project {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "GRADSTROKEANIM", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := aep.NewShapeLayer(comp, "GRADSTROKE")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	rect, err := l.RootGroup().AddRect()
	if err != nil {
		t.Fatalf("AddRect: %v", err)
	}
	_ = rect.SetSize([2]float64{900, 900})
	_ = l.Position().SetStaticValue([2]float64{960, 540})
	gs, err := l.RootGroup().AddGradientStroke()
	if err != nil {
		t.Fatalf("AddGradientStroke: %v", err)
	}
	_ = gs.SetStrokeWidth(80)
	_ = gs.SetStartPoint([2]float64{-450, 0})
	_ = gs.SetEndPoint([2]float64{450, 0})

	rbg := &aep.Gradient{
		Version: "4",
		ColorStops: []aep.GradientColorStop{
			{Offset: 0, Midpoint: 0.5, Color: [3]float64{1, 0, 0}},
			{Offset: 0.5, Midpoint: 0.5, Color: [3]float64{0, 0, 1}},
			{Offset: 1, Midpoint: 0.5, Color: [3]float64{0, 1, 0}},
		},
		AlphaStops: []aep.GradientAlphaStop{
			{Offset: 0, Midpoint: 0.5, Alpha: 1},
			{Offset: 1, Midpoint: 0.5, Alpha: 1},
		},
	}
	grb := &aep.Gradient{
		Version: "4",
		ColorStops: []aep.GradientColorStop{
			{Offset: 0, Midpoint: 0.5, Color: [3]float64{0, 1, 0}},
			{Offset: 0.5, Midpoint: 0.5, Color: [3]float64{1, 0, 0}},
			{Offset: 1, Midpoint: 0.5, Color: [3]float64{0, 0, 1}},
		},
		AlphaStops: []aep.GradientAlphaStop{
			{Offset: 0, Midpoint: 0.5, Alpha: 1},
			{Offset: 1, Midpoint: 0.5, Alpha: 1},
		},
	}
	if err := gs.AddGradientKeyframe(0, rbg); err != nil {
		t.Fatalf("AddGradientKeyframe 0: %v", err)
	}
	if err := gs.AddGradientKeyframe(1, grb); err != nil {
		t.Fatalf("AddGradientKeyframe 1: %v", err)
	}
	return p
}

// TestGradientStrokeAnimRoundtrip mirrors TestGradientAnimRoundtrip for a
// gradient STROKE: the animated `ADBE Vector Grad Colors` stream writes a
// keyframe time-table (lhd3 numKf=2, bpk=64) + 2 Utf8 stop-XML leaves.
func TestGradientStrokeAnimRoundtrip(t *testing.T) {
	p := buildGradStrokeAnim(t, aep.TargetAE2025)
	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	root, err := rifx.Parse(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	var gcst *rifx.Chunk
	var walk func(c *rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if c.IsList() && c.FormType == rifx.IDGCst {
			gcst = c
		}
		for _, ch := range c.Children {
			walk(ch)
		}
	}
	walk(root)
	if gcst == nil {
		t.Fatal("no GCst in animated gradient-stroke output")
	}
	var lhd3, ldat *rifx.Chunk
	var utf8Count int
	var dig func(c *rifx.Chunk)
	dig = func(c *rifx.Chunk) {
		for _, ch := range c.Children {
			switch ch.ID {
			case rifx.IDLhd3:
				lhd3 = ch
			case rifx.IDLdat:
				ldat = ch
			case rifx.IDUtf8:
				utf8Count++
			}
			if ch.IsList() {
				dig(ch)
			}
		}
	}
	dig(gcst)
	if lhd3 == nil || ldat == nil {
		t.Fatalf("animated gradient stroke missing time-table (lhd3=%v ldat=%v)", lhd3 != nil, ldat != nil)
	}
	if got := binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]); got != 2 {
		t.Errorf("lhd3 numKf = %d, want 2", got)
	}
	if got := binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]); got != 64 {
		t.Errorf("lhd3 bpk = %d, want 64", got)
	}
	if utf8Count != 2 {
		t.Errorf("GCky Utf8 leaves = %d, want 2", utf8Count)
	}
}

// TestGradientAnimRoundtrip is the pure-Go structural gate (no AE): an animated
// gradient writes a keyframe time-table (lhd3 numKf=2, bpk=64) in the tdbs and
// two Utf8 stop-XML leaves in the GCky, and re-parses cleanly.
func TestGradientAnimRoundtrip(t *testing.T) {
	p := buildGradAnim(t, aep.TargetAE2025)
	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	root, err := rifx.Parse(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}

	// Locate the GCst gradient-colors wrapper.
	var gcst *rifx.Chunk
	var walk func(c *rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if c.IsList() && c.FormType == rifx.IDGCst {
			gcst = c
		}
		for _, ch := range c.Children {
			walk(ch)
		}
	}
	walk(root)
	if gcst == nil {
		t.Fatal("no GCst in animated-gradient output")
	}

	// tdbs must carry the time-table (lhd3+ldat), GCky must carry 2 Utf8.
	var lhd3, ldat *rifx.Chunk
	var utf8Count int
	var dig func(c *rifx.Chunk)
	dig = func(c *rifx.Chunk) {
		for _, ch := range c.Children {
			switch ch.ID {
			case rifx.IDLhd3:
				lhd3 = ch
			case rifx.IDLdat:
				ldat = ch
			case rifx.IDUtf8:
				utf8Count++
			}
			if ch.IsList() {
				dig(ch)
			}
		}
	}
	dig(gcst)

	if lhd3 == nil || ldat == nil {
		t.Fatalf("animated gradient missing time-table (lhd3=%v ldat=%v)", lhd3 != nil, ldat != nil)
	}
	if got := binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]); got != 2 {
		t.Errorf("lhd3 numKf = %d, want 2", got)
	}
	if got := binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]); got != 64 {
		t.Errorf("lhd3 bpk = %d, want 64", got)
	}
	if len(ldat.Data) != 2*64 {
		t.Errorf("ldat len = %d, want 128", len(ldat.Data))
	}
	if utf8Count != 2 {
		t.Errorf("GCky Utf8 leaves = %d, want 2", utf8Count)
	}
	// kf1 time = round(1s * tickRate); the fixture's tickRate is 30720 (30fps).
	if t1 := binary.BigEndian.Uint32(ldat.Data[64:68]); t1 == 0 {
		t.Errorf("kf1 time = 0 — time not encoded")
	} else {
		t.Logf("kf0 time=%d kf1 time=%d (tickRate-scaled seconds)", binary.BigEndian.Uint32(ldat.Data[0:4]), t1)
	}
}
