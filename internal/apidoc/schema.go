// Package apidoc is the single source of truth for the exported-API annotation
// vocabulary. Both cmd/capindex and cmd/docgen import it; no enum or field rule
// is duplicated anywhere else.
//
// THIS FILE IS THE ONLY PLACE TO EDIT THE ANNOTATION VOCABULARY. To add a domain,
// append to Domains; to retire a verify value, remove it from Verifies. capindex,
// docgen, and `capindex --validate` all read these slices, so a change takes
// effect everywhere at once.
package apidoc

// Domains is the frozen set of capability buckets (16). "meta" is the lane for
// getters/readers/aliases/enum consts — present in the API but not verified
// capabilities.
var Domains = []string{
	"shape", "layer-set", "layer-create", "text", "mask", "effect",
	"gradient", "keyframe", "comp", "project", "render-queue", "structural",
	"eg", "expr", "io", "meta",
}

// Stabilities is the frozen set of API-maturity values.
var Stabilities = []string{"stable", "alpha"}

// Verifies is the frozen set of evidence levels.
var Verifies = []string{"ae-accept", "render-pixel", "roundtrip", "none"}

// SinceVersions is the frozen set of minimum-AE-version values (format AE<year>).
var SinceVersions = []string{"AE2020", "AE2025"}

func has(set []string, v string) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
}
