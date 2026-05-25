// tmp_debug/swap_reverse/main.go
//
// Reverse confirmation: take minfail_v2.aep (ours) as base + swap in
// tolerance's Transform Group. If layers.length=1, fully confirms Transform
// Group is the only silent-drop trigger in our builder.
package main

import (
	"fmt"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	baseFile, _ := os.Open("tmp_debug/minfail_v2.aep")
	baseRoot, _ := rifx.Parse(baseFile)
	baseFile.Close()

	donorFile, _ := os.Open("test_data/v2_2_shape_tolerance.aep")
	donorRoot, _ := rifx.Parse(donorFile)
	donorFile.Close()

	baseLayr := findFirstUserLayr(baseRoot)
	donorLayr := findFirstUserLayr(donorRoot)
	baseTdgp := findOuterTdgp(baseLayr)
	donorTdgp := findOuterTdgp(donorLayr)

	// Find tolerance's Transform Group body chunk.
	var donorTransformBody *rifx.Chunk
	for i, ch := range donorTdgp.Children {
		if ch.ID == rifx.IDTdmn && trimNUL(string(ch.Data)) == "ADBE Transform Group" && i+1 < len(donorTdgp.Children) {
			donorTransformBody = donorTdgp.Children[i+1]
			break
		}
	}
	if donorTransformBody == nil {
		fmt.Println("tolerance has no Transform Group")
		os.Exit(1)
	}

	// Replace ours Transform Group body with tolerance's.
	for i, ch := range baseTdgp.Children {
		if ch.ID == rifx.IDTdmn && trimNUL(string(ch.Data)) == "ADBE Transform Group" && i+1 < len(baseTdgp.Children) {
			baseTdgp.Children[i+1] = donorTransformBody
			break
		}
	}

	out, _ := os.Create("tmp_debug/swap_reverse.aep")
	defer out.Close()
	_ = baseRoot.Write(out)
	fmt.Println("wrote tmp_debug/swap_reverse.aep — base=ours, Transform Group from tolerance")
}

func findFirstUserLayr(root *rifx.Chunk) *rifx.Chunk {
	var found *rifx.Chunk
	var walk func(*rifx.Chunk)
	walk = func(c *rifx.Chunk) {
		if found != nil {
			return
		}
		for _, ch := range c.Children {
			if ch.IsList() && ch.FormType == rifx.IDLayr {
				found = ch
				return
			}
			if ch.IsList() {
				walk(ch)
				if found != nil {
					return
				}
			}
		}
	}
	walk(root)
	return found
}

func findOuterTdgp(layr *rifx.Chunk) *rifx.Chunk {
	for _, ch := range layr.Children {
		if ch.IsList() && ch.FormType == rifx.IDTdgp {
			return ch
		}
	}
	return nil
}

func trimNUL(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			return s[:i]
		}
	}
	return s
}
