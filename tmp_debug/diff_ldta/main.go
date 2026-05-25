// tmp_debug/diff_ldta/main.go
//
// Byte-by-byte diff of the FIRST Layr's ldta chunk between two .aep files.
// Prints "addr  ours  tolerance  ascii" for every differing byte. Identical
// stretches are summarized as "[NN bytes identical]".
//
// Usage: go run tmp_debug/diff_ldta/main.go <ours.aep> <tolerance.aep>
package main

import (
	"fmt"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: diff_ldta <ours.aep> <tolerance.aep>")
		os.Exit(1)
	}
	a := loadFirstLayrLdta(os.Args[1])
	b := loadFirstLayrLdta(os.Args[2])
	if a == nil || b == nil {
		fmt.Fprintln(os.Stderr, "first Layr's ldta not found in one or both files")
		os.Exit(1)
	}
	fmt.Printf("ours ldta = %d B; tolerance ldta = %d B\n\n", len(a), len(b))
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	diffs := 0
	run := 0
	for i := 0; i < n; i++ {
		if a[i] == b[i] {
			run++
			continue
		}
		if run > 0 {
			fmt.Printf("  [%d bytes identical]\n", run)
			run = 0
		}
		fmt.Printf("  @0x%02x  ours=0x%02x  tol=0x%02x\n", i, a[i], b[i])
		diffs++
	}
	if run > 0 {
		fmt.Printf("  [%d bytes identical]\n", run)
	}
	if len(a) != len(b) {
		fmt.Printf("\nsize mismatch: ours=%d tol=%d (tail %d B unique to %s)\n",
			len(a), len(b), abs(len(a)-len(b)),
			pickLonger(len(a), len(b)))
	}
	fmt.Printf("\n%d differing bytes (of first %d)\n", diffs, n)
}

func loadFirstLayrLdta(path string) []byte {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil {
		panic(err)
	}
	var found []byte
	var walk func(*rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if found != nil {
			return
		}
		for _, ch := range c.Children {
			if ch.IsList() && ch.FormType == rifx.IDLayr {
				ldta := ch.FindFirst(rifx.IDLdta)
				if ldta != nil {
					found = ldta.Data
					return
				}
			}
			if ch.IsList() {
				walk(ch)
				if found != nil {
					return
				}
			}
		}
	}
	walk(root)
	return found
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func pickLonger(a, b int) string {
	if a > b {
		return "ours"
	}
	return "tolerance"
}
