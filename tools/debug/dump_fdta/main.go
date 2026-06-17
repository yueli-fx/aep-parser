// Dump root Fold LIST's fdta chunk hex + sibling Item LIST count.
// Usage: go run ./tmp_debug/dump_fdta file.aep
package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: dump_fdta file.aep")
		os.Exit(2)
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil {
		panic(err)
	}

	fold := root.FindFirstList(rifx.IDFold)
	if fold == nil {
		fmt.Println("no Fold LIST")
		return
	}
	for _, ch := range fold.Children {
		if ch.ID == rifx.IDFdta {
			fmt.Printf("fdta (%d B) hex=%s\n", len(ch.Data), hex.EncodeToString(ch.Data))
			items := 0
			for _, sib := range fold.Children {
				if sib.IsList() && sib.FormType == rifx.IDItem {
					items++
				}
			}
			fmt.Printf("  sibling Item count: %d\n", items)
			return
		}
	}
	fmt.Println("no fdta chunk inside Fold LIST")
}
