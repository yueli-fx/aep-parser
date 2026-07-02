package aepmigrate

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/yueli-fx/aep-parser/internal/recipe"
)

func readMatrixRecipe(path string) (recipe.Recipe, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return recipe.Recipe{}, err
	}
	var rec recipe.Recipe
	if err := json.Unmarshal(data, &rec); err != nil {
		return recipe.Recipe{}, err
	}
	return rec, nil
}

func writeMatrixReport(path string, report MatrixReport) error {
	return writeMatrixJSON(path, report)
}

func writeMatrixJSON(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}
