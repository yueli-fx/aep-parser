package main

import (
	"fmt"
	"os"

	"github.com/yueli-fx/aep-parser/internal/rifx"
)

func trimNUL(b []byte) string {
	for i := 0; i < len(b); i++ {
		if b[i] == 0 {
			return string(b[:i])
		}
	}
	return string(b)
}

// findParade locates the "ADBE Effect Parade" tdgp LIST anywhere in the tree.
func findParade(c *rifx.Chunk) *rifx.Chunk {
	kids := c.Children
	for i := 0; i < len(kids); i++ {
		ch := kids[i]
		if ch.ID == rifx.IDTdmn && trimNUL(ch.Data) == "ADBE Effect Parade" && i+1 < len(kids) {
			if next := kids[i+1]; next.IsList() && next.FormType == rifx.IDTdgp {
				return next
			}
		}
		if ch.IsList() {
			if p := findParade(ch); p != nil {
				return p
			}
		}
	}
	return nil
}

func main() {
	f, _ := os.Open(os.Args[1])
	defer f.Close()
	root, _ := rifx.Parse(f)
	parade := findParade(root)
	if parade == nil {
		fmt.Println("no Effect Parade found")
		return
	}
	fmt.Printf("Effect Parade tdgp: %d direct children\n", len(parade.Children))
	for i, ch := range parade.Children {
		if ch.ID == rifx.IDTdmn {
			fmt.Printf("  [%2d] tdmn %q\n", i, trimNUL(ch.Data))
		} else if ch.IsList() {
			fmt.Printf("  [%2d] LIST:%s (payloadBytes=%d, children=%d)\n", i, ch.FormType.String(), len(ch.Data), len(ch.Children))
		} else {
			fmt.Printf("  [%2d] leaf:%s (%d bytes)\n", i, ch.ID.String(), len(ch.Data))
		}
	}
}
