package aep

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/example/aep-parser/internal/rifx"
)

// TestInjectAnimatedLayerPosition_Bpk128 pins the combined "ADBE Position"
// keyframe encoding (Path B) against the bytes RE'd from
// test_data/v2_2_shape_kf_re.aep (tmp_debug/dump_kf "ADBE Position"):
//
//	lhd3 bpk@0x10 = 128 (= 0x38 + 3*dim*8 for dim=3 spatial)
//	per-keyframe block (128 B):
//	  @0x07 = 0x07 (spatial header)
//	  @0x08 = 0x00000001 (motion-path marker)
//	  @0x38 = X (f64 BE)   @0x40 = Y   @0x48 = Z (= 0; layer pos is 2D-in/3D-on-disk)
//
// Fixture KF2 = (500, 300) at t=2s; tickRate 30720 → time field 61440.
func TestInjectAnimatedLayerPosition_Bpk128(t *testing.T) {
	body := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp}
	tdbs := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdbs}
	tdbs.Children = append(tdbs.Children,
		makeTdsb(),
		makeTdsn("Position"),
		makeTdb4(valueLayout{dim: 3, headerByte: 0x07, spatial: true}),
		makeCdat(encode3D([3]float64{0, 0, 0}), valueLayout{dim: 3}),
	)
	body.Children = append(body.Children, makeTdmn(MatchNamePosition), tdbs)

	ctx := &lowerCtx{tickRate: 30720}
	kfs := []StreamKeyframe[[2]float64]{
		{Time: 0, Value: [2]float64{0, 0}},
		{Time: 2, Value: [2]float64{500, 300}},
	}
	if err := injectAnimatedLayerPosition(body, kfs, ctx); err != nil {
		t.Fatalf("injectAnimatedLayerPosition: %v", err)
	}

	kfList := tdbs.FindFirstList(rifx.ChunkID{'l', 'i', 's', 't'})
	if kfList == nil {
		t.Fatal("position not flipped to animated: no LIST(list) keyframe container")
	}
	if tdbs.FindFirst(rifx.IDCdat) != nil {
		t.Fatal("static cdat should have been replaced by keyframe container")
	}
	lhd3 := kfList.FindFirst(rifx.IDLhd3)
	ldat := kfList.FindFirst(rifx.IDLdat)
	if lhd3 == nil || ldat == nil {
		t.Fatal("keyframe container missing lhd3/ldat")
	}

	if got := binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]); got != 2 {
		t.Errorf("lhd3 keyframe count = %d, want 2", got)
	}
	if got := binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]); got != 128 {
		t.Errorf("lhd3 bpk = %d, want 128", got)
	}
	if len(ldat.Data) != 256 {
		t.Fatalf("ldat len = %d, want 256 (2 × 128)", len(ldat.Data))
	}

	// KF2 = second 128-byte block.
	blk := ldat.Data[128:256]
	if blk[0x07] != 0x07 {
		t.Errorf("KF2 @0x07 = %#x, want 0x07 (spatial header)", blk[0x07])
	}
	if got := binary.BigEndian.Uint32(blk[0x08:0x0C]); got != 1 {
		t.Errorf("KF2 @0x08 motion-path marker = %d, want 1", got)
	}
	if got := binary.BigEndian.Uint32(blk[0x00:0x04]); got != 61440 {
		t.Errorf("KF2 time = %d, want 61440 (2s × 30720)", got)
	}
	rf := func(off int) float64 {
		return math.Float64frombits(binary.BigEndian.Uint64(blk[off : off+8]))
	}
	if x := rf(0x38); x != 500 {
		t.Errorf("KF2 X @0x38 = %v, want 500", x)
	}
	if y := rf(0x40); y != 300 {
		t.Errorf("KF2 Y @0x40 = %v, want 300", y)
	}
	if z := rf(0x48); z != 0 {
		t.Errorf("KF2 Z @0x48 = %v, want 0", z)
	}
}
