// tmp_debug/ge_duplicate_layer/main.go
//
// V3 Phase 3 Task 5 — produces 3 Go-emitted .aep fixtures from the
// shared 3-solid baseline, ready for AE 2020 + AE 2025 ship-gate
// (3 modes × 2 versions = 6 opens).
//
// Modes (mirror RE harness modes; matte by-design refused per
// specs/2026-05-28-v3-phase3-duplicatelayer-strategy.md §5):
//
//   solo       → Open(baseline) → DuplicateLayer(1, "L2_clone")
//                Clone of L2_mid pushed in at idx 1; source moves to 2.
//   dup_parent → Open(baseline) → L3.SetParent(L2.ID)
//                → DuplicateLayer(1, "L2_clone")
//                Exercises F6: child L3's ParentID still points to
//                ORIGINAL L2 (clone is a sibling shadow, not a promoted
//                parent).
//   dup_child  → Open(baseline) → L1.SetParent(L2.ID)
//                → DuplicateLayer(0, "L1_clone")
//                Exercises F7-inverse: clone inherits source's outgoing
//                ParentID (clone.parent = L2).
//
// Outputs:
//   test_data/ge_duplicate_layer_<mode>.aep   (3 files)
//
// Usage:
//   go run ./tmp_debug/ge_duplicate_layer
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

const (
	inputBaseline = "test_data/re_delete_layer_baseline.aep"
	outputDir     = "test_data"
)

type modeSpec struct {
	name string
	mut  func(*aep.Project) error
}

func main() {
	specs := []modeSpec{
		{
			name: "solo",
			mut: func(p *aep.Project) error {
				_, err := p.Compositions[0].DuplicateLayer(1, "L2_clone")
				return err
			},
		},
		{
			name: "dup_parent",
			mut: func(p *aep.Project) error {
				c := p.Compositions[0]
				l2, l3 := c.Layers[1], c.Layers[2]
				if err := l3.SetParent(l2.ID); err != nil {
					return fmt.Errorf("SetParent: %w", err)
				}
				_, err := c.DuplicateLayer(1, "L2_clone")
				return err
			},
		},
		{
			name: "dup_child",
			mut: func(p *aep.Project) error {
				c := p.Compositions[0]
				l1, l2 := c.Layers[0], c.Layers[1]
				if err := l1.SetParent(l2.ID); err != nil {
					return fmt.Errorf("SetParent: %w", err)
				}
				_, err := c.DuplicateLayer(0, "L1_clone")
				return err
			},
		},
	}

	for _, s := range specs {
		out := fmt.Sprintf("%s/ge_duplicate_layer_%s.aep", outputDir, s.name)
		if err := runMode(s, out); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", s.name, err)
			os.Exit(1)
		}
		fi, _ := os.Stat(out)
		size := int64(0)
		if fi != nil {
			size = fi.Size()
		}
		fmt.Printf("OK   %s -> %s (%d bytes)\n", s.name, out, size)
	}
}

func runMode(s modeSpec, out string) error {
	proj, err := aep.Open(inputBaseline)
	if err != nil {
		return fmt.Errorf("Open %s: %w", inputBaseline, err)
	}
	if err := s.mut(proj); err != nil {
		return fmt.Errorf("mut: %w", err)
	}
	if len(proj.Warnings) > 0 {
		return fmt.Errorf("parser warnings: %v", proj.Warnings)
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
