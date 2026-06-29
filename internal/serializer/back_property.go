package serializer

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/rifx"
	"github.com/yueli-fx/aep-parser/internal/scene"
)

// propertyBackrefs holds the rifx.Chunk references that power Property's
// length-preserving write paths (SetStaticValue / SetExpression / per-keyframe
// setters). Lives in a separate shard so scene types can mutate logical
// fields without dragging serialization state through every accessor.
//
// Lifecycle:
//   - Populated by parseLeafProperty when a Property is built from a parsed
//     .aep file.
//   - Nil for properties built outside the parser (NewProject builders set
//     it explicitly per their archetype).
//   - opaque captures parsed-but-undecoded sibling chunks under the property's
//     owning tdbs/tdgp container; serializer re-emits them in original order
//     to satisfy the opaque-preservation invariant. Populated
//     by a future phase; currently a nil map.
type propertyBackrefs struct {
	// tdbs is the property's owning tdbs LIST — used by SetExpression
	// to insert/remove the Utf8 chunk holding the JS source.
	tdbs *rifx.Chunk
	// tdb4 is the property metadata chunk under tdbs (124 bytes); holds
	// the dimension, type flags, spatial / animated / no_value / color /
	// integer / vector bits. Populated by parseLeafProperty. Used by
	// tdb4 flag readers (IsSpatial, IsAnimated, etc.).
	tdb4 *rifx.Chunk
	// cdat is the current/static value chunk (no keyframes).
	cdat *rifx.Chunk
	// cdatLE marks cdat as little-endian (AE stores a 3D layer's Orientation
	// value LE inside its otst wrapper — see parseOrientationProperty). The
	// default (big-endian) holds for every other property. SetStaticValue keys
	// its write endianness off this so the value round-trips and AE accepts it.
	cdatLE bool
	// otda is the orientation "default value" chunk (24 B = 3 × f64, BIG-endian)
	// living under the otst's otky sibling LIST. For a STATIC 3D orientation AE
	// reads its current value from HERE, not from cdat — so SetStaticValue must
	// mirror the value into otda (BE) or AE renders 0 despite a clean cdat write
	// (RE'd 2026-06-17 by byte-diffing an AE-authored golden). Nil for non-
	// orientation properties.
	otda *rifx.Chunk
	// ldat is the keyframe stream chunk (with keyframes).
	ldat *rifx.Chunk
	// lhd3 is the keyframe-list header chunk (count @0x08, bpk @0x10).
	lhd3 *rifx.Chunk
	// tdsb is the property subprop flags chunk (4 bytes); holds
	// locked_ratio (byte 2 bit 4), dimensions_separated (byte 3 bit 1),
	// enabled (byte 3 bit 0), roto_bezier (byte 0 bit 0).
	tdsb *rifx.Chunk
	// tdum / tduM are the property min/max value chunks under tdbs.
	// Populated by parseLeafProperty; nil when absent. Decoded by
	// MinValue() / MaxValue(). Layout depends on tdb4 type flags:
	// color → 4×f32, integer → 1×u32, otherwise N×f64.
	tdum *rifx.Chunk
	tduM *rifx.Chunk
	// exprChunk is the Utf8 chunk holding the expression JS source.
	// Nil when the property has no expression; SetExpression creates
	// or removes it as needed.
	exprChunk *rifx.Chunk

	// bytesPerKF mirrors the lhd3 @0x10 keyframe stride (decoded once during
	// parse, cached so write paths don't re-read the header on every setter).
	bytesPerKF int

	opaque map[rifx.ChunkID]*rifx.Chunk
}

var _ PropertyWriter = (*propertyBackrefs)(nil)

// propertyBack returns the concrete back-refs behind a Property's writer
// interface for serializer-stage (parse_/mutate_/write_) raw chunk access
// (keyframe stream ops, separate-dimensions splice, tdb4/tdsb flag readers,
// parse wiring). Returns nil when the property was built outside the parser.
func propertyBack(p *Property) *propertyBackrefs {
	if pb, ok := scene.PropertyBack(p).(*propertyBackrefs); ok {
		return pb
	}
	return nil
}

func (b *propertyBackrefs) Tdb4Byte(off int) (byte, bool) {
	if b == nil || b.tdb4 == nil || len(b.tdb4.Data) <= off {
		return 0, false
	}
	return b.tdb4.Data[off], true
}

func (b *propertyBackrefs) TdsbByte(off int) (byte, bool) {
	if b == nil || b.tdsb == nil || len(b.tdsb.Data) <= off {
		return 0, false
	}
	return b.tdsb.Data[off], true
}

func (b *propertyBackrefs) MinValueBytes() []byte {
	if b == nil || b.tdum == nil {
		return nil
	}
	return b.tdum.Data
}

func (b *propertyBackrefs) MaxValueBytes() []byte {
	if b == nil || b.tduM == nil {
		return nil
	}
	return b.tduM.Data
}

// SetStaticValue rewrites the property's constant value in the cdat chunk. The
// caller (Property delegate) validates the value's component count against
// Property.Components and syncs the scene field; this only writes the bytes.
func (b *propertyBackrefs) SetStaticValue(v any) error {
	if b.cdat == nil {
		return fmt.Errorf("property: no static-value chunk (has keyframes?)")
	}
	wf := writeFloat64
	if b.cdatLE {
		wf = writeFloat64LE
	}
	switch x := v.(type) {
	case float64:
		return wf(b.cdat.Data, 0, x)
	case []float64:
		for i, f := range x {
			if err := wf(b.cdat.Data, i*8, f); err != nil {
				return err
			}
			// Orientation: AE's authoritative static value lives in otda (BE),
			// not cdat — mirror it or AE renders 0 (see otda field doc).
			if b.otda != nil && i*8+8 <= len(b.otda.Data) {
				if err := writeFloat64(b.otda.Data, i*8, f); err != nil {
					return err
				}
			}
		}
		return nil
	default:
		return fmt.Errorf("property: unsupported value type %T", v)
	}
}

// SetExpressionEnabled writes the tdb4 @0x77 disabled byte (0 = AE
// evaluates the expression, 1 = expression kept but off). The neighbouring
// @0x78 is NOT the enabled flag — it is the has-expression marker owned by
// SetExpression (RE'd against AE-2025-native enabled/disabled fixture pair,
// expr_re 2026-06-12: enabled = 00 01, disabled = 01 01 at @0x77/@0x78).
// length-preserving.
func (b *propertyBackrefs) SetExpressionEnabled(enabled bool) error {
	if b.tdbs == nil {
		return fmt.Errorf("property: no tdbs reference")
	}
	tdb4 := findTdb4Chunk(b.tdbs)
	if tdb4 == nil {
		return fmt.Errorf("property: tdb4 chunk missing")
	}
	if len(tdb4.Data) <= 0x78 {
		return fmt.Errorf("property: tdb4 too short (%d bytes) for expressionEnabled write", len(tdb4.Data))
	}
	if enabled {
		tdb4.Data[0x77] = 0
	} else {
		tdb4.Data[0x77] = 1
	}
	return nil
}

// SetExpression rewrites the JS expression source: replaces the existing Utf8
// chunk's data, inserts a new Utf8 into the tdbs LIST when none existed, or
// removes it entirely when source == "". Keeps the tdb4 @0x78 has-expression
// marker in sync — without it AE treats the Utf8 as absent and silently
// drops the expression text on open (see SetExpressionEnabled for the
// @0x77/@0x78 pair semantics). length-variable.
func (b *propertyBackrefs) SetExpression(source string) error {
	if b.tdbs == nil {
		return fmt.Errorf("property: no tdbs reference (built outside parser?)")
	}
	tdb4 := findTdb4Chunk(b.tdbs)
	setMarker := func(v byte) {
		if tdb4 != nil && len(tdb4.Data) > 0x78 {
			tdb4.Data[0x78] = v
		}
	}
	if source == "" {
		if b.exprChunk != nil {
			out := b.tdbs.Children[:0]
			for _, ch := range b.tdbs.Children {
				if ch == b.exprChunk {
					continue
				}
				out = append(out, ch)
			}
			b.tdbs.Children = out
			b.exprChunk = nil
		}
		setMarker(0)
		return nil
	}
	if b.exprChunk != nil {
		b.exprChunk.Data = []byte(source)
	} else {
		newUtf8 := &rifx.Chunk{ID: rifx.IDUtf8, Data: []byte(source)}
		// The expression Utf8 belongs immediately AFTER the value cdat (before any
		// tdum/tduM min/max range chunks) — AE's canonical order. A plain
		// append() lands it after tdum/tduM, which AE 2020 reads as a corrupt
		// stream and skips the layer on open (a materialized effect param carries
		// tdum/tduM, where a bare Transform scalar does not, so append only
		// happened to work before). Insert right after the last cdat (or the
		// tdb4 when no cdat), falling back to append.
		insertAt := len(b.tdbs.Children)
		for i, ch := range b.tdbs.Children {
			if id := string(ch.ID[:]); ch.ID == rifx.IDCdat || id == "tdb4" || id == "Tdb4" {
				insertAt = i + 1
			}
		}
		kids := make([]*rifx.Chunk, 0, len(b.tdbs.Children)+1)
		kids = append(kids, b.tdbs.Children[:insertAt]...)
		kids = append(kids, newUtf8)
		kids = append(kids, b.tdbs.Children[insertAt:]...)
		b.tdbs.Children = kids
		b.exprChunk = newUtf8
	}
	setMarker(1)
	return nil
}

// SetLockedRatio flips the tdsb @0x02 bit 4 locked-ratio flag. length-preserving.
func (b *propertyBackrefs) SetLockedRatio(v bool) error {
	if b.tdsb == nil {
		return fmt.Errorf("property has no tdsb chunk (cannot set LockedRatio)")
	}
	if len(b.tdsb.Data) < 3 {
		return fmt.Errorf("tdsb chunk too short (len=%d)", len(b.tdsb.Data))
	}
	if v {
		b.tdsb.Data[0x02] |= 1 << 4
	} else {
		b.tdsb.Data[0x02] &^= 1 << 4
	}
	return nil
}
