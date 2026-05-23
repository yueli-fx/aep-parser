// tmp_debug/dump_ldta_hex/main.go
//
// Hex-dump every ldta chunk in a given .aep, with offsets and ASCII view.
// Use to diff per-Layer byte layout between AE-saved files and builder output.
//
// Run: go run tmp_debug/dump_ldta_hex/main.go <path>

package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: dump_ldta_hex <path.aep>")
		os.Exit(1)
	}
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
	idx := 0
	var visit func(c *rifx.Chunk, parentForm string)
	visit = func(c *rifx.Chunk, parentForm string) {
		if c.ID == rifx.IDLdta {
			fmt.Printf("=== ldta #%d  parent=%s  size=%d B ===\n", idx, parentForm, len(c.Data))
			idx++
			dumpHex(c.Data)
			fmt.Println()
			return
		}
		pf := parentForm
		if c.IsList() {
			pf = string(c.FormType[:])
		}
		for _, ch := range c.Children {
			visit(ch, pf)
		}
	}
	visit(root, "<root>")
}

func dumpHex(data []byte) {
	for i := 0; i < len(data); i += 16 {
		end := i + 16
		if end > len(data) {
			end = len(data)
		}
		hexStr := hex.EncodeToString(data[i:end])
		// space every 2 bytes
		var spaced strings.Builder
		for j := 0; j < len(hexStr); j += 4 {
			je := j + 4
			if je > len(hexStr) {
				je = len(hexStr)
			}
			spaced.WriteString(hexStr[j:je])
			spaced.WriteByte(' ')
		}
		ascii := ""
		for _, b := range data[i:end] {
			if b >= 0x20 && b < 0x7f {
				ascii += string(b)
			} else {
				ascii += "."
			}
		}
		fmt.Printf("  %04x: %-40s |%s|\n", i, spaced.String(), ascii)
	}
}
