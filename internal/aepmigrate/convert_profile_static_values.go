package aepmigrate

func propertyFloatValue(value any) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

func staticVector(value any, length int) ([]float64, bool) {
	var out []float64
	switch v := value.(type) {
	case []float64:
		out = append([]float64(nil), v...)
	case []any:
		out = make([]float64, 0, len(v))
		for _, item := range v {
			got, ok := item.(float64)
			if !ok {
				return nil, false
			}
			out = append(out, got)
		}
	case [2]float64:
		out = []float64{v[0], v[1]}
	case [3]float64:
		out = []float64{v[0], v[1], v[2]}
	case [4]float64:
		out = []float64{v[0], v[1], v[2], v[3]}
	default:
		return nil, false
	}
	if len(out) != length {
		return nil, false
	}
	return out, true
}

func staticVectorAtLeast(value any, length int) ([]float64, bool) {
	out, ok := staticVectorAny(value)
	if !ok || len(out) < length {
		return nil, false
	}
	return out, true
}

func staticVectorAny(value any) ([]float64, bool) {
	switch v := value.(type) {
	case []float64:
		return append([]float64(nil), v...), true
	case []any:
		out := make([]float64, 0, len(v))
		for _, item := range v {
			got, ok := item.(float64)
			if !ok {
				return nil, false
			}
			out = append(out, got)
		}
		return out, true
	case [2]float64:
		return []float64{v[0], v[1]}, true
	case [3]float64:
		return []float64{v[0], v[1], v[2]}, true
	case [4]float64:
		return []float64{v[0], v[1], v[2], v[3]}, true
	default:
		return nil, false
	}
}

func profileScaleToWriter(value float64) float64 {
	return value * 100
}

func profileOpacityToWriter(value float64) float64 {
	return value * 100
}
