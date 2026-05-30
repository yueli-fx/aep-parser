package aep

import "github.com/example/aep-parser/internal/rifx"

// projectBackrefs holds the rifx.Chunk references that power Project's
// length-preserving write paths (BPC + all P1 project-settings setters,
// CMS / GPU / expression engine writers) plus the parser-owned root chunk
// and the cached root Fold LIST used by structural mutation.
//
// Lifecycle:
//   - Populated by parseProject when a Project is built from a parsed .aep
//     file.
//   - For projects built via NewProject the template-loader populates these
//     fields from the bundled .aep template (parsed re-parse closed loop).
//   - opaque is reserved for future V3 phases that need to round-trip
//     unrecognized root-level chunks (per CLAUDE.md hard constraint #5);
//     currently nil.
type projectBackrefs struct {
	// root keeps the original RIFX chunk tree so callers can mutate leaf
	// chunks (e.g. Footage.SetPath) and re-serialize via WriteAEP.
	root *rifx.Chunk

	// rootFold is the cached root Fold LIST reference; derived cache, never
	// owned (see Invariants #8). Used by structural mutation (NewComposition
	// / future NewFootage etc.).
	rootFold *rifx.Chunk

	// Project header chunks. nhed @0x0F + nnhd @0x18 both hold BPC enum
	// (0=8 / 1=16 / 2=32). SetBitsPerChannel writes both for consistency.
	nhedChunk *rifx.Chunk
	nnhdChunk *rifx.Chunk

	// Project-level single-field setting chunks (P1 1D, py-aep parity).
	// Captured by parseProject when present; mutated by Set* methods.
	// All exist as direct root children — see project_settings.go.
	acerChunk *rifx.Chunk // 1B bool — compensate_for_scene_referred_profiles
	adfrChunk *rifx.Chunk // 8B f64 BE — audio_sample_rate
	dwgaChunk *rifx.Chunk // 1-4B — byte 0 = working_gamma selector
	gpugUtf8  *rifx.Chunk // Utf8 inside gpuG LIST — gpu_accel_type (UUID)
	exenUtf8  *rifx.Chunk // Utf8 inside ExEn LIST — expression_engine
	cmsUtf8   *rifx.Chunk // Utf8 — CMS settings JSON (AE 24+)

	opaque map[rifx.ChunkID]*rifx.Chunk
}
