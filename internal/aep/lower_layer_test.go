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

func TestLowerShapeLayer_LdtaIs164Bytes(t *testing.T) {
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
	// Iter 4 RE: AE 2025 saves ShapeLayer ldta as 164 B (trailing 4 B zero).
	// V2.2 builder bumped to 164 B so user Layr matches AE-saved fixture
	// byte-for-byte. AE 2020 has been observed to accept 164 B too.
	if len(ldta.Data) != 164 {
		t.Fatalf("ldta size = %d, want 164 (AE 2025 canonical per iter 4 RE)", len(ldta.Data))
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
	// With a shape: Layr's outer LIST(tdgp) wrapper (fix A) holds
	// tdmn("ADBE Root Vectors Group") + LIST(tdgp, root vectors children).
	// Walk descendants — the wrapper depth can change as more
	// property-group placeholders are added in future ship-gate iters.
	var walk func(c *rifx.Chunk) bool
	walk = func(c *rifx.Chunk) bool {
		for _, ch := range c.Children {
			if ch.ID == rifx.IDTdmn && len(ch.Data) >= 23 &&
				string(ch.Data[0:23]) == "ADBE Root Vectors Group" {
				return true
			}
			if ch.IsList() && walk(ch) {
				return true
			}
		}
		return false
	}
	if !walk(chunk) {
		t.Fatal("Layr missing tdmn(ADBE Root Vectors Group) when shape present")
	}
}
