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
	if string(c.FormType[:]) == "shap" {
		for _, child := range c.Children {
			if child.IsList() && string(child.FormType[:]) == "list" {
				for _, ch := range child.Children {
					if !ch.IsList() && string(ch.ID[:]) == "ldat" {
						fmt.Printf("ldat (%dB) hex=%s\n", len(ch.Data), hex.EncodeToString(ch.Data))
						n := len(ch.Data) / 4
						for i := 0; i < n; i++ {
							b := ch.Data[i*4 : i*4+4]
							u := binary.BigEndian.Uint32(b)
							v := math.Float32frombits(u)
							fmt.Printf("  [f32 %2d] %s = %v\n", i, hex.EncodeToString(b), v)
						}
					} else if !ch.IsList() && string(ch.ID[:]) == "lhd3" {
						fmt.Printf("lhd3 (%dB) hex=%s\n", len(ch.Data), hex.EncodeToString(ch.Data))
					}
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
