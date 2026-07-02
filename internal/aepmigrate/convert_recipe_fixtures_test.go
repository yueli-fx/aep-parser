package aepmigrate

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/recipe"
)

func writeTempRecipeDefaultNullLayer(t *testing.T) string {
	t.Helper()
	rec := recipe.Recipe{
		SchemaVersion: 1,
		Project:       recipe.ProjectSpec{TargetVersion: "AE2020", Name: "default-null"},
		Comps: []recipe.CompSpec{{
			Name:      "Main",
			Width:     640,
			Height:    360,
			FrameRate: 24,
			Duration:  2,
			Layers: []recipe.Layer{{
				Type: "null",
				Name: "Controller",
			}},
		}},
	}
	path := filepath.Join(t.TempDir(), "recipe-default-null.aep")
	report, err := recipe.CompileToFile(rec, path, recipe.StaticCapabilities{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report.Valid = false, refusals=%+v", report.Refusals)
	}
	return path
}

func writeTempRecipe(t *testing.T, recipePath string) string {
	t.Helper()
	raw, err := os.ReadFile(recipePath)
	if err != nil {
		t.Fatalf("ReadFile recipe: %v", err)
	}
	var rec recipe.Recipe
	if err := json.Unmarshal(raw, &rec); err != nil {
		t.Fatalf("Unmarshal recipe: %v", err)
	}
	path := filepath.Join(t.TempDir(), "recipe-source.aep")
	report, err := recipe.CompileToFile(rec, path, recipe.StaticCapabilities{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report.Valid = false, refusals=%+v", report.Refusals)
	}
	return path
}

func writeTempRecipeWithTarget(t *testing.T, recipePath, targetVersion string) string {
	t.Helper()
	raw, err := os.ReadFile(recipePath)
	if err != nil {
		t.Fatalf("ReadFile recipe: %v", err)
	}
	var rec recipe.Recipe
	if err := json.Unmarshal(raw, &rec); err != nil {
		t.Fatalf("Unmarshal recipe: %v", err)
	}
	rec.Project.TargetVersion = targetVersion
	path := filepath.Join(t.TempDir(), "recipe-source-"+targetVersion+".aep")
	report, err := recipe.CompileToFile(rec, path, recipe.StaticCapabilities{})
	if err != nil {
		t.Fatalf("CompileToFile: %v", err)
	}
	if !report.Valid {
		t.Fatalf("report.Valid = false, refusals=%+v", report.Refusals)
	}
	return path
}
