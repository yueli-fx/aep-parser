// internal/aep/shape_taper_wave_test.go
//
// Stroke Taper + Wave setters: model defaults + WriteAEP->FromReader roundtrip.
// All sub-streams are OneD float64-BE at cdat[0:8], one nesting level deeper
// than the top-level stroke scalars (inside the Taper/Wave LIST(tdgp) groups).
// V2.2 supports the always-active %/Wavelength-mode scalars only; Length Units /
// StartWidthPx / EndWidthPx / Wave Units / Cycles are AE-elided and not modeled.
// RE: incident-reports/stroke-line-cap-join-miter-re.md (Taper/Wave addendum).
package aep_test

import (
	"bytes"
	"math"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestStrokeTaperWave_Defaults(t *testing.T) {
	s := aep.NewStrokeNode()
	tp, wv := s.Taper(), s.Wave()
	if tp == nil || wv == nil {
		t.Fatal("Taper()/Wave() must be non-nil on a fresh StrokeNode")
	}
	// Taper defaults: all 0 (no taper).
	for _, c := range []struct {
		name string
		got  float64
	}{
		{"StartLength", tp.StartLength()}, {"EndLength", tp.EndLength()},
		{"StartWidth", tp.StartWidth()}, {"EndWidth", tp.EndWidth()},
		{"StartEase", tp.StartEase()}, {"EndEase", tp.EndEase()},
	} {
		if c.got != 0 {
			t.Errorf("taper %s default = %g, want 0", c.name, c.got)
		}
	}
	// Wave defaults: Amount 0, Wavelength 100, Phase 0.
	if wv.Amount() != 0 || wv.Wavelength() != 100 || wv.Phase() != 0 {
		t.Errorf("wave defaults: amount=%g wavelength=%g phase=%g, want 0/100/0",
			wv.Amount(), wv.Wavelength(), wv.Phase())
	}
}

func TestV2_2_StrokeTaperWave_Roundtrip(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "Main", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	l, err := aep.NewShapeLayer(comp, "TaperWave")
	if err != nil {
		t.Fatalf("NewShapeLayer: %v", err)
	}
	rect, _ := l.RootGroup().AddRect()
	_ = rect.SetSize([2]float64{200, 100})
	stroke, _ := l.RootGroup().AddStroke()
	_ = stroke.SetWidth(12)
	// Values distinct from the embedded template (25/35/50/60/40/45, 15/80/90)
	// so the roundtrip proves the cdat overwrite, not template passthrough.
	tp := stroke.Taper()
	_ = tp.SetStartLength(30)
	_ = tp.SetEndLength(40)
	_ = tp.SetStartWidth(55)
	_ = tp.SetEndWidth(65)
	_ = tp.SetStartEase(35)
	_ = tp.SetEndEase(50)
	wv := stroke.Wave()
	_ = wv.SetAmount(20)
	_ = wv.SetWavelength(70)
	_ = wv.SetPhase(45)

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	sl := aep.WrapShapeLayer(findLayerByName(re.Compositions[0], "TaperWave"))
	var ss *aep.StrokeNode
	for _, ch := range sl.RootGroup().Children {
		if n, ok := ch.(*aep.StrokeNode); ok {
			ss = n
		}
	}
	if ss == nil {
		t.Fatal("stroke node not recovered after roundtrip")
	}
	rtp, rwv := ss.Taper(), ss.Wave()
	eq := func(name string, got, want float64) {
		if math.Abs(got-want) > 1e-9 {
			t.Errorf("%s = %g, want %g", name, got, want)
		}
	}
	eq("taper StartLength", rtp.StartLength(), 30)
	eq("taper EndLength", rtp.EndLength(), 40)
	eq("taper StartWidth", rtp.StartWidth(), 55)
	eq("taper EndWidth", rtp.EndWidth(), 65)
	eq("taper StartEase", rtp.StartEase(), 35)
	eq("taper EndEase", rtp.EndEase(), 50)
	eq("wave Amount", rwv.Amount(), 20)
	eq("wave Wavelength", rwv.Wavelength(), 70)
	eq("wave Phase", rwv.Phase(), 45)
}
