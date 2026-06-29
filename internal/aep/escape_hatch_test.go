package aep_test

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestPropertyGroup_Vec2Stream(t *testing.T) {
	r := aep.NewRectNode()
	props := r.Properties()
	if props == nil {
		t.Fatal("RectNode.Properties() returned nil")
	}
	sizeStream, err := props.Vec2Stream("Size")
	if err != nil {
		t.Fatal(err)
	}
	if sizeStream != r.Size() {
		t.Fatal("escape hatch Vec2Stream(\"Size\") != typed accessor Size()")
	}
}

func TestPropertyGroup_NameNotFound_Error(t *testing.T) {
	r := aep.NewRectNode()
	props := r.Properties()
	if props == nil {
		t.Fatal("RectNode.Properties() returned nil")
	}
	if _, err := props.Vec2Stream("NotAFieldName"); err == nil {
		t.Fatal("unknown stream name should error")
	}
}

func TestPropertyGroup_Float64Stream(t *testing.T) {
	r := aep.NewRectNode()
	props := r.Properties()
	rndStream, err := props.Float64Stream("Roundness")
	if err != nil {
		t.Fatal(err)
	}
	if rndStream != r.Roundness() {
		t.Fatal("escape hatch Float64Stream(\"Roundness\") != typed accessor Roundness()")
	}
}

func TestPropertyGroup_TypeMismatch_Error(t *testing.T) {
	r := aep.NewRectNode()
	props := r.Properties()
	// "Size" is Vec2 — asking for Float64 should error type-mismatch
	if _, err := props.Float64Stream("Size"); err == nil {
		t.Fatal("Vec2 stream queried as Float64 should error")
	}
}

func TestPropertyGroup_FillColorStream(t *testing.T) {
	f := aep.NewFillNode()
	props := f.Properties()
	if props == nil {
		t.Fatal("FillNode.Properties() nil")
	}
	colorStream, err := props.ColorStream("Color")
	if err != nil {
		t.Fatal(err)
	}
	if colorStream != f.Color() {
		t.Fatal("ColorStream(\"Color\") != f.Color()")
	}
}

func TestPropertyGroup_PathStream(t *testing.T) {
	p := aep.NewPathNode()
	props := p.Properties()
	if props == nil {
		t.Fatal("PathNode.Properties() nil")
	}
	pathStream, err := props.PathStream("Path")
	if err != nil {
		t.Fatal(err)
	}
	if pathStream != p.Path() {
		t.Fatal("PathStream(\"Path\") != p.Path()")
	}
}

func TestPropertyGroup_EllipseStreams(t *testing.T) {
	e := aep.NewEllipseNode()
	props := e.Properties()
	if props == nil {
		t.Fatal("EllipseNode.Properties() nil")
	}
	if _, err := props.Vec2Stream("Size"); err != nil {
		t.Fatal(err)
	}
	if _, err := props.Vec2Stream("Position"); err != nil {
		t.Fatal(err)
	}
}

func TestPropertyGroup_StrokeStreams(t *testing.T) {
	s := aep.NewStrokeNode()
	props := s.Properties()
	if props == nil {
		t.Fatal("StrokeNode.Properties() nil")
	}
	if _, err := props.ColorStream("Color"); err != nil {
		t.Fatal(err)
	}
	if _, err := props.Float64Stream("Opacity"); err != nil {
		t.Fatal(err)
	}
	if _, err := props.Float64Stream("Width"); err != nil {
		t.Fatal(err)
	}
}
