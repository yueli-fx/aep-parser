package aep

import "github.com/example/aep-parser/internal/rifx"

// Project-level chunk navigation helpers. These read the root RIFX tree
// directly; the scene_ accessors that expose them (Project.EffectNames, …)
// reach the root via the unexported back field so they stay rifx-free.

// findRootListByType walks root's direct children looking for the
// first LIST chunk with the given formType.
func findRootListByType(root *rifx.Chunk, formType rifx.ChunkID) *rifx.Chunk {
	for _, ch := range root.Children {
		if ch.IsList() && ch.FormType == formType {
			return ch
		}
	}
	return nil
}

// effectNamesFromRoot collects effect match-names from the root-level
// `Pefl` LIST → `pjef` Utf8 entries. Returns nil when the project has no
// Pefl LIST (no effects applied anywhere, or builder-synthesized).
func effectNamesFromRoot(root *rifx.Chunk) []string {
	pefl := findRootListByType(root, rifx.IDPefl)
	if pefl == nil {
		return nil
	}
	var out []string
	for _, ch := range pefl.Children {
		if ch.ID == rifx.IDPjef {
			out = append(out, ch.Text())
		}
	}
	return out
}
