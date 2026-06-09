package serializer

import (
	"fmt"

	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

// propertyBackrefs holds the rifx.Chunk references that power Property's
// length-preserving write paths (SetStaticValue / SetExpression / per-keyframe
// setters). Lives in a separate shard so V3 scene types can mutate logical
// fields without dragging serialization state through every accessor.
//
// Lifecycle:
//   - Populated by parseLeafProperty when a Property is built from a parsed
//     .aep file.
//   - Nil for properties built outside the parser (NewProject builders set
//     it explicitly per their archetype).
//   - opaque captures parsed-but-undecoded sibling chunks under the property's
//     owning tdbs/tdgp container; serializer re-emits them in original order
//     to satisfy CLAUDE.md hard constraint #5 (opaque preservation). Populated
//     by future V3 phases; currently a nil map.
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
	switch x := v.(type) {
	case float64:
		return writeFloat64(b.cdat.Data, 0, x)
	case []float64:
		for i, f := range x {
			if err := writeFloat64(b.cdat.Data, i*8, f); err != nil {
				return err
			}
		}
		return nil
	default:
		return fmt.Errorf("property: unsupported value type %T", v)
	}
}

// SetExpressionEnabled flips the tdb4 @0x78 "disabled" byte (0 = enabled,
// 1 = disabled). length-preserving.
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
		tdb4.Data[0x78] = 0
	} else {
		tdb4.Data[0x78] = 1
	}
	return nil
}

// SetExpression rewrites the JS expression source: replaces the existing Utf8
// chunk's data, inserts a new Utf8 into the tdbs LIST when none existed, or
// removes it entirely when source == "". length-variable.
func (b *propertyBackrefs) SetExpression(source string) error {
	if b.tdbs == nil {
		return fmt.Errorf("property: no tdbs reference (built outside parser?)")
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
		return nil
	}
	if b.exprChunk != nil {
		b.exprChunk.Data = []byte(source)
	} else {
		newUtf8 := &rifx.Chunk{ID: rifx.IDUtf8, Data: []byte(source)}
		b.tdbs.Children = append(b.tdbs.Children, newUtf8)
		b.exprChunk = newUtf8
	}
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
