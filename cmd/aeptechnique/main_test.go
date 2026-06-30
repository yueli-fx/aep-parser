package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/technique"
)

func TestRunEmitsTechniqueFactsJSON(t *testing.T) {
	input := filepath.Join("..", "..", "flightdeck", "showcase", "text", "text.aep")
	var stdout, stderr bytes.Buffer

	code := run([]string{"-in", input}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run exit = %d, stderr=%s", code, stderr.String())
	}

	var facts technique.FactSet
	if err := json.Unmarshal(stdout.Bytes(), &facts); err != nil {
		t.Fatalf("json.Unmarshal: %v\nstdout=%s", err, stdout.String())
	}
	if facts.SchemaVersion != technique.SchemaVersion {
		t.Fatalf("SchemaVersion = %d, want %d", facts.SchemaVersion, technique.SchemaVersion)
	}
	if facts.SourcePath != input {
		t.Fatalf("SourcePath = %q, want %q", facts.SourcePath, input)
	}
	if facts.Summary.CompCount == 0 || facts.Summary.LayerCount == 0 || len(facts.Layers) == 0 {
		t.Fatalf("facts summary/layers = %+v / %d", facts.Summary, len(facts.Layers))
	}
}

func TestRunAcceptsJSONFlag(t *testing.T) {
	input := filepath.Join("..", "..", "flightdeck", "showcase", "text", "text.aep")
	var stdout, stderr bytes.Buffer

	code := run([]string{"-in", input, "-json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run exit = %d, stderr=%s", code, stderr.String())
	}
	if !json.Valid(stdout.Bytes()) {
		t.Fatalf("stdout is not valid json: %s", stdout.String())
	}
}
