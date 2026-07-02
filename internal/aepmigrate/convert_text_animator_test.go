package aepmigrate

import "testing"

func TestConvertWritesRecipeTextAnimatorOpacityProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-text-animator-opacity.json")
}

func TestConvertWritesRecipeTextAnimatorPositionProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-text-animator-position.json")
}

func TestConvertWritesRecipeTextAnimatorRangeOffsetProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-text-animator-range-offset.json")
}

func TestConvertWritesRecipeTextAnimatorColorValueKeyframesProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-text-animator-color-value-keyframes.json")
}

func TestConvertWritesRecipeTextAnimatorFillOpacityValueKeyframesProject(t *testing.T) {
	assertRecipeConvertsPass(t, "minimal-text-animator-fill-opacity-value-keyframes.json")
}

func TestConvertWritesRecipeTextAnimatorRemainingValueKeyframesProjects(t *testing.T) {
	for _, recipeName := range []string{
		"minimal-text-animator-stroke-opacity-value-keyframes.json",
		"minimal-text-animator-stroke-width-value-keyframes.json",
		"minimal-text-animator-skew-value-keyframes.json",
		"minimal-text-animator-rotation-x-value-keyframes.json",
		"minimal-text-animator-rotation-y-value-keyframes.json",
		"minimal-text-animator-stroke-color-value-keyframes.json",
	} {
		t.Run(recipeName, func(t *testing.T) {
			assertRecipeConvertsPass(t, recipeName)
		})
	}
}

func TestConvertWritesRecipeTextAnimatorRotationXYProjects(t *testing.T) {
	for _, recipeName := range []string{
		"minimal-text-animator-rotation-x.json",
		"minimal-text-animator-rotation-y.json",
	} {
		t.Run(recipeName, func(t *testing.T) {
			assertRecipeConvertsPass(t, recipeName)
		})
	}
}
