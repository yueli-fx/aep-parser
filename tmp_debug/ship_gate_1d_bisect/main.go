// ship_gate_1d_bisect generates 8 variants of re_cameralight.aep, each
// applying exactly ONE of the Task 1D Project setters. Used to bisect
// per-setter ship-gate failures. Driver: test_data/ship_gate_1d_bisect.jsx.
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

type variant struct {
	suffix string
	apply  func(*aep.Project) error
}

func main() {
	variants := []variant{
		{"lnrb", func(p *aep.Project) error { return p.SetLinearBlending(true) }},
		{"lnrp", func(p *aep.Project) error { return p.SetLinearizeWorkingSpace(true) }},
		{"acer", func(p *aep.Project) error { return p.SetCompensateForSceneReferredProfiles(false) }},
		{"adfr", func(p *aep.Project) error { return p.SetAudioSampleRate(44100) }},
		{"dwga", func(p *aep.Project) error { return p.SetWorkingGamma(2.2) }},
		{"gpug", func(p *aep.Project) error { return p.SetGpuAccelType("abcdef12-3456-7890-aaaa-bbbbccccdddd") }},
		{"exen", func(p *aep.Project) error { return p.SetExpressionEngine("extendscript") }},
		{"none", func(p *aep.Project) error { return nil }},
	}
	for _, v := range variants {
		proj, err := aep.Open("test_data/re_cameralight.aep")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := v.apply(proj); err != nil {
			fmt.Fprintf(os.Stderr, "apply %s: %v\n", v.suffix, err)
			continue
		}
		out := fmt.Sprintf("test_data/ship_gate_1d_bisect_%s.aep", v.suffix)
		f, err := os.Create(out)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if err := proj.WriteAEP(f); err != nil {
			fmt.Fprintf(os.Stderr, "WriteAEP %s: %v\n", v.suffix, err)
			f.Close()
			continue
		}
		f.Close()
		fmt.Printf("wrote %s\n", out)
	}
}
