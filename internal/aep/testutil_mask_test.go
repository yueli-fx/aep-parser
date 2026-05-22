package aep_test

import (
	"encoding/binary"
	"math"

	aep "github.com/example/aep-parser/internal/aep"
)

// Mask-atom builders for synthetic RIFX fixtures. Used by mask_test.go
// for path / mkif / animation / static-property scenarios.
//
// All builders return raw inner bytes (no enclosing LIST header) unless
// the name says otherwise (buildMaskShapBytes returns a full LIST chunk).
//
// Cross-helper dependencies (kept in testutil_keyframe_test.go / testutil_test.go):
//   - rifxBuilder, rb.chunk, rb.listChunk           — testutil_test.go
//   - buildLhd3, buildTdb4                          — testutil_keyframe_test.go

// buildMaskShap writes the "shap" LIST body (children: shph, list/lhd3+ldat,
// omtn) given a list of (anchor, in, out) triplets. Returns the inner bytes
// of the shap LIST (without the LIST header).
func buildMaskShap(rb *rifxBuilder, verts []aep.MaskVertex, closed bool, name string) []byte {
	shph := make([]byte, 24)
	if closed {
		shph[0x14] = 0x01
	}
	shphChunk := rb.chunk("shph", shph)

	// ldat: 3 float32 X/Y pairs per vertex.
	var ldatData []byte
	for _, v := range verts {
		for _, pt := range [][2]float64{v.Anchor, v.InTangent, v.OutTangent} {
			pair := make([]byte, 8)
			binary.BigEndian.PutUint32(pair[0:], math.Float32bits(float32(pt[0])))
			binary.BigEndian.PutUint32(pair[4:], math.Float32bits(float32(pt[1])))
			ldatData = append(ldatData, pair...)
		}
	}
	count := uint32(len(verts) * 3)
	lhd3 := rb.chunk("lhd3", buildLhd3(count, 8))
	ldat := rb.chunk("ldat", ldatData)
	var kflInner []byte
	kflInner = append(kflInner, lhd3...)
	kflInner = append(kflInner, ldat...)
	kflList := rb.listChunk("LIST", "list", kflInner)

	omtn := rb.chunk("omtn", []byte(name))

	var shapInner []byte
	shapInner = append(shapInner, shphChunk...)
	shapInner = append(shapInner, kflList...)
	shapInner = append(shapInner, omtn...)
	return shapInner
}

// buildMkif creates a 48-byte mkif chunk payload with the given mode /
// inverted flag / 1-based index / color.
func buildMkif(mode aep.MaskMode, inverted bool, index uint32, r, g, b uint8) []byte {
	b48 := make([]byte, 48)
	if inverted {
		b48[0x00] = 0x01
	}
	binary.BigEndian.PutUint32(b48[0x04:], uint32(mode))
	binary.BigEndian.PutUint32(b48[0x08:], index)
	b48[0x2C] = 0xFF
	b48[0x2D] = r
	b48[0x2E] = g
	b48[0x2F] = b
	return b48
}

// buildMaskAtom builds one "ADBE Mask Atom" entry: tdmn + mkif + tdgp body
// containing the "ADBE Mask Shape" om-s wrapper around the shap LIST.
func buildMaskAtom(rb *rifxBuilder, verts []aep.MaskVertex, closed bool, name string) []byte {
	return buildMaskAtomWithMkif(rb, verts, closed, name,
		buildMkif(aep.MaskModeAdd, false, 1, 0xE4, 0xD8, 0x4C))
}

func buildMaskAtomWithMkif(rb *rifxBuilder, verts []aep.MaskVertex, closed bool, name string, mkifBytes []byte) []byte {
	shapList := rb.listChunk("LIST", "shap", buildMaskShap(rb, verts, closed, name))
	omksList := rb.listChunk("LIST", "omks", shapList)

	tdb4 := rb.chunk("tdb4", buildTdb4(0x01))
	tdbsList := rb.listChunk("LIST", "tdbs", tdb4)
	var omSInner []byte
	omSInner = append(omSInner, tdbsList...)
	omSInner = append(omSInner, omksList...)
	omSList := rb.listChunk("LIST", "om-s", omSInner)

	var maskInnerBody []byte
	maskInnerBody = append(maskInnerBody, rb.chunk("tdmn", []byte("ADBE Mask Shape"))...)
	maskInnerBody = append(maskInnerBody, omSList...)
	maskInnerBody = append(maskInnerBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	maskInnerTdgp := rb.listChunk("LIST", "tdgp", maskInnerBody)

	mkif := rb.chunk("mkif", mkifBytes)

	var out []byte
	out = append(out, rb.chunk("tdmn", []byte("ADBE Mask Atom"))...)
	out = append(out, mkif...)
	out = append(out, maskInnerTdgp...)
	return out
}

// buildMaskShapBytes builds a single shap LIST chunk (with header) for the
// given vertices + closed flag + name. Returns the full LIST chunk bytes.
func buildMaskShapBytes(rb *rifxBuilder, verts []aep.MaskVertex, closed bool, name string) []byte {
	return rb.listChunk("LIST", "shap", buildMaskShap(rb, verts, closed, name))
}

// buildAnimatedMaskAtom builds an "ADBE Mask Atom" with N path keyframes.
// Each keyframe has its own vertices and a time (seconds).
func buildAnimatedMaskAtom(rb *rifxBuilder, snapshots []struct {
	Time     float64
	Vertices []aep.MaskVertex
}, closed bool, name string) []byte {
	// omks LIST: one shap per snapshot.
	var omksBody []byte
	for _, s := range snapshots {
		omksBody = append(omksBody, buildMaskShapBytes(rb, s.Vertices, closed, name)...)
	}
	omksList := rb.listChunk("LIST", "omks", omksBody)

	// om-s tdbs: kfl LIST with 64-byte timing blocks (time at 0x00, plus
	// scalar in/out temporal ease at 0x18/0x20/0x28/0x30 when bezier).
	var ldatData []byte
	for _, s := range snapshots {
		blk := make([]byte, 64)
		binary.BigEndian.PutUint32(blk[0:], uint32(math.Round(s.Time*8000)))
		blk[0x04] = byte(aep.InterpLinear)
		blk[0x05] = byte(aep.InterpLinear)
		blk[0x07] = 0x01
		ldatData = append(ldatData, blk...)
	}
	lhd3 := rb.chunk("lhd3", buildLhd3(uint32(len(snapshots)), 64))
	ldat := rb.chunk("ldat", ldatData)
	var kflInner []byte
	kflInner = append(kflInner, lhd3...)
	kflInner = append(kflInner, ldat...)
	kflList := rb.listChunk("LIST", "list", kflInner)
	tdb4 := rb.chunk("tdb4", buildTdb4(0x01))
	var tdbsInner []byte
	tdbsInner = append(tdbsInner, tdb4...)
	tdbsInner = append(tdbsInner, kflList...)
	tdbsList := rb.listChunk("LIST", "tdbs", tdbsInner)

	var omSInner []byte
	omSInner = append(omSInner, tdbsList...)
	omSInner = append(omSInner, omksList...)
	omSList := rb.listChunk("LIST", "om-s", omSInner)

	var maskInnerBody []byte
	maskInnerBody = append(maskInnerBody, rb.chunk("tdmn", []byte("ADBE Mask Shape"))...)
	maskInnerBody = append(maskInnerBody, omSList...)
	maskInnerBody = append(maskInnerBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	maskInnerTdgp := rb.listChunk("LIST", "tdgp", maskInnerBody)

	mkif := rb.chunk("mkif", make([]byte, 48))
	var out []byte
	out = append(out, rb.chunk("tdmn", []byte("ADBE Mask Atom"))...)
	out = append(out, mkif...)
	out = append(out, maskInnerTdgp...)
	return out
}

// buildMaskAtomWithProps is buildMaskAtom + extra static properties
// (Feather/Opacity/Expansion) inside the mask atom's body tdgp.
func buildMaskAtomWithProps(rb *rifxBuilder, verts []aep.MaskVertex, closed bool, name string,
	mkifBytes, featherProp, opacityProp, expansionProp []byte) []byte {

	shapList := rb.listChunk("LIST", "shap", buildMaskShap(rb, verts, closed, name))
	omksList := rb.listChunk("LIST", "omks", shapList)
	tdb4 := rb.chunk("tdb4", buildTdb4(0x01))
	tdbsList := rb.listChunk("LIST", "tdbs", tdb4)
	var omSInner []byte
	omSInner = append(omSInner, tdbsList...)
	omSInner = append(omSInner, omksList...)
	omSList := rb.listChunk("LIST", "om-s", omSInner)

	var maskInnerBody []byte
	maskInnerBody = append(maskInnerBody, rb.chunk("tdmn", []byte("ADBE Mask Shape"))...)
	maskInnerBody = append(maskInnerBody, omSList...)
	if featherProp != nil {
		maskInnerBody = append(maskInnerBody, featherProp...)
	}
	if opacityProp != nil {
		maskInnerBody = append(maskInnerBody, opacityProp...)
	}
	if expansionProp != nil {
		maskInnerBody = append(maskInnerBody, expansionProp...)
	}
	maskInnerBody = append(maskInnerBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	maskInnerTdgp := rb.listChunk("LIST", "tdgp", maskInnerBody)

	mkif := rb.chunk("mkif", mkifBytes)
	var out []byte
	out = append(out, rb.chunk("tdmn", []byte("ADBE Mask Atom"))...)
	out = append(out, mkif...)
	out = append(out, maskInnerTdgp...)
	return out
}

// buildAnimatedMaskAtomEased is like buildAnimatedMaskAtom but writes
// bezier interp + Easy-Ease scalar temporal-ease values in the 64-byte
// timing blocks.
func buildAnimatedMaskAtomEased(rb *rifxBuilder, snapshots []struct {
	Time     float64
	Vertices []aep.MaskVertex
}, closed bool, name string) []byte {
	var omksBody []byte
	for _, s := range snapshots {
		omksBody = append(omksBody, buildMaskShapBytes(rb, s.Vertices, closed, name)...)
	}
	omksList := rb.listChunk("LIST", "omks", omksBody)

	var ldatData []byte
	for _, s := range snapshots {
		blk := make([]byte, 64)
		binary.BigEndian.PutUint32(blk[0:], uint32(math.Round(s.Time*8000)))
		blk[0x04] = byte(aep.InterpBezier)
		blk[0x05] = byte(aep.InterpBezier)
		blk[0x07] = 0x01
		// In speed=0, in influence=0.333 (Easy Ease).
		binary.BigEndian.PutUint64(blk[0x18:], math.Float64bits(0.0))
		binary.BigEndian.PutUint64(blk[0x20:], math.Float64bits(0.333))
		binary.BigEndian.PutUint64(blk[0x28:], math.Float64bits(0.0))
		binary.BigEndian.PutUint64(blk[0x30:], math.Float64bits(0.333))
		ldatData = append(ldatData, blk...)
	}
	lhd3 := rb.chunk("lhd3", buildLhd3(uint32(len(snapshots)), 64))
	ldat := rb.chunk("ldat", ldatData)
	var kflInner []byte
	kflInner = append(kflInner, lhd3...)
	kflInner = append(kflInner, ldat...)
	kflList := rb.listChunk("LIST", "list", kflInner)
	tdb4 := rb.chunk("tdb4", buildTdb4(0x01))
	var tdbsInner []byte
	tdbsInner = append(tdbsInner, tdb4...)
	tdbsInner = append(tdbsInner, kflList...)
	tdbsList := rb.listChunk("LIST", "tdbs", tdbsInner)

	var omSInner []byte
	omSInner = append(omSInner, tdbsList...)
	omSInner = append(omSInner, omksList...)
	omSList := rb.listChunk("LIST", "om-s", omSInner)

	var maskInnerBody []byte
	maskInnerBody = append(maskInnerBody, rb.chunk("tdmn", []byte("ADBE Mask Shape"))...)
	maskInnerBody = append(maskInnerBody, omSList...)
	maskInnerBody = append(maskInnerBody, rb.chunk("tdmn", []byte("ADBE Group End"))...)
	maskInnerTdgp := rb.listChunk("LIST", "tdgp", maskInnerBody)

	mkif := rb.chunk("mkif", make([]byte, 48))
	var out []byte
	out = append(out, rb.chunk("tdmn", []byte("ADBE Mask Atom"))...)
	out = append(out, mkif...)
	out = append(out, maskInnerTdgp...)
	return out
}
