package aep

import (
	"bytes"
	"encoding/binary"
	"strings"

	"github.com/example/aep-parser/internal/rifx"
)

// parseLayer reads layer data from a Layr list.
//
// ldta layout (160 bytes; 164 for AE >= 23) — offsets verified against
// real .aep files plus py-aep's LdtaChunk reference:
//
//	0x00 : uint32 BE  — own layer ID (referenced by child layers' ParentID)
//	0x04 : uint16 BE  — quality (0=Wireframe, 1=Draft, 2=Best)
//	0x08 : int32  BE  — stretch_dividend
//	0x0C : int32  BE  — start_time_dividend
//	0x10 : uint32 BE  — start_time_divisor
//	0x14 : int32  BE  — in_point_dividend (source-media)
//	0x18 : uint32 BE  — in_point_divisor
//	0x1C : int32  BE  — out_point_dividend
//	0x20 : uint32 BE  — out_point_divisor
//	0x25–0x27 : 3-byte LayerAttrBits (see flag layout below)
//	0x28 : uint32 BE  — source item ID
//	0x3D : uint8      — label color index
//	0x40..0x5F        — legacy 32-byte NUL-terminated layer name (Windows
//	                    code-page text; the real name is in the sibling
//	                    Utf8 chunk so we ignore this)
//	0x63 : uint8      — blending mode
//	0x67 : uint8 bool — preserve transparency
//	0x6B : uint8      — track matte type
//	0x6C : uint32 BE  — stretch_divisor (paired with @0x08)
//	0x84 : uint32 BE  — parent layer ID (0 = no parent)
//	0x88 : uint32 BE  — light kind (only on light layers): 0=Parallel,
//	                    1=Spot, 2=Point, 3=Ambient. RE'd via
//	                    test_data/re_wave2_ae24.aep (RE_LIGHTS comp).
//	                    On non-light layers this slot is 0.
//	0xA0 : uint32 BE  — track matte SOURCE layer ID (AE 23+ explicit; 0 if
//	                    no explicit matte source — AE <= 22 used implicit
//	                    "layer immediately above" convention). RE'd via
//	                    test_data/re_trackmatte_ae24.aep.
//
// LayerAttrBits (3 bytes, per py-aep):
//
//	byte 0 (0x25): bit6=sampling-quality (0=Bilinear, 1=Bicubic),
//	               bit4=auto-orient characters-toward-camera,
//	               bit2=frame-blending-mode (0=FrameMix, 1=PixelMotion),
//	               bit1=guide-layer
//	byte 1 (0x26): bit7=null-layer, bit5=auto-orient camera-or-POI,
//	               bit4=markers-locked, bit3=solo, bit2=3D, bit1=adjustment,
//	               bit0=auto-orient along-path
//	byte 2 (0x27): bit7=collapse-transform, bit6=shy, bit5=lock,
//	               bit4=frame-blend-enabled, bit3=motion-blur,
//	               bit2=effects-active, bit1=audio-enabled, bit0=video (visible)
//
// The 3 auto-orient bits are mutually exclusive in AE's UI; we collapse
// them into a single Layer.AutoOrient enum.
func parseLayer(layr *rifx.Chunk, index int, ctx *parseCtx) (*Layer, error) {
	layer := &Layer{
		Index:   index,
		Visible: true,
		Stretch: 1.0,
		back:    &layerBackrefs{layrList: layr},
	}

	if utf8 := layr.FindFirst(rifx.IDUtf8); utf8 != nil {
		layer.Name = utf8.Text()
		layer.back.nameChunk = utf8
	}
	if cmta := layr.FindFirst(rifx.IDCmta); cmta != nil {
		layer.Comment = decodeCmta(cmta.Data)
		layer.back.commentChunk = cmta
	}

	ldta := layr.FindFirst(rifx.IDLdta)
	if ldta == nil {
		return layer, nil
	}
	layer.back.ldta = ldta

	d := ldta.Data

	if len(d) >= 0x04 {
		layer.ID = binary.BigEndian.Uint32(d[0x00:0x04])
	}
	if len(d) >= 0x06 {
		layer.Quality = LayerQuality(binary.BigEndian.Uint16(d[0x04:0x06]))
	}

	// Time fields are stored as dividend/divisor int/uint pairs. The
	// divisor is typically the comp's tick rate so dividing gives seconds.
	readFrac := func(off int) (val float64, ok bool) {
		if len(d) < off+8 {
			return 0, false
		}
		dividend := int32(binary.BigEndian.Uint32(d[off : off+4]))
		divisor := binary.BigEndian.Uint32(d[off+4 : off+8])
		if divisor == 0 {
			return 0, false
		}
		return float64(dividend) / float64(divisor), true
	}
	if v, ok := readFrac(0x0C); ok {
		layer.StartTime = v
	}
	// In-point / out-point on the source media give us the visible duration
	// in source seconds. Duration = (out − in).
	inSec, inOK := readFrac(0x14)
	outSec, outOK := readFrac(0x1C)
	if inOK && outOK {
		layer.Duration = outSec - inSec
	}
	// Stretch = stretch_dividend (0x08) / stretch_divisor (0x6C).
	if len(d) >= 0x70 {
		stretchDividend := int32(binary.BigEndian.Uint32(d[0x08:0x0C]))
		stretchDivisor := binary.BigEndian.Uint32(d[0x6C:0x70])
		if stretchDivisor > 0 {
			layer.Stretch = float64(stretchDividend) / float64(stretchDivisor)
		}
	}

	if len(d) >= 0x28 {
		b0, b1, b2 := d[0x25], d[0x26], d[0x27]
		layer.SamplingBicubic = b0&0x40 != 0
		layer.FrameBlendPixelMotion = b0&0x04 != 0
		layer.IsGuide = b0&0x02 != 0
		layer.IsNull = b1&0x80 != 0
		layer.MarkersLocked = b1&0x10 != 0
		layer.Solo = b1&0x08 != 0
		layer.Is3D = b1&0x04 != 0
		layer.IsAdjust = b1&0x02 != 0
		layer.CollapseTransform = b2&0x80 != 0
		layer.Shy = b2&0x40 != 0
		layer.Locked = b2&0x20 != 0
		layer.FrameBlendEnabled = b2&0x10 != 0
		layer.MotionBlur = b2&0x08 != 0
		layer.EffectsEnabled = b2&0x04 != 0
		layer.AudioEnabled = b2&0x02 != 0
		layer.Visible = b2&0x01 != 0

		switch {
		case b0&0x10 != 0:
			layer.AutoOrient = AutoOrientCharactersTowardCamera
		case b1&0x20 != 0:
			layer.AutoOrient = AutoOrientCameraOrPointOfInterest
		case b1&0x01 != 0:
			layer.AutoOrient = AutoOrientAlongPath
		}
	}

	if len(d) >= 0x2C {
		layer.SourceID = binary.BigEndian.Uint32(d[0x28:0x2C])
	}
	if len(d) >= 0x3E {
		layer.Label = d[0x3D]
	}
	if len(d) >= 0x64 {
		layer.BlendingMode = BlendingMode(d[0x63])
	}
	if len(d) >= 0x68 {
		layer.PreserveTransparency = d[0x67] != 0
	}
	if len(d) >= 0x6C {
		layer.TrackMatte = TrackMatteType(d[0x6B])
	}
	if len(d) >= 0x88 {
		layer.ParentID = binary.BigEndian.Uint32(d[0x84:0x88])
	}
	if len(d) >= 0x8C {
		// Light kind is at @0x88-0x8B (uint32 BE; small enum 0..3). Only
		// meaningful for light layers; for non-light layers AE writes 0
		// which coincidentally maps to LightKindParallel — callers must
		// gate on Type == LayerTypeLight before reading.
		layer.LightKind = LightKind(binary.BigEndian.Uint32(d[0x88:0x8C]) & 0xFF)
	}
	if len(d) >= 0xA4 {
		layer.TrackMatteLayerID = binary.BigEndian.Uint32(d[0xA0:0xA4])
	}

	if blsi := findAlternateSourceBlsi(layr); blsi != nil && len(blsi.Data) >= 4 {
		layer.back.alternateSourceBlsi = blsi
		layer.AlternateSourceID = binary.BigEndian.Uint32(blsi.Data[0:4])
	}

	layer.Type = inferLayerType(d, layer)
	layer.Properties, layer.Effects, layer.Markers = parseProperties(layr, ctx)
	layer.Masks = parseMasks(layr, ctx)
	layer.propertyTree = buildAEPropertyGroupTree(layr)
	wirePropertyTreeLeaves(layer.propertyTree, layer.Properties)
	layer.back.btdsChunk = findTextSourceChunk(layr)
	if layer.back.btdsChunk != nil {
		layer.TextSourceRaw = layer.back.btdsChunk.Data
		layer.Type = LayerTypeText
		ts, warn := decodeTextSource(layer.TextSourceRaw)
		layer.TextSource = ts
		if warn != "" && ctx != nil {
			ctx.warn("%s", warn)
		}
	}
	layer.IsShapeLayer = hasShapeLayerRoot(layr)
	if layer.IsShapeLayer {
		layer.Type = LayerTypeShape
		layer.ShapePaths = collectShapePaths(layr)
		layer.ShapePrimitives = collectShapePrimitives(layr, ctx)
		layer.shapeRootGroup = hydrateShapeNodes(layr, ctx)
		hydrateLayerTransform(layer)
	}
	return layer, nil
}

// findAlternateSourceBlsi scans a layer's properties for the Essential
// Properties → Media Replacement chunk pattern and returns the blsi chunk
// (4-byte BE uint32 holding the alternate-source AVItem id), or nil if
// the layer has no media-replacement slot.
//
// Pattern under the layer's properties tdgp (one slot per layer in the
// fixtures we have; we return the first match):
//
//	tdmn "ADBE Layer Overrides"
//	LIST OvG2                                  (CprC count + LIST CPrp slot UUID)
//	LIST tdgp (override container)
//	  tdsb / tdsn
//	  tdmn "ADBE Layer Source Alternate"
//	  blsv                                     (4 bytes, value=1)
//	  blsi                                     (4 bytes BE = alt AVItem id; 0 = unset)
//	  LIST tdbs                                (slot tdb4/cdat plus tdsn "<slot name>")
//	  tdmn "ADBE Group End"
//
// We ignore the second `ADBE Layer Source Alternate` occurrence that
// AE writes inside the sibling "ADBE Source Options Group" tdgp — its
// blsi is always 0 in every fixture observed; treating it as the
// authoritative override would lose the user-set value.
func findAlternateSourceBlsi(layr *rifx.Chunk) *rifx.Chunk {
	var propsTdgp *rifx.Chunk
	for _, ch := range layr.Children {
		if ch.IsList() && ch.FormType == rifx.IDTdgp {
			propsTdgp = ch
			break
		}
	}
	if propsTdgp == nil {
		return nil
	}
	kids := propsTdgp.Children
	for i := 0; i+2 < len(kids); i++ {
		if kids[i].ID != rifx.IDTdmn || trimNUL(kids[i].Data) != "ADBE Layer Overrides" {
			continue
		}
		ovg2 := kids[i+1]
		override := kids[i+2]
		if !ovg2.IsList() || ovg2.FormType != rifx.IDOvG2 {
			continue
		}
		if !override.IsList() || override.FormType != rifx.IDTdgp {
			continue
		}
		return findBlsiInOverrideTdgp(override)
	}
	return nil
}

// findBlsiInOverrideTdgp walks an Essential Properties override tdgp
// looking for the tdmn "ADBE Layer Source Alternate" / blsv / blsi run
// described in findAlternateSourceBlsi's header. Returns the blsi chunk
// or nil if the pattern isn't present (e.g., a non-source EGP override
// like Position or Color promoted to Essential Properties).
func findBlsiInOverrideTdgp(tdgp *rifx.Chunk) *rifx.Chunk {
	kids := tdgp.Children
	for i := 0; i+2 < len(kids); i++ {
		if kids[i].ID != rifx.IDTdmn || trimNUL(kids[i].Data) != "ADBE Layer Source Alternate" {
			continue
		}
		if kids[i+1].ID != rifx.IDBlsv || kids[i+2].ID != rifx.IDBlsi {
			continue
		}
		return kids[i+2]
	}
	return nil
}

// inferLayerType picks a LayerType from the ldta byte at 0x83 (py-aep's
// LayerType enum: 0=AV, 1=Light, 2=Camera, 3=Text, 4=Shape, 5=3DModel)
// plus the parsed flag bits and SourceID. The 0x83 byte distinguishes
// AE's built-in 3D layer kinds (Light/Camera/3DModel); text/shape are
// also detected via the property tree (findTextSource / hasShapeLayerRoot
// in parseLayer), so this function focuses on the cases the property tree
// can't tell us about.
func inferLayerType(ldta []byte, l *Layer) LayerType {
	if len(ldta) > 0x83 {
		switch ldta[0x83] {
		case 0x01:
			return LayerTypeLight
		case 0x02:
			return LayerTypeCamera
		case 0x05:
			return LayerType3DModel
		}
	}
	if l.IsAdjust {
		return LayerTypeAdjust
	}
	if l.IsNull || l.SourceID == 0 {
		return LayerTypeNull
	}
	return LayerTypeAV
}

// decodeCmta decodes a cmta chunk's payload (AE's layer comment).
// AE writes UTF-8 with CRLF line endings and a double-NUL terminator; we
// cut at the first NUL and normalize CRLF → LF.
func decodeCmta(d []byte) string {
	if i := bytes.IndexByte(d, 0); i >= 0 {
		d = d[:i]
	}
	return strings.ReplaceAll(string(d), "\r\n", "\n")
}

// findTextSourceChunk scans the layer's property tree for an
// "ADBE Text Document" payload (a btds opaque LIST) and returns the
// LIST chunk itself, or nil if the layer has no text source. Callers
// read the bytes via the returned `chunk.Data`; keeping the chunk
// reference around (vs. just the byte slice) lets length-variable
// writes update `chunk.Data` so WriteAEP sees the new payload.
func findTextSourceChunk(layr *rifx.Chunk) *rifx.Chunk {
	var found *rifx.Chunk
	var walk func(c *rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if found != nil {
			return
		}
		for _, ch := range c.Children {
			if ch.IsList() && ch.FormType == rifx.IDBtds {
				found = ch
				return
			}
			if ch.IsList() {
				walk(ch)
			}
		}
	}
	walk(layr)
	return found
}

// hasShapeLayerRoot reports whether the layer has an "ADBE Root Vectors
// Group" tdmn anywhere in its property tree — the marker that AE uses for
// Shape Layers.
func hasShapeLayerRoot(layr *rifx.Chunk) bool {
	var found bool
	var visit func(c *rifx.Chunk)
	visit = func(c *rifx.Chunk) {
		if found {
			return
		}
		for _, ch := range c.Children {
			if ch.ID == rifx.IDTdmn && trimNUL(ch.Data) == "ADBE Root Vectors Group" {
				found = true
				return
			}
			if ch.IsList() {
				visit(ch)
			}
		}
	}
	visit(layr)
	return found
}

// LdtaRawBytes returns the layer's ldta chunk Data slice, or nil if the
// layer has no ldta. Read-only access for debugging / RE tools — the
// underlying byte slice is the live chunk data; do not mutate.
func (l *Layer) LdtaRawBytes() []byte {
	if l.back == nil || l.back.ldta == nil {
		return nil
	}
	return l.back.ldta.Data
}
