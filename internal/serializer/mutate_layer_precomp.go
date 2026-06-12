// Public-API entry for precomp / nested-composition layer creation
// (`NewPrecompLayer`).
//
// A precomp layer is an ordinary AV layer whose ldta @0x28 SourceID points at
// an existing CompItem (instead of a footage item) — the source composition
// already lives in the project, so unlike Solid/Null/Adjustment there is no
// footage item to synthesize/import. We embed a single AE-native precomp AV
// Layr (templates/layer_precomp_body.bin, extracted from re_precomp.aep — a
// 4-child Layr identical in shape to camera/light, only its SourceID marks it
// as a precomp) and splice it via the same machinery as Camera/Light
// (newTemplatedLayer), then repoint its SourceID at the user's child comp.
package serializer

import (
	_ "embed"
	"fmt"

	"github.com/example/aep-parser/internal/scene"
)

//go:embed templates/layer_precomp_body.bin
var layerPrecompBodyBytes []byte

// NewPrecompLayer adds a layer to `parent` whose source is the composition
// `child` (a nested / pre-composed comp). Both comps must belong to the same
// project, and the nesting must not create a cycle.
// (Full contract lives on the aep.NewPrecompLayer facade — docgen source.)
func NewPrecompLayer(parent, child *Composition, name string) (*Layer, error) {
	if parent == nil || child == nil {
		return nil, fmt.Errorf("NewPrecompLayer: parent and child compositions cannot be nil")
	}
	if name == "" {
		return nil, fmt.Errorf("NewPrecompLayer: layer name cannot be empty")
	}
	if parent == child {
		return nil, fmt.Errorf("NewPrecompLayer: a composition cannot contain itself")
	}
	if child.ID == 0 {
		return nil, fmt.Errorf("NewPrecompLayer: child composition %q has no item ID (built outside parser?)", child.Name)
	}
	pProj := scene.CompositionProj(parent)
	if pProj == nil {
		return nil, fmt.Errorf("NewPrecompLayer: parent composition %q has no project back-ref", parent.Name)
	}
	if scene.CompositionProj(child) != pProj {
		return nil, fmt.Errorf("NewPrecompLayer: child composition %q is not in the same project as %q", child.Name, parent.Name)
	}
	// Cycle guard: nesting child under parent must not make parent reachable
	// from child (AE rejects circular comp references). DFS over child's
	// transitive precomp sources.
	if compReachable(child, parent, map[uint32]bool{}) {
		return nil, fmt.Errorf("NewPrecompLayer: nesting %q in %q would create a circular composition reference", child.Name, parent.Name)
	}

	// Clone + splice the AE-native precomp Layr (re-homes time span to parent
	// duration, ID/parent/name patched, atomic rollback) — identical path to
	// Camera/Light, then repoint the SourceID at the child comp.
	clone, err := newTemplatedLayer(parent, name, layerPrecompBodyBytes, LayerTypeAV)
	if err != nil {
		return nil, fmt.Errorf("NewPrecompLayer: %w", err)
	}
	if lb := layerBack(clone); lb != nil {
		if err := lb.SetSource(child.ID); err != nil {
			return nil, fmt.Errorf("NewPrecompLayer: set source: %w", err)
		}
	}
	clone.SourceID = child.ID
	return clone, nil
}

// compReachable reports whether `target` is reachable from `from` by following
// precomp source edges (a layer whose source is a composition). Used to reject
// circular nesting. `seen` guards against pre-existing cycles in the graph.
func compReachable(from, target *Composition, seen map[uint32]bool) bool {
	if from == nil || seen[from.ID] {
		return false
	}
	seen[from.ID] = true
	for _, l := range from.Layers {
		src := l.SourceComposition()
		if src == nil {
			continue
		}
		if src == target || compReachable(src, target, seen) {
			return true
		}
	}
	return false
}
