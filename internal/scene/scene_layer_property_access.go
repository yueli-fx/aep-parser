package scene

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

// @summary    Set the 3D plane curvature (Advanced 3D renderer)
// @param      v  the plane curvature value (typically 0..1)
// @domain     layer-set
// @stability  stable
// @verify     roundtrip
// @since      AE2020
// @boundary   length-preserving, low risk; requires a 3D layer with the
//   Advanced 3D renderer. On solid and extruded-shape layers the Plane
//   Curvature property is not present in the Extrusion group (it appears to
//   be footage-plane specific), so no carrier was found to author against AE.
// @incident   material-advanced-props-hidden
// @alias      plane curvature,平面曲率,3D geometry
func (l *Layer) SetGeometryPlaneCurvature(v float64) error {
	return setScalarProperty(l.GeometryPlaneCurvature(), l.Name, "Plane Curvature", v)
}

// @summary    Set the 3D plane subdivision (mesh quality)
// @param      v  the mesh subdivision level (integer, default 4)
// @domain     layer-set
// @stability  stable
// @verify     roundtrip
// @since      AE2020
// @boundary   length-preserving, low risk; requires a 3D layer with the
//   Advanced 3D renderer. On solid and extruded-shape layers the Plane
//   Subdivision property is not present in the Extrusion group (it appears to
//   be footage-plane specific), so no carrier was found to author against AE.
// @incident   material-advanced-props-hidden
// @alias      plane subdivision,平面细分,mesh quality,网格质量
func (l *Layer) SetGeometryPlaneSubdivision(v float64) error {
	return setScalarProperty(l.GeometryPlaneSubdivision(), l.Name, "Plane Subdivision", v)
}

// @summary    Set the 3D bevel direction enum
// @param      v  the bevel direction enum value (default 1)
// @domain     layer-set
// @stability  stable
// @verify     roundtrip
// @since      AE2020
// @boundary   length-preserving, low risk; requires a 3D layer. The
//   Extrusion group lists this property, but AE rejects setValue on it (and
//   on Extrusion Depth) with a hidden-parent-property error on both solid and
//   extruded-shape layers in AE 2020 and AE 2025, so no carrier was found to
//   author against AE.
// @incident   material-advanced-props-hidden
// @alias      bevel direction,斜面方向,3D bevel
func (l *Layer) SetGeometryBevelDirection(v float64) error {
	return setScalarProperty(l.GeometryBevelDirection(), l.Name, "Bevel Direction", v)
}

// ── Material Options setters (3D AV layers) ──

// @summary    Set the tri-state Casts Shadows mode on a 3D AV layer
// @param      mode  the shadow mode (Off, On, or Only)
// @domain     layer-set
// @stability  stable
// @verify     render-pixel
// @gate       TestLayer3DShadow_AEShipGate_AE2020,TestLayer3DShadow_AEShipGate_AE2025
// @since      AE2020
// @boundary   equivalent to SetLightCastsShadows(bool) when called with Off
//   or On; Only is AV-3D-specific and renders shadows while hiding the layer
//   itself.
// @alias      casts shadows,投射阴影,material shadows,3D shadow
func (l *Layer) SetMaterialCastsShadows(mode MaterialCastsShadowsMode) error {
	return setScalarProperty(l.MaterialCastsShadows(), l.Name, "Casts Shadows", float64(mode))
}

// @summary    Set the material's light transmission coefficient
// @param      v  the light transmission coefficient (0..1)
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestMaterialClassic_AEShipGate_AE2020,TestMaterialClassic_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; requires a 3D layer with the
//   classic renderer.
// @alias      light transmission,光传输,透光率
func (l *Layer) SetMaterialLightTransmission(v float64) error {
	return setScalarProperty(l.MaterialLightTransmission(), l.Name, "Light Transmission", v)
}

// @summary    Toggle the material's Accepts Shadows switch
// @param      enabled  whether the layer accepts shadows from other layers
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestMaterialClassic_AEShipGate_AE2020,TestMaterialClassic_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; requires a 3D layer with the
//   classic renderer. AE elides the value on resave when it equals the
//   default of 1, verified via DOM readback.
// @alias      accepts shadows,接受阴影,receive shadows
func (l *Layer) SetMaterialAcceptsShadows(enabled bool) error {
	v := 0.0
	if enabled {
		v = 1.0
	}
	return setScalarProperty(l.MaterialAcceptsShadows(), l.Name, "Accepts Shadows", v)
}

// @summary    Toggle the material's Accepts Lights switch
// @param      enabled  whether the layer accepts lighting from light layers
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestMaterialClassic_AEShipGate_AE2020,TestMaterialClassic_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; requires a 3D layer with the
//   classic renderer. AE elides the value on resave when it equals the
//   default of 1, verified via DOM readback.
// @alias      accepts lights,接受灯光,receive lights
func (l *Layer) SetMaterialAcceptsLights(enabled bool) error {
	v := 0.0
	if enabled {
		v = 1.0
	}
	return setScalarProperty(l.MaterialAcceptsLights(), l.Name, "Accepts Lights", v)
}

// @summary    Set the material's shadow color
// @param      rgba  the shadow color as 4 RGBA components
// @domain     layer-set
// @stability  stable
// @verify     roundtrip
// @since      AE2020
// @boundary   length-preserving, low risk; requires a 3D layer. AE 2020's
//   material group does not have this property at all (only AE 2025 does),
//   and on AE 2025 the Advanced 3D group also rejects setValue with a
//   hidden-parent-property error on both solid and shape layers, so no
//   carrier with a DOM-verifiable value was found.
// @incident   material-advanced-props-hidden
// @alias      shadow color,阴影颜色,3D shadow color
func (l *Layer) SetMaterialShadowColor(rgba []float64) error {
	p := l.MaterialShadowColor()
	if p == nil {
		return fmt.Errorf("layer %q: Shadow Color property not present", l.Name)
	}
	return p.SetStaticValue(rgba)
}

// @summary    Toggle whether the layer appears in other layers' reflections
// @param      enabled  whether the layer is visible in reflective surfaces
// @domain     layer-set
// @stability  stable
// @verify     roundtrip
// @since      AE2020
// @boundary   length-preserving, low risk; requires a 3D layer. The
//   Advanced 3D group lists this property, but AE rejects setValue on it with
//   a hidden-parent-property error on both solid and extruded-shape layers in
//   AE 2020 and AE 2025, so no carrier with a DOM-verifiable value was found.
// @incident   material-advanced-props-hidden
// @alias      appears in reflections,出现在反射中,reflection visibility
func (l *Layer) SetMaterialAppearsInReflections(enabled bool) error {
	v := 0.0
	if enabled {
		v = 1.0
	}
	return setScalarProperty(l.MaterialAppearsInReflections(), l.Name, "Appears in Reflections", v)
}

// @summary    Set the material's ambient coefficient
// @param      v  the ambient coefficient (typically 0..100); not clamped
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestMaterialClassic_AEShipGate_AE2020,TestMaterialClassic_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; requires a 3D layer with the
//   classic renderer.
// @alias      ambient,环境光系数,material ambient
func (l *Layer) SetMaterialAmbient(v float64) error {
	return setScalarProperty(l.MaterialAmbient(), l.Name, "Ambient Coefficient", v)
}

// @summary    Set the material's diffuse coefficient
// @param      v  the diffuse coefficient (typically 0..100); not clamped
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestMaterialClassic_AEShipGate_AE2020,TestMaterialClassic_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; requires a 3D layer with the
//   classic renderer.
// @alias      diffuse,漫反射系数,material diffuse
func (l *Layer) SetMaterialDiffuse(v float64) error {
	return setScalarProperty(l.MaterialDiffuse(), l.Name, "Diffuse Coefficient", v)
}

// @summary    Set the material's specular coefficient
// @param      v  the specular coefficient (typically 0..100); not clamped
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestMaterialClassic_AEShipGate_AE2020,TestMaterialClassic_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; requires a 3D layer with the
//   classic renderer.
// @alias      specular,镜面反射系数,material specular
func (l *Layer) SetMaterialSpecular(v float64) error {
	return setScalarProperty(l.MaterialSpecular(), l.Name, "Specular Coefficient", v)
}

// @summary    Set the material's shininess coefficient
// @param      v  the shininess coefficient (typically 0..150); not clamped
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestMaterialClassic_AEShipGate_AE2020,TestMaterialClassic_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; requires a 3D layer with the
//   classic renderer.
// @alias      shininess,光泽度,material shininess
func (l *Layer) SetMaterialShininess(v float64) error {
	return setScalarProperty(l.MaterialShininess(), l.Name, "Shininess Coefficient", v)
}

// @summary    Set the material's metal coefficient
// @param      v  the metal coefficient (typically 0..100); not clamped
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestMaterialClassic_AEShipGate_AE2020,TestMaterialClassic_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; requires a 3D layer with the
//   classic renderer.
// @alias      metal,金属感,material metal
func (l *Layer) SetMaterialMetal(v float64) error {
	return setScalarProperty(l.MaterialMetal(), l.Name, "Metal Coefficient", v)
}

// @summary    Set the ray-traced material's reflection coefficient
// @param      v  the reflection coefficient
// @domain     layer-set
// @stability  stable
// @verify     roundtrip
// @since      AE2020
// @boundary   length-preserving, low risk; requires a 3D layer. This is a
//   ray-traced material property. The Advanced 3D group lists it (its
//   enabled flag reads true, which is misleading) but AE rejects setValue on
//   it with a hidden-parent-property error across solid and extruded-shape
//   layers on both AE 2020 and AE 2025, and no ray-traced renderer is
//   available to un-hide it, so no carrier with a DOM-verifiable value was
//   found.
// @incident   material-advanced-props-hidden
// @alias      reflection,反射率,material reflection
func (l *Layer) SetMaterialReflection(v float64) error {
	return setScalarProperty(l.MaterialReflection(), l.Name, "Reflection Coefficient", v)
}

// @summary    Set the ray-traced material's glossiness coefficient
// @param      v  the glossiness coefficient
// @domain     layer-set
// @stability  stable
// @verify     roundtrip
// @since      AE2020
// @boundary   length-preserving, low risk; requires a 3D layer. This is a
//   ray-traced material property. The Advanced 3D group lists it (its
//   enabled flag reads true, which is misleading) but AE rejects setValue on
//   it with a hidden-parent-property error across solid and extruded-shape
//   layers on both AE 2020 and AE 2025, and no ray-traced renderer is
//   available to un-hide it, so no carrier with a DOM-verifiable value was
//   found.
// @incident   material-advanced-props-hidden
// @alias      glossiness,光泽,material glossiness
func (l *Layer) SetMaterialGlossiness(v float64) error {
	return setScalarProperty(l.MaterialGlossiness(), l.Name, "Glossiness Coefficient", v)
}

// @summary    Set the ray-traced material's Fresnel coefficient
// @param      v  the Fresnel coefficient
// @domain     layer-set
// @stability  stable
// @verify     roundtrip
// @since      AE2020
// @boundary   length-preserving, low risk; requires a 3D layer. This is a
//   ray-traced material property. The Advanced 3D group lists it (its
//   enabled flag reads true, which is misleading) but AE rejects setValue on
//   it with a hidden-parent-property error across solid and extruded-shape
//   layers on both AE 2020 and AE 2025, and no ray-traced renderer is
//   available to un-hide it, so no carrier with a DOM-verifiable value was
//   found.
// @incident   material-advanced-props-hidden
// @alias      fresnel,菲涅尔,material fresnel
func (l *Layer) SetMaterialFresnel(v float64) error {
	return setScalarProperty(l.MaterialFresnel(), l.Name, "Fresnel Coefficient", v)
}

// @summary    Set the ray-traced material's transparency coefficient
// @param      v  the transparency coefficient
// @domain     layer-set
// @stability  stable
// @verify     roundtrip
// @since      AE2020
// @boundary   length-preserving, low risk; requires a 3D layer. This is a
//   ray-traced material property. The Advanced 3D group lists it (its
//   enabled flag reads true, which is misleading) but AE rejects setValue on
//   it with a hidden-parent-property error across solid and extruded-shape
//   layers on both AE 2020 and AE 2025, and no ray-traced renderer is
//   available to un-hide it, so no carrier with a DOM-verifiable value was
//   found.
// @incident   material-advanced-props-hidden
// @alias      transparency,透明度,material transparency
func (l *Layer) SetMaterialTransparency(v float64) error {
	return setScalarProperty(l.MaterialTransparency(), l.Name, "Transparency Coefficient", v)
}

// @summary    Set the ray-traced material's transparency rolloff
// @param      v  the transparency rolloff coefficient
// @domain     layer-set
// @stability  stable
// @verify     roundtrip
// @since      AE2020
// @boundary   length-preserving, low risk; requires a 3D layer. This is a
//   ray-traced material property. The Advanced 3D group lists it (its
//   enabled flag reads true, which is misleading) but AE rejects setValue on
//   it with a hidden-parent-property error across solid and extruded-shape
//   layers on both AE 2020 and AE 2025, and no ray-traced renderer is
//   available to un-hide it, so no carrier with a DOM-verifiable value was
//   found.
// @incident   material-advanced-props-hidden
// @alias      transparency rolloff,透明衰减,transp rolloff
func (l *Layer) SetMaterialTranspRolloff(v float64) error {
	return setScalarProperty(l.MaterialTranspRolloff(), l.Name, "Transp Rolloff", v)
}

// @summary    Set the ray-traced material's index of refraction
// @param      v  the index of refraction (typically 1..3+)
// @domain     layer-set
// @stability  stable
// @verify     roundtrip
// @since      AE2020
// @boundary   length-preserving, low risk; requires a 3D layer. This is a
//   ray-traced material property. The Advanced 3D group lists it (its
//   enabled flag reads true, which is misleading) but AE rejects setValue on
//   it with a hidden-parent-property error across solid and extruded-shape
//   layers on both AE 2020 and AE 2025, and no ray-traced renderer is
//   available to un-hide it, so no carrier with a DOM-verifiable value was
//   found.
// @incident   material-advanced-props-hidden
// @alias      index of refraction,折射率,IOR
func (l *Layer) SetMaterialIndexOfRefraction(v float64) error {
	return setScalarProperty(l.MaterialIndexOfRefraction(), l.Name, "Index of Refraction", v)
}

// ── Camera Layer setters ──

// @summary    Set the camera's Zoom property
// @param      v  the zoom value in pixels (equivalent to focal length times
//   a 35mm conversion)
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestNewCameraLight_AEShipGate_AE2020,TestNewCameraLight_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; camera layers only. Errors on
//   non-camera layers or when Zoom is keyframed.
// @alias      camera zoom,摄像机缩放,focal length,焦距
func (l *Layer) SetCameraZoom(v float64) error {
	return setScalarProperty(l.CameraZoom(), l.Name, "Camera Zoom", v)
}

// @summary    Toggle the camera's Depth of Field switch
// @param      enabled  whether depth of field is enabled
// @domain     layer-set
// @stability  stable
// @verify     render-pixel
// @gate       TestLayer3DDoF_AEShipGate_AE2020,TestLayer3DDoF_AEShipGate_AE2025
// @since      AE2020
// @alias      depth of field,景深,DOF,camera dof
func (l *Layer) SetCameraDepthOfField(enabled bool) error {
	v := 0.0
	if enabled {
		v = 1.0
	}
	return setScalarProperty(l.CameraDepthOfField(), l.Name, "Camera Depth of Field", v)
}

// @summary    Set the camera's Focus Distance property
// @param      v  the focus distance in pixels
// @domain     layer-set
// @stability  stable
// @verify     render-pixel
// @gate       TestLayer3DDoF_AEShipGate_AE2020,TestLayer3DDoF_AEShipGate_AE2025
// @since      AE2020
// @alias      focus distance,对焦距离,camera focus
func (l *Layer) SetCameraFocusDistance(v float64) error {
	return setScalarProperty(l.CameraFocusDistance(), l.Name, "Camera Focus Distance", v)
}

// @summary    Set the camera's Aperture property
// @param      v  the aperture value in pixels
// @domain     layer-set
// @stability  stable
// @verify     render-pixel
// @gate       TestLayer3DDoF_AEShipGate_AE2020,TestLayer3DDoF_AEShipGate_AE2025
// @since      AE2020
// @alias      aperture,光圈,camera aperture
func (l *Layer) SetCameraAperture(v float64) error {
	return setScalarProperty(l.CameraAperture(), l.Name, "Camera Aperture", v)
}

// @summary    Set the camera's Blur Level property
// @param      v  the blur level as a percentage
// @domain     layer-set
// @stability  stable
// @verify     render-pixel
// @gate       TestLayer3DDoF_AEShipGate_AE2020,TestLayer3DDoF_AEShipGate_AE2025
// @since      AE2020
// @alias      blur level,模糊强度,camera blur,DOF blur
func (l *Layer) SetCameraBlurLevel(v float64) error {
	return setScalarProperty(l.CameraBlurLevel(), l.Name, "Camera Blur Level", v)
}

// @summary    Set the camera's Iris Shape enum
// @param      v  the iris shape enum (1 = Fast Rectangle, 3..10 =
//   Triangle..Decagon)
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestNewCameraLight_AEShipGate_AE2020,TestNewCameraLight_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; camera layers only.
// @alias      iris shape,光圈形状,bokeh shape
func (l *Layer) SetIrisShape(v float64) error {
	return setScalarProperty(l.IrisShape(), l.Name, "Iris Shape", v)
}

// @summary    Set the camera's Iris Rotation property
// @param      v  the iris rotation in degrees
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestNewCameraLight_AEShipGate_AE2020,TestNewCameraLight_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; camera layers only.
// @alias      iris rotation,光圈旋转,bokeh rotation
func (l *Layer) SetIrisRotation(v float64) error {
	return setScalarProperty(l.IrisRotation(), l.Name, "Iris Rotation", v)
}

// @summary    Set the camera's Iris Roundness property
// @param      v  the iris roundness as a percentage
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestNewCameraLight_AEShipGate_AE2020,TestNewCameraLight_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; camera layers only.
// @alias      iris roundness,光圈圆度,bokeh roundness
func (l *Layer) SetIrisRoundness(v float64) error {
	return setScalarProperty(l.IrisRoundness(), l.Name, "Iris Roundness", v)
}

// @summary    Set the camera's Iris Aspect Ratio property
// @param      v  the iris aspect ratio
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestNewCameraLight_AEShipGate_AE2020,TestNewCameraLight_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; camera layers only.
// @alias      iris aspect ratio,光圈纵横比,bokeh aspect
func (l *Layer) SetIrisAspectRatio(v float64) error {
	return setScalarProperty(l.IrisAspectRatio(), l.Name, "Iris Aspect Ratio", v)
}

// @summary    Set the camera's Iris Diffraction Fringe property
// @param      v  the diffraction fringe as a percentage
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestNewCameraLight_AEShipGate_AE2020,TestNewCameraLight_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; camera layers only.
// @alias      diffraction fringe,衍射边缘,iris fringe
func (l *Layer) SetIrisDiffractionFringe(v float64) error {
	return setScalarProperty(l.IrisDiffractionFringe(), l.Name, "Iris Diffraction Fringe", v)
}

// @summary    Set the camera's Iris Highlight Gain property
// @param      v  the highlight gain value
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestNewCameraLight_AEShipGate_AE2020,TestNewCameraLight_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; camera layers only.
// @alias      highlight gain,高光增益,iris highlight gain
func (l *Layer) SetIrisHighlightGain(v float64) error {
	return setScalarProperty(l.IrisHighlightGain(), l.Name, "Iris Highlight Gain", v)
}

// @summary    Set the camera's Iris Highlight Threshold property
// @param      v  the highlight threshold as a normalized luminance (0..1)
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestNewCameraLight_AEShipGate_AE2020,TestNewCameraLight_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; camera layers only.
// @alias      highlight threshold,高光阈值,iris threshold
func (l *Layer) SetIrisHighlightThreshold(v float64) error {
	return setScalarProperty(l.IrisHighlightThreshold(), l.Name, "Iris Highlight Threshold", v)
}

// @summary    Set the camera's Iris Highlight Saturation property
// @param      v  the highlight saturation value
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestNewCameraLight_AEShipGate_AE2020,TestNewCameraLight_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; camera layers only.
// @alias      highlight saturation,高光饱和度,iris saturation
func (l *Layer) SetIrisHighlightSaturation(v float64) error {
	return setScalarProperty(l.IrisHighlightSaturation(), l.Name, "Iris Highlight Saturation", v)
}

// ── Light Layer setters ──

// @summary    Set the light's Color property
// @param      rgba  the color value; its length must match the property's
//   component count (3 for RGB, 4 for RGBA)
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestNewCameraLight_AEShipGate_AE2020,TestNewCameraLight_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; light layers only.
// @alias      light color,灯光颜色,light colour
func (l *Layer) SetLightColor(rgba []float64) error {
	p := l.LightColor()
	if p == nil {
		return fmt.Errorf("layer %q: Light Color property not present", l.Name)
	}
	return p.SetStaticValue(rgba)
}

// @summary    Set the light's Intensity property
// @param      v  the light intensity as a percentage
// @domain     layer-set
// @stability  stable
// @verify     render-pixel
// @gate       TestLayer3DLight_AEShipGate_AE2020,TestLayer3DLight_AEShipGate_AE2025
// @since      AE2020
// @alias      light intensity,灯光强度,light brightness
func (l *Layer) SetLightIntensity(v float64) error {
	return setScalarProperty(l.LightIntensity(), l.Name, "Light Intensity", v)
}

// @summary    Set the spotlight's Cone Angle property
// @param      v  the cone angle in degrees
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestNewCameraLight_AEShipGate_AE2020,TestNewCameraLight_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; spot light layers only.
// @alias      cone angle,锥角,spotlight cone,聚光灯角度
func (l *Layer) SetLightConeAngle(v float64) error {
	return setScalarProperty(l.LightConeAngle(), l.Name, "Light Cone Angle", v)
}

// @summary    Set the spotlight's Cone Feather property
// @param      v  the cone feather as a percentage
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestNewCameraLight_AEShipGate_AE2020,TestNewCameraLight_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; spot light layers only.
// @alias      cone feather,锥形羽化,spotlight feather
func (l *Layer) SetLightConeFeather(v float64) error {
	return setScalarProperty(l.LightConeFeather(), l.Name, "Light Cone Feather", v)
}

// @summary    Set the light's Falloff Type enum
// @param      v  the falloff type enum value
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestNewCameraLight_AEShipGate_AE2020,TestNewCameraLight_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; light layers only.
// @alias      falloff type,衰减类型,light falloff
func (l *Layer) SetLightFalloffType(v float64) error {
	return setScalarProperty(l.LightFalloffType(), l.Name, "Light Falloff Type", v)
}

// @summary    Set the light's Falloff Start distance property
// @param      v  the falloff start distance in pixels
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestNewCameraLight_AEShipGate_AE2020,TestNewCameraLight_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; light layers only. Only
//   meaningful when the falloff type is not None.
// @alias      falloff start,衰减起始距离,light falloff start
func (l *Layer) SetLightFalloffStart(v float64) error {
	return setScalarProperty(l.LightFalloffStart(), l.Name, "Light Falloff Start", v)
}

// @summary    Set the light's Falloff Distance property
// @param      v  the falloff distance in pixels
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestNewCameraLight_AEShipGate_AE2020,TestNewCameraLight_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; light layers only.
// @alias      falloff distance,衰减距离,light falloff distance
func (l *Layer) SetLightFalloffDistance(v float64) error {
	return setScalarProperty(l.LightFalloffDistance(), l.Name, "Light Falloff Distance", v)
}

// @summary    Toggle the light's Casts Shadows switch
// @param      enabled  whether the light casts shadows
// @domain     layer-set
// @stability  stable
// @verify     render-pixel
// @gate       TestLayer3DShadow_AEShipGate_AE2020,TestLayer3DShadow_AEShipGate_AE2025
// @since      AE2020
// @alias      light casts shadows,灯光投射阴影,enable shadow,开启阴影
func (l *Layer) SetLightCastsShadows(enabled bool) error {
	v := 0.0
	if enabled {
		v = 1.0
	}
	return setScalarProperty(l.LightCastsShadows(), l.Name, "Casts Shadows", v)
}

// @summary    Set the light's Shadow Darkness property
// @param      v  the shadow darkness as a percentage
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestNewCameraLight_AEShipGate_AE2020,TestNewCameraLight_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; light layers only.
// @alias      shadow darkness,阴影暗度,light shadow darkness
func (l *Layer) SetLightShadowDarkness(v float64) error {
	return setScalarProperty(l.LightShadowDarkness(), l.Name, "Light Shadow Darkness", v)
}

// @summary    Set the light's Shadow Diffusion property
// @param      v  the shadow diffusion in pixels
// @domain     layer-set
// @stability  stable
// @verify     ae-accept
// @gate       TestNewCameraLight_AEShipGate_AE2020,TestNewCameraLight_AEShipGate_AE2025
// @since      AE2020
// @boundary   length-preserving, low risk; light layers only.
// @alias      shadow diffusion,阴影扩散,light shadow diffusion,soft shadow
func (l *Layer) SetLightShadowDiffusion(v float64) error {
	return setScalarProperty(l.LightShadowDiffusion(), l.Name, "Light Shadow Diffusion", v)
}
