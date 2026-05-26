// dump_root_compare lists the root chunk children of one or more .aep
// files. Used to detect chunk ordering / presence differences (e.g.
// where AE inserts a flag chunk vs. our builder).
package main

import (
	"fmt"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: <file1.aep> [file2.aep ...]")
		os.Exit(2)
	}
	for _, p := range os.Args[1:] {
		fmt.Printf("=== %s ===\n", p)
		f, _ := os.Open(p)
		root, _ := rifx.Parse(f)
		f.Close()
		for i, ch := range root.Children {
			ft := "----"
			if ch.IsList() {
				ft = ch.FormType.String()
			}
			extra := ""
			if !ch.IsList() && len(ch.Data) <= 16 {
				extra = fmt.Sprintf("  bytes=% 02x", ch.Data)
			}
			fmt.Printf("  [%2d] %s form=%s len=%d%s\n", i, ch.ID.String(), ft, len(ch.Data), extra)
		}
		fmt.Println()
	}
}
