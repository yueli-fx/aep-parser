package main
import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/example/aep-parser/internal/rifx"
)
func main() {
	path := "test_data/re_wave2_ae24.aep"
	wantComp := "RE_CDTA_DSF"
	if len(os.Args) > 1 { wantComp = os.Args[1] }
	f, _ := os.Open(path)
	defer f.Close()
	root, _ := rifx.Parse(f)
	var walk func(c *rifx.Chunk, parentName string)
	walk = func(c *rifx.Chunk, parentName string) {
		thisName := parentName
		if c.IsList() && c.FormType == rifx.IDItem {
			if utf8 := c.FindFirst(rifx.IDUtf8); utf8 != nil {
				thisName = utf8.Text()
			}
			if thisName == wantComp {
				fmt.Printf("=== Item %q ===\n", thisName)
				for _, ch := range c.Children {
					if !ch.IsList() {
						fmt.Printf("  child chunk ID=%q size=%d\n", ch.ID, len(ch.Data))
						fmt.Println(hex.Dump(ch.Data))
					}
				}
			}
		}
		for _, ch := range c.Children { walk(ch, thisName) }
	}
	walk(root, "")
}
