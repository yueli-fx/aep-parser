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
// Layers without an Effect Parade (AE only emits the parade once ≥1 effect
// exists, so every effect-less layer lacks it) get an empty parade spliced into
// their Layr property tree first — immediately before "ADBE Transform Group",
// matching AE's emitted group order (RE: re_shape_effect.aep +
// re_effect_library.aep). From-scratch layers built by NewShapeLayer have no
// parsed property tree to splice into (and dirty shape layers are re-lowered
// from scene state at write, discarding chunk edits) — AddEffect refuses those;
// aep.Reopen upgrades them to parsed layers.
package serializer

import (
	"bytes"
	"embed"
	"encoding/binary"
	"fmt"
	"sort"
	"sync"

	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

// effectTemplateFS holds the embedded effect templates. Each is a LIST(tdgp)
// wrapper around an effect's (tdmn match-name, LIST:sspc payload) pair,
// extracted verbatim from an AE-2020-saved layer that had the full curated set
// applied (test_data/re_effect_library.aep + re_effect_library2.aep, generated
// by re_effect_library.jsx / re_effect_library2.jsx; Point3D Control from
// re_effect_param_types.aep — all extracted by tmp_debug/extract_effect_lib).
// All are built-in effects present well before the 2020 read floor; their
// serialized form is version-portable (AE 2025 accepts the AE-2020 bytes —
// mirroring the gradient version-portability finding, confirmed by ship-gate
// across the full template library on both versions).
//
//go:embed templates/effect_adbe_gaussian_blur_2.bin templates/effect_adbe_fill.bin templates/effect_adbe_tint.bin templates/effect_adbe_brightness_contrast_2.bin templates/effect_adbe_tritone.bin templates/effect_adbe_easy_levels2.bin templates/effect_adbe_pro_levels2.bin templates/effect_adbe_hue_saturation.bin templates/effect_adbe_box_blur.bin templates/effect_adbe_glo2.bin templates/effect_adbe_invert.bin templates/effect_adbe_exposure2.bin
//go:embed templates/effect_adbe_drop_shadow.bin templates/effect_adbe_sharpen.bin templates/effect_adbe_mosaic.bin templates/effect_adbe_noise.bin templates/effect_adbe_geometry2.bin templates/effect_adbe_ramp.bin templates/effect_adbe_fractal_noise.bin templates/effect_adbe_tile.bin templates/effect_adbe_motion_blur.bin templates/effect_adbe_linear_wipe.bin templates/effect_adbe_wave_warp.bin templates/effect_adbe_curvescustom.bin templates/effect_adbe_slider_control.bin templates/effect_adbe_point_control.bin templates/effect_adbe_color_control.bin templates/effect_adbe_angle_control.bin templates/effect_adbe_checkbox_control.bin templates/effect_adbe_point3d_control.bin
//go:embed templates/effect_adbe_set_matte3.bin
//go:embed templates/effect_adbe_turbulent_displace.bin templates/effect_adbe_roughen_edges.bin templates/effect_adbe_echo.bin templates/effect_adbe_radial_blur.bin templates/effect_adbe_4colorgradient.bin templates/effect_adbe_checkerboard.bin templates/effect_adbe_grid.bin templates/effect_adbe_stroke.bin templates/effect_adbe_corner_pin.bin templates/effect_adbe_venetian_blinds.bin
var effectTemplateFS embed.FS

// Effect match-name constants for the addable built-in set. These are AE's
// stable internal match-names; use them with AddEffect instead of hardcoding
// strings. The display name (what the AE Effects panel shows) is in the comment.
const (
	EffectGaussianBlur       = "ADBE Gaussian Blur 2"         // Gaussian Blur
	EffectFill               = "ADBE Fill"                    // Fill
	EffectTint               = "ADBE Tint"                    // Tint
	EffectBrightnessContrast = "ADBE Brightness & Contrast 2" // Brightness & Contrast
	EffectTritone            = "ADBE Tritone"                 // Tritone
	EffectLevels             = "ADBE Easy Levels2"            // Levels
	EffectLevelsIndividual   = "ADBE Pro Levels2"             // Levels (Individual Controls)
	EffectHueSaturation      = "ADBE HUE SATURATION"          // Hue/Saturation
	EffectBoxBlur            = "ADBE Box Blur"                // Fast Box Blur
	EffectGlow               = "ADBE Glo2"                    // Glow
	EffectInvert             = "ADBE Invert"                  // Invert
	EffectExposure           = "ADBE Exposure2"               // Exposure
	EffectDropShadow         = "ADBE Drop Shadow"             // Drop Shadow
	EffectSharpen            = "ADBE Sharpen"                 // Sharpen
	EffectMosaic             = "ADBE Mosaic"                  // Mosaic
	EffectNoise              = "ADBE Noise"                   // Noise
	EffectTransform          = "ADBE Geometry2"               // Transform
	EffectGradientRamp       = "ADBE Ramp"                    // Gradient Ramp
	EffectFractalNoise       = "ADBE Fractal Noise"           // Fractal Noise
	EffectMotionTile         = "ADBE Tile"                    // Motion Tile
	EffectDirectionalBlur    = "ADBE Motion Blur"             // Directional Blur
	EffectLinearWipe         = "ADBE Linear Wipe"             // Linear Wipe
	EffectWaveWarp           = "ADBE Wave Warp"               // Wave Warp
	EffectCurves             = "ADBE CurvesCustom"            // Curves
	EffectSliderControl      = "ADBE Slider Control"          // Slider Control
	EffectPointControl       = "ADBE Point Control"           // Point Control
	EffectColorControl       = "ADBE Color Control"           // Color Control
	EffectAngleControl       = "ADBE Angle Control"           // Angle Control
	EffectCheckboxControl    = "ADBE Checkbox Control"        // Checkbox Control
	EffectPoint3DControl     = "ADBE Point3D Control"         // 3D Point Control
	EffectSetMatte           = "ADBE Set Matte3"              // Set Matte (layer-reference effect)
	// Wave 4 (2026-06-15, fixture re_effect_lib3.aep) — MG distort/generate/
	// stylize/transition.
	EffectTurbulentDisplace = "ADBE Turbulent Displace" // Turbulent Displace
	EffectRoughenEdges      = "ADBE Roughen Edges"      // Roughen Edges
	EffectEcho              = "ADBE Echo"               // Echo
	EffectRadialBlur        = "ADBE Radial Blur"        // Radial Blur
	EffectFourColorGradient = "ADBE 4ColorGradient"     // 4-Color Gradient
	EffectCheckerboard      = "ADBE Checkerboard"       // Checkerboard
	EffectGrid              = "ADBE Grid"               // Grid
	EffectStroke            = "ADBE Stroke"             // Stroke
	EffectCornerPin         = "ADBE Corner Pin"         // Corner Pin
	EffectVenetianBlinds    = "ADBE Venetian Blinds"    // Venetian Blinds
)

// effectTemplateFiles maps an effect match-name to its embedded template path.
// Each entry is an AE-native effect instance with AE's default parameter values;
// callers tune parameters afterward via the returned Effect.Parameters
// (Property.SetStaticValue works on effect params).
var effectTemplateFiles = map[string]string{
	EffectGaussianBlur:       "templates/effect_adbe_gaussian_blur_2.bin",
	EffectFill:               "templates/effect_adbe_fill.bin",
	EffectTint:               "templates/effect_adbe_tint.bin",
	EffectBrightnessContrast: "templates/effect_adbe_brightness_contrast_2.bin",
	EffectTritone:            "templates/effect_adbe_tritone.bin",
	EffectLevels:             "templates/effect_adbe_easy_levels2.bin",
	EffectLevelsIndividual:   "templates/effect_adbe_pro_levels2.bin",
	EffectHueSaturation:      "templates/effect_adbe_hue_saturation.bin",
	EffectBoxBlur:            "templates/effect_adbe_box_blur.bin",
	EffectGlow:               "templates/effect_adbe_glo2.bin",
	EffectInvert:             "templates/effect_adbe_invert.bin",
	EffectExposure:           "templates/effect_adbe_exposure2.bin",
	EffectDropShadow:         "templates/effect_adbe_drop_shadow.bin",
	EffectSharpen:            "templates/effect_adbe_sharpen.bin",
	EffectMosaic:             "templates/effect_adbe_mosaic.bin",
	EffectNoise:              "templates/effect_adbe_noise.bin",
	EffectTransform:          "templates/effect_adbe_geometry2.bin",
	EffectGradientRamp:       "templates/effect_adbe_ramp.bin",
	EffectFractalNoise:       "templates/effect_adbe_fractal_noise.bin",
	EffectMotionTile:         "templates/effect_adbe_tile.bin",
	EffectDirectionalBlur:    "templates/effect_adbe_motion_blur.bin",
	EffectLinearWipe:         "templates/effect_adbe_linear_wipe.bin",
	EffectWaveWarp:           "templates/effect_adbe_wave_warp.bin",
	EffectCurves:             "templates/effect_adbe_curvescustom.bin",
	EffectSliderControl:      "templates/effect_adbe_slider_control.bin",
	EffectPointControl:       "templates/effect_adbe_point_control.bin",
	EffectColorControl:       "templates/effect_adbe_color_control.bin",
	EffectAngleControl:       "templates/effect_adbe_angle_control.bin",
	EffectCheckboxControl:    "templates/effect_adbe_checkbox_control.bin",
	EffectPoint3DControl:     "templates/effect_adbe_point3d_control.bin",
	EffectSetMatte:           "templates/effect_adbe_set_matte3.bin",
	EffectTurbulentDisplace:  "templates/effect_adbe_turbulent_displace.bin",
	EffectRoughenEdges:       "templates/effect_adbe_roughen_edges.bin",
	EffectEcho:               "templates/effect_adbe_echo.bin",
	EffectRadialBlur:         "templates/effect_adbe_radial_blur.bin",
	EffectFourColorGradient:  "templates/effect_adbe_4colorgradient.bin",
	EffectCheckerboard:       "templates/effect_adbe_checkerboard.bin",
	EffectGrid:               "templates/effect_adbe_grid.bin",
	EffectStroke:             "templates/effect_adbe_stroke.bin",
	EffectCornerPin:          "templates/effect_adbe_corner_pin.bin",
	EffectVenetianBlinds:     "templates/effect_adbe_venetian_blinds.bin",
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
	names := make([]string, 0, len(effectTemplateFiles))
	for k := range effectTemplateFiles {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// cloneEffectTemplate parses the embedded template for matchName once (cached)
// and returns a deep clone of its (tdmn, sspc) pair so the spliced chunks never
// alias the cache (concurrent-mutate safe).
func cloneEffectTemplate(matchName string) (tdmn, sspc *rifx.Chunk, err error) {
	path, ok := effectTemplateFiles[matchName]
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
		raw, e := effectTemplateFS.ReadFile(path)
		if e != nil {
			ct.err = fmt.Errorf("read effect template %q (%s): %w", matchName, path, e)
			return
		}
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

// retargetEffectHostLayer rewrites every tdpi chunk (the 4-byte host-layer
// binding each effect param's tdbs carries) in the cloned template payload to
// the destination layer's ID. The embedded templates hold the extraction
// fixture's host layer id verbatim; AE validates the binding resolves on open
// and rejects the project with "cannot find layer ID=N in composition" when it
// dangles (RE'd 2026-06-10: re_effect_library host id 15, re_shape_effect host
// id 13 — tdpi tracks the host in both).
func retargetEffectHostLayer(c *rifx.Chunk, layerID uint32) {
	if c.ID == rifx.IDTdpi && len(c.Data) >= 4 {
		binary.BigEndian.PutUint32(c.Data[0:4], layerID)
	}
	for _, ch := range c.Children {
		retargetEffectHostLayer(ch, layerID)
	}
}

// RemoveEffect removes the effect at the given 0-based index from the layer's
// Effect Parade. Thin index-validated wrapper over RemovePropertyGroup (which is
// AE 2020 + AE 2025 ship-gate green for Effect-Parade child removal), giving
// AddEffect a symmetric inverse.
// (Full contract lives on the aep.RemoveEffect facade — docgen source.)
func RemoveEffect(layer *Layer, index int) error {
	if layer == nil {
		return fmt.Errorf("RemoveEffect: layer is nil")
	}
	parade := layer.EffectsParade()
	if parade == nil {
		return fmt.Errorf("RemoveEffect: layer %q has no Effect Parade group", layer.Name)
	}
	n := parade.NumProperties()
	if index < 0 || index >= n {
		return fmt.Errorf("RemoveEffect: index %d out of range (have %d effects)", index, n)
	}
	g, ok := parade.ChildByIndex(index).(*AEPropertyGroup)
	if !ok {
		return fmt.Errorf("RemoveEffect: effect at index %d is not a property group", index)
	}
	return RemovePropertyGroup(g)
}

// aeDefaultGroupName is the placeholder AE persists in a group's tdsn when the
// user never renamed it — "-_0_/-" observed on every AE-native Effect Parade
// (re_shape_effect.aep, re_effect_library.aep).
const aeDefaultGroupName = "-_0_/-"

// ensureEffectParade returns the layer's Effect Parade, splicing a fresh empty
// parade group (tdsb 0x01 + tdsn "-_0_/-" + Group End sentinel) into the Layr
// property tree when absent — immediately before "ADBE Transform Group",
// matching AE's emitted group order. The returned undo restores the pre-create
// chunk + scene-tree state (no-op when the parade already existed).
func ensureEffectParade(layer *Layer) (*AEPropertyGroup, func(), error) {
	if parade := layer.EffectsParade(); parade != nil {
		return parade, func() {}, nil
	}
	return spliceEmptyParade(layer, "AddEffect", MatchNameGroupEffectParade, []string{MatchNameGroupTransform})
}

// spliceEmptyParade splices a fresh empty parade group (tdsb 0x01 + tdsn
// "-_0_/-" + Group End sentinel) named paradeMatchName into the layer's Layr
// property tree, immediately before the first anchor group found (tried in
// anchorMatchNames order), mirroring the splice in the scene tree. The
// returned undo restores the pre-create chunk + scene-tree state.
func spliceEmptyParade(layer *Layer, opName, paradeMatchName string, anchorMatchNames []string) (*AEPropertyGroup, func(), error) {
	tree := layer.PropertyTree()
	if tree == nil {
		return nil, nil, fmt.Errorf("%s: layer %q was built outside the parser (no property tree to hold a %s); round-trip the project through aep.Reopen first, then mutate the re-parsed layer", opName, layer.Name, paradeMatchName)
	}
	lb := layerBack(layer)
	if lb == nil || lb.layrList == nil {
		return nil, nil, fmt.Errorf("%s: layer %q has no Layr chunk back-ref", opName, layer.Name)
	}
	var outer *rifx.Chunk
	for _, ch := range lb.layrList.Children {
		if ch.IsList() && ch.FormType == rifx.IDTdgp {
			outer = ch
			break
		}
	}
	if outer == nil {
		return nil, nil, fmt.Errorf("%s: layer %q has no property-group LIST in its Layr", opName, layer.Name)
	}
	anchor := -1
	for _, anchorName := range anchorMatchNames {
		for i, ch := range outer.Children {
			if ch.ID == rifx.IDTdmn && string(bytes.TrimRight(ch.Data, "\x00")) == anchorName {
				anchor = i
				break
			}
		}
		if anchor >= 0 {
			break
		}
	}
	if anchor < 0 {
		return nil, nil, fmt.Errorf("%s: layer %q has none of %v to anchor the %s position", opName, layer.Name, anchorMatchNames, paradeMatchName)
	}

	paradeTdgp := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp, Children: []*rifx.Chunk{
		makeTdsb(),
		makeTdsn(aeDefaultGroupName),
		makeTdmn("ADBE Group End"),
	}}

	oldOuterChildren := append([]*rifx.Chunk(nil), outer.Children...)
	oldTreeChildren := append([]PropertyBase(nil), tree.Children...)

	spliced := make([]*rifx.Chunk, 0, len(outer.Children)+2)
	spliced = append(spliced, outer.Children[:anchor]...)
	spliced = append(spliced, makeTdmn(paradeMatchName), paradeTdgp)
	spliced = append(spliced, outer.Children[anchor:]...)
	outer.Children = spliced

	parade := &AEPropertyGroup{MatchName: paradeMatchName, Name: paradeMatchName}
	scene.SetPropertyGroupParent(parade, tree)
	scene.SetPropertyGroupBack(parade, &propertyGroupBackrefs{chunk: paradeTdgp})
	treeAnchor := len(tree.Children)
	for _, anchorName := range anchorMatchNames {
		found := false
		for i, c := range tree.Children {
			if g, ok := c.(*AEPropertyGroup); ok && g.MatchName == anchorName {
				treeAnchor = i
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	newTreeChildren := make([]PropertyBase, 0, len(tree.Children)+1)
	newTreeChildren = append(newTreeChildren, tree.Children[:treeAnchor]...)
	newTreeChildren = append(newTreeChildren, parade)
	newTreeChildren = append(newTreeChildren, tree.Children[treeAnchor:]...)
	tree.Children = newTreeChildren

	return parade, func() {
		outer.Children = oldOuterChildren
		tree.Children = oldTreeChildren
	}, nil
}

// AddEffect appends an effect to the layer's Effect Parade (auto-creating the
// parade for effect-less parsed layers) and returns the parsed *Effect (so the
// caller can tune Effect.Parameters immediately).
// (Full contract + RE notes live on the aep.AddEffect facade — docgen source.)
func AddEffect(layer *Layer, effectMatchName string) (*Effect, error) {
	if layer == nil {
		return nil, fmt.Errorf("AddEffect: layer is nil")
	}
	if layer.Type == LayerTypeCamera || layer.Type == LayerTypeLight {
		return nil, fmt.Errorf("AddEffect: layer %q is a %s layer (AE does not allow effects on camera/light layers)", layer.Name, layer.Type)
	}
	parade, undoParadeCreate, err := ensureEffectParade(layer)
	if err != nil {
		return nil, err
	}
	pgb := propertyGroupBack(parade)
	if pgb == nil || pgb.chunk == nil {
		undoParadeCreate()
		return nil, fmt.Errorf("AddEffect: Effect Parade for layer %q has no chunk back-ref", layer.Name)
	}

	tdmnCh, sspcCh, err := cloneEffectTemplate(effectMatchName)
	if err != nil {
		undoParadeCreate()
		return nil, err
	}
	retargetEffectHostLayer(sspcCh, layer.ID)

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
		undoParadeCreate()
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
