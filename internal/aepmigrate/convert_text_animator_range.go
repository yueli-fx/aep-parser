package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/profile"
)

type textAnimatorRange struct {
	start  float64
	end    float64
	offset profile.Property
	value  float64
}

func textAnimatorRangeValues(layer profile.Layer) (textAnimatorRange, error) {
	startProperty, ok := propertyByMatchName(layer.Properties, "ADBE Text Percent Start")
	if !ok {
		return textAnimatorRange{}, fmt.Errorf("range start property missing")
	}
	start, ok := textAnimatorInitialFloatValue(startProperty)
	if !ok {
		return textAnimatorRange{}, fmt.Errorf("range start value unsupported (%T)", startProperty.StaticValue)
	}
	endProperty, ok := propertyByMatchName(layer.Properties, "ADBE Text Percent End")
	if !ok {
		return textAnimatorRange{}, fmt.Errorf("range end property missing")
	}
	end, ok := textAnimatorInitialFloatValue(endProperty)
	if !ok {
		return textAnimatorRange{}, fmt.Errorf("range end value unsupported (%T)", endProperty.StaticValue)
	}
	offsetProperty, ok := propertyByMatchName(layer.Properties, "ADBE Text Percent Offset")
	if !ok {
		return textAnimatorRange{}, fmt.Errorf("range offset property missing")
	}
	offset, ok := textAnimatorInitialFloatValue(offsetProperty)
	if !ok {
		return textAnimatorRange{}, fmt.Errorf("range offset value unsupported (%T)", offsetProperty.StaticValue)
	}
	return textAnimatorRange{
		start:  start,
		end:    end,
		offset: offsetProperty,
		value:  offset,
	}, nil
}
