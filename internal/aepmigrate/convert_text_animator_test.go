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
