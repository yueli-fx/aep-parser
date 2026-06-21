// Dump raw keyframe time-ticks + lhd3 hex for every keyframed property under a
// named composition's FIRST layer. Reveals the stored tick base (orig vs clone).
// Usage: go run ./tools/debug/dump_kf_ticks file.aep "comp name"
package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: dump_kf_ticks file.aep \"comp name\"")
		os.Exit(2)
	}
	compName := os.Args[2]
	f, _ := os.Open(os.Args[1])
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// find the Item LIST whose Utf8 == compName
	var compItem *rifx.Chunk
	var walk func(c *rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if c.IsList() && c.FormType == rifx.IDItem {
			if u := c.FindFirst(rifx.IDUtf8); u != nil && u.Text() == compName {
				compItem = c
			}
		}
		for _, ch := range c.Children {
			walk(ch)
		}
	}
	walk(root)
	if compItem == nil {
		fmt.Fprintln(os.Stderr, "no comp "+compName)
		os.Exit(1)
	}

	// walk every tdbs in the comp; if it carries a kfl list, print preceding
	// tdmn matchName + lhd3 + the per-keyframe @0x00 time tick.
	var lastTdmn string
	layerIdx := -1
	var walk2 func(c *rifx.Chunk)
	walk2 = func(c *rifx.Chunk) {
		for _, ch := range c.Children {
			switch {
			case ch.ID == rifx.IDLayr || (ch.IsList() && ch.FormType == rifx.IDLayr):
				layerIdx++
				// only dump first 1 layer to keep output focused
			case ch.ID == rifx.IDTdmn:
				lastTdmn = strings.TrimRight(string(ch.Data), "\x00")
			case ch.IsList() && ch.FormType == rifx.IDTdbs:
				mn := lastTdmn
				kfl := ch.FindFirstList(rifx.IDkfl)
				if kfl != nil && layerIdx == 0 {
					if tdb4 := ch.FindFirst(rifx.IDtdb4); tdb4 != nil {
						fmt.Printf("--- L0 prop %q tdb4(%d): %s\n", mn, len(tdb4.Data), hex.EncodeToString(tdb4.Data))
					}
					lhd3 := kfl.FindFirst(rifx.IDLhd3)
					ldat := kfl.FindFirst(rifx.IDLdat)
					if lhd3 != nil && ldat != nil {
						n := binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C])
						bpk := binary.BigEndian.Uint32(lhd3.Data[0x10:0x14])
						fmt.Printf("--- L0 prop %q : %d kf, bpk=%d ---\n", mn, n, bpk)
						fmt.Printf("    lhd3: %s\n", hex.EncodeToString(lhd3.Data))
						for i := 0; i < int(n) && (i+1)*int(bpk) <= len(ldat.Data); i++ {
							blk := ldat.Data[i*int(bpk) : (i+1)*int(bpk)]
							t := binary.BigEndian.Uint32(blk[0x00:0x04])
							fmt.Printf("    kf%d time-tick=%d  block=%s\n", i, t, hex.EncodeToString(blk))
						}
					}
				}
			}
			walk2(ch)
		}
	}
	walk2(compItem)
}
