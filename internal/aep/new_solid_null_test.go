// internal/aep/new_solid_null_test.go
//
// Go round-trip tests for NewSolidLayer / NewNullLayer / NewAdjustmentLayer
// (no AE). Proves the cross-Project template import splices a layer + backing
// solid footage item that survive WriteAEP → re-parse with the caller's name,
// color, dimensions, and ldta flags. AE acceptance is the ship-gate's job
// (new_solid_null_shipgate_test.go).
package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// buildSolidTrio creates a fresh project with one comp holding a solid + null
// + adjustment layer, round-trips it through WriteAEP → FromReader, and
// returns the re-parsed project + comp.
func buildSolidTrio(t *testing.T, target aep.AETarget) (*aep.Project, *aep.Composition) {
	t.Helper()
	p := aep.NewProject(target)
	comp, err := aep.NewComposition(p, "Main", 1280, 720, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aep.NewSolidLayer(comp, "Red BG", 1280, 720, [3]float64{0.75, 0.25, 0.5}); err != nil {
		t.Fatalf("NewSolidLayer: %v", err)
	}
	if _, err := aep.NewNullLayer(comp, "Controller"); err != nil {
		t.Fatalf("NewNullLayer: %v", err)
	}
	if _, err := aep.NewAdjustmentLayer(comp, "Grade"); err != nil {
		t.Fatalf("NewAdjustmentLayer: %v", err)
	}

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	var rc *aep.Composition
	for _, c := range re.Compositions {
		if c.Name == "Main" {
			rc = c
		}
	}
	if rc == nil {
		t.Fatalf("comp Main not found after round-trip")
	}
	return re, rc
}

func trioLayer(t *testing.T, c *aep.Composition, name string) *aep.Layer {
	t.Helper()
	for _, l := range c.Layers {
		if l.Name == name {
			return l
		}
	}
	t.Fatalf("layer %q not found (have %d layers)", name, len(c.Layers))
	return nil
}

func TestNewSolidLayer_RoundTrip(t *testing.T) {
	re, rc := buildSolidTrio(t, aep.TargetAE2020)
	l := trioLayer(t, rc, "Red BG")
	if l.Type != aep.LayerTypeAV {
		t.Errorf("Type = %v, want AV", l.Type)
	}
	if l.IsNull || l.IsAdjust {
		t.Errorf("solid layer flags: IsNull=%v IsAdjust=%v, want both false", l.IsNull, l.IsAdjust)
	}
	if l.SourceID == 0 {
		t.Fatal("solid layer SourceID = 0, want backing footage item")
	}
	f := trioFootage(t, re, l.SourceID)
	if !f.IsSolid {
		t.Errorf("backing footage IsSolid = false")
	}
	if f.Name != "Red BG" {
		t.Errorf("footage name = %q, want %q", f.Name, "Red BG")
	}
	if f.Width != 1280 || f.Height != 720 {
		t.Errorf("footage dims = %dx%d, want 1280x720", f.Width, f.Height)
	}
	if want := [3]float64{0.75, 0.25, 0.5}; f.SolidColor != want {
		t.Errorf("SolidColor = %v, want %v", f.SolidColor, want)
	}
	src, ok := f.MainSource().(*aep.SolidSource)
	if !ok {
		t.Fatalf("MainSource = %T, want *SolidSource", f.MainSource())
	}
	if src.Color != f.SolidColor {
		t.Errorf("MainSource color = %v, want %v", src.Color, f.SolidColor)
	}
}

func TestNewNullLayer_RoundTrip(t *testing.T) {
	re, rc := buildSolidTrio(t, aep.TargetAE2020)
	l := trioLayer(t, rc, "Controller")
	if !l.IsNull {
		t.Error("IsNull = false, want true")
	}
	if l.IsAdjust {
		t.Error("IsAdjust = true, want false")
	}
	f := trioFootage(t, re, l.SourceID)
	if f.Width != 100 || f.Height != 100 {
		t.Errorf("null footage dims = %dx%d, want 100x100", f.Width, f.Height)
	}
	if f.Name != "Controller" {
		t.Errorf("footage name = %q, want %q", f.Name, "Controller")
	}
}

func TestNewAdjustmentLayer_RoundTrip(t *testing.T) {
	re, rc := buildSolidTrio(t, aep.TargetAE2020)
	l := trioLayer(t, rc, "Grade")
	if !l.IsAdjust {
		t.Error("IsAdjust = false, want true")
	}
	if l.IsNull {
		t.Error("IsNull = true, want false")
	}
	f := trioFootage(t, re, l.SourceID)
	if f.Width != 1280 || f.Height != 720 {
		t.Errorf("adjustment footage dims = %dx%d, want comp-sized 1280x720", f.Width, f.Height)
	}
}

// Each call creates its own backing footage item (no dedup for solids —
// mirrors AE, which makes one solid source per New).
func TestNewSolidBacked_DistinctFootagePerCall(t *testing.T) {
	re, rc := buildSolidTrio(t, aep.TargetAE2020)
	seen := map[uint32]bool{}
	for _, name := range []string{"Red BG", "Controller", "Grade"} {
		l := trioLayer(t, rc, name)
		if seen[l.SourceID] {
			t.Errorf("layer %q shares SourceID %d with another layer", name, l.SourceID)
		}
		seen[l.SourceID] = true
		trioFootage(t, re, l.SourceID)
	}
}

// Time span must be re-homed into the dest comp (template carries a 10s span;
// dest comp here is 5s).
func TestNewSolidBacked_TimeSpanRehomed(t *testing.T) {
	_, rc := buildSolidTrio(t, aep.TargetAE2020)
	for _, name := range []string{"Red BG", "Controller", "Grade"} {
		l := trioLayer(t, rc, name)
		if l.StartTime != 0 {
			t.Errorf("layer %q StartTime = %v, want 0", name, l.StartTime)
		}
		if l.Duration != 5 {
			t.Errorf("layer %q Duration = %v, want 5 (comp duration)", name, l.Duration)
		}
	}
}

func TestNewSolidLayer_Validation(t *testing.T) {
	p := aep.NewProject(aep.TargetAE2020)
	comp, _ := aep.NewComposition(p, "Main", 1280, 720, 30, 5)
	cases := []struct {
		desc string
		fn   func() error
	}{
		{"empty name", func() error {
			_, err := aep.NewSolidLayer(comp, "", 100, 100, [3]float64{0, 0, 0})
			return err
		}},
		{"zero width", func() error {
			_, err := aep.NewSolidLayer(comp, "S", 0, 100, [3]float64{0, 0, 0})
			return err
		}},
		{"oversize height", func() error {
			_, err := aep.NewSolidLayer(comp, "S", 100, 30001, [3]float64{0, 0, 0})
			return err
		}},
		{"color out of range", func() error {
			_, err := aep.NewSolidLayer(comp, "S", 100, 100, [3]float64{0, 1.5, 0})
			return err
		}},
		{"null empty name", func() error {
			_, err := aep.NewNullLayer(comp, "")
			return err
		}},
		{"adjustment empty name", func() error {
			_, err := aep.NewAdjustmentLayer(comp, "")
			return err
		}},
	}
	for _, tc := range cases {
		if tc.fn() == nil {
			t.Errorf("%s: want error, got nil", tc.desc)
		}
	}
	if len(comp.Layers) != 0 {
		t.Errorf("failed validations left %d layers behind", len(comp.Layers))
	}
}

func trioFootage(t *testing.T, p *aep.Project, id uint32) *aep.Footage {
	t.Helper()
	for _, f := range p.Footage {
		if f.ID == id {
			return f
		}
	}
	t.Fatalf("footage id=%d not found (have %d footage items)", id, len(p.Footage))
	return nil
}
