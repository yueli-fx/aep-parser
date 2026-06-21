package main

import "github.com/example/aep-parser/internal/apidoc"

// Cap is a parsed `aep:cap` capability directive attached to an exported
// internal/aep symbol. It is the single source of truth for a capability's
// maturity/verification status (the docgen pipeline owns signature + prose).
type Cap struct {
	Domain   string   `json:"domain"`   // required bucket: layer-create, effect, shape, ...
	Tier     string   `json:"tier"`     // required: stable|alpha|planned|missing|negative
	Verify   string   `json:"verify"`   // required: none|roundtrip|ae-accept|render-pixel
	MinVer   string   `json:"minver"`   // AE version the capability requires (default "2020")
	Gate     []string `json:"gate"`     // ship-gate test function names
	Boundary string   `json:"boundary"` // known scale/combination limits
	Incident []string `json:"incident"` // incident slugs (gotcha sources)
	Alias    []string `json:"alias"`    // search keywords (may include CJK)
}

// Entry is one row of the generated capability index: an exported symbol plus
// its parsed Cap. HasCap is false for symbols that carry no aep:cap directive.
type Entry struct {
	Symbol    string `json:"symbol"`
	Kind      string `json:"kind"` // func|method|type|const
	Recv      string `json:"recv,omitempty"`
	Signature string `json:"signature"`
	Summary   string `json:"summary"`
	Example   string `json:"example,omitempty"`
	HasCap    bool   `json:"has_cap"`
	Cap       Cap    `json:"cap"`

	Pkg      string `json:"-"` // source package basename (aep|scene|codec); drives public-surface gating
	parseErr error  // unexported: malformed aep:cap directive, surfaced by validateEntries

	Pos             string             `json:"-"` // "file:line" for --validate messages
	Params          []string           `json:"-"` // non-receiver parameter names
	ReturnsNonError bool               `json:"-"` // signature returns a non-error value
	Ann             *apidoc.Annotation `json:"-"` // parsed @tag block (nil for legacy aep:cap)
}
