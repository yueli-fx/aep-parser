package aepmigrate

import (
	"fmt"
	"strings"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func normalizeTextEnum(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	return strings.ReplaceAll(value, "-", "_")
}

func textJustification(value string) (aep.TextJustification, error) {
	switch normalizeTextEnum(value) {
	case "", "left":
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
	switch normalizeTextEnum(value) {
	case "", "normal":
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
	switch normalizeTextEnum(value) {
	case "", "normal":
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
	switch normalizeTextEnum(value) {
	case "no_auto":
		return aep.TextAutoKernNoAuto, nil
	case "", "metric":
		return aep.TextAutoKernMetric, nil
	case "optical":
		return aep.TextAutoKernOptical, nil
	default:
		return 0, fmt.Errorf("unsupported auto_kern_type %q", value)
	}
}

func textLineJoinType(value string) (aep.TextLineJoinType, error) {
	switch normalizeTextEnum(value) {
	case "", "miter":
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
	switch normalizeTextEnum(value) {
	case "", "default":
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
	switch normalizeTextEnum(value) {
	case "", "roman":
		return aep.TextLeadingRoman, nil
	case "japanese":
		return aep.TextLeadingJapanese, nil
	default:
		return 0, fmt.Errorf("unsupported leading_type %q", value)
	}
}

func textParagraphDirection(value string) (aep.TextParagraphDirection, error) {
	switch normalizeTextEnum(value) {
	case "", "ltr":
		return aep.TextDirectionLeftToRight, nil
	case "rtl":
		return aep.TextDirectionRightToLeft, nil
	default:
		return 0, fmt.Errorf("unsupported paragraph_direction %q", value)
	}
}
