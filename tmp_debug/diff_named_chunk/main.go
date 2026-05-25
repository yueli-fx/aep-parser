// tmp_debug/diff_named_chunk/main.go
//
// Byte-by-byte diff of a named chunk (first occurrence) between two .aep files.
// Used to compare project-level registry chunks (AFsi / ARsi / Rhed / etc.).
//
// Usage: go run tmp_debug/diff_named_chunk/main.go <ours.aep> <tolerance.aep> <chunkID>
package main

import (
	"fmt"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "usage: diff_named_chunk <ours.aep> <tolerance.aep> <chunkID>")
		os.Exit(1)
	}
	want := os.Args[3]
	var wantID rifx.ChunkID
	copy(wantID[:], want)

	a := loadChunkData(os.Args[1], wantID)
	b := loadChunkData(os.Args[2], wantID)
	if a == nil || b == nil {
		fmt.Println("chunk not found in one or both files (a nil:", a == nil, " b nil:", b == nil, ")")
		return
	}
	fmt.Printf("ours=%dB  tol=%dB\n", len(a), len(b))
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
		fmt.Printf("  @0x%04x  ours=0x%02x  tol=0x%02x\n", i, a[i], b[i])
		diffs++
		if diffs > 50 {
			fmt.Println("  ...(>50 diffs, stopping)")
			return
		}
	}
	if run > 0 {
		fmt.Printf("  [%d bytes identical]\n", run)
	}
	fmt.Printf("%d differing bytes (of first %d)\n", diffs, n)
}

func loadChunkData(path string, id rifx.ChunkID) []byte {
	f, _ := os.Open(path)
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil {
		return nil
	}
	var out []byte
	var walk func(*rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if out != nil {
			return
		}
		for _, ch := range c.Children {
			if ch.ID == id {
				out = ch.Data
				return
			}
			if ch.IsList() {
				walk(ch)
				if out != nil {
					return
				}
			}
		}
	}
	walk(root)
	return out
}
