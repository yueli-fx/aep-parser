package main

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/yueli-fx/aep-parser/internal/rifx"
)

func main() {
	f, _ := os.Open(os.Args[1])
	defer f.Close()
	root, _ := rifx.Parse(f)
	var walk func(c *rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if c.IsList() && c.FormType == (rifx.ChunkID{'I','t','e','m'}) {
			for _, ch := range c.Children {
				if string(ch.ID[:]) == "idta" {
					fmt.Printf("idta (%d B): %s\n", len(ch.Data), hex.EncodeToString(ch.Data))
					return
				}
			}
		}
		for _, ch := range c.Children {
			walk(ch)
		}
	}
	walk(root)
}
