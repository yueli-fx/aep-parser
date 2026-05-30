// internal/aep/lower_property_stream_test.go
//
// Chunk-shape level tests only. Byte-exact validation is
// deferred to the roundtrip tests.
package aep_test

import (
	"encoding/binary"
	"math"
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
	// Path uses om-s + omks + shap encoding, not cdat.
	if oms := chunk.FindFirstList(rifx.IDOmS); oms == nil {
		t.Fatal("path stream: tdgp missing LIST(om-s)")
	}
}

// TestEncodeBezier_LdatMatchesAELayout pins the ldat per-vertex layout RE'd
// from AE-native fixtures (test_data/v2_2_shape_path_re.aep, decoded via
// tmp_debug/decode_path_ldat): each vertex stores 6 f32 (bbox-normalized) =
//   [ anchor_i , anchor_i+outTangent_i , anchor_{(i+1)%n}+inTangent_{(i+1)%n} ]
// i.e. anchor, THIS vertex's out-control, and the NEXT vertex's in-control
// (wraps mod n). The pre-V2.2.1 encoder wrote [anchor, in_i, out_i] (this
// vertex's own in/out) — AE renders that as the wrong shape.
func TestEncodeBezier_LdatMatchesAELayout(t *testing.T) {
	// Distinct coords (not the ambiguous 0/1 square) so each slot is identifiable.
	ps := aep.NewPropertyStream[aep.BezierPath]()
	if err := ps.SetStaticValue(aep.BezierPath{
		Vertices:    [][2]float64{{10, 20}, {70, 30}, {40, 90}},
		InTangents:  [][2]float64{{0, 0}, {0, 0}, {0, 0}},
		OutTangents: [][2]float64{{0, 0}, {0, 0}, {0, 0}},
		Closed:      true,
	}); err != nil {
		t.Fatal(err)
	}
	chunk, err := aep.LowerPathStream(ps, "ADBE Vector Shape", "Path", aep.NewLowerCtxForTest())
	if err != nil {
		t.Fatal(err)
	}
	ldat := findIDRec(chunk, rifx.IDLdat)
	if ldat == nil {
		t.Fatal("no ldat")
	}
	if len(ldat.Data) != 3*24 {
		t.Fatalf("ldat len = %d, want 72", len(ldat.Data))
	}
	rf := func(off int) float64 {
		return float64(math.Float32frombits(binary.BigEndian.Uint32(ldat.Data[off : off+4])))
	}
	// bbox min(10,20) max(70,90) → rx=60 ry=70. Expected normalized 6-tuples.
	want := [][6]float64{
		{0, 0, /*out*/ 0, 0, /*next-in*/ 1, 10.0 / 70.0},
		{1, 10.0 / 70.0, /*out*/ 1, 10.0 / 70.0, /*next-in*/ 0.5, 1},
		{0.5, 1, /*out*/ 0.5, 1, /*next-in*/ 0, 0},
	}
	for v := 0; v < 3; v++ {
		for k := 0; k < 6; k++ {
			got := rf(v*24 + k*4)
			if math.Abs(got-want[v][k]) > 1e-5 {
				t.Errorf("vertex %d slot %d = %.6g, want %.6g", v, k, got, want[v][k])
			}
		}
	}
}

func findIDRec(c *rifx.Chunk, id rifx.ChunkID) *rifx.Chunk {
	for _, ch := range c.Children {
		if ch.ID == id {
			return ch
		}
		if ch.IsList() {
			if g := findIDRec(ch, id); g != nil {
				return g
			}
		}
	}
	return nil
}
