package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/recipe"
)

func TestRunValidateReturnsZeroForValidRecipe(t *testing.T) {
	path := writeRecipe(t, minimalCLIRecipe())

	code := run([]string{"validate", "-recipe", path, "-json"})

	if code != 0 {
		t.Fatalf("run validate = %d, want 0", code)
	}
}

func TestRunValidateReturnsOneForUnsupportedEffect(t *testing.T) {
	rec := minimalCLIRecipe()
	rec.Comps[0].Layers[0].Effects = []recipe.Effect{{MatchName: "Third Party Magic"}}
	path := writeRecipe(t, rec)

	code := run([]string{"validate", "-recipe", path, "-json"})

	if code != 1 {
		t.Fatalf("run validate unsupported = %d, want 1", code)
	}
}

func TestRunValidateReturnsTwoForMalformedJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}

	code := run([]string{"validate", "-recipe", path})

	if code != 2 {
		t.Fatalf("run validate malformed = %d, want 2", code)
	}
}

func TestRunCompileWritesAEP(t *testing.T) {
	path := writeRecipe(t, minimalCLIRecipe())
	out := filepath.Join(t.TempDir(), "out.aep")

	code := run([]string{"compile", "-recipe", path, "-out", out, "-json"})

	if code != 0 {
		t.Fatalf("run compile = %d, want 0", code)
	}
	if info, err := os.Stat(out); err != nil || info.Size() == 0 {
		t.Fatalf("compiled AEP stat = %v/%v", info, err)
	}
}

func TestRunExplainFieldPrintsRecipeIndexEntry(t *testing.T) {
	output, code := captureStdout(t, func() int {
		return run([]string{"explain", "-field", "comps[].background_color"})
	})

	if code != 0 {
		t.Fatalf("run explain -field = %d, want 0", code)
	}
	if !strings.Contains(output, "comps[].background_color") {
		t.Fatalf("output missing field path:\n%s", output)
	}
	if !strings.Contains(output, "comp.set_background_color") {
		t.Fatalf("output missing capability key:\n%s", output)
	}
}

func TestRunExplainFieldReturnsOneForUnknownField(t *testing.T) {
	code := run([]string{"explain", "-field", "comps[].does_not_exist", "-json"})

	if code != 1 {
		t.Fatalf("run explain unknown field = %d, want 1", code)
	}
}

func minimalCLIRecipe() recipe.Recipe {
	return recipe.Recipe{
		SchemaVersion: recipe.SchemaVersion,
		Project:       recipe.ProjectSpec{Name: "CLI Recipe"},
		Comps: []recipe.CompSpec{{
			Name:      "Main",
			Width:     1920,
			Height:    1080,
			FrameRate: 30,
			Duration:  2,
			Layers: []recipe.Layer{
				{
					Type: "text",
					Name: "Title",
					Text: "HELLO",
					Transform: recipe.Transform{
						Position: []float64{960, 540},
					},
				},
				{
					Type: "shape",
					Name: "Box",
					Shape: &recipe.ShapeSpec{
						Kind:      "rect",
						Size:      []float64{320, 40},
						FillColor: []float64{255, 255, 255},
					},
					Transform: recipe.Transform{
						Position: []float64{960, 650},
					},
				},
			},
		}},
	}
}

func writeRecipe(t *testing.T, rec recipe.Recipe) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "recipe.json")
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func captureStdout(t *testing.T, fn func() int) (string, int) {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	code := fn()
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	os.Stdout = old
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(data), code
}
