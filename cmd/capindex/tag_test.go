package main

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/apidoc"
)

func TestCapFromAnnotation(t *testing.T) {
	a := &apidoc.Annotation{
		Domain: "mask", Stability: "stable", Verify: "ae-accept", Since: "AE2020",
		Gate: []string{"TestAddMask_AEShipGate_AE2020"}, Boundary: "empty parade ok",
		Incident: []string{"add-mask-create-re"}, Alias: []string{"add mask"}, HasTags: true,
	}
	c := capFromAnnotation(a)
	if c.Domain != "mask" || c.Tier != "stable" || c.Verify != "ae-accept" || c.MinVer != "2020" {
		t.Fatalf("cap = %+v", c)
	}
	if len(c.Gate) != 1 || c.Boundary != "empty parade ok" || len(c.Incident) != 1 {
		t.Errorf("cap fields = %+v", c)
	}
	// capFromAnnotation output must still satisfy the existing tier/verify rules.
	if err := validateCap(&c); err != nil {
		t.Errorf("validateCap on @tag-sourced cap: %v", err)
	}
}

func TestParseCapTag(t *testing.T) {
	// Single-line directive (the contract): placement within the comment is
	// irrelevant and following prose is never consumed.
	comment := "AddEffect appends a built-in effect.\n\n" +
		"aep:cap domain=effect tier=stable verify=render-pixel gate=TestA,TestB boundary=\"param via SetEffectParam\" alias=\"特效,blur\"\n" +
		"more prose after the directive that must be ignored.\n"
	c, ok, err := parseCapTag(comment)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !ok {
		t.Fatal("expected ok=true")
	}
	if c.Domain != "effect" || c.Tier != "stable" || c.Verify != "render-pixel" {
		t.Errorf("core fields wrong: %+v", c)
	}
	if len(c.Gate) != 2 || c.Gate[0] != "TestA" || c.Gate[1] != "TestB" {
		t.Errorf("gate: %v", c.Gate)
	}
	if c.Boundary != "param via SetEffectParam" {
		t.Errorf("boundary: %q", c.Boundary)
	}
	if len(c.Alias) != 2 || c.Alias[0] != "特效" || c.Alias[1] != "blur" {
		t.Errorf("alias: %v", c.Alias)
	}
	if c.MinVer != "2020" {
		t.Errorf("minver default: %q", c.MinVer)
	}
}

func TestParseCapTag_None(t *testing.T) {
	c, ok, err := parseCapTag("just a normal comment, no directive\n")
	if c != nil || ok || err != nil {
		t.Errorf("expected (nil,false,nil), got (%v,%v,%v)", c, ok, err)
	}
}

func TestParseCapTag_MinVerOverride(t *testing.T) {
	c, _, err := parseCapTag("x\n\naep:cap domain=project tier=alpha verify=roundtrip minver=2024\n")
	if err != nil {
		t.Fatal(err)
	}
	if c.MinVer != "2024" {
		t.Errorf("minver: %q", c.MinVer)
	}
}
