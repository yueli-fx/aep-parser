package scene

// transformFixedDefaults holds default values for transform properties
// whose defaults are fixed constants (not dependent on comp/layer
// dimensions). Ported from the reference parser's _TRANSFORM_FIXED_DEFAULTS table.
var transformFixedDefaults = map[string]any{
	MatchNameScale:       []float64{100, 100, 100},
	"ADBE Rotate X":      0.0,
	"ADBE Rotate Y":      0.0,
	MatchNameRotateZ:     0.0,
	MatchNameOpacity:     100.0,
	"ADBE Orientation":   []float64{0, 0, 0},
	"ADBE Position_2":    0.0,
	MatchNameEnvirAppear: 1.0,
}

// AssignTransformDefaults sets Property.DefaultValue on transform
// properties. Spatial defaults (Anchor Point, Position) depend on
// comp dimensions and layer type. Exported for the serializer stage
// (parse/mutate in internal/aep) which wires it during property-tree builds.
func AssignTransformDefaults(props []*Property, comp *Composition, layerType LayerType) {
	if comp == nil {
		return
	}

	compW := float64(comp.Width)
	compH := float64(comp.Height)

	// Anchor Point default depends on layer type.
	var anchorW, anchorH float64
	switch layerType {
	case LayerTypeShape, LayerTypeText, LayerTypeNull:
		anchorW, anchorH = 0, 0
	default:
		// For AV layers, anchor defaults to source center.
		// We don't have source dimensions here, so use comp dims
		// as a reasonable fallback (matches the reference parser for non-AV layers).
		anchorW, anchorH = compW, compH
	}

	for _, p := range props {
		if p.DefaultValue != nil {
			continue // already set (e.g. by pard)
		}
		if v, ok := transformFixedDefaults[p.MatchName]; ok {
			p.DefaultValue = v
			continue
		}
		// Spatial defaults.
		switch p.MatchName {
		case MatchNameAnchorPoint:
			p.DefaultValue = []float64{anchorW / 2, anchorH / 2, 0}
		case MatchNamePosition:
			p.DefaultValue = []float64{compW / 2, compH / 2, 0}
		case MatchNamePosition0:
			p.DefaultValue = compW / 2
		case MatchNamePosition1:
			p.DefaultValue = compH / 2
		}
	}
}
