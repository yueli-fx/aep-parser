// Dump ldta bytes for each layer in re_trackmatte_ae24.aep.
// Usage: go run ./tmp_debug/dump_ldta_trackmatte
package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func main() {
	path := "test_data/re_trackmatte_ae24.aep"
	if len(os.Args) > 1 {
		path = os.Args[1]
	}
	compFilter := "RE_TRACKMATTE"
	if len(os.Args) > 2 {
		compFilter = os.Args[2]
	}
	prefixFilter := ""
	if len(os.Args) > 3 {
		prefixFilter = os.Args[3]
	}

	p, err := aep.Open(path)
	if err != nil {
		panic(err)
	}
	for _, c := range p.Compositions {
		if c.Name != compFilter {
			continue
		}
		fmt.Printf("### comp=%q\n", c.Name)
		for _, l := range c.Layers {
			if prefixFilter != "" && !strings.HasPrefix(l.Name, prefixFilter) {
				continue
			}
			fmt.Printf("=== %q  id=%d  source=%d  parent=%d ===\n", l.Name, l.ID, l.SourceID, l.ParentID)
			ldta := l.LdtaRawBytes()
			if ldta == nil {
				fmt.Println("(no ldta)")
				continue
			}
			fmt.Printf("len(ldta)=%d\n", len(ldta))
			fmt.Println(hex.Dump(ldta))
		}
	}
}
