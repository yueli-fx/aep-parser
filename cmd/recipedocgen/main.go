package main

//go:generate go run . -out ../../docs

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yueli-fx/aep-parser/internal/recipedoc"
)

func main() {
	outDir := flag.String("out", "docs", "output docs directory")
	flag.Parse()
	schema, markdown, err := generate(*outDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "recipedocgen:", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "recipedocgen:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(filepath.Join(*outDir, "recipe_schema.json"), []byte(schema), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "recipedocgen:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(filepath.Join(*outDir, "recipe.md"), []byte(markdown), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, "recipedocgen:", err)
		os.Exit(1)
	}
}

func generate(docsDir string) (string, string, error) {
	doc, err := recipedoc.BuildDocumentWithCapabilities(filepath.Join(docsDir, "capabilities.json"))
	if err != nil {
		return "", "", err
	}
	schema, err := recipedoc.RenderJSONSchema(doc)
	if err != nil {
		return "", "", err
	}
	return schema, recipedoc.RenderMarkdown(doc), nil
}
