// Effect-parameter materialization (synthesis-lite) + typed set entry.
//
// AE persists an effect parameter only while its current value differs from
// the default — a default-valued param has NO tdbs stream in the effect's
// value tdgp (only its (tdmn, pard) definition in parT survives), so
// Property.SetStaticValue has nothing to write into. SetEffectParam mirrors
// AE's own semantics: when the target param is default-elided it splices a
// materialized (tdmn, LIST:tdbs) pair cloned from an embedded AE-native
// template into the value tdgp (definition order), then writes the caller's
// value — which is non-default by intent, exactly the state AE itself
// persists. The always-present "<effect>-0000" stream is the only one
// carrying a tdpi host binding, so spliced extras need no retarget.
// RE: flightdeck/incidents/effect-param-elision-synthesis-lite.md.
package serializer

import (
	"bytes"
	"embed"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

// effectParamTemplateFS holds embedded materialized-parameter templates: each
// a LIST(tdgp) wrapper around one (tdmn, LIST:tdbs) pair, extracted from an
// AE-2020-saved instance with the param touched
// (test_data/re_effect_param_elision.aep via tmp_debug/extract_effect_params).
//
//go:embed templates/effectparam_adbe_gaussian_blur_2_0001.bin templates/effectparam_adbe_gaussian_blur_2_0002.bin templates/effectparam_adbe_gaussian_blur_2_0003.bin
var effectParamTemplateFS embed.FS

// effectParamTemplateFiles maps a full parameter match-name to its embedded
// template path. The template's cdat carries the extraction fixture's value;
// SetEffectParam always overwrites it with the caller's value before
// returning, so the stale bytes never surface.
var effectParamTemplateFiles = map[string]string{
	"ADBE Gaussian Blur 2-0001": "templates/effectparam_adbe_gaussian_blur_2_0001.bin", // Blurriness (scalar)
	"ADBE Gaussian Blur 2-0002": "templates/effectparam_adbe_gaussian_blur_2_0002.bin", // Blur Dimensions (enum)
	"ADBE Gaussian Blur 2-0003": "templates/effectparam_adbe_gaussian_blur_2_0003.bin", // Repeat Edge Pixels (bool)
}

var effectParamTemplateCache = map[string]*cachedEffectTemplate{}
var effectParamTemplateCacheMu sync.Mutex

// SupportedEffectParams returns the sorted parameter match-names
// SetEffectParam can materialize from an embedded template when the target
// parameter is default-elided. Parameters already present on an effect are
// settable regardless of this list.
func SupportedEffectParams() []string {
	names := make([]string, 0, len(effectParamTemplateFiles))
	for k := range effectParamTemplateFiles {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func cloneEffectParamTemplate(paramMatchName string) (tdmn, tdbs *rifx.Chunk, err error) {
	path, ok := effectParamTemplateFiles[paramMatchName]
	if !ok {
		return nil, nil, fmt.Errorf("SetEffectParam: parameter %q is default-elided on this effect instance and no embedded template exists to materialize it (have: %v)", paramMatchName, SupportedEffectParams())
	}
	effectParamTemplateCacheMu.Lock()
	ct := effectParamTemplateCache[paramMatchName]
	if ct == nil {
		ct = &cachedEffectTemplate{}
		effectParamTemplateCache[paramMatchName] = ct
	}
	effectParamTemplateCacheMu.Unlock()

	ct.once.Do(func() {
		raw, e := effectParamTemplateFS.ReadFile(path)
		if e != nil {
			ct.err = fmt.Errorf("read effect param template %q (%s): %w", paramMatchName, path, e)
			return
		}
		wrapper, e := rifx.ReadChunk(bytes.NewReader(raw))
		if e != nil {
			ct.err = fmt.Errorf("parse effect param template %q: %w", paramMatchName, e)
			return
		}
		if len(wrapper.Children) != 2 {
			ct.err = fmt.Errorf("effect param template %q: want 2 wrapper children (tdmn, tdbs), got %d", paramMatchName, len(wrapper.Children))
			return
		}
		ct.chunk = wrapper
	})
	if ct.err != nil {
		return nil, nil, ct.err
	}
	return deepCloneChunk(ct.chunk.Children[0]), deepCloneChunk(ct.chunk.Children[1]), nil
}

// SetEffectParam sets an effect parameter's static value, materializing the
// parameter's value stream first when it is default-elided.
// (Full contract lives on the aep.SetEffectParam facade — docgen source.)
func SetEffectParam(layer *Layer, fx *Effect, paramMatchName string, value any) (*Property, error) {
	if layer == nil || fx == nil {
		return nil, fmt.Errorf("SetEffectParam: layer/effect is nil")
	}
	for _, p := range fx.Parameters {
		if p.MatchName == paramMatchName {
			if err := p.SetStaticValue(value); err != nil {
				return nil, err
			}
			return p, nil
		}
	}
	if !strings.HasPrefix(paramMatchName, fx.MatchName+"-") {
		return nil, fmt.Errorf("SetEffectParam: parameter %q does not belong to effect %q", paramMatchName, fx.MatchName)
	}

	idx := -1
	for i, e := range layer.Effects {
		if e == fx {
			idx = i
			break
		}
	}
	if idx < 0 {
		return nil, fmt.Errorf("SetEffectParam: effect %q is not on layer %q", fx.MatchName, layer.Name)
	}
	parade := layer.EffectsParade()
	if parade == nil || idx >= parade.NumProperties() {
		return nil, fmt.Errorf("SetEffectParam: layer %q has no parsed Effect Parade entry for effect %d; round-trip the project through aep.Reopen first", layer.Name, idx)
	}
	g, ok := parade.ChildByIndex(idx).(*AEPropertyGroup)
	if !ok {
		return nil, fmt.Errorf("SetEffectParam: parade child %d is not a property group", idx)
	}
	pgb := propertyGroupBack(g)
	if pgb == nil || pgb.chunk == nil {
		return nil, fmt.Errorf("SetEffectParam: effect %q has no chunk back-ref", fx.MatchName)
	}
	var valueGroup *rifx.Chunk
	for _, ch := range pgb.chunk.Children {
		if ch.IsList() && ch.FormType == rifx.IDTdgp {
			valueGroup = ch
			break
		}
	}
	if valueGroup == nil {
		return nil, fmt.Errorf("SetEffectParam: effect %q sspc has no value tdgp", fx.MatchName)
	}

	tdmnCh, tdbsCh, err := cloneEffectParamTemplate(paramMatchName)
	if err != nil {
		return nil, err
	}

	// Insertion point: definition order = ascending param match-name among the
	// effect's own pairs; fall back to just before the Built In Params group /
	// Group End sentinel.
	kids := valueGroup.Children
	insertIdx := len(kids)
	ordinal := 0
	for i := 0; i < len(kids); i++ {
		if kids[i].ID != rifx.IDTdmn {
			continue
		}
		name := string(bytes.TrimRight(kids[i].Data, "\x00"))
		if strings.HasPrefix(name, fx.MatchName+"-") {
			if name > paramMatchName {
				insertIdx = i
				break
			}
			ordinal++
			continue
		}
		if name == "ADBE Effect Built In Params" || name == "ADBE Group End" {
			insertIdx = i
			break
		}
	}

	oldChunkChildren := append([]*rifx.Chunk(nil), kids...)
	oldParams := append([]*Property(nil), fx.Parameters...)
	oldWarningsLen := warningsLen(layer)
	rollback := func() {
		valueGroup.Children = oldChunkChildren
		fx.Parameters = oldParams
		rollbackWarnings(layer, oldWarningsLen)
	}

	spliced := make([]*rifx.Chunk, 0, len(kids)+2)
	spliced = append(spliced, kids[:insertIdx]...)
	spliced = append(spliced, tdmnCh, tdbsCh)
	spliced = append(spliced, kids[insertIdx:]...)
	valueGroup.Children = spliced

	comp := scene.LayerComp(layer)
	if comp == nil || scene.CompositionProj(comp) == nil {
		rollback()
		return nil, fmt.Errorf("SetEffectParam: layer %q has no composition/project back-ref", layer.Name)
	}
	ctx := newParseCtxFPS(comp.TickRate, comp.FrameRate, comp.Name, &scene.CompositionProj(comp).Warnings)
	prop := parseLeafProperty(paramMatchName, tdbsCh, ctx)
	if prop == nil {
		rollback()
		return nil, fmt.Errorf("SetEffectParam: spliced param %q re-parse produced no property", paramMatchName)
	}
	if err := prop.SetStaticValue(value); err != nil {
		rollback()
		return nil, err
	}

	// Mirror into the flat Parameters slice at the same relative position
	// (ordinal = number of effect-own params preceding the splice point).
	pos := len(fx.Parameters)
	seen := 0
	for i, p := range fx.Parameters {
		if strings.HasPrefix(p.MatchName, fx.MatchName+"-") {
			if seen == ordinal {
				pos = i
				break
			}
			seen++
		}
	}
	newParams := make([]*Property, 0, len(fx.Parameters)+1)
	newParams = append(newParams, fx.Parameters[:pos]...)
	newParams = append(newParams, prop)
	newParams = append(newParams, fx.Parameters[pos:]...)
	fx.Parameters = newParams

	if newWarn := newWarningsSince(layer, oldWarningsLen); len(newWarn) > 0 {
		rollback()
		return nil, fmt.Errorf("SetEffectParam: produced %d parser warning(s), rolled back: %v", len(newWarn), newWarn)
	}
	return prop, nil
}
