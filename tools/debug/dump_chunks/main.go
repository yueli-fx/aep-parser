// tmp_debug/dump_chunks/main.go
//
// Phase 5 ship-gate debug aid — generic chunk-tree dump for any .aep.
//
// Usage:
//   go run tmp_debug/dump_chunks/main.go <path.aep> [max_depth]
//
// Output: indented tree. LIST chunks show `[ID FormType]`; leaf chunks
// show `ID (N B)` plus the trimmed tdmn data when applicable.
package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: dump_chunks <path.aep> [max_depth]")
		os.Exit(2)
	}
	maxDepth := 99
	if len(os.Args) >= 3 {
		if d, err := strconv.Atoi(os.Args[2]); err == nil {
			maxDepth = d
		}
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, "rifx.Parse:", err)
		os.Exit(1)
	}

	var walk func(c *rifx.Chunk, depth int)
	walk = func(c *rifx.Chunk, depth int) {
		ind := ""
		for i := 0; i < depth; i++ {
			ind += "  "
		}
		if c.IsList() {
			fmt.Printf("%s[%s %s]\n", ind, c.ID, c.FormType)
		} else {
			label := string(c.ID[:])
			if c.ID == rifx.IDTdmn {
				label += " = " + trimZ(c.Data)
			}
			fmt.Printf("%s%s (%d B)\n", ind, label, len(c.Data))
		}
		if depth >= maxDepth {
			return
		}
		for _, ch := range c.Children {
			walk(ch, depth+1)
		}
	}
	walk(root, 0)
}

func trimZ(b []byte) string {
	n := len(b)
	for n > 0 && b[n-1] == 0 {
		n--
	}
	return string(b[:n])
}
