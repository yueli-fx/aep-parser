package serializer

import (
	"encoding/binary"
	"math"

	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

// parseMasks walks the layer's property tree for the "ADBE Mask Parade"
// group and decodes each mask atom into a Mask.
//
// Structure (observed):
//
//	tdmn "ADBE Mask Parade" → tdgp
//	  tdmn "ADBE Mask Atom" + mkif (48-byte mode/inverted/color) + tdgp
//	    tdmn "ADBE Mask Shape" → om-s LIST
//	      tdbs (descriptor)
//	      omks LIST → shap LIST → { shph(24), list LIST { lhd3, ldat }, omtn }
//
// Each Bezier vertex on the path is stored as 3 consecutive float32 X/Y
// pairs in ldat (anchor, in-tangent, out-tangent). Straight-line segments
// have all three pairs identical.
func parseMasks(layr *rifx.Chunk, ctx *parseCtx) []*Mask {
	parade := findMaskParade(layr)
	if parade == nil {
		return nil
	}
	var masks []*Mask
	kids := parade.Children
	for i := 0; i < len(kids); i++ {
		ch := kids[i]
		if ch.ID != rifx.IDTdmn {
			continue
		}
		name := trimNUL(ch.Data)
		if name != "ADBE Mask Atom" {
			continue
		}
		// Next siblings: mkif chunk + tdgp containing the mask's inner properties.
		var mkif *rifx.Chunk
		var atomTdgp *rifx.Chunk
		for j := i + 1; j < len(kids); j++ {
			sib := kids[j]
			if sib.ID == rifx.IDMkif && mkif == nil {
				mkif = sib
			}
			if sib.IsList() && sib.FormType == rifx.IDTdgp {
				atomTdgp = sib
				break // tdgp marks end of this atom; next atom starts after it
			}
			if sib.ID == rifx.IDTdmn {
				break // hit next named entry
			}
		}
		if atomTdgp == nil {
			continue
		}
		mask := decodeMask(atomTdgp, ctx)
		if mask == nil {
			continue
		}
		if mask.Name == "" {
			// AE persists the user-visible mask name in the atom tdgp's tdsn
			// and leaves omtn empty; synthetic fixtures put it in omtn.
			mask.Name = vectorGroupName(atomTdgp)
		}
		if mb := maskBack(mask); mb != nil {
			mb.atomTdgp = atomTdgp
		}
		if mkif != nil {
			mask.MkifRaw = append([]byte(nil), mkif.Data...)
			if mb := maskBack(mask); mb != nil {
				mb.mkif = mkif
				mb.maskName = mask.Name
			}
			decodeMkif(mask, mkif.Data)
		}
		mask.Opacity = 1.0 // default if no cdat
		collectMaskProperties(atomTdgp, mask, ctx)
		masks = append(masks, mask)
	}
	return masks
}

func findMaskParade(layr *rifx.Chunk) *rifx.Chunk {
	var result *rifx.Chunk
	var visit func(c *rifx.Chunk)
	visit = func(c *rifx.Chunk) {
		if result != nil {
			return
		}
		kids := c.Children
		for i := 0; i < len(kids); i++ {
			ch := kids[i]
			if ch.ID == rifx.IDTdmn && trimNUL(ch.Data) == "ADBE Mask Parade" {
				if i+1 < len(kids) && kids[i+1].IsList() && kids[i+1].FormType == rifx.IDTdgp {
					result = kids[i+1]
					return
				}
			}
			if ch.IsList() {
				visit(ch)
			}
		}
	}
	visit(layr)
	return result
}

// decodeMask reads one mask's tdgp (the one inside an "ADBE Mask Atom").
// It locates the om-s wrapper around "ADBE Mask Shape" and pulls vertices
// out of the inner shap LIST(s).
//
// For animated mask paths there are N shap LISTs inside omks and N
// timing entries in the sibling tdbs.kfl LIST (64-byte blocks; time at
// 0x00-0x03 as uint32 1/8000-second ticks).
func decodeMask(maskTdgp *rifx.Chunk, ctx *parseCtx) *Mask {
	var omS *rifx.Chunk
	kids := maskTdgp.Children
	for i := 0; i < len(kids); i++ {
		ch := kids[i]
		if ch.ID != rifx.IDTdmn {
			continue
		}
		if trimNUL(ch.Data) == "ADBE Mask Shape" && i+1 < len(kids) {
			next := kids[i+1]
			if next.IsList() && next.FormType == rifx.IDOmS {
				omS = next
				break
			}
		}
	}
	if omS == nil {
		return nil
	}
	// Gather om-s children: an optional tdbs (path-keyframe times) and an
	// omks LIST containing one or more shap LISTs.
	var tdbs *rifx.Chunk
	var omks *rifx.Chunk
	for _, ch := range omS.Children {
		if !ch.IsList() {
			continue
		}
		switch ch.FormType {
		case rifx.IDTdbs:
			tdbs = ch
		case rifx.IDOmks:
			omks = ch
		}
	}
	if omks == nil {
		return nil
	}
	var shaps []*rifx.Chunk
	for _, ch := range omks.Children {
		if ch.IsList() && ch.FormType == rifx.IDShap {
			shaps = append(shaps, ch)
		}
	}
	if len(shaps) == 0 {
		return nil
	}

	mask := &Mask{}
	scene.SetMaskBack(mask, &maskBackrefs{})
	// First snapshot drives Mask.Closed / Mask.Name / Mask.ShphRaw.
	fillFromShap(mask, shaps[0])

	if len(shaps) == 1 {
		// Static path — Vertices is the only snapshot, PathKeyframes stays nil.
		return mask
	}
	// Animated: pull times from tdbs.kfl, vertices from each shap.
	times := readMaskPathTimes(tdbs, ctx)
	mask.PathKeyframes = make([]MaskPathKeyframe, len(shaps))
	for i, s := range shaps {
		kf := MaskPathKeyframe{Vertices: vertsFromShap(s)}
		if i < len(times) {
			kf.Time = times[i].time
			kf.InInterp = times[i].inInterp
			kf.OutInterp = times[i].outInterp
			kf.InTemporalEase = times[i].inEase
			kf.OutTemporalEase = times[i].outEase
		}
		mask.PathKeyframes[i] = kf
	}
	mask.Vertices = append([]MaskVertex(nil), mask.PathKeyframes[0].Vertices...)
	return mask
}

// collectMaskProperties walks a mask atom's inner tdgp and populates the
// well-known properties (Feather/Opacity/Expansion) plus a generic
// Properties slice for anything else / keyframed properties.
//
// Mask static cdat property names (verified against real .aep files):
//
//	ADBE Mask Shape    → handled by decodeMask (om-s wrapper)
//	ADBE Mask Feather  → 2D cdat (Feather X, Feather Y in pixels)
//	ADBE Mask Opacity  → 1D cdat (0..1)
//	ADBE Mask Offset   → 1D cdat (Expansion in pixels; called "Offset"
//	                     internally, "Expansion" in the AE UI)
func collectMaskProperties(atomTdgp *rifx.Chunk, mask *Mask, ctx *parseCtx) {
	walkTdmnPairs(atomTdgp, func(name string, payload *rifx.Chunk) bool {
		if payload.FormType != rifx.IDTdbs {
			return true
		}
		switch name {
		case "ADBE Mask Shape":
			return true // shape handled in decodeMask
		case "ADBE Mask Feather":
			if cdat := payload.FindFirst(rifx.IDCdat); cdat != nil && len(cdat.Data) >= 16 {
				x, _ := readFloat64BE(cdat.Data, 0)
				y, _ := readFloat64BE(cdat.Data, 8)
				mask.Feather = [2]float64{x, y}
			}
		case "ADBE Mask Opacity":
			if cdat := payload.FindFirst(rifx.IDCdat); cdat != nil && len(cdat.Data) >= 8 {
				v, _ := readFloat64BE(cdat.Data, 0)
				mask.Opacity = v
			}
		case "ADBE Mask Offset":
			if cdat := payload.FindFirst(rifx.IDCdat); cdat != nil && len(cdat.Data) >= 8 {
				v, _ := readFloat64BE(cdat.Data, 0)
				mask.Expansion = v
			}
		}
		// Always also surface the property in the generic slice (so
		// keyframed values on Feather/Opacity/etc. are reachable).
		if p := parseLeafProperty(name, payload, ctx); p != nil {
			mask.Properties = append(mask.Properties, p)
		}
		return true
	})
}

// decodeMkif populates Mode, Inverted, Index, and Color on the mask from
// its 48-byte mkif payload. Layout (verified by diffing 3 masks with
// known mode/inverted/color differences):
//
//	0x00 : uint8  — Inverted flag (0/1)
//	0x01 : uint8  — Locked flag (0/1)
//	0x02 : uint8  — MaskMotionBlur mode (0=SameAsLayer, 2=On, 3=Off)
//	0x03 : uint8  — MaskFeatherFalloff (0=Smooth, 1=Linear) — RE'd 2026-06-17
//	0x04 : uint32 BE — Mode enum (MaskModeNone..MaskModeDifference)
//	0x08 : uint32 BE — Mask index (1-based)
//	0x2C : uint8  — alpha (always 0xFF)
//	0x2D : uint8  — Color R
//	0x2E : uint8  — Color G
//	0x2F : uint8  — Color B
func decodeMkif(m *Mask, data []byte) {
	if len(data) < 0x30 {
		return
	}
	m.Inverted = data[0x00] != 0
	m.Locked = data[0x01] != 0
	m.MotionBlur = MaskMotionBlurMode(data[0x02])
	m.FeatherFalloff = MaskFeatherFalloff(data[0x03])
	m.Mode = MaskMode(binary.BigEndian.Uint32(data[0x04:0x08]))
	m.Index = binary.BigEndian.Uint32(data[0x08:0x0C])
	m.Color = [3]uint8{data[0x2D], data[0x2E], data[0x2F]}
}

// fillFromShap sets Closed / Name / ShphRaw / Vertices from a single shap.
func fillFromShap(m *Mask, shap *rifx.Chunk) {
	for _, ch := range shap.Children {
		switch {
		case ch.ID == rifx.IDShph:
			m.ShphRaw = append([]byte(nil), ch.Data...)
			if mb := maskBack(m); mb != nil && mb.shph == nil {
				mb.shph = ch
			}
			// Open flag = shph[3] bit3 (AE-2020-saved ground truth,
			// re_mask_open.aep: closed = 02 01, open = 02 09; [0x14] is 0x01
			// on every mask regardless — the earlier @0x14 reading only ever
			// matched closed masks by coincidence).
			if len(ch.Data) >= 4 {
				m.Closed = ch.Data[3]&0x08 == 0
			}
		case ch.IsList() && ch.FormType == rifx.IDkfl:
			m.Vertices = decodeMaskVertices(ch)
		case ch.ID == rifx.IDOmtn:
			m.Name = trimNUL(ch.Data)
		}
	}
}

func vertsFromShap(shap *rifx.Chunk) []MaskVertex {
	for _, ch := range shap.Children {
		if ch.IsList() && ch.FormType == rifx.IDkfl {
			return decodeMaskVertices(ch)
		}
	}
	return nil
}

// maskPathTime is one entry from the om-s.tdbs.kfl time stream.
type maskPathTime struct {
	time      float64
	inInterp  InterpType
	outInterp InterpType
	inEase    TemporalEase
	outEase   TemporalEase
}

// readMaskPathTimes pulls per-keyframe timing/interp/temporal ease from the
// om-s tdbs's kfl LIST. Each block is 64 bytes; layout matches the spatial
// keyframe scalar-ease pattern: time @0x00, in interp @0x04, out interp
// @0x05, in speed @0x18, in influence @0x20, out speed @0x28, out
// influence @0x30.
func readMaskPathTimes(tdbs *rifx.Chunk, ctx *parseCtx) []maskPathTime {
	if tdbs == nil {
		return nil
	}
	var kfl *rifx.Chunk
	for _, ch := range tdbs.Children {
		if ch.IsList() && ch.FormType == rifx.IDkfl {
			kfl = ch
			break
		}
	}
	if kfl == nil {
		return nil
	}
	lhd3 := kfl.FindFirst(rifx.IDLhd3)
	ldat := kfl.FindFirst(rifx.IDLdat)
	if lhd3 == nil || ldat == nil || len(lhd3.Data) < 0x14 {
		return nil
	}
	count := int(binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]))
	bpk := int(binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]))
	if count <= 0 || bpk <= 0 || count*bpk > len(ldat.Data) {
		return nil
	}
	out := make([]maskPathTime, count)
	for i := 0; i < count; i++ {
		off := i * bpk
		t := float64(binary.BigEndian.Uint32(ldat.Data[off:off+4])) / ctx.tickRate
		mpt := maskPathTime{
			time:      t,
			inInterp:  InterpType(ldat.Data[off+0x04]),
			outInterp: InterpType(ldat.Data[off+0x05]),
		}
		if bpk >= 0x38 {
			inSpd, _ := readFloat64BE(ldat.Data, off+0x18)
			inInf, _ := readFloat64BE(ldat.Data, off+0x20)
			outSpd, _ := readFloat64BE(ldat.Data, off+0x28)
			outInf, _ := readFloat64BE(ldat.Data, off+0x30)
			mpt.inEase = TemporalEase{Speed: inSpd, Influence: inInf}
			mpt.outEase = TemporalEase{Speed: outSpd, Influence: outInf}
		}
		out[i] = mpt
	}
	return out
}

// decodeMaskVertices reads the kfl LIST inside a shap (lhd3+ldat) and
// returns vertices. Each vertex consumes 3 float32 X/Y pairs (anchor,
// in-tangent, out-tangent).
func decodeMaskVertices(kfl *rifx.Chunk) []MaskVertex {
	lhd3 := kfl.FindFirst(rifx.IDLhd3)
	ldat := kfl.FindFirst(rifx.IDLdat)
	if lhd3 == nil || ldat == nil || len(lhd3.Data) < 0x14 {
		return nil
	}
	count := int(binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]))
	bpk := int(binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]))
	if count <= 0 || bpk != 8 || count*bpk > len(ldat.Data) {
		return nil
	}
	// 3 entries per vertex (anchor, in-tangent, out-tangent).
	if count%3 != 0 {
		// Unexpected; bail rather than mis-decode.
		return nil
	}
	nVerts := count / 3
	verts := make([]MaskVertex, 0, nVerts)
	for v := 0; v < nVerts; v++ {
		baseOff := v * 3 * bpk
		anchor := readFloat32XY(ldat.Data, baseOff)
		in := readFloat32XY(ldat.Data, baseOff+bpk)
		out := readFloat32XY(ldat.Data, baseOff+2*bpk)
		verts = append(verts, MaskVertex{Anchor: anchor, InTangent: in, OutTangent: out})
	}
	return verts
}

// readFloat32XY reads two big-endian float32 values starting at offset and
// returns them as a [2]float64 pair.
func readFloat32XY(b []byte, offset int) [2]float64 {
	if offset+8 > len(b) {
		return [2]float64{}
	}
	x := math.Float32frombits(binary.BigEndian.Uint32(b[offset:]))
	y := math.Float32frombits(binary.BigEndian.Uint32(b[offset+4:]))
	return [2]float64{float64(x), float64(y)}
}
