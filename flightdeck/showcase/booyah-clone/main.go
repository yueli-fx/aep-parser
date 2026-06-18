// flightdeck/showcase/booyah-clone/main.go — rebuild the entire "Booyah Glitch"
// project from scratch via our Go API, DAG leaf-first, the original as a read-only
// value oracle. Plan: flightdeck/plans/2026-06-19-booyah-glitch-replication.md.
//
// Phase 0 scaffold: open the original and list its comps to confirm the oracle
// reads it. Per-comp builders (gen_<comp>.go) land in later tasks and assemble
// into booyah-clone.aep in DAG topological order.
//
// Run from repo root: `go run ./flightdeck/showcase/booyah-clone`.
package main

import (
	"fmt"
	"os"
)

func main() {
	orc, err := openOriginal()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("original: %d compositions\n", len(orc.proj.Compositions))
	for _, c := range orc.proj.Compositions {
		fmt.Printf("  id=%-4d %3d layers  %dx%d %.0ffps %.1fs  %q\n",
			c.ID, len(c.Layers), c.Width, c.Height, c.FrameRate, c.Duration, c.Name)
	}
}
