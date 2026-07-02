package aepmigrate

import "github.com/yueli-fx/aep-parser/internal/profile"

func textAnimatorInitialFloatValue(property profile.Property) (float64, bool) {
	if len(property.Keyframes) != 0 {
		return propertyFloatValue(property.Keyframes[0].Value)
	}
	return propertyFloatValue(property.StaticValue)
}

func textAnimatorInitialVectorValue(property profile.Property, length int) ([]float64, bool) {
	if len(property.Keyframes) != 0 {
		return staticVectorAtLeast(property.Keyframes[0].Value, length)
	}
	return staticVectorAtLeast(property.StaticValue, length)
}
