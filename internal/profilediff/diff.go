// Package profilediff compares normalized AEP profiles by stable profile paths.
package profilediff

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/yueli-fx/aep-parser/internal/profile"
)

const SchemaVersion = 1

type Kind string

const (
	KindMissingObject        Kind = "missing_object"
	KindExtraObject          Kind = "extra_object"
	KindWrongValue           Kind = "wrong_value"
	KindUnsupportedConstruct Kind = "unsupported_construct"
	KindUnstablePath         Kind = "unstable_path"
	KindUnknownField         Kind = "unknown_field"
	KindRenderMismatch       Kind = "render_mismatch"
)

type Severity string

const (
	SeverityBlocker    Severity = "blocker"
	SeverityFidelity   Severity = "fidelity"
	SeverityPolish     Severity = "polish"
	SeverityAcceptable Severity = "acceptable"
	SeverityUnknown    Severity = "unknown"
)

type ActionType string

const (
	ActionParse       ActionType = "parse"
	ActionWrite       ActionType = "write"
	ActionSemantics   ActionType = "semantics"
	ActionRender      ActionType = "render"
	ActionAsset       ActionType = "asset"
	ActionPlugin      ActionType = "plugin"
	ActionIgnore      ActionType = "ignore"
	ActionInvestigate ActionType = "investigate"
)

type Options struct {
	IgnoreRules *IgnoreRules
}

type Report struct {
	SchemaVersion int    `json:"schema_version"`
	DiffCount     int    `json:"diff_count"`
	IgnoredCount  int    `json:"ignored_count,omitempty"`
	Diffs         []Diff `json:"diffs,omitempty"`
}

type Diff struct {
	Path       string           `json:"path"`
	Kind       Kind             `json:"kind"`
	Severity   Severity         `json:"severity"`
	Evidence   profile.Evidence `json:"evidence"`
	Expected   any              `json:"expected,omitempty"`
	Actual     any              `json:"actual,omitempty"`
	ActionType ActionType       `json:"action_type"`
	HumanNotes string           `json:"human_notes,omitempty"`
}

func Compare(expected, actual *profile.Profile, opts Options) (*Report, error) {
	if expected == nil {
		return nil, fmt.Errorf("profilediff: nil expected profile")
	}
	if actual == nil {
		return nil, fmt.Errorf("profilediff: nil actual profile")
	}
	if opts.IgnoreRules != nil {
		if err := opts.IgnoreRules.Validate(); err != nil {
			return nil, err
		}
	}

	var diffs []Diff
	add := func(path string, kind Kind, severity Severity, action ActionType, expectedValue, actualValue any) {
		diffs = append(diffs, Diff{
			Path:       path,
			Kind:       kind,
			Severity:   severity,
			ActionType: action,
			Expected:   expectedValue,
			Actual:     actualValue,
			Evidence: profile.Evidence{
				Level:      profile.EvidenceL1Parsed,
				Source:     "internal/profilediff",
				Confidence: "high",
			},
		})
	}
	compareValue := func(path string, expectedValue, actualValue any, severity Severity, action ActionType) {
		if !reflect.DeepEqual(expectedValue, actualValue) {
			add(path, KindWrongValue, severity, action, expectedValue, actualValue)
		}
	}

	compareValue("schema_version", expected.SchemaVersion, actual.SchemaVersion, SeverityUnknown, ActionInvestigate)
	compareValue("fingerprint.comp_count", expected.Fingerprint.CompCount, actual.Fingerprint.CompCount, SeverityUnknown, ActionInvestigate)
	compareValue("fingerprint.layer_count", expected.Fingerprint.LayerCount, actual.Fingerprint.LayerCount, SeverityUnknown, ActionInvestigate)
	compareValue("fingerprint.footage_count", expected.Fingerprint.FootageCount, actual.Fingerprint.FootageCount, SeverityUnknown, ActionInvestigate)

	compareItems("items.footage", expected.Items.Footage, actual.Items.Footage, add, compareValue)

	actualCompPaths := compPathMap(actual.Comps)
	actualCompNames := compNameMap(actual.Comps)
	usedComps := map[int]bool{}
	for _, ec := range expected.Comps {
		ac, ai, ok := matchComp(ec, actual.Comps, actualCompPaths, actualCompNames, usedComps)
		path := compObjectPath(ec)
		if !ok {
			add(path, KindMissingObject, SeverityFidelity, ActionWrite, ec.Name, nil)
			continue
		}
		usedComps[ai] = true
		compareComp(ec, ac, add, compareValue)
	}
	for i, ac := range actual.Comps {
		if !usedComps[i] {
			add(compObjectPath(ac), KindExtraObject, SeverityUnknown, ActionInvestigate, nil, ac.Name)
		}
	}

	sort.Slice(diffs, func(i, j int) bool {
		if diffs[i].Path != diffs[j].Path {
			return diffs[i].Path < diffs[j].Path
		}
		return diffs[i].Kind < diffs[j].Kind
	})

	report := &Report{SchemaVersion: SchemaVersion}
	for _, diff := range diffs {
		if opts.IgnoreRules != nil && opts.IgnoreRules.Matches(diff) {
			report.IgnoredCount++
			continue
		}
		report.Diffs = append(report.Diffs, diff)
	}
	report.DiffCount = len(report.Diffs)
	return report, nil
}

func compareComp(
	expected, actual profile.Composition,
	add func(string, Kind, Severity, ActionType, any, any),
	compareValue func(string, any, any, Severity, ActionType),
) {
	base := expected.Path.Path
	compareValue(base+".name", expected.Name, actual.Name, SeverityPolish, ActionSemantics)
	compareValue(base+".width", expected.Width, actual.Width, SeverityFidelity, ActionWrite)
	compareValue(base+".height", expected.Height, actual.Height, SeverityFidelity, ActionWrite)
	compareValue(base+".frame_rate", expected.FrameRate, actual.FrameRate, SeverityFidelity, ActionWrite)
	compareValue(base+".duration_seconds", expected.Duration, actual.Duration, SeverityFidelity, ActionWrite)
	compareValue(base+".label", expected.Label, actual.Label, SeverityPolish, ActionWrite)
	compareValue(base+".comment", expected.Comment, actual.Comment, SeverityPolish, ActionWrite)
	compareValue(base+".background_color", expected.BackgroundColor, actual.BackgroundColor, SeverityFidelity, ActionWrite)
	compareValue(base+".resolution_factor", expected.ResolutionFactor, actual.ResolutionFactor, SeverityFidelity, ActionWrite)
	compareValue(base+".pixel_aspect", expected.PixelAspect, actual.PixelAspect, SeverityFidelity, ActionWrite)
	compareValue(base+".display_start_time", expected.DisplayStartTime, actual.DisplayStartTime, SeverityFidelity, ActionWrite)
	compareValue(base+".renderer", expected.Renderer, actual.Renderer, SeverityFidelity, ActionWrite)
	compareValue(base+".draft_3d", expected.Draft3D, actual.Draft3D, SeverityFidelity, ActionWrite)
	compareValue(base+".frame_blending", expected.FrameBlending, actual.FrameBlending, SeverityFidelity, ActionWrite)
	compareValue(base+".hide_shy_layers", expected.HideShyLayers, actual.HideShyLayers, SeverityFidelity, ActionWrite)
	compareValue(base+".preserve_nested_frame_rate", expected.PreserveNestedFrameRate, actual.PreserveNestedFrameRate, SeverityFidelity, ActionWrite)
	compareValue(base+".preserve_nested_resolution", expected.PreserveNestedResolution, actual.PreserveNestedResolution, SeverityFidelity, ActionWrite)
	compareValue(base+".work_area", expected.WorkArea, actual.WorkArea, SeverityFidelity, ActionWrite)
	compareValue(base+".motion_blur", expected.MotionBlur, actual.MotionBlur, SeverityFidelity, ActionWrite)
	compareValue(base+".motion_graphics_template_name", expected.MotionGraphicsTemplateName, actual.MotionGraphicsTemplateName, SeverityFidelity, ActionWrite)

	actualLayerPaths := layerPathMap(actual.Layers)
	actualLayerKeys := layerKeyMap(actual.Layers)
	actualLayerIndexes := layerIndexMap(actual.Layers)
	usedLayers := map[int]bool{}
	for _, el := range expected.Layers {
		al, ai, ok := matchLayer(el, actual.Layers, actualLayerPaths, actualLayerKeys, actualLayerIndexes, usedLayers)
		path := layerObjectPath(el)
		if !ok {
			add(path, KindMissingObject, SeverityFidelity, ActionWrite, el.Name, nil)
			continue
		}
		usedLayers[ai] = true
		compareLayer(el, al, add, compareValue)
	}
	for i, al := range actual.Layers {
		if !usedLayers[i] {
			add(layerObjectPath(al), KindExtraObject, SeverityUnknown, ActionInvestigate, nil, al.Name)
		}
	}
}

func compareItems(
	base string,
	expected, actual []profile.Item,
	add func(string, Kind, Severity, ActionType, any, any),
	compareValue func(string, any, any, Severity, ActionType),
) {
	actualKeys := itemKeyMap(actual)
	used := map[int]bool{}
	for _, ei := range expected {
		ai, ok := matchItem(ei, actual, actualKeys, used)
		path := itemObjectPath(base, ei)
		if !ok {
			add(path, KindMissingObject, SeverityFidelity, ActionWrite, ei.Name, nil)
			continue
		}
		used[ai] = true
		compareItem(path, ei, actual[ai], compareValue)
	}
	for i, ai := range actual {
		if !used[i] {
			add(itemObjectPath(base, ai), KindExtraObject, SeverityUnknown, ActionInvestigate, nil, ai.Name)
		}
	}
}

func compareItem(
	base string,
	expected, actual profile.Item,
	compareValue func(string, any, any, Severity, ActionType),
) {
	compareValue(base+".name", expected.Name, actual.Name, SeverityPolish, ActionSemantics)
	compareValue(base+".type", expected.Type, actual.Type, SeverityFidelity, ActionWrite)
	if expected.Footage == nil || actual.Footage == nil {
		compareValue(base+".footage", expected.Footage, actual.Footage, SeverityFidelity, ActionWrite)
		return
	}
	compareValue(base+".footage.asset_type", expected.Footage.AssetType, actual.Footage.AssetType, SeverityFidelity, ActionWrite)
	compareValue(base+".footage.width", expected.Footage.Width, actual.Footage.Width, SeverityFidelity, ActionWrite)
	compareValue(base+".footage.height", expected.Footage.Height, actual.Footage.Height, SeverityFidelity, ActionWrite)
	compareValue(base+".footage.solid_color", expected.Footage.SolidColor, actual.Footage.SolidColor, SeverityFidelity, ActionWrite)
}

func compareLayer(
	expected, actual profile.Layer,
	add func(string, Kind, Severity, ActionType, any, any),
	compareValue func(string, any, any, Severity, ActionType),
) {
	base := expected.Path.Path
	compareValue(base+".name", expected.Name, actual.Name, SeverityPolish, ActionSemantics)
	compareValue(base+".type", expected.Type, actual.Type, SeverityFidelity, ActionWrite)
	compareValue(base+".label", expected.Label, actual.Label, SeverityPolish, ActionWrite)
	compareValue(base+".comment", expected.Comment, actual.Comment, SeverityPolish, ActionWrite)
	compareValue(base+".quality", expected.Quality, actual.Quality, SeverityFidelity, ActionWrite)
	compareValue(base+".blending_mode", expected.BlendingMode, actual.BlendingMode, SeverityFidelity, ActionWrite)
	compareValue(base+".auto_orient", expected.AutoOrient, actual.AutoOrient, SeverityFidelity, ActionWrite)
	compareValue(base+".light_kind", expected.LightKind, actual.LightKind, SeverityFidelity, ActionWrite)
	compareRefValue(base+".source_ref", expected.SourceRef, actual.SourceRef, SeverityFidelity, ActionWrite, compareValue)
	compareLayerRefValue(base+".light_source_ref", expected.LightSourceRef, actual.LightSourceRef, SeverityFidelity, ActionWrite, compareValue)
	compareLayerRefValue(base+".parent_ref", expected.ParentRef, actual.ParentRef, SeverityFidelity, ActionWrite, compareValue)
	compareLayerRefValue(base+".matte_ref", expected.MatteRef, actual.MatteRef, SeverityFidelity, ActionWrite, compareValue)
	compareValue(base+".timing.start_time_seconds", expected.Timing.StartTime, actual.Timing.StartTime, SeverityFidelity, ActionWrite)
	compareValue(base+".timing.duration_seconds", expected.Timing.Duration, actual.Timing.Duration, SeverityFidelity, ActionWrite)
	compareValue(base+".timing.in_point_seconds", expected.Timing.InPoint, actual.Timing.InPoint, SeverityFidelity, ActionWrite)
	compareValue(base+".timing.out_point_seconds", expected.Timing.OutPoint, actual.Timing.OutPoint, SeverityFidelity, ActionWrite)
	compareValue(base+".timing.stretch", expected.Timing.Stretch, actual.Timing.Stretch, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.visible", expected.Flags.Visible, actual.Flags.Visible, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.blend", expected.Flags.Blend, actual.Flags.Blend, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.blend_name", expected.Flags.BlendName, actual.Flags.BlendName, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.track_matte", expected.Flags.TrackMatte, actual.Flags.TrackMatte, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.track_matte_name", expected.Flags.TrackMatteName, actual.Flags.TrackMatteName, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.is_3d", expected.Flags.Is3D, actual.Flags.Is3D, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.solo", expected.Flags.Solo, actual.Flags.Solo, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.shy", expected.Flags.Shy, actual.Flags.Shy, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.locked", expected.Flags.Locked, actual.Flags.Locked, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.is_adjustment", expected.Flags.IsAdjustment, actual.Flags.IsAdjustment, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.is_null", expected.Flags.IsNull, actual.Flags.IsNull, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.is_guide", expected.Flags.IsGuide, actual.Flags.IsGuide, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.motion_blur", expected.Flags.MotionBlur, actual.Flags.MotionBlur, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.effects_enabled", expected.Flags.EffectsEnabled, actual.Flags.EffectsEnabled, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.audio_enabled", expected.Flags.AudioEnabled, actual.Flags.AudioEnabled, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.frame_blend_enabled", expected.Flags.FrameBlendEnabled, actual.Flags.FrameBlendEnabled, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.markers_locked", expected.Flags.MarkersLocked, actual.Flags.MarkersLocked, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.frame_blend_pixel_motion", expected.Flags.FrameBlendPixelMotion, actual.Flags.FrameBlendPixelMotion, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.collapse_transform", expected.Flags.CollapseTransform, actual.Flags.CollapseTransform, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.sampling_bicubic", expected.Flags.SamplingBicubic, actual.Flags.SamplingBicubic, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.preserve_transparency", expected.Flags.PreserveTransparency, actual.Flags.PreserveTransparency, SeverityFidelity, ActionWrite)

	actualEffects := effectKeyMap(actual.Effects)
	usedEffects := map[int]bool{}
	for _, ee := range expected.Effects {
		ae, ai, ok := matchEffect(ee, actual.Effects, actualEffects, usedEffects)
		path := effectObjectPath(ee)
		if !ok {
			add(path, KindMissingObject, SeverityFidelity, ActionWrite, ee.MatchName, nil)
			continue
		}
		usedEffects[ai] = true
		compareEffect(ee, ae, add, compareValue)
	}
	for i, ae := range actual.Effects {
		if !usedEffects[i] {
			add(effectObjectPath(ae), KindExtraObject, SeverityUnknown, ActionInvestigate, nil, ae.MatchName)
		}
	}

	actualProperties := propertyKeyMap(actual.Properties)
	usedProperties := map[int]bool{}
	for _, ep := range expected.Properties {
		ap, ai, ok := matchProperty(ep, actual.Properties, actualProperties, usedProperties)
		path := propertyObjectPath(ep)
		if !ok {
			add(path, KindMissingObject, SeverityFidelity, ActionWrite, ep.MatchName, nil)
			continue
		}
		usedProperties[ai] = true
		compareProperty(ep, ap, compareValue)
	}
	for i, ap := range actual.Properties {
		if !usedProperties[i] {
			add(propertyObjectPath(ap), KindExtraObject, SeverityUnknown, ActionInvestigate, nil, ap.MatchName)
		}
	}

	compareValue(base+".text", expected.Text, actual.Text, SeverityFidelity, ActionWrite)
	compareValue(base+".masks", len(expected.Masks), len(actual.Masks), SeverityFidelity, ActionWrite)
	compareValue(base+".shapes", len(expected.Shapes), len(actual.Shapes), SeverityFidelity, ActionWrite)
}

func compareRefValue(
	path string,
	expected *profile.ItemRef,
	actual *profile.ItemRef,
	severity Severity,
	action ActionType,
	compareValue func(string, any, any, Severity, ActionType),
) {
	if itemRefsEqual(expected, actual) {
		return
	}
	compareValue(path, expected, actual, severity, action)
}

func compareLayerRefValue(
	path string,
	expected *profile.LayerRef,
	actual *profile.LayerRef,
	severity Severity,
	action ActionType,
	compareValue func(string, any, any, Severity, ActionType),
) {
	if layerRefsEqual(expected, actual) {
		return
	}
	compareValue(path, expected, actual, severity, action)
}

func itemRefsEqual(expected, actual *profile.ItemRef) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}
	if expected.Kind == actual.Kind && expected.Name != "" && actual.Name != "" {
		return expected.Name == actual.Name
	}
	return reflect.DeepEqual(expected, actual)
}

func layerRefsEqual(expected, actual *profile.LayerRef) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}
	if expected.Name != "" && actual.Name != "" {
		return expected.Name == actual.Name
	}
	if expected.Index != 0 && actual.Index != 0 {
		return expected.Index == actual.Index
	}
	return reflect.DeepEqual(expected, actual)
}

func compareEffect(
	expected, actual profile.Effect,
	add func(string, Kind, Severity, ActionType, any, any),
	compareValue func(string, any, any, Severity, ActionType),
) {
	base := expected.Path.Path
	compareValue(base+".match_name", expected.MatchName, actual.MatchName, SeverityFidelity, ActionWrite)
	compareValue(base+".dependency_class", expected.DependencyClass, actual.DependencyClass, SeverityFidelity, ActionPlugin)

	actualParams := propertyKeyMap(actual.Params)
	usedParams := map[int]bool{}
	for _, ep := range expected.Params {
		ap, ai, ok := matchProperty(ep, actual.Params, actualParams, usedParams)
		path := propertyObjectPath(ep)
		if !ok {
			add(path, KindMissingObject, SeverityFidelity, ActionWrite, ep.MatchName, nil)
			continue
		}
		usedParams[ai] = true
		compareProperty(ep, ap, compareValue)
	}
	for i, ap := range actual.Params {
		if !usedParams[i] {
			add(propertyObjectPath(ap), KindExtraObject, SeverityUnknown, ActionInvestigate, nil, ap.MatchName)
		}
	}
}

func compareProperty(
	expected, actual profile.Property,
	compareValue func(string, any, any, Severity, ActionType),
) {
	base := expected.Path.Path
	compareValue(base+".name", expected.Name, actual.Name, SeverityPolish, ActionSemantics)
	compareValue(base+".match_name", expected.MatchName, actual.MatchName, SeverityFidelity, ActionWrite)
	compareValue(base+".occurrence", expected.Occurrence, actual.Occurrence, SeverityUnknown, ActionInvestigate)
	compareValue(base+".static_value", expected.StaticValue, actual.StaticValue, SeverityFidelity, ActionWrite)
	compareLayerRefValue(base+".layer_ref", expected.LayerRef, actual.LayerRef, SeverityFidelity, ActionWrite, compareValue)
	compareValue(base+".default", expected.Default, actual.Default, SeverityFidelity, ActionWrite)
	compareValue(base+".changed", expected.Changed, actual.Changed, SeverityFidelity, ActionWrite)
	compareValue(base+".expression", expected.Expression, actual.Expression, SeverityFidelity, ActionWrite)
	compareValue(base+".expression_enabled", expected.ExpressionEnabled, actual.ExpressionEnabled, SeverityFidelity, ActionWrite)
	if len(expected.Keyframes) != len(actual.Keyframes) {
		compareValue(base+".keyframes", len(expected.Keyframes), len(actual.Keyframes), SeverityFidelity, ActionWrite)
		return
	}
	for i := range expected.Keyframes {
		kpath := fmt.Sprintf("%s.keyframes[%d]", base, i)
		compareValue(kpath+".time_seconds", expected.Keyframes[i].Time, actual.Keyframes[i].Time, SeverityFidelity, ActionWrite)
		compareValue(kpath+".value", expected.Keyframes[i].Value, actual.Keyframes[i].Value, SeverityFidelity, ActionWrite)
		compareValue(kpath+".in_interp", expected.Keyframes[i].InInterp, actual.Keyframes[i].InInterp, SeverityFidelity, ActionWrite)
		compareValue(kpath+".out_interp", expected.Keyframes[i].OutInterp, actual.Keyframes[i].OutInterp, SeverityFidelity, ActionWrite)
		compareValue(kpath+".in_spatial_tangent", expected.Keyframes[i].InSpatialTangent, actual.Keyframes[i].InSpatialTangent, SeverityFidelity, ActionWrite)
		compareValue(kpath+".out_spatial_tangent", expected.Keyframes[i].OutSpatialTangent, actual.Keyframes[i].OutSpatialTangent, SeverityFidelity, ActionWrite)
		compareValue(kpath+".in_temporal_ease", expected.Keyframes[i].InTemporalEase, actual.Keyframes[i].InTemporalEase, SeverityFidelity, ActionWrite)
		compareValue(kpath+".out_temporal_ease", expected.Keyframes[i].OutTemporalEase, actual.Keyframes[i].OutTemporalEase, SeverityFidelity, ActionWrite)
	}
}

func objectPath(path, fallback string) string {
	if path != "" {
		return path
	}
	return fallback
}

func compObjectPath(comp profile.Composition) string {
	return objectPath(comp.Path.Path, fmt.Sprintf("comps.by_id[%d]", comp.ID))
}

func layerObjectPath(layer profile.Layer) string {
	return objectPath(layer.Path.Path, fmt.Sprintf("layers.by_id[%d]", layer.ID))
}

func effectObjectPath(effect profile.Effect) string {
	return objectPath(effect.Path.Path, fmt.Sprintf(`effects.by_match_name[%q]#%d`, effect.MatchName, effect.Occurrence))
}

func propertyObjectPath(prop profile.Property) string {
	return objectPath(prop.Path.Path, fmt.Sprintf(`properties.by_match_name[%q]#%d`, prop.MatchName, prop.Occurrence))
}

func itemObjectPath(base string, item profile.Item) string {
	return objectPath(item.Path.Path, fmt.Sprintf(`%s.by_name[%q]`, base, item.Name))
}

func itemKeyMap(items []profile.Item) map[string]int {
	return uniqueIndexMap(len(items), func(i int) string {
		return fmt.Sprintf("%s:%s", items[i].Type, items[i].Name)
	})
}

func compPathMap(comps []profile.Composition) map[string]int {
	out := map[string]int{}
	for i, comp := range comps {
		out[compObjectPath(comp)] = i
	}
	return out
}

func compNameMap(comps []profile.Composition) map[string]int {
	return uniqueIndexMap(len(comps), func(i int) string { return comps[i].Name })
}

func layerPathMap(layers []profile.Layer) map[string]int {
	out := map[string]int{}
	for i, layer := range layers {
		out[layerObjectPath(layer)] = i
	}
	return out
}

func layerKeyMap(layers []profile.Layer) map[string]int {
	return uniqueIndexMap(len(layers), func(i int) string {
		return fmt.Sprintf("%d:%s", layers[i].Index, layers[i].Name)
	})
}

func layerIndexMap(layers []profile.Layer) map[int]int {
	out := map[int]int{}
	dupes := map[int]bool{}
	for i, layer := range layers {
		if _, ok := out[layer.Index]; ok {
			dupes[layer.Index] = true
			continue
		}
		out[layer.Index] = i
	}
	for index := range dupes {
		delete(out, index)
	}
	return out
}

func effectKeyMap(effects []profile.Effect) map[string]int {
	return uniqueIndexMap(len(effects), func(i int) string {
		return fmt.Sprintf("%s#%d", effects[i].MatchName, effects[i].Occurrence)
	})
}

func propertyKeyMap(properties []profile.Property) map[string]int {
	return uniqueIndexMap(len(properties), func(i int) string {
		return fmt.Sprintf("%s#%d", properties[i].MatchName, properties[i].Occurrence)
	})
}

func uniqueIndexMap(n int, key func(int) string) map[string]int {
	out := map[string]int{}
	dupes := map[string]bool{}
	for i := 0; i < n; i++ {
		k := key(i)
		if k == "" {
			continue
		}
		if _, ok := out[k]; ok {
			dupes[k] = true
			continue
		}
		out[k] = i
	}
	for k := range dupes {
		delete(out, k)
	}
	return out
}

func matchItem(expected profile.Item, actual []profile.Item, actualKeys map[string]int, used map[int]bool) (int, bool) {
	if i, ok := actualKeys[fmt.Sprintf("%s:%s", expected.Type, expected.Name)]; ok && !used[i] {
		return i, true
	}
	return 0, false
}

func matchComp(
	expected profile.Composition,
	actual []profile.Composition,
	byPath map[string]int,
	byName map[string]int,
	used map[int]bool,
) (profile.Composition, int, bool) {
	if i, ok := byPath[compObjectPath(expected)]; ok && !used[i] {
		return actual[i], i, true
	}
	if i, ok := byName[expected.Name]; ok && !used[i] {
		return actual[i], i, true
	}
	return profile.Composition{}, 0, false
}

func matchLayer(
	expected profile.Layer,
	actual []profile.Layer,
	byPath map[string]int,
	byKey map[string]int,
	byIndex map[int]int,
	used map[int]bool,
) (profile.Layer, int, bool) {
	if i, ok := byPath[layerObjectPath(expected)]; ok && !used[i] {
		return actual[i], i, true
	}
	key := fmt.Sprintf("%d:%s", expected.Index, expected.Name)
	if i, ok := byKey[key]; ok && !used[i] {
		return actual[i], i, true
	}
	if i, ok := byIndex[expected.Index]; ok && !used[i] {
		return actual[i], i, true
	}
	return profile.Layer{}, 0, false
}

func matchEffect(
	expected profile.Effect,
	actual []profile.Effect,
	byKey map[string]int,
	used map[int]bool,
) (profile.Effect, int, bool) {
	key := fmt.Sprintf("%s#%d", expected.MatchName, expected.Occurrence)
	if i, ok := byKey[key]; ok && !used[i] {
		return actual[i], i, true
	}
	return profile.Effect{}, 0, false
}

func matchProperty(
	expected profile.Property,
	actual []profile.Property,
	byKey map[string]int,
	used map[int]bool,
) (profile.Property, int, bool) {
	key := fmt.Sprintf("%s#%d", expected.MatchName, expected.Occurrence)
	if i, ok := byKey[key]; ok && !used[i] {
		return actual[i], i, true
	}
	return profile.Property{}, 0, false
}
