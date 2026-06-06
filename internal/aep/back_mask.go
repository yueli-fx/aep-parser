package aep

import "github.com/example/aep-parser/internal/rifx"

// maskBackrefs holds the rifx.Chunk references that power a parsed Mask's
// length-preserving setters (SetMode / SetInverted / SetColor / SetLocked /
// SetMaskMotionBlur / SetClosed). Lives in a separate shard so the Mask scene
// type stays chunk-free.
//
// Lifecycle:
//   - Populated by parseMasks when a Mask is built from a parsed .aep file.
//   - nil for masks built outside the parser; every Set* refuses with an
//     error when the chunk it needs is missing.
type maskBackrefs struct {
	// mkif is the mask info chunk (Mode @0x04 / Inverted @0x00 / Locked @0x01
	// / MotionBlur @0x02 / Color @0x2D-0x2F).
	mkif *rifx.Chunk
	// shph is the first shph path-header chunk (Closed @0x14 write — only
	// meaningful for static masks).
	shph *rifx.Chunk
}
