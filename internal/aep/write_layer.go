package aep

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strings"

	"github.com/example/aep-parser/internal/rifx"
)

// Layer setters covered in this file are all length-preserving single-byte
// or single-bit edits inside the layer's ldta chunk. Each setter:
//
//   - validates the layer has an underlying ldta chunk (else returns an
//     error — happens for layers built outside the parser);
//   - mutates the ldta bytes in place;
//   - keeps the matching Go field in sync so later reads see the new value.
//
// The next call to Project.WriteAEP serializes the change.

// flag-bit positions inside ldta @0x25-0x27 — must mirror parse_layer.go's
// decode. Keeping them centralized so future ldta-version surprises only
// need updating in one place.
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
func (l *Layer) setFlagBit(b ldtaFlagBit, v bool) error {
	if l.ldta == nil {
		return fmt.Errorf("layer %q: no ldta chunk (built outside parser?)", l.Name)
	}
	if len(l.ldta.Data) <= b.off {
		return fmt.Errorf("layer %q: ldta @%#x out of range (len=%d)", l.Name, b.off, len(l.ldta.Data))
	}
	if v {
		l.ldta.Data[b.off] |= b.mask
	} else {
		l.ldta.Data[b.off] &^= b.mask
	}
	return nil
}

// SetVisible toggles the layer's video switch (the 👁️ icon).
// length-preserving (single bit @ldta 0x27).
func (l *Layer) SetVisible(v bool) error {
	if err := l.setFlagBit(flagVisible, v); err != nil {
		return err
	}
	l.Visible = v
	return nil
}

// SetSolo toggles the layer's Solo flag.
// length-preserving (single bit @ldta 0x26).
func (l *Layer) SetSolo(v bool) error {
	if err := l.setFlagBit(flagSolo, v); err != nil {
		return err
	}
	l.Solo = v
	return nil
}

// SetShy toggles the layer's Shy flag (hides from the shy-filter view).
// length-preserving (single bit @ldta 0x27).
func (l *Layer) SetShy(v bool) error {
	if err := l.setFlagBit(flagShy, v); err != nil {
		return err
	}
	l.Shy = v
	return nil
}

// SetLocked toggles the layer's Lock flag (🔒). When locked, AE refuses
// edits in the timeline UI; the AEP file itself is still mutable.
// length-preserving (single bit @ldta 0x27).
func (l *Layer) SetLocked(v bool) error {
	if err := l.setFlagBit(flagLocked, v); err != nil {
		return err
	}
	l.Locked = v
	return nil
}

// SetEffectsEnabled toggles the layer's fx switch (whether effects render).
// length-preserving (single bit @ldta 0x27).
func (l *Layer) SetEffectsEnabled(v bool) error {
	if err := l.setFlagBit(flagEffectsEnabled, v); err != nil {
		return err
	}
	l.EffectsEnabled = v
	return nil
}

// SetMotionBlur toggles the layer's motion-blur switch.
// length-preserving (single bit @ldta 0x27).
func (l *Layer) SetMotionBlur(v bool) error {
	if err := l.setFlagBit(flagMotionBlur, v); err != nil {
		return err
	}
	l.MotionBlur = v
	return nil
}

// SetAudioEnabled toggles the layer's audio switch.
// length-preserving (single bit @ldta 0x27).
func (l *Layer) SetAudioEnabled(v bool) error {
	if err := l.setFlagBit(flagAudioEnabled, v); err != nil {
		return err
	}
	l.AudioEnabled = v
	return nil
}

// SetFrameBlendEnabled toggles the layer's frame-blend switch.
// length-preserving (single bit @ldta 0x27).
func (l *Layer) SetFrameBlendEnabled(v bool) error {
	if err := l.setFlagBit(flagFrameBlendEnabled, v); err != nil {
		return err
	}
	l.FrameBlendEnabled = v
	return nil
}

// SetCollapseTransform toggles "Collapse Transformations" (for nested
// comps) or "Continuously Rasterize" (for Illustrator / shape layers).
// length-preserving (single bit @ldta 0x27).
func (l *Layer) SetCollapseTransform(v bool) error {
	if err := l.setFlagBit(flagCollapseTransform, v); err != nil {
		return err
	}
	l.CollapseTransform = v
	return nil
}

// SetIs3D toggles the layer's 3D-layer switch.
// length-preserving (single bit @ldta 0x26).
func (l *Layer) SetIs3D(v bool) error {
	if err := l.setFlagBit(flagIs3D, v); err != nil {
		return err
	}
	l.Is3D = v
	return nil
}

// SetIsAdjust toggles the layer's "Adjustment Layer" switch.
// length-preserving (single bit @ldta 0x26).
func (l *Layer) SetIsAdjust(v bool) error {
	if err := l.setFlagBit(flagIsAdjust, v); err != nil {
		return err
	}
	l.IsAdjust = v
	return nil
}

// SetIsGuide toggles the layer's "Guide Layer" switch (AE renders it
// in the comp viewer but excludes it from output).
// length-preserving (single bit @ldta 0x25).
func (l *Layer) SetIsGuide(v bool) error {
	if err := l.setFlagBit(flagIsGuide, v); err != nil {
		return err
	}
	l.IsGuide = v
	return nil
}

// SetIsNull toggles the Null-Object marker bit. AE's UI creates null
// layers via Layer > New > Null Object; flipping this post-hoc is
// supported by the byte but produces uncommon AE behavior — set only
// when you understand the consequences.
// length-preserving (single bit @ldta 0x26).
func (l *Layer) SetIsNull(v bool) error {
	if err := l.setFlagBit(flagIsNull, v); err != nil {
		return err
	}
	l.IsNull = v
	return nil
}

// SetMarkersLocked toggles "Lock markers" on the layer.
// length-preserving (single bit @ldta 0x26).
func (l *Layer) SetMarkersLocked(v bool) error {
	if err := l.setFlagBit(flagMarkersLocked, v); err != nil {
		return err
	}
	l.MarkersLocked = v
	return nil
}

// SetSamplingBicubic switches between Bilinear (false) and Bicubic (true)
// sampling for the layer.
// length-preserving (single bit @ldta 0x25).
func (l *Layer) SetSamplingBicubic(v bool) error {
	if err := l.setFlagBit(flagSamplingBicubic, v); err != nil {
		return err
	}
	l.SamplingBicubic = v
	return nil
}

// SetFrameBlendPixelMotion switches frame blending mode between
// Frame Mix (false) and Pixel Motion (true). Only relevant when
// FrameBlendEnabled is also true.
// length-preserving (single bit @ldta 0x25).
func (l *Layer) SetFrameBlendPixelMotion(v bool) error {
	if err := l.setFlagBit(flagFrameBlendPixelMotion, v); err != nil {
		return err
	}
	l.FrameBlendPixelMotion = v
	return nil
}

// SetShy / Solo / Locked / etc. cover bool toggles. Multi-value byte
// fields follow.

// SetBlendingMode writes a new blending-mode enum byte to ldta @0x63.
// length-preserving (single byte).
func (l *Layer) SetBlendingMode(m BlendingMode) error {
	if l.ldta == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if len(l.ldta.Data) <= 0x63 {
		return fmt.Errorf("layer %q: ldta too short for BlendingMode write (len=%d)", l.Name, len(l.ldta.Data))
	}
	l.ldta.Data[0x63] = byte(m)
	l.BlendingMode = m
	return nil
}

// SetTrackMatte writes a new track-matte type byte to ldta @0x6B.
// length-preserving (single byte).
//
// AE additionally requires the matte source layer to sit immediately
// above this layer in the comp; this setter only flips the mode byte
// and does NOT reorder layers.
func (l *Layer) SetTrackMatte(t TrackMatteType) error {
	if l.ldta == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if len(l.ldta.Data) <= 0x6B {
		return fmt.Errorf("layer %q: ldta too short for TrackMatte write (len=%d)", l.Name, len(l.ldta.Data))
	}
	l.ldta.Data[0x6B] = byte(t)
	l.TrackMatte = t
	return nil
}

// SetLabel writes a new timeline label-color index (0..16) to ldta @0x3D.
// Indices outside 0..16 are written verbatim (AE shows index 0 for any
// unknown value but the byte is preserved on round-trip).
// length-preserving (single byte).
func (l *Layer) SetLabel(index uint8) error {
	if l.ldta == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if len(l.ldta.Data) <= 0x3D {
		return fmt.Errorf("layer %q: ldta too short for Label write (len=%d)", l.Name, len(l.ldta.Data))
	}
	l.ldta.Data[0x3D] = index
	l.Label = index
	return nil
}

// SetQuality writes a new render-quality enum (Wireframe / Draft / Best)
// to ldta @0x04 (uint16 BE).
// length-preserving (2 bytes).
func (l *Layer) SetQuality(q LayerQuality) error {
	if l.ldta == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if len(l.ldta.Data) < 0x06 {
		return fmt.Errorf("layer %q: ldta too short for Quality write (len=%d)", l.Name, len(l.ldta.Data))
	}
	binary.BigEndian.PutUint16(l.ldta.Data[0x04:0x06], uint16(q))
	l.Quality = q
	return nil
}

// SetParent rewrites the layer's parent-layer ID (ldta @0x84) to
// `parentID`. Pass 0 to clear the parent (= "no parent", AE shows
// "None" in the timeline). length-preserving (4 bytes).
//
// When the layer was created via the parser and lives inside a
// composition, the new parentID is validated against same-comp layers
// — passing an ID that doesn't resolve returns an error and doesn't
// touch the bytes. Cross-comp parenting isn't allowed by AE.
//
// `parentID == l.ID` is rejected (would create a self-parent cycle).
func (l *Layer) SetParent(parentID uint32) error {
	if l.ldta == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if len(l.ldta.Data) < 0x88 {
		return fmt.Errorf("layer %q: ldta too short for ParentID write (len=%d)", l.Name, len(l.ldta.Data))
	}
	if parentID != 0 && parentID == l.ID {
		return fmt.Errorf("layer %q: self-parenting (parentID == own ID = %d) not allowed", l.Name, l.ID)
	}
	if parentID != 0 && l.comp != nil {
		if l.comp.LayerByID(parentID) == nil {
			return fmt.Errorf("layer %q: parentID %d not found in comp %q", l.Name, parentID, l.comp.Name)
		}
	}
	binary.BigEndian.PutUint32(l.ldta.Data[0x84:0x88], parentID)
	l.ParentID = parentID
	return nil
}

// SetSource rewrites the layer's source-item ID (ldta @0x28). For
// regular AV layers this is the footage or pre-comp ID. length-
// preserving (4 bytes).
//
// When the layer was created via the parser and the owning Project is
// known, `sourceID` is validated to match an existing Composition /
// Footage item — passing an ID that doesn't resolve returns an error
// and doesn't touch the bytes. Pass 0 to clear (rare; usually means
// the layer becomes a Null-like "no source" stub).
//
// Note: AE caches some source-derived metadata (Width/Height) on the
// layer at render time, not in ldta. Changing source preserves the
// rest of ldta verbatim, which is what we want.
func (l *Layer) SetSource(sourceID uint32) error {
	if l.ldta == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if len(l.ldta.Data) < 0x2C {
		return fmt.Errorf("layer %q: ldta too short for SourceID write (len=%d)", l.Name, len(l.ldta.Data))
	}
	if sourceID != 0 && l.comp != nil && l.comp.proj != nil {
		p := l.comp.proj
		if p.CompositionByID(sourceID) == nil && !footageWithID(p, sourceID) {
			return fmt.Errorf("layer %q: sourceID %d not found in project items", l.Name, sourceID)
		}
	}
	binary.BigEndian.PutUint32(l.ldta.Data[0x28:0x2C], sourceID)
	l.SourceID = sourceID
	return nil
}

// footageWithID reports whether any Footage item has the given ID.
// Used by SetSource for source-ID validation.
func footageWithID(p *Project, id uint32) bool {
	for _, f := range p.Footage {
		if f.ID == id {
			return true
		}
	}
	return false
}

// SetAutoOrient writes the layer's auto-orient mode by clearing the
// three mutually-exclusive bits across ldta @0x25/@0x26 and setting
// the one matching the requested AutoOrientType. length-preserving
// (touches 2 bytes).
//
// Bit positions (mirroring parse_layer.go's decode):
//   - CharactersTowardCamera → 0x25 bit 4
//   - CameraOrPointOfInterest → 0x26 bit 5
//   - AlongPath → 0x26 bit 0
//   - None → all three cleared
//
// Unknown enum values are treated as None (= no-op clearing).
func (l *Layer) SetAutoOrient(t AutoOrientType) error {
	if l.ldta == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if len(l.ldta.Data) <= 0x26 {
		return fmt.Errorf("layer %q: ldta too short for AutoOrient write (len=%d)", l.Name, len(l.ldta.Data))
	}
	// Clear all three bits first.
	l.ldta.Data[0x25] &^= 0x10
	l.ldta.Data[0x26] &^= 0x20 | 0x01
	switch t {
	case AutoOrientCharactersTowardCamera:
		l.ldta.Data[0x25] |= 0x10
	case AutoOrientCameraOrPointOfInterest:
		l.ldta.Data[0x26] |= 0x20
	case AutoOrientAlongPath:
		l.ldta.Data[0x26] |= 0x01
	case AutoOrientNone:
		// already cleared
	default:
		return fmt.Errorf("layer %q: unknown AutoOrientType %d", l.Name, int(t))
	}
	l.AutoOrient = t
	return nil
}

// SetPreserveTransparency toggles "Preserve Underlying Transparency"
// (ldta @0x67, single byte 0/1).
// length-preserving (single byte).
func (l *Layer) SetPreserveTransparency(v bool) error {
	if l.ldta == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if len(l.ldta.Data) <= 0x67 {
		return fmt.Errorf("layer %q: ldta too short for PreserveTransparency write (len=%d)", l.Name, len(l.ldta.Data))
	}
	if v {
		l.ldta.Data[0x67] = 1
	} else {
		l.ldta.Data[0x67] = 0
	}
	l.PreserveTransparency = v
	return nil
}

// ──────────────────────────────────────────────────────────────────
// Time-field setters — dividend/divisor pairs in ldta. All length-
// preserving (8 bytes each, fixed offsets). We keep the existing
// divisor when non-zero (typical: comp's tick rate) and fall back to
// 600 when missing, matching how AE writes these fields.
// ──────────────────────────────────────────────────────────────────

// setLdtaFrac writes (dividend, divisor) at ldta @off / @off+4 such
// that dividend/divisor ≈ seconds. Reuses the existing divisor when
// non-zero; otherwise picks 600 (AE's standard for layer time fields).
// Mutates 8 bytes. Returns the resulting (dividend, divisor) so the
// caller can mirror to its Go field.
func (l *Layer) setLdtaFrac(off int, seconds float64, label string) (int32, uint32, error) {
	if l.ldta == nil {
		return 0, 0, fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if len(l.ldta.Data) < off+8 {
		return 0, 0, fmt.Errorf("layer %q: ldta too short for %s write (len=%d, off=%#x)", l.Name, label, len(l.ldta.Data), off)
	}
	divisor := binary.BigEndian.Uint32(l.ldta.Data[off+4 : off+8])
	if divisor == 0 {
		divisor = 600
	}
	dividend := int32(math.Round(seconds * float64(divisor)))
	binary.BigEndian.PutUint32(l.ldta.Data[off:off+4], uint32(dividend))
	binary.BigEndian.PutUint32(l.ldta.Data[off+4:off+8], divisor)
	return dividend, divisor, nil
}

// SetStartTime writes the layer's start time (seconds) to
// ldta @0x0C/@0x10. length-preserving (8 bytes).
//
// AE allows negative start times (pre-roll). Mirrors the change to
// `Layer.StartTime`.
func (l *Layer) SetStartTime(seconds float64) error {
	dividend, divisor, err := l.setLdtaFrac(0x0C, seconds, "StartTime")
	if err != nil {
		return err
	}
	l.StartTime = float64(dividend) / float64(divisor)
	return nil
}

// SetInPoint writes the layer's source-media in-point (seconds) to
// ldta @0x14/@0x18, then refreshes `Layer.Duration` (= out − in).
// length-preserving (8 bytes).
func (l *Layer) SetInPoint(seconds float64) error {
	if _, _, err := l.setLdtaFrac(0x14, seconds, "InPoint"); err != nil {
		return err
	}
	l.recomputeDurationFromLdta()
	return nil
}

// SetOutPoint writes the layer's source-media out-point (seconds) to
// ldta @0x1C/@0x20, then refreshes `Layer.Duration`.
// length-preserving (8 bytes).
func (l *Layer) SetOutPoint(seconds float64) error {
	if _, _, err := l.setLdtaFrac(0x1C, seconds, "OutPoint"); err != nil {
		return err
	}
	l.recomputeDurationFromLdta()
	return nil
}

// recomputeDurationFromLdta re-reads in/out points from ldta and
// updates `Layer.Duration`. Called by SetInPoint / SetOutPoint so the
// in-memory Duration stays consistent with the underlying bytes.
func (l *Layer) recomputeDurationFromLdta() {
	if l.ldta == nil || len(l.ldta.Data) < 0x24 {
		return
	}
	d := l.ldta.Data
	inDivisor := binary.BigEndian.Uint32(d[0x18:0x1C])
	outDivisor := binary.BigEndian.Uint32(d[0x20:0x24])
	if inDivisor == 0 || outDivisor == 0 {
		return
	}
	in := float64(int32(binary.BigEndian.Uint32(d[0x14:0x18]))) / float64(inDivisor)
	out := float64(int32(binary.BigEndian.Uint32(d[0x1C:0x20]))) / float64(outDivisor)
	l.Duration = out - in
}

// SetStretch writes the layer's time-stretch ratio (1.0 = normal,
// 2.0 = 2× slow) to ldta @0x08 (dividend) / @0x6C (divisor). Unlike
// the other time fields the dividend/divisor pair is **split across
// the ldta** — dividend lives near the top, divisor in the trailing
// section.
// length-preserving (8 bytes total in two 4-byte writes).
func (l *Layer) SetStretch(ratio float64) error {
	if l.ldta == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if len(l.ldta.Data) < 0x70 {
		return fmt.Errorf("layer %q: ldta too short for Stretch write (len=%d)", l.Name, len(l.ldta.Data))
	}
	divisor := binary.BigEndian.Uint32(l.ldta.Data[0x6C:0x70])
	if divisor == 0 {
		divisor = 100 // AE writes 100 for stretch (1.0 → dividend=100)
	}
	dividend := int32(math.Round(ratio * float64(divisor)))
	binary.BigEndian.PutUint32(l.ldta.Data[0x08:0x0C], uint32(dividend))
	binary.BigEndian.PutUint32(l.ldta.Data[0x6C:0x70], divisor)
	l.Stretch = float64(dividend) / float64(divisor)
	return nil
}

// SetName rewrites the layer's display name (the AE timeline label).
// length-variable — the underlying Utf8 chunk's data slice is replaced
// wholesale; WriteAEP recomputes ancestor LIST sizes.
//
// Returns an error when the parsed layer has no Utf8 name chunk
// (unusual — most AE-written layers have one even for default names).
func (l *Layer) SetName(newName string) error {
	if l.nameChunk == nil {
		return fmt.Errorf("layer %q: no Utf8 name chunk to mutate", l.Name)
	}
	l.nameChunk.Data = []byte(newName)
	l.Name = newName
	return nil
}

// SetComment rewrites the layer's comment text (AE's "Comments"
// timeline column / Layer Settings dialog).
//
// length-variable — the underlying cmta chunk's data is replaced (or a
// new cmta chunk is inserted into the Layr LIST when none existed).
// Encoding mirrors what AE writes: LF → CRLF + a single NUL terminator.
func (l *Layer) SetComment(comment string) error {
	encoded := encodeCmta(comment)
	if l.commentChunk != nil {
		l.commentChunk.Data = encoded
	} else {
		if l.layrList == nil {
			return fmt.Errorf("layer %q: no Layr LIST reference to insert cmta into", l.Name)
		}
		newCmta := &rifx.Chunk{ID: rifx.IDCmta, Data: encoded}
		l.layrList.Children = append(l.layrList.Children, newCmta)
		l.commentChunk = newCmta
	}
	l.Comment = comment
	return nil
}

// encodeCmta encodes a Go-friendly LF-separated string into AE's cmta
// payload format: CRLF line endings + single NUL terminator.
func encodeCmta(s string) []byte {
	if s == "" {
		// AE typically writes a single NUL even for the empty case;
		// matches what decodeCmta gracefully reads back as "".
		return []byte{0}
	}
	// Normalize LF → CRLF; already-CRLF inputs are left alone via the
	// helper that converts only lone LFs.
	out := normalizeToCRLF(s) + "\x00"
	return []byte(out)
}

// normalizeToCRLF converts every \n that isn't preceded by \r into
// \r\n. \r\n in the input is preserved as-is.
func normalizeToCRLF(s string) string {
	if !containsLoneLF(s) {
		return s
	}
	out := make([]byte, 0, len(s)+8)
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' && (i == 0 || s[i-1] != '\r') {
			out = append(out, '\r', '\n')
			continue
		}
		out = append(out, s[i])
	}
	return string(out)
}

func containsLoneLF(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' && (i == 0 || s[i-1] != '\r') {
			return true
		}
	}
	return false
}

// SetTrackMatteLayer rewrites this layer's explicit track matte source
// (ldta @0xA0, AE 23+) to `sourceID` and updates the track matte mode
// (ldta @0x6B) to `mode`. Pass sourceID=0 with mode=TrackMatteNone to
// clear the matte. length-preserving (4 bytes + 1 byte).
//
// When the layer lives inside a parsed composition, `sourceID` is
// validated against same-comp layers — passing an ID that doesn't
// resolve returns an error and doesn't touch the bytes. AE 23+ requires
// the matte source layer to live in the same comp.
//
// `sourceID == l.ID` is rejected (self-matte is a no-op in AE; usually
// indicates a programmer error). Pass mode=TrackMatteNone with non-zero
// sourceID for "preserve target, no matte applied yet" — AE allows that
// (the source pointer stays but the matte channel is disabled).
//
// On AE 2020 / 2022 files (ldta 160 bytes), this call errors — the
// @0xA0 slot doesn't exist. Re-save the file through AE 23+ first to
// extend ldta, then this setter works.
func (l *Layer) SetTrackMatteLayer(sourceID uint32, mode TrackMatteType) error {
	if l.ldta == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if len(l.ldta.Data) < 0xA4 {
		return fmt.Errorf("layer %q: ldta too short for TrackMatteLayerID write (len=%d, need >=0xA4 — file written by AE <= 22?)", l.Name, len(l.ldta.Data))
	}
	if len(l.ldta.Data) < 0x6C {
		return fmt.Errorf("layer %q: ldta too short for TrackMatte mode write (len=%d)", l.Name, len(l.ldta.Data))
	}
	if sourceID != 0 && sourceID == l.ID {
		return fmt.Errorf("layer %q: self-matte (sourceID == own ID = %d) not allowed", l.Name, l.ID)
	}
	if sourceID != 0 && l.comp != nil {
		if l.comp.LayerByID(sourceID) == nil {
			return fmt.Errorf("layer %q: sourceID %d not found in comp %q", l.Name, sourceID, l.comp.Name)
		}
	}
	binary.BigEndian.PutUint32(l.ldta.Data[0xA0:0xA4], sourceID)
	l.ldta.Data[0x6B] = byte(mode)
	l.TrackMatteLayerID = sourceID
	l.TrackMatte = mode
	return nil
}

// ClearTrackMatteLayer is a shorthand for SetTrackMatteLayer(0, TrackMatteNone).
// Removes the explicit matte source AND the matte mode in one call.
func (l *Layer) ClearTrackMatteLayer() error {
	return l.SetTrackMatteLayer(0, TrackMatteNone)
}

// SetLightKind rewrites the light layer's kind (ldta @0x88, 4 bytes BE
// uint32). length-preserving. Only meaningful when Type == LayerTypeLight;
// AE silently accepts the byte change on non-light layers but it has
// no visual effect.
//
// On AE 22 / 2020 ldta (160 bytes), @0x88 is still present (parent ID
// at @0x84 + 4 bytes after = @0x88), so this setter works on older
// fixtures too.
func (l *Layer) SetLightKind(k LightKind) error {
	if l.ldta == nil {
		return fmt.Errorf("layer %q: no ldta chunk", l.Name)
	}
	if len(l.ldta.Data) < 0x8C {
		return fmt.Errorf("layer %q: ldta too short for LightKind write (len=%d)", l.Name, len(l.ldta.Data))
	}
	if k > LightKindAmbient {
		return fmt.Errorf("layer %q: invalid LightKind %d (want 0..3)", l.Name, int(k))
	}
	binary.BigEndian.PutUint32(l.ldta.Data[0x88:0x8C], uint32(k))
	l.LightKind = k
	return nil
}

// SetAlternateSource overrides this layer's source via the Essential
// Properties → Media Replacement slot (4-byte blsi write, length-
// preserving). Pass nil (or a zero-id item) to clear the override —
// equivalent to ClearAlternateSource. Mirrors AE script's
// Property.setAlternateSource.
//
// Requires the layer to already have an Essential Properties slot —
// AE only persists the blsi chunk after the source-side layer has been
// promoted via `AVLayer.addToMotionGraphicsTemplateAs()` in the wrapper
// precomp. If the slot is missing (HasAlternateSourceSlot() == false),
// this errors instead of attempting structural chunk insertion (which
// would break length-preserving roundtrip).
//
// When the layer's owning project is wired up, `item.ItemID()` is
// validated against the project's AV items; an id unknown to the
// project returns an error and doesn't touch bytes. AE rejects items
// whose `isMediaReplacementCompatible == false` (e.g., still images vs.
// video) at script time — we don't replicate that check; the rendered
// .aep is still parsed cleanly by AE regardless.
func (l *Layer) SetAlternateSource(item AVItem) error {
	if l.alternateSourceBlsi == nil {
		return fmt.Errorf("layer %q: no Essential Properties media-replacement slot (call addToMotionGraphicsTemplateAs in AE first)", l.Name)
	}
	var newID uint32
	if item != nil {
		newID = item.ItemID()
	}
	if newID != 0 && l.comp != nil && l.comp.proj != nil {
		if l.comp.proj.AVItemByID(newID) == nil {
			return fmt.Errorf("layer %q: alternate source item id %d not in project", l.Name, newID)
		}
	}
	if len(l.alternateSourceBlsi.Data) < 4 {
		return fmt.Errorf("layer %q: blsi chunk too short (len=%d)", l.Name, len(l.alternateSourceBlsi.Data))
	}
	binary.BigEndian.PutUint32(l.alternateSourceBlsi.Data[0:4], newID)
	l.AlternateSourceID = newID
	return nil
}

// ClearAlternateSource removes the media-replacement override (blsi
// = 0) so AE falls back to the wrapper precomp's default source.
// Equivalent to SetAlternateSource(nil).
func (l *Layer) ClearAlternateSource() error { return l.SetAlternateSource(nil) }

// SetText replaces a text layer's user-visible text in-place. The new
// text must encode to exactly the same number of bytes as the original
// PostScript string (length-preserving constraint — the surrounding
// btds/btdk chunk headers reference fixed byte counts that we won't
// rewrite). Returns an error when the byte budget doesn't match, when
// the layer isn't a text layer, or when the original text-string
// location wasn't recorded during parsing.
//
// Encoding parity with AE: input is split on '\n' (each segment becomes
// a paragraph terminated by AE's '\r' convention), then encoded as
// UTF-16BE with a leading FE FF BOM, with PostScript specials ( ) \
// escaped at the byte level. Use [TextEncodedByteLen] to predict
// whether a candidate string fits before calling.
//
// After a successful call Layer.TextSource is re-decoded so subsequent
// reads reflect the new value; Layer.WriteAEP serializes the change.
func (l *Layer) SetText(newText string) error {
	if l.TextSource == nil || l.TextSourceRaw == nil {
		return fmt.Errorf("layer %q: not a text layer (or text source failed to decode)", l.Name)
	}
	if l.TextSource.textStringEnd <= l.TextSource.textStringStart {
		return fmt.Errorf("layer %q: original text-string offset not recorded; can't splice", l.Name)
	}
	encoded := encodeAEPSText(newText)
	oldLen := l.TextSource.textStringEnd - l.TextSource.textStringStart
	if len(encoded) != oldLen {
		return fmt.Errorf("layer %q: SetText length mismatch (new=%d bytes, old=%d bytes — length-preserving only; pad input to match)",
			l.Name, len(encoded), oldLen)
	}
	copy(l.TextSourceRaw[l.TextSource.textStringStart:l.TextSource.textStringEnd], encoded)
	// Re-decode so Layer.TextSource.Text + per-run/paragraph stay accurate.
	ts, _ := decodeTextSource(l.TextSourceRaw)
	if ts != nil {
		l.TextSource = ts
	}
	return nil
}

// TextEncodedByteLen returns the encoded byte length that SetText
// would produce for s, so callers can check if a candidate value fits
// the length-preserving budget without trial-and-error. Use it like:
//
//	if aep.TextEncodedByteLen(candidate) != aep.TextEncodedByteLen(layer.TextSource.Text) {
//	    // doesn't fit — pad/truncate or skip
//	}
func TextEncodedByteLen(s string) int {
	return len(encodeAEPSText(s))
}

// encodeAEPSText encodes a Go string as the byte sequence AE writes
// for the text-document string field in btdk: open paren, FE FF BOM,
// UTF-16BE code units (with surrogate-pair expansion), '\r' at end of
// every paragraph (input '\n' becomes '\r'; a trailing '\r' is added
// if absent), then close paren. Byte values 0x28/0x29/0x5C anywhere in
// the resulting stream (high or low byte of any code unit) are
// backslash-escaped per PostScript string rules.
func encodeAEPSText(s string) []byte {
	// AE writes each paragraph terminated with \r; normalize \n → \r
	// and ensure a trailing \r.
	s = strings.ReplaceAll(s, "\n", "\r")
	if !strings.HasSuffix(s, "\r") {
		s += "\r"
	}
	var buf bytes.Buffer
	buf.WriteByte('(')
	buf.WriteByte(0xfe)
	buf.WriteByte(0xff)
	for _, r := range s {
		writeUTF16BEEscaped(&buf, r)
	}
	buf.WriteByte(')')
	return buf.Bytes()
}

// writeUTF16BEEscaped emits one rune as 2 or 4 UTF-16BE bytes,
// escaping any byte that is a PostScript string special (\, (, )).
func writeUTF16BEEscaped(buf *bytes.Buffer, r rune) {
	if r > 0xffff {
		// Encode as surrogate pair.
		v := uint32(r) - 0x10000
		hi := uint16(0xD800 | (v >> 10))
		lo := uint16(0xDC00 | (v & 0x3FF))
		writeBEEscaped(buf, hi)
		writeBEEscaped(buf, lo)
		return
	}
	writeBEEscaped(buf, uint16(r))
}

func writeBEEscaped(buf *bytes.Buffer, cu uint16) {
	for _, b := range [2]byte{byte(cu >> 8), byte(cu)} {
		switch b {
		case '\\', '(', ')':
			buf.WriteByte('\\')
		}
		buf.WriteByte(b)
	}
}
