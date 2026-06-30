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
