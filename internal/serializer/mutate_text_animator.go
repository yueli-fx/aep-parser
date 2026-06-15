// Public-API entry for creating a text animator from scratch — the kinetic-
// typography engine (per-character Opacity / Position / etc. driven by a Range
// Selector).
//
// A text layer's animators live in the "ADBE Text Animators" indexed group,
// itself nested inside the layer's "ADBE Text Properties" named group (NOT in
// the btdk document — btdk is the character string + layout cache; animators
// are property-tree leaves). Each animator is a (tdmn "ADBE Text Animator",
// LIST:tdgp) pair holding a "ADBE Text Selectors" sub-group (with one Range
// Selector: Percent Start/End/Offset) + an "ADBE Text Animator Properties"
// sub-group (the animatable leaves, e.g. "ADBE Text Opacity").
//
// AE elides every default-valued leaf, so the embedded template was authored
// with Start/End/Offset + Opacity all set non-default (RE_TXANIM_MODE=template
// in re_text_animator.jsx) — that materializes every cdat slot, which
// AddTextOpacityAnimator then overwrites with the caller's values (length-
// preserving f64 @ cdat[0:8], the same scalar layout shape/effect leaves use).
//
// Fresh text layers (NewTextLayer) carry NO Animators group at all (just
// Document / Path Options / More Options), so the first animator splices the
// whole "ADBE Text Animators" group into Text Properties (before its Group End
// sentinel); subsequent animators append into the existing group. The whole
// embed-AE-bytes + (tdmn, payload) splice is the same vein AddEffect /
// DuplicatePropertyGroup are ship-gate-green with.
package serializer

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"fmt"
	"math"
	"sync"

	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

//go:embed templates/text_animators_opacity_body.bin
var textAnimatorsOpacityBody []byte

const (
	matchNameTextAnimators     = "ADBE Text Animators"
	matchNameTextAnimator      = "ADBE Text Animator"
	matchNameTextPercentStart  = "ADBE Text Percent Start"
	matchNameTextPercentEnd    = "ADBE Text Percent End"
	matchNameTextPercentOffset = "ADBE Text Percent Offset"
	matchNameTextOpacity       = "ADBE Text Opacity"
)

var (
	textAnimatorsTmplOnce  sync.Once
	textAnimatorsTmplChunk *rifx.Chunk
	textAnimatorsTmplErr   error
)

// textAnimatorsTemplate parses the embedded "ADBE Text Animators" group body
// once (cached). Callers deep-clone before splicing so spliced chunks never
// alias the cache.
func textAnimatorsTemplate() (*rifx.Chunk, error) {
	textAnimatorsTmplOnce.Do(func() {
		c, e := rifx.ReadChunk(bytes.NewReader(textAnimatorsOpacityBody))
		if e != nil {
			textAnimatorsTmplErr = fmt.Errorf("parse text animators template: %w", e)
			return
		}
		textAnimatorsTmplChunk = c
	})
	return textAnimatorsTmplChunk, textAnimatorsTmplErr
}

// tdmnName returns a tdmn chunk's match-name (NUL-trimmed), or "" for non-tdmn.
func tdmnName(c *rifx.Chunk) string {
	if c.ID != rifx.IDTdmn {
		return ""
	}
	return string(bytes.TrimRight(c.Data, "\x00"))
}

// childGroupChunk returns the LIST:tdgp following the first DIRECT child
// tdmn == matchName in parent.Children, or nil.
func childGroupChunk(parent *rifx.Chunk, matchName string) *rifx.Chunk {
	for i, ch := range parent.Children {
		if tdmnName(ch) == matchName && i+1 < len(parent.Children) {
			if next := parent.Children[i+1]; next.IsList() {
				return next
			}
		}
	}
	return nil
}

// groupEndIndex returns the index of the "ADBE Group End" sentinel tdmn in
// group.Children, or len(Children) when absent (new payloads splice before it).
func groupEndIndex(group *rifx.Chunk) int {
	for i, ch := range group.Children {
		if tdmnName(ch) == "ADBE Group End" {
			return i
		}
	}
	return len(group.Children)
}

// textPropertiesChunk returns the layer's "ADBE Text Properties" LIST:tdgp from
// the Layr property tree, or nil when absent / not a parsed text layer.
func textPropertiesChunk(layer *Layer) *rifx.Chunk {
	lb := layerBack(layer)
	if lb == nil || lb.layrList == nil {
		return nil
	}
	var outer *rifx.Chunk
	for _, ch := range lb.layrList.Children {
		if ch.IsList() && ch.FormType == rifx.IDTdgp {
			outer = ch
			break
		}
	}
	if outer == nil {
		return nil
	}
	return childGroupChunk(outer, MatchNameGroupTextProperties)
}

// overwriteScalarCdat finds the first tdmn == matchName anywhere under root,
// takes the following tdbs LIST, locates its cdat, and writes value as a BE
// f64 at cdat[0:8] (the static 1D-scalar slot). Returns false when the slot
// is not present (the template elided it).
func overwriteScalarCdat(root *rifx.Chunk, matchName string, value float64) bool {
	for i, ch := range root.Children {
		if tdmnName(ch) == matchName && i+1 < len(root.Children) {
			if tdbs := root.Children[i+1]; tdbs.IsList() {
				for _, c := range tdbs.Children {
					if c.ID == rifx.IDCdat && len(c.Data) >= 8 {
						binary.BigEndian.PutUint64(c.Data[0:8], math.Float64bits(value))
						return true
					}
				}
			}
		}
		if ch.IsList() {
			if overwriteScalarCdat(ch, matchName, value) {
				return true
			}
		}
	}
	return false
}

// AddTextOpacityAnimator adds a per-character Opacity animator + Range Selector
// to a text layer and returns a stand-in group node referencing the spliced
// animator. opacity is the value applied to selected characters (0–100);
// rangeStart / rangeEnd / rangeOffset are the Range Selector bounds in percent.
// A static reveal frame: opacity 0 + start 0 + end 50 hides the first ~half of
// the characters. Animate the reveal over time with AnimateTextRangeOffset.
// (Full contract lives on the aep.AddTextOpacityAnimator facade — docgen source.)
func AddTextOpacityAnimator(layer *Layer, opacity, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	if layer == nil {
		return nil, fmt.Errorf("AddTextOpacityAnimator: layer is nil")
	}
	if layer.Type != LayerTypeText {
		return nil, fmt.Errorf("AddTextOpacityAnimator: layer %q is not a text layer", layer.Name)
	}
	tp := textPropertiesChunk(layer)
	if tp == nil {
		return nil, fmt.Errorf("AddTextOpacityAnimator: layer %q has no Text Properties group (built outside parser? round-trip via aep.Reopen first)", layer.Name)
	}

	tmpl, err := textAnimatorsTemplate()
	if err != nil {
		return nil, err
	}

	var animatorPayload *rifx.Chunk
	var undo func()

	if animators := childGroupChunk(tp, matchNameTextAnimators); animators == nil {
		// Fresh text layer: splice the whole Animators group (carrying one
		// animator) into Text Properties before its Group End sentinel.
		group := deepCloneChunk(tmpl)
		animatorPayload = childGroupChunk(group, matchNameTextAnimator)
		if animatorPayload == nil {
			return nil, fmt.Errorf("AddTextOpacityAnimator: template missing inner animator")
		}
		old := append([]*rifx.Chunk(nil), tp.Children...)
		at := groupEndIndex(tp)
		spliced := make([]*rifx.Chunk, 0, len(tp.Children)+2)
		spliced = append(spliced, tp.Children[:at]...)
		spliced = append(spliced, makeTdmn(matchNameTextAnimators), group)
		spliced = append(spliced, tp.Children[at:]...)
		tp.Children = spliced
		undo = func() { tp.Children = old }
	} else {
		// Existing Animators group: clone + append one animator before its
		// Group End sentinel.
		inner := childGroupChunk(tmpl, matchNameTextAnimator)
		if inner == nil {
			return nil, fmt.Errorf("AddTextOpacityAnimator: template missing inner animator")
		}
		animatorPayload = deepCloneChunk(inner)
		old := append([]*rifx.Chunk(nil), animators.Children...)
		at := groupEndIndex(animators)
		spliced := make([]*rifx.Chunk, 0, len(animators.Children)+2)
		spliced = append(spliced, animators.Children[:at]...)
		spliced = append(spliced, makeTdmn(matchNameTextAnimator), animatorPayload)
		spliced = append(spliced, animators.Children[at:]...)
		animators.Children = spliced
		undo = func() { animators.Children = old }
	}

	for _, sl := range []struct {
		name string
		val  float64
	}{
		{matchNameTextOpacity, opacity},
		{matchNameTextPercentStart, rangeStart},
		{matchNameTextPercentEnd, rangeEnd},
		{matchNameTextPercentOffset, rangeOffset},
	} {
		if !overwriteScalarCdat(animatorPayload, sl.name, sl.val) {
			undo()
			return nil, fmt.Errorf("AddTextOpacityAnimator: template missing %q cdat slot", sl.name)
		}
	}

	node := &AEPropertyGroup{MatchName: matchNameTextAnimator, Name: matchNameTextAnimator}
	scene.SetPropertyGroupBack(node, &propertyGroupBackrefs{chunk: animatorPayload})
	return node, nil
}

// scalarTdbs finds the first tdmn == matchName anywhere under root and returns
// the LIST:tdbs that follows it, or nil.
func scalarTdbs(root *rifx.Chunk, matchName string) *rifx.Chunk {
	for i, ch := range root.Children {
		if tdmnName(ch) == matchName && i+1 < len(root.Children) {
			if next := root.Children[i+1]; next.IsList() && next.FormType == rifx.IDTdbs {
				return next
			}
		}
		if ch.IsList() {
			if t := scalarTdbs(ch, matchName); t != nil {
				return t
			}
		}
	}
	return nil
}

// AnimateTextRangeOffset keyframes the Range Selector's "ADBE Text Percent
// Offset" on the text layer's first animator, turning a static reveal frame
// into an animated sweep (kinetic typography). With an Opacity-0 animator and
// Start=0/End=100, sweeping Offset 0→100 over time reveals the characters one by
// one (the selection window — and thus the invisibility — slides off the text).
// Builds a parsed *Property over the spliced Offset tdbs and delegates to
// AnimateScalarKeyframes (1D non-spatial, the same flip+stream the effect-param
// / shape-scalar animate paths use). Needs >= 2 keyframes.
// (Full contract lives on the aep.AnimateTextRangeOffset facade — docgen source.)
func AnimateTextRangeOffset(layer *Layer, tickRate float64, kfs []ScalarKeyframe) error {
	if layer == nil {
		return fmt.Errorf("AnimateTextRangeOffset: layer is nil")
	}
	tp := textPropertiesChunk(layer)
	if tp == nil {
		return fmt.Errorf("AnimateTextRangeOffset: layer %q has no Text Properties group", layer.Name)
	}
	animators := childGroupChunk(tp, matchNameTextAnimators)
	if animators == nil {
		return fmt.Errorf("AnimateTextRangeOffset: layer %q has no Text Animators (call AddTextOpacityAnimator first)", layer.Name)
	}
	tdbs := scalarTdbs(animators, matchNameTextPercentOffset)
	if tdbs == nil {
		return fmt.Errorf("AnimateTextRangeOffset: no Range Offset slot found")
	}
	if tickRate <= 0 {
		if comp := scene.LayerComp(layer); comp != nil && comp.TickRate > 0 {
			tickRate = comp.TickRate
		}
	}
	ctx := newParseCtx(tickRate, layer.Name, nil)
	p := parseLeafProperty(matchNameTextPercentOffset, tdbs, ctx)
	if p == nil {
		return fmt.Errorf("AnimateTextRangeOffset: failed to build property over Offset tdbs")
	}
	return AnimateScalarKeyframes(p, tickRate, kfs)
}
