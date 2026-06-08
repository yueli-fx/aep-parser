package aep

import (
	"encoding/binary"

	"github.com/example/aep-parser/internal/rifx"
)

// parseKeyframes decodes a keyframe stream into prop.Keyframes.
//
// lhd3 layout (52 bytes, observed):
//
//	0x00–0x07 : magic / version (constant `00 D0 0B EE 00 00 00 00`)
//	0x08–0x0B : keyframe count (uint32 BE)
//	0x10–0x13 : bytes per keyframe block (uint32 BE; varies by dim+style)
//
// Per-keyframe layout (within ldat):
//
//	0x00–0x03 : time (uint32 BE, per-comp ticks; conversion in ctx.tickRate)
//	0x04      : in-interpolation (1=linear, 2=bezier, 3=hold)
//	0x05      : out-interpolation
//	0x06      : reserved/flags
//	0x07      : 0x07 = spatial 3D (Position/Anchor); 0x01 = 4D color or
//	             mask-path-time; 0x00 = non-spatial N-D.
//
// See decodeEasing for the per-style ease/tangent layouts.
func parseKeyframes(prop *Property, lhd3, ldat *rifx.Chunk, ctx *parseCtx) {
	if len(lhd3.Data) < 0x14 {
		ctx.warn("keyframe header (lhd3) for property %q is %d bytes, need >= 20; skipping keyframes",
			prop.MatchName, len(lhd3.Data))
		return
	}
	count := int(binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]))
	bpk := int(binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]))
	if count <= 0 || bpk <= 0 || count*bpk > len(ldat.Data) {
		ctx.warn("keyframe stream for property %q inconsistent (count=%d, bytes_per_kf=%d, ldat=%d bytes); skipping keyframes",
			prop.MatchName, count, bpk, len(ldat.Data))
		return
	}
	if prop.back == nil {
		prop.back = &propertyBackrefs{}
	}
	pb := prop.propertyBack()
	pb.ldat = ldat
	pb.lhd3 = lhd3
	pb.bytesPerKF = bpk
	prop.Keyframes = make([]*Keyframe, 0, count)

	// Pre-flight: check bpk against the layout deduced from the first
	// keyframe's header byte. When bpk is smaller than the layout demands
	// decodeEasing silently skips ease + tangent decode for every block —
	// emit ONE warning here rather than letting easing data vanish without
	// a trace. We still decode Time + Value (which only need the first
	// 8 bytes + valueOff worth of the block).
	if bpk >= 8 {
		l := layoutFor(ldat.Data[0x07], prop.Components)
		need := 8 + 5*prop.Components*8 // non-spatial layout
		layoutName := "non-spatial"
		if l.spatialStyle {
			need = 0x38 + 3*prop.Components*8
			layoutName = "spatial-style"
		}
		if bpk < need {
			ctx.warn("keyframe block for property %q is %d bytes, %s %dD layout expects >= %d; easing/tangents will be skipped",
				prop.MatchName, bpk, layoutName, prop.Components, need)
		}
	}

	for i := 0; i < count; i++ {
		offset := i * bpk
		kf := &Keyframe{
			back: &keyframeBackrefs{
				ldat:     ldat,
				offset:   offset,
				dims:     prop.Components,
				tickRate: ctx.tickRate,
				compFps:  ctx.compFps,
			},
		}
		kf.Time = float64(binary.BigEndian.Uint32(ldat.Data[offset:offset+4])) / ctx.tickRate
		kf.Value = readKFValue(ldat.Data, offset, prop.Components)
		decodeEasing(kf, ldat.Data[offset:offset+bpk])
		prop.Keyframes = append(prop.Keyframes, kf)
	}
}

// kfLayout describes the byte layout of a single keyframe block. The
// header byte at offset 0x07 plus the property's component count fully
// determine the layout; everything (value position, ease shape, tangent
// presence) follows from those two facts.
//
// Two styles exist:
//
//   - SPATIAL-STYLE (Position, Anchor Point 3D, 4D color): scalar (single)
//     in/out temporal ease at 0x18/0x20/0x28/0x30; values at 0x38 (N
//     doubles); per-component in-tangent at 0x38+N*8 (N doubles);
//     per-component out-tangent at 0x38+2*N*8 (N doubles).
//     bpk = 0x38 + 3*N*8.
//
//   - NON-SPATIAL (Opacity 1D, Mask Feather 2D, Scale 3D, ...): values at
//     0x08 (N doubles); per-component temporal speed at 0x08+(N+i)*8;
//     per-component temporal influence at 0x08+(2N+i)*8; out-side mirrors
//     at 0x08+(3N+i)*8 and 0x08+(4N+i)*8.
//     bpk = 0x08 + 5*N*8 (= 48 for 1D, 88 for 2D, 128 for 3D).
type kfLayout struct {
	valueOff     int  // byte offset of value[0]
	spatialStyle bool // see doc above
}

// layoutFor selects the keyframe layout from the header byte + dim count.
// dims must be >= 1. Header values: 0x07 = spatial 3D (Position/Anchor);
// 0x01 = 4D color (Tritone-style) OR mask-path-time stream (callers
// avoid that overload); 0x00 = non-spatial N-D.
func layoutFor(header07 byte, dims int) kfLayout {
	if header07 == 0x07 || (header07 == 0x01 && dims >= 2) {
		return kfLayout{valueOff: 0x38, spatialStyle: true}
	}
	return kfLayout{valueOff: 0x08, spatialStyle: false}
}

// decodeEasing populates a keyframe's interpolation + tangent fields from
// its raw block bytes (already sliced to exactly bpk bytes). See kfLayout
// for the byte layouts.
func decodeEasing(kf *Keyframe, blk []byte) {
	if len(blk) < 8 {
		return
	}
	kf.InInterp = InterpType(blk[0x04])
	kf.OutInterp = InterpType(blk[0x05])
	dims := 0
	if kf.back != nil {
		if kb, ok := kf.back.(*keyframeBackrefs); ok {
			dims = kb.dims
		}
	}
	if dims <= 0 {
		dims = 1
	}
	l := layoutFor(blk[0x07], dims)

	if l.spatialStyle {
		need := 0x38 + 3*dims*8
		if len(blk) < need {
			// Header looks spatial-style but block is too short — likely
			// mask-path-time stream (bpk=64). Caller handles those.
			return
		}
		inSpd, _ := readFloat64BE(blk, 0x18)
		inInf, _ := readFloat64BE(blk, 0x20)
		outSpd, _ := readFloat64BE(blk, 0x28)
		outInf, _ := readFloat64BE(blk, 0x30)
		kf.InTemporalEase = []TemporalEase{{Speed: inSpd, Influence: inInf}}
		kf.OutTemporalEase = []TemporalEase{{Speed: outSpd, Influence: outInf}}
		inTanBase := l.valueOff + dims*8
		outTanBase := inTanBase + dims*8
		kf.InSpatialTangent = make([]float64, dims)
		kf.OutSpatialTangent = make([]float64, dims)
		for i := 0; i < dims; i++ {
			v, _ := readFloat64BE(blk, inTanBase+i*8)
			kf.InSpatialTangent[i] = v
			v2, _ := readFloat64BE(blk, outTanBase+i*8)
			kf.OutSpatialTangent[i] = v2
		}
		return
	}

	// Non-spatial: per-component temporal ease, no spatial tangents.
	need := 8 + 5*dims*8
	if len(blk) < need {
		return
	}
	kf.InTemporalEase = make([]TemporalEase, dims)
	kf.OutTemporalEase = make([]TemporalEase, dims)
	for i := 0; i < dims; i++ {
		inSpd, _ := readFloat64BE(blk, 8+(dims+i)*8)
		inInf, _ := readFloat64BE(blk, 8+(2*dims+i)*8)
		outSpd, _ := readFloat64BE(blk, 8+(3*dims+i)*8)
		outInf, _ := readFloat64BE(blk, 8+(4*dims+i)*8)
		kf.InTemporalEase[i] = TemporalEase{Speed: inSpd, Influence: inInf}
		kf.OutTemporalEase[i] = TemporalEase{Speed: outSpd, Influence: outInf}
	}
}

// readKFValue reads the value(s) of a keyframe block, dispatching on the
// per-block layout. `kfOffset` is the start of the block inside d.
func readKFValue(d []byte, kfOffset, components int) any {
	if kfOffset+8 > len(d) {
		return nil
	}
	l := layoutFor(d[kfOffset+0x07], components)
	if components <= 1 {
		v, _ := readFloat64BE(d, kfOffset+l.valueOff)
		return v
	}
	vals := make([]float64, components)
	for i := 0; i < components; i++ {
		v, _ := readFloat64BE(d, kfOffset+l.valueOff+i*8)
		vals[i] = v
	}
	return vals
}

// kfValueOffset is a thin wrapper around layoutFor for callers that only
// need the value offset (e.g. write.go's SetValue, which already knows
// the dim count via k.back.dims).
func kfValueOffset(d []byte, kfOffset, dims int) int {
	if kfOffset+8 > len(d) {
		return 0x08
	}
	return layoutFor(d[kfOffset+0x07], dims).valueOff
}
