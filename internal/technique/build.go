package technique

import (
	"fmt"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/profile"
)

func Build(prof *profile.Profile) (*FactSet, error) {
	if prof == nil {
		return nil, fmt.Errorf("technique: nil profile")
	}
	facts := &FactSet{
		SchemaVersion: SchemaVersion,
		SourcePath:    prof.Meta.Path,
		Summary: Summary{
			CompCount: len(prof.Comps),
		},
	}
	mainIndex := mainCompIndex(prof.Comps)
	for ci := range prof.Comps {
		comp := prof.Comps[ci]
		main := ci == mainIndex
		if main {
			facts.Summary.MainCompID = comp.ID
			facts.Summary.MainCompName = comp.Name
		}
		facts.Comps = append(facts.Comps, CompFact{
			ID:            comp.ID,
			Name:          comp.Name,
			Width:         comp.Width,
			Height:        comp.Height,
			FrameRate:     comp.FrameRate,
			Duration:      comp.Duration,
			LayerCount:    len(comp.Layers),
			MainCandidate: main,
			Path:          comp.Path,
			Evidence:      comp.Evidence,
		})
		for li := range comp.Layers {
			layer := comp.Layers[li]
			facts.Summary.LayerCount++
			role, confidence := layerRole(layer)
			if role == "text" {
				facts.Summary.TextLayerCount++
			}
			if role == "shape" {
				facts.Summary.ShapeLayerCount++
			}
			facts.Layers = append(facts.Layers, LayerFact{
				CompName:   comp.Name,
				ID:         layer.ID,
				Index:      layer.Index,
				Name:       layer.Name,
				Type:       layer.Type,
				Role:       role,
				Confidence: confidence,
				Path:       layer.Path,
				Evidence:   layer.Evidence,
			})
			addLayerDependencies(facts, comp, layer)
			addTextAnimatorFacts(facts, comp, layer)
			addShapeOperatorFacts(facts, comp, layer)
			for ei := range layer.Effects {
				effect := layer.Effects[ei]
				facts.Summary.EffectCount++
				facts.Effects = append(facts.Effects, effectFact(comp, layer, effect))
				addEffectDependencies(facts, comp, layer, effect)
				for _, unknown := range effect.UnknownParams {
					facts.Unknowns = append(facts.Unknowns, UnknownFact{
						CompName:   comp.Name,
						LayerName:  layer.Name,
						EffectName: effect.MatchName,
						Path:       unknown.Path,
						Reason:     unknown.Reason,
						Evidence:   unknown.Evidence,
					})
				}
			}
		}
	}
	for _, unknown := range prof.Unknowns {
		facts.Unknowns = append(facts.Unknowns, UnknownFact{
			Path:     unknown.Path,
			Reason:   unknown.Reason,
			Evidence: unknown.Evidence,
		})
	}
	facts.Summary.UnknownCount = len(facts.Unknowns)
	return facts, nil
}

func mainCompIndex(comps []profile.Composition) int {
	if len(comps) == 0 {
		return -1
	}
	best := 0
	for i := 1; i < len(comps); i++ {
		if len(comps[i].Layers) > len(comps[best].Layers) {
			best = i
		}
	}
	return best
}

func layerRole(layer profile.Layer) (string, string) {
	switch {
	case layer.Type == "text":
		return "text", "high"
	case layer.Type == "shape":
		return "shape", "high"
	case layer.Type == "camera":
		return "camera", "high"
	case layer.Type == "light":
		return "light", "high"
	case layer.Type == "null" || layer.Flags.IsNull:
		return "controller", "medium"
	case layer.Flags.IsAdjustment:
		return "adjustment", "high"
	case layer.SourceRef != nil && layer.SourceRef.Kind == "composition":
		return "precomp", "high"
	case layer.SourceRef != nil && layer.SourceRef.Kind == "footage":
		return "solid", "high"
	case layer.Type != "":
		return layer.Type, "medium"
	default:
		return "unknown", "low"
	}
}

func addTextAnimatorFacts(facts *FactSet, comp profile.Composition, layer profile.Layer) {
	if layer.Type != "text" {
		return
	}
	for _, prop := range layer.Properties {
		if !strings.HasPrefix(prop.MatchName, "ADBE Text ") {
			continue
		}
		facts.TextAnimators = append(facts.TextAnimators, TextAnimatorFact{
			CompName:       comp.Name,
			LayerName:      layer.Name,
			PropertyKind:   textPropertyKind(prop.MatchName),
			PropertyName:   prop.Name,
			MatchName:      prop.MatchName,
			HasStaticValue: prop.StaticValue != nil,
			HasExpression:  prop.Expression != "",
			HasKeyframes:   len(prop.Keyframes) > 0,
			Path:           prop.Path,
			Evidence:       prop.Evidence,
		})
	}
}

func textPropertyKind(matchName string) string {
	switch matchName {
	case "ADBE Text Opacity":
		return "opacity"
	case "ADBE Text Position 3D":
		return "position"
	case "ADBE Text Scale 3D":
		return "scale"
	case "ADBE Text Rotation":
		return "rotation"
	case "ADBE Text Fill Color":
		return "fill_color"
	case "ADBE Text Stroke Color":
		return "stroke_color"
	case "ADBE Text Tracking Amount":
		return "tracking"
	case "ADBE Text Character Offset":
		return "character_offset"
	case "ADBE Text Fill Opacity":
		return "fill_opacity"
	case "ADBE Text Stroke Opacity":
		return "stroke_opacity"
	case "ADBE Text Stroke Width":
		return "stroke_width"
	case "ADBE Text Skew":
		return "skew"
	case "ADBE Text Rotation X":
		return "rotation_x"
	case "ADBE Text Rotation Y":
		return "rotation_y"
	case "ADBE Text Percent Offset":
		return "range_offset"
	default:
		return slug(strings.TrimPrefix(matchName, "ADBE Text "))
	}
}

func addShapeOperatorFacts(facts *FactSet, comp profile.Composition, layer profile.Layer) {
	if layer.Type != "shape" && len(layer.Shapes) == 0 {
		return
	}
	for _, shape := range layer.Shapes {
		family := shapeFamily(shape.Kind, "")
		facts.ShapeOperators = append(facts.ShapeOperators, ShapeOperatorFact{
			CompName:  comp.Name,
			LayerName: layer.Name,
			Family:    family,
			Source:    "shape",
			MatchName: shape.Kind,
			Path:      shape.Path,
			Evidence:  shape.Evidence,
		})
		for _, prop := range shape.Properties {
			family := shapeFamily("", prop.MatchName)
			facts.ShapeOperators = append(facts.ShapeOperators, ShapeOperatorFact{
				CompName:  comp.Name,
				LayerName: layer.Name,
				Family:    family,
				Source:    "property",
				MatchName: prop.MatchName,
				Path:      prop.Path,
				Evidence:  prop.Evidence,
			})
		}
	}
	for _, prop := range layer.Properties {
		if !strings.HasPrefix(prop.MatchName, "ADBE Vector ") {
			continue
		}
		facts.ShapeOperators = append(facts.ShapeOperators, ShapeOperatorFact{
			CompName:  comp.Name,
			LayerName: layer.Name,
			Family:    shapeFamily("", prop.MatchName),
			Source:    "property",
			MatchName: prop.MatchName,
			Path:      prop.Path,
			Evidence:  prop.Evidence,
		})
	}
}

func shapeFamily(kind, matchName string) string {
	switch strings.ToLower(kind) {
	case "path":
		return "path"
	case "rect", "rectangle":
		return "rectangle"
	case "ellipse":
		return "ellipse"
	case "star", "polystar":
		return "star"
	}
	switch {
	case strings.Contains(matchName, "Graphic - Stroke"):
		return "stroke"
	case strings.Contains(matchName, "Graphic - Fill"):
		return "fill"
	case strings.Contains(matchName, "Graphic - G-Fill"):
		return "gradient_fill"
	case strings.Contains(matchName, "Graphic - G-Stroke"):
		return "gradient_stroke"
	case strings.Contains(matchName, "Filter - Trim"):
		return "trim"
	case strings.Contains(matchName, "Filter - Repeater"):
		return "repeater"
	case strings.Contains(matchName, "Filter - Round"):
		return "round_corners"
	case strings.Contains(matchName, "Filter - Offset"):
		return "offset_paths"
	case strings.Contains(matchName, "Filter - Zigzag"):
		return "zigzag"
	case strings.Contains(matchName, "Filter - PB"):
		return "pucker_bloat"
	case strings.Contains(matchName, "Filter - Twist"):
		return "twist"
	case strings.Contains(matchName, "Filter - Roughen"):
		return "wiggle_paths"
	case strings.Contains(matchName, "Filter - Transform"):
		return "wiggle_transform"
	case strings.Contains(matchName, "Filter - Merge"):
		return "merge_paths"
	default:
		return slug(matchName)
	}
}

func slug(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	replacer := strings.NewReplacer(" ", "_", "-", "_", "/", "_", ".", "_")
	s = replacer.Replace(s)
	for strings.Contains(s, "__") {
		s = strings.ReplaceAll(s, "__", "_")
	}
	return strings.Trim(s, "_")
}

func addLayerDependencies(facts *FactSet, comp profile.Composition, layer profile.Layer) {
	if layer.SourceRef != nil {
		facts.Dependencies = append(facts.Dependencies, DependencyFact{
			CompName:   comp.Name,
			SourceName: layer.Name,
			SourceID:   layer.ID,
			Relation:   "source",
			TargetName: layer.SourceRef.Name,
			TargetID:   layer.SourceRef.ID,
			TargetKind: layer.SourceRef.Kind,
			Path:       layer.Path,
			Evidence:   layer.Evidence,
		})
	}
	addLayerRefDependency(facts, comp, layer, "parent", layer.ParentRef, "")
	addLayerRefDependency(facts, comp, layer, "matte", layer.MatteRef, "")
	addLayerRefDependency(facts, comp, layer, "light_source", layer.LightSourceRef, "")
}

func addLayerRefDependency(facts *FactSet, comp profile.Composition, layer profile.Layer, relation string, ref *profile.LayerRef, property string) {
	if ref == nil {
		return
	}
	facts.Dependencies = append(facts.Dependencies, DependencyFact{
		CompName:   comp.Name,
		SourceName: layer.Name,
		SourceID:   layer.ID,
		Relation:   relation,
		TargetName: ref.Name,
		TargetID:   ref.ID,
		Property:   property,
		Path:       layer.Path,
		Evidence:   layer.Evidence,
	})
}

func effectFact(comp profile.Composition, layer profile.Layer, effect profile.Effect) EffectFact {
	fact := EffectFact{
		CompName:          comp.Name,
		LayerName:         layer.Name,
		MatchName:         effect.MatchName,
		DisplayName:       effect.DisplayName,
		DependencyClass:   effect.DependencyClass,
		Occurrence:        effect.Occurrence,
		TunedParamCount:   len(effect.TunedParams),
		UnknownParamCount: len(effect.UnknownParams),
		Path:              effect.Path,
		Evidence:          effect.Evidence,
	}
	for _, param := range effect.Params {
		if param.Changed {
			fact.ChangedParamCount++
		}
		if param.Expression != "" {
			fact.HasExpression = true
		}
		if len(param.Keyframes) > 0 {
			fact.HasKeyframes = true
		}
		if param.LayerRef != nil {
			fact.HasLayerRef = true
		}
	}
	return fact
}

func addEffectDependencies(facts *FactSet, comp profile.Composition, layer profile.Layer, effect profile.Effect) {
	for _, param := range effect.Params {
		if param.LayerRef == nil {
			continue
		}
		facts.Dependencies = append(facts.Dependencies, DependencyFact{
			CompName:   comp.Name,
			SourceName: layer.Name,
			SourceID:   layer.ID,
			Relation:   "effect_param_layer",
			TargetName: param.LayerRef.Name,
			TargetID:   param.LayerRef.ID,
			Property:   param.MatchName,
			Path:       param.Path,
			Evidence:   param.Evidence,
		})
	}
}
