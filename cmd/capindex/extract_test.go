package main

import (
	"strings"
	"testing"
)

func TestExtractEntries(t *testing.T) {
	entries, err := extractEntries("testdata/sample")
	if err != nil {
		t.Fatal(err)
	}
	byName := map[string]Entry{}
	for _, e := range entries {
		byName[e.Symbol] = e
	}

	tg, ok := byName["Tagged"]
	if !ok {
		t.Fatal("Tagged not extracted")
	}
	if !tg.HasCap {
		t.Fatal("Tagged should have a cap")
	}
	if tg.Cap.Domain != "layer-create" || tg.Cap.Verify != "ae-accept" || tg.Cap.Tier != "stable" {
		t.Errorf("Tagged cap wrong: %+v", tg.Cap)
	}
	if len(tg.Cap.Gate) != 1 || tg.Cap.Gate[0] != "TestFoo" {
		t.Errorf("Tagged gate: %v", tg.Cap.Gate)
	}
	if len(tg.Cap.Alias) != 2 || tg.Cap.Alias[0] != "测试" {
		t.Errorf("Tagged alias: %v", tg.Cap.Alias)
	}
	if !strings.HasPrefix(tg.Signature, "func Tagged(name string) error") {
		t.Errorf("Tagged signature: %q", tg.Signature)
	}
	// the //aep:cap directive must be stripped from the summary
	if tg.Summary != "Tagged does a thing worth indexing." {
		t.Errorf("Tagged summary (directive should be stripped): %q", tg.Summary)
	}

	ut, ok := byName["Untagged"]
	if !ok {
		t.Fatal("Untagged not extracted")
	}
	if ut.HasCap {
		t.Errorf("Untagged must not have a cap: %+v", ut.Cap)
	}
}
