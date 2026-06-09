// Public-API entry for adding an effect to a layer's Effect Parade.
//
// AE stores each effect as a (tdmn match-name, LIST:sspc payload) pair inside
// the layer's "ADBE Effect Parade" tdgp group, terminated by a lone
// "ADBE Group End" tdmn sentinel. Adding an effect = splice a fresh
// (tdmn, sspc) pair in just before that sentinel. The sspc payload (parameter
// tree, pard metadata, built-in-params group) is supplied verbatim from an
// embedded AE-native template — the same embed-AE-bytes strategy the V2.2 shape
// bodies use, and the same (tdmn, payload) splice DuplicatePropertyGroup is
// ship-gate-green with.
//
// Atomic: snapshot parade chunk + scene children + flat Effects slice, commit,
// re-parse the spliced pair to obtain a back-ref-correct *Effect, roll back on
// any parser warning. LIST sizes are recomputed bottom-up by rifx.Chunk.Write,
// so the byte-length growth needs no manual fixup (same as Footage.SetPath /
// gradient writes).
//
// Phase 1 requires the layer to already carry an Effect Parade group (every
// AE-parsed AV layer does). From-scratch shape layers built by NewShapeLayer
// have no parade yet (auto-create deferred to Phase 2).
package serializer

import (
	"bytes"
	_ "embed"
	"fmt"
	"sort"
	"sync"

	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

// effect_gaussian_blur_body.bin is a LIST(tdgp) wrapper holding the
// (tdmn "ADBE Gaussian Blur 2", LIST:sspc) pair extracted verbatim from an
// AE-2020-saved layer (test_data/re_property_struct_baseline.aep). Gaussian
// Blur is a built-in effect in every AE since well before the 2020 read floor
// and its serialized form is version-portable (AE 2025 accepts the AE-2020
// bytes — confirmed by ship-gate, mirroring the gradient version-portability
// finding).
//
//go:embed templates/effect_gaussian_blur_body.bin
var effectGaussianBlurBytes []byte

// effectTemplates maps an effect match-name to its embedded (tdmn+sspc) wrapper
// bytes. Each entry is an AE-native effect instance with default-ish parameter
// values; callers tune parameters afterward via the parsed Effect.Parameters
// (Property.SetStaticValue works on effect params).
var effectTemplates = map[string][]byte{
	"ADBE Gaussian Blur 2": effectGaussianBlurBytes,
}

type cachedEffectTemplate struct {
	once  sync.Once
	chunk *rifx.Chunk
	err   error
}

var effectTemplateCache = map[string]*cachedEffectTemplate{}
var effectTemplateCacheMu sync.Mutex

// SupportedEffects returns the sorted set of effect match-names AddEffect can
// currently add from an embedded template.
func SupportedEffects() []string {
	names := make([]string, 0, len(effectTemplates))
	for k := range effectTemplates {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// cloneEffectTemplate parses the embedded template for matchName once (cached)
// and returns a deep clone of its (tdmn, sspc) pair so the spliced chunks never
// alias the cache (concurrent-mutate safe).
func cloneEffectTemplate(matchName string) (tdmn, sspc *rifx.Chunk, err error) {
	raw, ok := effectTemplates[matchName]
	if !ok {
		return nil, nil, fmt.Errorf("AddEffect: unsupported effect %q (supported: %v)", matchName, SupportedEffects())
	}
	effectTemplateCacheMu.Lock()
	ct := effectTemplateCache[matchName]
	if ct == nil {
		ct = &cachedEffectTemplate{}
		effectTemplateCache[matchName] = ct
	}
	effectTemplateCacheMu.Unlock()

	ct.once.Do(func() {
		wrapper, e := rifx.ReadChunk(bytes.NewReader(raw))
		if e != nil {
			ct.err = fmt.Errorf("parse effect template %q: %w", matchName, e)
			return
		}
		if len(wrapper.Children) != 2 {
			ct.err = fmt.Errorf("effect template %q: want 2 wrapper children (tdmn, sspc), got %d", matchName, len(wrapper.Children))
			return
		}
		ct.chunk = wrapper
	})
	if ct.err != nil {
		return nil, nil, ct.err
	}
	return deepCloneChunk(ct.chunk.Children[0]), deepCloneChunk(ct.chunk.Children[1]), nil
}

// AddEffect appends an effect to the layer's Effect Parade and returns the
// parsed *Effect (so the caller can tune Effect.Parameters immediately).
// (Full contract + RE notes live on the aep.AddEffect facade — docgen source.)
func AddEffect(layer *Layer, effectMatchName string) (*Effect, error) {
	if layer == nil {
		return nil, fmt.Errorf("AddEffect: layer is nil")
	}
	parade := layer.EffectsParade()
	if parade == nil {
		return nil, fmt.Errorf("AddEffect: layer %q has no Effect Parade group (from-scratch layers have none yet; parade auto-create not implemented)", layer.Name)
	}
	pgb := propertyGroupBack(parade)
	if pgb == nil || pgb.chunk == nil {
		return nil, fmt.Errorf("AddEffect: Effect Parade for layer %q has no chunk back-ref", layer.Name)
	}

	tdmnCh, sspcCh, err := cloneEffectTemplate(effectMatchName)
	if err != nil {
		return nil, err
	}

	children := pgb.chunk.Children
	insertIdx := len(children)
	for i, ch := range children {
		if ch.ID == rifx.IDTdmn && string(bytes.TrimRight(ch.Data, "\x00")) == "ADBE Group End" {
			insertIdx = i
			break
		}
	}

	// Snapshot for atomic rollback.
	oldChunkChildren := append([]*rifx.Chunk(nil), children...)
	oldSceneChildren := append([]PropertyBase(nil), parade.Children...)
	oldEffects := append([]*Effect(nil), layer.Effects...)
	oldWarningsLen := warningsLen(layer)

	rollback := func() {
		pgb.chunk.Children = oldChunkChildren
		parade.Children = oldSceneChildren
		layer.Effects = oldEffects
		rollbackWarnings(layer, oldWarningsLen)
	}

	// Chunk: splice (tdmn, sspc) in just before the Group End sentinel.
	spliced := make([]*rifx.Chunk, 0, len(children)+2)
	spliced = append(spliced, children[:insertIdx]...)
	spliced = append(spliced, tdmnCh, sspcCh)
	spliced = append(spliced, children[insertIdx:]...)
	pgb.chunk.Children = spliced

	// Scene: append a stand-in group node (the parade's scene children are the
	// effect groups in order; the Group End sentinel is chunk-only).
	cloneNode := &AEPropertyGroup{MatchName: effectMatchName, Name: effectMatchName}
	scene.SetPropertyGroupParent(cloneNode, parade)
	scene.SetPropertyGroupBack(cloneNode, &propertyGroupBackrefs{chunk: sspcCh})
	parade.Children = append(parade.Children, cloneNode)

	// Flat mirror: re-parse the spliced pair so the typed Effect's back-refs
	// point at the spliced chunks (never aliased to the template cache).
	var newEffect *Effect
	if comp := scene.LayerComp(layer); comp != nil && scene.CompositionProj(comp) != nil {
		ctx := newParseCtxFPS(comp.TickRate, comp.FrameRate, comp.Name, &scene.CompositionProj(comp).Warnings)
		tmpParade := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp, Children: []*rifx.Chunk{tdmnCh, sspcCh}}
		var tmp []*Effect
		collectEffects(tmpParade, &tmp, ctx)
		if len(tmp) != 1 {
			rollback()
			return nil, fmt.Errorf("AddEffect: spliced effect re-parse produced %d effects (want 1)", len(tmp))
		}
		newEffect = tmp[0]
		layer.Effects = append(layer.Effects, newEffect)
	}

	if newWarn := newWarningsSince(layer, oldWarningsLen); len(newWarn) > 0 {
		rollback()
		return nil, fmt.Errorf("AddEffect: produced %d parser warning(s), rolled back: %v", len(newWarn), newWarn)
	}

	return newEffect, nil
}
