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
