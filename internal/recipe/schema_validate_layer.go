package recipe

import (
	"fmt"
)

func validateLayer(layer Layer, layerPath string, compDuration float64, recordCapability func(string, string) CapabilityLookup, addRefusal func(string, string, string)) {
	switch layer.Type {
	case "solid":
		recordCapability("NewSolidLayer", layerPath)
	case "precomp":
		recordCapability("NewPrecompLayer", layerPath)
	case "text":
		recordCapability("NewTextLayer", layerPath)
		if layer.Text != "" {
			recordCapability("Layer.SetText", layerPath+".text")
		}
	case "shape":
		recordCapability("NewShapeLayer", layerPath)
	case "null":
		recordCapability("NewNullLayer", layerPath)
	case "adjustment":
		recordCapability("NewAdjustmentLayer", layerPath)
	case "camera":
		recordCapability("NewCameraLayer", layerPath)
	case "light":
		recordCapability("NewLightLayer", layerPath)
	default:
		addRefusal("unsupported_layer_type", layerPath+".type", fmt.Sprintf("unsupported layer type %q", layer.Type))
	}
	if layer.Name == "" {
		addRefusal("missing_layer_name", layerPath+".name", "layer name is required")
	}
	if layer.Label != nil {
		recordCapability("Layer.SetLabel", layerPath+".label")
		validateLayerLabel(*layer.Label, layerPath+".label", addRefusal)
	}
	if layer.Comment != "" {
		recordCapability("Layer.SetComment", layerPath+".comment")
	}
	if layer.Visible != nil {
		recordCapability("Layer.SetVisible", layerPath+".visible")
	}
	if layer.Solo != nil {
		recordCapability("Layer.SetSolo", layerPath+".solo")
	}
	if layer.Locked != nil {
		recordCapability("Layer.SetLocked", layerPath+".locked")
	}
	if layer.MotionBlur != nil {
		recordCapability("Layer.SetMotionBlur", layerPath+".motion_blur")
	}
	if layer.Shy != nil {
		recordCapability("Layer.SetShy", layerPath+".shy")
	}
	if layer.EffectsEnabled != nil {
		recordCapability("Layer.SetEffectsEnabled", layerPath+".effects_enabled")
	}
	if layer.AudioEnabled != nil {
		recordCapability("Layer.SetAudioEnabled", layerPath+".audio_enabled")
	}
	if layer.FrameBlendEnabled != nil {
		recordCapability("Layer.SetFrameBlendEnabled", layerPath+".frame_blend_enabled")
	}
	if layer.MarkersLocked != nil {
		recordCapability("Layer.SetMarkersLocked", layerPath+".markers_locked")
	}
	if layer.CollapseTransform != nil {
		recordCapability("Layer.SetCollapseTransform", layerPath+".collapse_transform")
	}
	if layer.Is3D != nil {
		recordCapability("Layer.SetIs3D", layerPath+".is_3d")
	}
	if layer.IsAdjust != nil {
		recordCapability("Layer.SetIsAdjust", layerPath+".is_adjust")
	}
	if layer.IsNull != nil {
		recordCapability("Layer.SetIsNull", layerPath+".is_null")
	}
	if layer.IsGuide != nil {
		recordCapability("Layer.SetIsGuide", layerPath+".is_guide")
	}
	if layer.SamplingBicubic != nil {
		recordCapability("Layer.SetSamplingBicubic", layerPath+".sampling_bicubic")
	}
	if layer.FrameBlendPixelMotion != nil {
		recordCapability("Layer.SetFrameBlendPixelMotion", layerPath+".frame_blend_pixel_motion")
	}
	if layer.PreserveTransparency != nil {
		recordCapability("Layer.SetPreserveTransparency", layerPath+".preserve_transparency")
	}
	if layer.Quality != "" {
		recordCapability("Layer.SetQuality", layerPath+".quality")
		if _, err := layerQuality(layer.Quality); err != nil {
			addRefusal("invalid_layer_quality", layerPath+".quality", "quality must be wireframe, draft, or best")
		}
	}
	if layer.BlendingMode != "" {
		recordCapability("Layer.SetBlendingMode", layerPath+".blending_mode")
		if _, err := layerBlendingMode(layer.BlendingMode); err != nil {
			addRefusal("invalid_layer_blending_mode", layerPath+".blending_mode", "blending_mode is not supported")
		}
	}
	if layer.TrackMatte != "" {
		recordCapability("Layer.SetTrackMatte", layerPath+".track_matte")
		if _, err := layerTrackMatte(layer.TrackMatte); err != nil {
			addRefusal("invalid_layer_track_matte", layerPath+".track_matte", "track_matte must be none, alpha, alpha_inverse, luma, or luma_inverse")
		}
	}
	if layer.AutoOrient != "" {
		recordCapability("Layer.SetAutoOrient", layerPath+".auto_orient")
		if _, err := layerAutoOrient(layer.AutoOrient); err != nil {
			addRefusal("invalid_layer_auto_orient", layerPath+".auto_orient", "auto_orient must be none, along_path, camera_or_point_of_interest, or characters_toward_camera")
		}
	}
	if layer.Camera != nil {
		if layer.Type != "camera" {
			addRefusal("camera_options_on_non_camera_layer", layerPath+".camera", "camera options require type camera")
		}
		if layer.Camera.Zoom != nil {
			recordCapability("Layer.SetCameraZoom", layerPath+".camera.zoom")
		}
		if layer.Camera.DepthOfField != nil {
			recordCapability("Layer.SetCameraDepthOfField", layerPath+".camera.depth_of_field")
		}
		if layer.Camera.FocusDistance != nil {
			recordCapability("Layer.SetCameraFocusDistance", layerPath+".camera.focus_distance")
		}
		if layer.Camera.Aperture != nil {
			recordCapability("Layer.SetCameraAperture", layerPath+".camera.aperture")
		}
		if layer.Camera.BlurLevel != nil {
			recordCapability("Layer.SetCameraBlurLevel", layerPath+".camera.blur_level")
		}
		if layer.Camera.IrisShape != nil {
			recordCapability("SetIrisShape", layerPath+".camera.iris_shape")
		}
		if layer.Camera.IrisRotation != nil {
			recordCapability("SetIrisRotation", layerPath+".camera.iris_rotation")
		}
		if layer.Camera.IrisRoundness != nil {
			recordCapability("SetIrisRoundness", layerPath+".camera.iris_roundness")
		}
		if layer.Camera.IrisAspectRatio != nil {
			recordCapability("SetIrisAspectRatio", layerPath+".camera.iris_aspect_ratio")
		}
		if layer.Camera.IrisDiffractionFringe != nil {
			recordCapability("SetIrisDiffractionFringe", layerPath+".camera.iris_diffraction_fringe")
		}
		if layer.Camera.IrisHighlightGain != nil {
			recordCapability("SetIrisHighlightGain", layerPath+".camera.iris_highlight_gain")
		}
		if layer.Camera.IrisHighlightThreshold != nil {
			recordCapability("SetIrisHighlightThreshold", layerPath+".camera.iris_highlight_threshold")
		}
		if layer.Camera.IrisHighlightSaturation != nil {
			recordCapability("SetIrisHighlightSaturation", layerPath+".camera.iris_highlight_saturation")
		}
	}
	if layer.Light != nil {
		if layer.Type != "light" {
			addRefusal("light_options_on_non_light_layer", layerPath+".light", "light options require type light")
		}
		if layer.Light.Kind != "" {
			recordCapability("SetLightKind", layerPath+".light.kind")
		}
		if layer.Light.SourceLayer != "" {
			recordCapability("SetLightSource", layerPath+".light.source_layer")
		}
		if layer.Light.Intensity != nil {
			recordCapability("SetLightIntensity", layerPath+".light.intensity")
		}
		if len(layer.Light.Color) > 0 {
			recordCapability("SetLightColor", layerPath+".light.color")
			validateColor(layer.Light.Color, layerPath+".light.color", "invalid_light_color", addRefusal)
		}
		if layer.Light.CastsShadows != nil {
			recordCapability("SetLightCastsShadows", layerPath+".light.casts_shadows")
		}
		if layer.Light.ShadowDarkness != nil {
			recordCapability("SetLightShadowDarkness", layerPath+".light.shadow_darkness")
		}
		if layer.Light.ShadowDiffusion != nil {
			recordCapability("SetLightShadowDiffusion", layerPath+".light.shadow_diffusion")
		}
		if layer.Light.FalloffType != nil {
			recordCapability("SetLightFalloffType", layerPath+".light.falloff_type")
		}
		if layer.Light.FalloffStart != nil {
			recordCapability("SetLightFalloffStart", layerPath+".light.falloff_start")
		}
		if layer.Light.FalloffDistance != nil {
			recordCapability("SetLightFalloffDistance", layerPath+".light.falloff_distance")
		}
		if layer.Light.ConeAngle != nil {
			recordCapability("SetLightConeAngle", layerPath+".light.cone_angle")
		}
		if layer.Light.ConeFeather != nil {
			recordCapability("SetLightConeFeather", layerPath+".light.cone_feather")
		}
	}
	if layer.StartTime != nil {
		recordCapability("Layer.SetStartTime", layerPath+".start_time")
	}
	if layer.InPoint != nil {
		recordCapability("Layer.SetInPoint", layerPath+".in_point")
		if *layer.InPoint < 0 || *layer.InPoint > compDuration {
			addRefusal("invalid_layer_in_point", layerPath+".in_point", "in_point must be between 0 and comp duration")
		}
	}
	if layer.OutPoint != nil {
		recordCapability("Layer.SetOutPoint", layerPath+".out_point")
		if *layer.OutPoint < 0 || *layer.OutPoint > compDuration {
			addRefusal("invalid_layer_out_point", layerPath+".out_point", "out_point must be between 0 and comp duration")
		}
		if layer.InPoint != nil && *layer.OutPoint < *layer.InPoint {
			addRefusal("invalid_layer_out_point", layerPath+".out_point", "out_point must be greater than or equal to in_point")
		}
	}
	if layer.Stretch != nil {
		recordCapability("Layer.SetStretch", layerPath+".stretch")
		if *layer.Stretch <= 0 {
			addRefusal("invalid_layer_stretch", layerPath+".stretch", "stretch must be greater than 0")
		}
	}
	if layer.TextStyle != nil {
		if layer.Type != "text" {
			addRefusal("text_style_on_non_text_layer", layerPath+".text_style", "text_style is only supported on text layers")
		} else {
			validateTextStyle(*layer.TextStyle, layerPath+".text_style", recordCapability, addRefusal)
		}
	}
	for i, animator := range layer.TextAnimators {
		animatorPath := fmt.Sprintf("%s.text_animators[%d]", layerPath, i)
		if layer.Type != "text" {
			addRefusal("text_animator_on_non_text_layer", animatorPath, "text_animators are only supported on text layers")
		}
		validateTextAnimator(animator, animatorPath, recordCapability, addRefusal)
	}
	if layer.Type == "shape" && layer.Shape != nil {
		switch layer.Shape.Kind {
		case "rect", "ellipse":
			if layer.Shape.Kind == "rect" {
				recordCapability("RectNode.SetSize", layerPath+".shape.size")
				if len(layer.Shape.Position) > 0 {
					recordCapability("RectNode.SetPosition", layerPath+".shape.position")
				}
				if layer.Shape.Roundness != nil {
					recordCapability("RectNode.SetRoundness", layerPath+".shape.roundness")
					if *layer.Shape.Roundness < 0 {
						addRefusal("invalid_shape_roundness", layerPath+".shape.roundness", "shape roundness must be non-negative")
					}
				}
			} else {
				recordCapability("EllipseNode.SetSize", layerPath+".shape.size")
				if len(layer.Shape.Position) > 0 {
					recordCapability("EllipseNode.SetPosition", layerPath+".shape.position")
				}
				if layer.Shape.Roundness != nil {
					addRefusal("unsupported_shape_roundness", layerPath+".shape.roundness", "roundness is only supported for rect shapes")
				}
			}
		case "star", "polygon":
			recordCapability("VectorGroup.AddStar", layerPath+".shape.kind")
			if layer.Shape.Kind == "polygon" {
				recordCapability("StarNode.SetStarType", layerPath+".shape.kind")
			}
			if layer.Shape.Points != nil {
				recordCapability("StarNode.SetPoints", layerPath+".shape.points")
				if *layer.Shape.Points < 3 {
					addRefusal("invalid_shape_star_points", layerPath+".shape.points", "star points must be at least 3")
				}
			}
			if len(layer.Shape.Position) > 0 {
				recordCapability("StarNode.SetPosition", layerPath+".shape.position")
			}
			if layer.Shape.Rotation != nil {
				recordCapability("StarNode.SetRotation", layerPath+".shape.rotation")
			}
			if layer.Shape.InnerRadius != nil {
				recordCapability("StarNode.SetInnerRadius", layerPath+".shape.inner_radius")
				if *layer.Shape.InnerRadius < 0 {
					addRefusal("invalid_shape_star_inner_radius", layerPath+".shape.inner_radius", "star inner_radius must be non-negative")
				}
			}
			if layer.Shape.OuterRadius != nil {
				recordCapability("StarNode.SetOuterRadius", layerPath+".shape.outer_radius")
				if *layer.Shape.OuterRadius < 0 {
					addRefusal("invalid_shape_star_outer_radius", layerPath+".shape.outer_radius", "star outer_radius must be non-negative")
				}
			}
			if layer.Shape.InnerRoundness != nil {
				recordCapability("StarNode.SetInnerRoundness", layerPath+".shape.inner_roundness")
			}
			if layer.Shape.OuterRoundness != nil {
				recordCapability("StarNode.SetOuterRoundness", layerPath+".shape.outer_roundness")
			}
			if layer.Shape.Roundness != nil {
				addRefusal("unsupported_shape_roundness", layerPath+".shape.roundness", "roundness is only supported for rect shapes")
			}
		default:
			addRefusal("unsupported_shape_kind", layerPath+".shape.kind", fmt.Sprintf("unsupported shape kind %q", layer.Shape.Kind))
		}
		validateVec(layer.Shape.Position, 2, layerPath+".shape.position", addRefusal)
		if len(layer.Shape.FillColor) > 0 {
			recordCapability("FillNode.SetColor", layerPath+".shape.fill_color")
			validateColor(layer.Shape.FillColor, layerPath+".shape.fill_color", "invalid_shape_fill_color", addRefusal)
		}
		if layer.Shape.FillOpacity != nil {
			recordCapability("FillNode.SetOpacity", layerPath+".shape.fill_opacity")
			if *layer.Shape.FillOpacity < 0 || *layer.Shape.FillOpacity > 100 {
				addRefusal("invalid_shape_fill_opacity", layerPath+".shape.fill_opacity", "fill opacity must be between 0 and 100")
			}
		}
		if layer.Shape.FillBlendMode != nil {
			recordCapability("FillNode.SetBlendMode", layerPath+".shape.fill_blend_mode")
			if !validShapeBlendMode(*layer.Shape.FillBlendMode) {
				addRefusal("invalid_shape_fill_blend_mode", layerPath+".shape.fill_blend_mode", "fill_blend_mode must be an integer of at least 1")
			}
		}
		if layer.Shape.FillCompositeOrder != "" {
			recordCapability("FillNode.SetCompositeOrder", layerPath+".shape.fill_composite_order")
			if !validShapeCompositeOrder(layer.Shape.FillCompositeOrder) {
				addRefusal("invalid_shape_fill_composite_order", layerPath+".shape.fill_composite_order", "fill_composite_order must be above_previous or below_previous")
			}
		}
		if layer.Shape.FillRule != "" {
			recordCapability("FillNode.SetFillRule", layerPath+".shape.fill_rule")
			if !validFillRule(layer.Shape.FillRule) {
				addRefusal("invalid_shape_fill_rule", layerPath+".shape.fill_rule", "fill_rule must be nonzero_winding or even_odd")
			}
		}
		if layer.Shape.GradientFill != nil {
			gradientPath := layerPath + ".shape.gradient_fill"
			recordCapability("VectorGroup.AddGradientFill", gradientPath)
			if layer.Shape.GradientFill.Type != "" {
				recordCapability("GradientFillNode.SetGradientType", gradientPath+".type")
				if !validGradientType(layer.Shape.GradientFill.Type) {
					addRefusal("invalid_shape_gradient_fill_type", gradientPath+".type", "gradient_fill type must be linear or radial")
				}
			}
			if len(layer.Shape.GradientFill.StartPoint) > 0 {
				recordCapability("GradientFillNode.SetStartPoint", gradientPath+".start_point")
				validateVec(layer.Shape.GradientFill.StartPoint, 2, gradientPath+".start_point", addRefusal)
			}
			if len(layer.Shape.GradientFill.EndPoint) > 0 {
				recordCapability("GradientFillNode.SetEndPoint", gradientPath+".end_point")
				validateVec(layer.Shape.GradientFill.EndPoint, 2, gradientPath+".end_point", addRefusal)
			}
			if layer.Shape.GradientFill.HighlightLength != nil {
				recordCapability("GradientFillNode.SetHighlightLength", gradientPath+".highlight_length")
				if *layer.Shape.GradientFill.HighlightLength < -100 || *layer.Shape.GradientFill.HighlightLength > 100 {
					addRefusal("invalid_shape_gradient_fill_highlight_length", gradientPath+".highlight_length", "gradient_fill highlight_length must be between -100 and 100")
				}
			}
			if layer.Shape.GradientFill.HighlightAngle != nil {
				recordCapability("GradientFillNode.SetHighlightAngle", gradientPath+".highlight_angle")
			}
			if len(layer.Shape.GradientFill.ColorStops) > 0 {
				recordCapability("GradientFillNode.SetColorStops", gradientPath+".color_stops")
				validateGradientColorStops(layer.Shape.GradientFill.ColorStops, gradientPath+".color_stops", "invalid_shape_gradient_fill_color_stops", "gradient_fill", addRefusal)
			}
			if len(layer.Shape.GradientFill.AlphaStops) > 0 {
				recordCapability("GradientFillNode.SetAlphaStops", gradientPath+".alpha_stops")
				validateGradientAlphaStops(layer.Shape.GradientFill.AlphaStops, gradientPath+".alpha_stops", "invalid_shape_gradient_fill_alpha_stops", "gradient_fill", addRefusal)
			}
		}
		if layer.Shape.GradientStroke != nil {
			gradientPath := layerPath + ".shape.gradient_stroke"
			recordCapability("VectorGroup.AddGradientStroke", gradientPath)
			if layer.Shape.GradientStroke.Type != "" {
				recordCapability("GradientStrokeNode.SetGradientType", gradientPath+".type")
				if !validGradientType(layer.Shape.GradientStroke.Type) {
					addRefusal("invalid_shape_gradient_stroke_type", gradientPath+".type", "gradient_stroke type must be linear or radial")
				}
			}
			if len(layer.Shape.GradientStroke.StartPoint) > 0 {
				recordCapability("GradientStrokeNode.SetStartPoint", gradientPath+".start_point")
				validateVec(layer.Shape.GradientStroke.StartPoint, 2, gradientPath+".start_point", addRefusal)
			}
			if len(layer.Shape.GradientStroke.EndPoint) > 0 {
				recordCapability("GradientStrokeNode.SetEndPoint", gradientPath+".end_point")
				validateVec(layer.Shape.GradientStroke.EndPoint, 2, gradientPath+".end_point", addRefusal)
			}
			if layer.Shape.GradientStroke.HighlightLength != nil {
				recordCapability("GradientStrokeNode.SetHighlightLength", gradientPath+".highlight_length")
				if *layer.Shape.GradientStroke.HighlightLength < -100 || *layer.Shape.GradientStroke.HighlightLength > 100 {
					addRefusal("invalid_shape_gradient_stroke_highlight_length", gradientPath+".highlight_length", "gradient_stroke highlight_length must be between -100 and 100")
				}
			}
			if layer.Shape.GradientStroke.HighlightAngle != nil {
				recordCapability("GradientStrokeNode.SetHighlightAngle", gradientPath+".highlight_angle")
			}
			if layer.Shape.GradientStroke.Width != nil {
				recordCapability("GradientStrokeNode.SetStrokeWidth", gradientPath+".width")
				if *layer.Shape.GradientStroke.Width < 0 {
					addRefusal("invalid_shape_gradient_stroke_width", gradientPath+".width", "gradient_stroke width must be non-negative")
				}
			}
			if layer.Shape.GradientStroke.LineCap != "" {
				recordCapability("GradientStrokeNode.SetLineCap", gradientPath+".line_cap")
				if !validStrokeLineCap(layer.Shape.GradientStroke.LineCap) {
					addRefusal("invalid_shape_gradient_stroke_line_cap", gradientPath+".line_cap", "gradient_stroke line_cap must be butt, round, or projecting")
				}
			}
			if layer.Shape.GradientStroke.LineJoin != "" {
				recordCapability("GradientStrokeNode.SetLineJoin", gradientPath+".line_join")
				if !validStrokeLineJoin(layer.Shape.GradientStroke.LineJoin) {
					addRefusal("invalid_shape_gradient_stroke_line_join", gradientPath+".line_join", "gradient_stroke line_join must be miter, round, or bevel")
				}
			}
			if layer.Shape.GradientStroke.MiterLimit != nil {
				recordCapability("GradientStrokeNode.SetMiterLimit", gradientPath+".miter_limit")
				if *layer.Shape.GradientStroke.MiterLimit < 1 {
					addRefusal("invalid_shape_gradient_stroke_miter_limit", gradientPath+".miter_limit", "gradient_stroke miter_limit must be at least 1")
				}
			}
			if len(layer.Shape.GradientStroke.ColorStops) > 0 {
				recordCapability("GradientStrokeNode.SetColorStops", gradientPath+".color_stops")
				validateGradientColorStops(layer.Shape.GradientStroke.ColorStops, gradientPath+".color_stops", "invalid_shape_gradient_stroke_color_stops", "gradient_stroke", addRefusal)
			}
			if len(layer.Shape.GradientStroke.AlphaStops) > 0 {
				recordCapability("GradientStrokeNode.SetAlphaStops", gradientPath+".alpha_stops")
				validateGradientAlphaStops(layer.Shape.GradientStroke.AlphaStops, gradientPath+".alpha_stops", "invalid_shape_gradient_stroke_alpha_stops", "gradient_stroke", addRefusal)
			}
		}
		if layer.Shape.Stroke != nil {
			strokePath := layerPath + ".shape.stroke"
			recordCapability("VectorGroup.AddStroke", strokePath)
			if len(layer.Shape.Stroke.Color) > 0 {
				recordCapability("StrokeNode.SetColor", strokePath+".color")
				validateColor(layer.Shape.Stroke.Color, strokePath+".color", "invalid_shape_stroke_color", addRefusal)
			}
			if layer.Shape.Stroke.Width != nil {
				recordCapability("StrokeNode.SetWidth", strokePath+".width")
				if *layer.Shape.Stroke.Width < 0 {
					addRefusal("invalid_shape_stroke_width", strokePath+".width", "stroke width must be non-negative")
				}
			}
			if layer.Shape.Stroke.Opacity != nil {
				recordCapability("StrokeNode.SetOpacity", strokePath+".opacity")
				if *layer.Shape.Stroke.Opacity < 0 || *layer.Shape.Stroke.Opacity > 100 {
					addRefusal("invalid_shape_stroke_opacity", strokePath+".opacity", "stroke opacity must be between 0 and 100")
				}
			}
			if layer.Shape.Stroke.LineCap != "" {
				recordCapability("StrokeNode.SetLineCap", strokePath+".line_cap")
				if !validStrokeLineCap(layer.Shape.Stroke.LineCap) {
					addRefusal("invalid_shape_stroke_line_cap", strokePath+".line_cap", "stroke line_cap must be butt, round, or projecting")
				}
			}
			if layer.Shape.Stroke.LineJoin != "" {
				recordCapability("StrokeNode.SetLineJoin", strokePath+".line_join")
				if !validStrokeLineJoin(layer.Shape.Stroke.LineJoin) {
					addRefusal("invalid_shape_stroke_line_join", strokePath+".line_join", "stroke line_join must be miter, round, or bevel")
				}
			}
			if layer.Shape.Stroke.MiterLimit != nil {
				recordCapability("StrokeNode.SetMiterLimit", strokePath+".miter_limit")
				if *layer.Shape.Stroke.MiterLimit < 1 {
					addRefusal("invalid_shape_stroke_miter_limit", strokePath+".miter_limit", "stroke miter_limit must be at least 1")
				}
			}
			if layer.Shape.Stroke.CompositeOrder != "" {
				recordCapability("StrokeNode.SetCompositeOrder", strokePath+".composite_order")
				if !validShapeCompositeOrder(layer.Shape.Stroke.CompositeOrder) {
					addRefusal("invalid_shape_stroke_composite_order", strokePath+".composite_order", "stroke composite_order must be above_previous or below_previous")
				}
			}
			if layer.Shape.Stroke.Taper != nil {
				taperPath := strokePath + ".taper"
				if layer.Shape.Stroke.Taper.StartLength != nil {
					recordCapability("StrokeTaper.SetStartLength", taperPath+".start_length")
				}
				if layer.Shape.Stroke.Taper.EndLength != nil {
					recordCapability("StrokeTaper.SetEndLength", taperPath+".end_length")
				}
				if layer.Shape.Stroke.Taper.StartWidth != nil {
					recordCapability("StrokeTaper.SetStartWidth", taperPath+".start_width")
				}
				if layer.Shape.Stroke.Taper.EndWidth != nil {
					recordCapability("StrokeTaper.SetEndWidth", taperPath+".end_width")
				}
				if layer.Shape.Stroke.Taper.StartEase != nil {
					recordCapability("StrokeTaper.SetStartEase", taperPath+".start_ease")
				}
				if layer.Shape.Stroke.Taper.EndEase != nil {
					recordCapability("StrokeTaper.SetEndEase", taperPath+".end_ease")
				}
			}
			if layer.Shape.Stroke.Wave != nil {
				wavePath := strokePath + ".wave"
				if layer.Shape.Stroke.Wave.Amount != nil {
					recordCapability("StrokeWave.SetAmount", wavePath+".amount")
				}
				if layer.Shape.Stroke.Wave.Wavelength != nil {
					recordCapability("StrokeWave.SetWavelength", wavePath+".wavelength")
				}
				if layer.Shape.Stroke.Wave.Phase != nil {
					recordCapability("StrokeWave.SetPhase", wavePath+".phase")
				}
			}
			if layer.Shape.Stroke.Dashes != nil {
				dashesPath := strokePath + ".dashes"
				if layer.Shape.Stroke.Dashes.Dash != nil {
					recordCapability("StrokeDashes.SetDash", dashesPath+".dash")
					if *layer.Shape.Stroke.Dashes.Dash < 0 {
						addRefusal("invalid_shape_stroke_dash", dashesPath+".dash", "stroke dash must be non-negative")
					}
				}
				if layer.Shape.Stroke.Dashes.Gap != nil {
					recordCapability("StrokeDashes.SetGap", dashesPath+".gap")
					if *layer.Shape.Stroke.Dashes.Gap < 0 {
						addRefusal("invalid_shape_stroke_gap", dashesPath+".gap", "stroke gap must be non-negative")
					}
				}
			}
		}
		if layer.Shape.Trim != nil {
			trimPath := layerPath + ".shape.trim"
			recordCapability("VectorGroup.AddTrim", trimPath)
			if layer.Shape.Trim.Start != nil && (*layer.Shape.Trim.Start < 0 || *layer.Shape.Trim.Start > 100) {
				addRefusal("invalid_shape_trim_start", trimPath+".start", "trim start must be between 0 and 100")
			}
			if layer.Shape.Trim.End != nil && (*layer.Shape.Trim.End < 0 || *layer.Shape.Trim.End > 100) {
				addRefusal("invalid_shape_trim_end", trimPath+".end", "trim end must be between 0 and 100")
			}
		}
		if layer.Shape.RoundCorners != nil {
			roundPath := layerPath + ".shape.round_corners"
			recordCapability("VectorGroup.AddRoundCorners", roundPath)
			if layer.Shape.RoundCorners.Radius != nil {
				recordCapability("RoundCornersNode.SetRadius", roundPath+".radius")
				if *layer.Shape.RoundCorners.Radius < 0 {
					addRefusal("invalid_shape_round_corners_radius", roundPath+".radius", "round corners radius must be non-negative")
				}
			}
		}
		if layer.Shape.OffsetPaths != nil {
			offsetPath := layerPath + ".shape.offset_paths"
			recordCapability("VectorGroup.AddOffsetPaths", offsetPath)
			if layer.Shape.OffsetPaths.Amount != nil {
				recordCapability("OffsetPathsNode.SetAmount", offsetPath+".amount")
			}
			if layer.Shape.OffsetPaths.LineJoin != "" {
				recordCapability("OffsetPathsNode.SetLineJoin", offsetPath+".line_join")
				if !validOffsetLineJoin(layer.Shape.OffsetPaths.LineJoin) {
					addRefusal("invalid_shape_offset_line_join", offsetPath+".line_join", "offset line_join must be miter, round, or bevel")
				}
			}
			if layer.Shape.OffsetPaths.MiterLimit != nil {
				recordCapability("OffsetPathsNode.SetMiterLimit", offsetPath+".miter_limit")
				if *layer.Shape.OffsetPaths.MiterLimit < 1 {
					addRefusal("invalid_shape_offset_miter_limit", offsetPath+".miter_limit", "offset miter_limit must be at least 1")
				}
			}
			if layer.Shape.OffsetPaths.Copies != nil {
				recordCapability("OffsetPathsNode.SetCopies", offsetPath+".copies")
				if *layer.Shape.OffsetPaths.Copies < 1 {
					addRefusal("invalid_shape_offset_copies", offsetPath+".copies", "offset copies must be at least 1")
				}
			}
			if layer.Shape.OffsetPaths.CopyOffset != nil {
				recordCapability("OffsetPathsNode.SetCopyOffset", offsetPath+".copy_offset")
			}
		}
		if layer.Shape.Repeater != nil {
			repeaterPath := layerPath + ".shape.repeater"
			recordCapability("VectorGroup.AddRepeater", repeaterPath)
			if layer.Shape.Repeater.Copies != nil {
				recordCapability("RepeaterNode.SetCopies", repeaterPath+".copies")
				if *layer.Shape.Repeater.Copies < 1 {
					addRefusal("invalid_shape_repeater_copies", repeaterPath+".copies", "repeater copies must be at least 1")
				}
			}
			if layer.Shape.Repeater.Offset != nil {
				recordCapability("RepeaterNode.SetOffset", repeaterPath+".offset")
			}
			if layer.Shape.Repeater.Order != "" {
				recordCapability("RepeaterNode.SetOrder", repeaterPath+".order")
				if !validRepeaterOrder(layer.Shape.Repeater.Order) {
					addRefusal("invalid_shape_repeater_order", repeaterPath+".order", "repeater order must be below or above")
				}
			}
			if len(layer.Shape.Repeater.Anchor) > 0 {
				recordCapability("RepeaterTransform.SetAnchor", repeaterPath+".anchor")
				validateVec(layer.Shape.Repeater.Anchor, 2, repeaterPath+".anchor", addRefusal)
			}
			if len(layer.Shape.Repeater.Position) > 0 {
				recordCapability("RepeaterTransform.SetPosition", repeaterPath+".position")
				validateVec(layer.Shape.Repeater.Position, 2, repeaterPath+".position", addRefusal)
			}
			if len(layer.Shape.Repeater.Scale) > 0 {
				recordCapability("RepeaterTransform.SetScale", repeaterPath+".scale")
				validateVec(layer.Shape.Repeater.Scale, 2, repeaterPath+".scale", addRefusal)
			}
			if layer.Shape.Repeater.Rotation != nil {
				recordCapability("RepeaterTransform.SetRotation", repeaterPath+".rotation")
			}
			if layer.Shape.Repeater.StartOpacity != nil {
				recordCapability("RepeaterTransform.SetStartOpacity", repeaterPath+".start_opacity")
				if *layer.Shape.Repeater.StartOpacity < 0 || *layer.Shape.Repeater.StartOpacity > 100 {
					addRefusal("invalid_shape_repeater_start_opacity", repeaterPath+".start_opacity", "repeater start_opacity must be between 0 and 100")
				}
			}
			if layer.Shape.Repeater.EndOpacity != nil {
				recordCapability("RepeaterTransform.SetEndOpacity", repeaterPath+".end_opacity")
				if *layer.Shape.Repeater.EndOpacity < 0 || *layer.Shape.Repeater.EndOpacity > 100 {
					addRefusal("invalid_shape_repeater_end_opacity", repeaterPath+".end_opacity", "repeater end_opacity must be between 0 and 100")
				}
			}
		}
		if layer.Shape.MergePaths != nil {
			mergePath := layerPath + ".shape.merge_paths"
			recordCapability("VectorGroup.AddMergePaths", mergePath)
			if layer.Shape.MergePaths.Type != "" {
				recordCapability("MergePathsNode.SetType", mergePath+".type")
				if !validMergePathsType(layer.Shape.MergePaths.Type) {
					addRefusal("invalid_shape_merge_paths_type", mergePath+".type", "merge_paths type must be merge, add, subtract, intersect, or exclude")
				}
			}
		}
		if layer.Shape.ZigZag != nil {
			zigZagPath := layerPath + ".shape.zigzag"
			recordCapability("VectorGroup.AddZigZag", zigZagPath)
			if layer.Shape.ZigZag.Size != nil {
				recordCapability("ZigZagNode.SetSize", zigZagPath+".size")
				if *layer.Shape.ZigZag.Size < 0 {
					addRefusal("invalid_shape_zigzag_size", zigZagPath+".size", "zigzag size must be non-negative")
				}
			}
			if layer.Shape.ZigZag.Detail != nil {
				recordCapability("ZigZagNode.SetDetail", zigZagPath+".detail")
				if *layer.Shape.ZigZag.Detail < 0 {
					addRefusal("invalid_shape_zigzag_detail", zigZagPath+".detail", "zigzag detail must be non-negative")
				}
			}
			if layer.Shape.ZigZag.Points != "" {
				recordCapability("ZigZagNode.SetPoints", zigZagPath+".points")
				if !validZigZagPoints(layer.Shape.ZigZag.Points) {
					addRefusal("invalid_shape_zigzag_points", zigZagPath+".points", "zigzag points must be corner or smooth")
				}
			}
		}
		if layer.Shape.PuckerBloat != nil {
			puckerBloatPath := layerPath + ".shape.pucker_bloat"
			recordCapability("VectorGroup.AddPuckerBloat", puckerBloatPath)
			if layer.Shape.PuckerBloat.Amount != nil {
				recordCapability("PuckerBloatNode.SetAmount", puckerBloatPath+".amount")
			}
		}
		if layer.Shape.Twist != nil {
			twistPath := layerPath + ".shape.twist"
			recordCapability("VectorGroup.AddTwist", twistPath)
			if layer.Shape.Twist.Angle != nil {
				recordCapability("TwistNode.SetAngle", twistPath+".angle")
			}
			if len(layer.Shape.Twist.Center) > 0 {
				recordCapability("TwistNode.SetCenter", twistPath+".center")
				validateVec(layer.Shape.Twist.Center, 2, twistPath+".center", addRefusal)
			}
		}
		if layer.Shape.WigglePaths != nil {
			wigglePath := layerPath + ".shape.wiggle_paths"
			recordCapability("VectorGroup.AddWigglePaths", wigglePath)
			if layer.Shape.WigglePaths.Size != nil {
				recordCapability("WigglePathsNode.SetSize", wigglePath+".size")
			}
			if layer.Shape.WigglePaths.Detail != nil {
				recordCapability("WigglePathsNode.SetDetail", wigglePath+".detail")
			}
			if layer.Shape.WigglePaths.WigglesPerSecond != nil {
				recordCapability("WigglePathsNode.SetWigglesPerSecond", wigglePath+".wiggles_per_second")
			}
			if layer.Shape.WigglePaths.RandomSeed != nil {
				recordCapability("WigglePathsNode.SetRandomSeed", wigglePath+".random_seed")
			}
			if layer.Shape.WigglePaths.Points != "" {
				recordCapability("WigglePathsNode.SetPoints", wigglePath+".points")
				if !validRoughenPoints(layer.Shape.WigglePaths.Points) {
					addRefusal("invalid_shape_wiggle_paths_points", wigglePath+".points", "wiggle_paths points must be corner or smooth")
				}
			}
			if layer.Shape.WigglePaths.Correlation != nil {
				recordCapability("WigglePathsNode.SetCorrelation", wigglePath+".correlation")
				if *layer.Shape.WigglePaths.Correlation < 0 || *layer.Shape.WigglePaths.Correlation > 100 {
					addRefusal("invalid_shape_wiggle_paths_correlation", wigglePath+".correlation", "wiggle_paths correlation must be between 0 and 100")
				}
			}
			if layer.Shape.WigglePaths.TemporalPhase != nil {
				recordCapability("WigglePathsNode.SetTemporalPhase", wigglePath+".temporal_phase")
			}
			if layer.Shape.WigglePaths.SpatialPhase != nil {
				recordCapability("WigglePathsNode.SetSpatialPhase", wigglePath+".spatial_phase")
			}
		}
		if layer.Shape.WiggleTransform != nil {
			wigglePath := layerPath + ".shape.wiggle_transform"
			recordCapability("VectorGroup.AddWiggleTransform", wigglePath)
			if len(layer.Shape.WiggleTransform.Anchor) > 0 {
				recordCapability("WigglerTransform.SetAnchor", wigglePath+".anchor")
				validateVec(layer.Shape.WiggleTransform.Anchor, 2, wigglePath+".anchor", addRefusal)
			}
			if len(layer.Shape.WiggleTransform.Position) > 0 {
				recordCapability("WigglerTransform.SetPosition", wigglePath+".position")
				validateVec(layer.Shape.WiggleTransform.Position, 2, wigglePath+".position", addRefusal)
			}
			if len(layer.Shape.WiggleTransform.Scale) > 0 {
				recordCapability("WigglerTransform.SetScale", wigglePath+".scale")
				validateVec(layer.Shape.WiggleTransform.Scale, 2, wigglePath+".scale", addRefusal)
			}
			if layer.Shape.WiggleTransform.Rotation != nil {
				recordCapability("WigglerTransform.SetRotation", wigglePath+".rotation")
			}
			if layer.Shape.WiggleTransform.WigglesPerSecond != nil {
				recordCapability("WiggleTransformNode.SetWigglesPerSecond", wigglePath+".wiggles_per_second")
			}
			if layer.Shape.WiggleTransform.RandomSeed != nil {
				recordCapability("WiggleTransformNode.SetRandomSeed", wigglePath+".random_seed")
			}
			if layer.Shape.WiggleTransform.Correlation != nil {
				recordCapability("WiggleTransformNode.SetCorrelation", wigglePath+".correlation")
				if *layer.Shape.WiggleTransform.Correlation < 0 || *layer.Shape.WiggleTransform.Correlation > 100 {
					addRefusal("invalid_shape_wiggle_transform_correlation", wigglePath+".correlation", "wiggle_transform correlation must be between 0 and 100")
				}
			}
			if layer.Shape.WiggleTransform.TemporalPhase != nil {
				recordCapability("WiggleTransformNode.SetTemporalPhase", wigglePath+".temporal_phase")
			}
			if layer.Shape.WiggleTransform.SpatialPhase != nil {
				recordCapability("WiggleTransformNode.SetSpatialPhase", wigglePath+".spatial_phase")
			}
		}
	}
	for i, mask := range layer.Masks {
		maskPath := fmt.Sprintf("%s.masks[%d]", layerPath, i)
		recordCapability("AddMask", maskPath)
		if layer.Type == "camera" || layer.Type == "light" {
			addRefusal("mask_on_unsupported_layer_type", maskPath, "masks are not supported on camera or light layers")
		}
		if mask.Mode != "" {
			recordCapability("Mask.SetMode", maskPath+".mode")
			if !validMaskMode(mask.Mode) {
				addRefusal("invalid_mask_mode", maskPath+".mode", "mask mode must be add, subtract, intersect, lighten, darken, difference, or none")
			}
		}
		if mask.Inverted != nil {
			recordCapability("Mask.SetInverted", maskPath+".inverted")
		}
		if mask.Locked != nil {
			recordCapability("Mask.SetLocked", maskPath+".locked")
		}
		if len(mask.Color) > 0 {
			recordCapability("Mask.SetColor", maskPath+".color")
			validateRGBColor(mask.Color, maskPath+".color", "invalid_mask_color", addRefusal)
		}
		if mask.MotionBlur != "" {
			recordCapability("Mask.SetMaskMotionBlur", maskPath+".motion_blur")
			if !validMaskMotionBlur(mask.MotionBlur) {
				addRefusal("invalid_mask_motion_blur", maskPath+".motion_blur", "mask motion_blur must be same_as_layer, on, or off")
			}
		}
		if mask.FeatherFalloff != "" {
			recordCapability("Mask.SetFeatherFalloff", maskPath+".feather_falloff")
			if !validMaskFeatherFalloff(mask.FeatherFalloff) {
				addRefusal("invalid_mask_feather_falloff", maskPath+".feather_falloff", "mask feather_falloff must be smooth or linear")
			}
		}
		if mask.Opacity != nil {
			recordCapability("Mask.SetOpacity", maskPath+".opacity")
			if *mask.Opacity < 0 || *mask.Opacity > 1 {
				addRefusal("invalid_mask_opacity", maskPath+".opacity", "mask opacity must be between 0 and 1")
			}
		}
		if len(mask.Feather) > 0 {
			recordCapability("Mask.SetFeather", maskPath+".feather")
			validateMaskFeather(mask.Feather, maskPath+".feather", "invalid_mask_feather", addRefusal)
		}
		if mask.Expansion != nil {
			recordCapability("Mask.SetExpansion", maskPath+".expansion")
		}
		if len(mask.Vertices) < 3 {
			addRefusal("invalid_mask_vertices", maskPath+".vertices", "mask vertices must include at least 3 points")
		}
		for vi, vertex := range mask.Vertices {
			validateVec(vertex, 2, fmt.Sprintf("%s.vertices[%d]", maskPath, vi), addRefusal)
		}
		if len(mask.PathKeyframes) > 0 {
			recordCapability("SetMaskPathKeyframes", maskPath+".path_keyframes")
			if len(mask.PathKeyframes) < 2 {
				addRefusal("invalid_mask_path_keyframes", maskPath+".path_keyframes", "mask path_keyframes must include at least 2 keyframes")
			}
		}
		for ki, kf := range mask.PathKeyframes {
			kfPath := fmt.Sprintf("%s.path_keyframes[%d]", maskPath, ki)
			if kf.Time < 0 || kf.Time > compDuration {
				addRefusal("mask_path_keyframe_time_out_of_range", kfPath+".time", "mask path keyframe time must be within comp duration")
			}
			if ki > 0 && kf.Time < mask.PathKeyframes[ki-1].Time {
				addRefusal("mask_path_keyframes_not_sorted", kfPath+".time", "mask path keyframes must be sorted by time")
			}
			if len(kf.Vertices) < 3 {
				addRefusal("invalid_mask_path_keyframe_vertices", kfPath+".vertices", "mask path keyframe vertices must include at least 3 points")
			}
			for vi, vertex := range kf.Vertices {
				validateVec(vertex, 2, fmt.Sprintf("%s.vertices[%d]", kfPath, vi), addRefusal)
			}
		}
	}
	if usesTransform(layer.Transform) {
		recordCapability("SetLayerTransform", layerPath+".transform")
	}
	validateTransformExpressions(layer.Transform.Expressions, layerPath+".transform.expressions", recordCapability, addRefusal)
	validateVec(layer.Transform.Position, 2, layerPath+".transform.position", addRefusal)
	validateVec(layer.Transform.Scale, 2, layerPath+".transform.scale", addRefusal)
	validateVec(layer.Transform.AnchorPoint, 2, layerPath+".transform.anchor_point", addRefusal)
	for i, kf := range layer.Transform.PositionKeyframes {
		kfPath := fmt.Sprintf("%s.transform.position_keyframes[%d]", layerPath, i)
		if kf.Time < 0 || kf.Time > compDuration {
			addRefusal("keyframe_time_out_of_range", kfPath+".time", "keyframe time must be within comp duration")
		}
		if i > 0 && kf.Time < layer.Transform.PositionKeyframes[i-1].Time {
			addRefusal("keyframes_not_sorted", kfPath+".time", "keyframes must be sorted by time")
		}
		validateVec(kf.Value, 2, kfPath+".value", addRefusal)
		validateKeyframeEase(kf.InEase, kfPath+".in_ease", addRefusal)
		validateKeyframeEase(kf.OutEase, kfPath+".out_ease", addRefusal)
	}
	for i, kf := range layer.Transform.AnchorPointKeyframes {
		kfPath := fmt.Sprintf("%s.transform.anchor_point_keyframes[%d]", layerPath, i)
		if kf.Time < 0 || kf.Time > compDuration {
			addRefusal("keyframe_time_out_of_range", kfPath+".time", "keyframe time must be within comp duration")
		}
		if i > 0 && kf.Time < layer.Transform.AnchorPointKeyframes[i-1].Time {
			addRefusal("keyframes_not_sorted", kfPath+".time", "keyframes must be sorted by time")
		}
		validateVec(kf.Value, 2, kfPath+".value", addRefusal)
		validateKeyframeEase(kf.InEase, kfPath+".in_ease", addRefusal)
		validateKeyframeEase(kf.OutEase, kfPath+".out_ease", addRefusal)
	}
	for i, kf := range layer.Transform.ScaleKeyframes {
		kfPath := fmt.Sprintf("%s.transform.scale_keyframes[%d]", layerPath, i)
		if kf.Time < 0 || kf.Time > compDuration {
			addRefusal("keyframe_time_out_of_range", kfPath+".time", "keyframe time must be within comp duration")
		}
		if i > 0 && kf.Time < layer.Transform.ScaleKeyframes[i-1].Time {
			addRefusal("keyframes_not_sorted", kfPath+".time", "keyframes must be sorted by time")
		}
		validateVec(kf.Value, 2, kfPath+".value", addRefusal)
		validateKeyframeEase(kf.InEase, kfPath+".in_ease", addRefusal)
		validateKeyframeEase(kf.OutEase, kfPath+".out_ease", addRefusal)
	}
	for i, kf := range layer.Transform.RotationKeyframes {
		kfPath := fmt.Sprintf("%s.transform.rotation_keyframes[%d]", layerPath, i)
		if kf.Time < 0 || kf.Time > compDuration {
			addRefusal("keyframe_time_out_of_range", kfPath+".time", "keyframe time must be within comp duration")
		}
		if i > 0 && kf.Time < layer.Transform.RotationKeyframes[i-1].Time {
			addRefusal("keyframes_not_sorted", kfPath+".time", "keyframes must be sorted by time")
		}
		validateKeyframeEase(kf.InEase, kfPath+".in_ease", addRefusal)
		validateKeyframeEase(kf.OutEase, kfPath+".out_ease", addRefusal)
	}
	for i, kf := range layer.Transform.OpacityKeyframes {
		kfPath := fmt.Sprintf("%s.transform.opacity_keyframes[%d]", layerPath, i)
		if kf.Time < 0 || kf.Time > compDuration {
			addRefusal("keyframe_time_out_of_range", kfPath+".time", "keyframe time must be within comp duration")
		}
		if i > 0 && kf.Time < layer.Transform.OpacityKeyframes[i-1].Time {
			addRefusal("keyframes_not_sorted", kfPath+".time", "keyframes must be sorted by time")
		}
		if kf.Value < 0 || kf.Value > 100 {
			addRefusal("invalid_opacity_keyframe_value", kfPath+".value", "opacity keyframe value must be between 0 and 100")
		}
		validateKeyframeEase(kf.InEase, kfPath+".in_ease", addRefusal)
		validateKeyframeEase(kf.OutEase, kfPath+".out_ease", addRefusal)
	}
	for i, effect := range layer.Effects {
		effectPath := fmt.Sprintf("%s.effects[%d]", layerPath, i)
		lookup := recordCapability("AddEffect", effectPath)
		if effect.MatchName == "" {
			addRefusal("missing_effect_match_name", effectPath+".match_name", "effect match_name is required")
			continue
		}
		if lookup.Status == CapabilityUnsupported {
			addRefusal("unsupported_effect_api", effectPath, "AddEffect is not supported by the loaded capability index")
			continue
		}
		if !supportedEffect(effect.MatchName) {
			addRefusal("unsupported_effect", effectPath+".match_name", fmt.Sprintf("effect %q is not in SupportedEffects", effect.MatchName))
			continue
		}
		for pi, param := range effect.Params {
			paramPath := fmt.Sprintf("%s.params[%d]", effectPath, pi)
			if param.MatchName == "" {
				addRefusal("missing_effect_param_match_name", paramPath+".match_name", "effect param match_name is required")
			}
			if param.TargetLayer != "" {
				recordCapability("SetEffectLayerParam", paramPath+".target_layer")
				if param.Value != nil || len(param.Keyframes) > 0 || param.Expression != nil {
					addRefusal("invalid_effect_layer_param_combo", paramPath, "target_layer effect params cannot also set value, keyframes, or expression")
				}
			} else if len(param.Keyframes) > 0 {
				recordCapability("SetEffectParam", paramPath)
				validateEffectParamKeyframes(param.Keyframes, paramPath+".keyframes", recordCapability, addRefusal)
			} else if !validEffectParamValue(param.Value) {
				recordCapability("SetEffectParam", paramPath)
				addRefusal("unsupported_effect_param_value", paramPath+".value", "effect param value must be a number, boolean, or numeric array")
			} else {
				recordCapability("SetEffectParam", paramPath)
			}
			if param.EssentialGraphics != nil {
				recordCapability("AddEssentialProperty", paramPath+".essential_graphics")
				if param.Value == nil {
					addRefusal("missing_essential_graphics_param_value", paramPath+".value", "essential_graphics params require a static value in this recipe slice")
				}
				if param.TargetLayer != "" || len(param.Keyframes) > 0 || param.Expression != nil {
					addRefusal("invalid_essential_graphics_param_combo", paramPath+".essential_graphics", "essential_graphics params currently require a static value without target_layer, keyframes, or expression")
				}
			}
			if param.Expression != nil {
				recordCapability("Property.SetExpression", paramPath+".expression.source")
				if param.Expression.Source == "" {
					addRefusal("missing_effect_param_expression_source", paramPath+".expression.source", "effect param expression source is required")
				}
				if param.Expression.Enabled != nil {
					recordCapability("Property.SetExpressionEnabled", paramPath+".expression.enabled")
				}
			}
		}
	}
}

func validateEffectParamKeyframes(keyframes []ValueKeyframe, path string, recordCapability func(string, string) CapabilityLookup, addRefusal func(string, string, string)) {
	kind := "unknown"
	for i, kf := range keyframes {
		kfPath := fmt.Sprintf("%s[%d]", path, i)
		if kf.Time < 0 {
			addRefusal("keyframe_time_out_of_range", kfPath+".time", "keyframe time must be non-negative")
		}
		if i > 0 && kf.Time < keyframes[i-1].Time {
			addRefusal("keyframes_not_sorted", kfPath+".time", "keyframes must be sorted by time")
		}
		validateKeyframeEase(kf.InEase, kfPath+".in_ease", addRefusal)
		validateKeyframeEase(kf.OutEase, kfPath+".out_ease", addRefusal)

		if _, ok := numericValue(kf.Value); ok {
			if kind == "vector" {
				addRefusal("invalid_effect_param_keyframe_value", kfPath+".value", "effect param keyframe values must not mix scalar and vector values")
			}
			kind = "scalar"
			continue
		}
		values, ok := numericSliceValue(kf.Value)
		if !ok || len(values) < 2 || len(values) > 4 {
			addRefusal("invalid_effect_param_keyframe_value", kfPath+".value", "effect param keyframe value must be a number or a 2-, 3-, or 4-number array")
			continue
		}
		if kind == "scalar" {
			addRefusal("invalid_effect_param_keyframe_value", kfPath+".value", "effect param keyframe values must not mix scalar and vector values")
		}
		kind = "vector"
	}
	if len(keyframes) < 2 {
		addRefusal("invalid_effect_param_keyframes", path, "effect param keyframes must include at least 2 keyframes")
	}
	switch kind {
	case "scalar":
		recordCapability("AnimateEffectParam", path)
	case "vector":
		recordCapability("AnimateEffectParamVec", path)
	default:
		recordCapability("AnimateEffectParam", path)
	}
}
