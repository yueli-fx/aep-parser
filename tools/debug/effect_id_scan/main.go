// Scans an .aep (or effect template .bin) for 32-bit BE occurrences of a given
// value inside every chunk's data, printing chunk ID + offset; also prints each
// Layr's ldta @0x00 layer ID. Used to chase the "cannot find layer ID=N"
// AE reject on spliced effect templates.
//
// Usage:
//
//	go run ./tmp_debug/effect_id_scan <file> <decimalValue>
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()
	want64, _ := strconv.ParseUint(os.Args[2], 10, 32)
	want := uint32(want64)

	var root *rifx.Chunk
	root, err = rifx.Parse(f)
	if err != nil {
		// maybe a bare chunk template (.bin)
		f.Seek(0, 0)
		root, err = rifx.ReadChunk(f)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	ldtaID := chunkID("ldta")
	layrFT := chunkID("Layr")
	utf8ID := chunkID("Utf8")

	var path []string
	var walk func(c *rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		label := c.ID.String()
		if c.IsList() {
			label += "/" + c.FormType.String()
		}
		path = append(path, label)
		if c.IsList() && c.FormType == layrFT {
			for _, ch := range c.Children {
				if ch.ID == ldtaID && len(ch.Data) >= 4 {
					fmt.Printf("Layr ldta ID=%d", binary.BigEndian.Uint32(ch.Data[0:4]))
				}
				if ch.ID == utf8ID {
					fmt.Printf(" name=%q", string(ch.Data))
				}
			}
			fmt.Println()
		}
		for i := 0; i+4 <= len(c.Data); i++ {
			if binary.BigEndian.Uint32(c.Data[i:i+4]) == want {
				fmt.Printf("  HIT %d @ %s +0x%X (chunk len %d)\n", want, pathStr(path), i, len(c.Data))
			}
		}
		for _, ch := range c.Children {
			walk(ch)
		}
		path = path[:len(path)-1]
	}
	walk(root)
}

func pathStr(p []string) string { return fmt.Sprint(p) }

func chunkID(s string) rifx.ChunkID {
	var id rifx.ChunkID
	copy(id[:], s)
	return id
}

var _ = bytes.TrimRight
