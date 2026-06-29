package recipe

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func maskBezierPath(spec MaskSpec) aep.BezierPath {
	vertices := make([][2]float64, 0, len(spec.Vertices))
	for _, vertex := range spec.Vertices {
		if len(vertex) >= 2 {
			vertices = append(vertices, [2]float64{vertex[0], vertex[1]})
		}
	}
	closed := true
	if spec.Closed != nil {
		closed = *spec.Closed
	}
	return aep.BezierPath{Vertices: vertices, Closed: closed}
}

func maskPathKeys(spec MaskSpec) []aep.MaskPathKey {
	keys := make([]aep.MaskPathKey, 0, len(spec.PathKeyframes))
	closed := true
	if spec.Closed != nil {
		closed = *spec.Closed
	}
	for _, kf := range spec.PathKeyframes {
		path := aep.BezierPath{
			Vertices: make([][2]float64, 0, len(kf.Vertices)),
			Closed:   closed,
		}
		for _, vertex := range kf.Vertices {
			if len(vertex) >= 2 {
				path.Vertices = append(path.Vertices, [2]float64{vertex[0], vertex[1]})
			}
		}
		keys = append(keys, aep.MaskPathKey{Time: kf.Time, Path: path})
	}
	return keys
}

func hasEffects(comp CompSpec) bool {
	for _, layer := range comp.Layers {
		if len(layer.Effects) > 0 {
			return true
		}
	}
	return false
}

func hasMasks(comp CompSpec) bool {
	for _, layer := range comp.Layers {
		if len(layer.Masks) > 0 {
			return true
		}
	}
	return false
}

func hasTransformExpressions(comp CompSpec) bool {
	for _, layer := range comp.Layers {
		if hasLayerTransformExpressions(layer.Transform.Expressions) {
			return true
		}
	}
	return false
}

func hasLayerTransformExpressions(expressions TransformExpressions) bool {
	return expressions.Position != nil ||
		expressions.AnchorPoint != nil ||
		expressions.Scale != nil ||
		expressions.Rotation != nil ||
		expressions.Opacity != nil
}

func materializeMasks(project *aep.Project, compSpec CompSpec) (*aep.Project, error) {
	reopened, err := aep.Reopen(project)
	if err != nil {
		return nil, fmt.Errorf("recipe: reopen for masks: %w", err)
	}
	if len(reopened.Compositions) == 0 {
		return nil, fmt.Errorf("recipe: reopen for masks: no compositions")
	}
	comp := reopened.Compositions[0]
	for i, layerSpec := range compSpec.Layers {
		if len(layerSpec.Masks) == 0 {
			continue
		}
		if i >= len(comp.Layers) {
			return nil, fmt.Errorf("recipe: reopen for masks: layer index %d missing", i)
		}
		layer := comp.Layers[i]
		for mi, maskSpec := range layerSpec.Masks {
			mask, err := aep.AddMask(layer, maskSpec.Name, maskBezierPath(maskSpec))
			if err != nil {
				return nil, fmt.Errorf("recipe: layer %q add mask %d: %w", layerSpec.Name, mi, err)
			}
			if maskSpec.Mode != "" {
				mode, err := maskMode(maskSpec.Mode)
				if err != nil {
					return nil, fmt.Errorf("recipe: layer %q mask %q mode: %w", layerSpec.Name, maskSpec.Name, err)
				}
				if err := mask.SetMode(mode); err != nil {
					return nil, fmt.Errorf("recipe: layer %q mask %q mode: %w", layerSpec.Name, maskSpec.Name, err)
				}
			}
			if maskSpec.Inverted != nil {
				if err := mask.SetInverted(*maskSpec.Inverted); err != nil {
					return nil, fmt.Errorf("recipe: layer %q mask %q inverted: %w", layerSpec.Name, maskSpec.Name, err)
				}
			}
			if maskSpec.Locked != nil {
				if err := mask.SetLocked(*maskSpec.Locked); err != nil {
					return nil, fmt.Errorf("recipe: layer %q mask %q locked: %w", layerSpec.Name, maskSpec.Name, err)
				}
			}
			if len(maskSpec.Color) > 0 {
				if err := mask.SetColor(rgb8Color(maskSpec.Color)); err != nil {
					return nil, fmt.Errorf("recipe: layer %q mask %q color: %w", layerSpec.Name, maskSpec.Name, err)
				}
			}
			if maskSpec.MotionBlur != "" {
				mode, err := maskMotionBlur(maskSpec.MotionBlur)
				if err != nil {
					return nil, fmt.Errorf("recipe: layer %q mask %q motion_blur: %w", layerSpec.Name, maskSpec.Name, err)
				}
				if err := mask.SetMaskMotionBlur(mode); err != nil {
					return nil, fmt.Errorf("recipe: layer %q mask %q motion_blur: %w", layerSpec.Name, maskSpec.Name, err)
				}
			}
			if maskSpec.FeatherFalloff != "" {
				falloff, err := maskFeatherFalloff(maskSpec.FeatherFalloff)
				if err != nil {
					return nil, fmt.Errorf("recipe: layer %q mask %q feather_falloff: %w", layerSpec.Name, maskSpec.Name, err)
				}
				if err := mask.SetFeatherFalloff(falloff); err != nil {
					return nil, fmt.Errorf("recipe: layer %q mask %q feather_falloff: %w", layerSpec.Name, maskSpec.Name, err)
				}
			}
			if maskSpec.Opacity != nil {
				if err := mask.SetOpacity(*maskSpec.Opacity); err != nil {
					return nil, fmt.Errorf("recipe: layer %q mask %q opacity: %w", layerSpec.Name, maskSpec.Name, err)
				}
			}
			if len(maskSpec.Feather) == 2 {
				if err := mask.SetFeather([2]float64{maskSpec.Feather[0], maskSpec.Feather[1]}); err != nil {
					return nil, fmt.Errorf("recipe: layer %q mask %q feather: %w", layerSpec.Name, maskSpec.Name, err)
				}
			}
			if maskSpec.Expansion != nil {
				if err := mask.SetExpansion(*maskSpec.Expansion); err != nil {
					return nil, fmt.Errorf("recipe: layer %q mask %q expansion: %w", layerSpec.Name, maskSpec.Name, err)
				}
			}
			if len(maskSpec.PathKeyframes) > 0 {
				if err := aep.SetMaskPathKeyframes(layer, mask, maskPathKeys(maskSpec)); err != nil {
					return nil, fmt.Errorf("recipe: layer %q mask %q path_keyframes: %w", layerSpec.Name, maskSpec.Name, err)
				}
			}
		}
	}
	return reopened, nil
}

func materializeEffects(project *aep.Project, compSpec CompSpec) (*aep.Project, error) {
	reopened, err := aep.Reopen(project)
	if err != nil {
		return nil, fmt.Errorf("recipe: reopen for effects: %w", err)
	}
	if len(reopened.Compositions) == 0 {
		return nil, fmt.Errorf("recipe: reopen for effects: no compositions")
	}
	comp := reopened.Compositions[0]
	for i, layerSpec := range compSpec.Layers {
		if len(layerSpec.Effects) == 0 {
			continue
		}
		if i >= len(comp.Layers) {
			return nil, fmt.Errorf("recipe: reopen for effects: layer index %d missing", i)
		}
		layer := comp.Layers[i]
		for _, effect := range layerSpec.Effects {
			fx, err := aep.AddEffect(layer, effect.MatchName)
			if err != nil {
				return nil, fmt.Errorf("recipe: layer %q add effect %q: %w", layerSpec.Name, effect.MatchName, err)
			}
			for _, param := range effect.Params {
				value, err := normalizeEffectParamValue(param.Value)
				if err != nil {
					return nil, fmt.Errorf("recipe: layer %q effect %q param %q: %w", layerSpec.Name, effect.MatchName, param.MatchName, err)
				}
				property, err := aep.SetEffectParam(layer, fx, param.MatchName, value)
				if err != nil {
					return nil, fmt.Errorf("recipe: layer %q effect %q param %q: %w", layerSpec.Name, effect.MatchName, param.MatchName, err)
				}
				if param.Expression != nil {
					if err := applyPropertyExpression(property, *param.Expression); err != nil {
						return nil, fmt.Errorf("recipe: layer %q effect %q param %q expression: %w", layerSpec.Name, effect.MatchName, param.MatchName, err)
					}
				}
			}
		}
	}
	return reopened, nil
}

func materializeTransformExpressions(project *aep.Project, compSpec CompSpec) (*aep.Project, error) {
	reopened, err := aep.Reopen(project)
	if err != nil {
		return nil, fmt.Errorf("recipe: reopen for expressions: %w", err)
	}
	if len(reopened.Compositions) == 0 {
		return nil, fmt.Errorf("recipe: reopen for expressions: no compositions")
	}
	comp := reopened.Compositions[0]
	for i, layerSpec := range compSpec.Layers {
		if !hasLayerTransformExpressions(layerSpec.Transform.Expressions) {
			continue
		}
		if i >= len(comp.Layers) {
			return nil, fmt.Errorf("recipe: reopen for expressions: layer index %d missing", i)
		}
		if err := applyTransformExpressions(comp.Layers[i], layerSpec.Transform.Expressions); err != nil {
			return nil, fmt.Errorf("recipe: layer %q transform.expressions: %w", layerSpec.Name, err)
		}
	}
	return reopened, nil
}

func applyTransformExpressions(layer *aep.Layer, expressions TransformExpressions) error {
	if err := applyTransformExpression(layer, "position", expressions.Position); err != nil {
		return err
	}
	if err := applyTransformExpression(layer, "anchor_point", expressions.AnchorPoint); err != nil {
		return err
	}
	if err := applyTransformExpression(layer, "scale", expressions.Scale); err != nil {
		return err
	}
	if err := applyTransformExpression(layer, "rotation", expressions.Rotation); err != nil {
		return err
	}
	if err := applyTransformExpression(layer, "opacity", expressions.Opacity); err != nil {
		return err
	}
	return nil
}

func applyPropertyExpression(property *aep.Property, expression ExpressionSpec) error {
	if property == nil {
		return fmt.Errorf("property missing")
	}
	if err := property.SetExpression(expression.Source); err != nil {
		return fmt.Errorf("source: %w", err)
	}
	if expression.Enabled != nil {
		if err := property.SetExpressionEnabled(*expression.Enabled); err != nil {
			return fmt.Errorf("enabled: %w", err)
		}
	}
	return nil
}

func applyTransformExpression(layer *aep.Layer, name string, expression *ExpressionSpec) error {
	if expression == nil {
		return nil
	}
	property, err := transformExpressionProperty(layer, name)
	if err != nil {
		return err
	}
	if err := applyPropertyExpression(property, *expression); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

func transformExpressionProperty(layer *aep.Layer, name string) (*aep.Property, error) {
	switch name {
	case "position":
		return layer.Position(), nil
	case "anchor_point":
		return layer.AnchorPoint(), nil
	case "scale":
		return layer.Scale(), nil
	case "rotation":
		return layer.Rotation(), nil
	case "opacity":
		return layer.Opacity(), nil
	default:
		return nil, fmt.Errorf("unsupported transform expression property %q", name)
	}
}

func applyCompItemSettings(project *aep.Project, compSpec CompSpec) (*aep.Project, error) {
	if compSpec.Label == nil && compSpec.Comment == "" {
		return project, nil
	}
	reopened, err := aep.Reopen(project)
	if err != nil {
		return nil, fmt.Errorf("reopen: %w", err)
	}
	comp := findProjectComp(reopened, compSpec.Name)
	if comp == nil {
		return nil, fmt.Errorf("comp %q not found after reopen", compSpec.Name)
	}
	if compSpec.Label != nil {
		if err := comp.SetLabel(uint8(*compSpec.Label)); err != nil {
			return nil, fmt.Errorf("label: %w", err)
		}
	}
	if compSpec.Comment != "" {
		if err := comp.SetComment(compSpec.Comment); err != nil {
			return nil, fmt.Errorf("comment: %w", err)
		}
	}
	return reopened, nil
}

func findProjectComp(project *aep.Project, name string) *aep.Composition {
	for _, comp := range project.Compositions {
		if comp.Name == name {
			return comp
		}
	}
	return nil
}

func applyCompMotionBlur(comp *aep.Composition, spec *CompMotionBlurSpec) error {
	if spec == nil {
		return nil
	}
	if spec.Enabled != nil {
		if err := comp.SetCompMotionBlur(*spec.Enabled); err != nil {
			return err
		}
	}
	if spec.ShutterAngle != nil {
		if err := comp.SetShutterAngle(uint16(*spec.ShutterAngle)); err != nil {
			return err
		}
	}
	if spec.ShutterPhase != nil {
		if err := comp.SetShutterPhase(int32(*spec.ShutterPhase)); err != nil {
			return err
		}
	}
	if spec.AdaptiveSampleLimit != nil {
		if err := comp.SetMotionBlurAdaptiveSampleLimit(int32(*spec.AdaptiveSampleLimit)); err != nil {
			return err
		}
	}
	if spec.SamplesPerFrame != nil {
		if err := comp.SetMotionBlurSamplesPerFrame(int32(*spec.SamplesPerFrame)); err != nil {
			return err
		}
	}
	return nil
}

func normalizeEffectParamValue(value any) (any, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	case []float64:
		if len(v) == 0 {
			return nil, fmt.Errorf("empty numeric array")
		}
		return v, nil
	case []any:
		if len(v) == 0 {
			return nil, fmt.Errorf("empty numeric array")
		}
		out := make([]float64, 0, len(v))
		for i, item := range v {
			n, ok := item.(float64)
			if !ok {
				return nil, fmt.Errorf("array item %d is %T, want number", i, item)
			}
			out = append(out, n)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported value type %T", value)
	}
}
