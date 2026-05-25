// tmp_debug/diff_stream/main.go
//
// Detail-level dump of a named property stream's tdbs body. Locates the
// `LIST tdbs` chunk immediately following `tdmn(<matchName>)` in the first
// Layr, prints the full child list + chunk sizes + head bytes for each
// child. Side-by-side compare two .aep files.
//
// Usage: go run tmp_debug/diff_stream/main.go <ours.aep> <tolerance.aep> <matchName>
package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "usage: diff_stream <ours.aep> <tolerance.aep> <matchName>")
		os.Exit(1)
	}
	mn := os.Args[3]
	ours := loadStream(os.Args[1], mn)
	tol := loadStream(os.Args[2], mn)
	fmt.Printf("=== %s ===\n", mn)
	fmt.Println("--- OURS ---")
	dump(ours)
	fmt.Println("--- TOLERANCE ---")
	dump(tol)
}

func loadStream(path, matchName string) *rifx.Chunk {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil {
		panic(err)
	}
	var found *rifx.Chunk
	var visit func(*rifx.Chunk)
	visit = func(c *rifx.Chunk) {
		if found != nil {
			return
		}
		kids := c.Children
		for i := 0; i < len(kids); i++ {
			ch := kids[i]
			if ch.ID == rifx.IDTdmn && i+1 < len(kids) {
				n := bytes.IndexByte(ch.Data, 0)
				if n < 0 {
					n = len(ch.Data)
				}
				if string(ch.Data[:n]) == matchName && kids[i+1].IsList() {
					found = kids[i+1]
					return
				}
			}
			if ch.IsList() {
				visit(ch)
				if found != nil {
					return
				}
			}
		}
	}
	visit(root)
	return found
}

func dump(c *rifx.Chunk) {
	if c == nil {
		fmt.Println("  <not found>")
		return
	}
	fmt.Printf("  LIST %s (%d children)\n", string(c.FormType[:]), len(c.Children))
	for _, ch := range c.Children {
		if ch.IsList() {
			fmt.Printf("    LIST %s (%d children)\n", string(ch.FormType[:]), len(ch.Children))
			for _, gg := range ch.Children {
				fmt.Printf("      chunk %s (%dB) %s\n", string(gg.ID[:]), len(gg.Data), prevHex(gg.Data, 24))
			}
		} else {
			fmt.Printf("    chunk %s (%dB) %s\n", string(ch.ID[:]), len(ch.Data), prevHex(ch.Data, 24))
		}
	}
}

func prevHex(b []byte, max int) string {
	n := len(b)
	if n > max {
		n = max
	}
	suffix := ""
	if len(b) > n {
		suffix = "..."
	}
	return "hex=" + hex.EncodeToString(b[:n]) + suffix
}
