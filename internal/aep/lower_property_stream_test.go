// internal/aep/lower_property_stream_test.go
//
// Phase 2 Task 2.1 tests — chunk-shape level only. Byte-exact validation is
// deferred to Phase 4 roundtrip (per V2.2 plan §Phase 2 "Skeleton-level checks
// for chunk shape; byte-exact tests deferred to Phase 4 roundtrip.").
package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/rifx"
)

func TestLowerFloat64Stream_Static(t *testing.T) {
	ps := aep.NewPropertyStream[float64]()
	if err := ps.SetStaticValue(0.5); err != nil {
		t.Fatal(err)
	}
	chunk, err := aep.LowerFloat64Stream(ps, "ADBE Opacity", "Opacity", aep.NewLowerCtxForTest())
	if err != nil {
		t.Fatal(err)
	}
	if chunk == nil {
		t.Fatal("nil chunk")
	}
	if !chunk.IsList() || chunk.FormType != rifx.IDTdgp {
		t.Fatalf("expected LIST(tdgp), got id=%q ft=%q", chunk.ID, chunk.FormType)
	}
	// children: tdmn + LIST(tdbs)
	if len(chunk.Children) != 2 {
		t.Fatalf("tdgp child count = %d, want 2 (tdmn + LIST(tdbs))", len(chunk.Children))
	}
	if chunk.Children[0].ID != rifx.IDTdmn {
		t.Fatalf("child[0] id = %q, want tdmn", chunk.Children[0].ID)
	}
	tdbs := chunk.Children[1]
	if !tdbs.IsList() || tdbs.FormType != rifx.IDTdbs {
		t.Fatalf("child[1] not LIST(tdbs): id=%q ft=%q", tdbs.ID, tdbs.FormType)
	}
	// static mode: tdbs must contain a cdat (not a LIST(list)).
	if cdat := tdbs.FindFirst(rifx.IDCdat); cdat == nil {
		t.Fatal("static stream: tdbs missing cdat child")
	}
	if kfList := tdbs.FindFirstList(rifx.ChunkID{'l', 'i', 's', 't'}); kfList != nil {
		t.Fatal("static stream: unexpected LIST(list) keyframe container")
	}
}

func TestLowerVec2Stream_Animated(t *testing.T) {
	ps := aep.NewPropertyStream[[2]float64]()
	if err := ps.AddKeyframeLinear(0, [2]float64{0, 0}); err != nil {
		t.Fatal(err)
	}
	if err := ps.AddKeyframeLinear(2, [2]float64{500, 300}); err != nil {
		t.Fatal(err)
	}
	chunk, err := aep.LowerVec2Stream(ps, "ADBE Position", "Position", aep.NewLowerCtxForTest())
	if err != nil {
		t.Fatal(err)
	}
	if chunk == nil {
		t.Fatal("nil chunk")
	}
	tdbs := chunk.FindFirstList(rifx.IDTdbs)
	if tdbs == nil {
		t.Fatal("tdgp missing LIST(tdbs)")
	}
	// Animated mode: tdbs must hold a LIST(list) with lhd3 + ldat.
	kfList := tdbs.FindFirstList(rifx.ChunkID{'l', 'i', 's', 't'})
	if kfList == nil {
		t.Fatal("animated stream: tdbs missing LIST(list) keyframe container")
	}
	if kfList.FindFirst(rifx.IDLhd3) == nil {
		t.Fatal("animated stream: LIST(list) missing lhd3")
	}
	if kfList.FindFirst(rifx.IDLdat) == nil {
		t.Fatal("animated stream: LIST(list) missing ldat")
	}
}

func TestLowerColorStream_Static_CdatSize(t *testing.T) {
	ps := aep.NewPropertyStream[[4]float64]()
	if err := ps.SetStaticValue([4]float64{1, 0, 0, 1}); err != nil {
		t.Fatal(err)
	}
	chunk, err := aep.LowerColorStream(ps, "ADBE Vector Fill Color", "Color", aep.NewLowerCtxForTest())
	if err != nil {
		t.Fatal(err)
	}
	tdbs := chunk.FindFirstList(rifx.IDTdbs)
	cdat := tdbs.FindFirst(rifx.IDCdat)
	if cdat == nil {
		t.Fatal("color cdat missing")
	}
	if len(cdat.Data) < 4*8 {
		t.Fatalf("color cdat len = %d, want >= 32 (4 × float64)", len(cdat.Data))
	}
}

func TestLowerPathStream_Static_EmitsOmS(t *testing.T) {
	ps := aep.NewPropertyStream[aep.BezierPath]()
	if err := ps.SetStaticValue(aep.BezierPath{
		Vertices:    [][2]float64{{0, 0}, {100, 0}, {100, 100}, {0, 100}},
		InTangents:  [][2]float64{{0, 0}, {0, 0}, {0, 0}, {0, 0}},
		OutTangents: [][2]float64{{0, 0}, {0, 0}, {0, 0}, {0, 0}},
		Closed:      true,
	}); err != nil {
		t.Fatal(err)
	}
	chunk, err := aep.LowerPathStream(ps, "ADBE Vector Shape", "Path", aep.NewLowerCtxForTest())
	if err != nil {
		t.Fatal(err)
	}
	if chunk == nil {
		t.Fatal("nil chunk")
	}
	// Path uses om-s + omks + shap encoding, not cdat (RE-S5b).
	if oms := chunk.FindFirstList(rifx.IDOmS); oms == nil {
		t.Fatal("path stream: tdgp missing LIST(om-s)")
	}
}
