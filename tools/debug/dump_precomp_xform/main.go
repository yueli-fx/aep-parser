// Dump precomp-layer Scale + Rotation static values for a named comp, to feed the
// from-scratch gen. Usage: go run ./tools/debug/dump_precomp_xform file.aep "comp name"
package main

import (
	"fmt"
	"os"

	aep "github.com/example/aep-parser/internal/aep"
)

func findGroup(g *aep.AEPropertyGroup, mn string) *aep.AEPropertyGroup {
	if g == nil {
		return nil
	}
	for _, c := range g.Children {
		if cg, ok := c.(*aep.AEPropertyGroup); ok && cg.MatchName == mn {
			return cg
		}
	}
	return nil
}
func findProp(g *aep.AEPropertyGroup, mn string) *aep.Property {
	if g == nil {
		return nil
	}
	for _, c := range g.Children {
		if p, ok := c.(*aep.Property); ok && p.MatchName == mn {
			return p
		}
	}
	return nil
}

func main() {
	p, err := aep.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	comp := p.CompositionByName(os.Args[2])
	if comp == nil {
		fmt.Println("no comp")
		os.Exit(1)
	}
	for i, l := range comp.Layers {
		tg := findGroup(l.PropertyTree(), "ADBE Transform Group")
		var sc, rot, anc, pos any
		if s := findProp(tg, "ADBE Scale"); s != nil {
			sc = s.StaticValue
		}
		if r := findProp(tg, "ADBE Rotate Z"); r != nil {
			rot = r.StaticValue
		}
		if a := findProp(tg, "ADBE Anchor Point"); a != nil {
			anc = a.StaticValue
		}
		if po := findProp(tg, "ADBE Position"); po != nil {
			pos = po.StaticValue
		}
		fmt.Printf("L%d %q: Scale=%v Rotate Z=%v Anchor=%v Position(static)=%v\n", i, l.Name, sc, rot, anc, pos)
	}
}
