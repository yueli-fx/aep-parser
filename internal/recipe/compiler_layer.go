package recipe

import (
	"fmt"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/projectindex"
)

func compileLayer(comp *aep.Composition, spec Layer, compSpec CompSpec) (*aep.Layer, error) {
	var layer *aep.Layer
	switch spec.Type {
	case "text":
		l, err := aep.NewTextLayer(comp, spec.Name)
		if err != nil {
			return nil, fmt.Errorf("recipe: text layer %q: %w", spec.Name, err)
		}
		if spec.Text != "" {
			if err := l.SetText(spec.Text); err != nil {
				return nil, fmt.Errorf("recipe: text layer %q set text: %w", spec.Name, err)
			}
		}
		if spec.TextStyle != nil {
			if err := applyTextStyle(l, *spec.TextStyle); err != nil {
				return nil, fmt.Errorf("recipe: text layer %q style: %w", spec.Name, err)
			}
		}
		if len(spec.TextAnimators) > 0 {
			if err := applyTextAnimators(l, spec.TextAnimators); err != nil {
				return nil, fmt.Errorf("recipe: text layer %q animators: %w", spec.Name, err)
			}
		}
		layer = l
	case "shape":
		l, err := aep.NewShapeLayer(comp, spec.Name)
		if err != nil {
			return nil, fmt.Errorf("recipe: shape layer %q: %w", spec.Name, err)
		}
		if spec.Shape != nil {
			if err := compileShape(l.RootGroup(), *spec.Shape); err != nil {
				return nil, fmt.Errorf("recipe: shape layer %q: %w", spec.Name, err)
			}
		}
		layer = l.Layer
	case "solid":
		color := [3]float64{0, 0, 0}
		if spec.Shape != nil && len(spec.Shape.FillColor) >= 3 {
			color = [3]float64{toUnitColor(spec.Shape.FillColor[0]), toUnitColor(spec.Shape.FillColor[1]), toUnitColor(spec.Shape.FillColor[2])}
		}
		l, err := aep.NewSolidLayer(comp, spec.Name, compSpec.Width, compSpec.Height, color)
		if err != nil {
			return nil, fmt.Errorf("recipe: solid layer %q: %w", spec.Name, err)
		}
		layer = l
	case "null":
		l, err := aep.NewNullLayer(comp, spec.Name)
		if err != nil {
			return nil, fmt.Errorf("recipe: null layer %q: %w", spec.Name, err)
		}
		layer = l
	case "adjustment":
		l, err := aep.NewAdjustmentLayer(comp, spec.Name)
		if err != nil {
			return nil, fmt.Errorf("recipe: adjustment layer %q: %w", spec.Name, err)
		}
		layer = l
	case "camera":
		l, err := aep.NewCameraLayer(comp, spec.Name)
		if err != nil {
			return nil, fmt.Errorf("recipe: camera layer %q: %w", spec.Name, err)
		}
		layer = l
	case "light":
		l, err := aep.NewLightLayer(comp, spec.Name)
		if err != nil {
			return nil, fmt.Errorf("recipe: light layer %q: %w", spec.Name, err)
		}
		layer = l
	default:
		return nil, fmt.Errorf("recipe: unsupported layer type %q", spec.Type)
	}
	if layer != nil {
		if spec.Label != nil {
			if err := layer.SetLabel(uint8(*spec.Label)); err != nil {
				return nil, fmt.Errorf("recipe: layer %q label: %w", spec.Name, err)
			}
		}
		if spec.Comment != "" {
			if err := layer.SetComment(spec.Comment); err != nil {
				return nil, fmt.Errorf("recipe: layer %q comment: %w", spec.Name, err)
			}
		}
		if spec.Visible != nil {
			if err := layer.SetVisible(*spec.Visible); err != nil {
				return nil, fmt.Errorf("recipe: layer %q visible: %w", spec.Name, err)
			}
		}
		if spec.Solo != nil {
			if err := layer.SetSolo(*spec.Solo); err != nil {
				return nil, fmt.Errorf("recipe: layer %q solo: %w", spec.Name, err)
			}
		}
		if spec.Locked != nil {
			if err := layer.SetLocked(*spec.Locked); err != nil {
				return nil, fmt.Errorf("recipe: layer %q locked: %w", spec.Name, err)
			}
		}
		if spec.MotionBlur != nil {
			if err := layer.SetMotionBlur(*spec.MotionBlur); err != nil {
				return nil, fmt.Errorf("recipe: layer %q motion_blur: %w", spec.Name, err)
			}
		}
		if spec.Shy != nil {
			if err := layer.SetShy(*spec.Shy); err != nil {
				return nil, fmt.Errorf("recipe: layer %q shy: %w", spec.Name, err)
			}
		}
		if spec.EffectsEnabled != nil {
			if err := layer.SetEffectsEnabled(*spec.EffectsEnabled); err != nil {
				return nil, fmt.Errorf("recipe: layer %q effects_enabled: %w", spec.Name, err)
			}
		}
		if spec.AudioEnabled != nil {
			if err := layer.SetAudioEnabled(*spec.AudioEnabled); err != nil {
				return nil, fmt.Errorf("recipe: layer %q audio_enabled: %w", spec.Name, err)
			}
		}
		if spec.FrameBlendEnabled != nil {
			if err := layer.SetFrameBlendEnabled(*spec.FrameBlendEnabled); err != nil {
				return nil, fmt.Errorf("recipe: layer %q frame_blend_enabled: %w", spec.Name, err)
			}
		}
		if spec.MarkersLocked != nil {
			if err := layer.SetMarkersLocked(*spec.MarkersLocked); err != nil {
				return nil, fmt.Errorf("recipe: layer %q markers_locked: %w", spec.Name, err)
			}
		}
		if spec.CollapseTransform != nil {
			if err := layer.SetCollapseTransform(*spec.CollapseTransform); err != nil {
				return nil, fmt.Errorf("recipe: layer %q collapse_transform: %w", spec.Name, err)
			}
		}
		if spec.Is3D != nil {
			if err := layer.SetIs3D(*spec.Is3D); err != nil {
				return nil, fmt.Errorf("recipe: layer %q is_3d: %w", spec.Name, err)
			}
		}
		if spec.IsAdjust != nil {
			if err := layer.SetIsAdjust(*spec.IsAdjust); err != nil {
				return nil, fmt.Errorf("recipe: layer %q is_adjust: %w", spec.Name, err)
			}
		}
		if spec.IsNull != nil {
			if err := layer.SetIsNull(*spec.IsNull); err != nil {
				return nil, fmt.Errorf("recipe: layer %q is_null: %w", spec.Name, err)
			}
		}
		if spec.IsGuide != nil {
			if err := layer.SetIsGuide(*spec.IsGuide); err != nil {
				return nil, fmt.Errorf("recipe: layer %q is_guide: %w", spec.Name, err)
			}
		}
		if spec.SamplingBicubic != nil {
			if err := layer.SetSamplingBicubic(*spec.SamplingBicubic); err != nil {
				return nil, fmt.Errorf("recipe: layer %q sampling_bicubic: %w", spec.Name, err)
			}
		}
		if spec.FrameBlendPixelMotion != nil {
			if err := layer.SetFrameBlendPixelMotion(*spec.FrameBlendPixelMotion); err != nil {
				return nil, fmt.Errorf("recipe: layer %q frame_blend_pixel_motion: %w", spec.Name, err)
			}
		}
		if spec.PreserveTransparency != nil {
			if err := layer.SetPreserveTransparency(*spec.PreserveTransparency); err != nil {
				return nil, fmt.Errorf("recipe: layer %q preserve_transparency: %w", spec.Name, err)
			}
		}
		if spec.Quality != "" {
			quality, err := layerQuality(spec.Quality)
			if err != nil {
				return nil, fmt.Errorf("recipe: layer %q quality: %w", spec.Name, err)
			}
			if err := layer.SetQuality(quality); err != nil {
				return nil, fmt.Errorf("recipe: layer %q quality: %w", spec.Name, err)
			}
		}
		if spec.BlendingMode != "" {
			blendingMode, err := layerBlendingMode(spec.BlendingMode)
			if err != nil {
				return nil, fmt.Errorf("recipe: layer %q blending_mode: %w", spec.Name, err)
			}
			if err := layer.SetBlendingMode(blendingMode); err != nil {
				return nil, fmt.Errorf("recipe: layer %q blending_mode: %w", spec.Name, err)
			}
		}
		if spec.TrackMatte != "" {
			trackMatte, err := layerTrackMatte(spec.TrackMatte)
			if err != nil {
				return nil, fmt.Errorf("recipe: layer %q track_matte: %w", spec.Name, err)
			}
			if err := layer.SetTrackMatte(trackMatte); err != nil {
				return nil, fmt.Errorf("recipe: layer %q track_matte: %w", spec.Name, err)
			}
		}
		if spec.AutoOrient != "" {
			autoOrient, err := layerAutoOrient(spec.AutoOrient)
			if err != nil {
				return nil, fmt.Errorf("recipe: layer %q auto_orient: %w", spec.Name, err)
			}
			if err := layer.SetAutoOrient(autoOrient); err != nil {
				return nil, fmt.Errorf("recipe: layer %q auto_orient: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.Zoom != nil {
			if err := layer.SetCameraZoom(*spec.Camera.Zoom); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.zoom: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.DepthOfField != nil {
			if err := layer.SetCameraDepthOfField(*spec.Camera.DepthOfField); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.depth_of_field: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.FocusDistance != nil {
			if err := layer.SetCameraFocusDistance(*spec.Camera.FocusDistance); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.focus_distance: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.Aperture != nil {
			if err := layer.SetCameraAperture(*spec.Camera.Aperture); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.aperture: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.BlurLevel != nil {
			if err := layer.SetCameraBlurLevel(*spec.Camera.BlurLevel); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.blur_level: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.IrisShape != nil {
			if err := layer.SetIrisShape(*spec.Camera.IrisShape); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.iris_shape: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.IrisRotation != nil {
			if err := layer.SetIrisRotation(*spec.Camera.IrisRotation); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.iris_rotation: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.IrisRoundness != nil {
			if err := layer.SetIrisRoundness(*spec.Camera.IrisRoundness); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.iris_roundness: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.IrisAspectRatio != nil {
			if err := layer.SetIrisAspectRatio(*spec.Camera.IrisAspectRatio); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.iris_aspect_ratio: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.IrisDiffractionFringe != nil {
			if err := layer.SetIrisDiffractionFringe(*spec.Camera.IrisDiffractionFringe); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.iris_diffraction_fringe: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.IrisHighlightGain != nil {
			if err := layer.SetIrisHighlightGain(*spec.Camera.IrisHighlightGain); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.iris_highlight_gain: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.IrisHighlightThreshold != nil {
			if err := layer.SetIrisHighlightThreshold(*spec.Camera.IrisHighlightThreshold); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.iris_highlight_threshold: %w", spec.Name, err)
			}
		}
		if spec.Camera != nil && spec.Camera.IrisHighlightSaturation != nil {
			if err := layer.SetIrisHighlightSaturation(*spec.Camera.IrisHighlightSaturation); err != nil {
				return nil, fmt.Errorf("recipe: layer %q camera.iris_highlight_saturation: %w", spec.Name, err)
			}
		}
		if spec.Light != nil && spec.Light.Kind != "" {
			kind, err := lightKind(spec.Light.Kind)
			if err != nil {
				return nil, fmt.Errorf("recipe: layer %q light.kind: %w", spec.Name, err)
			}
			if err := layer.SetLightKind(kind); err != nil {
				return nil, fmt.Errorf("recipe: layer %q light.kind: %w", spec.Name, err)
			}
		}
		if spec.Light != nil && spec.Light.Intensity != nil {
			if err := layer.SetLightIntensity(*spec.Light.Intensity); err != nil {
				return nil, fmt.Errorf("recipe: layer %q light.intensity: %w", spec.Name, err)
			}
		}
		if spec.Light != nil && len(spec.Light.Color) > 0 {
			if err := layer.SetLightColor(spec.Light.Color); err != nil {
				return nil, fmt.Errorf("recipe: layer %q light.color: %w", spec.Name, err)
			}
		}
		if spec.Light != nil && spec.Light.CastsShadows != nil {
			if err := layer.SetLightCastsShadows(*spec.Light.CastsShadows); err != nil {
				return nil, fmt.Errorf("recipe: layer %q light.casts_shadows: %w", spec.Name, err)
			}
		}
		if spec.Light != nil && spec.Light.ShadowDarkness != nil {
			if err := layer.SetLightShadowDarkness(*spec.Light.ShadowDarkness); err != nil {
				return nil, fmt.Errorf("recipe: layer %q light.shadow_darkness: %w", spec.Name, err)
			}
		}
		if spec.Light != nil && spec.Light.ShadowDiffusion != nil {
			if err := layer.SetLightShadowDiffusion(*spec.Light.ShadowDiffusion); err != nil {
				return nil, fmt.Errorf("recipe: layer %q light.shadow_diffusion: %w", spec.Name, err)
			}
		}
		if spec.Light != nil && spec.Light.FalloffType != nil {
			if err := layer.SetLightFalloffType(*spec.Light.FalloffType); err != nil {
				return nil, fmt.Errorf("recipe: layer %q light.falloff_type: %w", spec.Name, err)
			}
		}
		if spec.Light != nil && spec.Light.FalloffStart != nil {
			if err := layer.SetLightFalloffStart(*spec.Light.FalloffStart); err != nil {
				return nil, fmt.Errorf("recipe: layer %q light.falloff_start: %w", spec.Name, err)
			}
		}
		if spec.Light != nil && spec.Light.FalloffDistance != nil {
			if err := layer.SetLightFalloffDistance(*spec.Light.FalloffDistance); err != nil {
				return nil, fmt.Errorf("recipe: layer %q light.falloff_distance: %w", spec.Name, err)
			}
		}
		if spec.Light != nil && spec.Light.ConeAngle != nil {
			if err := layer.SetLightConeAngle(*spec.Light.ConeAngle); err != nil {
				return nil, fmt.Errorf("recipe: layer %q light.cone_angle: %w", spec.Name, err)
			}
		}
		if spec.Light != nil && spec.Light.ConeFeather != nil {
			if err := layer.SetLightConeFeather(*spec.Light.ConeFeather); err != nil {
				return nil, fmt.Errorf("recipe: layer %q light.cone_feather: %w", spec.Name, err)
			}
		}
		if spec.StartTime != nil {
			if err := layer.SetStartTime(*spec.StartTime); err != nil {
				return nil, fmt.Errorf("recipe: layer %q start_time: %w", spec.Name, err)
			}
		}
		if spec.InPoint != nil {
			if err := layer.SetInPoint(*spec.InPoint); err != nil {
				return nil, fmt.Errorf("recipe: layer %q in_point: %w", spec.Name, err)
			}
		}
		if spec.OutPoint != nil {
			if err := layer.SetOutPoint(*spec.OutPoint); err != nil {
				return nil, fmt.Errorf("recipe: layer %q out_point: %w", spec.Name, err)
			}
		}
		if spec.Stretch != nil {
			if err := layer.SetStretch(*spec.Stretch); err != nil {
				return nil, fmt.Errorf("recipe: layer %q stretch: %w", spec.Name, err)
			}
		}
		if err := applyTransform(layer, spec.Transform); err != nil {
			return nil, fmt.Errorf("recipe: layer %q transform: %w", spec.Name, err)
		}
	}
	return layer, nil
}

func applyLayerParents(compSpec CompSpec, idx *projectindex.Index) error {
	for _, layerSpec := range compSpec.Layers {
		if layerSpec.Parent == "" {
			continue
		}
		layer := recipeLayerByName(idx, layerSpec.Name)
		parent := recipeLayerByName(idx, layerSpec.Parent)
		if layer == nil || parent == nil {
			return fmt.Errorf("recipe: layer %q parent %q not found", layerSpec.Name, layerSpec.Parent)
		}
		if err := layer.SetParent(parent.ID); err != nil {
			return fmt.Errorf("recipe: layer %q parent: %w", layerSpec.Name, err)
		}
	}
	return nil
}

func applyExplicitMattes(compSpec CompSpec, idx *projectindex.Index) error {
	for _, layerSpec := range compSpec.Layers {
		if layerSpec.Matte == "" {
			continue
		}
		layer := recipeLayerByName(idx, layerSpec.Name)
		matte := recipeLayerByName(idx, layerSpec.Matte)
		if layer == nil || matte == nil {
			return fmt.Errorf("recipe: layer %q matte %q not found", layerSpec.Name, layerSpec.Matte)
		}
		mode, err := layerTrackMatte(layerSpec.TrackMatte)
		if err != nil {
			return fmt.Errorf("recipe: layer %q matte track_matte: %w", layerSpec.Name, err)
		}
		if err := layer.SetTrackMatteSource(matte, mode); err != nil {
			return fmt.Errorf("recipe: layer %q matte: %w", layerSpec.Name, err)
		}
	}
	return nil
}

func applyLightSources(compSpec CompSpec, idx *projectindex.Index) error {
	for _, layerSpec := range compSpec.Layers {
		if layerSpec.Light == nil || layerSpec.Light.SourceLayer == "" {
			continue
		}
		layer := recipeLayerByName(idx, layerSpec.Name)
		target := recipeLayerByName(idx, layerSpec.Light.SourceLayer)
		if layer == nil || target == nil {
			return fmt.Errorf("recipe: layer %q light.source_layer %q not found", layerSpec.Name, layerSpec.Light.SourceLayer)
		}
		if err := layer.SetLightSource(target); err != nil {
			return fmt.Errorf("recipe: layer %q light.source_layer: %w", layerSpec.Name, err)
		}
	}
	return nil
}

func recipeLayerByName(idx *projectindex.Index, name string) *aep.Layer {
	layers := idx.LayersByName(name)
	if len(layers) == 0 {
		return nil
	}
	return layers[len(layers)-1]
}

func applyTextStyle(layer *aep.Layer, spec TextStyleSpec) error {
	if spec.FontSize != nil {
		if err := layer.SetRunFontSize(spec.RunIndex, *spec.FontSize); err != nil {
			return err
		}
	}
	if len(spec.FillColor) >= 3 {
		if err := layer.SetRunFillColor(spec.RunIndex, rgbaColor(spec.FillColor)); err != nil {
			return err
		}
	}
	if spec.AutoLeading != nil {
		if err := layer.SetRunAutoLeading(spec.RunIndex, *spec.AutoLeading); err != nil {
			return err
		}
	}
	if spec.Leading != nil {
		if err := layer.SetRunLeading(spec.RunIndex, *spec.Leading); err != nil {
			return err
		}
	}
	if spec.Tracking != nil {
		if err := layer.SetRunTracking(spec.RunIndex, *spec.Tracking); err != nil {
			return err
		}
	}
	if spec.BaselineShift != nil {
		if err := layer.SetRunBaselineShift(spec.RunIndex, *spec.BaselineShift); err != nil {
			return err
		}
	}
	if spec.HorizontalScale != nil {
		if err := layer.SetRunHorizontalScale(spec.RunIndex, *spec.HorizontalScale); err != nil {
			return err
		}
	}
	if spec.VerticalScale != nil {
		if err := layer.SetRunVerticalScale(spec.RunIndex, *spec.VerticalScale); err != nil {
			return err
		}
	}
	if spec.Tsume != nil {
		if err := layer.SetRunTsume(spec.RunIndex, *spec.Tsume); err != nil {
			return err
		}
	}
	if spec.CapsOption != "" {
		value, err := textCapsOption(spec.CapsOption)
		if err != nil {
			return err
		}
		if err := layer.SetRunCapsOption(spec.RunIndex, value); err != nil {
			return err
		}
	}
	if spec.BaselineOption != "" {
		value, err := textBaselineOption(spec.BaselineOption)
		if err != nil {
			return err
		}
		if err := layer.SetRunBaselineOption(spec.RunIndex, value); err != nil {
			return err
		}
	}
	if spec.AutoKernType != "" {
		value, err := textAutoKernType(spec.AutoKernType)
		if err != nil {
			return err
		}
		if err := layer.SetRunAutoKernType(spec.RunIndex, value); err != nil {
			return err
		}
	}
	if spec.LineJoinType != "" {
		value, err := textLineJoinType(spec.LineJoinType)
		if err != nil {
			return err
		}
		if err := layer.SetRunLineJoinType(spec.RunIndex, value); err != nil {
			return err
		}
	}
	if spec.DigitSet != "" {
		value, err := textDigitSet(spec.DigitSet)
		if err != nil {
			return err
		}
		if err := layer.SetRunDigitSet(spec.RunIndex, value); err != nil {
			return err
		}
	}
	if spec.NoBreak != nil {
		if err := layer.SetRunNoBreak(spec.RunIndex, *spec.NoBreak); err != nil {
			return err
		}
	}
	if spec.FauxBold != nil {
		if err := layer.SetRunFauxBold(spec.RunIndex, *spec.FauxBold); err != nil {
			return err
		}
	}
	if spec.FauxItalic != nil {
		if err := layer.SetRunFauxItalic(spec.RunIndex, *spec.FauxItalic); err != nil {
			return err
		}
	}
	if spec.ApplyStroke != nil {
		if err := layer.SetRunApplyStroke(spec.RunIndex, *spec.ApplyStroke); err != nil {
			return err
		}
	}
	if len(spec.StrokeColor) >= 3 {
		if err := layer.SetRunStrokeColor(spec.RunIndex, rgbaColor(spec.StrokeColor)); err != nil {
			return err
		}
	}
	if spec.StrokeWidth != nil {
		if err := layer.SetRunStrokeWidth(spec.RunIndex, *spec.StrokeWidth); err != nil {
			return err
		}
	}
	if spec.StrokeOverFill != nil {
		if err := layer.SetRunStrokeOverFill(spec.RunIndex, *spec.StrokeOverFill); err != nil {
			return err
		}
	}
	if spec.Justification != "" {
		justification, err := textJustification(spec.Justification)
		if err != nil {
			return err
		}
		if err := layer.SetParagraphJustification(spec.ParagraphIndex, justification); err != nil {
			return err
		}
	}
	if spec.FirstLineIndent != nil {
		if err := layer.SetParagraphFirstLineIndent(spec.ParagraphIndex, *spec.FirstLineIndent); err != nil {
			return err
		}
	}
	if spec.StartIndent != nil {
		if err := layer.SetParagraphStartIndent(spec.ParagraphIndex, *spec.StartIndent); err != nil {
			return err
		}
	}
	if spec.EndIndent != nil {
		if err := layer.SetParagraphEndIndent(spec.ParagraphIndex, *spec.EndIndent); err != nil {
			return err
		}
	}
	if spec.SpaceBefore != nil {
		if err := layer.SetParagraphSpaceBefore(spec.ParagraphIndex, *spec.SpaceBefore); err != nil {
			return err
		}
	}
	if spec.SpaceAfter != nil {
		if err := layer.SetParagraphSpaceAfter(spec.ParagraphIndex, *spec.SpaceAfter); err != nil {
			return err
		}
	}
	if spec.AutoHyphenate != nil {
		if err := layer.SetParagraphAutoHyphenate(spec.ParagraphIndex, *spec.AutoHyphenate); err != nil {
			return err
		}
	}
	if spec.LeadingType != "" {
		value, err := textLeadingType(spec.LeadingType)
		if err != nil {
			return err
		}
		if err := layer.SetParagraphLeadingType(spec.ParagraphIndex, value); err != nil {
			return err
		}
	}
	if spec.HangingRoman != nil {
		if err := layer.SetParagraphHangingRoman(spec.ParagraphIndex, *spec.HangingRoman); err != nil {
			return err
		}
	}
	if spec.ParagraphDirection != "" {
		value, err := textParagraphDirection(spec.ParagraphDirection)
		if err != nil {
			return err
		}
		if err := layer.SetParagraphDirection(spec.ParagraphIndex, value); err != nil {
			return err
		}
	}
	return nil
}

func applyTextAnimators(layer *aep.Layer, animators []TextAnimatorSpec) error {
	for i, animator := range animators {
		switch animator.Property {
		case "opacity":
			value, ok := animator.Value.(float64)
			if !ok {
				return fmt.Errorf("text_animators[%d].value must be a number", i)
			}
			if animator.RangeStart == nil || animator.RangeEnd == nil || animator.RangeOffset == nil {
				return fmt.Errorf("text_animators[%d] range_start, range_end, and range_offset are required", i)
			}
			if _, err := aep.AddTextOpacityAnimator(layer, value, *animator.RangeStart, *animator.RangeEnd, *animator.RangeOffset); err != nil {
				return err
			}
			if err := applyTextRangeOffsetKeyframes(layer, animator); err != nil {
				return err
			}
			if err := applyTextValueKeyframes(layer, animator); err != nil {
				return err
			}
		case "position":
			value, ok := numericSliceValue(animator.Value)
			if !ok || len(value) != 3 {
				return fmt.Errorf("text_animators[%d].value must be a 3-number array", i)
			}
			if animator.RangeStart == nil || animator.RangeEnd == nil || animator.RangeOffset == nil {
				return fmt.Errorf("text_animators[%d] range_start, range_end, and range_offset are required", i)
			}
			if _, err := aep.AddTextPositionAnimator(layer, value[0], value[1], value[2], *animator.RangeStart, *animator.RangeEnd, *animator.RangeOffset); err != nil {
				return err
			}
			if err := applyTextRangeOffsetKeyframes(layer, animator); err != nil {
				return err
			}
			if err := applyTextValueKeyframes(layer, animator); err != nil {
				return err
			}
		case "scale":
			value, ok := numericSliceValue(animator.Value)
			if !ok || len(value) != 3 {
				return fmt.Errorf("text_animators[%d].value must be a 3-number array", i)
			}
			if animator.RangeStart == nil || animator.RangeEnd == nil || animator.RangeOffset == nil {
				return fmt.Errorf("text_animators[%d] range_start, range_end, and range_offset are required", i)
			}
			if _, err := aep.AddTextScaleAnimator(layer, value[0], value[1], value[2], *animator.RangeStart, *animator.RangeEnd, *animator.RangeOffset); err != nil {
				return err
			}
			if err := applyTextRangeOffsetKeyframes(layer, animator); err != nil {
				return err
			}
			if err := applyTextValueKeyframes(layer, animator); err != nil {
				return err
			}
		case "rotation":
			value, ok := animator.Value.(float64)
			if !ok {
				return fmt.Errorf("text_animators[%d].value must be a number", i)
			}
			if animator.RangeStart == nil || animator.RangeEnd == nil || animator.RangeOffset == nil {
				return fmt.Errorf("text_animators[%d] range_start, range_end, and range_offset are required", i)
			}
			if _, err := aep.AddTextRotationAnimator(layer, value, *animator.RangeStart, *animator.RangeEnd, *animator.RangeOffset); err != nil {
				return err
			}
			if err := applyTextRangeOffsetKeyframes(layer, animator); err != nil {
				return err
			}
			if err := applyTextValueKeyframes(layer, animator); err != nil {
				return err
			}
		case "color":
			value, ok := numericSliceValue(animator.Value)
			if !ok || (len(value) != 3 && len(value) != 4) {
				return fmt.Errorf("text_animators[%d].value must be a 3- or 4-number array", i)
			}
			if animator.RangeStart == nil || animator.RangeEnd == nil || animator.RangeOffset == nil {
				return fmt.Errorf("text_animators[%d] range_start, range_end, and range_offset are required", i)
			}
			color := rgbaColor(value)
			if _, err := aep.AddTextColorAnimator(layer, color[0], color[1], color[2], color[3], *animator.RangeStart, *animator.RangeEnd, *animator.RangeOffset); err != nil {
				return err
			}
			if err := applyTextRangeOffsetKeyframes(layer, animator); err != nil {
				return err
			}
			if err := applyTextValueKeyframes(layer, animator); err != nil {
				return err
			}
		case "stroke_color":
			value, ok := numericSliceValue(animator.Value)
			if !ok || (len(value) != 3 && len(value) != 4) {
				return fmt.Errorf("text_animators[%d].value must be a 3- or 4-number array", i)
			}
			if animator.RangeStart == nil || animator.RangeEnd == nil || animator.RangeOffset == nil {
				return fmt.Errorf("text_animators[%d] range_start, range_end, and range_offset are required", i)
			}
			color := rgbaColor(value)
			if _, err := aep.AddTextStrokeColorAnimator(layer, color[0], color[1], color[2], color[3], *animator.RangeStart, *animator.RangeEnd, *animator.RangeOffset); err != nil {
				return err
			}
			if err := applyTextRangeOffsetKeyframes(layer, animator); err != nil {
				return err
			}
		case "tracking":
			value, ok := animator.Value.(float64)
			if !ok {
				return fmt.Errorf("text_animators[%d].value must be a number", i)
			}
			if animator.RangeStart == nil || animator.RangeEnd == nil || animator.RangeOffset == nil {
				return fmt.Errorf("text_animators[%d] range_start, range_end, and range_offset are required", i)
			}
			if _, err := aep.AddTextTrackingAnimator(layer, value, *animator.RangeStart, *animator.RangeEnd, *animator.RangeOffset); err != nil {
				return err
			}
			if err := applyTextRangeOffsetKeyframes(layer, animator); err != nil {
				return err
			}
			if err := applyTextValueKeyframes(layer, animator); err != nil {
				return err
			}
		case "character_offset":
			value, ok := animator.Value.(float64)
			if !ok {
				return fmt.Errorf("text_animators[%d].value must be a number", i)
			}
			if animator.RangeStart == nil || animator.RangeEnd == nil || animator.RangeOffset == nil {
				return fmt.Errorf("text_animators[%d] range_start, range_end, and range_offset are required", i)
			}
			if _, err := aep.AddTextCharacterOffsetAnimator(layer, value, *animator.RangeStart, *animator.RangeEnd, *animator.RangeOffset); err != nil {
				return err
			}
			if err := applyTextRangeOffsetKeyframes(layer, animator); err != nil {
				return err
			}
			if err := applyTextValueKeyframes(layer, animator); err != nil {
				return err
			}
		case "fill_opacity":
			value, ok := animator.Value.(float64)
			if !ok {
				return fmt.Errorf("text_animators[%d].value must be a number", i)
			}
			if animator.RangeStart == nil || animator.RangeEnd == nil || animator.RangeOffset == nil {
				return fmt.Errorf("text_animators[%d] range_start, range_end, and range_offset are required", i)
			}
			if _, err := aep.AddTextFillOpacityAnimator(layer, value, *animator.RangeStart, *animator.RangeEnd, *animator.RangeOffset); err != nil {
				return err
			}
			if err := applyTextRangeOffsetKeyframes(layer, animator); err != nil {
				return err
			}
		case "stroke_opacity":
			value, ok := animator.Value.(float64)
			if !ok {
				return fmt.Errorf("text_animators[%d].value must be a number", i)
			}
			if animator.RangeStart == nil || animator.RangeEnd == nil || animator.RangeOffset == nil {
				return fmt.Errorf("text_animators[%d] range_start, range_end, and range_offset are required", i)
			}
			if _, err := aep.AddTextStrokeOpacityAnimator(layer, value, *animator.RangeStart, *animator.RangeEnd, *animator.RangeOffset); err != nil {
				return err
			}
			if err := applyTextRangeOffsetKeyframes(layer, animator); err != nil {
				return err
			}
		case "stroke_width":
			value, ok := animator.Value.(float64)
			if !ok {
				return fmt.Errorf("text_animators[%d].value must be a number", i)
			}
			if animator.RangeStart == nil || animator.RangeEnd == nil || animator.RangeOffset == nil {
				return fmt.Errorf("text_animators[%d] range_start, range_end, and range_offset are required", i)
			}
			if _, err := aep.AddTextStrokeWidthAnimator(layer, value, *animator.RangeStart, *animator.RangeEnd, *animator.RangeOffset); err != nil {
				return err
			}
			if err := applyTextRangeOffsetKeyframes(layer, animator); err != nil {
				return err
			}
		case "skew":
			value, ok := animator.Value.(float64)
			if !ok {
				return fmt.Errorf("text_animators[%d].value must be a number", i)
			}
			if animator.RangeStart == nil || animator.RangeEnd == nil || animator.RangeOffset == nil {
				return fmt.Errorf("text_animators[%d] range_start, range_end, and range_offset are required", i)
			}
			if _, err := aep.AddTextSkewAnimator(layer, value, *animator.RangeStart, *animator.RangeEnd, *animator.RangeOffset); err != nil {
				return err
			}
			if err := applyTextRangeOffsetKeyframes(layer, animator); err != nil {
				return err
			}
		case "rotation_x":
			value, ok := animator.Value.(float64)
			if !ok {
				return fmt.Errorf("text_animators[%d].value must be a number", i)
			}
			if animator.RangeStart == nil || animator.RangeEnd == nil || animator.RangeOffset == nil {
				return fmt.Errorf("text_animators[%d] range_start, range_end, and range_offset are required", i)
			}
			if _, err := aep.AddTextRotationXAnimator(layer, value, *animator.RangeStart, *animator.RangeEnd, *animator.RangeOffset); err != nil {
				return err
			}
			if err := applyTextRangeOffsetKeyframes(layer, animator); err != nil {
				return err
			}
		case "rotation_y":
			value, ok := animator.Value.(float64)
			if !ok {
				return fmt.Errorf("text_animators[%d].value must be a number", i)
			}
			if animator.RangeStart == nil || animator.RangeEnd == nil || animator.RangeOffset == nil {
				return fmt.Errorf("text_animators[%d] range_start, range_end, and range_offset are required", i)
			}
			if _, err := aep.AddTextRotationYAnimator(layer, value, *animator.RangeStart, *animator.RangeEnd, *animator.RangeOffset); err != nil {
				return err
			}
			if err := applyTextRangeOffsetKeyframes(layer, animator); err != nil {
				return err
			}
		default:
			return fmt.Errorf("text_animators[%d].property %q is not supported", i, animator.Property)
		}
	}
	return nil
}

func applyTextRangeOffsetKeyframes(layer *aep.Layer, animator TextAnimatorSpec) error {
	if len(animator.RangeOffsetKeyframes) == 0 {
		return nil
	}
	keyframes := make([]aep.ScalarKeyframe, 0, len(animator.RangeOffsetKeyframes))
	for _, kf := range animator.RangeOffsetKeyframes {
		keyframes = append(keyframes, aep.ScalarKeyframe{
			Time:    kf.Time,
			Value:   kf.Value,
			InEase:  temporalEase(kf.InEase),
			OutEase: temporalEase(kf.OutEase),
		})
	}
	return aep.AnimateTextRangeOffset(layer, 0, keyframes)
}

func applyTextValueKeyframes(layer *aep.Layer, animator TextAnimatorSpec) error {
	if len(animator.ValueKeyframes) == 0 {
		return nil
	}
	switch animator.Property {
	case "opacity":
		keyframes, err := scalarValueKeyframes(animator.ValueKeyframes)
		if err != nil {
			return err
		}
		return aep.AnimateTextOpacity(layer, 0, keyframes)
	case "position":
		keyframes, err := vectorValueKeyframes(animator.ValueKeyframes, 3)
		if err != nil {
			return err
		}
		return aep.AnimateTextPosition(layer, 0, keyframes)
	case "scale":
		keyframes, err := vectorValueKeyframes(animator.ValueKeyframes, 3)
		if err != nil {
			return err
		}
		return aep.AnimateTextScale(layer, 0, keyframes)
	case "rotation":
		keyframes, err := scalarValueKeyframes(animator.ValueKeyframes)
		if err != nil {
			return err
		}
		return aep.AnimateTextRotation(layer, 0, keyframes)
	case "color":
		keyframes, err := colorValueKeyframes(animator.ValueKeyframes)
		if err != nil {
			return err
		}
		return aep.AnimateTextColor(layer, 0, keyframes)
	case "tracking":
		keyframes, err := scalarValueKeyframes(animator.ValueKeyframes)
		if err != nil {
			return err
		}
		return aep.AnimateTextTracking(layer, 0, keyframes)
	case "character_offset":
		keyframes, err := scalarValueKeyframes(animator.ValueKeyframes)
		if err != nil {
			return err
		}
		return aep.AnimateTextCharacterOffset(layer, 0, keyframes)
	default:
		return fmt.Errorf("text_animators[].value_keyframes are not supported for property %q", animator.Property)
	}
}

func scalarValueKeyframes(in []ValueKeyframe) ([]aep.ScalarKeyframe, error) {
	out := make([]aep.ScalarKeyframe, 0, len(in))
	for i, kf := range in {
		value, ok := numericValue(kf.Value)
		if !ok {
			return nil, fmt.Errorf("text_animators[].value_keyframes[%d].value must be a number", i)
		}
		out = append(out, aep.ScalarKeyframe{
			Time:    kf.Time,
			Value:   value,
			InEase:  temporalEase(kf.InEase),
			OutEase: temporalEase(kf.OutEase),
		})
	}
	return out, nil
}

func vectorValueKeyframes(in []ValueKeyframe, want int) ([]aep.VectorKeyframe, error) {
	out := make([]aep.VectorKeyframe, 0, len(in))
	for i, kf := range in {
		value, ok := numericSliceValue(kf.Value)
		if !ok || len(value) != want {
			return nil, fmt.Errorf("text_animators[].value_keyframes[%d].value must be a %d-number array", i, want)
		}
		out = append(out, aep.VectorKeyframe{
			Time:    kf.Time,
			Value:   value,
			InEase:  temporalEase(kf.InEase),
			OutEase: temporalEase(kf.OutEase),
		})
	}
	return out, nil
}

func colorValueKeyframes(in []ValueKeyframe) ([]aep.VectorKeyframe, error) {
	out := make([]aep.VectorKeyframe, 0, len(in))
	for i, kf := range in {
		value, ok := numericSliceValue(kf.Value)
		if !ok || (len(value) != 3 && len(value) != 4) {
			return nil, fmt.Errorf("text_animators[].value_keyframes[%d].value must be a 3- or 4-number color array", i)
		}
		color := rgbaColor(value)
		out = append(out, aep.VectorKeyframe{
			Time:    kf.Time,
			Value:   []float64{color[0], color[1], color[2], color[3]},
			InEase:  temporalEase(kf.InEase),
			OutEase: temporalEase(kf.OutEase),
		})
	}
	return out, nil
}

func textJustification(value string) (aep.TextJustification, error) {
	switch value {
	case "left":
		return aep.TextJustifyLeft, nil
	case "right":
		return aep.TextJustifyRight, nil
	case "center":
		return aep.TextJustifyCenter, nil
	default:
		return 0, fmt.Errorf("unsupported justification %q", value)
	}
}

func textCapsOption(value string) (aep.TextCapsOption, error) {
	switch normalizeEnum(value) {
	case "normal":
		return aep.TextCapsNormal, nil
	case "small_caps":
		return aep.TextCapsSmall, nil
	case "all_caps":
		return aep.TextCapsAll, nil
	case "all_small_caps":
		return aep.TextCapsAllSmall, nil
	default:
		return 0, fmt.Errorf("unsupported caps_option %q", value)
	}
}

func textBaselineOption(value string) (aep.TextBaselineOption, error) {
	switch normalizeEnum(value) {
	case "normal":
		return aep.TextBaselineNormal, nil
	case "superscript":
		return aep.TextBaselineSuperscript, nil
	case "subscript":
		return aep.TextBaselineSubscript, nil
	default:
		return 0, fmt.Errorf("unsupported baseline_option %q", value)
	}
}

func textAutoKernType(value string) (aep.TextAutoKernType, error) {
	switch normalizeEnum(value) {
	case "no_auto":
		return aep.TextAutoKernNoAuto, nil
	case "metric":
		return aep.TextAutoKernMetric, nil
	case "optical":
		return aep.TextAutoKernOptical, nil
	default:
		return 0, fmt.Errorf("unsupported auto_kern_type %q", value)
	}
}

func textLineJoinType(value string) (aep.TextLineJoinType, error) {
	switch normalizeEnum(value) {
	case "miter":
		return aep.TextLineJoinMiter, nil
	case "round":
		return aep.TextLineJoinRound, nil
	case "bevel":
		return aep.TextLineJoinBevel, nil
	default:
		return 0, fmt.Errorf("unsupported line_join_type %q", value)
	}
}

func textDigitSet(value string) (aep.TextDigitSet, error) {
	switch normalizeEnum(value) {
	case "default":
		return aep.TextDigitSetDefault, nil
	case "arabic":
		return aep.TextDigitSetArabic, nil
	case "hindi":
		return aep.TextDigitSetHindi, nil
	case "farsi":
		return aep.TextDigitSetFarsi, nil
	case "arabic_rtl":
		return aep.TextDigitSetArabicRTL, nil
	default:
		return 0, fmt.Errorf("unsupported digit_set %q", value)
	}
}

func textLeadingType(value string) (aep.TextLeadingType, error) {
	switch normalizeEnum(value) {
	case "roman":
		return aep.TextLeadingRoman, nil
	case "japanese":
		return aep.TextLeadingJapanese, nil
	default:
		return 0, fmt.Errorf("unsupported leading_type %q", value)
	}
}

func textParagraphDirection(value string) (aep.TextParagraphDirection, error) {
	switch normalizeEnum(value) {
	case "ltr":
		return aep.TextDirectionLeftToRight, nil
	case "rtl":
		return aep.TextDirectionRightToLeft, nil
	default:
		return 0, fmt.Errorf("unsupported paragraph_direction %q", value)
	}
}

func normalizeEnum(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.ReplaceAll(value, "-", "_")
}

func layerQuality(value string) (aep.LayerQuality, error) {
	switch value {
	case "wireframe":
		return aep.LayerQualityWireframe, nil
	case "draft":
		return aep.LayerQualityDraft, nil
	case "best":
		return aep.LayerQualityBest, nil
	default:
		return 0, fmt.Errorf("unsupported quality %q", value)
	}
}

func layerAutoOrient(value string) (aep.AutoOrientType, error) {
	switch value {
	case "none":
		return aep.AutoOrientNone, nil
	case "along_path":
		return aep.AutoOrientAlongPath, nil
	case "camera_or_point_of_interest":
		return aep.AutoOrientCameraOrPointOfInterest, nil
	case "characters_toward_camera":
		return aep.AutoOrientCharactersTowardCamera, nil
	default:
		return 0, fmt.Errorf("unsupported auto_orient %q", value)
	}
}

func layerTrackMatte(value string) (aep.TrackMatteType, error) {
	switch value {
	case "none":
		return aep.TrackMatteNone, nil
	case "alpha":
		return aep.TrackMatteAlpha, nil
	case "alpha_inverse":
		return aep.TrackMatteAlphaInverse, nil
	case "luma":
		return aep.TrackMatteLuma, nil
	case "luma_inverse":
		return aep.TrackMatteLumaInverse, nil
	default:
		return 0, fmt.Errorf("unsupported track_matte %q", value)
	}
}

func maskMode(value string) (aep.MaskMode, error) {
	switch value {
	case "none":
		return aep.MaskModeNone, nil
	case "add":
		return aep.MaskModeAdd, nil
	case "subtract":
		return aep.MaskModeSubtract, nil
	case "intersect":
		return aep.MaskModeIntersect, nil
	case "lighten":
		return aep.MaskModeLighten, nil
	case "darken":
		return aep.MaskModeDarken, nil
	case "difference":
		return aep.MaskModeDifference, nil
	default:
		return 0, fmt.Errorf("unsupported mask mode %q", value)
	}
}

func maskMotionBlur(value string) (aep.MaskMotionBlurMode, error) {
	switch value {
	case "same_as_layer":
		return aep.MaskMotionBlurSameAsLayer, nil
	case "on":
		return aep.MaskMotionBlurOn, nil
	case "off":
		return aep.MaskMotionBlurOff, nil
	default:
		return 0, fmt.Errorf("unsupported mask motion_blur %q", value)
	}
}

func maskFeatherFalloff(value string) (aep.MaskFeatherFalloff, error) {
	switch value {
	case "smooth":
		return aep.MaskFeatherFalloffSmooth, nil
	case "linear":
		return aep.MaskFeatherFalloffLinear, nil
	default:
		return 0, fmt.Errorf("unsupported mask feather_falloff %q", value)
	}
}

func layerBlendingMode(value string) (aep.BlendingMode, error) {
	switch value {
	case "normal_camera":
		return aep.BlendingModeNormalCamera, nil
	case "normal":
		return aep.BlendingModeNormal, nil
	case "dissolve":
		return aep.BlendingModeDissolve, nil
	case "add":
		return aep.BlendingModeAdd, nil
	case "multiply":
		return aep.BlendingModeMultiply, nil
	case "screen":
		return aep.BlendingModeScreen, nil
	case "overlay":
		return aep.BlendingModeOverlay, nil
	case "soft_light":
		return aep.BlendingModeSoftLight, nil
	case "hard_light":
		return aep.BlendingModeHardLight, nil
	case "darken":
		return aep.BlendingModeDarken, nil
	case "lighten":
		return aep.BlendingModeLighten, nil
	case "classic_difference":
		return aep.BlendingModeClassicDifference, nil
	case "hue":
		return aep.BlendingModeHue, nil
	case "saturation":
		return aep.BlendingModeSaturation, nil
	case "color":
		return aep.BlendingModeColor, nil
	case "luminosity":
		return aep.BlendingModeLuminosity, nil
	case "stencil_alpha":
		return aep.BlendingModeStencilAlpha, nil
	case "stencil_luma":
		return aep.BlendingModeStencilLuma, nil
	case "silhouette_alpha":
		return aep.BlendingModeSilhouetteAlpha, nil
	case "silhouette_luma":
		return aep.BlendingModeSilhouetteLuma, nil
	case "luminescent_premul":
		return aep.BlendingModeLuminescentPremul, nil
	case "alpha_add":
		return aep.BlendingModeAlphaAdd, nil
	case "classic_color_dodge":
		return aep.BlendingModeClassicColorDodge, nil
	case "classic_color_burn":
		return aep.BlendingModeClassicColorBurn, nil
	case "exclusion":
		return aep.BlendingModeExclusion, nil
	case "difference":
		return aep.BlendingModeDifference, nil
	case "color_dodge":
		return aep.BlendingModeColorDodge, nil
	case "color_burn":
		return aep.BlendingModeColorBurn, nil
	case "linear_dodge":
		return aep.BlendingModeLinearDodge, nil
	case "linear_burn":
		return aep.BlendingModeLinearBurn, nil
	case "linear_light":
		return aep.BlendingModeLinearLight, nil
	case "vivid_light":
		return aep.BlendingModeVividLight, nil
	case "pin_light":
		return aep.BlendingModePinLight, nil
	case "hard_mix":
		return aep.BlendingModeHardMix, nil
	case "lighter_color":
		return aep.BlendingModeLighterColor, nil
	case "darker_color":
		return aep.BlendingModeDarkerColor, nil
	case "subtract":
		return aep.BlendingModeSubtract, nil
	case "divide":
		return aep.BlendingModeDivide, nil
	default:
		return 0, fmt.Errorf("unsupported blending_mode %q", value)
	}
}

func lightKind(value string) (aep.LightKind, error) {
	switch strings.ToLower(value) {
	case "parallel":
		return aep.LightKindParallel, nil
	case "spot":
		return aep.LightKindSpot, nil
	case "point":
		return aep.LightKindPoint, nil
	case "ambient":
		return aep.LightKindAmbient, nil
	default:
		return 0, fmt.Errorf("unsupported kind %q", value)
	}
}
