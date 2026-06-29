package main

import (
	"fmt"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: probe_effects file.aep")
		os.Exit(2)
	}
	proj, err := aep.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, c := range proj.Compositions {
		for _, l := range c.Layers {
			if len(l.Effects) == 0 {
				continue
			}
			fmt.Printf("comp=%q layer=%q effects=%d\n", c.Name, l.Name, len(l.Effects))
			for _, e := range l.Effects {
				fmt.Printf("  effect %s (%d params):\n", e.MatchName, len(e.Parameters))
				for _, p := range e.Parameters {
					sv := "nil"
					if p.StaticValue != nil {
						sv = fmt.Sprintf("%v", p.StaticValue)
					}
					fmt.Printf("    %s  components=%d static=%s default=%v kfs=%d\n",
						p.MatchName, p.Components, sv, p.DefaultValue, len(p.Keyframes))
				}
			}
		}
	}
}
