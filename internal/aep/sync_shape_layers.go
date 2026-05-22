// internal/aep/sync_shape_layers.go
//
// V2.2 Phase 4 — write-time bridge from runtime ShapeLayer state to the
// owning Layr chunk. NewShapeLayer pre-lowers an empty Layr at construction
// time; subsequent VectorGroup mutations (AddRect, SetSize, ...) only touch
// the runtime tree. This pass walks every layer where shapeRootGroup is
// populated AND a back-ref'd Layr chunk exists, re-lowers, and overwrites
// the chunk's Children in place. Chunk sizes recompute at WriteAEP time
// (rifx.Chunk.Write does Children → bytes from scratch).
package aep

import (
	"fmt"
)

// syncShapeLayerChunks re-lowers every shape layer that has a runtime
// VectorGroup + a back-ref'd Layr chunk, overwriting the chunk's Children
// in place. Safe to call on parse-then-write paths too: parsed layers
// have shapeRootGroup populated by parseLayer/hydrateShapeNodes and
// layrList populated by parseLayer; if the user never wrapped/mutated,
// the re-lowering reconstructs structurally-equivalent chunks (Phase 4
// roundtrip is the byte-fidelity gate).
func (p *Project) syncShapeLayerChunks() error {
	for _, c := range p.Compositions {
		if err := p.syncCompositionShapeLayers(c); err != nil {
			return err
		}
	}
	return nil
}

func (p *Project) syncCompositionShapeLayers(c *Composition) error {
	for _, l := range c.Layers {
		if l.Type != LayerTypeShape {
			continue
		}
		if l.shapeRootGroup == nil || l.layrList == nil {
			continue
		}
		if !l.shapeDirty {
			continue // parser-loaded, not user-mutated; on-disk chunks are authoritative
		}
		s := WrapShapeLayer(l)
		ctx := &lowerCtx{
			tickRate:     c.TickRate,
			capabilities: Capabilities(p.target),
			nextLayerID:  p.allocItemID,
		}
		fresh, err := lowerShapeLayer(s, ctx)
		if err != nil {
			return fmt.Errorf("layer %q (ID %d): %w", l.Name, l.ID, err)
		}
		// Overwrite the existing Layr LIST's children in place. The
		// chunk pointer stays the same, so c.itemList.Children
		// (and any other holder of the layrList pointer) remains
		// consistent.
		l.layrList.Children = fresh.Children
	}
	return nil
}
