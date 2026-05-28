// tmp_debug/diff_dup_blocks/main.go
//
// V3 Phase 3 RE helper. Given a duplicate-fixture, byte-compares the
// source layer's 16-chunk block (Layr + Ewst + 14 followers) against
// the clone's 16-chunk block. Answers: does AE byte-clone the 14
// follower chunks (fvdv/fiop/...) or does it mutate any field?
//
// Usage:
//   go run ./tmp_debug/diff_dup_blocks <fixture.aep> <clone_block_idx> <source_block_idx>
//
// Block indices = position in Item LIST.Children of the Layr LIST
// for each layer. For solo fixture (clone at parse idx 1, source at
// parse idx 2): clone block at children[25], source block at children[41].
package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: diff_dup_blocks <file.aep> <clone_idx> <source_idx>")
		os.Exit(2)
	}
	path := os.Args[1]
	cloneIdx, _ := strconv.Atoi(os.Args[2])
	srcIdx, _ := strconv.Atoi(os.Args[3])

	f, err := os.Open(path)
	must(err)
	defer f.Close()
	root, err := rifx.Parse(f)
	must(err)

	item := findFirstItemWithLayrs(root)
	if item == nil {
		fmt.Fprintln(os.Stderr, "no Item LIST with Layr children found")
		os.Exit(1)
	}
	kids := item.Children
	fmt.Printf("Item LIST has %d children\n", len(kids))
	if cloneIdx >= len(kids) || srcIdx >= len(kids) {
		fmt.Fprintf(os.Stderr, "indices out of range\n")
		os.Exit(1)
	}

	// For each pair, walk forward until the next LIST chunk. That defines
	// the layer's full block (Layr + Ewst + N followers, where N is
	// adaptive). Then compare per-position.
	cloneBlock := readBlock(kids, cloneIdx)
	srcBlock := readBlock(kids, srcIdx)

	fmt.Printf("\nClone block @ children[%d..%d) — %d chunks\n", cloneIdx, cloneIdx+len(cloneBlock), len(cloneBlock))
	fmt.Printf("Source block @ children[%d..%d) — %d chunks\n", srcIdx, srcIdx+len(srcBlock), len(srcBlock))
	if len(cloneBlock) != len(srcBlock) {
		fmt.Printf("BLOCK LENGTH MISMATCH — clone=%d source=%d\n", len(cloneBlock), len(srcBlock))
	}

	min := len(cloneBlock)
	if len(srcBlock) < min {
		min = len(srcBlock)
	}

	for i := 0; i < min; i++ {
		c, s := cloneBlock[i], srcBlock[i]
		label := chunkLabel(c)
		srcLabel := chunkLabel(s)
		if label != srcLabel {
			fmt.Printf("  [%d] TYPE MISMATCH clone=%s src=%s\n", i, label, srcLabel)
			continue
		}
		if c.IsList() {
			// LIST — compare children count + recurse first level
			fmt.Printf("  [%d] %s (LIST, clone=%d kids, src=%d kids)\n", i, label, len(c.Children), len(s.Children))
			if len(c.Children) != len(s.Children) {
				fmt.Printf("       LIST KIDS COUNT MISMATCH\n")
			}
			continue
		}
		// Leaf — byte compare Data
		if equalBytes(c.Data, s.Data) {
			fmt.Printf("  [%d] %s (%d bytes) BYTE-IDENTICAL\n", i, label, len(c.Data))
		} else {
			fmt.Printf("  [%d] %s (%d bytes) DIFFER:\n", i, label, len(c.Data))
			fmt.Printf("       clone: %s\n", hex.EncodeToString(c.Data))
			fmt.Printf("       src:   %s\n", hex.EncodeToString(s.Data))
			// Show byte-level diffs
			min2 := len(c.Data)
			if len(s.Data) < min2 {
				min2 = len(s.Data)
			}
			for off := 0; off < min2; off++ {
				if c.Data[off] != s.Data[off] {
					fmt.Printf("         @0x%02x clone=0x%02x src=0x%02x\n", off, c.Data[off], s.Data[off])
				}
			}
		}
	}
}

// readBlock returns chunks [startIdx, endIdx) where endIdx is the next
// LIST chunk after startIdx+1 (skip the Layr LIST itself + its Ewst at
// startIdx+1; consume leaves until the next LIST).
func readBlock(children []*rifx.Chunk, startIdx int) []*rifx.Chunk {
	// Sanity: Children[startIdx] should be a Layr LIST.
	end := startIdx + 2 // past Layr + Ewst
	for end < len(children) && !children[end].IsList() {
		end++
	}
	return children[startIdx:end]
}

func findFirstItemWithLayrs(root *rifx.Chunk) *rifx.Chunk {
	var found *rifx.Chunk
	var walk func(c *rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if found != nil {
			return
		}
		if c.IsList() && c.FormType == rifx.IDItem {
			for _, kid := range c.Children {
				if kid.IsList() && kid.FormType == rifx.IDLayr {
					found = c
					return
				}
			}
		}
		for _, kid := range c.Children {
			walk(kid)
		}
	}
	walk(root)
	return found
}

func chunkLabel(c *rifx.Chunk) string {
	if c.IsList() {
		return "LIST(" + string(c.FormType[:]) + ")"
	}
	return string(c.ID[:])
}

func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
