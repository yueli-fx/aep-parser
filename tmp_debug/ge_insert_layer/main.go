// tmp_debug/ge_insert_layer/main.go
//
// V3 Phase 5C — produces 3 Go-emitted .aep fixtures for the
// InsertLayer AE 2020 + AE 2025 ship-gate (3 modes × 2 versions = 6 opens).
//
// For each mode it opens the RE baseline re_insert_layer_<mode>_before.aep
// (two sibling comps in one Project: compA_src_holder + compB_dest, plus a
// shared asset for footage/precomp), then calls:
//
//   compB.InsertLayer(compA.Layers[0], 0)
//
// i.e. deep-clone compA's single src layer into the top of compB — the same
// operation AE's src.copyToComp(compB) performs to build the _after baseline.
//
// Modes:
//   basic    — src is a plain solid; clone refs the shared solid footage.
//   footage  — src is a solid whose footage is also referenced (disabled) in
//              compB; clone must ref the SAME footage item (SourceID verbatim).
//   precomp  — src's source is precomp compC (also referenced in compB);
//              clone must ref compC (SourceID verbatim, no precomp loop).
//
// Outputs:
//   test_data/ge_insert_layer_<mode>.aep   (3 files)
//
// Usage:
//   go run ./tmp_debug/ge_insert_layer
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

const outputDir = "test_data"

var modes = []string{"basic", "footage", "precomp"}

func main() {
	for _, mode := range modes {
		in := fmt.Sprintf("%s/re_insert_layer_%s_before.aep", outputDir, mode)
		out := fmt.Sprintf("%s/ge_insert_layer_%s.aep", outputDir, mode)
		if err := runMode(in, out); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", mode, err)
			os.Exit(1)
		}
		fi, _ := os.Stat(out)
		size := int64(0)
		if fi != nil {
			size = fi.Size()
		}
		fmt.Printf("OK   %s -> %s (%d bytes)\n", mode, out, size)
	}
}

func findComp(p *aep.Project, name string) *aep.Composition {
	for _, c := range p.Compositions {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func runMode(in, out string) error {
	proj, err := aep.Open(in)
	if err != nil {
		return fmt.Errorf("Open %s: %w", in, err)
	}
	compA := findComp(proj, "compA_src_holder")
	compB := findComp(proj, "compB_dest")
	if compA == nil || compB == nil {
		return fmt.Errorf("baseline missing compA_src_holder/compB_dest (have %d comps)", len(proj.Compositions))
	}
	if len(compA.Layers) == 0 {
		return fmt.Errorf("compA has no src layer")
	}
	src := compA.Layers[0]

	clone, err := compB.InsertLayer(src, 0)
	if err != nil {
		return fmt.Errorf("InsertLayer: %w", err)
	}
	if len(proj.Warnings) > 0 {
		return fmt.Errorf("parser warnings: %v", proj.Warnings)
	}
	// Sanity: clone identity must be reset cross-comp.
	if clone.ParentID != 0 || clone.TrackMatteLayerID != 0 {
		return fmt.Errorf("clone refs not reset: ParentID=%d TrackMatteLayerID=%d", clone.ParentID, clone.TrackMatteLayerID)
	}
	if clone.SourceID != src.SourceID {
		return fmt.Errorf("clone.SourceID=%d != src.SourceID=%d (must be verbatim)", clone.SourceID, src.SourceID)
	}

	f, err := os.Create(out)
	if err != nil {
		return fmt.Errorf("Create %s: %w", out, err)
	}
	defer f.Close()
	if err := proj.WriteAEP(f); err != nil {
		return fmt.Errorf("WriteAEP: %w", err)
	}
	return nil
}
