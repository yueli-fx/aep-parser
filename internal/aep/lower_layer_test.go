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
	// iter-5b: every AE-saved Layr carries a 4th LIST(Gide) boilerplate
	// child. Absent → AE 2025 silently drops layer from comp.layers at
	// instantiation stage (verified via tmp_debug/bisect_v2_2 variant #2
	// empty ShapeLayer also dropping).
	gide := chunk.FindFirstList(rifx.IDGide)
	if gide == nil {
		t.Fatal("Layr missing LIST(Gide) boilerplate (iter-5b)")
	}
	if len(gide.Children) != 2 {
		t.Fatalf("Gide children = %d, want 2 (gdta + LIST list)", len(gide.Children))
	}
	if gide.Children[0].ID != rifx.IDGdta || len(gide.Children[0].Data) != 8 {
		t.Errorf("Gide[0] = %q (%dB), want gdta (8B)", gide.Children[0].ID, len(gide.Children[0].Data))
	}
}

func TestLowerShapeLayer_LdtaSizeByTarget(t *testing.T) {
	// AE 2020's ldta reader rejects a 164-B ShapeLayer ldta as corrupt
	// ("项目文件似乎已损坏（跳过部分：1）"), silently skipping the layer; AE 2025
	// reads its native 164 B. Each target must emit its native ldta size:
	// 160 B for AE 2020/2022, 164 B for AE 2025. Confirmed by AE-2020-native
	// reference (ldta 160 B) + AE 2020 ship-gate (see incident report
	// flightdeck/incident-reports/ae2020-shape-ldta-164-corrupt.md).
	cases := []struct {
		target aep.AETarget
		want   int
	}{
		{aep.TargetAE2020, 160},
		{aep.TargetAE2022, 160},
		{aep.TargetAE2025, 164},
	}
	for _, tc := range cases {
		base := &aep.Layer{Type: aep.LayerTypeShape, Name: "S1", ID: 13}
		s := aep.WrapShapeLayer(base)
		chunk, err := aep.LowerShapeLayerForTargetForTest(s, tc.target)
		if err != nil {
			t.Fatal(err)
		}
		ldta := chunk.FindFirst(rifx.IDLdta)
		if ldta == nil {
			t.Fatalf("target %d: ldta missing", tc.target)
		}
		if len(ldta.Data) != tc.want {
			t.Errorf("target %d: ldta size = %d, want %d", tc.target, len(ldta.Data), tc.want)
		}
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
