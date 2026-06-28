package profile

import (
	"encoding/json"
	"math"
	"os"
)

type EffectDictionary struct {
	AEVersion string                `json:"aeVersion"`
	Effects   map[string]DictEffect `json:"effects"`
}

type DictEffect struct {
	Name   string               `json:"name"`
	Params map[string]DictParam `json:"params"`
}

type DictParam struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Default any    `json:"default"`
}

func LoadEffectDictionary(path string) (*EffectDictionary, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var dict EffectDictionary
	if err := json.Unmarshal(data, &dict); err != nil {
		return nil, err
	}
	return &dict, nil
}

func (d *EffectDictionary) effect(matchName string) (DictEffect, bool) {
	if d == nil || d.Effects == nil {
		return DictEffect{}, false
	}
	effect, ok := d.Effects[matchName]
	return effect, ok
}

func propertyChanged(value, def any) bool {
	if def == nil || value == nil {
		return false
	}
	equal, ok := valuesEqual(value, def)
	return ok && !equal
}

func valuesEqual(value, def any) (bool, bool) {
	if vf, ok := number(value); ok {
		if df, ok := number(def); ok {
			return math.Abs(vf-df) < 1e-6, true
		}
	}
	vs, vok := numberSlice(value)
	ds, dok := numberSlice(def)
	if vok && dok {
		if len(vs) != len(ds) {
			return false, true
		}
		for i := range vs {
			if math.Abs(vs[i]-ds[i]) > 1e-6 {
				return false, true
			}
		}
		return true, true
	}
	if vb, ok := value.(bool); ok {
		if db, ok := def.(bool); ok {
			return vb == db, true
		}
	}
	if vs, ok := value.(string); ok {
		if ds, ok := def.(string); ok {
			return vs == ds, true
		}
	}
	return false, false
}

func number(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint32:
		return float64(n), true
	}
	return 0, false
}

func numberSlice(v any) ([]float64, bool) {
	switch s := v.(type) {
	case []float64:
		return s, true
	case []int:
		out := make([]float64, len(s))
		for i, n := range s {
			out[i] = float64(n)
		}
		return out, true
	case []any:
		out := make([]float64, 0, len(s))
		for _, item := range s {
			n, ok := number(item)
			if !ok {
				return nil, false
			}
			out = append(out, n)
		}
		return out, true
	}
	return nil, false
}
