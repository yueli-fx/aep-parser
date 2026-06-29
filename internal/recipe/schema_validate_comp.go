package recipe

import (
	"fmt"
)

func validateExpectedLayerTiming(timing *ExpectedLayerTiming, path string, addRefusal func(string, string, string)) {
	if timing.StartTime != nil && *timing.StartTime < 0 {
		addRefusal("invalid_expected_profile", path+".start_time", "start_time must be non-negative")
	}
	if timing.InPoint != nil && *timing.InPoint < 0 {
		addRefusal("invalid_expected_profile", path+".in_point", "in_point must be non-negative")
	}
	if timing.OutPoint != nil && *timing.OutPoint < 0 {
		addRefusal("invalid_expected_profile", path+".out_point", "out_point must be non-negative")
	}
	if timing.InPoint != nil && timing.OutPoint != nil && *timing.OutPoint < *timing.InPoint {
		addRefusal("invalid_expected_profile", path+".out_point", "out_point must be greater than or equal to in_point")
	}
	if timing.Duration != nil && *timing.Duration < 0 {
		addRefusal("invalid_expected_profile", path+".duration", "duration must be non-negative")
	}
}

func validateCompMotionBlur(spec *CompMotionBlurSpec, path string, recordCapability func(string, string) CapabilityLookup, addRefusal func(string, string, string)) {
	if spec.Enabled != nil {
		recordCapability("SetCompMotionBlur", path+".enabled")
	}
	if spec.ShutterAngle != nil {
		recordCapability("SetShutterAngle", path+".shutter_angle")
		if *spec.ShutterAngle < 0 || *spec.ShutterAngle > 720 || !isWholeNumber(*spec.ShutterAngle) {
			addRefusal("invalid_comp_motion_blur_shutter_angle", path+".shutter_angle", "motion_blur shutter_angle must be an integer between 0 and 720")
		}
	}
	if spec.ShutterPhase != nil {
		recordCapability("SetShutterPhase", path+".shutter_phase")
		if !isWholeNumber(*spec.ShutterPhase) {
			addRefusal("invalid_comp_motion_blur_shutter_phase", path+".shutter_phase", "motion_blur shutter_phase must be an integer")
		}
	}
	if spec.AdaptiveSampleLimit != nil {
		recordCapability("SetMotionBlurAdaptiveSampleLimit", path+".adaptive_sample_limit")
		if *spec.AdaptiveSampleLimit < 0 || !isWholeNumber(*spec.AdaptiveSampleLimit) {
			addRefusal("invalid_comp_motion_blur_adaptive_sample_limit", path+".adaptive_sample_limit", "motion_blur adaptive_sample_limit must be a non-negative integer")
		}
	}
	if spec.SamplesPerFrame != nil {
		recordCapability("SetMotionBlurSamplesPerFrame", path+".samples_per_frame")
		if *spec.SamplesPerFrame < 0 || !isWholeNumber(*spec.SamplesPerFrame) {
			addRefusal("invalid_comp_motion_blur_samples_per_frame", path+".samples_per_frame", "motion_blur samples_per_frame must be a non-negative integer")
		}
	}
}

func validateCompWorkArea(spec *CompWorkAreaSpec, path string, compDuration float64, recordCapability func(string, string) CapabilityLookup, addRefusal func(string, string, string)) {
	recordCapability("SetWorkArea", path)
	if spec.Start == nil || spec.End == nil {
		addRefusal("invalid_comp_work_area", path, "work_area start and end are required")
		return
	}
	if *spec.Start < 0 || *spec.End < *spec.Start || *spec.End > compDuration {
		addRefusal("invalid_comp_work_area", path, "work_area must satisfy 0 <= start <= end <= comp duration")
	}
}

func validateResolutionFactor(values []float64, path string, addRefusal func(string, string, string)) {
	if len(values) != 2 {
		addRefusal("invalid_comp_resolution_factor", path, "resolution_factor must have exactly two values")
		return
	}
	for i, value := range values {
		if value <= 0 || value > 65535 || !isWholeNumber(value) {
			addRefusal("invalid_comp_resolution_factor", fmt.Sprintf("%s[%d]", path, i), "resolution_factor values must be positive integers in uint16 range")
		}
	}
}

func validatePixelAspect(value float64, path string, addRefusal func(string, string, string)) {
	if value <= 0 {
		addRefusal("invalid_comp_pixel_aspect", path, "pixel_aspect must be positive")
	}
}

func validateDisplayStartTime(value float64, path string, addRefusal func(string, string, string)) {
	if value < 0 {
		addRefusal("invalid_comp_display_start_time", path, "display_start_time must be non-negative")
	}
}

func validateCompLabel(value float64, path string, addRefusal func(string, string, string)) {
	if value < 0 || value > 16 || !isWholeNumber(value) {
		addRefusal("invalid_comp_label", path, "label must be an integer between 0 and 16")
	}
}

func validateLayerLabel(value float64, path string, addRefusal func(string, string, string)) {
	if value < 0 || value > 16 || !isWholeNumber(value) {
		addRefusal("invalid_layer_label", path, "label must be an integer between 0 and 16")
	}
}
