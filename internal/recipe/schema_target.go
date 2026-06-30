package recipe

import (
	"fmt"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

type recipeProjectTarget int

const (
	aepTargetAE2020 recipeProjectTarget = 2020
	aepTargetAE2022 recipeProjectTarget = 2022
	aepTargetAE2025 recipeProjectTarget = 2025
)

func parseRecipeProjectTarget(version string) (recipeProjectTarget, error) {
	switch strings.ToUpper(strings.TrimSpace(version)) {
	case "", "AE2020", "2020":
		return aepTargetAE2020, nil
	case "AE2022", "2022":
		return aepTargetAE2022, nil
	case "AE2025", "2025":
		return aepTargetAE2025, nil
	default:
		return 0, fmt.Errorf("target_version must be AE2020, AE2022, or AE2025")
	}
}

func projectBitsPerChannel(value string) (aep.BitsPerChannel, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "8", "8bpc":
		return aep.BPC8, nil
	case "16", "16bpc":
		return aep.BPC16, nil
	case "32", "32bpc":
		return aep.BPC32, nil
	default:
		return 0, fmt.Errorf("bits_per_channel must be 8, 16, or 32")
	}
}

func projectBitsPerChannelProfileValue(value string) (string, error) {
	bpc, err := projectBitsPerChannel(value)
	if err != nil {
		return "", err
	}
	return bpc.String(), nil
}

func projectTimeDisplayType(value string) (aep.TimeDisplayType, error) {
	switch normalizeEnum(value) {
	case "timecode":
		return aep.TimeDisplayTypeTimecode, nil
	case "frames":
		return aep.TimeDisplayTypeFrames, nil
	default:
		return 0, fmt.Errorf("time_display_type must be timecode or frames")
	}
}

func projectTimeDisplayTypeProfileValue(value string) (string, error) {
	v, err := projectTimeDisplayType(value)
	if err != nil {
		return "", err
	}
	switch v {
	case aep.TimeDisplayTypeFrames:
		return "frames", nil
	default:
		return "timecode", nil
	}
}

func projectFramesCountType(value string) (aep.FramesCountType, error) {
	switch normalizeEnum(value) {
	case "start_0", "start0":
		return aep.FramesCountTypeStart0, nil
	case "start_1", "start1":
		return aep.FramesCountTypeStart1, nil
	case "timecode_conversion":
		return aep.FramesCountTypeTimecodeConversion, nil
	default:
		return 0, fmt.Errorf("frames_count_type must be start_0, start_1, or timecode_conversion")
	}
}

func projectFramesCountTypeProfileValue(value string) (string, error) {
	v, err := projectFramesCountType(value)
	if err != nil {
		return "", err
	}
	switch v {
	case aep.FramesCountTypeStart1:
		return "start_1", nil
	case aep.FramesCountTypeTimecodeConversion:
		return "timecode_conversion", nil
	default:
		return "start_0", nil
	}
}

func projectFeetFramesFilmType(value string) (aep.FeetFramesFilmType, error) {
	switch normalizeEnum(value) {
	case "35mm", "mm35":
		return aep.FeetFramesFilmTypeMM35, nil
	case "16mm", "mm16":
		return aep.FeetFramesFilmTypeMM16, nil
	default:
		return 0, fmt.Errorf("feet_frames_film_type must be 35mm or 16mm")
	}
}

func projectFeetFramesFilmTypeProfileValue(value string) (string, error) {
	v, err := projectFeetFramesFilmType(value)
	if err != nil {
		return "", err
	}
	switch v {
	case aep.FeetFramesFilmTypeMM16:
		return "16mm", nil
	default:
		return "35mm", nil
	}
}

func projectFootageTimecodeDisplayStartType(value string) (aep.FootageTimecodeDisplayStartType, error) {
	switch normalizeEnum(value) {
	case "start_0", "start0":
		return aep.FootageTimecodeDisplayStartTypeStart0, nil
	case "source_media", "use_source_media":
		return aep.FootageTimecodeDisplayStartTypeUseSourceMedia, nil
	default:
		return 0, fmt.Errorf("footage_timecode_display_start_type must be start_0 or source_media")
	}
}

func projectFootageTimecodeDisplayStartTypeProfileValue(value string) (string, error) {
	v, err := projectFootageTimecodeDisplayStartType(value)
	if err != nil {
		return "", err
	}
	switch v {
	case aep.FootageTimecodeDisplayStartTypeUseSourceMedia:
		return "source_media", nil
	default:
		return "start_0", nil
	}
}

func projectExpressionEngine(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "extendscript", "javascript-1.0":
		return strings.ToLower(strings.TrimSpace(value)), nil
	default:
		return "", fmt.Errorf("expression_engine must be extendscript or javascript-1.0")
	}
}

func projectAudioSampleRate(value float64) error {
	switch value {
	case 22050, 32000, 44100, 48000, 96000:
		return nil
	default:
		return fmt.Errorf("audio_sample_rate must be one of 22050, 32000, 44100, 48000, or 96000")
	}
}

func projectWorkingGamma(value float64) error {
	switch value {
	case 2.2, 2.4:
		return nil
	default:
		return fmt.Errorf("working_gamma must be 2.2 or 2.4")
	}
}

func projectTimecodeDefaultBase(value int) error {
	if value < 1 || value > 999 {
		return fmt.Errorf("timecode_default_base must be between 1 and 999")
	}
	return nil
}

func (target recipeProjectTarget) aepTarget() aep.AETarget {
	switch target {
	case aepTargetAE2022:
		return aep.TargetAE2022
	case aepTargetAE2025:
		return aep.TargetAE2025
	default:
		return aep.TargetAE2020
	}
}
