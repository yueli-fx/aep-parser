package main

import (
	"bytes"
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
	"github.com/example/aep-parser/internal/rifx"
)

func main() {
	p := aep.NewProject()
	c, _ := p.NewComposition("M", 1920, 1080, 30, 5)
	s, _ := c.NewShapeLayer("S1")
	r, _ := s.RootGroup().AddRect()
	_ = r.SetSize([2]float64{200, 100})

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		fmt.Println("WriteAEP err:", err)
		os.Exit(1)
	}

	root, err := rifx.Parse(bytes.NewReader(buf.Bytes()))
	if err != nil {
		fmt.Println("rifx.Parse err:", err)
		os.Exit(1)
	}

	var walk func(c *rifx.Chunk, depth int)
	walk = func(c *rifx.Chunk, depth int) {
		ind := ""
		for i := 0; i < depth; i++ {
			ind += "  "
		}
		if c.IsList() {
			fmt.Printf("%s[%s %s]\n", ind, c.ID, c.FormType)
		} else {
			label := string(c.ID[:])
			if c.ID == rifx.IDTdmn {
				label += " = " + trimZ(c.Data)
			}
			fmt.Printf("%s%s (%d B)\n", ind, label, len(c.Data))
		}
		if depth > 8 {
			return
		}
		for _, ch := range c.Children {
			walk(ch, depth+1)
		}
	}
	walk(root, 0)
}

func trimZ(b []byte) string {
	n := len(b)
	for n > 0 && b[n-1] == 0 {
		n--
	}
	return string(b[:n])
}
