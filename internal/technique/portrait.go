package technique

import (
	"fmt"
	"sort"
)

func BuildPortrait(facts *FactSet) (*Portrait, error) {
	if facts == nil {
		return nil, fmt.Errorf("technique: nil facts")
	}
	portrait := &Portrait{
		SchemaVersion: SchemaVersion,
		SourcePath:    facts.SourcePath,
		Fingerprint: FingerprintSummary{
			CompCount:          facts.Summary.CompCount,
			LayerCount:         facts.Summary.LayerCount,
			EffectCount:        facts.Summary.EffectCount,
			TextLayerCount:     facts.Summary.TextLayerCount,
			ShapeLayerCount:    facts.Summary.ShapeLayerCount,
			TextAnimatorCount:  len(facts.TextAnimators),
			ShapeOperatorCount: len(facts.ShapeOperators),
			DependencyCount:    len(facts.Dependencies),
			UnknownCount:       len(facts.Unknowns),
			LayerRoleCounts:    map[string]int{},
		},
		Mechanisms: MechanismSummary{
			EffectClassCounts:           map[string]int{},
			EffectMatchCounts:           map[string]int{},
			ThirdPartyEffectMatchCounts: map[string]int{},
			TextAnimatorKindCounts:      map[string]int{},
			ShapeFamilyCounts:           map[string]int{},
			ReproducibilityCounts:       map[string]int{},
		},
		Graph: GraphSummary{
			RelationCounts: map[string]int{},
		},
		Unknowns: UnknownSummary{Count: len(facts.Unknowns)},
	}

	layerRefs := map[layerKey]LayerFact{}
	signals := map[layerKey]map[string]bool{}
	hints := hintCollector{seen: map[string]TechniqueHint{}}

	for _, layer := range facts.Layers {
		key := layerKey{CompName: layer.CompName, LayerName: layer.Name}
		layerRefs[key] = layer
		if layer.Role != "" {
			portrait.Fingerprint.LayerRoleCounts[layer.Role]++
		}
	}
	for _, effect := range facts.Effects {
		key := layerKey{CompName: effect.CompName, LayerName: effect.LayerName}
		addSignal(signals, key, "effect:"+effectClass(effect.DependencyClass))
		portrait.Mechanisms.EffectClassCounts[effectClass(effect.DependencyClass)]++
		portrait.Mechanisms.EffectMatchCounts[effect.MatchName]++
		reproducibility := reproducibilityClass(effect.DependencyClass)
		portrait.Mechanisms.ReproducibilityCounts[reproducibility]++
		if reproducibility == "third_party" {
			portrait.Mechanisms.ThirdPartyEffectMatchCounts[effect.MatchName]++
		}
		if effect.HasKeyframes {
			addSignal(signals, key, "effect:keyframed")
		}
		if effect.HasExpression {
			addSignal(signals, key, "effect:expression")
		}
		if effect.HasLayerRef {
			addSignal(signals, key, "effect:layer_ref")
		}
		if effect.ChangedParamCount > 0 || effect.TunedParamCount > 0 || effect.HasKeyframes || effect.HasExpression || effect.HasLayerRef {
			hints.add("effect_driven_layer", "high", "effect:"+effect.MatchName)
		}
		if reproducibility == "third_party" {
			hints.add("plugin_dependent", "high", "effect:"+effect.MatchName)
		}
	}
	for _, animator := range facts.TextAnimators {
		key := layerKey{CompName: animator.CompName, LayerName: animator.LayerName}
		addSignal(signals, key, "text_animator:"+animator.PropertyKind)
		portrait.Mechanisms.TextAnimatorKindCounts[animator.PropertyKind]++
		if animator.HasKeyframes || animator.HasExpression {
			hints.add("kinetic_text", "high", "text_animator:"+animator.PropertyKind)
		}
	}
	shapeByLayer := map[layerKey]int{}
	for _, operator := range facts.ShapeOperators {
		key := layerKey{CompName: operator.CompName, LayerName: operator.LayerName}
		addSignal(signals, key, "shape_operator:"+operator.Family)
		portrait.Mechanisms.ShapeFamilyCounts[operator.Family]++
		shapeByLayer[key]++
	}
	for key, count := range shapeByLayer {
		if count >= 2 {
			hints.add("shape_operator_stack", "medium", "layer:"+key.LayerName)
		}
	}
	for _, dep := range facts.Dependencies {
		key := layerKey{CompName: dep.CompName, LayerName: dep.SourceName}
		addSignal(signals, key, "dependency:"+dep.Relation)
		portrait.Graph.EdgeCount++
		portrait.Graph.RelationCounts[dep.Relation]++
		switch dep.Relation {
		case "source":
			if dep.TargetKind == "composition" {
				hints.add("precomp_assembly", "high", "source:"+dep.TargetName)
			}
		case "matte":
			hints.add("matte_composite", "high", "matte:"+dep.TargetName)
		case "effect_param_layer":
			hints.add("controller_rig", "medium", "effect_param_layer:"+dep.TargetName)
		}
		if layer, ok := layerRefs[key]; ok && layer.Role == "controller" {
			hints.add("controller_rig", "medium", "controller:"+layer.Name)
		}
	}

	portrait.SignalLayers = buildSignalLayers(signals, layerRefs)
	portrait.TechniqueHints = hints.list()
	return portrait, nil
}

type layerKey struct {
	CompName  string
	LayerName string
}

func addSignal(signals map[layerKey]map[string]bool, key layerKey, signal string) {
	if key.LayerName == "" || signal == "" {
		return
	}
	if signals[key] == nil {
		signals[key] = map[string]bool{}
	}
	signals[key][signal] = true
}

func buildSignalLayers(signals map[layerKey]map[string]bool, layers map[layerKey]LayerFact) []SignalLayer {
	out := make([]SignalLayer, 0, len(signals))
	for key, signalSet := range signals {
		layer := layers[key]
		values := sortedKeys(signalSet)
		out = append(out, SignalLayer{
			CompName:  key.CompName,
			LayerName: key.LayerName,
			Role:      layer.Role,
			Score:     len(values),
			Signals:   values,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		if out[i].CompName != out[j].CompName {
			return out[i].CompName < out[j].CompName
		}
		return out[i].LayerName < out[j].LayerName
	})
	return out
}

func effectClass(value string) string {
	if value == "" {
		return "unknown"
	}
	return value
}

func reproducibilityClass(value string) string {
	switch value {
	case "native":
		return "native"
	case "cycore", "cycore_bundled", "Cycore(bundled)":
		return "cycore"
	case "third_party", "third-party", "⚠THIRD-PARTY":
		return "third_party"
	default:
		return "unknown"
	}
}

type hintCollector struct {
	seen map[string]TechniqueHint
}

func (h hintCollector) add(id, confidence, signal string) {
	hint := h.seen[id]
	if hint.ID == "" {
		hint = TechniqueHint{ID: id, Confidence: confidence}
	}
	if signal != "" && !stringInSlice(hint.Signals, signal) {
		hint.Signals = append(hint.Signals, signal)
		sort.Strings(hint.Signals)
	}
	h.seen[id] = hint
}

func (h hintCollector) list() []TechniqueHint {
	out := make([]TechniqueHint, 0, len(h.seen))
	for _, hint := range h.seen {
		out = append(out, hint)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func sortedKeys(values map[string]bool) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func stringInSlice(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
