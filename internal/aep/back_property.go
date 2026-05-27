package aep

import "github.com/example/aep-parser/internal/rifx"

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
//     by future V3 phases; nil-map in Phase 1.
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
