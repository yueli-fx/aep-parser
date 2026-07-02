package aepmigrate

func profileARGBToRGBA(value []float64) [4]float64 {
	return [4]float64{
		colorByteToUnit(value[1]),
		colorByteToUnit(value[2]),
		colorByteToUnit(value[3]),
		colorByteToUnit(value[0]),
	}
}

func colorByteToUnit(value float64) float64 {
	if value > 1 {
		return value / 255
	}
	return value
}
