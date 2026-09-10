// Compile builds an AEP from versioned Recipe JSON using only the public SDK.
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
	input := flag.String("recipe", "examples/recipes/minimal-null-layer.json", "Recipe JSON")
	output := flag.String("out", "tmp/sdk-compiled.aep", "new output file; must not exist")
	flag.Parse()
	if err := run(*input, *output); err != nil {
		log.Fatal(err)
	}
}

func run(input, output string) error {
	spec, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	var result bytes.Buffer
	report, err := aep.Compile(context.Background(), spec, &result)
	if encodeErr := json.NewEncoder(os.Stderr).Encode(report); encodeErr != nil {
		return encodeErr
	}
	if err != nil {
		return err
	}
	if report.Status != aep.CompileStatusVerified {
		return fmt.Errorf("compile status: %s", report.Status)
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
