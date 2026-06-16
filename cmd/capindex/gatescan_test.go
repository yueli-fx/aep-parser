package main

import "testing"

func TestScanTests(t *testing.T) {
	gates, err := scanTests("testdata/gatesample")
	if err != nil {
		t.Fatal(err)
	}
	if skipped, ok := gates["TestRealGate"]; !ok || skipped {
		t.Errorf("TestRealGate: ok=%v skipped=%v (env-conditional skip must not count)", ok, skipped)
	}
	if skipped, ok := gates["TestDisabledGate"]; !ok || !skipped {
		t.Errorf("TestDisabledGate: ok=%v skipped=%v (top-level skip must count)", ok, skipped)
	}
	if skipped, ok := gates["TestPlain"]; !ok || skipped {
		t.Errorf("TestPlain: ok=%v skipped=%v", ok, skipped)
	}
	if _, ok := gates["helper"]; ok {
		t.Error("helper should not be collected (not Test*)")
	}
}
