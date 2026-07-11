package aep_test

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestRawByteAccessorsReturnDetachedSnapshots(t *testing.T) {
	proj, err := aep.Open("../../test_data/fixtures/re_solidnull.aep")
	if err != nil {
		t.Skipf("re_solidnull.aep not present: %v", err)
	}
	comp := proj.Compositions[0]
	assertDetachedBytes(t, "cdta", comp.CdtaRawBytes)
	assertDetachedBytes(t, "ldta", comp.Layers[0].LdtaRawBytes)
	assertDetachedBytes(t, "sspc", proj.Footage[0].SspcData)
}

func assertDetachedBytes(t *testing.T, name string, get func() []byte) {
	t.Helper()
	first := get()
	if len(first) == 0 {
		t.Fatalf("%s accessor returned no bytes", name)
	}
	original := first[0]
	first[0] ^= 0xff
	second := get()
	if len(second) == 0 || second[0] != original {
		t.Fatalf("%s accessor exposed mutable backing data", name)
	}
}
