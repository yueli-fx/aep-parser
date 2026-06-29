// Demo runner: parses the .aep file given as argv[1], prints keyframes per
// layer/property, then demonstrates Keyframe.SetTime / SetValue +
// Project.WriteAEP by:
//   1. Doubling each keyframe's value on the first layer's Opacity
//   2. Shifting all Position keyframes 1 second earlier
//   3. Writing the result to <file>.kf.aep
//   4. Re-parsing the output to confirm the edits survived
//
// Usage:
//
//	go run ./main path/to/project.aep
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <path/to/project.aep>\n", os.Args[0])
		os.Exit(2)
	}
	aepPath := os.Args[1]

	project, err := aep.Open(aepPath)
	check(err, "parse")

	for _, w := range project.Warnings {
		fmt.Fprintln(os.Stderr, "warning:", w)
	}

	fmt.Println("=== BEFORE ===")
	dumpKeyframes(project)

	// Demo edits on the first layer of the first composition.
	if len(project.Compositions) == 0 || len(project.Compositions[0].Layers) == 0 {
		fmt.Fprintln(os.Stderr, "no layers; nothing to edit")
		return
	}
	layer := project.Compositions[0].Layers[0]

	edits := 0
	for _, p := range layer.Properties {
		switch p.MatchName {
		case "ADBE Opacity":
			for _, kf := range p.Keyframes {
				cur := kf.Value.(float64)
				if err := kf.SetValue(cur / 2); err != nil {
					fmt.Fprintln(os.Stderr, "opacity set:", err)
					continue
				}
				edits++
			}
		case "ADBE Position":
			for _, kf := range p.Keyframes {
				if err := kf.SetTime(kf.Time - 1.0); err != nil {
					fmt.Fprintln(os.Stderr, "position time:", err)
					continue
				}
				edits++
			}
		}
	}
	fmt.Printf("\n(applied %d keyframe edits)\n", edits)

	base := strings.TrimSuffix(aepPath, filepath.Ext(aepPath))
	outAEP := base + ".kf.aep"
	check(writeTo(outAEP, project.WriteAEP), "write aep")
	fmt.Println("wrote", outAEP)

	verify, err := aep.Open(outAEP)
	check(err, "re-parse modified")

	fmt.Println("\n=== AFTER (re-parsed) ===")
	dumpKeyframes(verify)
}

func dumpKeyframes(project *aep.Project) {
	for _, comp := range project.Compositions {
		for _, l := range comp.Layers {
			hasKF := false
			for _, p := range l.Properties {
				if len(p.Keyframes) > 0 {
					hasKF = true
					break
				}
			}
			if !hasKF {
				continue
			}
			fmt.Printf("Layer[%d] src=%d\n", l.Index+1, l.SourceID)
			for _, p := range l.Properties {
				if len(p.Keyframes) == 0 {
					continue
				}
				fmt.Printf("  %s (%dD):\n", p.MatchName, p.Components)
				for i, kf := range p.Keyframes {
					fmt.Printf("    kf[%d]  t=%.4fs  v=%v\n", i, kf.Time, kf.Value)
				}
			}
		}
	}
}

func writeTo(path string, fn func(io.Writer) error) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return fn(f)
}

func check(err error, ctx string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", ctx, err)
		os.Exit(1)
	}
}
