package aep_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	aep "github.com/yueli-fx/aep-parser"
)

func TestExternalPackageCanInspectProfileAndRoundTrip(t *testing.T) {
	path := filepath.Join("test_data", "fixtures", "v2_smoke_ae2020.aep")
	document, err := aep.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	inspection := document.Inspect()
	if inspection.SchemaVersion != 1 || inspection.CompositionCount == 0 || len(inspection.Compositions) == 0 {
		t.Fatalf("inspection = %+v", inspection)
	}
	profileJSON, err := document.ProfileJSON()
	if err != nil {
		t.Fatalf("ProfileJSON: %v", err)
	}
	var profileEnvelope struct {
		SchemaVersion int `json:"schema_version"`
	}
	if err := json.Unmarshal(profileJSON, &profileEnvelope); err != nil || profileEnvelope.SchemaVersion != 1 {
		t.Fatalf("profile JSON schema = %+v, err = %v", profileEnvelope, err)
	}

	var output bytes.Buffer
	if err := document.Write(&output); err != nil {
		t.Fatalf("Write: %v", err)
	}
	roundTripped, err := aep.Parse(bytes.NewReader(output.Bytes()))
	if err != nil {
		t.Fatalf("Parse round trip: %v", err)
	}
	if got := roundTripped.Inspect().CompositionCount; got != inspection.CompositionCount {
		t.Fatalf("round-trip compositions = %d, want %d", got, inspection.CompositionCount)
	}
}

func TestNilDocumentMethodsReturnStableErrors(t *testing.T) {
	var document *aep.Document
	if _, err := document.ProfileJSON(); err == nil {
		t.Fatal("ProfileJSON succeeded on nil document")
	}
	if err := document.Write(&bytes.Buffer{}); err == nil {
		t.Fatal("Write succeeded on nil document")
	}
	if got := document.Inspect().SchemaVersion; got != aep.InspectionSchemaVersion {
		t.Fatalf("nil inspection schema = %d", got)
	}
}

func TestExternalPackageCanTightenParseLimits(t *testing.T) {
	path := filepath.Join("test_data", "fixtures", "v2_smoke_ae2020.aep")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	limits := aep.DefaultLimits()
	if limits.MaxInputBytes == 0 || limits.MaxNodes == 0 || limits.MaxDepth == 0 {
		t.Fatalf("DefaultLimits returned zero budgets: %+v", limits)
	}
	limits.MaxInputBytes = uint64(len(data) - 1)
	if _, err := aep.ParseWithLimits(bytes.NewReader(data), limits); err == nil {
		t.Fatal("ParseWithLimits accepted input above MaxInputBytes")
	}
	if _, err := aep.OpenWithLimits(path, limits); err == nil {
		t.Fatal("OpenWithLimits accepted input above MaxInputBytes")
	}
}
