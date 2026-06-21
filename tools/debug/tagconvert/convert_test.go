package main

import (
	"strings"
	"testing"
)

func TestConvertSource_Basic(t *testing.T) {
	src := `package aep

// AddThing adds a thing to the project. It does the needful.
//
//aep:cap domain=structural tier=stable verify=roundtrip alias="add thing,加东西"
func AddThing(p *Project, x int) (*Thing, error) { return nil, nil }
`
	out, err := convertSource([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	for _, want := range []string{
		"// @summary",               // a summary tag was emitted
		"// @param   p ",            // each signature param gets a placeholder
		"// @param   x ",
		"// @returns ",              // non-error return → @returns placeholder
		"// @domain     structural", // machine fields carried over
		"// @stability  stable",
		"// @verify     roundtrip",
		"// @since      AE2020", // minver default 2020 → AE2020
		"// @alias      add thing,加东西",
		"TODO", // placeholders present for manual fill
	} {
		if !strings.Contains(got, want) {
			t.Errorf("convert missing %q in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "aep:cap") {
		t.Errorf("legacy aep:cap not removed:\n%s", got)
	}
	// The rewritten source must still parse.
	if !strings.Contains(got, "func AddThing(p *Project, x int) (*Thing, error)") {
		t.Errorf("signature mangled:\n%s", got)
	}
}

func TestConvertSource_ChineseBoundaryFlagged(t *testing.T) {
	src := `package aep

// NewThing builds a thing.
//
//aep:cap domain=project tier=stable verify=ae-accept gate=TestX boundary="零参=默认值" minver=2025
func NewThing() *Thing { return nil }
`
	out, err := convertSource([]byte(src))
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	if !strings.Contains(got, "@boundary   TODO translate: 零参=默认值") {
		t.Errorf("non-ASCII boundary should be flagged for translation:\n%s", got)
	}
	if !strings.Contains(got, "@since      AE2025") {
		t.Errorf("minver=2025 → @since AE2025:\n%s", got)
	}
	if !strings.Contains(got, "@gate       TestX") {
		t.Errorf("gate carried over:\n%s", got)
	}
}
