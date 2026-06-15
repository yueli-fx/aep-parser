package serializer

import (
	"encoding/binary"
	"fmt"
	"math"

	"github.com/example/aep-parser/internal/codec"
	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

// Property-level keyframe insert/delete (length-variable ldat rebuild that
// re-parses the property's keyframe stream). These are structural serializer
// free-functions reaching the concrete property/keyframe back-refs; the pure
// scene setters (SetStaticValue / SetExpressionEnabled / SetExpression) live
// in internal/scene.

// InsertKeyframe builds a new bpk-byte keyframe block and inserts it
// into the property's ldat stream, then updates the lhd3 count header.
// Returns the new Keyframe and its index in Property.Keyframes
// (insertion is time-sorted; ties land after existing keys at the
// same time).
// (Full contract + RE notes live on the aep.InsertKeyframe facade — docgen source.)
func InsertKeyframe(p *Property, time float64, value any) (*Keyframe, int, error) {
	pb := propertyBack(p)
	if pb == nil || pb.ldat == nil || pb.lhd3 == nil {
		return nil, -1, fmt.Errorf("property %q: no existing keyframes (insert from scratch not supported)", p.MatchName)
	}
	if pb.bytesPerKF <= 0 {
		return nil, -1, fmt.Errorf("property %q: bytesPerKF is zero", p.MatchName)
	}
	if time < 0 {
		return nil, -1, fmt.Errorf("property %q: negative time %v", p.MatchName, time)
	}

	// Decide insertion index (time-sorted; ties go after).
	insertIdx := len(p.Keyframes)
	for i, kf := range p.Keyframes {
		if time < kf.Time {
			insertIdx = i
			break
		}
	}

	tickRate := aeLegacyTimeBase
	if len(p.Keyframes) > 0 {
		if kb := keyframeBack(p.Keyframes[0]); kb != nil {
			tickRate = kb.tickRate
		}
	}
	if tickRate == 0 {
		tickRate = aeLegacyTimeBase
	}

	// Build the new block: clone header byte @0x07 from an existing
	// keyframe (sample[0]) so layout matches; zero everything else.
	bpk := pb.bytesPerKF
	if len(pb.ldat.Data) < bpk {
		return nil, -1, fmt.Errorf("property %q: ldat shorter than one block (%d bytes, bpk=%d)", p.MatchName, len(pb.ldat.Data), bpk)
	}
	headerByte := pb.ldat.Data[0x07]
	block := make([]byte, bpk)
	binary.BigEndian.PutUint32(block[0:4], uint32(math.Round(time*tickRate)))
	block[0x04] = byte(scene.InterpLinear)
	block[0x05] = byte(scene.InterpLinear)
	block[0x07] = headerByte

	// Write value at the layout-appropriate offset.
	l := layoutFor(headerByte, p.Components)
	switch v := value.(type) {
	case float64:
		if p.Components != 1 {
			return nil, -1, fmt.Errorf("property %q is %dD; expected []float64", p.MatchName, p.Components)
		}
		if err := writeFloat64(block, l.valueOff, v); err != nil {
			return nil, -1, err
		}
	case []float64:
		if len(v) != p.Components {
			return nil, -1, fmt.Errorf("property %q is %dD; got %d-component value", p.MatchName, p.Components, len(v))
		}
		for i, x := range v {
			if err := writeFloat64(block, l.valueOff+i*8, x); err != nil {
				return nil, -1, err
			}
		}
	default:
		return nil, -1, fmt.Errorf("property %q: unsupported value type %T", p.MatchName, value)
	}

	// Splice the block into ldat at insertIdx * bpk.
	splice := insertIdx * bpk
	old := pb.ldat.Data
	newData := make([]byte, 0, len(old)+bpk)
	newData = append(newData, old[:splice]...)
	newData = append(newData, block...)
	newData = append(newData, old[splice:]...)
	pb.ldat.Data = newData

	// Bump count in lhd3 @0x08.
	if len(pb.lhd3.Data) >= 0x0C {
		newCount := binary.BigEndian.Uint32(pb.lhd3.Data[0x08:0x0C]) + 1
		binary.BigEndian.PutUint32(pb.lhd3.Data[0x08:0x0C], newCount)
	}

	// Rebuild Property.Keyframes — the simplest correct path. Re-points
	// every Keyframe.offset to its new position in the resized ldat.
	if err := reparseKeyframes(p, tickRate); err != nil {
		return nil, -1, fmt.Errorf("property %q: reparse after InsertKeyframe: %w", p.MatchName, err)
	}
	if insertIdx >= len(p.Keyframes) {
		insertIdx = len(p.Keyframes) - 1
	}
	return p.Keyframes[insertIdx], insertIdx, nil
}

// DeleteKeyframe removes the keyframe at index i from the property's
// ldat stream and decrements the lhd3 count header. Returns an error
// when i is out of range or the property has no keyframe stream.
// (Full contract + RE notes live on the aep.DeleteKeyframe facade — docgen source.)
func DeleteKeyframe(p *Property, i int) error {
	pb := propertyBack(p)
	if pb == nil || pb.ldat == nil || pb.lhd3 == nil {
		return fmt.Errorf("property %q: no keyframe stream", p.MatchName)
	}
	if i < 0 || i >= len(p.Keyframes) {
		return fmt.Errorf("property %q: keyframe index %d out of range [0,%d)", p.MatchName, i, len(p.Keyframes))
	}
	bpk := pb.bytesPerKF
	splice := i * bpk
	old := pb.ldat.Data
	if splice+bpk > len(old) {
		return fmt.Errorf("property %q: ldat shorter than expected (off=%d, bpk=%d, len=%d)", p.MatchName, splice, bpk, len(old))
	}
	newData := make([]byte, 0, len(old)-bpk)
	newData = append(newData, old[:splice]...)
	newData = append(newData, old[splice+bpk:]...)
	pb.ldat.Data = newData

	if len(pb.lhd3.Data) >= 0x0C {
		newCount := binary.BigEndian.Uint32(pb.lhd3.Data[0x08:0x0C])
		if newCount > 0 {
			newCount--
		}
		binary.BigEndian.PutUint32(pb.lhd3.Data[0x08:0x0C], newCount)
	}

	tickRate := aeLegacyTimeBase
	if len(p.Keyframes) > 0 {
		if kb := keyframeBack(p.Keyframes[0]); kb != nil {
			tickRate = kb.tickRate
		}
	}
	if tickRate == 0 {
		tickRate = aeLegacyTimeBase
	}
	return reparseKeyframes(p, tickRate)
}

// AnimateScalarKeyframes converts a STATIC 1D-scalar property into an animated
// one by synthesizing the keyframe container from scratch (the case
// InsertKeyframe refuses). It replaces the property's static cdat with a
// LIST(list){lhd3, ldat} keyframe stream built by encodeKeyframes (non-spatial
// 1D layout: bpk 48, time@0x00, value@0x08) and flips the tdb4 static→animated
// flags (@0x05 clear bit0, @0x44 = 1, @0x4f clear bit0 — byte-verified against
// an AE-saved animated Gaussian-Blur-Blurriness fixture, identical to the shape
// injectAnimatedStream patch). Used for animated effect parameters (animate a
// blur amount / a Slider Control). The property must be parsed (it needs its
// tdbs back-ref) and currently static (no existing keyframes — use InsertKeyframe
// to add to an already-animated stream). 1D scalar only.
// (Full contract lives on the aep.AnimateEffectParam facade — docgen source.)
func AnimateScalarKeyframes(p *Property, tickRate float64, kfs []ScalarKeyframe) error {
	if p == nil {
		return fmt.Errorf("AnimateScalarKeyframes: nil property")
	}
	if len(kfs) < 2 {
		return fmt.Errorf("AnimateScalarKeyframes: need >= 2 keyframes, got %d", len(kfs))
	}
	if p.Components != 1 {
		return fmt.Errorf("AnimateScalarKeyframes: property %q is %dD; only 1D scalars supported", p.MatchName, p.Components)
	}
	pb := propertyBack(p)
	if pb == nil || pb.tdbs == nil {
		return fmt.Errorf("AnimateScalarKeyframes: property %q has no tdbs back-ref (built outside parser?)", p.MatchName)
	}
	if pb.ldat != nil {
		return fmt.Errorf("AnimateScalarKeyframes: property %q already animated; use InsertKeyframe", p.MatchName)
	}
	if pb.cdat == nil {
		return fmt.Errorf("AnimateScalarKeyframes: property %q has no cdat to convert", p.MatchName)
	}
	if pb.tdb4 == nil || len(pb.tdb4.Data) <= 0x4f {
		return fmt.Errorf("AnimateScalarKeyframes: property %q tdb4 missing/short", p.MatchName)
	}
	if tickRate <= 0 {
		tickRate = aeLegacyTimeBase
	}

	streamKfs := make([]codec.StreamKeyframe[float64], len(kfs))
	for i, kf := range kfs {
		streamKfs[i] = codec.StreamKeyframe[float64]{
			Time:    kf.Time,
			Value:   kf.Value,
			InEase:  kf.InEase,
			OutEase: kf.OutEase,
		}
	}
	kfList, err := encodeKeyframes(streamKfs, valueLayout{dim: 1, headerByte: 0x00, spatial: false}, encode1D, &lowerCtx{tickRate: tickRate})
	if err != nil {
		return fmt.Errorf("AnimateScalarKeyframes: %w", err)
	}

	// Flip tdb4 static→animated (same patch as injectAnimatedStream).
	pb.tdb4.Data[0x05] &^= 0x01
	pb.tdb4.Data[0x44] = 0x01
	pb.tdb4.Data[0x4f] &^= 0x01

	// Swap the static cdat for the keyframe container, in place (AE keeps it
	// where the cdat was — between tdb4 and any tdum/tduM).
	replaced := false
	for i, ch := range pb.tdbs.Children {
		if ch == pb.cdat {
			pb.tdbs.Children[i] = kfList
			replaced = true
			break
		}
	}
	if !replaced {
		return fmt.Errorf("AnimateScalarKeyframes: property %q cdat not found in tdbs", p.MatchName)
	}
	pb.cdat = nil
	pb.lhd3 = kfList.FindFirst(rifx.IDLhd3)
	pb.ldat = kfList.FindFirst(rifx.IDLdat)
	if pb.lhd3 == nil || pb.ldat == nil {
		return fmt.Errorf("AnimateScalarKeyframes: built keyframe container missing lhd3/ldat")
	}
	return reparseKeyframes(p, tickRate)
}

// vectorKeyframeLayout returns the AE-native animated layout for a
// multi-component effect parameter, keyed by component count. Effect color (4D)
// and 2D/3D point params all use the SPATIAL keyframe block (value at 0x38,
// bpk = 0x38 + 3*dim*8) with a per-type @0x08 marker — RE'd byte-for-byte from
// re_anim_effect_colorpoint.aep: color = block header 0x01 / marker 2; point =
// block header 0x07 / marker 3.
func vectorKeyframeLayout(dim int) valueLayout {
	if dim == 4 {
		return valueLayout{dim: 4, headerByte: 0x01, spatial: true, spatialMarker: 2}
	}
	return valueLayout{dim: dim, headerByte: 0x07, spatial: true, spatialMarker: 3}
}

// AnimateVectorKeyframes converts a STATIC 2/3/4-component property into an
// animated one — the multi-component analogue of AnimateScalarKeyframes (the
// case InsertKeyframe refuses). It replaces the property's static cdat with a
// LIST(list){lhd3, ldat} keyframe stream built by encodeKeyframes (the SPATIAL
// layout effect color/point params use on disk) and flips the same tdb4
// static→animated bits (@0x05 clear bit0, @0x44 = 1, @0x4f clear bit0 — verified
// type-agnostic: color 07/00/01→06/01/00, point 0f/00/01→0e/01/00, identical
// transition to the 1D scalar 01/00/00→00/01/00). Keyframe values are taken in
// the property's on-disk units (same as SetEffectParam: color [A,R,G,B] 0-255,
// 2D/3D point = fraction of the layer coord space). The property must be parsed,
// currently static, and 2/3/4-component.
// (Full contract lives on the aep.AnimateEffectParamVec facade — docgen source.)
func AnimateVectorKeyframes(p *Property, tickRate float64, kfs []VectorKeyframe) error {
	return animateVectorKeyframes(p, tickRate, kfs, false)
}

// AnimateVectorKeyframesNonSpatial is AnimateVectorKeyframes for a NON-SPATIAL
// multi-component property (e.g. the text Scale 3D animator leaf): its animated
// keyframe block places the value at 0x08 with per-component temporal ease
// (bpk = 0x08 + 5*dim*8), NOT the spatial 0x38 tangent block effect color/point
// params use. RE'd byte-matching the AE-saved text Scale 3D leaf
// (re_text_animator_animatedvec): bpk 128, value@0x08, header byte 0x00, no
// marker. Same parsed-property contract as AnimateVectorKeyframes.
func AnimateVectorKeyframesNonSpatial(p *Property, tickRate float64, kfs []VectorKeyframe) error {
	return animateVectorKeyframes(p, tickRate, kfs, true)
}

func animateVectorKeyframes(p *Property, tickRate float64, kfs []VectorKeyframe, nonSpatial bool) error {
	who := "AnimateVectorKeyframes"
	if nonSpatial {
		who = "AnimateVectorKeyframesNonSpatial"
	}
	if p == nil {
		return fmt.Errorf("%s: nil property", who)
	}
	if len(kfs) < 2 {
		return fmt.Errorf("%s: need >= 2 keyframes, got %d", who, len(kfs))
	}
	dim := p.Components
	if dim < 2 || dim > 4 {
		return fmt.Errorf("%s: property %q is %dD; only 2/3/4D supported (use AnimateScalarKeyframes for 1D)", who, p.MatchName, dim)
	}
	for i, kf := range kfs {
		if len(kf.Value) != dim {
			return fmt.Errorf("%s: keyframe %d has %d-component value, property %q is %dD", who, i, len(kf.Value), p.MatchName, dim)
		}
	}
	pb := propertyBack(p)
	if pb == nil || pb.tdbs == nil {
		return fmt.Errorf("%s: property %q has no tdbs back-ref (built outside parser?)", who, p.MatchName)
	}
	if pb.ldat != nil {
		return fmt.Errorf("%s: property %q already animated; use InsertKeyframe", who, p.MatchName)
	}
	if pb.cdat == nil {
		return fmt.Errorf("%s: property %q has no cdat to convert", who, p.MatchName)
	}
	if pb.tdb4 == nil || len(pb.tdb4.Data) <= 0x4f {
		return fmt.Errorf("%s: property %q tdb4 missing/short", who, p.MatchName)
	}
	if tickRate <= 0 {
		tickRate = aeLegacyTimeBase
	}

	streamKfs := make([]codec.StreamKeyframe[[]float64], len(kfs))
	for i, kf := range kfs {
		streamKfs[i] = codec.StreamKeyframe[[]float64]{
			Time:    kf.Time,
			Value:   kf.Value,
			InEase:  kf.InEase,
			OutEase: kf.OutEase,
		}
	}
	enc := func(v []float64) []byte {
		b := make([]byte, dim*8)
		for i := range dim {
			binary.BigEndian.PutUint64(b[i*8:(i+1)*8], math.Float64bits(v[i]))
		}
		return b
	}
	layout := vectorKeyframeLayout(dim)
	if nonSpatial {
		// Non-spatial multi-dim block: value@0x08, header byte 0x00, no marker
		// (text Scale 3D leaf — re_text_animator_animatedvec).
		layout = valueLayout{dim: dim, headerByte: 0x00, spatial: false}
	}
	kfList, err := encodeKeyframes(streamKfs, layout, enc, &lowerCtx{tickRate: tickRate})
	if err != nil {
		return fmt.Errorf("%s: %w", who, err)
	}

	// Flip tdb4 static→animated (same patch as the scalar / shape paths).
	pb.tdb4.Data[0x05] &^= 0x01
	pb.tdb4.Data[0x44] = 0x01
	pb.tdb4.Data[0x4f] &^= 0x01

	replaced := false
	for i, ch := range pb.tdbs.Children {
		if ch == pb.cdat {
			pb.tdbs.Children[i] = kfList
			replaced = true
			break
		}
	}
	if !replaced {
		return fmt.Errorf("%s: property %q cdat not found in tdbs", who, p.MatchName)
	}
	pb.cdat = nil
	pb.lhd3 = kfList.FindFirst(rifx.IDLhd3)
	pb.ldat = kfList.FindFirst(rifx.IDLdat)
	if pb.lhd3 == nil || pb.ldat == nil {
		return fmt.Errorf("%s: built keyframe container missing lhd3/ldat", who)
	}
	return reparseKeyframes(p, tickRate)
}

// reparseKeyframes rebuilds Property.Keyframes from the current
// ldat/lhd3 bytes. Used after InsertKeyframe / DeleteKeyframe so
// Keyframe.offset / Value / etc. reflect the new stream layout.
func reparseKeyframes(p *Property, tickRate float64) error {
	pb := propertyBack(p)
	if pb == nil || pb.ldat == nil || pb.lhd3 == nil {
		return fmt.Errorf("property %q: missing ldat/lhd3", p.MatchName)
	}
	if len(pb.lhd3.Data) < 0x14 {
		return fmt.Errorf("property %q: lhd3 too short (%d bytes)", p.MatchName, len(pb.lhd3.Data))
	}
	count := int(binary.BigEndian.Uint32(pb.lhd3.Data[0x08:0x0C]))
	bpk := int(binary.BigEndian.Uint32(pb.lhd3.Data[0x10:0x14]))
	if count < 0 || bpk <= 0 || count*bpk > len(pb.ldat.Data) {
		return fmt.Errorf("property %q: lhd3 says count=%d bpk=%d but ldat has %d bytes", p.MatchName, count, bpk, len(pb.ldat.Data))
	}
	pb.bytesPerKF = bpk
	p.Keyframes = make([]*Keyframe, 0, count)
	for i := 0; i < count; i++ {
		off := i * bpk
		kf := &Keyframe{}
		setKeyframeBack(kf, &keyframeBackrefs{
			ldat:     pb.ldat,
			offset:   off,
			dims:     p.Components,
			tickRate: tickRate,
		})
		kf.Time = float64(binary.BigEndian.Uint32(pb.ldat.Data[off:off+4])) / tickRate
		kf.Value = readKFValue(pb.ldat.Data, off, p.Components)
		decodeEasing(kf, pb.ldat.Data[off:off+bpk])
		p.Keyframes = append(p.Keyframes, kf)
	}
	return nil
}
