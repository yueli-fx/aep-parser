package serializer

import (
	"path/filepath"
	"testing"
)

func TestStaticValuePreservationDigestRedactsOnlyClaimedBytes(t *testing.T) {
	project, err := Open(filepath.Join("..", "..", "test_data", "fixtures", "transform_unseparated.aep"))
	if err != nil {
		t.Fatal(err)
	}
	if len(project.Compositions) == 0 || len(project.Compositions[0].Layers) == 0 {
		t.Fatal("fixture has no layer")
	}
	var target *Property
	for _, property := range project.Compositions[0].Layers[0].Properties {
		if property.MatchName == "ADBE Time Remapping" {
			target = property
			break
		}
	}
	if target == nil {
		t.Fatal("time remapping property missing")
	}
	before, err := StaticValuePreservationDigest(project, []*Property{target})
	if err != nil {
		t.Fatal(err)
	}
	if err := target.SetStaticValue(2.5); err != nil {
		t.Fatal(err)
	}
	afterClaimedChange, err := StaticValuePreservationDigest(project, []*Property{target})
	if err != nil {
		t.Fatal(err)
	}
	if before != afterClaimedChange || before.RedactedBytes != 8 || before.NodeCount == 0 || before.DataBytes == 0 {
		t.Fatalf("claimed change affected digest: before=%+v after=%+v", before, afterClaimedChange)
	}

	back := projectBack(project)
	if back == nil || len(back.root.Trailing) == 0 {
		t.Fatal("fixture needs opaque trailing bytes")
	}
	back.root.Trailing[0] ^= 0xff
	afterOpaqueChange, err := StaticValuePreservationDigest(project, []*Property{target})
	if err != nil {
		t.Fatal(err)
	}
	if before.SHA256 == afterOpaqueChange.SHA256 {
		t.Fatal("opaque trailing-byte mutation was not detected")
	}
}
