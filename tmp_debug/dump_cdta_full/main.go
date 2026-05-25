package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	f, _ := os.Open(os.Args[1])
	defer f.Close()
	root, _ := rifx.Parse(f)
	var find func(*rifx.Chunk) []byte
	find = func(c *rifx.Chunk) []byte {
		if c.IsList() && string(c.FormType[:]) == "Item" {
			for _, ch := range c.Children {
				if string(ch.ID[:]) == "cdta" {
					return ch.Data
				}
			}
		}
		for _, ch := range c.Children {
			if ch.IsList() {
				if d := find(ch); d != nil {
					return d
				}
			}
		}
		return nil
	}
	d := find(root)
	if d == nil {
		fmt.Println("no cdta")
		return
	}
	for i := 0; i < len(d); i += 16 {
		end := i + 16
		if end > len(d) {
			end = len(d)
		}
		fmt.Printf("@0x%02x  %s\n", i, strings.Join(splitHex(d[i:end]), " "))
	}
}

func splitHex(b []byte) []string {
	out := make([]string, len(b))
	for i, x := range b {
		out[i] = hex.EncodeToString([]byte{x})
	}
	return out
}
