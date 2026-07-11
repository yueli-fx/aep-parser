package recipe

import (
	"github.com/yueli-fx/aep-parser/internal/aep"
)

func applyTransform(layer *aep.Layer, spec Transform) error {
	t := aep.NewLayerTransform()
	if err := applyTransformStreams(t, spec); err != nil {
		return err
	}
	return aep.SetLayerTransform(layer, t)
}

func applyShapeLayerTransform(layer *aep.ShapeLayer, spec Transform) error {
	return applyTransformStreams(layer.Transform(), spec)
}

func applyTransformStreams(t *aep.LayerTransform, spec Transform) error {
	if len(spec.AnchorPoint) == 2 {
		if err := t.AnchorPoint().SetStaticValue([2]float64{spec.AnchorPoint[0], spec.AnchorPoint[1]}); err != nil {
			return err
		}
	}
	if len(spec.Position) == 2 {
		if err := t.Position().SetStaticValue([2]float64{spec.Position[0], spec.Position[1]}); err != nil {
			return err
		}
	}
	if len(spec.Scale) == 2 {
		if err := t.Scale().SetStaticValue([2]float64{spec.Scale[0], spec.Scale[1]}); err != nil {
			return err
		}
	}
	if spec.Rotation != nil {
		if err := t.Rotation().SetStaticValue(*spec.Rotation); err != nil {
			return err
		}
	}
	if spec.Opacity != nil {
		if err := t.Opacity().SetStaticValue(*spec.Opacity); err != nil {
			return err
		}
	}
	for _, kf := range spec.PositionKeyframes {
		value := [2]float64{kf.Value[0], kf.Value[1]}
		if hasKeyframeEase(kf.InEase, kf.OutEase) {
			if err := t.Position().AddKeyframeWithEase(kf.Time, value, temporalEase(kf.InEase), temporalEase(kf.OutEase)); err != nil {
				return err
			}
		} else if err := t.Position().AddKeyframeLinear(kf.Time, value); err != nil {
			return err
		}
	}
	for _, kf := range spec.AnchorPointKeyframes {
		value := [2]float64{kf.Value[0], kf.Value[1]}
		if hasKeyframeEase(kf.InEase, kf.OutEase) {
			if err := t.AnchorPoint().AddKeyframeWithEase(kf.Time, value, temporalEase(kf.InEase), temporalEase(kf.OutEase)); err != nil {
				return err
			}
		} else if err := t.AnchorPoint().AddKeyframeLinear(kf.Time, value); err != nil {
			return err
		}
	}
	for _, kf := range spec.ScaleKeyframes {
		value := [2]float64{kf.Value[0], kf.Value[1]}
		if hasKeyframeEase(kf.InEase, kf.OutEase) {
			if err := t.Scale().AddKeyframeWithEase(kf.Time, value, temporalEase(kf.InEase), temporalEase(kf.OutEase)); err != nil {
				return err
			}
		} else if err := t.Scale().AddKeyframeLinear(kf.Time, value); err != nil {
			return err
		}
	}
	for _, kf := range spec.RotationKeyframes {
		if hasKeyframeEase(kf.InEase, kf.OutEase) {
			if err := t.Rotation().AddKeyframeWithEase(kf.Time, kf.Value, temporalEase(kf.InEase), temporalEase(kf.OutEase)); err != nil {
				return err
			}
		} else if err := t.Rotation().AddKeyframeLinear(kf.Time, kf.Value); err != nil {
			return err
		}
	}
	for _, kf := range spec.OpacityKeyframes {
		if hasKeyframeEase(kf.InEase, kf.OutEase) {
			if err := t.Opacity().AddKeyframeWithEase(kf.Time, kf.Value, temporalEase(kf.InEase), temporalEase(kf.OutEase)); err != nil {
				return err
			}
		} else if err := t.Opacity().AddKeyframeLinear(kf.Time, kf.Value); err != nil {
			return err
		}
	}
	return nil
}

func hasKeyframeEase(in, out *TemporalEase) bool {
	return in != nil || out != nil
}

func temporalEase(ease *TemporalEase) aep.TemporalEase {
	if ease == nil {
		return aep.TemporalEase{}
	}
	return aep.TemporalEase{
		Speed:     ease.Speed,
		Influence: ease.Influence,
	}
}

func toUnitColor(v float64) float64 {
	if v > 1 {
		return v / 255
	}
	return v
}

func rgbaColor(values []float64) [4]float64 {
	alpha := 1.0
	if len(values) >= 4 {
		alpha = toUnitColor(values[3])
	}
	return [4]float64{
		toUnitColor(values[0]),
		toUnitColor(values[1]),
		toUnitColor(values[2]),
		alpha,
	}
}

func rgb8Color(values []float64) [3]uint8 {
	return [3]uint8{uint8(values[0]), uint8(values[1]), uint8(values[2])}
}
