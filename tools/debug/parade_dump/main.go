// Dumps each Layr's outer LIST(tdgp) child order (tdmn names) and, for the
// "ADBE Effect Parade" group, hex of its tdsb/tdsn header chunks.
//
// Usage:
//
//	go run ./tmp_debug/parade_dump path/to/project.aep
package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/yueli-fx/aep-parser/internal/rifx"
)

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	layrFT := chunkID("Layr")
	tdgpFT := chunkID("tdgp")
	tdmnID := chunkID("tdmn")
	utf8ID := chunkID("Utf8")

	var walk func(c *rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if c.IsList() && c.FormType == layrFT {
			name := "?"
			for _, ch := range c.Children {
				if ch.ID == utf8ID {
					name = string(ch.Data)
				}
				if ch.IsList() && ch.FormType == tdgpFT {
					fmt.Printf("\n=== Layr %q outer tdgp (%d children) ===\n", name, len(ch.Children))
					dumpGroup(ch, tdmnID)
				}
			}
		}
		for _, ch := range c.Children {
			walk(ch)
		}
	}
	walk(root)
}

func dumpGroup(g *rifx.Chunk, tdmnID rifx.ChunkID) {
	var lastTdmn string
	for i, ch := range g.Children {
		switch {
		case ch.ID == tdmnID:
			lastTdmn = string(bytes.TrimRight(ch.Data, "\x00"))
			fmt.Printf("  [%2d] tdmn %q (len=%d)\n", i, lastTdmn, len(ch.Data))
		case ch.IsList():
			fmt.Printf("  [%2d] LIST(%s) %d children\n", i, ch.FormType, len(ch.Children))
			if lastTdmn == "ADBE Effect Parade" {
				fmt.Printf("       --- parade body ---\n")
				for j, pc := range ch.Children {
					if pc.IsList() {
						fmt.Printf("       [%2d] LIST(%s) %d children\n", j, pc.FormType, len(pc.Children))
					} else {
						fmt.Printf("       [%2d] %s len=%d hex=% x\n", j, pc.ID, len(pc.Data), pc.Data)
					}
				}
			}
		default:
			fmt.Printf("  [%2d] %s len=%d", i, ch.ID, len(ch.Data))
			if len(ch.Data) <= 16 {
				fmt.Printf(" hex=% x", ch.Data)
			}
			fmt.Println()
		}
	}
}

func chunkID(s string) rifx.ChunkID {
	var id rifx.ChunkID
	copy(id[:], s)
	return id
}
