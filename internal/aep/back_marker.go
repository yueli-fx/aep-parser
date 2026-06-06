package aep

import "github.com/example/aep-parser/internal/rifx"

// markerBackrefs holds the per-marker rifx.Chunk references that power a
// parsed Marker's write paths (SetTime / SetDuration / SetLabel / the five
// SetComment-family text setters). Lives in a separate shard so the Marker
// scene type stays chunk-free.
//
// Lifecycle:
//   - Populated by parseMarkers (and AddMarker) when a Marker is built from
//     a parsed .aep file.
//   - nil for markers built outside the parser; every Set* refuses with an
//     error when the reference it needs is missing.
type markerBackrefs struct {
	// ldat is the shared keyframe-block chunk; this marker's 16-byte block
	// starts at Marker.ldatOffset. SetTime patches its time slot in place.
	ldat *rifx.Chunk
	// nmHd is the per-marker NmHd chunk (Duration @0x08 / Label @0x10).
	nmHd *rifx.Chunk
	// nmrd is the per-marker Nmrd LIST holding the Utf8 text children.
	nmrd *rifx.Chunk
}

// markerList is the internal container behind one "ADBE Marker" set (a comp's
// or a layer's markers). It holds the chunk references the structural ops
// splice — the keyframe count (lhd3), the keyframe blocks (ldat), and the
// Nmrd-bearing mrky LIST — plus owner, a pointer to the public Markers field
// these markers live in, so Remove/Add can keep that slice in sync. All
// markers in one set share a single *markerList.
type markerList struct {
	lhd3  *rifx.Chunk // kfl count chunk: count @0x08, bpk @0x10 = 16
	ldat  *rifx.Chunk // keyframe blocks: count × 16 bytes
	mrky  *rifx.Chunk // Nmrd container (nil when the set has no mrky branch)
	owner *[]*Marker  // the public Composition.Markers / Layer.Markers field
}
