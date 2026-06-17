package serializer

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/example/aep-parser/internal/codec"
	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

// layerBackrefs holds the rifx.Chunk references that power Layer's
// length-preserving write paths (SetName / SetComment / SetVisible /
// SetBlendingMode / SetAlternateSource / per-text-run setters).
//
// Lifecycle:
//   - Populated by parseLayer when a Layer is built from a parsed .aep file.
//   - Nil for layers built outside the parser (NewProject / NewShapeLayer
//     builders construct it explicitly per their archetype).
//   - opaque is reserved for future V3 phases that need to round-trip
//     unrecognized sibling chunks under the layer's owning Layr LIST
//     (per CLAUDE.md hard constraint #5); currently nil.
type layerBackrefs struct {
	// ldta is the underlying ldta chunk reference, captured by parseLayer.
	// Used by SetVisible/SetBlendingMode/etc. for length-preserving
	// flag-bit and byte-field writes.
	ldta *rifx.Chunk

	// nameChunk is the layer's name Utf8 chunk. Used by SetName for
	// length-variable text replacement.
	nameChunk *rifx.Chunk

	// commentChunk is the layer's cmta chunk (may be nil when no
	// comment was set). SetComment replaces its data or creates one.
	commentChunk *rifx.Chunk

	// layrList is the owning Layr LIST itself — needed when SetComment
	// has to insert a fresh cmta chunk (no existing one to mutate).
	layrList *rifx.Chunk

	// btdsChunk is the btds LIST holding the text source bytes
	// (TextSourceRaw is an alias of this chunk's Data). Length-variable
	// text writes (per-run setters) update this chunk's Data to point
	// at a fresh splice; WriteAEP recomputes parent LIST sizes.
	btdsChunk *rifx.Chunk

	// alternateSourceBlsi is the underlying blsi chunk (4-byte BE uint32
	// holding the alt source AVItem id). nil for layers without an
	// Essential Properties media-replacement slot. SetAlternateSource
	// rewrites its first 4 data bytes in place (length-preserving).
	alternateSourceBlsi *rifx.Chunk

	opaque map[rifx.ChunkID]*rifx.Chunk
}

var _ LayerWriter = (*layerBackrefs)(nil)

// layerBack returns the concrete backrefs behind a Layer's writer interface
// for serializer-stage (parse_/mutate_/write_) raw chunk access. Returns nil
// when the layer was built outside the parser. Free function (the receiver is
// a scene type, so the accessor can't be a method on it post package-split).
func layerBack(l *Layer) *layerBackrefs {
	if lb, ok := scene.LayerBack(l).(*layerBackrefs); ok {
		return lb
	}
	return nil
}

func (b *layerBackrefs) LdtaFrac(off int) (float64, bool) {
	if b == nil || b.ldta == nil || len(b.ldta.Data) < off+8 {
		return 0, false
	}
	d := b.ldta.Data
	dividend := int32(binary.BigEndian.Uint32(d[off : off+4]))
	divisor := binary.BigEndian.Uint32(d[off+4 : off+8])
	if divisor == 0 {
		return 0, false
	}
	return float64(dividend) / float64(divisor), true
}

func (b *layerBackrefs) HasAlternateSourceSlot() bool {
	return b != nil && b.alternateSourceBlsi != nil
}

// StretchFrac decodes the split time-stretch dividend (@0x08) / divisor
// (@0x6C) pair. Unlike LdtaFrac the two halves are not contiguous, so it has
// its own accessor.
func (b *layerBackrefs) StretchFrac() (float64, bool) {
	if b == nil || b.ldta == nil || len(b.ldta.Data) < 0x70 {
		return 0, false
	}
	dividend := int32(binary.BigEndian.Uint32(b.ldta.Data[0x08:0x0C]))
	divisor := binary.BigEndian.Uint32(b.ldta.Data[0x6C:0x70])
	if divisor == 0 {
		return 0, false
	}
	return float64(dividend) / float64(divisor), true
}

// BtdsData exposes the live btds chunk bytes (text-source raw) so the scene
// text resync re-decodes after a back-side splice. Returns nil for non-text
// layers / layers built outside the parser.
func (b *layerBackrefs) BtdsData() []byte {
	if b == nil || b.btdsChunk == nil {
		return nil
	}
	return b.btdsChunk.Data
}

// LdtaRaw exposes the live ldta chunk bytes. Returns nil for layers built
// outside the parser / without an ldta.
func (b *layerBackrefs) LdtaRaw() []byte {
	if b == nil || b.ldta == nil {
		return nil
	}
	return b.ldta.Data
}

// flag-bit positions inside ldta @0x25-0x27 — must mirror parse_layer.go's
// decode. Keeping them centralized so future ldta-version surprises only
// need updating in one place. Serializer-side (the byte-patch impl); the
// scene Set* methods delegate here via LayerWriter.
type ldtaFlagBit struct {
	off  int  // byte offset (0x25, 0x26, or 0x27)
	mask byte // single-bit mask within that byte
}

var (
	flagSamplingBicubic       = ldtaFlagBit{0x25, 0x40}
	flagFrameBlendPixelMotion = ldtaFlagBit{0x25, 0x04}
	flagIsGuide               = ldtaFlagBit{0x25, 0x02}
	flagIsNull                = ldtaFlagBit{0x26, 0x80}
	flagMarkersLocked         = ldtaFlagBit{0x26, 0x10}
	flagSolo                  = ldtaFlagBit{0x26, 0x08}
	flagIs3D                  = ldtaFlagBit{0x26, 0x04}
	flagIsAdjust              = ldtaFlagBit{0x26, 0x02}
	flagCollapseTransform     = ldtaFlagBit{0x27, 0x80}
	flagShy                   = ldtaFlagBit{0x27, 0x40}
	flagLocked                = ldtaFlagBit{0x27, 0x20}
	flagFrameBlendEnabled     = ldtaFlagBit{0x27, 0x10}
	flagMotionBlur            = ldtaFlagBit{0x27, 0x08}
	flagEffectsEnabled        = ldtaFlagBit{0x27, 0x04}
	flagAudioEnabled          = ldtaFlagBit{0x27, 0x02}
	flagVisible               = ldtaFlagBit{0x27, 0x01}
)

// setFlagBit flips a single ldta flag bit to match v. Returns an error
// when the ldta chunk is missing or too short for the targeted byte
// (length-preserving so chunk size never changes).
func (b *layerBackrefs) setFlagBit(bit ldtaFlagBit, v bool) error {
	if b.ldta == nil {
		return fmt.Errorf("layer: no ldta chunk (built outside parser?)")
	}
	if len(b.ldta.Data) <= bit.off {
		return fmt.Errorf("layer: ldta @%#x out of range (len=%d)", bit.off, len(b.ldta.Data))
	}
	if v {
		b.ldta.Data[bit.off] |= bit.mask
	} else {
		b.ldta.Data[bit.off] &^= bit.mask
	}
	return nil
}

func (b *layerBackrefs) SetVisible(v bool) error { return b.setFlagBit(flagVisible, v) }

func (b *layerBackrefs) SetSolo(v bool) error { return b.setFlagBit(flagSolo, v) }

func (b *layerBackrefs) SetShy(v bool) error { return b.setFlagBit(flagShy, v) }

func (b *layerBackrefs) SetLocked(v bool) error { return b.setFlagBit(flagLocked, v) }

func (b *layerBackrefs) SetEffectsEnabled(v bool) error { return b.setFlagBit(flagEffectsEnabled, v) }

func (b *layerBackrefs) SetMotionBlur(v bool) error { return b.setFlagBit(flagMotionBlur, v) }

func (b *layerBackrefs) SetAudioEnabled(v bool) error { return b.setFlagBit(flagAudioEnabled, v) }

func (b *layerBackrefs) SetFrameBlendEnabled(v bool) error {
	return b.setFlagBit(flagFrameBlendEnabled, v)
}

func (b *layerBackrefs) SetCollapseTransform(v bool) error {
	return b.setFlagBit(flagCollapseTransform, v)
}

func (b *layerBackrefs) SetIs3D(v bool) error { return b.setFlagBit(flagIs3D, v) }

func (b *layerBackrefs) SetIsAdjust(v bool) error { return b.setFlagBit(flagIsAdjust, v) }

func (b *layerBackrefs) SetIsGuide(v bool) error { return b.setFlagBit(flagIsGuide, v) }

func (b *layerBackrefs) SetIsNull(v bool) error { return b.setFlagBit(flagIsNull, v) }

func (b *layerBackrefs) SetMarkersLocked(v bool) error { return b.setFlagBit(flagMarkersLocked, v) }

func (b *layerBackrefs) SetSamplingBicubic(v bool) error { return b.setFlagBit(flagSamplingBicubic, v) }

func (b *layerBackrefs) SetFrameBlendPixelMotion(v bool) error {
	return b.setFlagBit(flagFrameBlendPixelMotion, v)
}

func (b *layerBackrefs) SetBlendingMode(m BlendingMode) error {
	if b.ldta == nil {
		return fmt.Errorf("layer: no ldta chunk")
	}
	if len(b.ldta.Data) <= 0x63 {
		return fmt.Errorf("layer: ldta too short for BlendingMode write (len=%d)", len(b.ldta.Data))
	}
	b.ldta.Data[0x63] = byte(m)
	return nil
}

func (b *layerBackrefs) SetTrackMatte(t TrackMatteType) error {
	if b.ldta == nil {
		return fmt.Errorf("layer: no ldta chunk")
	}
	if len(b.ldta.Data) <= 0x6B {
		return fmt.Errorf("layer: ldta too short for TrackMatte write (len=%d)", len(b.ldta.Data))
	}
	b.ldta.Data[0x6B] = byte(t)
	return nil
}

func (b *layerBackrefs) SetLabel(index uint8) error {
	if b.ldta == nil {
		return fmt.Errorf("layer: no ldta chunk")
	}
	if len(b.ldta.Data) <= 0x3D {
		return fmt.Errorf("layer: ldta too short for Label write (len=%d)", len(b.ldta.Data))
	}
	b.ldta.Data[0x3D] = index
	return nil
}

func (b *layerBackrefs) SetQuality(q LayerQuality) error {
	if b.ldta == nil {
		return fmt.Errorf("layer: no ldta chunk")
	}
	if len(b.ldta.Data) < 0x06 {
		return fmt.Errorf("layer: ldta too short for Quality write (len=%d)", len(b.ldta.Data))
	}
	binary.BigEndian.PutUint16(b.ldta.Data[0x04:0x06], uint16(q))
	return nil
}

func (b *layerBackrefs) SetParent(parentID uint32) error {
	if b.ldta == nil {
		return fmt.Errorf("layer: no ldta chunk")
	}
	if len(b.ldta.Data) < 0x88 {
		return fmt.Errorf("layer: ldta too short for ParentID write (len=%d)", len(b.ldta.Data))
	}
	binary.BigEndian.PutUint32(b.ldta.Data[0x84:0x88], parentID)
	return nil
}

func (b *layerBackrefs) SetSource(sourceID uint32) error {
	if b.ldta == nil {
		return fmt.Errorf("layer: no ldta chunk")
	}
	if len(b.ldta.Data) < 0x2C {
		return fmt.Errorf("layer: ldta too short for SourceID write (len=%d)", len(b.ldta.Data))
	}
	binary.BigEndian.PutUint32(b.ldta.Data[0x28:0x2C], sourceID)
	return nil
}

func (b *layerBackrefs) SetAutoOrient(t AutoOrientType) error {
	if b.ldta == nil {
		return fmt.Errorf("layer: no ldta chunk")
	}
	if len(b.ldta.Data) <= 0x26 {
		return fmt.Errorf("layer: ldta too short for AutoOrient write (len=%d)", len(b.ldta.Data))
	}
	// Clear all three bits first.
	b.ldta.Data[0x25] &^= 0x10
	b.ldta.Data[0x26] &^= 0x20 | 0x01
	switch t {
	case AutoOrientCharactersTowardCamera:
		b.ldta.Data[0x25] |= 0x10
	case AutoOrientCameraOrPointOfInterest:
		b.ldta.Data[0x26] |= 0x20
	case AutoOrientAlongPath:
		b.ldta.Data[0x26] |= 0x01
	case AutoOrientNone:
		// already cleared
	default:
		return fmt.Errorf("layer: unknown AutoOrientType %d", int(t))
	}
	return nil
}

func (b *layerBackrefs) SetPreserveTransparency(v bool) error {
	if b.ldta == nil {
		return fmt.Errorf("layer: no ldta chunk")
	}
	if len(b.ldta.Data) <= 0x67 {
		return fmt.Errorf("layer: ldta too short for PreserveTransparency write (len=%d)", len(b.ldta.Data))
	}
	if v {
		b.ldta.Data[0x67] = 1
	} else {
		b.ldta.Data[0x67] = 0
	}
	return nil
}

// setLdtaFrac writes (dividend, divisor) at ldta @off / @off+4 such
// that dividend/divisor ≈ seconds. Reuses the existing divisor when
// non-zero; otherwise picks 600 (AE's standard for layer time fields).
// Mutates 8 bytes. Returns the resulting (dividend, divisor) so the
// caller can mirror to its Go field.
func (b *layerBackrefs) setLdtaFrac(off int, seconds float64, label string) (int32, uint32, error) {
	if b.ldta == nil {
		return 0, 0, fmt.Errorf("layer: no ldta chunk")
	}
	if len(b.ldta.Data) < off+8 {
		return 0, 0, fmt.Errorf("layer: ldta too short for %s write (len=%d, off=%#x)", label, len(b.ldta.Data), off)
	}
	divisor := binary.BigEndian.Uint32(b.ldta.Data[off+4 : off+8])
	if divisor == 0 {
		divisor = 600
	}
	dividend := int32(math.Round(seconds * float64(divisor)))
	binary.BigEndian.PutUint32(b.ldta.Data[off:off+4], uint32(dividend))
	binary.BigEndian.PutUint32(b.ldta.Data[off+4:off+8], divisor)
	return dividend, divisor, nil
}

func (b *layerBackrefs) SetStartTime(seconds float64) error {
	_, _, err := b.setLdtaFrac(0x0C, seconds, "StartTime")
	return err
}

func (b *layerBackrefs) SetInPoint(seconds float64) error {
	_, _, err := b.setLdtaFrac(0x14, seconds, "InPoint")
	return err
}

func (b *layerBackrefs) SetOutPoint(seconds float64) error {
	_, _, err := b.setLdtaFrac(0x1C, seconds, "OutPoint")
	return err
}

func (b *layerBackrefs) SetStretch(ratio float64) error {
	if b.ldta == nil {
		return fmt.Errorf("layer: no ldta chunk")
	}
	if len(b.ldta.Data) < 0x70 {
		return fmt.Errorf("layer: ldta too short for Stretch write (len=%d)", len(b.ldta.Data))
	}
	divisor := binary.BigEndian.Uint32(b.ldta.Data[0x6C:0x70])
	if divisor == 0 {
		divisor = 100 // AE writes 100 for stretch (1.0 → dividend=100)
	}
	dividend := int32(math.Round(ratio * float64(divisor)))
	binary.BigEndian.PutUint32(b.ldta.Data[0x08:0x0C], uint32(dividend))
	binary.BigEndian.PutUint32(b.ldta.Data[0x6C:0x70], divisor)
	return nil
}

func (b *layerBackrefs) SetName(newName string) error {
	if b.nameChunk == nil {
		return fmt.Errorf("layer: no Utf8 name chunk to mutate")
	}
	b.nameChunk.Data = []byte(newName)
	return nil
}

func (b *layerBackrefs) SetComment(comment string) error {
	encoded := codec.EncodeCmta(comment)
	if b.commentChunk != nil {
		b.commentChunk.Data = encoded
	} else {
		if b.layrList == nil {
			return fmt.Errorf("layer: no Layr LIST reference to insert cmta into")
		}
		newCmta := &rifx.Chunk{ID: rifx.IDCmta, Data: encoded}
		b.layrList.Children = append(b.layrList.Children, newCmta)
		b.commentChunk = newCmta
	}
	// AE marks "this layer has a comment" with a flag byte at ldta @0x3C
	// (1 = present, 0 = none). Without it AE ignores the cmta chunk even
	// when the bytes are otherwise identical — see
	// incidents/layer-setcomment-cmta-append-position.md.
	if b.ldta != nil && len(b.ldta.Data) > 0x3C {
		if comment == "" {
			b.ldta.Data[0x3C] = 0
		} else {
			b.ldta.Data[0x3C] = 1
		}
	}
	return nil
}

func (b *layerBackrefs) SetTrackMatteLayer(sourceID uint32, mode TrackMatteType) error {
	if b.ldta == nil {
		return fmt.Errorf("layer: no ldta chunk")
	}
	if len(b.ldta.Data) < 0xA4 {
		return fmt.Errorf("layer: ldta too short for TrackMatteLayerID write (len=%d, need >=0xA4 — file written by AE <= 22?)", len(b.ldta.Data))
	}
	if len(b.ldta.Data) < 0x6C {
		return fmt.Errorf("layer: ldta too short for TrackMatte mode write (len=%d)", len(b.ldta.Data))
	}
	binary.BigEndian.PutUint32(b.ldta.Data[0xA0:0xA4], sourceID)
	b.ldta.Data[0x6B] = byte(mode)
	return nil
}

func (b *layerBackrefs) SetLightKind(k LightKind) error {
	if b.ldta == nil {
		return fmt.Errorf("layer: no ldta chunk")
	}
	if len(b.ldta.Data) < 0x8C {
		return fmt.Errorf("layer: ldta too short for LightKind write (len=%d)", len(b.ldta.Data))
	}
	if k > LightKindAmbient {
		return fmt.Errorf("layer: invalid LightKind %d (want 0..3)", int(k))
	}
	binary.BigEndian.PutUint32(b.ldta.Data[0x88:0x8C], uint32(k))
	return nil
}

func (b *layerBackrefs) SetLightSource(target *Layer) error {
	if b.ldta == nil {
		return fmt.Errorf("layer: no ldta chunk")
	}
	if len(b.ldta.Data) < 0x2C {
		return fmt.Errorf("layer: ldta too short for LightSource write (len=%d)", len(b.ldta.Data))
	}
	if target == nil {
		binary.BigEndian.PutUint32(b.ldta.Data[0x28:0x2C], codec.LightSourceUndefined)
		return nil
	}
	binary.BigEndian.PutUint32(b.ldta.Data[0x28:0x2C], target.ID)
	return nil
}

func (b *layerBackrefs) SetAlternateSource(item AVItem) error {
	if b.alternateSourceBlsi == nil {
		return fmt.Errorf("layer: no Essential Properties media-replacement slot (call addToMotionGraphicsTemplateAs in AE first)")
	}
	var newID uint32
	if item != nil {
		newID = item.ItemID()
	}
	if len(b.alternateSourceBlsi.Data) < 4 {
		return fmt.Errorf("layer: blsi chunk too short (len=%d)", len(b.alternateSourceBlsi.Data))
	}
	binary.BigEndian.PutUint32(b.alternateSourceBlsi.Data[0:4], newID)
	return nil
}

// SetText replaces the text-string value at btdk path /1/1/0/0/0.
//
// Fast path: when the replacement keeps the encoded byte length AND the
// per-paragraph UTF-16 character counts (same line-break positions), it is an
// in-place copy — safe for any run/paragraph structure (no counter changes).
// Equal totals are not enough: "A\rB" → "XYZ" moves a paragraph boundary,
// which the paragraph entries would have to mirror.
//
// Length-variable path: splices the text string at /1/1/0/0/0, then rebuilds
// both entry arrays that are coupled to the text. The paragraph array at
// /1/1/0/0/5/0 gets one byte-identical clone of its entry [0] per target
// paragraph, each with its /1 unit count patched (the paragraph's text plus its
// trailing \r). The style-run array at /1/1/0/0/6/0 is collapsed to a single
// clone of run [0] carrying the total unit count — matching what AE itself does
// on a whole-text replace: it drops the extra runs and keeps the first run's
// style (RE 2026-06-12, re_text_multirun.aep). Counts are UTF-16 code units
// (trailing \r included). Empty text ("") stores as the single paragraph "\r"
// (count 1). The btdk layout cache (/1/1/0/1) is left stale on purpose: AE
// recomputes it on load (ship-gate-proven, incl. paragraph-count and run-count
// changes; see incidents/text-btdk-length-variable-write-scoping.md). A
// per-character manual-kerning table (/1/1/0/0/8) is dropped before the
// rewrite — AE does the same on a whole-text replace, since per-char kerning is
// meaningless for the new characters. The length-variable path has no remaining
// refuse cases.
func (b *layerBackrefs) SetText(newText string) error {
	if b.btdsChunk == nil {
		return fmt.Errorf("layer: not a text layer")
	}
	body, bodyOff, err := codec.ExtractBtdkBody(b.btdsChunk.Data)
	if err != nil {
		return fmt.Errorf("layer: %w", err)
	}
	root := codec.ParsePSDict(body)
	if root == nil {
		return fmt.Errorf("layer: btdk dict empty")
	}
	t := codec.PsPath(root, "/1/1/0/0/0")
	if t == nil || t.Kind != codec.PsStr {
		return fmt.Errorf("layer: no text-string value at /1/1/0/0/0")
	}
	start := bodyOff + t.SrcStart
	end := bodyOff + t.SrcEnd
	encoded := codec.EncodeAEPSText(newText)
	normalized := codec.NormalizeAEParagraphText(newText)
	totalCount := codec.UTF16CodeUnitLen(normalized)
	if len(encoded) == end-start && sameParagraphProfile(t.Str, normalized) {
		copy(b.btdsChunk.Data[start:end], encoded)
		return nil
	}

	// Snapshot the whole payload before any mutation: every step below builds a
	// fresh slice, so restoring this header undoes the kerning drop and all
	// splices atomically.
	snapshot := b.btdsChunk.Data

	// Drop a per-character manual-kerning table if present: AE discards it on a
	// whole-text replace (RE 2026-06-12, re_text_kern_resize.aep), and a stale
	// per-char table would desync against the new character count. Delete the
	// whole /1/1/0/0/8 subtree, then re-parse since byte offsets shift.
	if codec.PsPath(root, "/1/1/0/0/8") != nil {
		if err := b.deleteBtdkKey("/1/1/0/0/8"); err != nil {
			b.btdsChunk.Data = snapshot
			return fmt.Errorf("layer: SetText drop kerning: %w", err)
		}
		body, _, err = codec.ExtractBtdkBody(b.btdsChunk.Data)
		if err != nil {
			b.btdsChunk.Data = snapshot
			return fmt.Errorf("layer: %w", err)
		}
		root = codec.ParsePSDict(body)
		if root == nil {
			b.btdsChunk.Data = snapshot
			return fmt.Errorf("layer: btdk dict empty after kerning drop")
		}
	}

	// Paragraph array: one clone of entry [0] per paragraph, each /1 count
	// patched (the paragraph's text plus its trailing \r). normalized ends with
	// \r, so the split drops a trailing empty element.
	pa := codec.PsPath(root, "/1/1/0/0/5/0")
	if pa == nil || pa.Kind != codec.PsArr || len(pa.Arr) == 0 {
		b.btdsChunk.Data = snapshot
		return fmt.Errorf("layer: paragraph array /1/1/0/0/5/0 missing")
	}
	paras := strings.Split(normalized, "\r")
	paras = paras[:len(paras)-1]
	counts := make([]int, len(paras))
	for i, p := range paras {
		counts[i] = codec.UTF16CodeUnitLen(p) + 1
	}
	newParaArray, err := buildEntryArray(append([]byte(nil), body[pa.Arr[0].SrcStart:pa.Arr[0].SrcEnd]...), counts)
	if err != nil {
		b.btdsChunk.Data = snapshot
		return fmt.Errorf("layer: SetText paragraph array: %w", err)
	}

	// Style-run array: collapse to a single clone of run [0] carrying the total
	// count (AE's whole-text-replace behavior — keep run [0]'s style, drop the
	// rest).
	ra := codec.PsPath(root, "/1/1/0/0/6/0")
	if ra == nil || ra.Kind != codec.PsArr || len(ra.Arr) == 0 {
		b.btdsChunk.Data = snapshot
		return fmt.Errorf("layer: style-run array /1/1/0/0/6/0 missing")
	}
	newRunArray, err := buildEntryArray(append([]byte(nil), body[ra.Arr[0].SrcStart:ra.Arr[0].SrcEnd]...), []int{totalCount})
	if err != nil {
		b.btdsChunk.Data = snapshot
		return fmt.Errorf("layer: SetText run array: %w", err)
	}

	// Three splices, each of which re-parses (offsets refresh) and patches the
	// inner btdk size header.
	for _, sp := range []struct {
		path string
		src  []byte
	}{
		{"/1/1/0/0/0", encoded},
		{"/1/1/0/0/5/0", newParaArray},
		{"/1/1/0/0/6/0", newRunArray},
	} {
		if _, err := b.splicePSValue(sp.path, sp.src); err != nil {
			b.btdsChunk.Data = snapshot
			return fmt.Errorf("layer: SetText splice %s: %w", sp.path, err)
		}
	}
	return nil
}

// buildEntryArray emits a btdk paragraph- or style-run array holding one clone
// of the template entry per element of counts, with each clone's top-level /1
// unit count replaced. Both entry kinds share the layout (a dict whose /1 is
// the unit count). Entries are newline-separated (PostScript is whitespace-
// insensitive); AE rebuilds its layout cache from these on load.
func buildEntryArray(template []byte, counts []int) ([]byte, error) {
	cs, ce, err := entryCountRange(template)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte('[')
	for _, c := range counts {
		buf.WriteByte('\n')
		buf.Write(template[:cs])
		buf.WriteString(strconv.Itoa(c))
		buf.Write(template[ce:])
	}
	buf.WriteString("\n]")
	return buf.Bytes(), nil
}

// entryCountRange parses a single btdk paragraph/run entry (a dict) and
// returns the byte range of its top-level /1 unit-count value within entry.
func entryCountRange(entry []byte) (int, int, error) {
	lx := &codec.PsLexer{D: entry}
	v := codec.ParsePSValue(lx, lx.Next())
	if v == nil || v.Kind != codec.PsDict {
		return 0, 0, fmt.Errorf("text entry is not a dict")
	}
	cnt := codec.PsPath(v, "/1")
	if cnt == nil || cnt.Kind != codec.PsNum {
		return 0, 0, fmt.Errorf("text entry has no /1 count")
	}
	return cnt.SrcStart, cnt.SrcEnd, nil
}

// sameParagraphProfile reports whether two normalized paragraph strings
// (\r-terminated lines) have identical per-paragraph UTF-16 unit counts —
// the condition under which an in-place string swap leaves every btdk
// character counter valid.
func sameParagraphProfile(a, b string) bool {
	as, bs := strings.Split(a, "\r"), strings.Split(b, "\r")
	if len(as) != len(bs) {
		return false
	}
	for i := range as {
		if codec.UTF16CodeUnitLen(as[i]) != codec.UTF16CodeUnitLen(bs[i]) {
			return false
		}
	}
	return true
}

// splicePSValue locates the codec.PsValue at `path` (relative to the btdk
// root dict, e.g. "/1/1/0/0/6/0/0/0/0/6/1") and replaces its on-disk
// bytes with newSrc. Returns the rewritten btds raw bytes and an error if
// the path doesn't resolve or the layer has no decoded text source.
func (b *layerBackrefs) splicePSValue(path string, newSrc []byte) ([]byte, error) {
	if b.btdsChunk == nil {
		return nil, fmt.Errorf("layer: not a text layer")
	}
	body, bodyOff, err := codec.ExtractBtdkBody(b.btdsChunk.Data)
	if err != nil {
		return nil, fmt.Errorf("layer: %w", err)
	}
	root := codec.ParsePSDict(body)
	if root == nil {
		return nil, fmt.Errorf("layer: btdk dict empty")
	}
	target := codec.PsPath(root, path)
	if target == nil {
		return nil, fmt.Errorf("layer: no codec.PsValue at %q", path)
	}
	if target.SrcEnd <= target.SrcStart {
		return nil, fmt.Errorf("layer: target at %q has zero-width src range", path)
	}
	absStart := bodyOff + target.SrcStart
	absEnd := bodyOff + target.SrcEnd
	return b.spliceBtdkRange(absStart, absEnd, newSrc, bodyOff)
}

// spliceBtdkRange replaces b.btdsChunk.Data[absStart:absEnd] with newSrc (nil
// deletes the range) and patches the inner btdk LIST size header so the next
// ExtractBtdkBody reads the correct length. bodyOff is the btdk body start, so
// bodyOff-8 holds that uint32 BE size.
//
//	bodyOff-12: "LIST"   bodyOff-8: size (incl. formType)   bodyOff-4: "btdk"
func (b *layerBackrefs) spliceBtdkRange(absStart, absEnd int, newSrc []byte, bodyOff int) ([]byte, error) {
	old := b.btdsChunk.Data
	if absStart < 0 || absEnd > len(old) || absStart > absEnd {
		return nil, fmt.Errorf("layer: bad btdk splice range [%d:%d] (len %d)", absStart, absEnd, len(old))
	}
	delta := len(newSrc) - (absEnd - absStart)
	newRaw := make([]byte, 0, len(old)+delta)
	newRaw = append(newRaw, old[:absStart]...)
	newRaw = append(newRaw, newSrc...)
	newRaw = append(newRaw, old[absEnd:]...)

	if delta != 0 && bodyOff >= 8 {
		sizeOff := bodyOff - 8
		oldSize := binary.BigEndian.Uint32(newRaw[sizeOff : sizeOff+4])
		binary.BigEndian.PutUint32(newRaw[sizeOff:sizeOff+4], uint32(int(oldSize)+delta))
	}

	b.btdsChunk.Data = newRaw
	return newRaw, nil
}

// deleteBtdkKey removes the entire dict entry (the "/key" name token plus its
// value) at path from its parent dict. Used to drop the manual-kerning table on
// a length-changing SetText, which AE itself discards on a whole-text replace.
func (b *layerBackrefs) deleteBtdkKey(path string) error {
	if b.btdsChunk == nil {
		return fmt.Errorf("layer: not a text layer")
	}
	body, bodyOff, err := codec.ExtractBtdkBody(b.btdsChunk.Data)
	if err != nil {
		return fmt.Errorf("layer: %w", err)
	}
	root := codec.ParsePSDict(body)
	if root == nil {
		return fmt.Errorf("layer: btdk dict empty")
	}
	v := codec.PsPath(root, path)
	if v == nil {
		return fmt.Errorf("layer: no value at %q", path)
	}
	// Walk back from the value start over whitespace, then over the key-name
	// token, to its leading '/'.
	i := v.SrcStart - 1
	for i >= 0 && isBtdkWS(body[i]) {
		i--
	}
	for i >= 0 && body[i] != '/' {
		i--
	}
	if i < 0 {
		return fmt.Errorf("layer: key token before %q not found", path)
	}
	_, err = b.spliceBtdkRange(bodyOff+i, bodyOff+v.SrcEnd, nil, bodyOff)
	return err
}

func isBtdkWS(c byte) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == 0
}

func (b *layerBackrefs) SetRunFontSize(runIdx int, sizePts float64) error {
	_, err := b.splicePSValue(codec.RunStylePath(runIdx)+"/1", []byte(codec.FormatPSReal(sizePts)))
	return err
}

func (b *layerBackrefs) SetRunTracking(runIdx int, tracking float64) error {
	_, err := b.splicePSValue(codec.RunStylePath(runIdx)+"/8", []byte(codec.FormatPSNumber(tracking)))
	return err
}

func (b *layerBackrefs) SetRunBaselineShift(runIdx int, shift float64) error {
	_, err := b.splicePSValue(codec.RunStylePath(runIdx)+"/9", []byte(codec.FormatPSReal(shift)))
	return err
}

func (b *layerBackrefs) SetRunLeading(runIdx int, leading float64) error {
	_, err := b.splicePSValue(codec.RunStylePath(runIdx)+"/5", []byte(codec.FormatPSReal(leading)))
	return err
}

func (b *layerBackrefs) SetRunAutoLeading(runIdx int, auto bool) error {
	v := "false"
	if auto {
		v = "true"
	}
	_, err := b.splicePSValue(codec.RunStylePath(runIdx)+"/4", []byte(v))
	return err
}

func (b *layerBackrefs) SetRunFontIndex(runIdx, fontIdx int) error {
	_, err := b.splicePSValue(codec.RunStylePath(runIdx)+"/0", []byte(strconv.Itoa(fontIdx)))
	return err
}

func (b *layerBackrefs) SetRunFauxBold(runIdx int, on bool) error {
	v := "false"
	if on {
		v = "true"
	}
	_, err := b.splicePSValue(codec.RunStylePath(runIdx)+"/2", []byte(v))
	return err
}

func (b *layerBackrefs) SetRunFauxItalic(runIdx int, on bool) error {
	v := "false"
	if on {
		v = "true"
	}
	_, err := b.splicePSValue(codec.RunStylePath(runIdx)+"/3", []byte(v))
	return err
}

func (b *layerBackrefs) SetRunHorizontalScale(runIdx int, scale float64) error {
	_, err := b.splicePSValue(codec.RunStylePath(runIdx)+"/6", []byte(codec.FormatPSReal(scale)))
	return err
}

func (b *layerBackrefs) SetRunVerticalScale(runIdx int, scale float64) error {
	_, err := b.splicePSValue(codec.RunStylePath(runIdx)+"/7", []byte(codec.FormatPSReal(scale)))
	return err
}

func (b *layerBackrefs) SetRunTsume(runIdx int, tsume float64) error {
	// tsume key /36 is a REAL: AE/CoolType reads a decimal-point-less number as
	// 16.16 fixed-point (value/65536), so FormatPSNumber(0.5)="0.5" is fine but an
	// integer like "50" would read as 50/65536. Use FormatPSReal so it always
	// carries the point (matches BaselineShift/Leading/Scale siblings).
	_, err := b.splicePSValue(codec.RunStylePath(runIdx)+"/36", []byte(codec.FormatPSReal(tsume)))
	return err
}

func (b *layerBackrefs) SetRunFillColor(runIdx int, rgba [4]float64) error {
	_, err := b.splicePSValue(codec.RunStylePath(runIdx)+"/53/0/1", codec.FormatPSColorArray(rgba))
	return err
}

func (b *layerBackrefs) SetRunStrokeColor(runIdx int, rgba [4]float64) error {
	_, err := b.splicePSValue(codec.RunStylePath(runIdx)+"/54/0/1", codec.FormatPSColorArray(rgba))
	return err
}

func (b *layerBackrefs) SetRunApplyStroke(runIdx int, apply bool) error {
	v := "false"
	if apply {
		v = "true"
	}
	_, err := b.splicePSValue(codec.RunStylePath(runIdx)+"/57", []byte(v))
	return err
}

func (b *layerBackrefs) SetRunStrokeWidth(runIdx int, width float64) error {
	_, err := b.splicePSValue(codec.RunStylePath(runIdx)+"/63", []byte(codec.FormatPSReal(width)))
	return err
}

func (b *layerBackrefs) SetRunCapsOption(runIdx int, caps TextCapsOption) error {
	if caps < TextCapsNormal || caps > TextCapsAllSmall {
		return fmt.Errorf("layer: invalid TextCapsOption %d", int(caps))
	}
	_, err := b.splicePSValue(codec.RunStylePath(runIdx)+"/12", []byte(strconv.Itoa(int(caps))))
	return err
}

func (b *layerBackrefs) SetRunBaselineOption(runIdx int, base TextBaselineOption) error {
	if base < TextBaselineNormal || base > TextBaselineSubscript {
		return fmt.Errorf("layer: invalid TextBaselineOption %d", int(base))
	}
	_, err := b.splicePSValue(codec.RunStylePath(runIdx)+"/13", []byte(strconv.Itoa(int(base))))
	return err
}

func (b *layerBackrefs) SetRunStrokeOverFill(runIdx int, over bool) error {
	v := "false"
	if over {
		v = "true"
	}
	_, err := b.splicePSValue(codec.RunStylePath(runIdx)+"/58", []byte(v))
	return err
}

func (b *layerBackrefs) SetRunAutoKernType(runIdx int, kt TextAutoKernType) error {
	if kt < TextAutoKernNoAuto || kt > TextAutoKernOptical {
		return fmt.Errorf("layer: invalid TextAutoKernType %d", int(kt))
	}
	_, err := b.splicePSValue(codec.RunStylePath(runIdx)+"/11", []byte(strconv.Itoa(int(kt))))
	return err
}

func (b *layerBackrefs) SetRunNoBreak(runIdx int, on bool) error {
	v := "false"
	if on {
		v = "true"
	}
	_, err := b.splicePSValue(codec.RunStylePath(runIdx)+"/52", []byte(v))
	return err
}

func (b *layerBackrefs) SetRunLineJoinType(runIdx int, j TextLineJoinType) error {
	if j < TextLineJoinMiter || j > TextLineJoinBevel {
		return fmt.Errorf("layer: invalid TextLineJoinType %d", int(j))
	}
	_, err := b.splicePSValue(codec.RunStylePath(runIdx)+"/62", []byte(strconv.Itoa(int(j))))
	return err
}

func (b *layerBackrefs) SetRunDigitSet(runIdx int, d TextDigitSet) error {
	if d < TextDigitSetDefault || d > TextDigitSetArabicRTL {
		return fmt.Errorf("layer: invalid TextDigitSet %d", int(d))
	}
	_, err := b.splicePSValue(codec.RunStylePath(runIdx)+"/70", []byte(strconv.Itoa(int(d))))
	return err
}

func (b *layerBackrefs) SetParagraphJustification(paraIdx int, j TextJustification) error {
	_, err := b.splicePSValue(codec.ParagraphStylePath(paraIdx)+"/0", []byte(strconv.Itoa(int(j))))
	return err
}

// The five paragraph indent/spacing keys /1../5 are point-measurement REALs:
// AE/CoolType reads a decimal-point-less number as 16.16 fixed-point (value/65536),
// so these MUST use FormatPSReal (not FormatPSNumber). Same bug class as
// SetRunTsume — fixed + gated by text_para_shipgate (batch 7b).
func (b *layerBackrefs) SetParagraphFirstLineIndent(paraIdx int, v float64) error {
	_, err := b.splicePSValue(codec.ParagraphStylePath(paraIdx)+"/1", []byte(codec.FormatPSReal(v)))
	return err
}

func (b *layerBackrefs) SetParagraphStartIndent(paraIdx int, v float64) error {
	_, err := b.splicePSValue(codec.ParagraphStylePath(paraIdx)+"/2", []byte(codec.FormatPSReal(v)))
	return err
}

func (b *layerBackrefs) SetParagraphEndIndent(paraIdx int, v float64) error {
	_, err := b.splicePSValue(codec.ParagraphStylePath(paraIdx)+"/3", []byte(codec.FormatPSReal(v)))
	return err
}

func (b *layerBackrefs) SetParagraphSpaceBefore(paraIdx int, v float64) error {
	_, err := b.splicePSValue(codec.ParagraphStylePath(paraIdx)+"/4", []byte(codec.FormatPSReal(v)))
	return err
}

func (b *layerBackrefs) SetParagraphSpaceAfter(paraIdx int, v float64) error {
	_, err := b.splicePSValue(codec.ParagraphStylePath(paraIdx)+"/5", []byte(codec.FormatPSReal(v)))
	return err
}

func (b *layerBackrefs) SetParagraphAutoHyphenate(paraIdx int, on bool) error {
	v := "false"
	if on {
		v = "true"
	}
	_, err := b.splicePSValue(codec.ParagraphStylePath(paraIdx)+"/9", []byte(v))
	return err
}

func (b *layerBackrefs) SetParagraphLeadingType(paraIdx int, lt TextLeadingType) error {
	if lt < TextLeadingRoman || lt > TextLeadingJapanese {
		return fmt.Errorf("layer: invalid TextLeadingType %d", int(lt))
	}
	_, err := b.splicePSValue(codec.ParagraphStylePath(paraIdx)+"/8", []byte(strconv.Itoa(int(lt))))
	return err
}

func (b *layerBackrefs) SetParagraphHangingRoman(paraIdx int, on bool) error {
	v := "false"
	if on {
		v = "true"
	}
	_, err := b.splicePSValue(codec.ParagraphStylePath(paraIdx)+"/21", []byte(v))
	return err
}

func (b *layerBackrefs) SetParagraphDirection(paraIdx int, d TextParagraphDirection) error {
	if d < TextDirectionLeftToRight || d > TextDirectionRightToLeft {
		return fmt.Errorf("layer: invalid TextParagraphDirection %d", int(d))
	}
	_, err := b.splicePSValue(codec.ParagraphStylePath(paraIdx)+"/33", []byte(strconv.Itoa(int(d))))
	return err
}

func (b *layerBackrefs) SetManualKerning(values []int) error {
	// Write per-char values first. splicePSValue re-parses the body
	// each call, so length-changing splices stay consistent across the
	// loop. Path: /1/1/0/0/8/0/{i}/0/0.
	for i, v := range values {
		path := fmt.Sprintf("/1/1/0/0/8/0/%d/0/0", i)
		if _, err := b.splicePSValue(path, []byte(strconv.Itoa(v))); err != nil {
			return err
		}
	}
	// Mirror values[0] to the sibling first-char scalar /1/1[0]/0/7.
	_, err := b.splicePSValue("/1/1/0/0/7", []byte(strconv.Itoa(values[0])))
	return err
}

func (b *layerBackrefs) AddFont(fontName string) (int, error) {
	if b.btdsChunk == nil {
		return -1, fmt.Errorf("layer: not a text layer")
	}
	if fontName == "" {
		return -1, fmt.Errorf("layer: fontName must be non-empty")
	}
	body, bodyOff, err := codec.ExtractBtdkBody(b.btdsChunk.Data)
	if err != nil {
		return -1, fmt.Errorf("layer: %w", err)
	}
	root := codec.ParsePSDict(body)
	if root == nil {
		return -1, fmt.Errorf("layer: btdk dict empty")
	}
	arr := codec.PsPath(root, "/0/1/0")
	if arr == nil || arr.Kind != codec.PsArr {
		return -1, fmt.Errorf("layer: Fonts array at /0/1/0 not found")
	}
	// We splice just inside the closing `]`. Find the array's last
	// child srcEnd and insert " <new>" between it and `]`. When the
	// array is empty (rare), insert directly after `[`.
	var insertAt int
	if len(arr.Arr) == 0 {
		// arr.SrcStart points at the `[`; insert one byte after.
		insertAt = arr.SrcStart + 1
	} else {
		last := arr.Arr[len(arr.Arr)-1]
		insertAt = last.SrcEnd
	}
	newEntry := codec.SerializeFontEntry(fontName)

	// Build the byte sequence to inject: leading space + serialized entry.
	var injected []byte
	injected = append(injected, ' ')
	injected = append(injected, newEntry...)

	old := b.btdsChunk.Data
	abs := bodyOff + insertAt
	newRaw := make([]byte, 0, len(old)+len(injected))
	newRaw = append(newRaw, old[:abs]...)
	newRaw = append(newRaw, injected...)
	newRaw = append(newRaw, old[abs:]...)

	// Update inner LIST btdk size header (same invariant as splicePSValue).
	if bodyOff >= 8 {
		sizeOff := bodyOff - 8
		oldSize := binary.BigEndian.Uint32(newRaw[sizeOff : sizeOff+4])
		binary.BigEndian.PutUint32(newRaw[sizeOff:sizeOff+4], uint32(int(oldSize)+len(injected)))
	}

	b.btdsChunk.Data = newRaw
	return len(arr.Arr), nil // index of newly-added font (was len-1 + 1 → len of old)
}
