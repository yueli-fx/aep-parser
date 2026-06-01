package aep

import (
	"fmt"
	"sort"
	"strings"
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
// Alpha: NOT yet double-version ship-gated. prin/prda templates were RE'd
// from py-aep's renderer_{classic_3d,advanced_3d,cinema_4d,ray_traced}.aep
// fixtures (single comp each). AE 2020 + AE 2025 acceptance pending.
//
// Concurrency: like all mutate paths these touch shared chunk bytes; callers
// serialize their own access (see incidents/concurrency-unsafe-shared-chunk-bytes).

const (
	prinMatchNameOff = 4    // match-name field start
	prinDisplayOff   = 0x34 // display-name field start (= 52)
	prinDisplayEnd   = 0x60 // display-name field end (= 96); @96..103 constant trailer
	prinSize         = 104
)

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

// SetRenderer switches the composition's 3D rendering engine to the given
// internal match-name (one of "ADBE Escher" / "ADBE Calder" / "ADBE Ernst" /
// "ADBE Picasso"). The match-name + localized display name are rewritten in
// the prin chunk (length-preserving) and the prda chunk is replaced with the
// engine's default options (structural). Returns an error for an unknown
// renderer, a comp built outside the parser (no prin/prda back-ref), a comp
// whose prin is not the expected 104 bytes, or if the mutation surfaces a
// parser warning (rolled back).
//
// Alpha — not yet AE-ship-gated.
func (c *Composition) SetRenderer(matchName string) error {
	tmpl, ok := rendererTemplates[matchName]
	if !ok {
		return fmt.Errorf("SetRenderer: unknown renderer %q (known: %s)", matchName, knownRenderers())
	}
	if c.back == nil || c.back.prinChunk == nil || c.back.prdaChunk == nil {
		return fmt.Errorf("SetRenderer: comp %q has no prin/prda back-ref (built outside parser, or has no PRin LIST)", c.Name)
	}
	prin := c.back.prinChunk
	prda := c.back.prdaChunk
	if len(prin.Data) != prinSize {
		return fmt.Errorf("SetRenderer: comp %q prin is %d bytes, expected %d", c.Name, len(prin.Data), prinSize)
	}

	// Snapshot for rollback.
	oldPrin := append([]byte(nil), prin.Data...)
	oldPrda := append([]byte(nil), prda.Data...)
	oldRenderer := c.Renderer
	oldWarningsLen := 0
	if c.proj != nil {
		oldWarningsLen = len(c.proj.Warnings)
	}

	// Apply: rewrite prin's two name fields in place (preserve constants),
	// replace prda wholesale.
	writePrinField(prin.Data[prinMatchNameOff:prinDisplayOff], matchName)
	writePrinField(prin.Data[prinDisplayOff:prinDisplayEnd], tmpl.display)
	prda.Data = append([]byte(nil), tmpl.prda...)
	c.Renderer = matchName

	// Warnings-as-failure. No re-parse here, so warnings won't grow in
	// practice — defensive rollback path matching the V2.1 mutate pattern.
	if c.proj != nil && len(c.proj.Warnings) > oldWarningsLen {
		prin.Data = oldPrin
		prda.Data = oldPrda
		c.Renderer = oldRenderer
		newWarnings := append([]string(nil), c.proj.Warnings[oldWarningsLen:]...)
		c.proj.Warnings = c.proj.Warnings[:oldWarningsLen]
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
