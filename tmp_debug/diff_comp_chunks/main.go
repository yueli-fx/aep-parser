// tmp_debug/diff_comp_chunks/main.go
//
// Byte-by-byte diff of the FIRST CompItem's structural chunks (cdta + idta)
// plus the root head chunk between two .aep files. Catches per-comp metadata
// drift that ldta-only diff misses.
package main

import (
	"fmt"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)

type compChunks struct {
	idta []byte
	cdta []byte
	head []byte
}

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: diff_comp_chunks <ours.aep> <tolerance.aep>")
		os.Exit(1)
	}
	a := load(os.Args[1])
	b := load(os.Args[2])
	fmt.Println("=== head ===")
	diff(a.head, b.head)
	fmt.Println("\n=== idta (first comp) ===")
	diff(a.idta, b.idta)
	fmt.Println("\n=== cdta (first comp) ===")
	diff(a.cdta, b.cdta)
}

func load(path string) compChunks {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil {
		panic(err)
	}
	out := compChunks{}
	// head: root child
	for _, ch := range root.Children {
		if string(ch.ID[:]) == "head" {
			out.head = ch.Data
			break
		}
	}
	// idta + cdta: first Item LIST's chunks
	var visit func(*rifx.Chunk) bool
	visit = func(c *rifx.Chunk) bool {
		if c.IsList() && string(c.FormType[:]) == "Item" {
			for _, ch := range c.Children {
				switch string(ch.ID[:]) {
				case "idta":
					if out.idta == nil {
						out.idta = ch.Data
					}
				case "cdta":
					if out.cdta == nil {
						out.cdta = ch.Data
					}
				}
			}
			if out.cdta != nil {
				return true
			}
		}
		for _, ch := range c.Children {
			if ch.IsList() {
				if visit(ch) {
					return true
				}
			}
		}
		return false
	}
	visit(root)
	return out
}

func diff(a, b []byte) {
	if a == nil || b == nil {
		fmt.Println("  missing one side")
		return
	}
	fmt.Printf("  ours=%dB  tol=%dB\n", len(a), len(b))
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	count := 0
	run := 0
	for i := 0; i < n; i++ {
		if a[i] == b[i] {
			run++
			continue
		}
		if run > 0 {
			fmt.Printf("    [%d bytes identical]\n", run)
			run = 0
		}
		fmt.Printf("    @0x%02x  ours=0x%02x  tol=0x%02x\n", i, a[i], b[i])
		count++
	}
	if run > 0 {
		fmt.Printf("    [%d bytes identical]\n", run)
	}
	fmt.Printf("  %d differing bytes\n", count)
}
