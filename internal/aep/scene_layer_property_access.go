package aep

import "fmt"

// TimeRemap returns the layer's Time Remap property (1D, seconds) when
// time remapping is enabled. Returns the property slot even when time
// remap is disabled — check TimeRemapEnabled() for the on/off bit.
// To toggle time remap on / off requires structural property-tree
// changes (AE adds 2 default keyframes); not currently supported.
func (l *Layer) TimeRemap() *Property {
	return l.PropertyByMatchName(MatchNameTimeRemap)
}

// AudioLevels returns the layer's Audio Levels property (2D, [left, right]
// channel levels in dB). Present on layers with audio content; nil
// otherwise.
func (l *Layer) AudioLevels() *Property {
	return l.PropertyByMatchName(MatchNameAudioLevels)
}

// ── Geometry Options property getters (3D AV layers) ──

// GeometryPlaneCurvature returns the Plane Curvature property (1D, AE
// 24+ Advanced 3D renderer; for the "Curved" plane preset on a 3D solid).
func (l *Layer) GeometryPlaneCurvature() *Property {
	return l.PropertyByMatchName(MatchNameGeometryPlaneCurvature)
}

// GeometryPlaneSubdivision returns the Plane Subdivision property (1D
// int, default 4; mesh quality for the curved plane).
func (l *Layer) GeometryPlaneSubdivision() *Property {
	return l.PropertyByMatchName(MatchNameGeometryPlaneSubdivision)
}

// GeometryBevelDirection returns the Bevel Direction property (1D enum,
// default 1; relevant for shape/text 3D layers with bevel — present on
// solid layers as a stub).
func (l *Layer) GeometryBevelDirection() *Property {
	return l.PropertyByMatchName(MatchNameGeometryBevelDirection)
}

// ── Material Options property getters (3D AV layers) ──

// MaterialCastsShadows returns the (shared with Light) Casts Shadows
// property. On 3D AV layers it carries a tri-state (Off/On/Only); on
// light layers it's bool. Same property as LightCastsShadows().
func (l *Layer) MaterialCastsShadows() *Property {
	return l.PropertyByMatchName(MatchNameMaterialCastsShadows)
}
func (l *Layer) MaterialLightTransmission() *Property {
	return l.PropertyByMatchName(MatchNameMaterialLightTransmission)
}
func (l *Layer) MaterialAcceptsShadows() *Property {
	return l.PropertyByMatchName(MatchNameMaterialAcceptsShadows)
}
func (l *Layer) MaterialAcceptsLights() *Property {
	return l.PropertyByMatchName(MatchNameMaterialAcceptsLights)
}
func (l *Layer) MaterialShadowColor() *Property {
	return l.PropertyByMatchName(MatchNameMaterialShadowColor)
}
func (l *Layer) MaterialAppearsInReflections() *Property {
	return l.PropertyByMatchName(MatchNameMaterialAppearsInRefl)
}
func (l *Layer) MaterialAmbient() *Property {
	return l.PropertyByMatchName(MatchNameMaterialAmbient)
}
func (l *Layer) MaterialDiffuse() *Property {
	return l.PropertyByMatchName(MatchNameMaterialDiffuse)
}
func (l *Layer) MaterialSpecular() *Property {
	return l.PropertyByMatchName(MatchNameMaterialSpecular)
}
func (l *Layer) MaterialShininess() *Property {
	return l.PropertyByMatchName(MatchNameMaterialShininess)
}
func (l *Layer) MaterialMetal() *Property {
	return l.PropertyByMatchName(MatchNameMaterialMetal)
}
func (l *Layer) MaterialReflection() *Property {
	return l.PropertyByMatchName(MatchNameMaterialReflection)
}
func (l *Layer) MaterialGlossiness() *Property {
	return l.PropertyByMatchName(MatchNameMaterialGlossiness)
}
func (l *Layer) MaterialFresnel() *Property {
	return l.PropertyByMatchName(MatchNameMaterialFresnel)
}
func (l *Layer) MaterialTransparency() *Property {
	return l.PropertyByMatchName(MatchNameMaterialTransparency)
}
func (l *Layer) MaterialTranspRolloff() *Property {
	return l.PropertyByMatchName(MatchNameMaterialTranspRolloff)
}
func (l *Layer) MaterialIndexOfRefraction() *Property {
	return l.PropertyByMatchName(MatchNameMaterialIndexOfRefraction)
}

// ── Transform group property getters (3D-only) ──

// RotateX / RotateY are the per-axis 3D rotation properties (degrees).
// Present only when Layer.Is3D; nil otherwise.
func (l *Layer) RotateX() *Property {
	return l.PropertyByMatchName(MatchNameRotateX)
}
func (l *Layer) RotateY() *Property {
	return l.PropertyByMatchName(MatchNameRotateY)
}

// Orientation returns the layer's 3D Orientation property (3 components,
// degrees per axis — separate from the animatable per-axis RotateX/Y/Z).
// Present only when Layer.Is3D.
func (l *Layer) Orientation() *Property {
	return l.PropertyByMatchName(MatchNameOrientation)
}

// ── Camera Layer property getters ──

// CameraZoom returns the camera's Zoom property (1D, pixels — AE's
// "Zoom" field; equivalent to focal length × 35mm conversion). Nil for
// non-camera layers.
func (l *Layer) CameraZoom() *Property {
	return l.PropertyByMatchName(MatchNameCameraZoom)
}

// CameraDepthOfField returns the camera's Depth of Field toggle
// (1D bool-as-float, 0 = off, 1 = on). Nil for non-camera layers.
func (l *Layer) CameraDepthOfField() *Property {
	return l.PropertyByMatchName(MatchNameCameraDepthOfField)
}

// CameraFocusDistance returns the camera's Focus Distance property
// (1D, pixels). Nil for non-camera layers.
func (l *Layer) CameraFocusDistance() *Property {
	return l.PropertyByMatchName(MatchNameCameraFocusDistance)
}

// CameraAperture returns the camera's Aperture property (1D, pixels).
// Nil for non-camera layers.
func (l *Layer) CameraAperture() *Property {
	return l.PropertyByMatchName(MatchNameCameraAperture)
}

// CameraBlurLevel returns the camera's Blur Level property (1D, %).
// Nil for non-camera layers.
func (l *Layer) CameraBlurLevel() *Property {
	return l.PropertyByMatchName(MatchNameCameraBlurLevel)
}

// IrisShape returns the camera's Iris Shape menu property (1D, enum:
// 1=Fast Rectangle, 3..10 = Triangle..Decagon). Nil for non-camera layers.
func (l *Layer) IrisShape() *Property {
	return l.PropertyByMatchName(MatchNameCameraIrisShape)
}

// IrisRotation returns the camera's Iris Rotation property (1D, degrees).
func (l *Layer) IrisRotation() *Property {
	return l.PropertyByMatchName(MatchNameCameraIrisRotation)
}

// IrisRoundness returns the camera's Iris Roundness property (1D, %).
func (l *Layer) IrisRoundness() *Property {
	return l.PropertyByMatchName(MatchNameCameraIrisRoundness)
}

// IrisAspectRatio returns the camera's Iris Aspect Ratio property (1D).
func (l *Layer) IrisAspectRatio() *Property {
	return l.PropertyByMatchName(MatchNameCameraIrisAspectRatio)
}

// IrisDiffractionFringe returns the camera's Iris Diffraction Fringe
// property (1D, %).
func (l *Layer) IrisDiffractionFringe() *Property {
	return l.PropertyByMatchName(MatchNameCameraIrisDiffractionFringe)
}

// IrisHighlightGain returns the camera's Iris Highlight Gain property (1D).
func (l *Layer) IrisHighlightGain() *Property {
	return l.PropertyByMatchName(MatchNameCameraIrisHighlightGain)
}

// IrisHighlightThreshold returns the camera's Iris Highlight Threshold
// property (1D, 0..1 normalized luminance).
func (l *Layer) IrisHighlightThreshold() *Property {
	return l.PropertyByMatchName(MatchNameCameraIrisHighlightThreshold)
}

// IrisHighlightSaturation returns the camera's Iris Highlight Saturation
// property (1D).
func (l *Layer) IrisHighlightSaturation() *Property {
	return l.PropertyByMatchName(MatchNameCameraIrisHighlightSaturation)
}

// ── Light Layer property getters ──

// LightType returns the light's Type menu property if present in the
// property tree. NOTE: AE typically stores the light type at the
// layer level (similar to LayerType for camera vs light disambig)
// rather than as a regular property, so this accessor often returns
// nil for real-world projects — the type field's exact ldta location
// is not RE'd yet. Kept for future use / older AE versions that may
// have surfaced it as a property.
func (l *Layer) LightType() *Property {
	return l.PropertyByMatchName(MatchNameLightType)
}

// LightColor returns the light's Color property (4D color or 3D RGB
// depending on AE version — read Components to disambiguate). Nil
// for non-light layers.
func (l *Layer) LightColor() *Property {
	return l.PropertyByMatchName(MatchNameLightColor)
}

// LightIntensity returns the light's Intensity property (1D, %). Nil
// for non-light layers.
func (l *Layer) LightIntensity() *Property {
	return l.PropertyByMatchName(MatchNameLightIntensity)
}

// LightConeAngle returns the spotlight cone-angle property (1D,
// degrees). Only meaningful for spot lights. Nil otherwise.
func (l *Layer) LightConeAngle() *Property {
	return l.PropertyByMatchName(MatchNameLightConeAngle)
}

// LightConeFeather returns the spotlight cone-feather property (1D, %).
// Only meaningful for spot lights.
func (l *Layer) LightConeFeather() *Property {
	return l.PropertyByMatchName(MatchNameLightConeFeather)
}

// LightFalloffType returns the falloff-type menu property (1D enum).
func (l *Layer) LightFalloffType() *Property {
	return l.PropertyByMatchName(MatchNameLightFalloffType)
}

// LightFalloffStart returns the falloff-start distance property (1D,
// pixels). Only meaningful when LightFalloffType != None.
func (l *Layer) LightFalloffStart() *Property {
	return l.PropertyByMatchName(MatchNameLightFalloffStart)
}

// LightFalloffDistance returns the falloff distance property (1D,
// pixels).
func (l *Layer) LightFalloffDistance() *Property {
	return l.PropertyByMatchName(MatchNameLightFalloffDistance)
}

// LightCastsShadows returns the Casts Shadows toggle if present in the
// property tree. NOTE: AE may store this as a checkbox flag at the
// layer level rather than a property, so this accessor returns nil for
// many real-world projects. The exact ldta location is not RE'd yet.
func (l *Layer) LightCastsShadows() *Property {
	return l.PropertyByMatchName(MatchNameLightCastsShadows)
}

// LightShadowDarkness returns the Shadow Darkness property (1D, %).
func (l *Layer) LightShadowDarkness() *Property {
	return l.PropertyByMatchName(MatchNameLightShadowDarkness)
}

// LightShadowDiffusion returns the Shadow Diffusion property (1D,
// pixels).
func (l *Layer) LightShadowDiffusion() *Property {
	return l.PropertyByMatchName(MatchNameLightShadowDiffusion)
}

// ── Geometry Options setters (3D AV layers) ──

// SetGeometryPlaneCurvature writes Plane Curvature (typically 0..1).
func (l *Layer) SetGeometryPlaneCurvature(v float64) error {
	return setScalarProperty(l.GeometryPlaneCurvature(), l.Name, "Plane Curvature", v)
}

// SetGeometryPlaneSubdivision writes Plane Subdivision (integer mesh quality).
func (l *Layer) SetGeometryPlaneSubdivision(v float64) error {
	return setScalarProperty(l.GeometryPlaneSubdivision(), l.Name, "Plane Subdivision", v)
}

// SetGeometryBevelDirection writes Bevel Direction enum.
func (l *Layer) SetGeometryBevelDirection(v float64) error {
	return setScalarProperty(l.GeometryBevelDirection(), l.Name, "Bevel Direction", v)
}

// ── Material Options setters (3D AV layers) ──

// SetMaterialCastsShadows writes the tri-state Casts Shadows enum
// (Off/On/Only). Equivalent to SetLightCastsShadows(bool) when called
// with Off/On; "Only" is AV-3D-specific (renders shadows but hides the
// layer itself).
func (l *Layer) SetMaterialCastsShadows(mode MaterialCastsShadowsMode) error {
	return setScalarProperty(l.MaterialCastsShadows(), l.Name, "Casts Shadows", float64(mode))
}

// SetMaterialLightTransmission writes light transmission (0..1).
func (l *Layer) SetMaterialLightTransmission(v float64) error {
	return setScalarProperty(l.MaterialLightTransmission(), l.Name, "Light Transmission", v)
}

// SetMaterialAcceptsShadows toggles the Accepts Shadows switch.
func (l *Layer) SetMaterialAcceptsShadows(enabled bool) error {
	v := 0.0
	if enabled {
		v = 1.0
	}
	return setScalarProperty(l.MaterialAcceptsShadows(), l.Name, "Accepts Shadows", v)
}

// SetMaterialAcceptsLights toggles the Accepts Lights switch.
func (l *Layer) SetMaterialAcceptsLights(enabled bool) error {
	v := 0.0
	if enabled {
		v = 1.0
	}
	return setScalarProperty(l.MaterialAcceptsLights(), l.Name, "Accepts Lights", v)
}

// SetMaterialShadowColor writes the shadow color (4-component RGBA).
func (l *Layer) SetMaterialShadowColor(rgba []float64) error {
	p := l.MaterialShadowColor()
	if p == nil {
		return fmt.Errorf("layer %q: Shadow Color property not present", l.Name)
	}
	return p.SetStaticValue(rgba)
}

// SetMaterialAppearsInReflections toggles whether this layer appears in
// reflective surfaces of other 3D layers.
func (l *Layer) SetMaterialAppearsInReflections(enabled bool) error {
	v := 0.0
	if enabled {
		v = 1.0
	}
	return setScalarProperty(l.MaterialAppearsInReflections(), l.Name, "Appears in Reflections", v)
}

// SetMaterialAmbient / Diffuse / Specular / Shininess / Metal / Reflection /
// Glossiness / Fresnel / Transparency / TranspRolloff / IndexOfRefraction
// — physically-based material coefficients. All scalar; range varies by
// field (Ambient/Diffuse/Specular/Metal/Reflection typically 0..1 or 0..100,
// Shininess 0..150, IndexOfRefraction 1..3+). No range clamping is done;
// AE will accept any float.
func (l *Layer) SetMaterialAmbient(v float64) error {
	return setScalarProperty(l.MaterialAmbient(), l.Name, "Ambient Coefficient", v)
}
func (l *Layer) SetMaterialDiffuse(v float64) error {
	return setScalarProperty(l.MaterialDiffuse(), l.Name, "Diffuse Coefficient", v)
}
func (l *Layer) SetMaterialSpecular(v float64) error {
	return setScalarProperty(l.MaterialSpecular(), l.Name, "Specular Coefficient", v)
}
func (l *Layer) SetMaterialShininess(v float64) error {
	return setScalarProperty(l.MaterialShininess(), l.Name, "Shininess Coefficient", v)
}
func (l *Layer) SetMaterialMetal(v float64) error {
	return setScalarProperty(l.MaterialMetal(), l.Name, "Metal Coefficient", v)
}
func (l *Layer) SetMaterialReflection(v float64) error {
	return setScalarProperty(l.MaterialReflection(), l.Name, "Reflection Coefficient", v)
}
func (l *Layer) SetMaterialGlossiness(v float64) error {
	return setScalarProperty(l.MaterialGlossiness(), l.Name, "Glossiness Coefficient", v)
}
func (l *Layer) SetMaterialFresnel(v float64) error {
	return setScalarProperty(l.MaterialFresnel(), l.Name, "Fresnel Coefficient", v)
}
func (l *Layer) SetMaterialTransparency(v float64) error {
	return setScalarProperty(l.MaterialTransparency(), l.Name, "Transparency Coefficient", v)
}
func (l *Layer) SetMaterialTranspRolloff(v float64) error {
	return setScalarProperty(l.MaterialTranspRolloff(), l.Name, "Transp Rolloff", v)
}
func (l *Layer) SetMaterialIndexOfRefraction(v float64) error {
	return setScalarProperty(l.MaterialIndexOfRefraction(), l.Name, "Index of Refraction", v)
}

// ── Camera Layer setters ──

// SetCameraZoom writes a new static value to the camera's Zoom property
// (pixels). Errors on non-camera layers or keyframed Zoom.
func (l *Layer) SetCameraZoom(v float64) error {
	return setScalarProperty(l.CameraZoom(), l.Name, "Camera Zoom", v)
}

// SetCameraDepthOfField toggles the Depth-of-Field switch (true=1, false=0).
func (l *Layer) SetCameraDepthOfField(enabled bool) error {
	v := 0.0
	if enabled {
		v = 1.0
	}
	return setScalarProperty(l.CameraDepthOfField(), l.Name, "Camera Depth of Field", v)
}

// SetCameraFocusDistance writes the focus-distance property (pixels).
func (l *Layer) SetCameraFocusDistance(v float64) error {
	return setScalarProperty(l.CameraFocusDistance(), l.Name, "Camera Focus Distance", v)
}

// SetCameraAperture writes the aperture property (pixels).
func (l *Layer) SetCameraAperture(v float64) error {
	return setScalarProperty(l.CameraAperture(), l.Name, "Camera Aperture", v)
}

// SetCameraBlurLevel writes the blur-level property (%).
func (l *Layer) SetCameraBlurLevel(v float64) error {
	return setScalarProperty(l.CameraBlurLevel(), l.Name, "Camera Blur Level", v)
}

// SetIrisShape writes the Iris Shape enum (1=Fast Rect, 3..10 = Triangle..Decagon).
func (l *Layer) SetIrisShape(v float64) error {
	return setScalarProperty(l.IrisShape(), l.Name, "Iris Shape", v)
}

// SetIrisRotation writes the Iris Rotation (degrees).
func (l *Layer) SetIrisRotation(v float64) error {
	return setScalarProperty(l.IrisRotation(), l.Name, "Iris Rotation", v)
}

// SetIrisRoundness writes the Iris Roundness (%).
func (l *Layer) SetIrisRoundness(v float64) error {
	return setScalarProperty(l.IrisRoundness(), l.Name, "Iris Roundness", v)
}

// SetIrisAspectRatio writes the Iris Aspect Ratio.
func (l *Layer) SetIrisAspectRatio(v float64) error {
	return setScalarProperty(l.IrisAspectRatio(), l.Name, "Iris Aspect Ratio", v)
}

// SetIrisDiffractionFringe writes the Iris Diffraction Fringe (%).
func (l *Layer) SetIrisDiffractionFringe(v float64) error {
	return setScalarProperty(l.IrisDiffractionFringe(), l.Name, "Iris Diffraction Fringe", v)
}

// SetIrisHighlightGain writes the Iris Highlight Gain.
func (l *Layer) SetIrisHighlightGain(v float64) error {
	return setScalarProperty(l.IrisHighlightGain(), l.Name, "Iris Highlight Gain", v)
}

// SetIrisHighlightThreshold writes the Iris Highlight Threshold (0..1).
func (l *Layer) SetIrisHighlightThreshold(v float64) error {
	return setScalarProperty(l.IrisHighlightThreshold(), l.Name, "Iris Highlight Threshold", v)
}

// SetIrisHighlightSaturation writes the Iris Highlight Saturation.
func (l *Layer) SetIrisHighlightSaturation(v float64) error {
	return setScalarProperty(l.IrisHighlightSaturation(), l.Name, "Iris Highlight Saturation", v)
}

// ── Light Layer setters ──

// SetLightColor writes the light's Color property. The value's length
// must match the property's component count (3 for RGB, 4 for RGBA);
// SetStaticValue enforces this.
func (l *Layer) SetLightColor(rgba []float64) error {
	p := l.LightColor()
	if p == nil {
		return fmt.Errorf("layer %q: Light Color property not present", l.Name)
	}
	return p.SetStaticValue(rgba)
}

// SetLightIntensity writes the light intensity (%).
func (l *Layer) SetLightIntensity(v float64) error {
	return setScalarProperty(l.LightIntensity(), l.Name, "Light Intensity", v)
}

// SetLightConeAngle writes the spotlight cone angle (degrees).
func (l *Layer) SetLightConeAngle(v float64) error {
	return setScalarProperty(l.LightConeAngle(), l.Name, "Light Cone Angle", v)
}

// SetLightConeFeather writes the spotlight cone-feather (%).
func (l *Layer) SetLightConeFeather(v float64) error {
	return setScalarProperty(l.LightConeFeather(), l.Name, "Light Cone Feather", v)
}

// SetLightFalloffType writes the falloff type enum.
func (l *Layer) SetLightFalloffType(v float64) error {
	return setScalarProperty(l.LightFalloffType(), l.Name, "Light Falloff Type", v)
}

// SetLightFalloffStart writes the falloff-start distance (pixels).
func (l *Layer) SetLightFalloffStart(v float64) error {
	return setScalarProperty(l.LightFalloffStart(), l.Name, "Light Falloff Start", v)
}

// SetLightFalloffDistance writes the falloff distance (pixels).
func (l *Layer) SetLightFalloffDistance(v float64) error {
	return setScalarProperty(l.LightFalloffDistance(), l.Name, "Light Falloff Distance", v)
}

// SetLightCastsShadows toggles the Casts Shadows switch (true=1, false=0).
func (l *Layer) SetLightCastsShadows(enabled bool) error {
	v := 0.0
	if enabled {
		v = 1.0
	}
	return setScalarProperty(l.LightCastsShadows(), l.Name, "Casts Shadows", v)
}

// SetLightShadowDarkness writes the shadow darkness (%).
func (l *Layer) SetLightShadowDarkness(v float64) error {
	return setScalarProperty(l.LightShadowDarkness(), l.Name, "Light Shadow Darkness", v)
}

// SetLightShadowDiffusion writes the shadow diffusion (pixels).
func (l *Layer) SetLightShadowDiffusion(v float64) error {
	return setScalarProperty(l.LightShadowDiffusion(), l.Name, "Light Shadow Diffusion", v)
}
