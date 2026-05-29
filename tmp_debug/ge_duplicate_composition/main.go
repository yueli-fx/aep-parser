// tmp_debug/ge_duplicate_composition/main.go
//
// V3 Phase 5D — produces a Go-emitted .aep for the DuplicateComposition AE
// ship-gate. Opens the re_duplicate_item baseline (compA_main: 3 AV layers
// with an intra-comp parent ref + shared footage/precomp sources), calls
// p.DuplicateComposition(compA_main, "compA_dup"), and writes
// ge_duplicate_composition.aep.
//
// AE must accept the file and the dup must contain: 3 layers, the parented
// layer pointing at the DUP's own A_precomp (remapped, not the original
// comp's), and sources shared with the original.
//
// Usage:
//   go run ./tmp_debug/ge_duplicate_composition
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

const outputDir = "test_data"

func main() {
	var in string
	for _, fn := range []string{"re_duplicate_item_before.aep", "re_duplicate_item_after.aep"} {
		p := outputDir + "/" + fn
		if _, err := os.Stat(p); err == nil {
			in = p
			break
		}
	}
	if in == "" {
		fmt.Fprintln(os.Stderr, "FAIL: no re_duplicate_item_{before,after}.aep baseline found")
		os.Exit(1)
	}

	proj, err := aep.Open(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL Open %s: %v\n", in, err)
		os.Exit(1)
	}
	var src *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "compA_main" {
			src = c
			break
		}
	}
	if src == nil {
		fmt.Fprintln(os.Stderr, "FAIL: compA_main not found in baseline")
		os.Exit(1)
	}

	dup, err := proj.DuplicateComposition(src, "compA_dup")
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL DuplicateComposition: %v\n", err)
		os.Exit(1)
	}
	if len(proj.Warnings) > 0 {
		fmt.Fprintf(os.Stderr, "FAIL parser warnings: %v\n", proj.Warnings)
		os.Exit(1)
	}

	// Go-side sanity before handing to AE.
	srcIDs := map[uint32]bool{}
	for _, l := range src.Layers {
		srcIDs[l.ID] = true
	}
	for i, l := range dup.Layers {
		if srcIDs[l.ID] {
			fmt.Fprintf(os.Stderr, "FAIL dup layer %d ID %d collides with src\n", i, l.ID)
			os.Exit(1)
		}
		if l.SourceID != src.Layers[i].SourceID {
			fmt.Fprintf(os.Stderr, "FAIL dup layer %d SourceID %d != src %d\n", i, l.SourceID, src.Layers[i].SourceID)
			os.Exit(1)
		}
	}

	out := outputDir + "/ge_duplicate_composition.aep"
	f, err := os.Create(out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL Create %s: %v\n", out, err)
		os.Exit(1)
	}
	defer f.Close()
	if err := proj.WriteAEP(f); err != nil {
		fmt.Fprintf(os.Stderr, "FAIL WriteAEP: %v\n", err)
		os.Exit(1)
	}
	fi, _ := os.Stat(out)
	fmt.Printf("OK   in=%s -> %s (dup id=%d, %d layers)\n", in, out, dup.ID, len(dup.Layers))
	if fi != nil {
		fmt.Printf("     size=%d bytes\n", fi.Size())
	}
}
