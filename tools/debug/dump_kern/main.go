package main

import (
	"fmt"
	"os"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func main() {
	p, err := aep.Open(os.Args[1])
	if err != nil { fmt.Println(err); os.Exit(1) }
	target := os.Args[2]
	for _, c := range p.Compositions {
		for _, l := range c.Layers {
			if l.Name != target { continue }
			if l.TextSource == nil {
				fmt.Println("no TextSource")
				return
			}
			fmt.Printf("text=%q (len=%d)\n", l.TextSource.Text, len([]rune(l.TextSource.Text)))
			fmt.Printf("Kerning=%d ManualKerning=%v\n", l.TextSource.Kerning, l.TextSource.ManualKerning)
			return
		}
	}
	fmt.Println("layer not found")
}
