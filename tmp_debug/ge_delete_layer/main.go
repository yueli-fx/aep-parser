// tmp_debug/ge_delete_layer/main.go
//
// V3 Phase 2 Task 5 — produces 4 Go-emitted .aep fixtures from the
// AE-saved RE inputs, ready for AE 2020 + AE 2025 ship-gate (8 opens).
//
// Modes (matches the JSX harness modes):
//   baseline → Open(re_delete_layer_baseline.aep) → WriteAEP (no-op
//              round-trip; validates write path doesn't corrupt)
//   middle   → Open(baseline) → DeleteLayer(1) → WriteAEP
//   parent   → Open(baseline) → L3.SetParent(L2.ID) → DeleteLayer(1)
//              → WriteAEP
//   matte_ae20 → Open(re_delete_layer_matte_predelete_ae20.aep, AE 2020
//                saved w/ implicit "layer above" matte) → DeleteLayer(0)
//                of L1 → WriteAEP. Exercises the ldta-too-short skip
//                branch in TrackMatteLayerID cleanup.
//   matte_ae25 → Open(re_delete_layer_matte_predelete_ae25.aep, AE 2025
//                saved w/ explicit TrackMatteLayerID @0xA0) →
//                DeleteLayer(0) of L1 → WriteAEP. Exercises the actual
//                @0xA0 byte clear + the F3 "leave @0x6B alone" rule.
//
// Outputs:
//   test_data/ge_delete_layer_<mode>.aep   (4 files)
//
// Usage:
//   go run ./tmp_debug/ge_delete_layer
//
// Fixtures required at test_data/:
//   re_delete_layer_baseline.aep         (run JSX with RE_DELETE_MODE=baseline)
//   re_delete_layer_matte_predelete.aep  (run JSX with RE_DELETE_MODE=matte_predelete on AE 2025)
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

const (
	inputBaseline           = "test_data/re_delete_layer_baseline.aep"
	inputMattePredeleteAE20 = "test_data/re_delete_layer_matte_predelete_ae20.aep"
	inputMattePredeleteAE25 = "test_data/re_delete_layer_matte_predelete_ae25.aep"
	outputDir               = "test_data"
)

type modeSpec struct {
	name  string
	input string
	mut   func(*aep.Project) error
}

func main() {
	specs := []modeSpec{
		{
			name:  "baseline",
			input: inputBaseline,
			mut:   nil, // no-op — just round-trip
		},
		{
			name:  "middle",
			input: inputBaseline,
			mut: func(p *aep.Project) error {
				return p.Compositions[0].DeleteLayer(1) // L2_mid
			},
		},
		{
			name:  "parent",
			input: inputBaseline,
			mut: func(p *aep.Project) error {
				c := p.Compositions[0]
				l2, l3 := c.Layers[1], c.Layers[2]
				if err := l3.SetParent(l2.ID); err != nil {
					return fmt.Errorf("SetParent: %w", err)
				}
				return c.DeleteLayer(1) // L2_mid (now L3's parent)
			},
		},
		{
			name:  "matte_ae20",
			input: inputMattePredeleteAE20,
			mut: func(p *aep.Project) error {
				return p.Compositions[0].DeleteLayer(0) // L1_top (implicit matte source)
			},
		},
		{
			name:  "matte_ae25",
			input: inputMattePredeleteAE25,
			mut: func(p *aep.Project) error {
				return p.Compositions[0].DeleteLayer(0) // L1_top (explicit matte source)
			},
		},
	}

	for _, s := range specs {
		out := fmt.Sprintf("%s/ge_delete_layer_%s.aep", outputDir, s.name)
		if err := runMode(s, out); err != nil {
			fmt.Fprintf(os.Stderr, "FAIL %s: %v\n", s.name, err)
			os.Exit(1)
		}
		fi, _ := os.Stat(out)
		size := int64(0)
		if fi != nil {
			size = fi.Size()
		}
		fmt.Printf("OK   %s → %s (%d bytes)\n", s.name, out, size)
	}
}

func runMode(s modeSpec, out string) error {
	proj, err := aep.Open(s.input)
	if err != nil {
		return fmt.Errorf("Open %s: %w", s.input, err)
	}
	if s.mut != nil {
		if err := s.mut(proj); err != nil {
			return fmt.Errorf("mut: %w", err)
		}
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
