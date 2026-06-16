package main

import "testing"

func TestParseCapTag(t *testing.T) {
	comment := "AddEffect appends a built-in effect.\n\n" +
		"aep:cap domain=effect tier=stable verify=render-pixel\n" +
		"  gate=TestA,TestB boundary=\"param via SetEffectParam\" alias=\"特效,blur\"\n"
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
