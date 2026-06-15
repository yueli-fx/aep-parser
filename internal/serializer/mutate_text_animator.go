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

// Each embedded body is an AE-native "ADBE Text Animators" group carrying ONE
// animator whose single driven leaf (Opacity / Position 3D / …) plus the Range
// Selector's Start/End/Offset are all authored non-default, so every cdat slot
// is materialized for AddText*Animator to overwrite parametrically. One body per
// leaf value-type because AE elides default leaves and the cdat layout differs
// (scalar 40B f64 @ [0:8] vs spatial 3D 72B with three f64 @ [0:24]).
//
//go:embed templates/text_animators_opacity_body.bin
var textAnimatorsOpacityBody []byte

//go:embed templates/text_animators_position_body.bin
var textAnimatorsPositionBody []byte

//go:embed templates/text_animators_scale_body.bin
var textAnimatorsScaleBody []byte

//go:embed templates/text_animators_rotation_body.bin
var textAnimatorsRotationBody []byte

//go:embed templates/text_animators_color_body.bin
var textAnimatorsColorBody []byte

// Free-neighbor leaves (same on-disk layout as the five above): every one is a
// 1D scalar (vtype 6417, 40B cdat f64 @ [0:8] — identical to Opacity/Rotation)
// except Stroke Color (vtype 6418 colour, 96B cdat [A,R,G,B]×255 — identical to
// Fill Color). Each body materializes only its own leaf (AE elides the rest);
// the Rotation X/Y bodies additionally carry an inert companion Rotation (Z=0,
// default) that AE auto-adds — overwriteScalarCdat targets the X/Y match-name so
// the companion stays default. Regen via test_data/gen_text_anim_neighbor_templates.jsx.
//
//go:embed templates/text_animators_fillopacity_body.bin
var textAnimatorsFillOpacityBody []byte

//go:embed templates/text_animators_strokeopacity_body.bin
var textAnimatorsStrokeOpacityBody []byte

//go:embed templates/text_animators_strokecolor_body.bin
var textAnimatorsStrokeColorBody []byte

//go:embed templates/text_animators_strokewidth_body.bin
var textAnimatorsStrokeWidthBody []byte

//go:embed templates/text_animators_skew_body.bin
var textAnimatorsSkewBody []byte

//go:embed templates/text_animators_rotx_body.bin
var textAnimatorsRotXBody []byte

//go:embed templates/text_animators_roty_body.bin
var textAnimatorsRotYBody []byte

const (
	matchNameTextAnimators     = "ADBE Text Animators"
	matchNameTextAnimator      = "ADBE Text Animator"
	matchNameTextPercentStart  = "ADBE Text Percent Start"
	matchNameTextPercentEnd    = "ADBE Text Percent End"
	matchNameTextPercentOffset = "ADBE Text Percent Offset"
	matchNameTextOpacity       = "ADBE Text Opacity"
	matchNameTextPosition3D    = "ADBE Text Position 3D"
	matchNameTextScale3D       = "ADBE Text Scale 3D"
	matchNameTextRotation      = "ADBE Text Rotation"
	matchNameTextFillColor     = "ADBE Text Fill Color"
	matchNameTextFillOpacity   = "ADBE Text Fill Opacity"
	matchNameTextStrokeOpacity = "ADBE Text Stroke Opacity"
	matchNameTextStrokeColor   = "ADBE Text Stroke Color"
	matchNameTextStrokeWidth   = "ADBE Text Stroke Width"
	matchNameTextSkew          = "ADBE Text Skew"
	matchNameTextRotationX     = "ADBE Text Rotation X"
	matchNameTextRotationY     = "ADBE Text Rotation Y"
)

var animatorTmplCache sync.Map // first-byte ptr → *animatorTmplEntry

type animatorTmplEntry struct {
	chunk *rifx.Chunk
	err   error
}

// animatorTemplate parses an embedded "ADBE Text Animators" group body once per
// body (cached by backing-array pointer). Callers deep-clone before splicing so
// spliced chunks never alias the cache.
func animatorTemplate(body []byte) (*rifx.Chunk, error) {
	key := &body[0]
	if v, ok := animatorTmplCache.Load(key); ok {
		e := v.(*animatorTmplEntry)
		return e.chunk, e.err
	}
	c, err := rifx.ReadChunk(bytes.NewReader(body))
	if err != nil {
		err = fmt.Errorf("parse text animators template: %w", err)
	}
	e := &animatorTmplEntry{chunk: c, err: err}
	animatorTmplCache.Store(key, e)
	return e.chunk, e.err
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

// overwriteVectorCdat is overwriteScalarCdat for an N-component value: it writes
// each vals[i] as a BE f64 at cdat[8*i:8*i+8] (the spatial/vector slot, e.g.
// Position 3D's three doubles at [0:24]; trailing tangent bytes stay zero).
// Returns false when the slot is absent or too short for len(vals) doubles.
func overwriteVectorCdat(root *rifx.Chunk, matchName string, vals []float64) bool {
	for i, ch := range root.Children {
		if tdmnName(ch) == matchName && i+1 < len(root.Children) {
			if tdbs := root.Children[i+1]; tdbs.IsList() {
				for _, c := range tdbs.Children {
					if c.ID == rifx.IDCdat && len(c.Data) >= 8*len(vals) {
						for j, v := range vals {
							binary.BigEndian.PutUint64(c.Data[8*j:8*j+8], math.Float64bits(v))
						}
						return true
					}
				}
			}
		}
		if ch.IsList() {
			if overwriteVectorCdat(ch, matchName, vals) {
				return true
			}
		}
	}
	return false
}

// spliceTextAnimator inserts one animator from tmpl into the layer's text-
// property tree and returns the spliced animator payload + an undo. A fresh text
// layer (no Animators group) gets the whole "ADBE Text Animators" group spliced
// into Text Properties before its Group End sentinel; a layer that already has
// the group gets one inner "ADBE Text Animator" appended. The caller overwrites
// the payload's cdat slots, then sets the Range Selector Start/End/Offset.
func spliceTextAnimator(tp, tmpl *rifx.Chunk) (*rifx.Chunk, func(), error) {
	if animators := childGroupChunk(tp, matchNameTextAnimators); animators == nil {
		group := deepCloneChunk(tmpl)
		payload := childGroupChunk(group, matchNameTextAnimator)
		if payload == nil {
			return nil, nil, fmt.Errorf("text animator template missing inner animator")
		}
		old := append([]*rifx.Chunk(nil), tp.Children...)
		at := groupEndIndex(tp)
		spliced := make([]*rifx.Chunk, 0, len(tp.Children)+2)
		spliced = append(spliced, tp.Children[:at]...)
		spliced = append(spliced, makeTdmn(matchNameTextAnimators), group)
		spliced = append(spliced, tp.Children[at:]...)
		tp.Children = spliced
		return payload, func() { tp.Children = old }, nil
	} else {
		inner := childGroupChunk(tmpl, matchNameTextAnimator)
		if inner == nil {
			return nil, nil, fmt.Errorf("text animator template missing inner animator")
		}
		payload := deepCloneChunk(inner)
		old := append([]*rifx.Chunk(nil), animators.Children...)
		at := groupEndIndex(animators)
		spliced := make([]*rifx.Chunk, 0, len(animators.Children)+2)
		spliced = append(spliced, animators.Children[:at]...)
		spliced = append(spliced, makeTdmn(matchNameTextAnimator), payload)
		spliced = append(spliced, animators.Children[at:]...)
		animators.Children = spliced
		return payload, func() { animators.Children = old }, nil
	}
}

// setRangeSelector overwrites the spliced animator's Range Selector Start/End/
// Offset (percent) scalar cdats, rolling back via undo on a missing slot.
func setRangeSelector(payload *rifx.Chunk, undo func(), who string, start, end, offset float64) error {
	for _, sl := range []struct {
		name string
		val  float64
	}{
		{matchNameTextPercentStart, start},
		{matchNameTextPercentEnd, end},
		{matchNameTextPercentOffset, offset},
	} {
		if !overwriteScalarCdat(payload, sl.name, sl.val) {
			undo()
			return fmt.Errorf("%s: template missing %q cdat slot", who, sl.name)
		}
	}
	return nil
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
	tmpl, err := animatorTemplate(textAnimatorsOpacityBody)
	if err != nil {
		return nil, err
	}
	payload, undo, err := spliceTextAnimator(tp, tmpl)
	if err != nil {
		return nil, err
	}
	if !overwriteScalarCdat(payload, matchNameTextOpacity, opacity) {
		undo()
		return nil, fmt.Errorf("AddTextOpacityAnimator: template missing %q cdat slot", matchNameTextOpacity)
	}
	if err := setRangeSelector(payload, undo, "AddTextOpacityAnimator", rangeStart, rangeEnd, rangeOffset); err != nil {
		return nil, err
	}

	node := &AEPropertyGroup{MatchName: matchNameTextAnimator, Name: matchNameTextAnimator}
	scene.SetPropertyGroupBack(node, &propertyGroupBackrefs{chunk: payload})
	return node, nil
}

// AddTextPositionAnimator adds a per-character Position 3D animator + Range
// Selector to a text layer and returns a stand-in group node referencing the
// spliced animator. x / y / z is the position offset (pixels) applied to
// selected characters; rangeStart / rangeEnd / rangeOffset are the Range
// Selector bounds in percent. The canonical "characters slide/drop into place"
// reveal: set an offset like (0, -100, 0), Start=0/End=100, then sweep the Range
// Offset 0→100 over time with AnimateTextRangeOffset — the displacement applies
// to the not-yet-revealed characters and lands them as the window slides off.
// (Full contract lives on the aep.AddTextPositionAnimator facade — docgen source.)
func AddTextPositionAnimator(layer *Layer, x, y, z, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	if layer == nil {
		return nil, fmt.Errorf("AddTextPositionAnimator: layer is nil")
	}
	if layer.Type != LayerTypeText {
		return nil, fmt.Errorf("AddTextPositionAnimator: layer %q is not a text layer", layer.Name)
	}
	tp := textPropertiesChunk(layer)
	if tp == nil {
		return nil, fmt.Errorf("AddTextPositionAnimator: layer %q has no Text Properties group (built outside parser? round-trip via aep.Reopen first)", layer.Name)
	}
	tmpl, err := animatorTemplate(textAnimatorsPositionBody)
	if err != nil {
		return nil, err
	}
	payload, undo, err := spliceTextAnimator(tp, tmpl)
	if err != nil {
		return nil, err
	}
	if !overwriteVectorCdat(payload, matchNameTextPosition3D, []float64{x, y, z}) {
		undo()
		return nil, fmt.Errorf("AddTextPositionAnimator: template missing %q cdat slot", matchNameTextPosition3D)
	}
	if err := setRangeSelector(payload, undo, "AddTextPositionAnimator", rangeStart, rangeEnd, rangeOffset); err != nil {
		return nil, err
	}

	node := &AEPropertyGroup{MatchName: matchNameTextAnimator, Name: matchNameTextAnimator}
	scene.SetPropertyGroupBack(node, &propertyGroupBackrefs{chunk: payload})
	return node, nil
}

// AddTextScaleAnimator adds a per-character Scale 3D animator + Range Selector
// to a text layer and returns a stand-in group node referencing the spliced
// animator. sx / sy / sz is the scale percent (100 = unchanged) applied to
// selected characters; rangeStart / rangeEnd / rangeOffset are the Range
// Selector bounds in percent. The canonical "characters pop / grow into place"
// reveal: scale (0,0,100) (or, for emphasis, an oversize like 220), Start=0/
// End=100, then sweep the Range Offset 0→100 over time with
// AnimateTextRangeOffset — the scale applies to the not-yet-revealed characters
// and resolves to 100% as the window slides off.
// (Full contract lives on the aep.AddTextScaleAnimator facade — docgen source.)
func AddTextScaleAnimator(layer *Layer, sx, sy, sz, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	if layer == nil {
		return nil, fmt.Errorf("AddTextScaleAnimator: layer is nil")
	}
	if layer.Type != LayerTypeText {
		return nil, fmt.Errorf("AddTextScaleAnimator: layer %q is not a text layer", layer.Name)
	}
	tp := textPropertiesChunk(layer)
	if tp == nil {
		return nil, fmt.Errorf("AddTextScaleAnimator: layer %q has no Text Properties group (built outside parser? round-trip via aep.Reopen first)", layer.Name)
	}
	tmpl, err := animatorTemplate(textAnimatorsScaleBody)
	if err != nil {
		return nil, err
	}
	payload, undo, err := spliceTextAnimator(tp, tmpl)
	if err != nil {
		return nil, err
	}
	if !overwriteVectorCdat(payload, matchNameTextScale3D, []float64{sx, sy, sz}) {
		undo()
		return nil, fmt.Errorf("AddTextScaleAnimator: template missing %q cdat slot", matchNameTextScale3D)
	}
	if err := setRangeSelector(payload, undo, "AddTextScaleAnimator", rangeStart, rangeEnd, rangeOffset); err != nil {
		return nil, err
	}

	node := &AEPropertyGroup{MatchName: matchNameTextAnimator, Name: matchNameTextAnimator}
	scene.SetPropertyGroupBack(node, &propertyGroupBackrefs{chunk: payload})
	return node, nil
}

// AddTextRotationAnimator adds a per-character Rotation animator + Range
// Selector to a text layer and returns a stand-in group node referencing the
// spliced animator. rotation is the angle in degrees applied to selected
// characters (each rotates about its own anchor); rangeStart / rangeEnd /
// rangeOffset are the Range Selector bounds in percent. The canonical "letters
// spin into place" reveal: rotation 90, Start=0/End=100, then sweep the Range
// Offset 0→100 over time with AnimateTextRangeOffset — the rotation resolves to
// 0° as the selection window slides off the characters.
// (Full contract lives on the aep.AddTextRotationAnimator facade — docgen source.)
func AddTextRotationAnimator(layer *Layer, rotation, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	if layer == nil {
		return nil, fmt.Errorf("AddTextRotationAnimator: layer is nil")
	}
	if layer.Type != LayerTypeText {
		return nil, fmt.Errorf("AddTextRotationAnimator: layer %q is not a text layer", layer.Name)
	}
	tp := textPropertiesChunk(layer)
	if tp == nil {
		return nil, fmt.Errorf("AddTextRotationAnimator: layer %q has no Text Properties group (built outside parser? round-trip via aep.Reopen first)", layer.Name)
	}
	tmpl, err := animatorTemplate(textAnimatorsRotationBody)
	if err != nil {
		return nil, err
	}
	payload, undo, err := spliceTextAnimator(tp, tmpl)
	if err != nil {
		return nil, err
	}
	if !overwriteScalarCdat(payload, matchNameTextRotation, rotation) {
		undo()
		return nil, fmt.Errorf("AddTextRotationAnimator: template missing %q cdat slot", matchNameTextRotation)
	}
	if err := setRangeSelector(payload, undo, "AddTextRotationAnimator", rangeStart, rangeEnd, rangeOffset); err != nil {
		return nil, err
	}

	node := &AEPropertyGroup{MatchName: matchNameTextAnimator, Name: matchNameTextAnimator}
	scene.SetPropertyGroupBack(node, &propertyGroupBackrefs{chunk: payload})
	return node, nil
}

// AddTextColorAnimator adds a per-character Fill Color animator + Range Selector
// to a text layer and returns a stand-in group node referencing the spliced
// animator. r / g / b / a is the colour applied to selected characters (each
// channel 0..1); rangeStart / rangeEnd / rangeOffset are the Range Selector
// bounds in percent. The canonical "characters tint in" reveal: set a target
// colour, Start=0/End=100, then sweep the Range Offset 0→100 over time with
// AnimateTextRangeOffset — the colour applies to the selected characters and
// resolves to the base text colour as the window slides off.
// AE stores the colour as [A,R,G,B] × 255 f64 BE (same on-disk encoding as shape
// Fill/Stroke), so the four channels are written at cdat[0:32].
// (Full contract lives on the aep.AddTextColorAnimator facade — docgen source.)
func AddTextColorAnimator(layer *Layer, r, g, b, a, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	if layer == nil {
		return nil, fmt.Errorf("AddTextColorAnimator: layer is nil")
	}
	if layer.Type != LayerTypeText {
		return nil, fmt.Errorf("AddTextColorAnimator: layer %q is not a text layer", layer.Name)
	}
	tp := textPropertiesChunk(layer)
	if tp == nil {
		return nil, fmt.Errorf("AddTextColorAnimator: layer %q has no Text Properties group (built outside parser? round-trip via aep.Reopen first)", layer.Name)
	}
	tmpl, err := animatorTemplate(textAnimatorsColorBody)
	if err != nil {
		return nil, err
	}
	payload, undo, err := spliceTextAnimator(tp, tmpl)
	if err != nil {
		return nil, err
	}
	if !overwriteVectorCdat(payload, matchNameTextFillColor, []float64{a * 255, r * 255, g * 255, b * 255}) {
		undo()
		return nil, fmt.Errorf("AddTextColorAnimator: template missing %q cdat slot", matchNameTextFillColor)
	}
	if err := setRangeSelector(payload, undo, "AddTextColorAnimator", rangeStart, rangeEnd, rangeOffset); err != nil {
		return nil, err
	}

	node := &AEPropertyGroup{MatchName: matchNameTextAnimator, Name: matchNameTextAnimator}
	scene.SetPropertyGroupBack(node, &propertyGroupBackrefs{chunk: payload})
	return node, nil
}

// addTextScalarLeafAnimator is the shared body of every 1D-scalar
// AddText*Animator facade (Fill/Stroke Opacity, Stroke Width, Skew, Rotation
// X·Y): it splices one animator from body into the layer's text-property tree,
// overwrites the leaf's scalar cdat (f64 @ [0:8]) with value, and sets the Range
// Selector. matchName is the driven leaf; who is the caller name for errors.
func addTextScalarLeafAnimator(layer *Layer, body []byte, matchName, who string, value, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	if layer == nil {
		return nil, fmt.Errorf("%s: layer is nil", who)
	}
	if layer.Type != LayerTypeText {
		return nil, fmt.Errorf("%s: layer %q is not a text layer", who, layer.Name)
	}
	tp := textPropertiesChunk(layer)
	if tp == nil {
		return nil, fmt.Errorf("%s: layer %q has no Text Properties group (built outside parser? round-trip via aep.Reopen first)", who, layer.Name)
	}
	tmpl, err := animatorTemplate(body)
	if err != nil {
		return nil, err
	}
	payload, undo, err := spliceTextAnimator(tp, tmpl)
	if err != nil {
		return nil, err
	}
	if !overwriteScalarCdat(payload, matchName, value) {
		undo()
		return nil, fmt.Errorf("%s: template missing %q cdat slot", who, matchName)
	}
	if err := setRangeSelector(payload, undo, who, rangeStart, rangeEnd, rangeOffset); err != nil {
		return nil, err
	}
	node := &AEPropertyGroup{MatchName: matchNameTextAnimator, Name: matchNameTextAnimator}
	scene.SetPropertyGroupBack(node, &propertyGroupBackrefs{chunk: payload})
	return node, nil
}

// addTextColorLeafAnimator is addTextScalarLeafAnimator for a 4-channel colour
// leaf (Stroke Color): r/g/b/a are 0..1 and written as the on-disk [A,R,G,B]×255
// cdat (the same encoding shape Fill/Stroke and Fill Color use).
func addTextColorLeafAnimator(layer *Layer, body []byte, matchName, who string, r, g, b, a, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	if layer == nil {
		return nil, fmt.Errorf("%s: layer is nil", who)
	}
	if layer.Type != LayerTypeText {
		return nil, fmt.Errorf("%s: layer %q is not a text layer", who, layer.Name)
	}
	tp := textPropertiesChunk(layer)
	if tp == nil {
		return nil, fmt.Errorf("%s: layer %q has no Text Properties group (built outside parser? round-trip via aep.Reopen first)", who, layer.Name)
	}
	tmpl, err := animatorTemplate(body)
	if err != nil {
		return nil, err
	}
	payload, undo, err := spliceTextAnimator(tp, tmpl)
	if err != nil {
		return nil, err
	}
	if !overwriteVectorCdat(payload, matchName, []float64{a * 255, r * 255, g * 255, b * 255}) {
		undo()
		return nil, fmt.Errorf("%s: template missing %q cdat slot", who, matchName)
	}
	if err := setRangeSelector(payload, undo, who, rangeStart, rangeEnd, rangeOffset); err != nil {
		return nil, err
	}
	node := &AEPropertyGroup{MatchName: matchNameTextAnimator, Name: matchNameTextAnimator}
	scene.SetPropertyGroupBack(node, &propertyGroupBackrefs{chunk: payload})
	return node, nil
}

// AddTextFillOpacityAnimator adds a per-character Fill Opacity animator + Range
// Selector to a text layer. opacity (0–100) is applied to selected characters;
// rangeStart/rangeEnd/rangeOffset are the Range Selector bounds in percent.
// (Full contract lives on the aep.AddTextFillOpacityAnimator facade.)
func AddTextFillOpacityAnimator(layer *Layer, opacity, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	return addTextScalarLeafAnimator(layer, textAnimatorsFillOpacityBody, matchNameTextFillOpacity, "AddTextFillOpacityAnimator", opacity, rangeStart, rangeEnd, rangeOffset)
}

// AddTextStrokeOpacityAnimator adds a per-character Stroke Opacity animator +
// Range Selector to a text layer. opacity (0–100) is applied to selected
// characters' stroke; the text must carry a stroke for this to be visible.
// (Full contract lives on the aep.AddTextStrokeOpacityAnimator facade.)
func AddTextStrokeOpacityAnimator(layer *Layer, opacity, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	return addTextScalarLeafAnimator(layer, textAnimatorsStrokeOpacityBody, matchNameTextStrokeOpacity, "AddTextStrokeOpacityAnimator", opacity, rangeStart, rangeEnd, rangeOffset)
}

// AddTextStrokeWidthAnimator adds a per-character Stroke Width animator + Range
// Selector to a text layer. width (pixels) is applied to selected characters'
// stroke; the text must carry a stroke for this to be visible.
// (Full contract lives on the aep.AddTextStrokeWidthAnimator facade.)
func AddTextStrokeWidthAnimator(layer *Layer, width, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	return addTextScalarLeafAnimator(layer, textAnimatorsStrokeWidthBody, matchNameTextStrokeWidth, "AddTextStrokeWidthAnimator", width, rangeStart, rangeEnd, rangeOffset)
}

// AddTextSkewAnimator adds a per-character Skew animator + Range Selector to a
// text layer. skew is the shear angle (degrees) applied to selected characters;
// rangeStart/rangeEnd/rangeOffset are the Range Selector bounds in percent.
// (Full contract lives on the aep.AddTextSkewAnimator facade.)
func AddTextSkewAnimator(layer *Layer, skew, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	return addTextScalarLeafAnimator(layer, textAnimatorsSkewBody, matchNameTextSkew, "AddTextSkewAnimator", skew, rangeStart, rangeEnd, rangeOffset)
}

// AddTextRotationXAnimator adds a per-character Rotation X (3D, about the
// horizontal axis) animator + Range Selector to a text layer. rotation is the
// angle in degrees applied to selected characters; rangeStart/rangeEnd/
// rangeOffset are the Range Selector bounds in percent.
// (Full contract lives on the aep.AddTextRotationXAnimator facade.)
func AddTextRotationXAnimator(layer *Layer, rotation, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	return addTextScalarLeafAnimator(layer, textAnimatorsRotXBody, matchNameTextRotationX, "AddTextRotationXAnimator", rotation, rangeStart, rangeEnd, rangeOffset)
}

// AddTextRotationYAnimator adds a per-character Rotation Y (3D, about the
// vertical axis) animator + Range Selector to a text layer. rotation is the
// angle in degrees applied to selected characters; rangeStart/rangeEnd/
// rangeOffset are the Range Selector bounds in percent.
// (Full contract lives on the aep.AddTextRotationYAnimator facade.)
func AddTextRotationYAnimator(layer *Layer, rotation, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	return addTextScalarLeafAnimator(layer, textAnimatorsRotYBody, matchNameTextRotationY, "AddTextRotationYAnimator", rotation, rangeStart, rangeEnd, rangeOffset)
}

// AddTextStrokeColorAnimator adds a per-character Stroke Color animator + Range
// Selector to a text layer. r/g/b/a (0..1) is the colour applied to selected
// characters' stroke; the text must carry a stroke for this to be visible.
// (Full contract lives on the aep.AddTextStrokeColorAnimator facade.)
func AddTextStrokeColorAnimator(layer *Layer, r, g, b, a, rangeStart, rangeEnd, rangeOffset float64) (*AEPropertyGroup, error) {
	return addTextColorLeafAnimator(layer, textAnimatorsStrokeColorBody, matchNameTextStrokeColor, "AddTextStrokeColorAnimator", r, g, b, a, rangeStart, rangeEnd, rangeOffset)
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

// animateTextScalarLeaf keyframes a 1D-scalar animator-properties leaf (Opacity
// / Rotation / Skew / …) on the layer's first animator, converting its static
// value into a keyframed stream. Same parse-the-clone + AnimateScalarKeyframes
// vein as AnimateTextRangeOffset, but it targets a DRIVEN leaf rather than the
// Range Selector Offset — so all selected characters share one value curve over
// time (a synchronized pulse / spin / fade, which a Range-Offset sweep cannot
// express). The animator must already carry the leaf (call the matching
// AddText*Animator first); the leaf must still be static. Needs >= 2 keyframes.
func animateTextScalarLeaf(layer *Layer, matchName, who string, tickRate float64, kfs []ScalarKeyframe) error {
	if layer == nil {
		return fmt.Errorf("%s: layer is nil", who)
	}
	if layer.Type != LayerTypeText {
		return fmt.Errorf("%s: layer %q is not a text layer", who, layer.Name)
	}
	tp := textPropertiesChunk(layer)
	if tp == nil {
		return fmt.Errorf("%s: layer %q has no Text Properties group", who, layer.Name)
	}
	animators := childGroupChunk(tp, matchNameTextAnimators)
	if animators == nil {
		return fmt.Errorf("%s: layer %q has no Text Animators (add the matching animator first)", who, layer.Name)
	}
	tdbs := scalarTdbs(animators, matchName)
	if tdbs == nil {
		return fmt.Errorf("%s: no %q leaf found (add the matching animator first)", who, matchName)
	}
	if tickRate <= 0 {
		if comp := scene.LayerComp(layer); comp != nil && comp.TickRate > 0 {
			tickRate = comp.TickRate
		}
	}
	ctx := newParseCtx(tickRate, layer.Name, nil)
	p := parseLeafProperty(matchName, tdbs, ctx)
	if p == nil {
		return fmt.Errorf("%s: failed to build property over %q tdbs", who, matchName)
	}
	return AnimateScalarKeyframes(p, tickRate, kfs)
}

// AnimateTextOpacity keyframes the per-character Opacity leaf of the layer's
// first text animator (added via AddTextOpacityAnimator), fading the selected
// characters as one synchronized group over time. Unlike AnimateTextRangeOffset
// (which sweeps the selection window), this animates the driven value itself, so
// every selected character shares the opacity curve. Needs >= 2 keyframes;
// tickRate <= 0 uses the comp's.
// (Full contract lives on the aep.AnimateTextOpacity facade — docgen source.)
func AnimateTextOpacity(layer *Layer, tickRate float64, kfs []ScalarKeyframe) error {
	return animateTextScalarLeaf(layer, matchNameTextOpacity, "AnimateTextOpacity", tickRate, kfs)
}

// AnimateTextRotation keyframes the per-character Rotation leaf of the layer's
// first text animator (added via AddTextRotationAnimator), spinning the selected
// characters as one synchronized group over time (e.g. a continuous 0→360 spin,
// which a Range-Offset sweep cannot express). Unlike AnimateTextRangeOffset, this
// animates the driven angle itself. Needs >= 2 keyframes; tickRate <= 0 uses the
// comp's.
// (Full contract lives on the aep.AnimateTextRotation facade — docgen source.)
func AnimateTextRotation(layer *Layer, tickRate float64, kfs []ScalarKeyframe) error {
	return animateTextScalarLeaf(layer, matchNameTextRotation, "AnimateTextRotation", tickRate, kfs)
}

// animateTextVectorLeaf is animateTextScalarLeaf for a 2/3/4-component leaf
// (Position 3D / Fill Color): it locates the leaf, wraps it as a parsed Property,
// and delegates to AnimateVectorKeyframes (the SPATIAL keyframe block effect
// color/point params use — verified byte-matching the AE-saved text leaf for the
// spatial Position 3D and 4-channel Fill Color leaves; the non-spatial Scale 3D
// leaf uses a different block and is NOT routed here). Values are in the leaf's
// on-disk units (Position = pixels, Fill Color = [A,R,G,B]×255).
func animateTextVectorLeaf(layer *Layer, matchName, who string, nonSpatial bool, tickRate float64, kfs []VectorKeyframe) error {
	if layer == nil {
		return fmt.Errorf("%s: layer is nil", who)
	}
	if layer.Type != LayerTypeText {
		return fmt.Errorf("%s: layer %q is not a text layer", who, layer.Name)
	}
	tp := textPropertiesChunk(layer)
	if tp == nil {
		return fmt.Errorf("%s: layer %q has no Text Properties group", who, layer.Name)
	}
	animators := childGroupChunk(tp, matchNameTextAnimators)
	if animators == nil {
		return fmt.Errorf("%s: layer %q has no Text Animators (add the matching animator first)", who, layer.Name)
	}
	tdbs := scalarTdbs(animators, matchName)
	if tdbs == nil {
		return fmt.Errorf("%s: no %q leaf found (add the matching animator first)", who, matchName)
	}
	if tickRate <= 0 {
		if comp := scene.LayerComp(layer); comp != nil && comp.TickRate > 0 {
			tickRate = comp.TickRate
		}
	}
	ctx := newParseCtx(tickRate, layer.Name, nil)
	p := parseLeafProperty(matchName, tdbs, ctx)
	if p == nil {
		return fmt.Errorf("%s: failed to build property over %q tdbs", who, matchName)
	}
	if nonSpatial {
		return AnimateVectorKeyframesNonSpatial(p, tickRate, kfs)
	}
	return AnimateVectorKeyframes(p, tickRate, kfs)
}

// AnimateTextPosition keyframes the per-character Position 3D leaf of the layer's
// first text animator (added via AddTextPositionAnimator), translating the
// selected characters as one synchronized group over time. Each keyframe Value is
// the [x, y, z] offset in pixels. Needs >= 2 keyframes; tickRate <= 0 uses the
// comp's.
// (Full contract lives on the aep.AnimateTextPosition facade — docgen source.)
func AnimateTextPosition(layer *Layer, tickRate float64, kfs []VectorKeyframe) error {
	return animateTextVectorLeaf(layer, matchNameTextPosition3D, "AnimateTextPosition", false, tickRate, kfs)
}

// AnimateTextScale keyframes the per-character Scale 3D leaf of the layer's first
// text animator (added via AddTextScaleAnimator), scaling the selected characters
// as one synchronized group over time (e.g. a pulse). Each keyframe Value is the
// [sx, sy, sz] scale percent (100 = unchanged). Scale 3D is a NON-SPATIAL
// 3-component leaf, so it routes through AnimateVectorKeyframesNonSpatial (value
// @0x08 block) rather than the spatial block Position/Color use. Needs >= 2
// keyframes; tickRate <= 0 uses the comp's.
// (Full contract lives on the aep.AnimateTextScale facade — docgen source.)
func AnimateTextScale(layer *Layer, tickRate float64, kfs []VectorKeyframe) error {
	return animateTextVectorLeaf(layer, matchNameTextScale3D, "AnimateTextScale", true, tickRate, kfs)
}

// AnimateTextColor keyframes the per-character Fill Color leaf of the layer's
// first text animator (added via AddTextColorAnimator), tinting the selected
// characters through a colour curve over time (e.g. red→blue cycling). Each
// keyframe Value is an [r, g, b, a] colour with channels 0..1; it is converted to
// the on-disk [A,R,G,B]×255 encoding before keyframing. Needs >= 2 keyframes;
// tickRate <= 0 uses the comp's.
// (Full contract lives on the aep.AnimateTextColor facade — docgen source.)
func AnimateTextColor(layer *Layer, tickRate float64, kfs []VectorKeyframe) error {
	conv := make([]VectorKeyframe, len(kfs))
	for i, kf := range kfs {
		if len(kf.Value) != 4 {
			return fmt.Errorf("AnimateTextColor: keyframe %d Value must be [r,g,b,a] (4 channels), got %d", i, len(kf.Value))
		}
		r, g, b, a := kf.Value[0], kf.Value[1], kf.Value[2], kf.Value[3]
		conv[i] = VectorKeyframe{
			Time:    kf.Time,
			Value:   []float64{a * 255, r * 255, g * 255, b * 255},
			InEase:  kf.InEase,
			OutEase: kf.OutEase,
		}
	}
	return animateTextVectorLeaf(layer, matchNameTextFillColor, "AnimateTextColor", false, tickRate, conv)
}
