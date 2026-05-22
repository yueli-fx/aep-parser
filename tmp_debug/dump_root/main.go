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
	root, err := rifx.Parse(f)
	if err != nil { panic(err) }
	var dump func(c *rifx.Chunk, depth int)
	dump = func(c *rifx.Chunk, depth int) {
		ind := strings.Repeat("  ", depth)
		tag := string(c.ID[:])
		if c.IsList() {
			fmt.Printf("%sLIST %s (formType=%s, %d children)\n", ind, tag, string(c.FormType[:]), len(c.Children))
			for _, ch := range c.Children { dump(ch, depth+1) }
		} else {
			snip := ""
			if len(c.Data) > 0 && len(c.Data) <= 32 {
				snip = " hex=" + hex.EncodeToString(c.Data)
			} else if len(c.Data) > 0 {
				snip = fmt.Sprintf(" (%d B) head=%s", len(c.Data), hex.EncodeToString(c.Data[:16]))
			}
			fmt.Printf("%schunk %s%s\n", ind, tag, snip)
		}
	}
	dump(root, 0)
}
