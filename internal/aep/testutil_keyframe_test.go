package aep_test

import (
	"encoding/binary"
	"math"

	aep "github.com/example/aep-parser/internal/aep"
)

// Keyframe block builders + rifxBuilder leaf helpers used across tests.
// The "kf block" funcs return raw bytes for a single keyframe inside an
// ldat stream; the "leaf*" methods compose them into full tdmn+tdbs
// property leaves.

// ───── lhd3 / tdb4 headers ─────

func buildLhd3(count, bpk uint32) []byte {
	blk := make([]byte, 0x14)
	binary.BigEndian.PutUint32(blk[0x08:], count)
	binary.BigEndian.PutUint32(blk[0x10:], bpk)
	return blk
}

func buildTdb4(dim byte) []byte {
	// decodeTdb4Components only inspects byte[3]: 0x01=1D, 0x03/0x07=3D.
	return []byte{0, 0, 0, dim}
}

// ───── keyframe block builders ─────

func buildKF1D(timeSec, value float64) []byte {
	blk := make([]byte, 48)
	binary.BigEndian.PutUint32(blk[0:], uint32(math.Round(timeSec*8000)))
	// blk[0x07] left as 0 → kfValueOffset returns 0x08
	binary.BigEndian.PutUint64(blk[0x08:], math.Float64bits(value))
	return blk
}

func buildKF3DSpatial(timeSec, x, y, z float64) []byte {
	blk := make([]byte, 128)
	binary.BigEndian.PutUint32(blk[0:], uint32(math.Round(timeSec*8000)))
	blk[0x07] = 0x07 // spatial marker → kfValueOffset returns 0x38
	binary.BigEndian.PutUint64(blk[0x38:], math.Float64bits(x))
	binary.BigEndian.PutUint64(blk[0x40:], math.Float64bits(y))
	binary.BigEndian.PutUint64(blk[0x48:], math.Float64bits(z))
	return blk
}

// buildKF3DSpatialBezier writes a 128-byte spatial keyframe block with
// bezier in/out interpolation, value (X, Y, Z), and explicit spatial
// tangents (each a 3-vector).
func buildKF3DSpatialBezier(timeSec float64, val, inTan, outTan [3]float64, inInf, outInf float64) []byte {
	blk := make([]byte, 128)
	binary.BigEndian.PutUint32(blk[0:], uint32(math.Round(timeSec*8000)))
	blk[0x04] = byte(aep.InterpBezier)
	blk[0x05] = byte(aep.InterpBezier)
	blk[0x07] = 0x07
	binary.BigEndian.PutUint64(blk[0x20:], math.Float64bits(inInf))
	binary.BigEndian.PutUint64(blk[0x30:], math.Float64bits(outInf))
	for i := 0; i < 3; i++ {
		binary.BigEndian.PutUint64(blk[0x38+i*8:], math.Float64bits(val[i]))
		binary.BigEndian.PutUint64(blk[0x50+i*8:], math.Float64bits(inTan[i]))
		binary.BigEndian.PutUint64(blk[0x68+i*8:], math.Float64bits(outTan[i]))
	}
	return blk
}

// buildKF1DBezier builds a 1D non-spatial bezier-eased keyframe block
// (bpk=48) with the given speed/influence pairs per side.
func buildKF1DBezier(timeSec, value, inSpeed, inInf, outSpeed, outInf float64) []byte {
	blk := make([]byte, 48)
	binary.BigEndian.PutUint32(blk[0:], uint32(math.Round(timeSec*8000)))
	blk[0x04] = byte(aep.InterpBezier)
	blk[0x05] = byte(aep.InterpBezier)
	binary.BigEndian.PutUint64(blk[0x08:], math.Float64bits(value))
	binary.BigEndian.PutUint64(blk[0x10:], math.Float64bits(inSpeed))
	binary.BigEndian.PutUint64(blk[0x18:], math.Float64bits(inInf))
	binary.BigEndian.PutUint64(blk[0x20:], math.Float64bits(outSpeed))
	binary.BigEndian.PutUint64(blk[0x28:], math.Float64bits(outInf))
	return blk
}

// buildKF3DBezierNonSpatial builds a 3D non-spatial bezier-eased keyframe
// block (bpk=128) with per-component speed/influence values.
func buildKF3DBezierNonSpatial(timeSec float64, value, inSpd, inInf, outSpd, outInf [3]float64) []byte {
	blk := make([]byte, 128)
	binary.BigEndian.PutUint32(blk[0:], uint32(math.Round(timeSec*8000)))
	blk[0x04] = byte(aep.InterpBezier)
	blk[0x05] = byte(aep.InterpBezier)
	// blk[0x07] = 0x00 (non-spatial)
	for i := 0; i < 3; i++ {
		binary.BigEndian.PutUint64(blk[0x08+i*8:], math.Float64bits(value[i]))
		binary.BigEndian.PutUint64(blk[0x20+i*8:], math.Float64bits(inSpd[i]))
		binary.BigEndian.PutUint64(blk[0x38+i*8:], math.Float64bits(inInf[i]))
		binary.BigEndian.PutUint64(blk[0x50+i*8:], math.Float64bits(outSpd[i]))
		binary.BigEndian.PutUint64(blk[0x68+i*8:], math.Float64bits(outInf[i]))
	}
	return blk
}

// buildKF4DColor builds a 4D color keyframe block (bpk=152, header[0x07]=0x01)
// with Easy Ease defaults on both sides and the supplied RGBA values
// (0..255 range, stored as float64).
func buildKF4DColor(timeSec float64, r, g, b, a float64) []byte {
	blk := make([]byte, 152)
	binary.BigEndian.PutUint32(blk[0:], uint32(math.Round(timeSec*8000)))
	blk[0x04] = byte(aep.InterpBezier)
	blk[0x05] = byte(aep.InterpBezier)
	blk[0x07] = 0x01 // 4D / color marker
	// Easy Ease default temporal ease
	binary.BigEndian.PutUint64(blk[0x20:], math.Float64bits(0.333))
	binary.BigEndian.PutUint64(blk[0x30:], math.Float64bits(0.333))
	// Values at 0x38, 0x40, 0x48, 0x50
	binary.BigEndian.PutUint64(blk[0x38:], math.Float64bits(r))
	binary.BigEndian.PutUint64(blk[0x40:], math.Float64bits(g))
	binary.BigEndian.PutUint64(blk[0x48:], math.Float64bits(b))
	binary.BigEndian.PutUint64(blk[0x50:], math.Float64bits(a))
	return blk
}

// buildKF2DNonSpatial builds a 2D non-spatial keyframe block (bpk=88).
// Layout: 0x08+i*8 value, 0x18+i*8 in-speed, 0x28+i*8 in-influence,
// 0x38+i*8 out-speed, 0x48+i*8 out-influence (i=0..1).
func buildKF2DNonSpatial(timeSec, x, y, inSpdX, inSpdY, inInfX, inInfY, outSpdX, outSpdY, outInfX, outInfY float64) []byte {
	blk := make([]byte, 88)
	binary.BigEndian.PutUint32(blk[0:], uint32(math.Round(timeSec*8000)))
	blk[0x04] = byte(aep.InterpBezier)
	blk[0x05] = byte(aep.InterpBezier)
	binary.BigEndian.PutUint64(blk[0x08:], math.Float64bits(x))
	binary.BigEndian.PutUint64(blk[0x10:], math.Float64bits(y))
	binary.BigEndian.PutUint64(blk[0x18:], math.Float64bits(inSpdX))
	binary.BigEndian.PutUint64(blk[0x20:], math.Float64bits(inSpdY))
	binary.BigEndian.PutUint64(blk[0x28:], math.Float64bits(inInfX))
	binary.BigEndian.PutUint64(blk[0x30:], math.Float64bits(inInfY))
	binary.BigEndian.PutUint64(blk[0x38:], math.Float64bits(outSpdX))
	binary.BigEndian.PutUint64(blk[0x40:], math.Float64bits(outSpdY))
	binary.BigEndian.PutUint64(blk[0x48:], math.Float64bits(outInfX))
	binary.BigEndian.PutUint64(blk[0x50:], math.Float64bits(outInfY))
	return blk
}

// ───── rifxBuilder leaf-property helpers ─────

// leafKeyframed builds the chunk bytes for one keyframed property:
// tdmn(matchName) + tdbs LIST { tdb4 + list LIST { lhd3 + ldat } }
func (b *rifxBuilder) leafKeyframed(matchName string, dim byte, bpk uint32, kfBlocks [][]byte) []byte {
	tdmn := b.chunk("tdmn", []byte(matchName))
	tdb4 := b.chunk("tdb4", buildTdb4(dim))
	var ldatData []byte
	for _, kf := range kfBlocks {
		ldatData = append(ldatData, kf...)
	}
	lhd3 := b.chunk("lhd3", buildLhd3(uint32(len(kfBlocks)), bpk))
	ldat := b.chunk("ldat", ldatData)
	var kflInner []byte
	kflInner = append(kflInner, lhd3...)
	kflInner = append(kflInner, ldat...)
	// IDkfl is the lowercase "list" formType.
	kflList := b.listChunk("LIST", "list", kflInner)
	var tdbsInner []byte
	tdbsInner = append(tdbsInner, tdb4...)
	tdbsInner = append(tdbsInner, kflList...)
	tdbsList := b.listChunk("LIST", "tdbs", tdbsInner)
	var out []byte
	out = append(out, tdmn...)
	out = append(out, tdbsList...)
	return out
}

// leafStatic builds the chunk bytes for one cdat-only static property:
// tdmn(matchName) + tdbs LIST { tdb4 + cdat }
func (b *rifxBuilder) leafStatic(matchName string, dim byte, value float64) []byte {
	tdmn := b.chunk("tdmn", []byte(matchName))
	tdb4 := b.chunk("tdb4", buildTdb4(dim))
	cdatData := make([]byte, 40) // 1D static cdat is 40-byte padded
	binary.BigEndian.PutUint64(cdatData[0:], math.Float64bits(value))
	cdat := b.chunk("cdat", cdatData)
	var tdbsInner []byte
	tdbsInner = append(tdbsInner, tdb4...)
	tdbsInner = append(tdbsInner, cdat...)
	tdbsList := b.listChunk("LIST", "tdbs", tdbsInner)
	var out []byte
	out = append(out, tdmn...)
	out = append(out, tdbsList...)
	return out
}

// leafStatic2D builds a static 2D property leaf (e.g. ADBE Mask Feather).
func (b *rifxBuilder) leafStatic2D(matchName string, x, y float64) []byte {
	tdmn := b.chunk("tdmn", []byte(matchName))
	tdb4 := b.chunk("tdb4", buildTdb4(0x03))
	cdatData := make([]byte, 80)
	binary.BigEndian.PutUint64(cdatData[0:], math.Float64bits(x))
	binary.BigEndian.PutUint64(cdatData[8:], math.Float64bits(y))
	cdat := b.chunk("cdat", cdatData)
	var tdbsInner []byte
	tdbsInner = append(tdbsInner, tdb4...)
	tdbsInner = append(tdbsInner, cdat...)
	tdbsList := b.listChunk("LIST", "tdbs", tdbsInner)
	return append(tdmn, tdbsList...)
}

// leafKeyframedWithExpr is like leafKeyframed but also includes a Utf8 chunk
// inside the tdbs holding the JS expression source.
func (b *rifxBuilder) leafKeyframedWithExpr(matchName string, dim byte, bpk uint32, kfBlocks [][]byte, expr string) []byte {
	tdmn := b.chunk("tdmn", []byte(matchName))
	tdb4 := b.chunk("tdb4", buildTdb4(dim))
	utf8 := b.chunk("Utf8", []byte(expr))
	var ldatData []byte
	for _, kf := range kfBlocks {
		ldatData = append(ldatData, kf...)
	}
	lhd3 := b.chunk("lhd3", buildLhd3(uint32(len(kfBlocks)), bpk))
	ldat := b.chunk("ldat", ldatData)
	var kflInner []byte
	kflInner = append(kflInner, lhd3...)
	kflInner = append(kflInner, ldat...)
	kflList := b.listChunk("LIST", "list", kflInner)
	var tdbsInner []byte
	tdbsInner = append(tdbsInner, tdb4...)
	tdbsInner = append(tdbsInner, utf8...)
	tdbsInner = append(tdbsInner, kflList...)
	tdbsList := b.listChunk("LIST", "tdbs", tdbsInner)
	var out []byte
	out = append(out, tdmn...)
	out = append(out, tdbsList...)
	return out
}

// buildCorruptKeyframedLeaf is leafKeyframed with a caller-controlled lhd3
// header. Tests of the warning paths use this to inject inconsistent
// count/bpk pairs that the normal helper would never produce.
func (b *rifxBuilder) buildCorruptKeyframedLeaf(matchName string, dim byte, ldatPayload []byte, lhd3Bytes []byte) []byte {
	tdmn := b.chunk("tdmn", []byte(matchName))
	tdb4 := b.chunk("tdb4", buildTdb4(dim))
	lhd3 := b.chunk("lhd3", lhd3Bytes)
	ldat := b.chunk("ldat", ldatPayload)
	var kflInner []byte
	kflInner = append(kflInner, lhd3...)
	kflInner = append(kflInner, ldat...)
	kflList := b.listChunk("LIST", "list", kflInner)
	var tdbsInner []byte
	tdbsInner = append(tdbsInner, tdb4...)
	tdbsInner = append(tdbsInner, kflList...)
	tdbsList := b.listChunk("LIST", "tdbs", tdbsInner)
	var out []byte
	out = append(out, tdmn...)
	out = append(out, tdbsList...)
	return out
}
