package aep_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	aep "github.com/yueli-fx/aep-parser"
)

type exportProjectSnapshot struct {
	PropertyIntegrity struct {
		PreservedOpaque  int `json:"preserved_opaque_count"`
		PreservedUnknown int `json:"preserved_unknown_count"`
		DroppedUnknown   int `json:"dropped_unknown_count"`
	} `json:"property_integrity"`
	PropertyRecords []struct {
		MatchName   string           `json:"match_name"`
		StaticValue any              `json:"static_value"`
		WriteTarget *aep.WriteTarget `json:"write_target"`
	} `json:"property_records"`
}

func TestDocumentExportAppliesStaticValueAndPreservesUnclaimedBytes(t *testing.T) {
	path := filepath.Join("test_data", "fixtures", "transform_unseparated.aep")
	sourceBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	document, err := aep.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	before := exportSnapshot(t, document)
	record := exportRecord(t, before, "ADBE Time Remapping")
	if record.WriteTarget == nil || record.WriteTarget.DocumentRef == "" || record.WriteTarget.CanonicalPath == "" {
		t.Fatalf("time-remap write target = %+v", record.WriteTarget)
	}

	var output bytes.Buffer
	report, err := document.Export(context.Background(), aep.ExportRequest{
		SchemaVersion: aep.ExportSchemaVersion,
		Changes: []aep.ExportChange{{
			ID: "time-remap", Operation: aep.ExportSetStaticValue,
			Target: *record.WriteTarget, Value: []float64{2.5},
		}},
	}, &output)
	if err != nil {
		t.Fatalf("Export: %v (report=%+v)", err, report)
	}
	if report.Status != aep.ExportStatusVerified || !report.Preservation.Verified ||
		report.Preservation.Mode != aep.ExportPreservationModeByteExact ||
		report.Preservation.ClaimedBytes != 8 || !report.Verification.Reparsed ||
		report.Verification.OperationsMatched != 1 || report.OutputBytes != output.Len() {
		t.Fatalf("export report = %+v", report)
	}
	if len(report.Operations) != 1 || report.Operations[0].Status != aep.ExportOperationApplied {
		t.Fatalf("operation report = %+v", report.Operations)
	}
	if len(sourceBytes) != output.Len() {
		t.Fatalf("length-preserving export changed file size: source=%d output=%d", len(sourceBytes), output.Len())
	}
	differentBytes := 0
	for index, sourceByte := range sourceBytes {
		if sourceByte != output.Bytes()[index] {
			differentBytes++
		}
	}
	if differentBytes == 0 || differentBytes > report.Preservation.ClaimedBytes {
		t.Fatalf("raw byte differences=%d claimed=%d", differentBytes, report.Preservation.ClaimedBytes)
	}

	written, err := aep.Parse(bytes.NewReader(output.Bytes()))
	if err != nil {
		t.Fatalf("Parse exported AEP: %v", err)
	}
	after := exportSnapshot(t, written)
	writtenRecord := exportRecord(t, after, "ADBE Time Remapping")
	if value, ok := writtenRecord.StaticValue.(float64); !ok || value != 2.5 {
		t.Fatalf("exported static value = %#v", writtenRecord.StaticValue)
	}
	if before.PropertyIntegrity != after.PropertyIntegrity || after.PropertyIntegrity.DroppedUnknown != 0 {
		t.Fatalf("property preservation before=%+v after=%+v", before.PropertyIntegrity, after.PropertyIntegrity)
	}

	// Export works on a private clone. The source Document remains reusable and
	// still reports its original value and document-scoped target.
	originalAgain := exportRecord(t, exportSnapshot(t, document), "ADBE Time Remapping")
	if value, ok := originalAgain.StaticValue.(float64); !ok || value != 0 ||
		originalAgain.WriteTarget == nil || originalAgain.WriteTarget.DocumentRef != record.WriteTarget.DocumentRef {
		t.Fatalf("source document changed after export: %+v", originalAgain)
	}
}

func TestDocumentExportRejectsDuplicateWriteTargetAsAtomicBatch(t *testing.T) {
	document, err := aep.Open(filepath.Join("test_data", "fixtures", "transform_unseparated.aep"))
	if err != nil {
		t.Fatal(err)
	}
	record := exportRecord(t, exportSnapshot(t, document), "ADBE Time Remapping")

	var output bytes.Buffer
	report, err := document.Export(context.Background(), aep.ExportRequest{
		SchemaVersion: aep.ExportSchemaVersion,
		Changes: []aep.ExportChange{
			{ID: "first", Operation: aep.ExportSetStaticValue, Target: *record.WriteTarget, Value: []float64{1}},
			{ID: "second", Operation: aep.ExportSetStaticValue, Target: *record.WriteTarget, Value: []float64{2}},
		},
	}, &output)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != aep.ExportStatusRejected || output.Len() != 0 || len(report.Operations) != 2 ||
		report.Operations[0].Code != "batch-rejected" ||
		report.Operations[1].Code != "duplicate-write-target" {
		t.Fatalf("duplicate-target report=%+v output=%d", report, output.Len())
	}
}

func TestDocumentExportRejectsForeignWriteTargetWithoutOutput(t *testing.T) {
	document, err := aep.Open(filepath.Join("test_data", "fixtures", "transform_unseparated.aep"))
	if err != nil {
		t.Fatal(err)
	}
	record := exportRecord(t, exportSnapshot(t, document), "ADBE Time Remapping")
	target := *record.WriteTarget
	target.DocumentRef = "sha256:foreign"

	var output bytes.Buffer
	report, err := document.Export(context.Background(), aep.ExportRequest{
		SchemaVersion: aep.ExportSchemaVersion,
		Changes: []aep.ExportChange{{
			ID: "foreign", Operation: aep.ExportSetStaticValue, Target: target, Value: []float64{1},
		}},
	}, &output)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != aep.ExportStatusRejected || output.Len() != 0 || len(report.Operations) != 1 ||
		report.Operations[0].Code != "document-ref-mismatch" ||
		report.Operations[0].Status != aep.ExportOperationRejected {
		t.Fatalf("foreign-target report=%+v output=%d", report, output.Len())
	}
}

func exportSnapshot(t *testing.T, document *aep.Document) exportProjectSnapshot {
	t.Helper()
	payload, err := document.ProjectJSON()
	if err != nil {
		t.Fatalf("ProjectJSON: %v", err)
	}
	var snapshot exportProjectSnapshot
	if err := json.Unmarshal(payload, &snapshot); err != nil {
		t.Fatalf("decode ProjectJSON: %v", err)
	}
	return snapshot
}

func exportRecord(t *testing.T, snapshot exportProjectSnapshot, matchName string) struct {
	MatchName   string           `json:"match_name"`
	StaticValue any              `json:"static_value"`
	WriteTarget *aep.WriteTarget `json:"write_target"`
} {
	t.Helper()
	for _, record := range snapshot.PropertyRecords {
		if record.MatchName == matchName {
			return record
		}
	}
	t.Fatalf("property record %q not found", matchName)
	return struct {
		MatchName   string           `json:"match_name"`
		StaticValue any              `json:"static_value"`
		WriteTarget *aep.WriteTarget `json:"write_target"`
	}{}
}
