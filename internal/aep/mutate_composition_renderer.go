package aep

import (
	"fmt"
	"sort"
	"strings"

	"github.com/example/aep-parser/internal/scene"
)

// mutate_composition_renderer.go — SetRenderer switches a comp's 3D rendering
// engine. This is a STRUCTURAL write: the renderer lives in the comp's
// `LIST:PRin` sibling as two chunks —
//
//   - prin: fixed 104B. @0..3 constant, match-name NUL-padded @4 (field
//     [4,52)), display name NUL-padded @0x34 (field [52,96)), @96..103
//     constant trailer. Switching renderers rewrites only the two name
//     fields in place — length-preserving, opaque-preserving (the constant
//     regions, which may carry AE-version-specific bytes, are untouched).
//   - prda: renderer-specific options, VARIABLE length (12/52/20/16B for the
//     four engines). Switching renderers replaces it wholesale with the
//     target engine's default-option template → parent PRin LIST size
//     changes → structural. WriteAEP recomputes all sizes from Data.
//
// Switching the renderer resets its options to defaults (same as AE's own
// behavior when you change the renderer in Composition Settings).
//
// Atomic mutation: snapshot prin.Data / prda.Data / c.Renderer / warnings;
// on any new parser warning, roll all back and return them as an error.
//
// prin/prda templates were RE'd from py-aep's renderer_{classic_3d,advanced_3d,
// cinema_4d,ray_traced}.aep fixtures (single comp each). The binary match_name
// is the stable engine identity; the prda template is keyed by it (Escher 12B /
// Calder 52B / Ernst 20B / Picasso 16B).
//
// Ship-gate: AE 2025 green for all 4 engines (Calder/Ernst exact, legacy
// Escher/Picasso auto-promoted to Advanced 3D on load); AE 2020 green for the
// engines it exposes (Ernst exact, Escher→"ADBE Advanced 3d"). See
// composition_renderer_shipgate_test.go.
//
// Concurrency: like all mutate paths these touch shared chunk bytes; callers
// serialize their own access (see incidents/concurrency-unsafe-shared-chunk-bytes).

// prin layout (104B), per py-aep binary/misc_chunks.py PrinChunk:
//
//	[0,4)    reserved (constant 00000000)
//	[4,52)   match_name   — ASCII NUL-padded, 48B
//	[52,100) display_name — ASCII NUL-padded, 48B (cosmetic; AE re-derives)
//	[100,103) reserved
//	[103]    end marker 0x01
const (
	prinMatchNameOff = 4    // match_name field start
	prinDisplayOff   = 0x34 // display_name field start (= 52)
	prinDisplayEnd   = 0x64 // display_name field end (= 100); [100,104) reserved + end marker
	prinSize         = 104
)

// rendererExtendscriptToBinary maps the ExtendScript module name (what
// CompItem.renderer exposes) to the binary prin match_name (what's stored on
// disk). Mirrors py-aep composition.py _RENDERER_EXTENDSCRIPT_TO_BINARY. Only
// "ADBE Advanced 3d" differs from its binary name ("ADBE Escher"); the other
// three are identical in both namespaces. SetRenderer accepts either name.
var rendererExtendscriptToBinary = map[string]string{
	"ADBE Advanced 3d": "ADBE Escher",
}

// normalizeRendererName resolves an ExtendScript module name to its binary
// match_name; binary names (and unknowns) pass through unchanged.
func normalizeRendererName(name string) string {
	if bin, ok := rendererExtendscriptToBinary[name]; ok {
		return bin
	}
	return name
}

// rendererTemplate carries the per-engine bytes SetRenderer emits: the
// localized display name written into prin's display field, and the full
// prda chunk payload.
type rendererTemplate struct {
	display string
	prda    []byte
}

// rendererTemplates maps a renderer match-name to its prin display name and
// prda default-option payload. RE'd from py-aep renderer fixtures.
var rendererTemplates = map[string]rendererTemplate{
	"ADBE Escher": { // Classic 3D
		display: "Classic 3D",
		prda:    []byte{0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0},
	},
	"ADBE Calder": { // Advanced 3D
		display: "Advanced 3D",
		prda: []byte{
			0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00, 0x00, 0x01,
			0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x80, 0x3f,
			0x00, 0x00, 0x80, 0x3f, 0x00, 0x00, 0x80, 0x3f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00,
		},
	},
	"ADBE Ernst": { // Cinema 4D
		display: "Cinema 4D",
		prda:    []byte{0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 0x19, 0, 0, 0, 1, 0, 0, 0, 0},
	},
	"ADBE Picasso": { // Ray-traced 3D
		display: "Ray-traced 3D",
		prda:    []byte{0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 3, 0, 0, 0, 1},
	},
}

// knownRenderers returns the supported match-names, sorted, for error text.
func knownRenderers() string {
	names := make([]string, 0, len(rendererTemplates))
	for k := range rendererTemplates {
		names = append(names, k)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}

// SetRenderer switches the composition's 3D rendering engine. The name may be
// either a binary prin match_name ("ADBE Escher" / "ADBE Calder" /
// "ADBE Ernst" / "ADBE Picasso") or an ExtendScript module name
// ("ADBE Advanced 3d" → "ADBE Escher"); it is normalized to the binary name.
// The binary match_name + display name are rewritten in the prin chunk
// (length-preserving) and the prda chunk is replaced with the engine's default
// options (structural). Returns an error for an unknown renderer, a comp built
// outside the parser (no prin/prda back-ref), a comp whose prin is not the
// expected 104 bytes, or if the mutation surfaces a parser warning (rolled back).
//
// Which engines a given AE version actually exposes differs (AE 2020:
// Escher/Ernst + a Standard variant; AE 2025: Calder/Ernst + Picasso; AE 2025
// auto-promotes legacy Escher/Picasso to Advanced 3D on load). The binary
// match_name is the stable engine identity — see
// sketches/2026-06-01-renderer-write-re-findings.md.
//
// Ship-gated: AE 2025 (4/4) + AE 2020 (Ernst + Escher) green.
//
// Free function (not a method) so the rollback path can reach the concrete
// comp back-ref (prin/prda chunks) after the M8 split; the aep facade
// re-exports it. BREAKING vs the former Composition.SetRenderer method form.
func SetRenderer(c *Composition, name string) error {
	cb := compositionBack(c)
	if cb == nil {
		return fmt.Errorf("SetRenderer: comp %q has no prin/prda back-ref (built outside parser, or has no PRin LIST)", c.Name)
	}
	matchName := normalizeRendererName(name)
	oldRenderer := c.Renderer
	proj := scene.CompositionProj(c)
	oldWarningsLen := 0
	if proj != nil {
		oldWarningsLen = len(proj.Warnings)
	}

	// Snapshot prin/prda for rollback.
	var oldPrin, oldPrda []byte
	if cb.prinChunk != nil {
		oldPrin = append([]byte(nil), cb.prinChunk.Data...)
	}
	if cb.prdaChunk != nil {
		oldPrda = append([]byte(nil), cb.prdaChunk.Data...)
	}

	if err := cb.SetRenderer(name); err != nil {
		return err
	}
	c.Renderer = matchName

	// Warnings-as-failure. No re-parse here, so warnings won't grow in
	// practice — defensive rollback path matching the V2.1 mutate pattern.
	if proj != nil && len(proj.Warnings) > oldWarningsLen {
		c.Renderer = oldRenderer
		if cb.prinChunk != nil && oldPrin != nil {
			cb.prinChunk.Data = oldPrin
		}
		if cb.prdaChunk != nil && oldPrda != nil {
			cb.prdaChunk.Data = oldPrda
		}
		newWarnings := append([]string(nil), proj.Warnings[oldWarningsLen:]...)
		proj.Warnings = proj.Warnings[:oldWarningsLen]
		return fmt.Errorf("SetRenderer: produced %d parser warning(s), rolled back: %v", len(newWarnings), newWarnings)
	}

	return nil
}

// writePrinField writes s as ASCII into field (zeroing it first, NUL-padded).
// Bytes beyond len(field) are dropped; non-ASCII runes (>= 0x80) are skipped.
func writePrinField(field []byte, s string) {
	for i := range field {
		field[i] = 0
	}
	n := 0
	for _, r := range s {
		if n >= len(field) {
			break
		}
		if r < 0x80 {
			field[n] = byte(r)
			n++
		}
	}
}
