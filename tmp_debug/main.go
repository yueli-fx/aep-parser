// Diagnostic dumper: walks the RIFX tree of an .aep file and hex-dumps
// every tdbs that contains a LIST/list (i.e. keyframed properties).
//
// Usage:
//
//	go run ./tmp_debug path/to/project.aep
package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <path/to/project.aep>\n", os.Args[0])
		os.Exit(2)
	}
	aepPath := os.Args[1]

	f, err := os.Open(aepPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "open:", err)
		os.Exit(1)
	}
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, "rifx parse:", err)
		os.Exit(1)
	}

	tdbsFT := chunkID("tdbs")
	listFT := chunkID("list")

	var lastTdmn string
	tdmnID := chunkID("tdmn")
	count := 0
	var walk func(c *rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		for _, ch := range c.Children {
			if ch.ID == tdmnID {
				lastTdmn = trimNul(ch.Data)
			}
			if ch.IsList() && ch.FormType == tdbsFT {
				// Check if this tdbs contains a LIST/list
				hasKF := false
				for _, gc := range ch.Children {
					if gc.IsList() && gc.FormType == listFT {
						hasKF = true
						break
					}
				}
				if hasKF {
					count++
					fmt.Printf("\n=== keyframed property #%d: %q ===\n", count, lastTdmn)
					dump(ch, 0)
				}
			}
			walk(ch)
		}
	}
	walk(root)
}

func chunkID(s string) rifx.ChunkID {
	var id rifx.ChunkID
	copy(id[:], s)
	return id
}
func trimNul(b []byte) string { return string(bytes.TrimRight(b, "\x00")) }

func dump(c *rifx.Chunk, depth int) {
	indent := strings.Repeat("  ", depth)
	if c.IsList() {
		fmt.Printf("%s[%s/%s] (%d ch)\n", indent, c.ID, c.FormType, len(c.Children))
		for _, ch := range c.Children {
			dump(ch, depth+1)
		}
		return
	}
	full := isFull(c.ID)
	if full {
		fmt.Printf("%s%s  size=%d  FULL:\n", indent, c.ID, len(c.Data))
		for i := 0; i < len(c.Data); i += 16 {
			end := min(i+16, len(c.Data))
			fmt.Printf("%s  %04X  %s\n", indent, i, hexLine(c.Data[i:end]))
		}
		return
	}
	limit := min(32, len(c.Data))
	fmt.Printf("%s%s  size=%-4d  %s\n", indent, c.ID, len(c.Data), hexLine(c.Data[:limit]))
}

func isFull(id rifx.ChunkID) bool {
	s := id.String()
	return s == "tdb4" || s == "cdat" || s == "lhd3" || s == "ldat"
}

func hexLine(d []byte) string {
	var sb strings.Builder
	for i, b := range d {
		fmt.Fprintf(&sb, "%02X ", b)
		if i == 7 {
			sb.WriteByte(' ')
		}
	}
	sb.WriteString("  ")
	for _, b := range d {
		if b >= 0x20 && b < 0x7F {
			sb.WriteByte(b)
		} else {
			sb.WriteByte('.')
		}
	}
	return sb.String()
}
