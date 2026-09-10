// Edit-static discovers a document-scoped target and exports a verified edit.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	aep "github.com/yueli-fx/aep-parser"
)

func main() {
	input := flag.String("in", "test_data/fixtures/transform_unseparated.aep", "input AEP")
	output := flag.String("out", "tmp/sdk-edited.aep", "new output file; must not exist")
	match := flag.String("match", "ADBE Time Remapping", "property match name; must identify one write target")
	value := flag.String("value", "[2.5]", "numeric components as JSON, e.g. [2.5] or [640,360]")
	flag.Parse()
	if err := run(*input, *output, *match, *value); err != nil {
		log.Fatal(err)
	}
}

func run(input, output, match, value string) error {
	var components []float64
	if err := json.Unmarshal([]byte(value), &components); err != nil {
		return err
	}
	doc, err := aep.Open(input)
	if err != nil {
		return err
	}
	data, err := doc.ProjectJSON()
	if err != nil {
		return err
	}
	var snapshot struct {
		Properties []struct {
			MatchName string           `json:"match_name"`
			Target    *aep.WriteTarget `json:"write_target"`
		} `json:"property_records"`
	}
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return err
	}
	var target *aep.WriteTarget
	for _, property := range snapshot.Properties {
		if property.MatchName != match || property.Target == nil {
			continue
		}
		if target != nil {
			return fmt.Errorf("multiple targets for %q; choose a specific property from the snapshot", match)
		}
		target = property.Target
	}
	if target == nil {
		return fmt.Errorf("no write target for %q", match)
	}
	var result bytes.Buffer
	report, err := doc.Export(context.Background(), aep.ExportRequest{
		SchemaVersion: aep.ExportSchemaVersion,
		Changes: []aep.ExportChange{{
			ID: "example-edit", Operation: aep.ExportSetStaticValue,
			Target: *target, Value: components,
		}},
	}, &result)
	if encodeErr := json.NewEncoder(os.Stderr).Encode(report); encodeErr != nil {
		return encodeErr
	}
	if err != nil {
		return err
	}
	if report.Status != aep.ExportStatusVerified || !report.Preservation.Verified {
		return fmt.Errorf("export not verified: %s", report.Status)
	}
	if err := os.MkdirAll(filepath.Dir(output), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := result.WriteTo(f); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	fmt.Println(output)
	return nil
}
