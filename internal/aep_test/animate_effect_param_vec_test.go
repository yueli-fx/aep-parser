// internal/aep/animate_effect_param_vec_test.go
//
// Go round-trip + byte-structural verification for AnimateEffectParamVec
// (animated color / 2D-point / 3D-point effect params). Non-AE: a cheap
// pre-check before the AE render gate. Asserts each param survives WriteAEP +
// re-parse as a 2-keyframe animated stream with the right values, AND that the
// raw written bytes use the AE-native spatial layout (bpk 152/104/128, block
// header 0x01/0x07, per-type @0x08 marker 2/3, value at 0x38).
package aep_test

import (
	"encoding/binary"
	"github.com/yueli-fx/aep-parser/internal/serializer"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/rifx"
)

type vecAnimCase struct {
	effect, param string
	comps, bpk    int
	marker        uint32
	header        byte
	kfs           []aep.VectorKeyframe
}

var vecAnimCases = []vecAnimCase{
	{aep.EffectColorControl, "ADBE Color Control-0001", 4, 152, 2, 0x01,
		[]aep.VectorKeyframe{{Time: 0, Value: []float64{255, 255, 0, 0}}, {Time: 2, Value: []float64{255, 0, 0, 255}}}},
	{aep.EffectPointControl, "ADBE Point Control-0001", 2, 104, 3, 0x07,
		[]aep.VectorKeyframe{{Time: 0, Value: []float64{0.1, 0.2}}, {Time: 2, Value: []float64{0.8, 0.6}}}},
	{aep.EffectPoint3DControl, "ADBE Point3D Control-0001", 3, 128, 3, 0x07,
		[]aep.VectorKeyframe{{Time: 0, Value: []float64{0.1, 0.2, 0.05}}, {Time: 2, Value: []float64{0.8, 0.6, 0.25}}}},
}

func TestAnimateEffectParamVec_RoundTrip(t *testing.T) {
	p := aep.NewProject(aep.TargetAE2020)
	comp, err := aep.NewComposition(p, "AV", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := aep.NewShapeLayer(comp, "S"); err != nil {
		t.Fatal(err)
	}
	rp, err := aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	l := rp.Compositions[0].LayerByName("S")
	if l == nil {
		t.Fatal("layer S missing")
	}
	for _, c := range vecAnimCases {
		fx, err := aep.AddEffect(l, c.effect)
		if err != nil {
			t.Fatalf("AddEffect(%s): %v", c.effect, err)
		}
		if _, err := serializer.AnimateEffectParamVec(l, fx, c.param, c.kfs); err != nil {
			t.Fatalf("AnimateEffectParamVec(%s): %v", c.param, err)
		}
	}

	dir := t.TempDir()
	fpath := filepath.Join(dir, "av.aep")
	out, err := os.Create(fpath)
	if err != nil {
		t.Fatal(err)
	}
	if err := rp.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	// Parser-level readback.
	re, err := aep.Open(fpath)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	rl := re.Compositions[0].LayerByName("S")
	if rl == nil {
		t.Fatal("reopened layer S missing")
	}
	for _, c := range vecAnimCases {
		var prop *aep.Property
		for _, e := range rl.Effects {
			for _, pr := range e.Parameters {
				if pr.MatchName == c.param {
					prop = pr
				}
			}
		}
		if prop == nil {
			t.Errorf("%s not found after round-trip", c.param)
			continue
		}
		if prop.Components != c.comps {
			t.Errorf("%s Components=%d want %d", c.param, prop.Components, c.comps)
		}
		if len(prop.Keyframes) != 2 {
			t.Errorf("%s keyframes=%d want 2 (static→animated flip failed?)", c.param, len(prop.Keyframes))
			continue
		}
		got0, ok := prop.Keyframes[0].Value.([]float64)
		if !ok || len(got0) != c.comps {
			t.Errorf("%s kf0 value = %v (%T), want %d-component []float64", c.param, prop.Keyframes[0].Value, prop.Keyframes[0].Value, c.comps)
			continue
		}
		for i, w := range c.kfs[0].Value {
			if math.Abs(got0[i]-w) > 1e-6 {
				t.Errorf("%s kf0[%d]=%v want %v", c.param, i, got0[i], w)
			}
		}
		if math.Abs(prop.Keyframes[1].Time-2) > 1e-6 {
			t.Errorf("%s kf1 time=%v want 2", c.param, prop.Keyframes[1].Time)
		}
	}

	// Byte-structural: the written stream must use the AE-native spatial layout.
	root := parseAEP(t, fpath)
	for _, c := range vecAnimCases {
		kfl := findShipList(root, c.param)
		if kfl == nil {
			t.Errorf("%s: no animated kfl in written bytes", c.param)
			continue
		}
		var lhd3, ldat *rifx.Chunk
		for _, ch := range kfl.Children {
			if ch.ID == rifx.IDLhd3 {
				lhd3 = ch
			}
			if ch.ID == rifx.IDLdat {
				ldat = ch
			}
		}
		if lhd3 == nil || ldat == nil || len(ldat.Data) < c.bpk {
			t.Errorf("%s: kfl missing lhd3/ldat or short", c.param)
			continue
		}
		if cnt := binary.BigEndian.Uint32(lhd3.Data[0x08:]); cnt != 2 {
			t.Errorf("%s: lhd3 count=%d want 2", c.param, cnt)
		}
		if bpk := int(binary.BigEndian.Uint32(lhd3.Data[0x10:])); bpk != c.bpk {
			t.Errorf("%s: bpk=%d want %d", c.param, bpk, c.bpk)
		}
		blk := ldat.Data[:c.bpk]
		if blk[0x07] != c.header {
			t.Errorf("%s: block hdr@0x07=%02x want %02x", c.param, blk[0x07], c.header)
		}
		if m := binary.BigEndian.Uint32(blk[0x08:]); m != c.marker {
			t.Errorf("%s: spatial marker@0x08=%d want %d", c.param, m, c.marker)
		}
		if v := math.Float64frombits(binary.BigEndian.Uint64(blk[0x38:])); math.Abs(v-c.kfs[0].Value[0]) > 1e-6 {
			t.Errorf("%s: kf0 value@0x38=%v want %v (spatial value offset wrong)", c.param, v, c.kfs[0].Value[0])
		}
	}
}
