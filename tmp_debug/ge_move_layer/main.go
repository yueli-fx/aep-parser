// tmp_debug/ge_move_layer/main.go
//
// V3 Phase 4 Task 3 — produces 3 Go-emitted .aep fixtures from the
// shared 3-solid baseline, ready for AE 2020 + AE 2025 ship-gate
// (3 modes × 2 versions = 6 opens).
//
// Modes:
//   first_to_last → Open(baseline) → MoveLayer(0, 2)
//                   L1_top moves to end: [L2_mid, L3_bot, L1_top]
//   last_to_first → Open(baseline) → MoveLayer(2, 0)
//                   L3_bot moves to start: [L3_bot, L1_top, L2_mid]
//   mid_swap      → Open(baseline) → MoveLayer(1, 0)
//                   L2_mid moves up one: [L2_mid, L1_top, L3_bot]
//
// Outputs:
//   test_data/ge_move_layer_<mode>.aep   (3 files)
//
// Usage:
//   go run ./tmp_debug/ge_move_layer
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
			name: "first_to_last",
			mut: func(p *aep.Project) error {
				return p.Compositions[0].MoveLayer(0, 2)
			},
		},
		{
			name: "last_to_first",
			mut: func(p *aep.Project) error {
				return p.Compositions[0].MoveLayer(2, 0)
			},
		},
		{
			name: "mid_swap",
			mut: func(p *aep.Project) error {
				return p.Compositions[0].MoveLayer(1, 0)
			},
		},
	}

	for _, s := range specs {
		out := fmt.Sprintf("%s/ge_move_layer_%s.aep", outputDir, s.name)
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
