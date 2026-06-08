// internal/aep/shape_atomicity_test.go
//
// Atomicity / rollback tests. Verifies that
// failed mutations leave no observable side effect.
package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/codec"
)

func TestShapeLayer_NewWithEmptyName_DoesNotPolluteComp(t *testing.T) {
	p := aep.NewProject()
	c, err := p.NewComposition("M", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewShapeLayer(c, "ok"); err != nil {
		t.Fatalf("NewShapeLayer(ok): %v", err)
	}
	beforeLen := len(c.Layers)

	if _, err := aep.NewShapeLayer(c, ""); err == nil {
		t.Fatal("empty name should error")
	}
	if len(c.Layers) != beforeLen {
		t.Fatalf("comp.Layers polluted: was %d, now %d", beforeLen, len(c.Layers))
	}
}

func TestPropertyStream_AddKeyframe_NegativeTime_NoSideEffect(t *testing.T) {
	ps := codec.NewPropertyStream[float64]()
	if err := ps.AddKeyframeLinear(1.0, 100); err != nil {
		t.Fatalf("seed kf: %v", err)
	}
	beforeKf := len(ps.Keyframes())
	beforeMode := ps.Mode()

	if err := ps.AddKeyframeLinear(-1, 0); err == nil {
		t.Fatal("negative time should error")
	}
	if len(ps.Keyframes()) != beforeKf {
		t.Fatalf("kf count changed: was %d, now %d", beforeKf, len(ps.Keyframes()))
	}
	if ps.Mode() != beforeMode {
		t.Fatalf("mode changed: was %v, now %v", beforeMode, ps.Mode())
	}
}

func TestPropertyStream_AddKeyframe_DuplicateTime_NoSideEffect(t *testing.T) {
	ps := codec.NewPropertyStream[float64]()
	if err := ps.AddKeyframeLinear(1.0, 100); err != nil {
		t.Fatalf("seed kf: %v", err)
	}
	beforeKf := len(ps.Keyframes())

	if err := ps.AddKeyframeLinear(1.0, 999); err == nil {
		t.Fatal("duplicate time should error")
	}
	if got := len(ps.Keyframes()); got != beforeKf {
		t.Fatalf("kf count changed: was %d, now %d", beforeKf, got)
	}
	if v := ps.Keyframes()[0].Value; v != 100 {
		t.Fatalf("original kf value mutated: got %v, want 100", v)
	}
}
