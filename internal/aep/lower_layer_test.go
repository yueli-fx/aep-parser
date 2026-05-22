// internal/aep/lower_layer_test.go
//
// Phase 2 Task 2.3 tests — chunk-shape level only. ShapeLayer → LIST(Layr).
package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/rifx"
)

func TestLowerShapeLayer_EmptyHasLayrChunk(t *testing.T) {
	base := &aep.Layer{Type: aep.LayerTypeShape, Name: "S1"}
	s := aep.WrapShapeLayer(base)
	chunk, err := aep.LowerShapeLayerForTest(s)
	if err != nil {
		t.Fatal(err)
	}
	if chunk == nil {
		t.Fatal("nil")
	}
	if !chunk.IsList() || chunk.FormType != rifx.IDLayr {
		t.Fatalf("expected LIST(Layr), got id=%q ft=%q", chunk.ID, chunk.FormType)
	}
	// Empty ShapeLayer (no root-vector children): ldta + Utf8 + LIST(tdgp).
	if chunk.FindFirst(rifx.IDLdta) == nil {
		t.Fatal("Layr missing ldta")
	}
	if chunk.FindFirst(rifx.IDUtf8) == nil {
		t.Fatal("Layr missing Utf8 (layer name)")
	}
	if chunk.FindFirstList(rifx.IDTdgp) == nil {
		t.Fatal("Layr missing LIST(tdgp) — Transform Group")
	}
}

func TestLowerShapeLayer_LdtaIs160Bytes(t *testing.T) {
	base := &aep.Layer{Type: aep.LayerTypeShape, Name: "S1", ID: 13}
	s := aep.WrapShapeLayer(base)
	chunk, err := aep.LowerShapeLayerForTest(s)
	if err != nil {
		t.Fatal(err)
	}
	ldta := chunk.FindFirst(rifx.IDLdta)
	if ldta == nil {
		t.Fatal("ldta missing")
	}
	// AE 2020 canonical = 160 bytes (RE-S1; spec §4.2).
	if len(ldta.Data) != 160 {
		t.Fatalf("ldta size = %d, want 160 (AE 2020 canonical per RE-S1)", len(ldta.Data))
	}
}

func TestLowerShapeLayer_WithShape_HasRootVectorsGroup(t *testing.T) {
	base := &aep.Layer{Type: aep.LayerTypeShape, Name: "S1"}
	s := aep.WrapShapeLayer(base)
	s.RootGroup().Children = append(s.RootGroup().Children, aep.NewRectNode())
	chunk, err := aep.LowerShapeLayerForTest(s)
	if err != nil {
		t.Fatal(err)
	}
	// With a shape: Layr should hold tdmn("ADBE Root Vectors Group") +
	// LIST(tdgp, root vectors children).
	foundRoot := false
	for _, ch := range chunk.Children {
		if ch.ID == rifx.IDTdmn && len(ch.Data) >= 23 {
			// "ADBE Root Vectors Group" = 23 ASCII bytes.
			if string(ch.Data[0:23]) == "ADBE Root Vectors Group" {
				foundRoot = true
				break
			}
		}
	}
	if !foundRoot {
		t.Fatal("Layr missing tdmn(ADBE Root Vectors Group) when shape present")
	}
}
