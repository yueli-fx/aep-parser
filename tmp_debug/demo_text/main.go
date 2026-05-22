// Read all text layers and print their decoded TextSource.
// Usage: go run ./tmp_debug/demo_text test_data/re_text.aep
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: demo_text file.aep")
		os.Exit(2)
	}
	p, err := aep.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	for _, c := range p.Compositions {
		for _, l := range c.Layers {
			if l.TextSource == nil {
				continue
			}
			ts := l.TextSource
			run0 := aep.TextStyleRun{FontIndex: -1}
			if len(ts.Runs) > 0 {
				run0 = ts.Runs[0]
			}
			boxStr := "point"
			if ts.IsBoxText {
				boxStr = fmt.Sprintf("box%v", ts.BoxBounds)
			}
			fmt.Printf("[%s/%s] text=%q just=%s font=%q size=%.2f color=%v lead=%v(auto=%v) track=%g stroke=%v width=%g paras=%d %s\n",
				c.Name, l.Name,
				ts.Text,
				ts.Justification.String(),
				run0.FontName,
				run0.FontSize,
				run0.FillColor,
				run0.Leading, run0.AutoLeading,
				run0.Tracking,
				run0.ApplyStroke, run0.StrokeWidth,
				len(ts.Paragraphs),
				boxStr,
			)
		}
	}
	for _, w := range p.Warnings {
		fmt.Fprintln(os.Stderr, "warn:", w)
	}
}
