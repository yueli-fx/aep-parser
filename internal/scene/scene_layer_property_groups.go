package scene

// P2c PropertyGroup hierarchy — Layer-level accessors mirroring py-aep's
// `layer.transform()` / `layer.effects()` / etc. shortcuts. Each returns
// the corresponding *AEPropertyGroup from the parsed tree, or nil when:
//
//   - the layer was built outside the parser (no tree was built), or
//   - the group is not present (e.g. LightOptionsGroup on an AV layer).
//
// The flat Layer.Properties / Effects / Masks / Markers slices remain the
// canonical way to enumerate decoded values; this hierarchy is for
// chained AE-script-style lookup.

// Group match-names for AE's standard layer-level property groups. The
// exact strings AE writes in tdmn chunks; verified against multiple
// fixtures (re_cameralight.aep, re_text.aep, re_material_options.aep).
const (
	MatchNameGroupTransform       = "ADBE Transform Group"
	MatchNameGroupAudio           = "ADBE Audio Group"
	MatchNameGroupLayerStyles     = "ADBE Layer Styles"
	MatchNameGroupEffectParade    = "ADBE Effect Parade"
	MatchNameGroupMaskParade      = "ADBE Mask Parade"
	MatchNameGroupTextProperties  = "ADBE Text Properties"
	MatchNameGroupCameraOptions   = "ADBE Camera Options Group"
	MatchNameGroupLightOptions    = "ADBE Light Options Group"
	MatchNameGroupMaterialOptions = "ADBE Material Options Group"
	MatchNameGroupGeometryOptions = "ADBE Extrsn Options Group"
	MatchNameGroupShapeContents   = "ADBE Root Vectors Group"
)

// PropertyTree returns the layer's hierarchical PropertyGroup root, or
// nil for layers built outside the parser. The root has empty MatchName /
// Name; its Children are the named top-level groups AE writes under the
// Layr (Transform, Mask Parade, Effect Parade, etc.).
func (l *Layer) PropertyTree() *AEPropertyGroup { return l.propertyTree }

// PropertyGroupByMatchName returns the top-level subgroup of the layer's
// property tree whose match-name equals name. Equivalent to
// `l.PropertyTree().Group(name)`. Returns nil when the layer has no tree
// or the named subgroup isn't present.
func (l *Layer) PropertyGroupByMatchName(matchName string) *AEPropertyGroup {
	if l.propertyTree == nil {
		return nil
	}
	return l.propertyTree.Group(matchName)
}

// PropertyByPath walks the property hierarchy following each match-name
// in turn. The last name must resolve to a leaf property; intermediate
// names must resolve to subgroups. Returns nil if the layer has no tree
// or any step fails.
//
// Example:
//
//	pos := layer.PropertyByPath("ADBE Transform Group", "ADBE Position")
func (l *Layer) PropertyByPath(matchNames ...string) *Property {
	if l.propertyTree == nil {
		return nil
	}
	return l.propertyTree.PropertyByPath(matchNames...)
}

// TransformGroup returns the layer's "ADBE Transform Group" subgroup, or
// nil if absent. Standard layers always have it; effect-only layers do
// not.
func (l *Layer) TransformGroup() *AEPropertyGroup {
	return l.PropertyGroupByMatchName(MatchNameGroupTransform)
}

// AudioGroup returns the layer's "ADBE Audio Group" subgroup, or nil if
// absent. Present on layers with audio content.
func (l *Layer) AudioGroup() *AEPropertyGroup {
	return l.PropertyGroupByMatchName(MatchNameGroupAudio)
}

// LayerStylesGroup returns the layer's "ADBE Layer Styles" subgroup, or
// nil if absent. Carries Drop Shadow / Bevel / Glow / etc. effect-style
// sub-properties (separate from the "ADBE Effect Parade" effect list).
func (l *Layer) LayerStylesGroup() *AEPropertyGroup {
	return l.PropertyGroupByMatchName(MatchNameGroupLayerStyles)
}

// EffectsParade returns the layer's "ADBE Effect Parade" subgroup, or
// nil if absent. The parade carries each effect as a sub-PropertyGroup
// named by the effect's match-name. For decoded effect parameter values
// prefer Layer.Effects (the parade tree captures the layout but does
// not re-parse effect parameters).
func (l *Layer) EffectsParade() *AEPropertyGroup {
	return l.PropertyGroupByMatchName(MatchNameGroupEffectParade)
}

// MaskParade returns the layer's "ADBE Mask Parade" subgroup, or nil if
// absent. For decoded mask data prefer Layer.Masks.
func (l *Layer) MaskParade() *AEPropertyGroup {
	return l.PropertyGroupByMatchName(MatchNameGroupMaskParade)
}

// TextPropertiesGroup returns the layer's "ADBE Text Properties"
// subgroup, or nil for non-text layers.
func (l *Layer) TextPropertiesGroup() *AEPropertyGroup {
	return l.PropertyGroupByMatchName(MatchNameGroupTextProperties)
}

// CameraOptionsGroup returns the camera layer's "ADBE Camera Options
// Group", or nil for non-camera layers.
func (l *Layer) CameraOptionsGroup() *AEPropertyGroup {
	return l.PropertyGroupByMatchName(MatchNameGroupCameraOptions)
}

// LightOptionsGroup returns the light layer's "ADBE Light Options
// Group", or nil for non-light layers.
func (l *Layer) LightOptionsGroup() *AEPropertyGroup {
	return l.PropertyGroupByMatchName(MatchNameGroupLightOptions)
}

// MaterialOptionsGroup returns the 3D AV layer's "ADBE Material Options
// Group", or nil for 2D layers (Is3D == false).
func (l *Layer) MaterialOptionsGroup() *AEPropertyGroup {
	return l.PropertyGroupByMatchName(MatchNameGroupMaterialOptions)
}

// GeometryOptionsGroup returns the 3D layer's "ADBE Extrsn Options
// Group", or nil if absent (2D layers, AE 23 and older).
func (l *Layer) GeometryOptionsGroup() *AEPropertyGroup {
	return l.PropertyGroupByMatchName(MatchNameGroupGeometryOptions)
}

// ShapeContentsGroup returns a shape layer's "ADBE Root Vectors Group",
// or nil for non-shape layers. For typed access prefer the V2.2
// ShapeLayer wrapper (see WrapShapeLayer).
func (l *Layer) ShapeContentsGroup() *AEPropertyGroup {
	return l.PropertyGroupByMatchName(MatchNameGroupShapeContents)
}
