package aep

import (
	"fmt"

	"github.com/example/aep-parser/internal/rifx"
)

// PropertyBase structural ops (py-aep parity P3 §3C): Remove / MoveTo on an
// AEPropertyGroup that is a direct child of an INDEXED_GROUP parent.
//
// Behavior contract (RE'd from AE 2020, see
// incidents/property-indexed-group-structural-re.md):
//
//   - AE refuses .remove()/.moveTo() unless the node's parentProperty is an
//     INDEXED_GROUP (the ScriptingAPI throws "无法使用 remove 方法... 因为父
//     属性不是 INDEXED_GROUP"). INDEXED_GROUP is the fixed match-name set
//     below; every other group is NAMED_GROUP and its children are fixed.
//   - The indexed groups AE ships (Effect Parade / Mask Parade / Root Vectors
//     Group / Text Animators) hold their children as group payloads (sspc for
//     effects, tdgp for masks/shape/animators) — never bare leaves — so these
//     ops live on *AEPropertyGroup, not *Property.
//   - Chunk mechanics are pure tdmn+payload pair splice/reorder within the
//     parent group LIST; AE touches NO count/index chunk (baseline→after
//     diffs of re_property_struct_*.aep confirm 9→7 on remove, 9↔9 on move).
//
// Alpha: Effect-Parade remove/move are AE 2020 + AE 2025 ship-gate green.
// Mask Parade / Root Vectors Group / Text Animators share the identical
// mechanic and pass Go round-trip, but are not yet ship-gated — treat as
// Alpha until their own gate runs.

// indexedGroupMatchNames is AE's fixed set of INDEXED_GROUP match-names
// (mirrors py-aep _INDEXED_GROUP_MATCH_NAMES). Membership is the sole
// predicate for whether a group's direct children may be removed / reordered /
// duplicated.
var indexedGroupMatchNames = map[string]bool{
	"ADBE Effect Parade":      true,
	"ADBE Mask Parade":        true,
	"ADBE Effect Mask Parade": true,
	"ADBE Text Animators":     true,
	"ADBE Root Vectors Group": true,
}

// IsIndexedGroup reports whether this group is one of AE's INDEXED_GROUP
// containers, i.e. whether its direct children support structural Remove /
// MoveTo / Duplicate. Named groups (Transform, Material Options, …) and leaf
// properties return false.
func (g *AEPropertyGroup) IsIndexedGroup() bool {
	return g != nil && indexedGroupMatchNames[g.MatchName]
}

// ownerLayer walks up to the synthetic property-tree root and returns the
// owning Layer, or nil for groups built outside the parser.
func (g *AEPropertyGroup) ownerLayer() *Layer {
	for cur := g; cur != nil; cur = cur.parent {
		if cur.layer != nil {
			return cur.layer
		}
	}
	return nil
}

// childPayloadChunk returns the on-disk payload chunk for a direct child of an
// indexed group: the group LIST for a subgroup, or the tdbs LIST for a leaf.
// nil when the child carries no parser back-ref.
func childPayloadChunk(c PropertyBase) *rifx.Chunk {
	switch v := c.(type) {
	case *AEPropertyGroup:
		return v.chunk
	case *Property:
		if v.back != nil {
			return v.back.tdbs
		}
	}
	return nil
}

// childTdmnPayload locates a direct child's (tdmn, payload) chunk pair inside
// parent.chunk.Children by payload pointer identity. ok is false when the
// parent has no chunk, the payload isn't found, or it isn't preceded by a
// tdmn.
func (parent *AEPropertyGroup) childTdmnPayload(child PropertyBase) (tdmn, payload *rifx.Chunk, ok bool) {
	if parent == nil || parent.chunk == nil {
		return nil, nil, false
	}
	pc := childPayloadChunk(child)
	if pc == nil {
		return nil, nil, false
	}
	pi := indexOfChunk(parent.chunk.Children, pc)
	if pi < 1 {
		return nil, nil, false
	}
	t := parent.chunk.Children[pi-1]
	if t.ID != rifx.IDTdmn {
		return nil, nil, false
	}
	return t, pc, true
}

// rebuildIndexedGroupChunk re-emits parent.chunk.Children from the current
// parent.Children scene order. The leading header chunks (tdsb / tdsn) and the
// trailing "ADBE Group End" sentinel are preserved verbatim; each scene child
// contributes its original (tdmn, payload) chunk pair — the SAME pointers, so
// any opaque/undecoded content rides along unchanged (CLAUDE.md #5). pairs
// supplies the tdmn+payload for each child (captured before any scene-order
// mutation). Returns an error without mutating when a child's pair is missing.
func (parent *AEPropertyGroup) rebuildIndexedGroupChunk(pairs map[PropertyBase][2]*rifx.Chunk) error {
	children := parent.chunk.Children

	// Boundaries: prefix = chunks before the first child's tdmn; suffix =
	// chunks after the last child's payload (the Group End sentinel + any
	// trailing chunks). Derived from the current (pre-rebuild) pair positions.
	first, last := len(children), -1
	for _, c := range parent.Children {
		pair, ok := pairs[c]
		if !ok {
			return fmt.Errorf("rebuildIndexedGroupChunk: child %q has no captured chunk pair", c.PropertyMatchName())
		}
		ti := indexOfChunk(children, pair[0])
		pi := indexOfChunk(children, pair[1])
		if ti < 0 || pi < 0 {
			return fmt.Errorf("rebuildIndexedGroupChunk: child %q chunk pair not in group LIST", c.PropertyMatchName())
		}
		if ti < first {
			first = ti
		}
		if pi > last {
			last = pi
		}
	}
	if last < 0 {
		// No children to anchor against — leave the chunk untouched.
		return nil
	}

	prefix := append([]*rifx.Chunk(nil), children[:first]...)
	suffix := append([]*rifx.Chunk(nil), children[last+1:]...)

	rebuilt := make([]*rifx.Chunk, 0, len(prefix)+2*len(parent.Children)+len(suffix))
	rebuilt = append(rebuilt, prefix...)
	for _, c := range parent.Children {
		pair := pairs[c]
		rebuilt = append(rebuilt, pair[0], pair[1])
	}
	rebuilt = append(rebuilt, suffix...)
	parent.chunk.Children = rebuilt
	return nil
}

// capturePairs snapshots each current scene child's (tdmn, payload) chunk pair
// from parent.chunk.Children, keyed by the scene node. Children whose pair
// can't be located are omitted; callers validate completeness as needed.
func (parent *AEPropertyGroup) capturePairs() map[PropertyBase][2]*rifx.Chunk {
	pairs := make(map[PropertyBase][2]*rifx.Chunk, len(parent.Children))
	for _, c := range parent.Children {
		if t, p, ok := parent.childTdmnPayload(c); ok {
			pairs[c] = [2]*rifx.Chunk{t, p}
		}
	}
	return pairs
}

// flatSliceFor returns the layer's flat slice that mirrors an indexed group
// 1:1 in order (Effects for the Effect Parade, Masks for the Mask Parade), and
// a setter to write it back. Returns ok=false for indexed groups with no flat
// mirror (Root Vectors Group / Text Animators) — their on-disk chunk tree is
// the source of truth and no flat sync is needed.
func flatSliceFor(parent *AEPropertyGroup, layer *Layer) (get func() int, reorder func(newOrder []int), drop func(i int), ok bool) {
	if layer == nil {
		return nil, nil, nil, false
	}
	switch parent.MatchName {
	case "ADBE Effect Parade":
		return func() int { return len(layer.Effects) },
			func(order []int) { layer.Effects = reorderEffects(layer.Effects, order) },
			func(i int) { layer.Effects = dropEffect(layer.Effects, i) },
			true
	case "ADBE Mask Parade":
		return func() int { return len(layer.Masks) },
			func(order []int) { layer.Masks = reorderMasks(layer.Masks, order) },
			func(i int) { layer.Masks = dropMask(layer.Masks, i) },
			true
	}
	return nil, nil, nil, false
}

func reorderEffects(s []*Effect, order []int) []*Effect {
	if len(order) != len(s) {
		return s
	}
	out := make([]*Effect, len(s))
	for newIdx, oldIdx := range order {
		out[newIdx] = s[oldIdx]
	}
	return out
}

func dropEffect(s []*Effect, i int) []*Effect {
	if i < 0 || i >= len(s) {
		return s
	}
	return append(append([]*Effect(nil), s[:i]...), s[i+1:]...)
}

func reorderMasks(s []*Mask, order []int) []*Mask {
	if len(order) != len(s) {
		return s
	}
	out := make([]*Mask, len(s))
	for newIdx, oldIdx := range order {
		out[newIdx] = s[oldIdx]
	}
	return out
}

func dropMask(s []*Mask, i int) []*Mask {
	if i < 0 || i >= len(s) {
		return s
	}
	return append(append([]*Mask(nil), s[:i]...), s[i+1:]...)
}

// Remove deletes this group from its parent INDEXED_GROUP. The receiver must
// be a direct child of an indexed group (Effect Parade / Mask Parade / Root
// Vectors Group / Text Animators); Remove returns an error otherwise, mirroring
// AE's ScriptingAPI refuse.
//
// Atomic: snapshots the parent chunk LIST, scene children, the mirrored flat
// slice, and Project.Warnings; on any new parser warning everything rolls back
// and the warnings are returned as an error.
//
// Alpha — see file header for ship-gate status.
func (g *AEPropertyGroup) Remove() error {
	parent := g.parent
	if parent == nil {
		return fmt.Errorf("Remove: property group %q has no parent (root or built outside parser)", g.MatchName)
	}
	if !parent.IsIndexedGroup() {
		return fmt.Errorf("Remove: parent group %q is not an INDEXED_GROUP; only children of indexed groups can be removed", parent.MatchName)
	}
	idx := parent.PropertyIndex(g)
	if idx < 0 {
		return fmt.Errorf("Remove: group %q not found among parent %q children", g.MatchName, parent.MatchName)
	}
	if _, _, ok := parent.childTdmnPayload(g); !ok {
		return fmt.Errorf("Remove: group %q chunk pair not located in parent LIST", g.MatchName)
	}

	layer := parent.ownerLayer()
	pairs := parent.capturePairs()

	// Snapshot for rollback.
	oldChunkChildren := append([]*rifx.Chunk(nil), parent.chunk.Children...)
	oldSceneChildren := append([]PropertyBase(nil), parent.Children...)
	var oldEffects []*Effect
	var oldMasks []*Mask
	if layer != nil {
		oldEffects = append([]*Effect(nil), layer.Effects...)
		oldMasks = append([]*Mask(nil), layer.Masks...)
	}
	oldWarningsLen := warningsLen(layer)

	// Scene: drop the child.
	parent.Children = filterPropertyBase(parent.Children, map[PropertyBase]bool{g: true})
	// Chunk: rebuild from the reduced scene order.
	if err := parent.rebuildIndexedGroupChunk(pairs); err != nil {
		parent.chunk.Children = oldChunkChildren
		parent.Children = oldSceneChildren
		return fmt.Errorf("Remove: %w", err)
	}
	// Flat mirror.
	if _, _, drop, ok := flatSliceFor(parent, layer); ok {
		drop(idx)
	}

	if newWarn := newWarningsSince(layer, oldWarningsLen); len(newWarn) > 0 {
		parent.chunk.Children = oldChunkChildren
		parent.Children = oldSceneChildren
		if layer != nil {
			layer.Effects = oldEffects
			layer.Masks = oldMasks
		}
		rollbackWarnings(layer, oldWarningsLen)
		return fmt.Errorf("Remove: produced %d parser warning(s), rolled back: %v", len(newWarn), newWarn)
	}
	return nil
}

// MoveTo reorders this group to position index (0-based) among its parent
// INDEXED_GROUP's children. index is clamped-checked against the current child
// count. Mirrors AE's PropertyBase.moveTo (which is 1-based; the Go API is
// 0-based per project convention).
//
// Alpha — see file header for ship-gate status.
func (g *AEPropertyGroup) MoveTo(index int) error {
	parent := g.parent
	if parent == nil {
		return fmt.Errorf("MoveTo: property group %q has no parent (root or built outside parser)", g.MatchName)
	}
	if !parent.IsIndexedGroup() {
		return fmt.Errorf("MoveTo: parent group %q is not an INDEXED_GROUP; only children of indexed groups can be reordered", parent.MatchName)
	}
	cur := parent.PropertyIndex(g)
	if cur < 0 {
		return fmt.Errorf("MoveTo: group %q not found among parent %q children", g.MatchName, parent.MatchName)
	}
	n := len(parent.Children)
	if index < 0 || index >= n {
		return fmt.Errorf("MoveTo: index %d out of range (have %d children)", index, n)
	}
	if index == cur {
		return nil
	}
	if _, _, ok := parent.childTdmnPayload(g); !ok {
		return fmt.Errorf("MoveTo: group %q chunk pair not located in parent LIST", g.MatchName)
	}

	layer := parent.ownerLayer()
	pairs := parent.capturePairs()

	oldChunkChildren := append([]*rifx.Chunk(nil), parent.chunk.Children...)
	oldSceneChildren := append([]PropertyBase(nil), parent.Children...)
	var oldEffects []*Effect
	var oldMasks []*Mask
	if layer != nil {
		oldEffects = append([]*Effect(nil), layer.Effects...)
		oldMasks = append([]*Mask(nil), layer.Masks...)
	}
	oldWarningsLen := warningsLen(layer)

	// Scene: move element cur → index. Build the new order as a permutation of
	// the OLD indices so the flat mirror can apply the identical permutation.
	order := moveIndexPermutation(n, cur, index)
	newChildren := make([]PropertyBase, n)
	for newIdx, oldIdx := range order {
		newChildren[newIdx] = oldSceneChildren[oldIdx]
	}
	parent.Children = newChildren

	if err := parent.rebuildIndexedGroupChunk(pairs); err != nil {
		parent.chunk.Children = oldChunkChildren
		parent.Children = oldSceneChildren
		return fmt.Errorf("MoveTo: %w", err)
	}
	if _, reorder, _, ok := flatSliceFor(parent, layer); ok {
		reorder(order)
	}

	if newWarn := newWarningsSince(layer, oldWarningsLen); len(newWarn) > 0 {
		parent.chunk.Children = oldChunkChildren
		parent.Children = oldSceneChildren
		if layer != nil {
			layer.Effects = oldEffects
			layer.Masks = oldMasks
		}
		rollbackWarnings(layer, oldWarningsLen)
		return fmt.Errorf("MoveTo: produced %d parser warning(s), rolled back: %v", len(newWarn), newWarn)
	}
	return nil
}

// moveIndexPermutation returns the new→old index mapping for moving the element
// at `from` to position `to` in a slice of length n (other elements keep their
// relative order).
func moveIndexPermutation(n, from, to int) []int {
	base := make([]int, 0, n)
	for i := 0; i < n; i++ {
		if i != from {
			base = append(base, i)
		}
	}
	out := make([]int, 0, n)
	out = append(out, base[:to]...)
	out = append(out, from)
	out = append(out, base[to:]...)
	return out
}

// warningsLen / newWarningsSince / rollbackWarnings centralize the
// Project.Warnings bookkeeping used by the atomic structural ops, tolerating a
// nil layer / comp / project chain (groups built outside the parser).
func warningsLen(layer *Layer) int {
	if layer != nil && layer.comp != nil && layer.comp.proj != nil {
		return len(layer.comp.proj.Warnings)
	}
	return 0
}

func newWarningsSince(layer *Layer, oldLen int) []string {
	if layer == nil || layer.comp == nil || layer.comp.proj == nil {
		return nil
	}
	w := layer.comp.proj.Warnings
	if len(w) <= oldLen {
		return nil
	}
	return append([]string(nil), w[oldLen:]...)
}

func rollbackWarnings(layer *Layer, oldLen int) {
	if layer == nil || layer.comp == nil || layer.comp.proj == nil {
		return
	}
	layer.comp.proj.Warnings = layer.comp.proj.Warnings[:oldLen]
}
