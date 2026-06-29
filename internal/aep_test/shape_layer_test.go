package aep_test

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestShapeLayer_Wraps_Layer(t *testing.T) {
	base := &aep.Layer{Type: aep.LayerTypeShape, Name: "test"}
	s := aep.WrapShapeLayer(base)
	if s.Name != "test" {
		t.Fatalf("ShapeLayer.Name (via embedded Layer) = %q, want %q", s.Name, "test")
	}
	if s.RootGroup() == nil {
		t.Fatal("ShapeLayer.RootGroup() nil")
	}
	if s.Transform() == nil {
		t.Fatal("ShapeLayer.Transform() nil")
	}
}

func TestShapeLayer_Transform_AccessorsReturnNonNilStreams(t *testing.T) {
	base := &aep.Layer{Type: aep.LayerTypeShape}
	s := aep.WrapShapeLayer(base)
	if s.Transform().Position() == nil {
		t.Fatal("Transform().Position() nil")
	}
	if s.Transform().AnchorPoint() == nil {
		t.Fatal("Transform().AnchorPoint() nil")
	}
	if s.Transform().Scale() == nil {
		t.Fatal("Transform().Scale() nil")
	}
	if s.Transform().Rotation() == nil {
		t.Fatal("Transform().Rotation() nil")
	}
	if s.Transform().Opacity() == nil {
		t.Fatal("Transform().Opacity() nil")
	}
	if s.Position() == nil {
		t.Fatal("shorthand Position() nil")
	}
	if s.Scale() == nil {
		t.Fatal("shorthand Scale() nil")
	}
	if s.Rotation() == nil {
		t.Fatal("shorthand Rotation() nil")
	}
	if s.Opacity() == nil {
		t.Fatal("shorthand Opacity() nil")
	}
}

func TestLayerTransform_Defaults(t *testing.T) {
	base := &aep.Layer{Type: aep.LayerTypeShape}
	s := aep.WrapShapeLayer(base)
	t0 := s.Transform()

	ap, _ := t0.AnchorPoint().StaticValue()
	if ap != [2]float64{0, 0} {
		t.Fatalf("AnchorPoint default = %v, want [0,0]", ap)
	}
	pos, _ := t0.Position().StaticValue()
	if pos != [2]float64{0, 0} {
		t.Fatalf("Position default = %v, want [0,0]", pos)
	}
	sc, _ := t0.Scale().StaticValue()
	if sc != [2]float64{100, 100} {
		t.Fatalf("Scale default = %v, want [100,100]", sc)
	}
	rot, _ := t0.Rotation().StaticValue()
	if rot != 0 {
		t.Fatalf("Rotation default = %v, want 0", rot)
	}
	op, _ := t0.Opacity().StaticValue()
	if op != 100 {
		t.Fatalf("Opacity default = %v, want 100", op)
	}
}
