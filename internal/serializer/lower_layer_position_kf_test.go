package serializer

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/example/aep-parser/internal/codec"
	"github.com/example/aep-parser/internal/rifx"
)

// minimalTransformLeaf builds a tdmn(name) + LIST(tdbs)(tdsb+tdsn+tdb4+cdat)
// scaffold so the inject/overwrite helpers have a flippable static stream.
func minimalTransformLeaf(body *rifx.Chunk, name string, layout valueLayout) {
	tdbs := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdbs}
	tdbs.Children = append(tdbs.Children,
		makeTdsb(),
		makeTdsn(name),
		makeTdb4(layout),
		makeCdat(make([]byte, layout.dim*8), layout),
	)
	body.Children = append(body.Children, makeTdmn(name), tdbs)
}

func kfListOf(t *testing.T, body *rifx.Chunk, name string) (lhd3, ldat *rifx.Chunk) {
	t.Helper()
	for i := 0; i+1 < len(body.Children); i++ {
		if body.Children[i].ID == rifx.IDTdmn && trimChunkNUL(body.Children[i].Data) == name {
			tdbs := body.Children[i+1]
			kfl := tdbs.FindFirstList(rifx.ChunkID{'l', 'i', 's', 't'})
			if kfl == nil {
				t.Fatalf("%s: not flipped to animated (no LIST(list))", name)
			}
			if tdbs.FindFirst(rifx.IDCdat) != nil {
				t.Fatalf("%s: static cdat should have been replaced", name)
			}
			return kfl.FindFirst(rifx.IDLhd3), kfl.FindFirst(rifx.IDLdat)
		}
	}
	t.Fatalf("%s: tdmn not found", name)
	return nil, nil
}

func f64At(b []byte, off int) float64 {
	return math.Float64frombits(binary.BigEndian.Uint64(b[off : off+8]))
}

// TestLowerTransformVec2Spatial_Bpk128 pins Anchor/Position encoding (combined
// 3D spatial motion-path, bpk-128) against the RE'd fixture
// (test_data/v2_2_transform_kf_re.aep / v2_2_shape_kf_re.aep): value@0x38 X/Y/Z
// (Z=0), motion-path marker@0x08, header@0x07=0x07. KF2 = (500,300) at t=2s.
func TestLowerTransformVec2Spatial_Bpk128(t *testing.T) {
	body := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	minimalTransformLeaf(body, MatchNamePosition, valueLayout{dim: 3, headerByte: 0x07, spatial: true})
	ps := codec.NewPropertyStream[[2]float64]()
	_ = ps.AddKeyframeLinear(0, [2]float64{0, 0})
	_ = ps.AddKeyframeLinear(2, [2]float64{500, 300})

	if err := lowerTransformVec2Spatial(body, MatchNamePosition, ps, &lowerCtx{tickRate: 30720}); err != nil {
		t.Fatal(err)
	}
	lhd3, ldat := kfListOf(t, body, MatchNamePosition)
	if binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]) != 128 {
		t.Errorf("bpk = %d, want 128", binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]))
	}
	if len(ldat.Data) != 256 {
		t.Fatalf("ldat len = %d, want 256", len(ldat.Data))
	}
	blk := ldat.Data[128:256]
	if blk[0x07] != 0x07 {
		t.Errorf("@0x07 = %#x, want 0x07", blk[0x07])
	}
	if binary.BigEndian.Uint32(blk[0x08:0x0C]) != 1 {
		t.Errorf("motion-path marker@0x08 != 1")
	}
	if f64At(blk, 0x38) != 500 || f64At(blk, 0x40) != 300 || f64At(blk, 0x48) != 0 {
		t.Errorf("value = (%v,%v,%v), want (500,300,0)", f64At(blk, 0x38), f64At(blk, 0x40), f64At(blk, 0x48))
	}
}

// TestLowerTransformScale_Bpk128NonSpatial pins Scale: 3D non-spatial (bpk-128,
// value@0x08), percent÷100 with Z=1.0. KF2 = [150,200] → (1.5, 2.0, 1.0).
func TestLowerTransformScale_Bpk128NonSpatial(t *testing.T) {
	body := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	minimalTransformLeaf(body, MatchNameScale, valueLayout{dim: 3, headerByte: 0x00})
	ps := codec.NewPropertyStream[[2]float64]()
	_ = ps.AddKeyframeLinear(0, [2]float64{100, 100})
	_ = ps.AddKeyframeLinear(2, [2]float64{150, 200})

	if err := lowerTransformScale(body, ps, &lowerCtx{tickRate: 30720}); err != nil {
		t.Fatal(err)
	}
	lhd3, ldat := kfListOf(t, body, MatchNameScale)
	if binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]) != 128 {
		t.Errorf("bpk = %d, want 128", binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]))
	}
	blk := ldat.Data[128:256]
	if blk[0x07] != 0x00 {
		t.Errorf("@0x07 = %#x, want 0x00 (non-spatial)", blk[0x07])
	}
	if f64At(blk, 0x08) != 1.5 || f64At(blk, 0x10) != 2.0 || f64At(blk, 0x18) != 1.0 {
		t.Errorf("value = (%v,%v,%v), want (1.5,2.0,1.0)", f64At(blk, 0x08), f64At(blk, 0x10), f64At(blk, 0x18))
	}
}

// TestLowerShapeRectSubProps pins Rect Position (spatial Vec2 motion-path,
// bpk-104, value@0x38, like Ellipse Position) + Roundness (1D non-spatial,
// bpk-48) — RE'd from test_data/v2_2_rect_subprops.aep. KF2: pos=(40,50), rnd=20.
func TestLowerShapeRectSubProps(t *testing.T) {
	body := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	minimalTransformLeaf(body, "ADBE Vector Rect Position", valueLayout{dim: 2, headerByte: 0x07, spatial: true})
	minimalTransformLeaf(body, "ADBE Vector Rect Roundness", valueLayout{dim: 1, headerByte: 0x00})

	pos := codec.NewPropertyStream[[2]float64]()
	_ = pos.AddKeyframeLinear(0, [2]float64{0, 0})
	_ = pos.AddKeyframeLinear(2, [2]float64{40, 50})
	rnd := codec.NewPropertyStream[float64]()
	_ = rnd.AddKeyframeLinear(0, 0)
	_ = rnd.AddKeyframeLinear(2, 20)
	ctx := &lowerCtx{tickRate: 30720}

	if err := lowerShapeVec2(body, "ADBE Vector Rect Position", pos, ctx, valueLayout{dim: 2, headerByte: 0x07, spatial: true, motionPath: true}); err != nil {
		t.Fatal(err)
	}
	if err := lowerShapeScalar(body, "ADBE Vector Rect Roundness", rnd, ctx); err != nil {
		t.Fatal(err)
	}

	plhd3, pldat := kfListOf(t, body, "ADBE Vector Rect Position")
	if binary.BigEndian.Uint32(plhd3.Data[0x10:0x14]) != 104 {
		t.Errorf("Position bpk = %d, want 104", binary.BigEndian.Uint32(plhd3.Data[0x10:0x14]))
	}
	pblk := pldat.Data[104:208]
	if pblk[0x07] != 0x07 || binary.BigEndian.Uint32(pblk[0x08:0x0C]) != 1 {
		t.Errorf("Position KF2 header/motion-path wrong")
	}
	if f64At(pblk, 0x38) != 40 || f64At(pblk, 0x40) != 50 {
		t.Errorf("Position KF2 = (%v,%v), want (40,50)", f64At(pblk, 0x38), f64At(pblk, 0x40))
	}

	rlhd3, rldat := kfListOf(t, body, "ADBE Vector Rect Roundness")
	if binary.BigEndian.Uint32(rlhd3.Data[0x10:0x14]) != 48 {
		t.Errorf("Roundness bpk = %d, want 48", binary.BigEndian.Uint32(rlhd3.Data[0x10:0x14]))
	}
	if v := f64At(rldat.Data[48:96], 0x08); v != 20 {
		t.Errorf("Roundness KF2 = %v, want 20", v)
	}
}

// TestLowerTransformScalar pins Rotation (degrees as-is) + Opacity (÷100):
// 1D non-spatial, bpk-48, value@0x08.
func TestLowerTransformScalar(t *testing.T) {
	cases := []struct {
		name  string
		match string
		scale float64
		in    float64
		want  float64
	}{
		{"rotation", MatchNameRotateZ, 1, 90, 90},
		{"opacity", MatchNameOpacity, 0.01, 50, 0.5},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			body := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
			minimalTransformLeaf(body, tc.match, valueLayout{dim: 1, headerByte: 0x00})
			ps := codec.NewPropertyStream[float64]()
			_ = ps.AddKeyframeLinear(0, 0)
			_ = ps.AddKeyframeLinear(2, tc.in)

			if err := lowerTransformScalar(body, tc.match, ps, &lowerCtx{tickRate: 30720}, tc.scale); err != nil {
				t.Fatal(err)
			}
			lhd3, ldat := kfListOf(t, body, tc.match)
			if binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]) != 48 {
				t.Errorf("bpk = %d, want 48", binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]))
			}
			blk := ldat.Data[48:96]
			if got := f64At(blk, 0x08); got != tc.want {
				t.Errorf("KF2 value = %v, want %v", got, tc.want)
			}
		})
	}
}
