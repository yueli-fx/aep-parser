package scene

import (
	"fmt"

	"github.com/yueli-fx/aep-parser/internal/codec"
)

// LdtaRawBytes returns a detached copy of the layer's ldta chunk data, or nil
// if the layer has no ldta. It is intended for debugging and RE tools.
func (l *Layer) LdtaRawBytes() []byte {
	if l.runtime.back == nil {
		return nil
	}
	return append([]byte(nil), l.runtime.back.LdtaRaw()...)
}

// This file adds typed convenience accessors for properties unique to
// Camera Layers and Light Layers. The properties themselves are
// flattened into Layer.Properties by parseProperties, so every
// accessor here is just a PropertyByMatchName lookup with a strongly-
// typed name.
//
// Each accessor returns nil if the property isn't present (e.g.
// calling LightColor() on a camera or AV layer).

// Camera Options match-names. The "Camera Options" group sits under
// the camera layer's property tree; AE writes these match-names
// regardless of the comp's render engine.
const (
	MatchNameCameraZoom                   = "ADBE Camera Zoom"
	MatchNameCameraDepthOfField           = "ADBE Camera Depth of Field"
	MatchNameCameraFocusDistance          = "ADBE Camera Focus Distance"
	MatchNameCameraAperture               = "ADBE Camera Aperture"
	MatchNameCameraBlurLevel              = "ADBE Camera Blur Level"
	MatchNameCameraIrisShape              = "ADBE Iris Shape"
	MatchNameCameraIrisRotation           = "ADBE Iris Rotation"
	MatchNameCameraIrisRoundness          = "ADBE Iris Roundness"
	MatchNameCameraIrisAspectRatio        = "ADBE Iris Aspect Ratio"
	MatchNameCameraIrisDiffractionFringe  = "ADBE Iris Diffraction Fringe"
	MatchNameCameraIrisHighlightGain      = "ADBE Iris Highlight Gain"
	MatchNameCameraIrisHighlightThreshold = "ADBE Iris Highlight Threshold"
	// NB: Adobe misspelled this match-name as "Hightlight" — it is the real
	// on-disk name (verified via a generated camera-iris probe script); the
	// correct spelling never matches, so the accessor was silently
	// always-nil before.
	MatchNameCameraIrisHighlightSaturation = "ADBE Iris Hightlight Saturation"
)

// 3D-layer-only Transform match-names. AnchorPoint / Position / Scale /
// RotateZ / Opacity are declared in types_core.go (shared 2D + 3D). The
// 3 names below appear only on 3D-enabled layers.
const (
	MatchNameRotateX     = "ADBE Rotate X"
	MatchNameRotateY     = "ADBE Rotate Y"
	MatchNameOrientation = "ADBE Orientation"
)

// AV-layer match-names that aren't part of the Transform group.
const (
	MatchNameTimeRemap   = "ADBE Time Remapping"
	MatchNameAudioLevels = "ADBE Audio Levels"
)

// Geometry Options match-names. Present on 3D-enabled AV layers
// (regardless of comp renderer — Classic / Advanced 3D / Cinema 4D).
// Shape / text 3D layers carry additional bevel + extrusion match
// names not enumerated here; this set is the universal solid-layer set.
const (
	MatchNameGeometryPlaneCurvature   = "ADBE Plane Curvature"
	MatchNameGeometryPlaneSubdivision = "ADBE Plane Subdivision"
	MatchNameGeometryBevelDirection   = "ADBE Bevel Direction"
)

// Material Options match-names. These properties exist only on 3D-enabled
// AV layers (solid / footage / precomp / shape / text with Is3D = true).
// `ADBE Casts Shadows` is name-shared with Light Layer's CastsShadows;
// `PropertyByMatchName` returns whichever exists on the given layer.
const (
	MatchNameMaterialCastsShadows      = "ADBE Casts Shadows"
	MatchNameMaterialLightTransmission = "ADBE Light Transmission"
	MatchNameMaterialAcceptsShadows    = "ADBE Accepts Shadows"
	MatchNameMaterialAcceptsLights     = "ADBE Accepts Lights"
	MatchNameMaterialShadowColor       = "ADBE Shadow Color"
	MatchNameMaterialAppearsInRefl     = "ADBE Appears in Reflections"
	MatchNameMaterialAmbient           = "ADBE Ambient Coefficient"
	MatchNameMaterialDiffuse           = "ADBE Diffuse Coefficient"
	MatchNameMaterialSpecular          = "ADBE Specular Coefficient"
	MatchNameMaterialShininess         = "ADBE Shininess Coefficient"
	MatchNameMaterialMetal             = "ADBE Metal Coefficient"
	MatchNameMaterialReflection        = "ADBE Reflection Coefficient"
	MatchNameMaterialGlossiness        = "ADBE Glossiness Coefficient"
	MatchNameMaterialFresnel           = "ADBE Fresnel Coefficient"
	MatchNameMaterialTransparency      = "ADBE Transparency Coefficient"
	MatchNameMaterialTranspRolloff     = "ADBE Transp Rolloff"
	MatchNameMaterialIndexOfRefraction = "ADBE Index of Refraction"
)

// MaterialCastsShadowsMode is the tri-state enum for AV 3D layer's Casts
// Shadows option. Light layers use a 2-state bool (use SetLightCastsShadows).
type MaterialCastsShadowsMode uint8

const (
	MaterialCastsOff  MaterialCastsShadowsMode = 0
	MaterialCastsOn   MaterialCastsShadowsMode = 1
	MaterialCastsOnly MaterialCastsShadowsMode = 2 // shadow visible but layer hidden
)

// Light Options match-names.
const (
	MatchNameLightType            = "ADBE Light Type"
	MatchNameLightColor           = "ADBE Light Color"
	MatchNameLightIntensity       = "ADBE Light Intensity"
	MatchNameLightConeAngle       = "ADBE Light Cone Angle"
	MatchNameLightConeFeather     = "ADBE Light Cone Feather 2"
	MatchNameLightFalloffType     = "ADBE Light Falloff Type"
	MatchNameLightFalloffStart    = "ADBE Light Falloff Start"
	MatchNameLightFalloffDistance = "ADBE Light Falloff Distance"
	MatchNameLightCastsShadows    = "ADBE Casts Shadows"
	MatchNameLightShadowDarkness  = "ADBE Light Shadow Darkness"
	MatchNameLightShadowDiffusion = "ADBE Light Shadow Diffusion"
)

// LightKind is the layer-level light type stored in ldta @0x88 (4-byte
// BE uint32, AE 23+ ldta layout). For light layers only; zero for
// non-light layers (the parser leaves Layer.LightKind as default 0 ==
// Parallel for them, so always cross-check Layer.Type == LayerTypeLight
// before reading LightKind).
type LightKind uint8

const (
	LightKindParallel LightKind = 0
	LightKindSpot     LightKind = 1
	LightKindPoint    LightKind = 2
	LightKindAmbient  LightKind = 3
)

// String returns a human-readable name for the light kind.
func (k LightKind) String() string {
	switch k {
	case LightKindParallel:
		return "parallel"
	case LightKindSpot:
		return "spot"
	case LightKindPoint:
		return "point"
	case LightKindAmbient:
		return "ambient"
	}
	return "unknown"
}

// LightTypeID and its constants are kept for backwards compatibility
// but are deprecated. The 4101..4104 values were never observed in real
// .aep files (the actual storage is at ldta @0x88 as small 0..3 enum;
// see LightKind). Prefer LightKind / LightKindParallel etc.
//
// Deprecated: use LightKind. The 4101..4104 values are wrong (no AE
// version writes them) and the actual storage was reverse-engineered
// to ldta @0x88 by re_wave2_ae24.jsx. //nolint:jargon
type LightTypeID int

// Deprecated: use LightKindParallel.
const LightTypeParallel LightTypeID = 4101

// Deprecated: use LightKindSpot.
const LightTypeSpot LightTypeID = 4102

// Deprecated: use LightKindPoint.
const LightTypePoint LightTypeID = 4103

// Deprecated: use LightKindAmbient.
const LightTypeAmbient LightTypeID = 4104

// String returns a human-readable name for the deprecated light type id.
func (t LightTypeID) String() string {
	switch t {
	case LightTypeParallel:
		return "parallel"
	case LightTypeSpot:
		return "spot"
	case LightTypePoint:
		return "point"
	case LightTypeAmbient:
		return "ambient"
	}
	return "unknown"
}

// ──────────────────────────────────────────────
// AV-layer accessors
// ──────────────────────────────────────────────

// TimeRemapEnabled reports whether the layer has time remapping turned
// on. AE always writes the TimeRemap property slot in the property tree
// (so PropertyByMatchName always finds it on AV layers / precomps), but
// only populates it with keyframes when remap is enabled — AE adds two
// default identity keyframes (t=0, t=duration) on enable. When disabled,
// Keyframes is empty.
//
// Mirrors AE script's TimeRemapEnabled getter.
func (l *Layer) TimeRemapEnabled() bool {
	p := l.TimeRemap()
	if p == nil {
		return false
	}
	return len(p.Keyframes) > 0
}

// @summary     Enable or disable time remapping on a layer
// @description Enabling sets a static value of 0.0 (identity mapping). For
//
//	full remapping, call TimeRemap().SetStaticValue() or insert keyframes
//	after enabling. Disabling clears the static value.
//
// @param       enabled  the new time-remap enabled state
// @domain      layer-set
// @stability   alpha
// @verify      roundtrip
// @since       AE2020
// @boundary    suspected false green — setting a bare static value of 0.0
//
//	does not match how AE (and this library's TimeRemapEnabled getter)
//	detects enablement, which is "TimeRemap property carries 2 identity
//	keyframes". AE may read timeRemapEnabled back as false. Enabling
//	correctly likely requires synthesizing 2 keyframes (at the in/out
//	points) plus the enable flag, not a bare static value.
//
// @incident    layer-settimeremapenabled-needs-keyframes
// @alias       time remap,time remapping,时间重映射,帧速率重映射
func (l *Layer) SetTimeRemapEnabled(enabled bool) error {
	p := l.TimeRemap()
	if p == nil {
		return fmt.Errorf("layer %q has no TimeRemap property", l.Name)
	}
	if enabled {
		if len(p.Keyframes) > 0 {
			return nil // already enabled with keyframes
		}
		if p.StaticValue != nil {
			return nil // already enabled with static value
		}
		return p.SetStaticValue(0.0)
	}
	// Disable: clear static value.
	if len(p.Keyframes) > 0 {
		return fmt.Errorf("layer %q: delete TimeRemap keyframes before disabling", l.Name)
	}
	p.StaticValue = nil
	return nil
}

// @summary     Set a layer's per-channel audio gain
// @param       lr  the [left, right] gain in dB
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerAudio_AEShipGate_AE2020,TestLayerAudio_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving static-value write; errors when the layer
//
//	has no audio track, since the Audio Levels property is elided by
//	default and must already be materialized
//
// @alias       audio levels,audio gain,音频电平,声道增益,音量
func (l *Layer) SetAudioLevels(lr []float64) error {
	p := l.AudioLevels()
	if p == nil {
		return fmt.Errorf("layer %q: Audio Levels property not present", l.Name)
	}
	return p.SetStaticValue(lr)
}

// ── Transform group accessors (3D-only — shared 2D/3D ones in types_core.go) ──

// @summary     Set a layer's static anchor point
// @param       v  the new anchor-point components (2 for 2D, 3 for 3D)
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerXform_AEShipGate_AE2020,TestLayerXform_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving; v's length must match the property's
//
//	component count (2D or 3D); the property must already be materialized
//
// @alias       anchor point,锚点,中心点
func (l *Layer) SetAnchorPoint(v []float64) error {
	p := l.AnchorPoint()
	if p == nil {
		return fmt.Errorf("layer %q: Anchor Point property not present", l.Name)
	}
	return p.SetStaticValue(v)
}

// @summary     Set a layer's static position
// @param       v  the new position components (2 for 2D, 3 for 3D)
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerXform_AEShipGate_AE2020,TestLayerXform_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving; v's length must match the property's
//
//	component count; the property must already be materialized
//
// @alias       position,位置,坐标,移动,平移
func (l *Layer) SetPosition(v []float64) error {
	p := l.Position()
	if p == nil {
		return fmt.Errorf("layer %q: Position property not present", l.Name)
	}
	return p.SetStaticValue(v)
}

// @summary     Set a layer's static scale
// @param       v  the new scale components, normalized (1.0 = 100%)
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerXform_AEShipGate_AE2020,TestLayerXform_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving; v's length must match the property's
//
//	component count; the property must already be materialized
//
// @alias       scale,缩放,大小,尺寸
func (l *Layer) SetScale(v []float64) error {
	p := l.Scale()
	if p == nil {
		return fmt.Errorf("layer %q: Scale property not present", l.Name)
	}
	return p.SetStaticValue(v)
}

// @summary     Set a layer's Z-axis rotation
// @description Available on both 2D and 3D layers — it's the only rotation
//
//	axis 2D layers have.
//
// @param       deg  the new rotation in degrees
// @domain      layer-set
// @stability   stable
// @verify      render-pixel
// @gate        TestLayer3DRotateZ_AEShipGate_AE2020,TestLayer3DRotateZ_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving; shared by 2D and 3D layers
// @alias       rotation,旋转,Z 旋转,角度,转动
func (l *Layer) SetRotation(deg float64) error {
	return setScalarProperty(l.Rotation(), l.Name, "Rotate Z", deg)
}

// @summary     Set a layer's X-axis 3D rotation
// @param       deg  the new rotation in degrees
// @domain      layer-set
// @stability   stable
// @verify      render-pixel
// @gate        TestLayer3DRotateX_AEShipGate_AE2020,TestLayer3DRotateX_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving; 3D layers only — errors with "property not
//
//	present" on 2D layers
//
// @alias       rotate x,3D X 旋转,X 轴旋转
func (l *Layer) SetRotateX(deg float64) error {
	return setScalarProperty(l.RotateX(), l.Name, "Rotate X", deg)
}

// @summary     Set a layer's Y-axis 3D rotation
// @param       deg  the new rotation in degrees
// @domain      layer-set
// @stability   stable
// @verify      render-pixel
// @gate        TestLayer3DRotateY_AEShipGate_AE2020,TestLayer3DRotateY_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving; 3D layers only — errors with "property not
//
//	present" on 2D layers
//
// @alias       rotate y,3D Y 旋转,Y 轴旋转
func (l *Layer) SetRotateY(deg float64) error {
	return setScalarProperty(l.RotateY(), l.Name, "Rotate Y", deg)
}

// @summary     Set a layer's 3D orientation
// @param       v  the new orientation, 3 components in degrees per axis
// @domain      layer-set
// @stability   stable
// @verify      render-pixel
// @gate        TestLayer3DOrientation_AEShipGate_AE2020,TestLayer3DOrientation_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving; 3D layers only (errors on 2D layers); the
//
//	static value must be written to both cdat (little-endian) and otda
//	(big-endian), otherwise AE renders the orientation as zero
//
// @alias       orientation,方向,3D 朝向,定向
func (l *Layer) SetOrientation(v []float64) error {
	p := l.Orientation()
	if p == nil {
		return fmt.Errorf("layer %q: Orientation property not present", l.Name)
	}
	return p.SetStaticValue(v)
}

// @summary     Set a layer's opacity
// @description Callers that think in percent should divide by 100 first.
// @param       v  the new opacity, normalized 0..1 (1 = fully opaque)
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerXform_AEShipGate_AE2020,TestLayerXform_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving scalar write; the property must already be
//
//	materialized
//
// @alias       opacity,不透明度,透明度,layer opacity,淡入淡出
func (l *Layer) SetOpacity(v float64) error {
	return setScalarProperty(l.Opacity(), l.Name, "Opacity", v)
}

// @summary     Report whether a layer has an Essential Properties media-replacement slot
// @description The slot exists when AE persisted an "ADBE Layer Overrides" +
//
//	"ADBE Layer Source Alternate" pattern, typically created by calling
//	addToMotionGraphicsTemplateAs() on the source-side layer while the parent
//	composition uses the precomp as a layer. Without the slot, the
//	alternate-source setter cannot length-preservingly write a new id.
//
// @domain      meta
// @stability   stable
// @verify      roundtrip
// @since       AE2020
// @returns     true when the media-replacement slot is present
// @alias       alternate source slot,EG 媒体替换槽,essential properties
func (l *Layer) HasAlternateSourceSlot() bool {
	return l.runtime.back != nil && l.runtime.back.HasAlternateSourceSlot()
}

// AlternateSource returns the AVItem (*Composition or *Footage) used as
// this layer's media-replacement override, or nil when no override is
// active (AlternateSourceID == 0), the layer has no slot, the layer was
// built outside the parser (no owning project), or the override id
// doesn't resolve to any project AV item. Mirrors AE script's
// Property.alternateSource — except we surface the override on the
// layer rather than on a sub-property since AE persists one id per
// instance.
func (l *Layer) AlternateSource() AVItem {
	if l.AlternateSourceID == 0 {
		return nil
	}
	if l.runtime.comp == nil || l.runtime.comp.proj == nil {
		return nil
	}
	return l.runtime.comp.proj.AVItemByID(l.AlternateSourceID)
}

// LightSource returns the layer used as the environment-light source for
// this Light layer (AE 24+). Returns nil for non-light layers, lights with
// no source (sentinels 0 / 0xFFFFFFFF), or when the source ID doesn't
// resolve to a layer in the owning composition.
//
// AE stores the source layer ID in ldta @0x28 (same slot as Layer.SourceID
// for AV layers; the field is repurposed per Layer.Type). Only meaningful
// when Type == LayerTypeLight and LightKind == LightKindAmbient (the
// "Environment" light in AE 24+).
func (l *Layer) LightSource() *Layer {
	if l.Type != LayerTypeLight || l.runtime.comp == nil {
		return nil
	}
	sid := l.SourceID
	if sid == 0 || sid == codec.LightSourceUndefined {
		return nil
	}
	return l.runtime.comp.LayerByID(sid)
}

// setScalarProperty centralizes the nil-check + delegation logic shared
// by every typed scalar Set* method.
func setScalarProperty(p *Property, layerName, propName string, v float64) error {
	if p == nil {
		return fmt.Errorf("layer %q: %s property not present", layerName, propName)
	}
	return p.SetStaticValue(v)
}

// AVSource resolves this layer's source to an AVItem (Composition or
// Footage), or nil when SourceID is 0 or the layer was built outside
// the parser.
func (l *Layer) AVSource() AVItem {
	if l.SourceID == 0 || l.runtime.comp == nil || l.runtime.comp.proj == nil {
		return nil
	}
	return l.runtime.comp.proj.AVItemByID(l.SourceID)
}

// CanSetCollapseTransformation reports whether AE will let the user
// toggle CollapseTransform on this layer:
//
//   - precomp source → true
//   - solid footage source → true
//   - otherwise (file footage, placeholder, no source, text/shape/null/camera/light) → false
//
// Conservative: returns false when the source can't be resolved (layer
// built outside parser).
func (l *Layer) CanSetCollapseTransformation() bool {
	src := l.AVSource()
	if src == nil {
		return false
	}
	if _, isComp := src.(*Composition); isComp {
		return true
	}
	if f, isFootage := src.(*Footage); isFootage {
		return f.IsSolid
	}
	return false
}

// CanSetTimeRemapEnabled reports whether AE will let the user enable
// time remapping on this layer: true when the source has a non-zero
// duration. Still images, text layers, and shape sources don't qualify.
func (l *Layer) CanSetTimeRemapEnabled() bool {
	src := l.AVSource()
	if src == nil {
		return false
	}
	if c, isComp := src.(*Composition); isComp {
		return c.Duration > 0
	}
	if f, isFootage := src.(*Footage); isFootage {
		return f.Duration > 0 && !f.IsStill
	}
	return false
}

// @summary     Replace a layer's source with another AV item
// @description Internally calls the id-based source setter with the target
//
//	item's id.
//
// @param       target          the new source (a Composition or Footage)
// @param       fixExpressions  accepted for API compatibility but not
//
//	implemented — symbolic rewriting of expressions referencing the old
//	source is out of scope; when true, a warning is appended to the
//	project's warnings instead
//
// @domain      layer-set
// @stability   stable
// @verify      ae-accept
// @gate        TestLayerSource_AEShipGate_AE2020,TestLayerSource_AEShipGate_AE2025
// @since       AE2020
// @boundary    length-preserving; delegates to the id-based source setter
// @alias       replace source,替换来源,替换素材,换源,swap footage
func (l *Layer) ReplaceSource(target AVItem, fixExpressions bool) error {
	if target == nil {
		return fmt.Errorf("layer %q: target is nil", l.Name)
	}
	if fixExpressions && l.runtime.comp != nil && l.runtime.comp.proj != nil {
		l.runtime.comp.proj.recordWarning(
			fmt.Sprintf("layer %q: fixExpressions=true not implemented (expressions not auto-updated)", l.Name))
	}
	return l.SetSource(target.ItemID())
}
