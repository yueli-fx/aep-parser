// Package profilediff compares normalized AEP profiles by stable profile paths.
package profilediff

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/example/aep-parser/internal/profile"
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
	compareValue(base+".renderer", expected.Renderer, actual.Renderer, SeverityFidelity, ActionWrite)

	actualLayerPaths := layerPathMap(actual.Layers)
	actualLayerKeys := layerKeyMap(actual.Layers)
	usedLayers := map[int]bool{}
	for _, el := range expected.Layers {
		al, ai, ok := matchLayer(el, actual.Layers, actualLayerPaths, actualLayerKeys, usedLayers)
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

func compareLayer(
	expected, actual profile.Layer,
	add func(string, Kind, Severity, ActionType, any, any),
	compareValue func(string, any, any, Severity, ActionType),
) {
	base := expected.Path.Path
	compareValue(base+".type", expected.Type, actual.Type, SeverityFidelity, ActionWrite)
	compareValue(base+".source_ref", expected.SourceRef, actual.SourceRef, SeverityFidelity, ActionWrite)
	compareValue(base+".parent_ref", expected.ParentRef, actual.ParentRef, SeverityFidelity, ActionWrite)
	compareValue(base+".matte_ref", expected.MatteRef, actual.MatteRef, SeverityFidelity, ActionWrite)
	compareValue(base+".timing.start_time_seconds", expected.Timing.StartTime, actual.Timing.StartTime, SeverityFidelity, ActionWrite)
	compareValue(base+".timing.duration_seconds", expected.Timing.Duration, actual.Timing.Duration, SeverityFidelity, ActionWrite)
	compareValue(base+".timing.in_point_seconds", expected.Timing.InPoint, actual.Timing.InPoint, SeverityFidelity, ActionWrite)
	compareValue(base+".timing.out_point_seconds", expected.Timing.OutPoint, actual.Timing.OutPoint, SeverityFidelity, ActionWrite)
	compareValue(base+".timing.stretch", expected.Timing.Stretch, actual.Timing.Stretch, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.visible", expected.Flags.Visible, actual.Flags.Visible, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.blend", expected.Flags.Blend, actual.Flags.Blend, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.track_matte", expected.Flags.TrackMatte, actual.Flags.TrackMatte, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.is_3d", expected.Flags.Is3D, actual.Flags.Is3D, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.solo", expected.Flags.Solo, actual.Flags.Solo, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.shy", expected.Flags.Shy, actual.Flags.Shy, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.locked", expected.Flags.Locked, actual.Flags.Locked, SeverityFidelity, ActionWrite)
	compareValue(base+".flags.motion_blur", expected.Flags.MotionBlur, actual.Flags.MotionBlur, SeverityFidelity, ActionWrite)

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

	compareValue(base+".text", expected.Text, actual.Text, SeverityFidelity, ActionWrite)
	compareValue(base+".masks", len(expected.Masks), len(actual.Masks), SeverityFidelity, ActionWrite)
	compareValue(base+".shapes", len(expected.Shapes), len(actual.Shapes), SeverityFidelity, ActionWrite)
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
	compareValue(base+".static_value", expected.StaticValue, actual.StaticValue, SeverityFidelity, ActionWrite)
	compareValue(base+".expression", expected.Expression, actual.Expression, SeverityFidelity, ActionWrite)
	if len(expected.Keyframes) != len(actual.Keyframes) {
		compareValue(base+".keyframes", len(expected.Keyframes), len(actual.Keyframes), SeverityFidelity, ActionWrite)
		return
	}
	for i := range expected.Keyframes {
		kpath := fmt.Sprintf("%s.keyframes[%d]", base, i)
		compareValue(kpath+".time_seconds", expected.Keyframes[i].Time, actual.Keyframes[i].Time, SeverityFidelity, ActionWrite)
		compareValue(kpath+".value", expected.Keyframes[i].Value, actual.Keyframes[i].Value, SeverityFidelity, ActionWrite)
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
	used map[int]bool,
) (profile.Layer, int, bool) {
	if i, ok := byPath[layerObjectPath(expected)]; ok && !used[i] {
		return actual[i], i, true
	}
	key := fmt.Sprintf("%d:%s", expected.Index, expected.Name)
	if i, ok := byKey[key]; ok && !used[i] {
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
