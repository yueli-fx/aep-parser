package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

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
