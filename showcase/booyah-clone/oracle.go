// showcase/booyah-clone/oracle.go — the ORIGINAL "Booyah Glitch" project
// opened READ-ONLY as a value oracle. Replication reads structure/values from here
// and rebuilds every chunk via our own write API (New*/Add*/Set*/Animate*). The
// original is NEVER a byte source — copying its chunks would be round-trip, not
// replication (spec 2026-06-19-booyah-glitch-full-replication §4).
//
// Read side is rich: p.Compositions / c.Layers / l.Effects / l.Masks /
// l.PropertyTree() / fx.Parameters / pr.StaticValue / pr.Expression / pr.Keyframes
// (see cmd/aepdissect for the canonical read patterns).
package main

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

const originalPath = "data/samples/motionbox/glitch/booyah-glitch/Booyah Glitch.aep"

// oracle wraps the parsed original for value lookups during reconstruction.
type oracle struct {
	proj *aep.Project
}

func openOriginal() (*oracle, error) {
	p, err := aep.Open(originalPath)
	if err != nil {
		return nil, fmt.Errorf("open original %q: %w", originalPath, err)
	}
	return &oracle{proj: p}, nil
}

// comp returns the original composition by name (nil if absent).
func (o *oracle) comp(name string) *aep.Composition {
	return o.proj.CompositionByName(name)
}

// mustComp is comp() that panics if the comp is missing — reconstruction can't
// proceed without its value source, so fail loud.
func (o *oracle) mustComp(name string) *aep.Composition {
	c := o.comp(name)
	if c == nil {
		panic(fmt.Sprintf("oracle: original comp %q not found", name))
	}
	return c
}
