package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func maskBezierPathFromProfile(source profile.Mask) aep.BezierPath {
	return maskPathFromVertices(source.Vertices, source.Closed)
}

func maskPathKeysFromProfile(source profile.Mask) []aep.MaskPathKey {
	keys := make([]aep.MaskPathKey, 0, len(source.PathKeyframes))
	for _, keyframe := range source.PathKeyframes {
		keys = append(keys, aep.MaskPathKey{
			Time:    keyframe.Time,
			Path:    maskPathFromVertices(keyframe.Vertices, source.Closed),
			InEase:  temporalEaseFromProfile(keyframe.InTemporalEase),
			OutEase: temporalEaseFromProfile(keyframe.OutTemporalEase),
		})
	}
	return keys
}

func maskPathFromVertices(vertices []profile.MaskVertex, closed bool) aep.BezierPath {
	path := aep.BezierPath{
		Closed:      closed,
		Vertices:    make([][2]float64, 0, len(vertices)),
		InTangents:  make([][2]float64, 0, len(vertices)),
		OutTangents: make([][2]float64, 0, len(vertices)),
	}
	for _, vertex := range vertices {
		path.Vertices = append(path.Vertices, vertex.Anchor)
		path.InTangents = append(path.InTangents, vertex.InTangent)
		path.OutTangents = append(path.OutTangents, vertex.OutTangent)
	}
	return path
}

func temporalEaseFromProfile(ease *profile.TemporalEase) aep.TemporalEase {
	if ease == nil {
		return aep.TemporalEase{}
	}
	return aep.TemporalEase{Speed: ease.Speed, Influence: ease.Influence}
}

func convertMaskMode(value string) (aep.MaskMode, error) {
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

func convertMaskMotionBlur(value string) (aep.MaskMotionBlurMode, error) {
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

func convertMaskFeatherFalloff(value string) (aep.MaskFeatherFalloff, error) {
	switch value {
	case "smooth":
		return aep.MaskFeatherFalloffSmooth, nil
	case "linear":
		return aep.MaskFeatherFalloffLinear, nil
	default:
		return 0, fmt.Errorf("unsupported mask feather_falloff %q", value)
	}
}

func profileRGBColor(value []float64) ([3]uint8, bool) {
	if len(value) != 3 {
		return [3]uint8{}, false
	}
	return [3]uint8{profileColorByte(value[0]), profileColorByte(value[1]), profileColorByte(value[2])}, true
}

func profileColorByte(value float64) uint8 {
	if value <= 1 {
		value *= 255
	}
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}
	return uint8(value + 0.5)
}
