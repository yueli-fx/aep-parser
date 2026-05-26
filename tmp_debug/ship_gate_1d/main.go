// ship_gate_1d generates two fixtures for the Task 1D ship gate:
//   - test_data/ship_gate_1d_baseline.aep — copy of re_cameralight.aep
//   - test_data/ship_gate_1d_modified.aep — same with all 8 settings flipped
//
// Driver: test_data/ship_gate_1d.jsx (reads AE-visible values and
// writes ship_gate_1d.done). Required to match AE 2020 since the
// source fixture is AE 2020 (avoid version-conversion GUI prompt).
package main

import (
	"fmt"
	"io"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	const src = "test_data/re_cameralight.aep"
	const baseline = "test_data/ship_gate_1d_baseline.aep"
	const modified = "test_data/ship_gate_1d_modified.aep"

	if err := copyFile(src, baseline); err != nil {
		return fmt.Errorf("copy baseline: %w", err)
	}

	proj, err := aep.Open(src)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	fmt.Printf("baseline (%s):\n", src)
	fmt.Printf("  Rev=%d LB=%v LWS=%v CSR=%v ASR=%v WG=%v GAT=%q EE=%q\n",
		proj.Revision(), proj.LinearBlending(), proj.LinearizeWorkingSpace(),
		proj.CompensateForSceneReferredProfiles(), proj.AudioSampleRate(),
		proj.WorkingGamma(), proj.GpuAccelType(), proj.ExpressionEngine())

	if err := proj.SetLinearBlending(true); err != nil {
		return err
	}
	if err := proj.SetLinearizeWorkingSpace(true); err != nil {
		return err
	}
	if err := proj.SetCompensateForSceneReferredProfiles(false); err != nil {
		return err
	}
	if err := proj.SetAudioSampleRate(44100); err != nil {
		return err
	}
	if err := proj.SetWorkingGamma(2.2); err != nil {
		return err
	}
	if err := proj.SetExpressionEngine("extendscript"); err != nil {
		return err
	}

	f, err := os.Create(modified)
	if err != nil {
		return fmt.Errorf("create modified: %w", err)
	}
	defer f.Close()
	if err := proj.WriteAEP(f); err != nil {
		return fmt.Errorf("WriteAEP: %w", err)
	}
	fmt.Printf("wrote %s\n", modified)
	fmt.Println("next: AfterFX.exe -r test_data/ship_gate_1d.jsx (use AE 2020 to match fixture version)")
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
