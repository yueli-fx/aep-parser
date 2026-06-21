// Public-API entry for replacing a templated layer's Transform Group with a
// caller-built LayerTransform — the from-scratch path for ANIMATED layer
// transform on layers that are NOT generic ShapeLayers (text / precomp / footage /
// solid / null …). Those layers come from embedded templates whose Transform
// Group ELIDES default channels (AE default-omission — see incident
// transform-group-default-omission): a fresh NewTextLayer has no "ADBE Position"
// / "ADBE Opacity" / "ADBE Anchor Point" chunk at all, only Position_0/_1,
// Orientation, RotateX/Y, Envir. So neither SetPosition (no materialized prop)
// nor AnimateScalarKeyframes/InsertKeyframe (no cdat to convert) can reach them.
//
// SetLayerTransform sidesteps that by lowering a full canonical Transform Group
// (the exact path generic ShapeLayers ship, already AE 2020/2025 gated) from a
// LayerTransform the caller fills with static values and/or keyframes, then
// swapping it in for the layer's elided Transform Group body. rifx recomputes
// the ancestor LIST sizes on write (same as the text-animator splice).
package serializer

import (
	"fmt"

	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

// SetLayerTransform replaces layer's Transform Group with a lowering of t.
// (Full contract lives on the aep.SetLayerTransform facade — docgen source.)
func SetLayerTransform(layer *Layer, t *LayerTransform) error {
	if layer == nil {
		return fmt.Errorf("SetLayerTransform: nil layer")
	}
	if t == nil {
		return fmt.Errorf("SetLayerTransform: nil transform")
	}
	lb := layerBack(layer)
	if lb == nil || lb.layrList == nil {
		return fmt.Errorf("SetLayerTransform: layer %q has no chunk back-ref (built outside parser? round-trip via aep.Reopen first)", layer.Name)
	}
	var outer *rifx.Chunk
	for _, ch := range lb.layrList.Children {
		if ch.IsList() && ch.FormType == rifx.IDTdgp {
			outer = ch
			break
		}
	}
	if outer == nil {
		return fmt.Errorf("SetLayerTransform: layer %q has no property tree", layer.Name)
	}
	tickRate := aeLegacyTimeBase
	if comp := scene.LayerComp(layer); comp != nil && comp.TickRate > 0 {
		tickRate = comp.TickRate
	}
	wrapper, err := lowerLayerTransform(t, &lowerCtx{tickRate: tickRate})
	if err != nil {
		return fmt.Errorf("SetLayerTransform: %w", err)
	}
	// wrapper.Children[0] = tdmn("ADBE Transform Group"); [1:] = the body chunks.
	newBody := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp, Children: wrapper.Children[1:]}
	for i, ch := range outer.Children {
		if tdmnName(ch) == "ADBE Transform Group" && i+1 < len(outer.Children) && outer.Children[i+1].IsList() {
			outer.Children[i+1] = newBody
			return nil
		}
	}
	return fmt.Errorf("SetLayerTransform: layer %q has no Transform Group", layer.Name)
}
