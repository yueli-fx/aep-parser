package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestNewShapeLayer_BasicCreation(t *testing.T) {
	p := aep.NewProject()
	c, err := p.NewComposition("Main", 1920, 1080, 30, 10)
	if err != nil {
		t.Fatal(err)
	}
	s, err := c.NewShapeLayer("S1")
	if err != nil {
		t.Fatal(err)
	}
	if s == nil {
		t.Fatal("nil ShapeLayer")
	}
	if s.Name != "S1" {
		t.Fatalf("Name = %q, want S1", s.Name)
	}
	if s.Type != aep.LayerTypeShape {
		t.Fatalf("Type = %v, want LayerTypeShape", s.Type)
	}
	if len(c.Layers) != 1 || c.Layers[0] != s.Layer {
		t.Fatalf("comp.Layers not updated correctly: len=%d", len(c.Layers))
	}
}

func TestNewShapeLayer_EmptyName_Error(t *testing.T) {
	p := aep.NewProject()
	c, err := p.NewComposition("Main", 1920, 1080, 30, 10)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.NewShapeLayer(""); err == nil {
		t.Fatal("empty name should error")
	}
	if len(c.Layers) != 0 {
		t.Fatalf("comp polluted on failure: %d layers", len(c.Layers))
	}
}
