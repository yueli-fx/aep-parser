package apidoc

import "testing"

func TestParse_FullBlock(t *testing.T) {
	raw := `@summary    Add a vector mask to a layer
@description Appends a closed Bezier mask to the layer's Mask Parade.
  The outline is in layer-pixel space.
@param      layer  owning layer (from a parsed project)
@param      path   outline in layer-pixel space (closed Bezier)
@returns    the created *Mask
@domain     mask
@stability  stable
@verify     ae-accept
@gate       TestAddMask_AEShipGate_AE2020,TestAddMask_AEShipGate_AE2025
@since      AE2020
@boundary   removing the last mask leaves an empty parade
@incident   add-mask-create-re
@alias      remove mask,删蒙版
`
	a, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !a.HasTags {
		t.Fatal("HasTags = false, want true")
	}
	if a.Summary != "Add a vector mask to a layer" {
		t.Errorf("Summary = %q", a.Summary)
	}
	if want := "Appends a closed Bezier mask to the layer's Mask Parade.\nThe outline is in layer-pixel space."; a.Description != want {
		t.Errorf("Description = %q, want %q", a.Description, want)
	}
	if len(a.Params) != 2 || a.Params[0].Name != "layer" || a.Params[1].Name != "path" {
		t.Fatalf("Params = %+v", a.Params)
	}
	if a.Params[1].Desc != "outline in layer-pixel space (closed Bezier)" {
		t.Errorf("Params[1].Desc = %q", a.Params[1].Desc)
	}
	if a.Returns != "the created *Mask" {
		t.Errorf("Returns = %q", a.Returns)
	}
	if a.Domain != "mask" || a.Stability != "stable" || a.Verify != "ae-accept" || a.Since != "AE2020" {
		t.Errorf("caps = %q/%q/%q/%q", a.Domain, a.Stability, a.Verify, a.Since)
	}
	if len(a.Gate) != 2 || a.Gate[0] != "TestAddMask_AEShipGate_AE2020" {
		t.Errorf("Gate = %v", a.Gate)
	}
	if len(a.Incident) != 1 || a.Incident[0] != "add-mask-create-re" {
		t.Errorf("Incident = %v", a.Incident)
	}
	if len(a.Alias) != 2 || a.Alias[1] != "删蒙版" {
		t.Errorf("Alias = %v", a.Alias)
	}
}

func TestParse_DescriptionParagraphs(t *testing.T) {
	raw := "@summary do a thing\n@description First paragraph line one.\n  line two of paragraph one.\n\n  Second paragraph after a blank line.\n@domain io"
	a, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	want := "First paragraph line one.\nline two of paragraph one.\n\nSecond paragraph after a blank line."
	if a.Description != want {
		t.Errorf("Description = %q, want %q", a.Description, want)
	}
}

func TestParse_NoTags(t *testing.T) {
	a, err := Parse("AddMask adds a vector mask to a layer.\nNo @ lines here.")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if a.HasTags {
		t.Error("HasTags = true, want false (no @tag lines)")
	}
}

func TestParse_UnknownTag(t *testing.T) {
	// An unknown @tag is a typo only inside a converted block (one with a known
	// tag); there it must error.
	if _, err := Parse("@summary do a thing\n@bogus something"); err == nil {
		t.Fatal("Parse: want error for unknown @tag in a converted block")
	}
}

func TestParse_LegacyChunkOffsetProse(t *testing.T) {
	// A legacy (un-converted) comment whose wrapped lines start with chunk-offset
	// refs must NOT be mistaken for @tags, and must not be flagged as converted.
	raw := "SetColor sets the fill color at\n@0x2D/@0x2E/@0x2F. AE keeps alpha at 0xFF.\n\naep:cap domain=mask tier=stable verify=roundtrip"
	a, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if a.HasTags {
		t.Error("HasTags = true, want false (chunk-offset prose is not a @tag)")
	}
}

func TestHasLegacyCap(t *testing.T) {
	if !HasLegacyCap("Open parses an .aep file.\n\naep:cap domain=meta tier=stable verify=roundtrip") {
		t.Error("HasLegacyCap = false, want true")
	}
	if HasLegacyCap("@summary Open an aep file\n@domain io") {
		t.Error("HasLegacyCap = true, want false")
	}
}
