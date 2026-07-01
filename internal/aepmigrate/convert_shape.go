package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeRectGraphicShapeLayer(comp *aep.Composition, source profile.Layer) (*aep.Layer, error) {
	shapeLayer, err := aep.NewShapeLayer(comp, source.Name)
	if err != nil {
		return nil, err
	}
	if err := materializeShapePrimitive(shapeLayer, source.Shapes[0]); err != nil {
		return nil, err
	}
	if hasProperty(source, "ADBE Vector RoundCorner Radius") {
		if err := materializeShapeRoundCorners(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasProperty(source, "ADBE Vector Offset Amount") {
		if err := materializeShapeOffsetPaths(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasProperty(source, "ADBE Vector Trim Start") ||
		hasProperty(source, "ADBE Vector Trim End") ||
		hasProperty(source, "ADBE Vector Trim Offset") {
		if err := materializeShapeTrim(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasProperty(source, "ADBE Vector Zigzag Size") ||
		hasProperty(source, "ADBE Vector Zigzag Detail") ||
		hasProperty(source, "ADBE Vector Zigzag Points") {
		if err := materializeShapeZigZag(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasProperty(source, "ADBE Vector PuckerBloat Amount") {
		if err := materializeShapePuckerBloat(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasProperty(source, "ADBE Vector Twist Angle") ||
		hasProperty(source, "ADBE Vector Twist Center") {
		if err := materializeShapeTwist(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasWigglePathsFilter(source) {
		if err := materializeShapeWigglePaths(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasWiggleTransformFilter(source) {
		if err := materializeShapeWiggleTransform(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasRepeaterFilter(source) {
		if err := materializeShapeRepeater(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasProperty(source, "ADBE Vector Merge Type") {
		if err := materializeShapeMergePaths(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasGradientFillGraphic(source) {
		if err := materializeShapeGradientFill(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasGradientStrokeGraphic(source) {
		if err := materializeShapeGradientStroke(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasProperty(source, "ADBE Vector Fill Color") {
		if err := materializeShapeFill(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	if hasProperty(source, "ADBE Vector Stroke Color") {
		if err := materializeShapeStroke(shapeLayer, source); err != nil {
			return nil, err
		}
	}
	return shapeLayer.Layer, nil
}

func materializeShapePrimitive(shapeLayer *aep.ShapeLayer, shape profile.Shape) error {
	switch shape.Kind {
	case "rect":
		return materializeRectPrimitive(shapeLayer, shape)
	case "ellipse":
		return materializeEllipsePrimitive(shapeLayer, shape)
	case "star":
		return materializeStarPrimitive(shapeLayer, shape)
	default:
		return fmt.Errorf("unsupported shape primitive %q", shape.Kind)
	}
}

func materializeRectPrimitive(shapeLayer *aep.ShapeLayer, shape profile.Shape) error {
	rect, err := shapeLayer.RootGroup().AddRect()
	if err != nil {
		return err
	}
	if value, ok := propertyVector(shape.Properties, "ADBE Vector Rect Size", 2); ok {
		if err := rect.SetSize([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(shape.Properties, "ADBE Vector Rect Position", 2); ok {
		if err := rect.SetPosition([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Rect Roundness"); ok {
		if err := rect.SetRoundness(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Shape Direction"); ok {
		if err := rect.SetDirection(aep.ShapeDirection(int(value))); err != nil {
			return err
		}
	}
	return nil
}

func materializeEllipsePrimitive(shapeLayer *aep.ShapeLayer, shape profile.Shape) error {
	ellipse, err := shapeLayer.RootGroup().AddEllipse()
	if err != nil {
		return err
	}
	if value, ok := propertyVector(shape.Properties, "ADBE Vector Ellipse Size", 2); ok {
		if err := ellipse.SetSize([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(shape.Properties, "ADBE Vector Ellipse Position", 2); ok {
		if err := ellipse.SetPosition([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Shape Direction"); ok {
		if err := ellipse.SetDirection(aep.ShapeDirection(int(value))); err != nil {
			return err
		}
	}
	return nil
}

func materializeStarPrimitive(shapeLayer *aep.ShapeLayer, shape profile.Shape) error {
	star, err := shapeLayer.RootGroup().AddStar()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Star Type"); ok {
		if err := star.SetStarType(aep.StarType(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Star Points"); ok {
		if err := star.SetPoints(value); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(shape.Properties, "ADBE Vector Star Position", 2); ok {
		if err := star.SetPosition([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Star Rotation"); ok {
		if err := star.SetRotation(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Star Inner Radius"); ok {
		if err := star.SetInnerRadius(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Star Outer Radius"); ok {
		if err := star.SetOuterRadius(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Star Inner Roundess"); ok {
		if err := star.SetInnerRoundness(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(shape.Properties, "ADBE Vector Star Outer Roundess"); ok {
		if err := star.SetOuterRoundness(value); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeRoundCorners(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	roundCorners, err := shapeLayer.RootGroup().AddRoundCorners()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector RoundCorner Radius"); ok {
		if err := roundCorners.SetRadius(value); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeOffsetPaths(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	offsetPaths, err := shapeLayer.RootGroup().AddOffsetPaths()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Offset Amount"); ok {
		if err := offsetPaths.SetAmount(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Offset Line Join"); ok {
		if err := offsetPaths.SetLineJoin(aep.StrokeLineJoin(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Offset Miter Limit"); ok {
		if err := offsetPaths.SetMiterLimit(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Offset Copies"); ok {
		if err := offsetPaths.SetCopies(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Offset Copy Offset"); ok {
		if err := offsetPaths.SetCopyOffset(value); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeTrim(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	trim, err := shapeLayer.RootGroup().AddTrim()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Trim Start"); ok {
		if err := trim.SetStart(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Trim End"); ok {
		if err := trim.SetEnd(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Trim Offset"); ok {
		if err := trim.SetOffset(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Trim Type"); ok {
		if err := trim.SetType(aep.TrimType(int(value))); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeZigZag(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	zigZag, err := shapeLayer.RootGroup().AddZigZag()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Zigzag Size"); ok {
		if err := zigZag.SetSize(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Zigzag Detail"); ok {
		if err := zigZag.SetDetail(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Zigzag Points"); ok {
		if err := zigZag.SetPoints(aep.ZigZagPoints(int(value))); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapePuckerBloat(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	puckerBloat, err := shapeLayer.RootGroup().AddPuckerBloat()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector PuckerBloat Amount"); ok {
		if err := puckerBloat.SetAmount(value); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeTwist(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	twist, err := shapeLayer.RootGroup().AddTwist()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Twist Angle"); ok {
		if err := twist.SetAngle(value); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Twist Center", 2); ok {
		if err := twist.SetCenter([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeWigglePaths(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	wigglePaths, err := shapeLayer.RootGroup().AddWigglePaths()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Roughen Size"); ok {
		if err := wigglePaths.SetSize(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Roughen Detail"); ok {
		if err := wigglePaths.SetDetail(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Temporal Freq"); ok {
		if err := wigglePaths.SetWigglesPerSecond(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Random Seed"); ok {
		if err := wigglePaths.SetRandomSeed(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Roughen Points"); ok {
		if err := wigglePaths.SetPoints(aep.RoughenPoints(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Correlation"); ok {
		if err := wigglePaths.SetCorrelation(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Temporal Phase"); ok {
		if err := wigglePaths.SetTemporalPhase(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Spatial Phase"); ok {
		if err := wigglePaths.SetSpatialPhase(value); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeWiggleTransform(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	wiggleTransform, err := shapeLayer.RootGroup().AddWiggleTransform()
	if err != nil {
		return err
	}
	transform := wiggleTransform.Transform()
	if value, ok := propertyVector(source.Properties, "ADBE Vector Wiggler Anchor", 2); ok {
		if err := transform.SetAnchor([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Wiggler Position", 2); ok {
		if err := transform.SetPosition([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Wiggler Scale", 2); ok {
		if err := transform.SetScale([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Wiggler Rotation"); ok {
		if err := transform.SetRotation(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Xform Temporal Freq"); ok {
		if err := wiggleTransform.SetWigglesPerSecond(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Random Seed"); ok {
		if err := wiggleTransform.SetRandomSeed(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Correlation"); ok {
		if err := wiggleTransform.SetCorrelation(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Temporal Phase"); ok {
		if err := wiggleTransform.SetTemporalPhase(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Spatial Phase"); ok {
		if err := wiggleTransform.SetSpatialPhase(value); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeRepeater(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	repeater, err := shapeLayer.RootGroup().AddRepeater()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Repeater Copies"); ok {
		if err := repeater.SetCopies(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Repeater Offset"); ok {
		if err := repeater.SetOffset(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Repeater Order"); ok {
		if err := repeater.SetOrder(aep.RepeaterOrder(int(value))); err != nil {
			return err
		}
	}
	transform := repeater.Transform()
	if value, ok := propertyVector(source.Properties, "ADBE Vector Repeater Anchor", 2); ok {
		if err := transform.SetAnchor([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Repeater Position", 2); ok {
		if err := transform.SetPosition([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Repeater Scale", 2); ok {
		if err := transform.SetScale([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Repeater Rotation"); ok {
		if err := transform.SetRotation(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Repeater Opacity 1"); ok {
		if err := transform.SetStartOpacity(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Repeater Opacity 2"); ok {
		if err := transform.SetEndOpacity(value); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeMergePaths(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	mergePaths, err := shapeLayer.RootGroup().AddMergePaths()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Merge Type"); ok {
		if err := mergePaths.SetType(aep.MergeType(int(value))); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeGradientFill(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	fill, err := shapeLayer.RootGroup().AddGradientFill()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Grad Type"); ok {
		if err := fill.SetGradientType(aep.GradientType(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Grad Start Pt", 2); ok {
		if err := fill.SetStartPoint([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Grad End Pt", 2); ok {
		if err := fill.SetEndPoint([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Grad HiLite Length"); ok {
		if err := fill.SetHighlightLength(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Grad HiLite Angle"); ok {
		if err := fill.SetHighlightAngle(value); err != nil {
			return err
		}
	}
	if gradient, ok := propertyGradient(source.Properties, "ADBE Vector Grad Colors"); ok {
		if len(gradient.ColorStops) > 0 {
			if err := fill.SetColorStops(gradient.ColorStops); err != nil {
				return err
			}
		}
		if len(gradient.AlphaStops) > 0 {
			if err := fill.SetAlphaStops(gradient.AlphaStops); err != nil {
				return err
			}
		}
	}
	return nil
}

func materializeShapeGradientStroke(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	stroke, err := shapeLayer.RootGroup().AddGradientStroke()
	if err != nil {
		return err
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Grad Type"); ok {
		if err := stroke.SetGradientType(aep.GradientType(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Grad Start Pt", 2); ok {
		if err := stroke.SetStartPoint([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Grad End Pt", 2); ok {
		if err := stroke.SetEndPoint([2]float64{value[0], value[1]}); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Grad HiLite Length"); ok {
		if err := stroke.SetHighlightLength(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Grad HiLite Angle"); ok {
		if err := stroke.SetHighlightAngle(value); err != nil {
			return err
		}
	}
	if gradient, ok := propertyGradient(source.Properties, "ADBE Vector Grad Colors"); ok {
		if len(gradient.ColorStops) > 0 {
			if err := stroke.SetColorStops(gradient.ColorStops); err != nil {
				return err
			}
		}
		if len(gradient.AlphaStops) > 0 {
			if err := stroke.SetAlphaStops(gradient.AlphaStops); err != nil {
				return err
			}
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Width"); ok {
		if err := stroke.SetStrokeWidth(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Line Cap"); ok {
		if err := stroke.SetLineCap(aep.StrokeLineCap(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Line Join"); ok {
		if err := stroke.SetLineJoin(aep.StrokeLineJoin(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Miter Limit"); ok {
		if err := stroke.SetMiterLimit(value); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeFill(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	fill, err := shapeLayer.RootGroup().AddFill()
	if err != nil {
		return err
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Fill Color", 4); ok {
		if err := fill.SetColor(profileARGBToRGBA(value)); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Fill Opacity"); ok {
		if err := fill.SetOpacity(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Blend Mode"); ok {
		if err := fill.SetBlendMode(aep.ShapeBlendMode(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Composite Order"); ok {
		if err := fill.SetCompositeOrder(aep.ShapeCompositeOrder(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Fill Rule"); ok {
		if err := fill.SetFillRule(aep.FillRule(int(value))); err != nil {
			return err
		}
	}
	return nil
}

func materializeShapeStroke(shapeLayer *aep.ShapeLayer, source profile.Layer) error {
	stroke, err := shapeLayer.RootGroup().AddStroke()
	if err != nil {
		return err
	}
	if value, ok := propertyVector(source.Properties, "ADBE Vector Stroke Color", 4); ok {
		if err := stroke.SetColor(profileARGBToRGBA(value)); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Opacity"); ok {
		if err := stroke.SetOpacity(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Width"); ok {
		if err := stroke.SetWidth(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Line Cap"); ok {
		if err := stroke.SetLineCap(aep.StrokeLineCap(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Line Join"); ok {
		if err := stroke.SetLineJoin(aep.StrokeLineJoin(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Miter Limit"); ok {
		if err := stroke.SetMiterLimit(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Blend Mode"); ok {
		if err := stroke.SetBlendMode(aep.ShapeBlendMode(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Composite Order"); ok {
		if err := stroke.SetCompositeOrder(aep.ShapeCompositeOrder(int(value))); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Dash 1"); ok {
		if err := stroke.Dashes().SetDash(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Stroke Gap 1"); ok {
		if err := stroke.Dashes().SetGap(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper Start Length"); ok {
		if err := stroke.Taper().SetStartLength(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper End Length"); ok {
		if err := stroke.Taper().SetEndLength(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper Start Width"); ok {
		if err := stroke.Taper().SetStartWidth(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper End Width"); ok {
		if err := stroke.Taper().SetEndWidth(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper Start Ease"); ok {
		if err := stroke.Taper().SetStartEase(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper End Ease"); ok {
		if err := stroke.Taper().SetEndEase(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper Wave Amount"); ok {
		if err := stroke.Wave().SetAmount(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper Wavelength"); ok {
		if err := stroke.Wave().SetWavelength(value); err != nil {
			return err
		}
	}
	if value, ok := propertyFloat(source.Properties, "ADBE Vector Taper Wave Phase"); ok {
		if err := stroke.Wave().SetPhase(value); err != nil {
			return err
		}
	}
	return nil
}
