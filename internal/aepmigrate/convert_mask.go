package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeProjectMasks(project *aep.Project, prof *profile.Profile) (*aep.Project, error) {
	if !hasProfileMasks(prof) {
		return project, nil
	}
	reopened, err := aep.Reopen(project)
	if err != nil {
		return nil, fmt.Errorf("reopen for masks: %w", err)
	}
	for compIndex, sourceComp := range prof.Comps {
		targetComp, err := targetCompBySource(reopened, compIndex, sourceComp)
		if err != nil {
			return nil, err
		}
		for layerIndex, sourceLayer := range sourceComp.Layers {
			if len(sourceLayer.Masks) == 0 {
				continue
			}
			if layerIndex >= len(targetComp.Layers) {
				return nil, fmt.Errorf("comp %q masks: target layer index %d missing", sourceComp.Name, layerIndex)
			}
			targetLayer := targetComp.Layers[layerIndex]
			for maskIndex, sourceMask := range sourceLayer.Masks {
				mask, err := aep.AddMask(targetLayer, sourceMask.Name, maskBezierPathFromProfile(sourceMask))
				if err != nil {
					return nil, fmt.Errorf("comp %q layer %q add mask %d: %w", sourceComp.Name, sourceLayer.Name, maskIndex, err)
				}
				if err := materializeMaskSurface(targetLayer, mask, sourceMask); err != nil {
					return nil, fmt.Errorf("comp %q layer %q mask %q: %w", sourceComp.Name, sourceLayer.Name, sourceMask.Name, err)
				}
			}
		}
	}
	return reopened, nil
}

func hasProfileMasks(prof *profile.Profile) bool {
	if prof == nil {
		return false
	}
	for _, comp := range prof.Comps {
		for _, layer := range comp.Layers {
			if len(layer.Masks) != 0 {
				return true
			}
		}
	}
	return false
}

func isSupportedMaskSurface(masks []profile.Mask) bool {
	for _, mask := range masks {
		if len(mask.Vertices) == 0 {
			return false
		}
		if _, err := convertMaskMode(mask.Mode); mask.Mode != "" && err != nil {
			return false
		}
		if _, err := convertMaskMotionBlur(mask.MotionBlur); mask.MotionBlur != "" && err != nil {
			return false
		}
		if _, err := convertMaskFeatherFalloff(mask.FeatherFalloff); mask.FeatherFalloff != "" && err != nil {
			return false
		}
		if len(mask.Color) != 0 {
			if _, ok := profileRGBColor(mask.Color); !ok {
				return false
			}
		}
		if len(mask.Feather) != 0 && len(mask.Feather) != 2 {
			return false
		}
		if len(mask.PathKeyframes) == 1 {
			return false
		}
		for _, keyframe := range mask.PathKeyframes {
			if len(keyframe.Vertices) == 0 {
				return false
			}
		}
	}
	return true
}

func materializeMaskSurface(layer *aep.Layer, mask *aep.Mask, source profile.Mask) error {
	if source.Mode != "" {
		mode, err := convertMaskMode(source.Mode)
		if err != nil {
			return fmt.Errorf("mode: %w", err)
		}
		if err := mask.SetMode(mode); err != nil {
			return fmt.Errorf("mode: %w", err)
		}
	}
	if err := mask.SetInverted(source.Inverted); err != nil {
		return fmt.Errorf("inverted: %w", err)
	}
	if err := mask.SetLocked(source.Locked); err != nil {
		return fmt.Errorf("locked: %w", err)
	}
	if len(source.Color) != 0 {
		color, ok := profileRGBColor(source.Color)
		if !ok {
			return fmt.Errorf("color: unsupported value %v", source.Color)
		}
		if err := mask.SetColor(color); err != nil {
			return fmt.Errorf("color: %w", err)
		}
	}
	if source.MotionBlur != "" {
		mode, err := convertMaskMotionBlur(source.MotionBlur)
		if err != nil {
			return fmt.Errorf("motion_blur: %w", err)
		}
		if err := mask.SetMaskMotionBlur(mode); err != nil {
			return fmt.Errorf("motion_blur: %w", err)
		}
	}
	if source.FeatherFalloff != "" {
		falloff, err := convertMaskFeatherFalloff(source.FeatherFalloff)
		if err != nil {
			return fmt.Errorf("feather_falloff: %w", err)
		}
		if err := mask.SetFeatherFalloff(falloff); err != nil {
			return fmt.Errorf("feather_falloff: %w", err)
		}
	}
	if err := mask.SetOpacity(source.Opacity); err != nil {
		return fmt.Errorf("opacity: %w", err)
	}
	if len(source.Feather) == 2 {
		if err := mask.SetFeather([2]float64{source.Feather[0], source.Feather[1]}); err != nil {
			return fmt.Errorf("feather: %w", err)
		}
	}
	if err := mask.SetExpansion(source.Expansion); err != nil {
		return fmt.Errorf("expansion: %w", err)
	}
	if err := mask.SetClosed(source.Closed); err != nil {
		return fmt.Errorf("closed: %w", err)
	}
	if len(source.PathKeyframes) != 0 {
		if err := aep.SetMaskPathKeyframes(layer, mask, maskPathKeysFromProfile(source)); err != nil {
			return fmt.Errorf("path_keyframes: %w", err)
		}
	}
	return nil
}
