package aep

import "github.com/example/aep-parser/internal/rifx"

// keyframeBackrefs holds the rifx.Chunk reference plus the cached layout
// metadata (offset within ldat.Data, dimensionality, owning comp's TickRate
// and FrameRate) that powers Keyframe.SetTime / SetValue / SetFrameTime
// in-place writes.
//
// Cached fields (offset / dims / tickRate / compFps) duplicate state that
// could be re-derived (offset from Property.lhd3 bpk + keyframe index;
// dims from Property.Components; rates from owning composition). They live
// here because the parser already computed them once per keyframe block.
//
// Lifecycle:
//   - Populated by parseKeyframes when keyframes are decoded from an .aep file.
//   - Also populated by write_property.go::reparseKeyframes after a
//     length-variable splice rewrites the ldat.Data.
//   - Nil for keyframes built outside the parser.
//   - opaque is reserved for future V3 phases (per-keyframe-block ancillary
//     chunks if/when we identify any); currently nil.
type keyframeBackrefs struct {
	ldat     *rifx.Chunk
	offset   int
	dims     int
	tickRate float64
	compFps  float64

	opaque map[rifx.ChunkID]*rifx.Chunk
}
