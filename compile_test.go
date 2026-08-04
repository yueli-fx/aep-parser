package aep_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	aep "github.com/yueli-fx/aep-parser"
)

func TestCompileBuildsAndReparsesVersionedRecipe(t *testing.T) {
	specification, err := os.ReadFile(filepath.Join("examples", "recipes", "minimal-default-text-layer.json"))
	if err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	report, err := aep.Compile(context.Background(), specification, &output)
	if err != nil {
		t.Fatalf("Compile: %v (report=%+v)", err, report)
	}
	if report.Status != aep.CompileStatusVerified || !report.Verification.Reparsed ||
		report.OutputBytes != output.Len() || output.Len() == 0 || len(report.Capabilities) == 0 {
		t.Fatalf("compile report=%+v output=%d", report, output.Len())
	}

	document, err := aep.Parse(bytes.NewReader(output.Bytes()))
	if err != nil {
		t.Fatalf("Parse compiled AEP: %v", err)
	}
	inspection := document.Inspect()
	if inspection.CompositionCount != 1 || inspection.LayerCount != 1 {
		t.Fatalf("compiled inspection=%+v", inspection)
	}
}

func TestCompileStrictlyRejectsUnknownRecipeFieldWithoutOutput(t *testing.T) {
	specification := []byte(`{
  "schema_version": 1,
  "project": {},
  "comps": [{"name":"Main","width":640,"height":360,"frame_rate":24,"duration":1}],
  "unexpected": true
}`)

	var output bytes.Buffer
	report, err := aep.Compile(context.Background(), specification, &output)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != aep.CompileStatusRejected || output.Len() != 0 || len(report.Rejected) != 1 ||
		report.Rejected[0].Code != "invalid-json" {
		t.Fatalf("strict rejection report=%+v output=%d", report, output.Len())
	}
}
