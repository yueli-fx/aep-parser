package aepmigrate

import (
	"math"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/profile"
)

func materializeLayerTiming(layer *aep.Layer, source profile.Layer) error {
	if source.Timing.StartTime != 0 {
		if err := layer.SetStartTime(source.Timing.StartTime); err != nil {
			return err
		}
	}
	if source.Timing.InPoint != 0 {
		if err := layer.SetInPoint(source.Timing.InPoint); err != nil {
			return err
		}
	}
	if source.Timing.OutPoint != 0 && !isDefaultLayerOutPoint(source.Timing) {
		if err := layer.SetOutPoint(source.Timing.OutPoint); err != nil {
			return err
		}
	}
	if source.Timing.Stretch != 0 && source.Timing.Stretch != 1 {
		if err := layer.SetStretch(source.Timing.Stretch); err != nil {
			return err
		}
	}
	return nil
}

func isDefaultLayerOutPoint(timing profile.LayerTiming) bool {
	return timing.InPoint == 0 && math.Abs(timing.OutPoint-timing.Duration) < 1e-6
}
