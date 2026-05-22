package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	f, _ := os.Open(os.Args[1])
	defer f.Close()
	root, err := rifx.Parse(f)
	if err != nil {
		panic(err)
	}
	walk(root)
}

func walk(c *rifx.Chunk) {
	if !c.IsList() {
		return
	}
	if string(c.FormType[:]) == "shap" || string(c.FormType[:]) == "list" || string(c.FormType[:]) == "om-s" || string(c.FormType[:]) == "omks" {
		fmt.Printf("LIST form=%s children=%d\n", string(c.FormType[:]), len(c.Children))
		for _, ch := range c.Children {
			if !ch.IsList() {
				if string(ch.ID[:]) == "ldat" || string(ch.ID[:]) == "shph" {
					fmt.Printf("  chunk %s (%dB) hex=%s\n", string(ch.ID[:]), len(ch.Data), hex.EncodeToString(ch.Data))
					if string(ch.ID[:]) == "ldat" {
						n := len(ch.Data) / 8
						for i := 0; i < n; i++ {
							b := ch.Data[i*8 : i*8+8]
							u := binary.BigEndian.Uint64(b)
							v := math.Float64frombits(u)
							fmt.Printf("    [%2d] %s = %v\n", i, hex.EncodeToString(b), v)
						}
					}
				} else if string(ch.ID[:]) == "lhd3" {
					fmt.Printf("  chunk lhd3 (%dB) hex=%s\n", len(ch.Data), hex.EncodeToString(ch.Data))
				}
			}
		}
	}
	for _, ch := range c.Children {
		if ch.IsList() {
			walk(ch)
		}
	}
}
