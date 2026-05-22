// internal/aep/new_layer.go
//
// V2.2 Phase 3 — public-API entry for ShapeLayer creation. Atomic mutation
// pattern mirrors V2.1 Project.NewComposition (Inv-10 atomicity, Inv-11
// warnings-as-failure): lower runtime → chunk, snapshot mutable state,
// commit, rollback on any warning produced by downstream parse.
//
// Phase 3 wires only the construction side; parse-after-build full hydration
// (typed wrapper rebuilt from chunks) lands in Phase 4 roundtrip closure.
// For now the wrapper returned to the caller is the freshly constructed
// runtime instance, and comp.Layers carries the embedded *Layer base so V1
// readers (Layer.Name / Type / etc.) work.
package aep

import "fmt"

// NewShapeLayer adds a new empty ShapeLayer to the composition.
//
// Required:
//
//	name — non-empty string (matches NewComposition validation contract)
//
// Returns the typed *ShapeLayer wrapper; the embedded *Layer is also
// appended to comp.Layers so V1 lookup paths (Composition.LayerByID /
// LayerByName) work immediately. ID is auto-assigned via the project's
// monotonic item-ID counter (Inv-9 — never reused; layer IDs share the
// item-ID namespace per V1 parser convention).
//
// Atomic mutation (Inv-10): if lowering fails, or downstream parse emits
// any warning, all state mutated by this call is rolled back to the
// pre-call snapshot before the error is returned.
func (c *Composition) NewShapeLayer(name string) (*ShapeLayer, error) {
	if name == "" {
		return nil, fmt.Errorf("ShapeLayer name cannot be empty")
	}
	if c.proj == nil {
		return nil, fmt.Errorf("internal: comp has no project back-ref")
	}
	if c.itemList == nil {
		return nil, fmt.Errorf("internal: comp has no itemList chunk")
	}

	// 1. Construct runtime ShapeLayer (Phase 1 types).
	base := &Layer{
		Type: LayerTypeShape,
		Name: name,
		ID:   c.proj.allocItemID(), // layer ID shares item-ID namespace (V1 convention)
		comp: c,
	}
	s := WrapShapeLayer(base)

	// 2. Lower runtime → LIST(Layr) chunk (Phase 2).
	ctx := &lowerCtx{
		tickRate:     c.TickRate,
		capabilities: Capabilities(c.proj.target),
		nextLayerID:  c.proj.allocItemID,
	}
	layrChunk, err := lowerShapeLayer(s, ctx)
	if err != nil {
		return nil, fmt.Errorf("lower ShapeLayer: %w", err)
	}

	// 3. Snapshot pre-mutation state (Inv-10 atomicity).
	oldChildLen := len(c.itemList.Children)
	oldLayersLen := len(c.Layers)
	oldWarningsLen := len(c.proj.Warnings)

	// 4. Commit: append Layr to itemList + base layer to typed index.
	//    base.layrList back-ref lets the write-time sync pass
	//    (syncShapeLayerChunks) re-lower the runtime tree into this same
	//    chunk in place before WriteAEP serializes it.
	base.layrList = layrChunk
	base.shapeDirty = true // gate for syncShapeLayerChunks
	c.itemList.Children = append(c.itemList.Children, layrChunk)
	c.Layers = append(c.Layers, base)

	// 5. Warnings-as-failure (Inv-11): if any warnings appeared, rollback.
	// Phase 3 doesn't re-parse, so this is a defensive guard for Phase 4 to
	// rely on (re-parse closed loop arrives there).
	if len(c.proj.Warnings) != oldWarningsLen {
		c.itemList.Children = c.itemList.Children[:oldChildLen]
		c.Layers = c.Layers[:oldLayersLen]
		newWarnings := append([]string(nil), c.proj.Warnings[oldWarningsLen:]...)
		c.proj.Warnings = c.proj.Warnings[:oldWarningsLen]
		return nil, fmt.Errorf("internal: NewShapeLayer produced %d parser warning(s): %v", len(newWarnings), newWarnings)
	}

	return s, nil
}
