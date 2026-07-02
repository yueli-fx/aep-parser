package aepmigrate

import (
	"path/filepath"
	"sort"
)

func matrixRecipePaths(opts MatrixOptions) ([]string, error) {
	seen := map[string]bool{}
	var paths []string
	add := func(path string) {
		if path == "" || seen[path] {
			return
		}
		seen[path] = true
		paths = append(paths, path)
	}
	for _, path := range opts.RecipePaths {
		add(path)
	}
	if opts.RecipeDir != "" {
		matches, err := filepath.Glob(filepath.Join(opts.RecipeDir, "*.json"))
		if err != nil {
			return nil, err
		}
		for _, path := range matches {
			add(path)
		}
	}
	sort.Strings(paths)
	return paths, nil
}
