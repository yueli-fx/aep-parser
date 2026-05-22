package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)

// Dump lhd3 + ldat for the FIRST keyframe-bearing property in the AEP.
// Usage: go run ./tmp_debug/dump_kf <file.aep>
func main() {
	f, err := os.Open(os.Args[1])
	must(err)
	defer f.Close()
	root, err := rifx.Parse(f)
	must(err)
	walk(root, "")
}

func walk(c *rifx.Chunk, path string) {
	if !c.IsList() {
		return
	}
	if string(c.FormType[:]) == "tdbs" {
		// look for nested 'list' with lhd3 + ldat
		for _, child := range c.Children {
			if child.IsList() && string(child.FormType[:]) == "list" {
				var lhd3, ldat *rifx.Chunk
				for _, ch := range child.Children {
					if !ch.IsList() && string(ch.ID[:]) == "lhd3" {
						lhd3 = ch
					}
					if !ch.IsList() && string(ch.ID[:]) == "ldat" {
						ldat = ch
					}
				}
				if lhd3 != nil && ldat != nil {
					fmt.Printf("=== path=%s ===\n", path)
					dumpLHD3(lhd3)
					dumpLDAT(ldat, lhd3)
					fmt.Println()
				}
			}
		}
	}
	for _, ch := range c.Children {
		if ch.IsList() {
			label := string(ch.FormType[:])
			walk(ch, path+"/"+label)
		}
	}
}

func dumpLHD3(c *rifx.Chunk) {
	d := c.Data
	fmt.Printf("lhd3 (%dB) hex=%s\n", len(d), hex.EncodeToString(d))
	if len(d) < 0x34 {
		fmt.Printf("  WARN: lhd3 < 52 bytes\n")
		return
	}
	for i := 0; i < len(d); i += 4 {
		end := i + 4
		if end > len(d) {
			end = len(d)
		}
		raw := d[i:end]
		u := uint32(0)
		if len(raw) == 4 {
			u = binary.BigEndian.Uint32(raw)
		}
		fmt.Printf("  [0x%02x] %-8s u32=%d\n", i, hex.EncodeToString(raw), u)
	}
}

func dumpLDAT(c *rifx.Chunk, lhd3 *rifx.Chunk) {
	d := c.Data
	count := int(binary.BigEndian.Uint32(lhd3.Data[0x08:0x0C]))
	bpk := int(binary.BigEndian.Uint32(lhd3.Data[0x10:0x14]))
	fmt.Printf("ldat (%dB) count=%d bpk=%d (=count*bpk=%d, match=%v)\n", len(d), count, bpk, count*bpk, count*bpk == len(d))
	for k := 0; k < count; k++ {
		off := k * bpk
		end := off + bpk
		if end > len(d) {
			end = len(d)
		}
		blk := d[off:end]
		fmt.Printf("--- keyframe[%d] off=0x%x bpk=%d hex=%s\n", k, off, bpk, hex.EncodeToString(blk))
		dumpBlock(blk)
	}
}

func dumpBlock(blk []byte) {
	if len(blk) < 8 {
		return
	}
	fmt.Printf("  time(u32 BE @0x00)=%d  inInterp=0x%02x outInterp=0x%02x flags=0x%02x header07=0x%02x\n",
		binary.BigEndian.Uint32(blk[0:4]), blk[4], blk[5], blk[6], blk[7])
	for i := 0; i+8 <= len(blk); i += 8 {
		raw := blk[i : i+8]
		u := binary.BigEndian.Uint64(raw)
		fv := math.Float64frombits(u)
		fmt.Printf("  [0x%02x] %s  f64=%v\n", i, hex.EncodeToString(raw), fv)
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
