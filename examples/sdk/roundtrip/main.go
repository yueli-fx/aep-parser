// Roundtrip writes an unmodified project, reparses it, and compares profiles.
package main

import (
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	aep "github.com/yueli-fx/aep-parser"
)

func main() {
	input := flag.String("in", "test_data/fixtures/v2_smoke.aep", "input AEP")
	output := flag.String("out", "tmp/sdk-roundtrip.aep", "new output file; must not exist")
	flag.Parse()
	if err := run(*input, *output); err != nil {
		log.Fatal(err)
	}
}

func run(input, output string) error {
	source, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	doc, err := aep.Parse(bytes.NewReader(source))
	if err != nil {
		return err
	}
	var result bytes.Buffer
	if err := doc.Write(&result); err != nil {
		return err
	}
	reparsed, err := aep.Parse(bytes.NewReader(result.Bytes()))
	if err != nil {
		return err
	}
	beforeProfile, err := doc.ProfileJSON()
	if err != nil {
		return err
	}
	afterProfile, err := reparsed.ProfileJSON()
	if err != nil {
		return err
	}
	if !bytes.Equal(beforeProfile, afterProfile) {
		return fmt.Errorf("round-trip profiles differ; refusing to write output")
	}
	byteIdentical := bytes.Equal(source, result.Bytes())
	outputHash := sha256.Sum256(result.Bytes())
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
	fmt.Printf("profile-identical; byte-identical: %t\ninput SHA-256: %x\noutput SHA-256: %x\n", byteIdentical, sha256.Sum256(source), outputHash)
	return nil
}
