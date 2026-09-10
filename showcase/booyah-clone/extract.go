// showcase/booyah-clone/extract.go — tiny readers over the original's
// raw property tree (the value-oracle side). gen_<comp>.go uses these to pull
// shape/keyframe/param values out of the parsed original and feed them to our
// from-scratch write API. The original is read-only; nothing here copies bytes.
package main

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// findGroup returns the first child group of g whose MatchName matches, or nil.
func findGroup(g *aep.AEPropertyGroup, matchName string) *aep.AEPropertyGroup {
	if g == nil {
		return nil
	}
	for _, c := range g.Children {
		if cg, ok := c.(*aep.AEPropertyGroup); ok && cg.MatchName == matchName {
			return cg
		}
	}
	return nil
}

// findProp returns the first child property of g whose MatchName matches, or nil.
func findProp(g *aep.AEPropertyGroup, matchName string) *aep.Property {
	if g == nil {
		return nil
	}
	for _, c := range g.Children {
		if p, ok := c.(*aep.Property); ok && p.MatchName == matchName {
			return p
		}
	}
	return nil
}

// toFloats coerces a parser value (any holding a float slice/array) to []float64.
func toFloats(v any) []float64 {
	switch s := v.(type) {
	case []float64:
		return s
	case [2]float64:
		return s[:]
	case [3]float64:
		return s[:]
	case [4]float64:
		return s[:]
	}
	panic(fmt.Sprintf("toFloats: unhandled value type %T (%v)", v, v))
}

func to2(v any) [2]float64 { f := toFloats(v); return [2]float64{f[0], f[1]} }
func to4(v any) [4]float64 { f := toFloats(v); return [4]float64{f[0], f[1], f[2], f[3]} }

// toScalar coerces a parser value to a single float64 (a bare float, or the first
// component of a 1-element slice/array).
func toScalar(v any) float64 {
	switch s := v.(type) {
	case float64:
		return s
	case []float64:
		return s[0]
	case [1]float64:
		return s[0]
	}
	return toFloats(v)[0]
}
