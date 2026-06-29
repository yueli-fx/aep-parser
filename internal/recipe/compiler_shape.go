package recipe

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func compileShape(group *aep.VectorGroup, shape ShapeSpec) error {
	switch shape.Kind {
	case "rect":
		rect, err := group.AddRect()
		if err != nil {
			return err
		}
		if len(shape.Size) == 2 {
			if err := rect.SetSize([2]float64{shape.Size[0], shape.Size[1]}); err != nil {
				return err
			}
		}
		if len(shape.Position) == 2 {
			if err := rect.SetPosition([2]float64{shape.Position[0], shape.Position[1]}); err != nil {
				return err
			}
		}
		if shape.Roundness != nil {
			if err := rect.SetRoundness(*shape.Roundness); err != nil {
				return err
			}
		}
	case "ellipse":
		ellipse, err := group.AddEllipse()
		if err != nil {
			return err
		}
		if len(shape.Size) == 2 {
			if err := ellipse.SetSize([2]float64{shape.Size[0], shape.Size[1]}); err != nil {
				return err
			}
		}
		if len(shape.Position) == 2 {
			if err := ellipse.SetPosition([2]float64{shape.Position[0], shape.Position[1]}); err != nil {
				return err
			}
		}
	case "star", "polygon":
		star, err := group.AddStar()
		if err != nil {
			return err
		}
		if shape.Kind == "polygon" {
			if err := star.SetStarType(aep.StarTypePolygon); err != nil {
				return err
			}
		}
		if shape.Points != nil {
			if err := star.SetPoints(*shape.Points); err != nil {
				return err
			}
		}
		if len(shape.Position) == 2 {
			if err := star.SetPosition([2]float64{shape.Position[0], shape.Position[1]}); err != nil {
				return err
			}
		}
		if shape.Rotation != nil {
			if err := star.SetRotation(*shape.Rotation); err != nil {
				return err
			}
		}
		if shape.InnerRadius != nil {
			if err := star.SetInnerRadius(*shape.InnerRadius); err != nil {
				return err
			}
		}
		if shape.OuterRadius != nil {
			if err := star.SetOuterRadius(*shape.OuterRadius); err != nil {
				return err
			}
		}
		if shape.InnerRoundness != nil {
			if err := star.SetInnerRoundness(*shape.InnerRoundness); err != nil {
				return err
			}
		}
		if shape.OuterRoundness != nil {
			if err := star.SetOuterRoundness(*shape.OuterRoundness); err != nil {
				return err
			}
		}
	}
	if shape.RoundCorners != nil {
		roundCorners, err := group.AddRoundCorners()
		if err != nil {
			return err
		}
		if shape.RoundCorners.Radius != nil {
			if err := roundCorners.SetRadius(*shape.RoundCorners.Radius); err != nil {
				return err
			}
		}
	}
	if shape.OffsetPaths != nil {
		offsetPaths, err := group.AddOffsetPaths()
		if err != nil {
			return err
		}
		if shape.OffsetPaths.Amount != nil {
			if err := offsetPaths.SetAmount(*shape.OffsetPaths.Amount); err != nil {
				return err
			}
		}
		if shape.OffsetPaths.LineJoin != "" {
			lineJoin, err := offsetLineJoin(shape.OffsetPaths.LineJoin)
			if err != nil {
				return err
			}
			if err := offsetPaths.SetLineJoin(lineJoin); err != nil {
				return err
			}
		}
		if shape.OffsetPaths.MiterLimit != nil {
			if err := offsetPaths.SetMiterLimit(*shape.OffsetPaths.MiterLimit); err != nil {
				return err
			}
		}
		if shape.OffsetPaths.Copies != nil {
			if err := offsetPaths.SetCopies(*shape.OffsetPaths.Copies); err != nil {
				return err
			}
		}
		if shape.OffsetPaths.CopyOffset != nil {
			if err := offsetPaths.SetCopyOffset(*shape.OffsetPaths.CopyOffset); err != nil {
				return err
			}
		}
	}
	if shape.Repeater != nil {
		repeater, err := group.AddRepeater()
		if err != nil {
			return err
		}
		if shape.Repeater.Copies != nil {
			if err := repeater.SetCopies(*shape.Repeater.Copies); err != nil {
				return err
			}
		}
		if shape.Repeater.Offset != nil {
			if err := repeater.SetOffset(*shape.Repeater.Offset); err != nil {
				return err
			}
		}
		if shape.Repeater.Order != "" {
			order, err := repeaterOrder(shape.Repeater.Order)
			if err != nil {
				return err
			}
			if err := repeater.SetOrder(order); err != nil {
				return err
			}
		}
		transform := repeater.Transform()
		if len(shape.Repeater.Anchor) == 2 {
			if err := transform.SetAnchor([2]float64{shape.Repeater.Anchor[0], shape.Repeater.Anchor[1]}); err != nil {
				return err
			}
		}
		if len(shape.Repeater.Position) == 2 {
			if err := transform.SetPosition([2]float64{shape.Repeater.Position[0], shape.Repeater.Position[1]}); err != nil {
				return err
			}
		}
		if len(shape.Repeater.Scale) == 2 {
			if err := transform.SetScale([2]float64{shape.Repeater.Scale[0], shape.Repeater.Scale[1]}); err != nil {
				return err
			}
		}
		if shape.Repeater.Rotation != nil {
			if err := transform.SetRotation(*shape.Repeater.Rotation); err != nil {
				return err
			}
		}
		if shape.Repeater.StartOpacity != nil {
			if err := transform.SetStartOpacity(*shape.Repeater.StartOpacity); err != nil {
				return err
			}
		}
		if shape.Repeater.EndOpacity != nil {
			if err := transform.SetEndOpacity(*shape.Repeater.EndOpacity); err != nil {
				return err
			}
		}
	}
	if shape.MergePaths != nil {
		mergePaths, err := group.AddMergePaths()
		if err != nil {
			return err
		}
		if shape.MergePaths.Type != "" {
			mergeType, err := mergePathsType(shape.MergePaths.Type)
			if err != nil {
				return err
			}
			if err := mergePaths.SetType(mergeType); err != nil {
				return err
			}
		}
	}
	if shape.ZigZag != nil {
		zigZag, err := group.AddZigZag()
		if err != nil {
			return err
		}
		if shape.ZigZag.Size != nil {
			if err := zigZag.SetSize(*shape.ZigZag.Size); err != nil {
				return err
			}
		}
		if shape.ZigZag.Detail != nil {
			if err := zigZag.SetDetail(*shape.ZigZag.Detail); err != nil {
				return err
			}
		}
		if shape.ZigZag.Points != "" {
			points, err := zigZagPoints(shape.ZigZag.Points)
			if err != nil {
				return err
			}
			if err := zigZag.SetPoints(points); err != nil {
				return err
			}
		}
	}
	if shape.PuckerBloat != nil {
		puckerBloat, err := group.AddPuckerBloat()
		if err != nil {
			return err
		}
		if shape.PuckerBloat.Amount != nil {
			if err := puckerBloat.SetAmount(*shape.PuckerBloat.Amount); err != nil {
				return err
			}
		}
	}
	if shape.Twist != nil {
		twist, err := group.AddTwist()
		if err != nil {
			return err
		}
		if shape.Twist.Angle != nil {
			if err := twist.SetAngle(*shape.Twist.Angle); err != nil {
				return err
			}
		}
		if len(shape.Twist.Center) == 2 {
			if err := twist.SetCenter([2]float64{shape.Twist.Center[0], shape.Twist.Center[1]}); err != nil {
				return err
			}
		}
	}
	if shape.WigglePaths != nil {
		wigglePaths, err := group.AddWigglePaths()
		if err != nil {
			return err
		}
		if shape.WigglePaths.Size != nil {
			if err := wigglePaths.SetSize(*shape.WigglePaths.Size); err != nil {
				return err
			}
		}
		if shape.WigglePaths.Detail != nil {
			if err := wigglePaths.SetDetail(*shape.WigglePaths.Detail); err != nil {
				return err
			}
		}
		if shape.WigglePaths.WigglesPerSecond != nil {
			if err := wigglePaths.SetWigglesPerSecond(*shape.WigglePaths.WigglesPerSecond); err != nil {
				return err
			}
		}
		if shape.WigglePaths.RandomSeed != nil {
			if err := wigglePaths.SetRandomSeed(*shape.WigglePaths.RandomSeed); err != nil {
				return err
			}
		}
		if shape.WigglePaths.Points != "" {
			points, err := roughenPoints(shape.WigglePaths.Points)
			if err != nil {
				return err
			}
			if err := wigglePaths.SetPoints(points); err != nil {
				return err
			}
		}
		if shape.WigglePaths.Correlation != nil {
			if err := wigglePaths.SetCorrelation(*shape.WigglePaths.Correlation); err != nil {
				return err
			}
		}
		if shape.WigglePaths.TemporalPhase != nil {
			if err := wigglePaths.SetTemporalPhase(*shape.WigglePaths.TemporalPhase); err != nil {
				return err
			}
		}
		if shape.WigglePaths.SpatialPhase != nil {
			if err := wigglePaths.SetSpatialPhase(*shape.WigglePaths.SpatialPhase); err != nil {
				return err
			}
		}
	}
	if shape.WiggleTransform != nil {
		wiggleTransform, err := group.AddWiggleTransform()
		if err != nil {
			return err
		}
		transform := wiggleTransform.Transform()
		if len(shape.WiggleTransform.Anchor) == 2 {
			if err := transform.SetAnchor([2]float64{shape.WiggleTransform.Anchor[0], shape.WiggleTransform.Anchor[1]}); err != nil {
				return err
			}
		}
		if len(shape.WiggleTransform.Position) == 2 {
			if err := transform.SetPosition([2]float64{shape.WiggleTransform.Position[0], shape.WiggleTransform.Position[1]}); err != nil {
				return err
			}
		}
		if len(shape.WiggleTransform.Scale) == 2 {
			if err := transform.SetScale([2]float64{shape.WiggleTransform.Scale[0], shape.WiggleTransform.Scale[1]}); err != nil {
				return err
			}
		}
		if shape.WiggleTransform.Rotation != nil {
			if err := transform.SetRotation(*shape.WiggleTransform.Rotation); err != nil {
				return err
			}
		}
		if shape.WiggleTransform.WigglesPerSecond != nil {
			if err := wiggleTransform.SetWigglesPerSecond(*shape.WiggleTransform.WigglesPerSecond); err != nil {
				return err
			}
		}
		if shape.WiggleTransform.RandomSeed != nil {
			if err := wiggleTransform.SetRandomSeed(*shape.WiggleTransform.RandomSeed); err != nil {
				return err
			}
		}
		if shape.WiggleTransform.Correlation != nil {
			if err := wiggleTransform.SetCorrelation(*shape.WiggleTransform.Correlation); err != nil {
				return err
			}
		}
		if shape.WiggleTransform.TemporalPhase != nil {
			if err := wiggleTransform.SetTemporalPhase(*shape.WiggleTransform.TemporalPhase); err != nil {
				return err
			}
		}
		if shape.WiggleTransform.SpatialPhase != nil {
			if err := wiggleTransform.SetSpatialPhase(*shape.WiggleTransform.SpatialPhase); err != nil {
				return err
			}
		}
	}
	if shape.Trim != nil {
		trim, err := group.AddTrim()
		if err != nil {
			return err
		}
		if shape.Trim.Start != nil {
			if err := trim.SetStart(*shape.Trim.Start); err != nil {
				return err
			}
		}
		if shape.Trim.End != nil {
			if err := trim.SetEnd(*shape.Trim.End); err != nil {
				return err
			}
		}
		if shape.Trim.Offset != nil {
			if err := trim.SetOffset(*shape.Trim.Offset); err != nil {
				return err
			}
		}
	}
	if len(shape.FillColor) >= 3 || shape.FillOpacity != nil || shape.FillBlendMode != nil || shape.FillCompositeOrder != "" || shape.FillRule != "" {
		fill, err := group.AddFill()
		if err != nil {
			return err
		}
		if len(shape.FillColor) >= 3 {
			if err := fill.SetColor(rgbaColor(shape.FillColor)); err != nil {
				return err
			}
		}
		if shape.FillOpacity != nil {
			if err := fill.SetOpacity(*shape.FillOpacity); err != nil {
				return err
			}
		}
		if shape.FillBlendMode != nil {
			mode, err := shapeBlendMode(*shape.FillBlendMode)
			if err != nil {
				return err
			}
			if err := fill.SetBlendMode(mode); err != nil {
				return err
			}
		}
		if shape.FillCompositeOrder != "" {
			order, err := shapeCompositeOrder(shape.FillCompositeOrder)
			if err != nil {
				return err
			}
			if err := fill.SetCompositeOrder(order); err != nil {
				return err
			}
		}
		if shape.FillRule != "" {
			rule, err := fillRule(shape.FillRule)
			if err != nil {
				return err
			}
			if err := fill.SetFillRule(rule); err != nil {
				return err
			}
		}
	}
	if shape.GradientFill != nil {
		fill, err := group.AddGradientFill()
		if err != nil {
			return err
		}
		if shape.GradientFill.Type != "" {
			typ, err := gradientType(shape.GradientFill.Type)
			if err != nil {
				return err
			}
			if err := fill.SetGradientType(typ); err != nil {
				return err
			}
		}
		if len(shape.GradientFill.StartPoint) == 2 {
			if err := fill.SetStartPoint([2]float64{shape.GradientFill.StartPoint[0], shape.GradientFill.StartPoint[1]}); err != nil {
				return err
			}
		}
		if len(shape.GradientFill.EndPoint) == 2 {
			if err := fill.SetEndPoint([2]float64{shape.GradientFill.EndPoint[0], shape.GradientFill.EndPoint[1]}); err != nil {
				return err
			}
		}
		if shape.GradientFill.HighlightLength != nil {
			if err := fill.SetHighlightLength(*shape.GradientFill.HighlightLength); err != nil {
				return err
			}
		}
		if shape.GradientFill.HighlightAngle != nil {
			if err := fill.SetHighlightAngle(*shape.GradientFill.HighlightAngle); err != nil {
				return err
			}
		}
		if len(shape.GradientFill.ColorStops) > 0 {
			stops, err := gradientColorStops(shape.GradientFill.ColorStops)
			if err != nil {
				return err
			}
			if err := fill.SetColorStops(stops); err != nil {
				return err
			}
		}
		if len(shape.GradientFill.AlphaStops) > 0 {
			stops := gradientAlphaStops(shape.GradientFill.AlphaStops)
			if err := fill.SetAlphaStops(stops); err != nil {
				return err
			}
		}
	}
	if shape.GradientStroke != nil {
		stroke, err := group.AddGradientStroke()
		if err != nil {
			return err
		}
		if shape.GradientStroke.Type != "" {
			typ, err := gradientType(shape.GradientStroke.Type)
			if err != nil {
				return err
			}
			if err := stroke.SetGradientType(typ); err != nil {
				return err
			}
		}
		if len(shape.GradientStroke.StartPoint) == 2 {
			if err := stroke.SetStartPoint([2]float64{shape.GradientStroke.StartPoint[0], shape.GradientStroke.StartPoint[1]}); err != nil {
				return err
			}
		}
		if len(shape.GradientStroke.EndPoint) == 2 {
			if err := stroke.SetEndPoint([2]float64{shape.GradientStroke.EndPoint[0], shape.GradientStroke.EndPoint[1]}); err != nil {
				return err
			}
		}
		if shape.GradientStroke.HighlightLength != nil {
			if err := stroke.SetHighlightLength(*shape.GradientStroke.HighlightLength); err != nil {
				return err
			}
		}
		if shape.GradientStroke.HighlightAngle != nil {
			if err := stroke.SetHighlightAngle(*shape.GradientStroke.HighlightAngle); err != nil {
				return err
			}
		}
		if shape.GradientStroke.Width != nil {
			if err := stroke.SetStrokeWidth(*shape.GradientStroke.Width); err != nil {
				return err
			}
		}
		if shape.GradientStroke.LineCap != "" {
			lineCap, err := strokeLineCap(shape.GradientStroke.LineCap)
			if err != nil {
				return err
			}
			if err := stroke.SetLineCap(lineCap); err != nil {
				return err
			}
		}
		if shape.GradientStroke.LineJoin != "" {
			lineJoin, err := strokeLineJoin(shape.GradientStroke.LineJoin)
			if err != nil {
				return err
			}
			if err := stroke.SetLineJoin(lineJoin); err != nil {
				return err
			}
		}
		if shape.GradientStroke.MiterLimit != nil {
			if err := stroke.SetMiterLimit(*shape.GradientStroke.MiterLimit); err != nil {
				return err
			}
		}
		if len(shape.GradientStroke.ColorStops) > 0 {
			stops, err := gradientColorStops(shape.GradientStroke.ColorStops)
			if err != nil {
				return err
			}
			if err := stroke.SetColorStops(stops); err != nil {
				return err
			}
		}
		if len(shape.GradientStroke.AlphaStops) > 0 {
			stops := gradientAlphaStops(shape.GradientStroke.AlphaStops)
			if err := stroke.SetAlphaStops(stops); err != nil {
				return err
			}
		}
	}
	if shape.Stroke != nil {
		stroke, err := group.AddStroke()
		if err != nil {
			return err
		}
		if len(shape.Stroke.Color) >= 3 {
			color := rgbaColor(shape.Stroke.Color)
			if err := stroke.SetColor(color); err != nil {
				return err
			}
		}
		if shape.Stroke.Width != nil {
			if err := stroke.SetWidth(*shape.Stroke.Width); err != nil {
				return err
			}
		}
		if shape.Stroke.Opacity != nil {
			if err := stroke.SetOpacity(*shape.Stroke.Opacity); err != nil {
				return err
			}
		}
		if shape.Stroke.LineCap != "" {
			lineCap, err := strokeLineCap(shape.Stroke.LineCap)
			if err != nil {
				return err
			}
			if err := stroke.SetLineCap(lineCap); err != nil {
				return err
			}
		}
		if shape.Stroke.LineJoin != "" {
			lineJoin, err := strokeLineJoin(shape.Stroke.LineJoin)
			if err != nil {
				return err
			}
			if err := stroke.SetLineJoin(lineJoin); err != nil {
				return err
			}
		}
		if shape.Stroke.MiterLimit != nil {
			if err := stroke.SetMiterLimit(*shape.Stroke.MiterLimit); err != nil {
				return err
			}
		}
		if shape.Stroke.CompositeOrder != "" {
			order, err := shapeCompositeOrder(shape.Stroke.CompositeOrder)
			if err != nil {
				return err
			}
			if err := stroke.SetCompositeOrder(order); err != nil {
				return err
			}
		}
		if shape.Stroke.Taper != nil {
			taper := stroke.Taper()
			if shape.Stroke.Taper.StartLength != nil {
				if err := taper.SetStartLength(*shape.Stroke.Taper.StartLength); err != nil {
					return err
				}
			}
			if shape.Stroke.Taper.EndLength != nil {
				if err := taper.SetEndLength(*shape.Stroke.Taper.EndLength); err != nil {
					return err
				}
			}
			if shape.Stroke.Taper.StartWidth != nil {
				if err := taper.SetStartWidth(*shape.Stroke.Taper.StartWidth); err != nil {
					return err
				}
			}
			if shape.Stroke.Taper.EndWidth != nil {
				if err := taper.SetEndWidth(*shape.Stroke.Taper.EndWidth); err != nil {
					return err
				}
			}
			if shape.Stroke.Taper.StartEase != nil {
				if err := taper.SetStartEase(*shape.Stroke.Taper.StartEase); err != nil {
					return err
				}
			}
			if shape.Stroke.Taper.EndEase != nil {
				if err := taper.SetEndEase(*shape.Stroke.Taper.EndEase); err != nil {
					return err
				}
			}
		}
		if shape.Stroke.Wave != nil {
			wave := stroke.Wave()
			if shape.Stroke.Wave.Amount != nil {
				if err := wave.SetAmount(*shape.Stroke.Wave.Amount); err != nil {
					return err
				}
			}
			if shape.Stroke.Wave.Wavelength != nil {
				if err := wave.SetWavelength(*shape.Stroke.Wave.Wavelength); err != nil {
					return err
				}
			}
			if shape.Stroke.Wave.Phase != nil {
				if err := wave.SetPhase(*shape.Stroke.Wave.Phase); err != nil {
					return err
				}
			}
		}
		if shape.Stroke.Dashes != nil {
			dashes := stroke.Dashes()
			if shape.Stroke.Dashes.Dash != nil {
				if err := dashes.SetDash(*shape.Stroke.Dashes.Dash); err != nil {
					return err
				}
			}
			if shape.Stroke.Dashes.Gap != nil {
				if err := dashes.SetGap(*shape.Stroke.Dashes.Gap); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func offsetLineJoin(value string) (aep.StrokeLineJoin, error) {
	switch value {
	case "miter":
		return aep.StrokeLineJoinMiter, nil
	case "round":
		return aep.StrokeLineJoinRound, nil
	case "bevel":
		return aep.StrokeLineJoinBevel, nil
	default:
		return 0, fmt.Errorf("unsupported offset line_join %q", value)
	}
}

func strokeLineCap(value string) (aep.StrokeLineCap, error) {
	switch value {
	case "butt":
		return aep.StrokeLineCapButt, nil
	case "round":
		return aep.StrokeLineCapRound, nil
	case "projecting":
		return aep.StrokeLineCapProjecting, nil
	default:
		return 0, fmt.Errorf("unsupported stroke line_cap %q", value)
	}
}

func strokeLineJoin(value string) (aep.StrokeLineJoin, error) {
	switch value {
	case "miter":
		return aep.StrokeLineJoinMiter, nil
	case "round":
		return aep.StrokeLineJoinRound, nil
	case "bevel":
		return aep.StrokeLineJoinBevel, nil
	default:
		return 0, fmt.Errorf("unsupported stroke line_join %q", value)
	}
}

func shapeCompositeOrder(value string) (aep.ShapeCompositeOrder, error) {
	switch value {
	case "above_previous":
		return aep.ShapeCompositeOrderAbovePrevious, nil
	case "below_previous":
		return aep.ShapeCompositeOrderBelowPrevious, nil
	default:
		return 0, fmt.Errorf("unsupported shape composite order %q", value)
	}
}

func shapeBlendMode(value float64) (aep.ShapeBlendMode, error) {
	if value < 1 || !isWholeNumber(value) {
		return 0, fmt.Errorf("shape blend mode must be an integer of at least 1, got %g", value)
	}
	return aep.ShapeBlendMode(value), nil
}

func fillRule(value string) (aep.FillRule, error) {
	switch value {
	case "nonzero_winding":
		return aep.FillRuleNonzeroWinding, nil
	case "even_odd":
		return aep.FillRuleEvenOdd, nil
	default:
		return 0, fmt.Errorf("unsupported fill rule %q", value)
	}
}

func gradientType(value string) (aep.GradientType, error) {
	switch value {
	case "linear":
		return aep.GradientLinear, nil
	case "radial":
		return aep.GradientRadial, nil
	default:
		return 0, fmt.Errorf("unsupported gradient type %q", value)
	}
}

func gradientColorStops(specs []GradientColorStopSpec) ([]aep.GradientColorStop, error) {
	stops := make([]aep.GradientColorStop, 0, len(specs))
	for i, spec := range specs {
		if len(spec.Color) < 3 {
			return nil, fmt.Errorf("gradient color stop %d needs at least 3 color channels", i)
		}
		midpoint := 0.5
		if spec.Midpoint != nil {
			midpoint = *spec.Midpoint
		}
		stops = append(stops, aep.GradientColorStop{
			Offset:   spec.Offset,
			Midpoint: midpoint,
			Color: [3]float64{
				toUnitColor(spec.Color[0]),
				toUnitColor(spec.Color[1]),
				toUnitColor(spec.Color[2]),
			},
		})
	}
	return stops, nil
}

func gradientAlphaStops(specs []GradientAlphaStopSpec) []aep.GradientAlphaStop {
	stops := make([]aep.GradientAlphaStop, 0, len(specs))
	for _, spec := range specs {
		midpoint := 0.5
		if spec.Midpoint != nil {
			midpoint = *spec.Midpoint
		}
		stops = append(stops, aep.GradientAlphaStop{
			Offset:   spec.Offset,
			Midpoint: midpoint,
			Alpha:    spec.Alpha,
		})
	}
	return stops
}

func repeaterOrder(value string) (aep.RepeaterOrder, error) {
	switch value {
	case "below":
		return aep.RepeaterOrderBelow, nil
	case "above":
		return aep.RepeaterOrderAbove, nil
	default:
		return 0, fmt.Errorf("unsupported repeater order %q", value)
	}
}

func mergePathsType(value string) (aep.MergeType, error) {
	switch value {
	case "merge":
		return aep.MergeTypeMerge, nil
	case "add":
		return aep.MergeTypeAdd, nil
	case "subtract":
		return aep.MergeTypeSubtract, nil
	case "intersect":
		return aep.MergeTypeIntersect, nil
	case "exclude":
		return aep.MergeTypeExclude, nil
	default:
		return 0, fmt.Errorf("unsupported merge_paths type %q", value)
	}
}

func zigZagPoints(value string) (aep.ZigZagPoints, error) {
	switch value {
	case "corner":
		return aep.ZigZagPointsCorner, nil
	case "smooth":
		return aep.ZigZagPointsSmooth, nil
	default:
		return 0, fmt.Errorf("unsupported zigzag points %q", value)
	}
}

func roughenPoints(value string) (aep.RoughenPoints, error) {
	switch value {
	case "corner":
		return aep.RoughenPointsCorner, nil
	case "smooth":
		return aep.RoughenPointsSmooth, nil
	default:
		return 0, fmt.Errorf("unsupported wiggle_paths points %q", value)
	}
}
