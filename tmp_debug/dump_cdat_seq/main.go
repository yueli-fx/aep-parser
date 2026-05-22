package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)

// Sequential cdat / tdb4 dump under a specified named-tdgp match-prefix.
// Usage: dump_cdat_seq <aep> <matchprefix>
// Walks the tree, finds the named property (tdmn) whose ASCII match starts with matchprefix,
// then prints its sibling LIST tdgp tree showing tdmn names + tdb4 dims + cdat full bytes (hex + float64 BE decoded).
func main() {
	f, _ := os.Open(os.Args[1])
	defer f.Close()
	target := os.Args[2]
	root, err := rifx.Parse(f)
	if err != nil {
		panic(err)
	}
	walk(root, target, 0, false)
}

func walk(c *rifx.Chunk, target string, depth int, active bool) {
	if !c.IsList() {
		return
	}
	tag := string(c.ID[:])
	form := string(c.FormType[:])
	_ = tag
	// Look ahead: find tdmn whose data matches target prefix
	for i, ch := range c.Children {
		if !ch.IsList() && string(ch.ID[:]) == "tdmn" && len(ch.Data) >= len(target) {
			name := string(trim(ch.Data))
			if startsWith(name, target) {
				// Find following LIST sibling tdgp
				if i+1 < len(c.Children) && c.Children[i+1].IsList() {
					sibling := c.Children[i+1]
					fmt.Printf("=== found '%s' at depth %d, following sibling: %s formType=%s children=%d ===\n", name, depth, string(sibling.ID[:]), string(sibling.FormType[:]), len(sibling.Children))
					dumpDetail(sibling, depth+1)
				}
				return
			}
		}
	}
	for _, ch := range c.Children {
		if ch.IsList() {
			walk(ch, target, depth+1, active)
		}
	}
	_ = form
}

func dumpDetail(c *rifx.Chunk, depth int) {
	ind := repeat("  ", depth)
	if c.IsList() {
		fmt.Printf("%sLIST %s (form=%s, %d children)\n", ind, string(c.ID[:]), string(c.FormType[:]), len(c.Children))
		for _, ch := range c.Children {
			dumpDetail(ch, depth+1)
		}
		return
	}
	tag := string(c.ID[:])
	switch tag {
	case "tdmn":
		name := string(trim(c.Data))
		fmt.Printf("%schunk tdmn name=%q\n", ind, name)
	case "tdb4":
		dim := byte(0)
		if len(c.Data) > 3 {
			dim = c.Data[3]
		}
		fmt.Printf("%schunk tdb4 (%dB) dim@03=%d head=%s\n", ind, len(c.Data), dim, hex.EncodeToString(c.Data[:min(16, len(c.Data))]))
	case "cdat":
		fmt.Printf("%schunk cdat (%dB) hex=%s\n", ind, len(c.Data), hex.EncodeToString(c.Data))
		// Decode as float64 BE × N
		n := len(c.Data) / 8
		vals := make([]float64, n)
		for i := 0; i < n; i++ {
			b := c.Data[i*8 : i*8+8]
			u := binary.BigEndian.Uint64(b)
			vals[i] = math.Float64frombits(u)
		}
		fmt.Printf("%s   decoded float64 BE x%d = %v\n", ind, n, vals)
	default:
		if len(c.Data) <= 32 {
			fmt.Printf("%schunk %s hex=%s\n", ind, tag, hex.EncodeToString(c.Data))
		} else {
			fmt.Printf("%schunk %s (%dB) head=%s\n", ind, tag, len(c.Data), hex.EncodeToString(c.Data[:16]))
		}
	}
}

func trim(b []byte) []byte {
	end := len(b)
	for end > 0 && b[end-1] == 0x00 {
		end--
	}
	return b[:end]
}

func startsWith(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return s[:len(prefix)] == prefix
}

func repeat(s string, n int) string {
	r := ""
	for i := 0; i < n; i++ {
		r += s
	}
	return r
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
