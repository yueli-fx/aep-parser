package aepmigrate

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeTextAnimator(targetLayer *aep.Layer, property profile.Property, rangeValues textAnimatorRange) error {
	if handled, err := materializeTextAnimatorTransform(targetLayer, property, rangeValues); handled {
		return err
	}
	if handled, err := materializeTextAnimatorStyle(targetLayer, property, rangeValues); handled {
		return err
	}
	return fmt.Errorf("unsupported text animator property")
}
