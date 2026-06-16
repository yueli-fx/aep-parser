// Public-API entry for adding an effect to a layer's Effect Parade.
//
// AE stores each effect as a (tdmn match-name, LIST:sspc payload) pair inside
// the layer's "ADBE Effect Parade" tdgp group, terminated by a lone
// "ADBE Group End" tdmn sentinel. Adding an effect = splice a fresh
// (tdmn, sspc) pair in just before that sentinel. The sspc payload (parameter
// tree, pard metadata, built-in-params group) is supplied verbatim from an
// embedded AE-native template — the same embed-AE-bytes strategy the V2.2 shape
// bodies use, and the same (tdmn, payload) splice DuplicatePropertyGroup is
// ship-gate-green with.
//
// Atomic: snapshot parade chunk + scene children + flat Effects slice, commit,
// re-parse the spliced pair to obtain a back-ref-correct *Effect, roll back on
// any parser warning. LIST sizes are recomputed bottom-up by rifx.Chunk.Write,
// so the byte-length growth needs no manual fixup (same as Footage.SetPath /
// gradient writes).
//
// Layers without an Effect Parade (AE only emits the parade once ≥1 effect
// exists, so every effect-less layer lacks it) get an empty parade spliced into
// their Layr property tree first — immediately before "ADBE Transform Group",
// matching AE's emitted group order (RE: re_shape_effect.aep +
// re_effect_library.aep). From-scratch layers built by NewShapeLayer have no
// parsed property tree to splice into (and dirty shape layers are re-lowered
// from scene state at write, discarding chunk edits) — AddEffect refuses those;
// aep.Reopen upgrades them to parsed layers.
package serializer

import (
	"bytes"
	"embed"
	"encoding/binary"
	"fmt"
	"sort"
	"sync"

	"github.com/example/aep-parser/internal/rifx"
	"github.com/example/aep-parser/internal/scene"
)

// effectTemplateFS holds the embedded effect templates. Each is a LIST(tdgp)
// wrapper around an effect's (tdmn match-name, LIST:sspc payload) pair,
// extracted verbatim from an AE-2020-saved layer that had the full curated set
// applied (test_data/re_effect_library.aep + re_effect_library2.aep, generated
// by re_effect_library.jsx / re_effect_library2.jsx; Point3D Control from
// re_effect_param_types.aep — all extracted by tmp_debug/extract_effect_lib).
// All are built-in effects present well before the 2020 read floor; their
// serialized form is version-portable (AE 2025 accepts the AE-2020 bytes —
// mirroring the gradient version-portability finding, confirmed by ship-gate
// across the full template library on both versions).
//
//go:embed templates/effect_adbe_gaussian_blur_2.bin templates/effect_adbe_fill.bin templates/effect_adbe_tint.bin templates/effect_adbe_brightness_contrast_2.bin templates/effect_adbe_tritone.bin templates/effect_adbe_easy_levels2.bin templates/effect_adbe_pro_levels2.bin templates/effect_adbe_hue_saturation.bin templates/effect_adbe_box_blur.bin templates/effect_adbe_glo2.bin templates/effect_adbe_invert.bin templates/effect_adbe_exposure2.bin
//go:embed templates/effect_adbe_drop_shadow.bin templates/effect_adbe_sharpen.bin templates/effect_adbe_mosaic.bin templates/effect_adbe_noise.bin templates/effect_adbe_geometry2.bin templates/effect_adbe_ramp.bin templates/effect_adbe_fractal_noise.bin templates/effect_adbe_tile.bin templates/effect_adbe_motion_blur.bin templates/effect_adbe_linear_wipe.bin templates/effect_adbe_wave_warp.bin templates/effect_adbe_curvescustom.bin templates/effect_adbe_slider_control.bin templates/effect_adbe_point_control.bin templates/effect_adbe_color_control.bin templates/effect_adbe_angle_control.bin templates/effect_adbe_checkbox_control.bin templates/effect_adbe_point3d_control.bin
//go:embed templates/effect_adbe_set_matte3.bin
//go:embed templates/effect_adbe_turbulent_displace.bin templates/effect_adbe_roughen_edges.bin templates/effect_adbe_echo.bin templates/effect_adbe_radial_blur.bin templates/effect_adbe_4colorgradient.bin templates/effect_adbe_checkerboard.bin templates/effect_adbe_grid.bin templates/effect_adbe_stroke.bin templates/effect_adbe_corner_pin.bin templates/effect_adbe_venetian_blinds.bin
//go:embed templates/effect_adbe_twirl.bin templates/effect_adbe_polar_coordinates.bin templates/effect_adbe_spherize.bin templates/effect_adbe_magnify.bin templates/effect_adbe_ripple.bin templates/effect_adbe_optics_compensation.bin templates/effect_adbe_posterize.bin templates/effect_adbe_threshold2.bin templates/effect_adbe_find_edges.bin templates/effect_adbe_color_emboss.bin templates/effect_adbe_emboss.bin templates/effect_adbe_strobe.bin templates/effect_adbe_brush_strokes.bin templates/effect_adbe_bevel_alpha.bin templates/effect_adbe_bevel_edges.bin templates/effect_adbe_photo_filter.bin templates/effect_adbe_vibrance.bin templates/effect_adbe_color_balance_2.bin templates/effect_adbe_color_balance_hls.bin
//go:embed templates/effect_adbe_black_white.bin templates/effect_adbe_gamma_pedestal_gain2.bin templates/effect_adbe_channel_blur.bin templates/effect_adbe_bilateral.bin templates/effect_adbe_smart_blur.bin templates/effect_adbe_unsharp_mask2.bin templates/effect_adbe_shift_channels.bin templates/effect_adbe_solid_composite.bin templates/effect_adbe_minimax.bin templates/effect_adbe_arithmetic.bin templates/effect_adbe_circle.bin templates/effect_adbe_lens_flare.bin templates/effect_adbe_cell_pattern.bin templates/effect_adbe_lightning_2.bin templates/effect_adbe_laser.bin templates/effect_adbe_paint_bucket.bin templates/effect_adbe_posterize_time.bin templates/effect_adbe_simple_choker.bin templates/effect_adbe_matte_choker.bin
//go:embed templates/effect_adbe_bulge.bin templates/effect_adbe_offset.bin templates/effect_adbe_mirror.bin templates/effect_adbe_fractal.bin templates/effect_adbe_write_on.bin templates/effect_adbe_scribble_fill.bin templates/effect_adbe_eyedropper_fill.bin templates/effect_adbe_audspect.bin templates/effect_adbe_audwave.bin templates/effect_adbe_autolevels.bin templates/effect_adbe_autocolor.bin templates/effect_adbe_autocontrast.bin
//go:embed templates/effect_adbe_equalize.bin templates/effect_adbe_leave_color.bin templates/effect_adbe_change_to_color.bin templates/effect_adbe_change_color.bin templates/effect_adbe_radial_shadow.bin templates/effect_adbe_remove_color_matting.bin templates/effect_adbe_dust_scratches.bin templates/effect_adbe_noise_alpha2.bin templates/effect_adbe_noise_hls2.bin templates/effect_adbe_radial_wipe.bin templates/effect_adbe_block_dissolve.bin
//go:embed templates/effect_adbe_lumetri.bin templates/effect_adbe_lightning.bin templates/effect_cc_radial_fast_blur.bin templates/effect_cc_radial_blur.bin templates/effect_cs_crossblur.bin templates/effect_cc_bend_it.bin templates/effect_cc_bender.bin templates/effect_cc_blobbylize.bin templates/effect_cc_flo_motion.bin templates/effect_cc_griddler.bin templates/effect_cc_lens.bin templates/effect_cc_page_turn.bin templates/effect_cc_power_pin.bin templates/effect_cc_ripple_pulse.bin templates/effect_cc_slant.bin
//go:embed templates/effect_cc_smear.bin templates/effect_cc_split.bin templates/effect_cc_split_2.bin templates/effect_cc_tiler.bin templates/effect_cc_warpomatic.bin templates/effect_cc_light_burst_2_5.bin templates/effect_cc_light_rays.bin templates/effect_cc_light_sweep.bin templates/effect_cs_threads.bin templates/effect_cc_cylinder.bin templates/effect_cc_sphere.bin templates/effect_cc_spotlight.bin templates/effect_cc_glass.bin templates/effect_cs_hextile.bin templates/effect_cc_kaleida.bin
//go:embed templates/effect_cc_mr_smoothie.bin templates/effect_cc_plastic.bin templates/effect_cc_repetile.bin templates/effect_cc_threshold.bin templates/effect_cc_threshold_rgb.bin templates/effect_cc_pixel_polly.bin templates/effect_cc_scatterize.bin templates/effect_cc_star_burst.bin templates/effect_cc_force_motion_blur.bin templates/effect_cc_wide_time.bin templates/effect_cc_color_offset.bin templates/effect_cc_toner.bin templates/effect_cc_burn_film.bin templates/effect_cs_vignette.bin templates/effect_cc_simple_wire_removal.bin
//go:embed templates/effect_adbe_displacement_map.bin templates/effect_adbe_compound_blur.bin templates/effect_cc_vector_blur.bin
//go:embed templates/effect_adbe_basic_3d.bin templates/effect_adbe_broadcast_colors.bin templates/effect_adbe_channel_combiner.bin templates/effect_adbe_cineon_converter2.bin templates/effect_adbe_color_key.bin templates/effect_adbe_color_range.bin templates/effect_adbe_extract.bin templates/effect_adbe_geometry.bin templates/effect_adbe_gradient_wipe.bin templates/effect_adbe_grow_bounds.bin templates/effect_adbe_keycleaner.bin templates/effect_adbe_layer_control.bin
//go:embed templates/effect_adbe_luma_key.bin templates/effect_adbe_median.bin templates/effect_adbe_noise_hls_auto2.bin templates/effect_adbe_profiletoprofile.bin templates/effect_adbe_time_displacement.bin templates/effect_adbe_timecode.bin templates/effect_cc_ball_action.bin templates/effect_cc_bubbles.bin templates/effect_cc_composite.bin templates/effect_cc_drizzle.bin templates/effect_cc_environment.bin templates/effect_cc_glass_wipe.bin
//go:embed templates/effect_cc_glue_gun.bin templates/effect_cc_grid_wipe.bin templates/effect_cc_hair.bin templates/effect_cc_image_wipe.bin templates/effect_cc_jaws.bin templates/effect_cc_light_wipe.bin templates/effect_cc_mr_mercury.bin templates/effect_cc_particle_systems_ii.bin templates/effect_cc_radial_scalewipe.bin templates/effect_cc_rain.bin templates/effect_cc_scale_wipe.bin templates/effect_cc_snow.bin
//go:embed templates/effect_cc_twister.bin templates/effect_cs_blockload.bin templates/effect_cs_color_neutralizer.bin templates/effect_cs_kernel.bin templates/effect_cs_linesweep.bin templates/effect_csrainfall.bin templates/effect_cssnowfall.bin
//go:embed templates/effect_adbe_aud_reverse.bin templates/effect_adbe_aud_bt.bin templates/effect_adbe_aud_delay.bin templates/effect_adbe_aud_flange.bin templates/effect_adbe_aud_hilo.bin templates/effect_adbe_aud_modulator.bin templates/effect_adbe_param_eq.bin templates/effect_adbe_aud_reverb.bin templates/effect_adbe_aud_stereo_mixer.bin templates/effect_adbe_aud_tone.bin
var effectTemplateFS embed.FS

// Effect match-name constants for the addable built-in set. These are AE's
// stable internal match-names; use them with AddEffect instead of hardcoding
// strings. The display name (what the AE Effects panel shows) is in the comment.
const (
	EffectGaussianBlur       = "ADBE Gaussian Blur 2"         // Gaussian Blur
	EffectFill               = "ADBE Fill"                    // Fill
	EffectTint               = "ADBE Tint"                    // Tint
	EffectBrightnessContrast = "ADBE Brightness & Contrast 2" // Brightness & Contrast
	EffectTritone            = "ADBE Tritone"                 // Tritone
	EffectLevels             = "ADBE Easy Levels2"            // Levels
	EffectLevelsIndividual   = "ADBE Pro Levels2"             // Levels (Individual Controls)
	EffectHueSaturation      = "ADBE HUE SATURATION"          // Hue/Saturation
	EffectBoxBlur            = "ADBE Box Blur"                // Fast Box Blur
	EffectGlow               = "ADBE Glo2"                    // Glow
	EffectInvert             = "ADBE Invert"                  // Invert
	EffectExposure           = "ADBE Exposure2"               // Exposure
	EffectDropShadow         = "ADBE Drop Shadow"             // Drop Shadow
	EffectSharpen            = "ADBE Sharpen"                 // Sharpen
	EffectMosaic             = "ADBE Mosaic"                  // Mosaic
	EffectNoise              = "ADBE Noise"                   // Noise
	EffectTransform          = "ADBE Geometry2"               // Transform
	EffectGradientRamp       = "ADBE Ramp"                    // Gradient Ramp
	EffectFractalNoise       = "ADBE Fractal Noise"           // Fractal Noise
	EffectMotionTile         = "ADBE Tile"                    // Motion Tile
	EffectDirectionalBlur    = "ADBE Motion Blur"             // Directional Blur
	EffectLinearWipe         = "ADBE Linear Wipe"             // Linear Wipe
	EffectWaveWarp           = "ADBE Wave Warp"               // Wave Warp
	EffectCurves             = "ADBE CurvesCustom"            // Curves
	EffectSliderControl      = "ADBE Slider Control"          // Slider Control
	EffectPointControl       = "ADBE Point Control"           // Point Control
	EffectColorControl       = "ADBE Color Control"           // Color Control
	EffectAngleControl       = "ADBE Angle Control"           // Angle Control
	EffectCheckboxControl    = "ADBE Checkbox Control"        // Checkbox Control
	EffectPoint3DControl     = "ADBE Point3D Control"         // 3D Point Control
	EffectSetMatte           = "ADBE Set Matte3"              // Set Matte (layer-reference effect)
	// Wave 4 (2026-06-15, fixture re_effect_lib3.aep) — MG distort/generate/
	// stylize/transition.
	EffectTurbulentDisplace = "ADBE Turbulent Displace" // Turbulent Displace
	EffectRoughenEdges      = "ADBE Roughen Edges"      // Roughen Edges
	EffectEcho              = "ADBE Echo"               // Echo
	EffectRadialBlur        = "ADBE Radial Blur"        // Radial Blur
	EffectFourColorGradient = "ADBE 4ColorGradient"     // 4-Color Gradient
	EffectCheckerboard      = "ADBE Checkerboard"       // Checkerboard
	EffectGrid              = "ADBE Grid"               // Grid
	EffectStroke            = "ADBE Stroke"             // Stroke
	EffectCornerPin         = "ADBE Corner Pin"         // Corner Pin
	EffectVenetianBlinds    = "ADBE Venetian Blinds"    // Venetian Blinds
	// Wave 5 (2026-06-16, fixture re_effect_lib5.aep) — MG distort / stylize /
	// perspective / color-correction / blur / channel / generate / time / matte.
	// All parameter-only (no layer pickwhip); extracted host id 15 uniform.
	EffectTwirl              = "ADBE Twirl"                // Twirl
	EffectPolarCoordinates   = "ADBE Polar Coordinates"    // Polar Coordinates
	EffectSpherize           = "ADBE Spherize"             // Spherize
	EffectMagnify            = "ADBE Magnify"              // Magnify
	EffectRipple             = "ADBE Ripple"               // Ripple
	EffectOpticsCompensation = "ADBE Optics Compensation"  // Optics Compensation
	EffectPosterize          = "ADBE Posterize"            // Posterize
	EffectThreshold          = "ADBE Threshold2"           // Threshold
	EffectFindEdges          = "ADBE Find Edges"           // Find Edges
	EffectColorEmboss        = "ADBE Color Emboss"         // Color Emboss
	EffectEmboss             = "ADBE Emboss"               // Emboss
	EffectStrobeLight        = "ADBE Strobe"               // Strobe Light
	EffectBrushStrokes       = "ADBE Brush Strokes"        // Brush Strokes
	EffectBevelAlpha         = "ADBE Bevel Alpha"          // Bevel Alpha
	EffectBevelEdges         = "ADBE Bevel Edges"          // Bevel Edges
	EffectPhotoFilter        = "ADBE Photo Filter"         // Photo Filter
	EffectVibrance           = "ADBE Vibrance"             // Vibrance
	EffectColorBalance       = "ADBE Color Balance 2"      // Color Balance
	EffectColorBalanceHLS    = "ADBE Color Balance (HLS)"  // Color Balance (HLS)
	EffectBlackAndWhite      = "ADBE Black&White"          // Black & White
	EffectGammaPedestalGain  = "ADBE Gamma/Pedestal/Gain2" // Gamma/Pedestal/Gain
	EffectChannelBlur        = "ADBE Channel Blur"         // Channel Blur
	EffectBilateralBlur      = "ADBE Bilateral"            // Bilateral Blur
	EffectSmartBlur          = "ADBE Smart Blur"           // Smart Blur
	EffectUnsharpMask        = "ADBE Unsharp Mask2"        // Unsharp Mask
	EffectShiftChannels      = "ADBE Shift Channels"       // Shift Channels
	EffectSolidComposite     = "ADBE Solid Composite"      // Solid Composite
	EffectMinimax            = "ADBE Minimax"              // Minimax
	EffectArithmetic         = "ADBE Arithmetic"           // Arithmetic
	EffectCircle             = "ADBE Circle"               // Circle
	EffectLensFlare          = "ADBE Lens Flare"           // Lens Flare
	EffectCellPattern        = "ADBE Cell Pattern"         // Cell Pattern
	EffectAdvancedLightning  = "ADBE Lightning 2"          // Advanced Lightning
	EffectBeam               = "ADBE Laser"                // Beam
	EffectPaintBucket        = "ADBE Paint Bucket"         // Paint Bucket
	EffectPosterizeTime      = "ADBE Posterize Time"       // Posterize Time
	EffectSimpleChoker       = "ADBE Simple Choker"        // Simple Choker
	EffectMatteChoker        = "ADBE Matte Choker"         // Matte Choker
	// Wave 6 (2026-06-16, fixture re_effect_lib6.aep) — more MG parameter-only
	// built-ins (+ wave-5 Bulge correction). All tdpi-host uniform (audio
	// spectrum/waveform default Audio Layer = None → no foreign tdpi).
	EffectBulge            = "ADBE Bulge"                // Bulge
	EffectOffset           = "ADBE Offset"               // Offset
	EffectMirror           = "ADBE Mirror"               // Mirror
	EffectFractal          = "ADBE Fractal"              // Fractal
	EffectWriteOn          = "ADBE Write-on"             // Write-on
	EffectScribble         = "ADBE Scribble Fill"        // Scribble
	EffectEyedropperFill   = "ADBE Eyedropper Fill"      // Eyedropper Fill
	EffectAudioSpectrum    = "ADBE AudSpect"             // Audio Spectrum
	EffectAudioWaveform    = "ADBE AudWave"              // Audio Waveform
	EffectAutoLevels       = "ADBE AutoLevels"           // Auto Levels
	EffectAutoColor        = "ADBE AutoColor"            // Auto Color
	EffectAutoContrast     = "ADBE AutoContrast"         // Auto Contrast
	EffectEqualize         = "ADBE Equalize"             // Equalize
	EffectLeaveColor       = "ADBE Leave Color"          // Leave Color
	EffectChangeToColor    = "ADBE Change To Color"      // Change to Color
	EffectChangeColor      = "ADBE Change Color"         // Change Color
	EffectRadialShadow     = "ADBE Radial Shadow"        // Radial Shadow
	EffectRemoveColorMatte = "ADBE Remove Color Matting" // Remove Color Matting
	EffectDustAndScratches = "ADBE Dust & Scratches"     // Dust & Scratches
	EffectNoiseAlpha       = "ADBE Noise Alpha2"         // Noise Alpha
	EffectNoiseHLS         = "ADBE Noise HLS2"           // Noise HLS
	EffectRadialWipe       = "ADBE Radial Wipe"          // Radial Wipe
	EffectBlockDissolve    = "ADBE Block Dissolve"       // Block Dissolve
	// Wave 7 (2026-06-16, fixture re_effect_lib7.aep) — Lumetri Color + Lightning
	// + the Cycore (CC) effect family. Probe-swept via canAddProperty. NOTE 4 CC
	// effects store a "CS …" internal match-name (≠ the "CC …" addProperty alias):
	// the const value is the STORED name (what AddEffect must be given + what AE
	// reads back). CC Vector Blur excluded (Vector Map layer pickwhip → foreign tdpi).
	EffectLumetri             = "ADBE Lumetri"             // Lumetri Color
	EffectLightning           = "ADBE Lightning"           // Lightning
	EffectCCRadialFastBlur    = "CC Radial Fast Blur"      // CC Radial Fast Blur
	EffectCCRadialBlur        = "CC Radial Blur"           // CC Radial Blur
	EffectCCCrossBlur         = "CS CrossBlur"             // CC Cross Blur
	EffectCCBendIt            = "CC Bend It"               // CC Bend It
	EffectCCBender            = "CC Bender"                // CC Bender
	EffectCCBlobbylize        = "CC Blobbylize"            // CC Blobbylize
	EffectCCFloMotion         = "CC Flo Motion"            // CC Flo Motion
	EffectCCGriddler          = "CC Griddler"              // CC Griddler
	EffectCCLens              = "CC Lens"                  // CC Lens
	EffectCCPageTurn          = "CC Page Turn"             // CC Page Turn
	EffectCCPowerPin          = "CC Power Pin"             // CC Power Pin
	EffectCCRipplePulse       = "CC Ripple Pulse"          // CC Ripple Pulse
	EffectCCSlant             = "CC Slant"                 // CC Slant
	EffectCCSmear             = "CC Smear"                 // CC Smear
	EffectCCSplit             = "CC Split"                 // CC Split
	EffectCCSplit2            = "CC Split 2"               // CC Split 2
	EffectCCTiler             = "CC Tiler"                 // CC Tiler
	EffectCCWarpoMatic        = "CC WarpoMatic"            // CC WarpoMatic
	EffectCCLightBurst        = "CC Light Burst 2.5"       // CC Light Burst 2.5
	EffectCCLightRays         = "CC Light Rays"            // CC Light Rays
	EffectCCLightSweep        = "CC Light Sweep"           // CC Light Sweep
	EffectCCThreads           = "CS Threads"               // CC Threads
	EffectCCCylinder          = "CC Cylinder"              // CC Cylinder
	EffectCCSphere            = "CC Sphere"                // CC Sphere
	EffectCCSpotlight         = "CC Spotlight"             // CC Spotlight
	EffectCCGlass             = "CC Glass"                 // CC Glass
	EffectCCHexTile           = "CS HexTile"               // CC HexTile
	EffectCCKaleida           = "CC Kaleida"               // CC Kaleida
	EffectCCMrSmoothie        = "CC Mr. Smoothie"          // CC Mr. Smoothie
	EffectCCPlastic           = "CC Plastic"               // CC Plastic
	EffectCCRepeTile          = "CC RepeTile"              // CC RepeTile
	EffectCCThreshold         = "CC Threshold"             // CC Threshold
	EffectCCThresholdRGB      = "CC Threshold RGB"         // CC Threshold RGB
	EffectCCPixelPolly        = "CC Pixel Polly"           // CC Pixel Polly
	EffectCCScatterize        = "CC Scatterize"            // CC Scatterize
	EffectCCStarBurst         = "CC Star Burst"            // CC Star Burst
	EffectCCForceMotionBlur   = "CC Force Motion Blur"     // CC Force Motion Blur
	EffectCCWideTime          = "CC Wide Time"             // CC Wide Time
	EffectCCColorOffset       = "CC Color Offset"          // CC Color Offset
	EffectCCToner             = "CC Toner"                 // CC Toner
	EffectCCBurnFilm          = "CC Burn Film"             // CC Burn Film
	EffectCCVignette          = "CS Vignette"              // CC Vignette
	EffectCCSimpleWireRemoval = "CC Simple Wire Removal"   // CC Simple Wire Removal
	// Wave 8 (2026-06-16, fixture re_effect_layerref.aep) — LAYER-REFERENCE
	// effects: their templates carry a materialized layer-ref param (with tdpi);
	// AddEffect retargets all tdpi to host (self), then SetEffectLayerParam aims
	// the layer-ref param at the real source (same flow as Set Matte). The
	// layer-ref param match-name for each is the …LayerParam const below.
	EffectDisplacementMap = "ADBE Displacement Map" // Displacement Map
	EffectCompoundBlur    = "ADBE Compound Blur"    // Compound Blur
	EffectCCVectorBlur    = "CC Vector Blur"        // CC Vector Blur
	// Wave 9 (2026-06-16, fixture re_effect_lib9.aep) — big probe sweep:
	// keying / simulation / utility / more Cycore CC. parameter-only (4 layer-ref
	// effects — 3D Glasses/Warp Stabilizer/Timewarp/CC Particle World — deferred:
	// foreign tdpi). NOTE several CC store a "CS …" internal name (≠ "CC …" alias).
	EffectBasic3D               = "ADBE Basic 3D"
	EffectBroadcastColors       = "ADBE Broadcast Colors"
	EffectChannelCombiner       = "ADBE Channel Combiner"
	EffectCineonConverter       = "ADBE Cineon Converter2"
	EffectColorKey              = "ADBE Color Key"
	EffectColorRange            = "ADBE Color Range"
	EffectExtract               = "ADBE Extract"
	EffectGeometryLegacy        = "ADBE Geometry"
	EffectGradientWipe          = "ADBE Gradient Wipe"
	EffectGrowBounds            = "ADBE GROW BOUNDS"
	EffectKeyCleaner            = "ADBE KeyCleaner"
	EffectLayerControl          = "ADBE Layer Control"
	EffectLumaKey               = "ADBE Luma Key"
	EffectMedian                = "ADBE Median"
	EffectNoiseHLSAuto          = "ADBE Noise HLS Auto2"
	EffectColorProfileConverter = "ADBE ProfileToProfile"
	EffectTimeDisplacement      = "ADBE Time Displacement"
	EffectTimecode              = "ADBE Timecode"
	EffectCCBallAction          = "CC Ball Action"
	EffectCCBubbles             = "CC Bubbles"
	EffectCCComposite           = "CC Composite"
	EffectCCDrizzle             = "CC Drizzle"
	EffectCCEnvironment         = "CC Environment"
	EffectCCGlassWipe           = "CC Glass Wipe"
	EffectCCGlueGun             = "CC Glue Gun"
	EffectCCGridWipe            = "CC Grid Wipe"
	EffectCCHair                = "CC Hair"
	EffectCCImageWipe           = "CC Image Wipe"
	EffectCCJaws                = "CC Jaws"
	EffectCCLightWipe           = "CC Light Wipe"
	EffectCCMrMercury           = "CC Mr. Mercury"
	EffectCCParticleSystemsII   = "CC Particle Systems II"
	EffectCCRadialScaleWipe     = "CC Radial ScaleWipe"
	EffectCCRain                = "CC Rain"
	EffectCCScaleWipe           = "CC Scale Wipe"
	EffectCCSnow                = "CC Snow"
	EffectCCTwister             = "CC Twister"
	EffectCCBlockLoad           = "CS BlockLoad"
	EffectCCColorNeutralizer    = "CS Color Neutralizer"
	EffectCCKernel              = "CS Kernel"
	EffectCCLineSweep           = "CS LineSweep"
	EffectCCRainfall            = "CSRainfall"
	EffectCCSnowfall            = "CSSnowfall"
	// Wave 10 (2026-06-17, fixture re_effect_audio.aep) — AUDIO-processing effects.
	// Unlike every prior wave these can ONLY be applied to a layer that HAS audio
	// (AE's canAddProperty returns false on a solid), so the fixture imports an mp3
	// and hosts them on that audio layer (tdpi-host=14 uniform). Non-visual: the
	// ship-gate is AE-accept + DOM readback (no render pixels). Match-names probed
	// via canAddProperty (note the irregular "ADBE Aud_Flange" underscore and the
	// "ADBE Param EQ" / "ADBE Aud Reverse" off-pattern names).
	EffectAudioBackwards    = "ADBE Aud Reverse"       // Backwards
	EffectAudioBassTreble   = "ADBE Aud BT"            // Bass & Treble
	EffectAudioDelay        = "ADBE Aud Delay"         // Delay
	EffectAudioFlangeChorus = "ADBE Aud_Flange"        // Flange & Chorus
	EffectAudioHighLowPass  = "ADBE Aud HiLo"          // High-Low Pass
	EffectAudioModulator    = "ADBE Aud Modulator"     // Modulator
	EffectAudioParametricEQ = "ADBE Param EQ"          // Parametric EQ
	EffectAudioReverb       = "ADBE Aud Reverb"        // Reverb
	EffectAudioStereoMixer  = "ADBE Aud Stereo Mixer"  // Stereo Mixer
	EffectAudioTone         = "ADBE Aud Tone"          // Tone
)

// Layer-reference parameter match-names for the wave-8 layer-ref effects — pass
// these to SetEffectLayerParam(layer, fx, paramMatchName, target).
const (
	EffectDisplacementMapLayer = "ADBE Displacement Map-0001" // Displacement Map Layer
	EffectCompoundBlurLayer    = "ADBE Compound Blur-0001"    // Blur Layer
	EffectCCVectorBlurMap      = "CC Vector Blur-0005"        // Vector Map
)

// effectTemplateFiles maps an effect match-name to its embedded template path.
// Each entry is an AE-native effect instance with AE's default parameter values;
// callers tune parameters afterward via the returned Effect.Parameters
// (Property.SetStaticValue works on effect params).
var effectTemplateFiles = map[string]string{
	EffectGaussianBlur:       "templates/effect_adbe_gaussian_blur_2.bin",
	EffectFill:               "templates/effect_adbe_fill.bin",
	EffectTint:               "templates/effect_adbe_tint.bin",
	EffectBrightnessContrast: "templates/effect_adbe_brightness_contrast_2.bin",
	EffectTritone:            "templates/effect_adbe_tritone.bin",
	EffectLevels:             "templates/effect_adbe_easy_levels2.bin",
	EffectLevelsIndividual:   "templates/effect_adbe_pro_levels2.bin",
	EffectHueSaturation:      "templates/effect_adbe_hue_saturation.bin",
	EffectBoxBlur:            "templates/effect_adbe_box_blur.bin",
	EffectGlow:               "templates/effect_adbe_glo2.bin",
	EffectInvert:             "templates/effect_adbe_invert.bin",
	EffectExposure:           "templates/effect_adbe_exposure2.bin",
	EffectDropShadow:         "templates/effect_adbe_drop_shadow.bin",
	EffectSharpen:            "templates/effect_adbe_sharpen.bin",
	EffectMosaic:             "templates/effect_adbe_mosaic.bin",
	EffectNoise:              "templates/effect_adbe_noise.bin",
	EffectTransform:          "templates/effect_adbe_geometry2.bin",
	EffectGradientRamp:       "templates/effect_adbe_ramp.bin",
	EffectFractalNoise:       "templates/effect_adbe_fractal_noise.bin",
	EffectMotionTile:         "templates/effect_adbe_tile.bin",
	EffectDirectionalBlur:    "templates/effect_adbe_motion_blur.bin",
	EffectLinearWipe:         "templates/effect_adbe_linear_wipe.bin",
	EffectWaveWarp:           "templates/effect_adbe_wave_warp.bin",
	EffectCurves:             "templates/effect_adbe_curvescustom.bin",
	EffectSliderControl:      "templates/effect_adbe_slider_control.bin",
	EffectPointControl:       "templates/effect_adbe_point_control.bin",
	EffectColorControl:       "templates/effect_adbe_color_control.bin",
	EffectAngleControl:       "templates/effect_adbe_angle_control.bin",
	EffectCheckboxControl:    "templates/effect_adbe_checkbox_control.bin",
	EffectPoint3DControl:     "templates/effect_adbe_point3d_control.bin",
	EffectSetMatte:           "templates/effect_adbe_set_matte3.bin",
	EffectTurbulentDisplace:  "templates/effect_adbe_turbulent_displace.bin",
	EffectRoughenEdges:       "templates/effect_adbe_roughen_edges.bin",
	EffectEcho:               "templates/effect_adbe_echo.bin",
	EffectRadialBlur:         "templates/effect_adbe_radial_blur.bin",
	EffectFourColorGradient:  "templates/effect_adbe_4colorgradient.bin",
	EffectCheckerboard:       "templates/effect_adbe_checkerboard.bin",
	EffectGrid:               "templates/effect_adbe_grid.bin",
	EffectStroke:             "templates/effect_adbe_stroke.bin",
	EffectCornerPin:          "templates/effect_adbe_corner_pin.bin",
	EffectVenetianBlinds:     "templates/effect_adbe_venetian_blinds.bin",
	EffectTwirl:              "templates/effect_adbe_twirl.bin",
	EffectPolarCoordinates:   "templates/effect_adbe_polar_coordinates.bin",
	EffectSpherize:           "templates/effect_adbe_spherize.bin",
	EffectMagnify:            "templates/effect_adbe_magnify.bin",
	EffectRipple:             "templates/effect_adbe_ripple.bin",
	EffectOpticsCompensation: "templates/effect_adbe_optics_compensation.bin",
	EffectPosterize:          "templates/effect_adbe_posterize.bin",
	EffectThreshold:          "templates/effect_adbe_threshold2.bin",
	EffectFindEdges:          "templates/effect_adbe_find_edges.bin",
	EffectColorEmboss:        "templates/effect_adbe_color_emboss.bin",
	EffectEmboss:             "templates/effect_adbe_emboss.bin",
	EffectStrobeLight:        "templates/effect_adbe_strobe.bin",
	EffectBrushStrokes:       "templates/effect_adbe_brush_strokes.bin",
	EffectBevelAlpha:         "templates/effect_adbe_bevel_alpha.bin",
	EffectBevelEdges:         "templates/effect_adbe_bevel_edges.bin",
	EffectPhotoFilter:        "templates/effect_adbe_photo_filter.bin",
	EffectVibrance:           "templates/effect_adbe_vibrance.bin",
	EffectColorBalance:       "templates/effect_adbe_color_balance_2.bin",
	EffectColorBalanceHLS:    "templates/effect_adbe_color_balance_hls.bin",
	EffectBlackAndWhite:      "templates/effect_adbe_black_white.bin",
	EffectGammaPedestalGain:  "templates/effect_adbe_gamma_pedestal_gain2.bin",
	EffectChannelBlur:        "templates/effect_adbe_channel_blur.bin",
	EffectBilateralBlur:      "templates/effect_adbe_bilateral.bin",
	EffectSmartBlur:          "templates/effect_adbe_smart_blur.bin",
	EffectUnsharpMask:        "templates/effect_adbe_unsharp_mask2.bin",
	EffectShiftChannels:      "templates/effect_adbe_shift_channels.bin",
	EffectSolidComposite:     "templates/effect_adbe_solid_composite.bin",
	EffectMinimax:            "templates/effect_adbe_minimax.bin",
	EffectArithmetic:         "templates/effect_adbe_arithmetic.bin",
	EffectCircle:             "templates/effect_adbe_circle.bin",
	EffectLensFlare:          "templates/effect_adbe_lens_flare.bin",
	EffectCellPattern:        "templates/effect_adbe_cell_pattern.bin",
	EffectAdvancedLightning:  "templates/effect_adbe_lightning_2.bin",
	EffectBeam:               "templates/effect_adbe_laser.bin",
	EffectPaintBucket:        "templates/effect_adbe_paint_bucket.bin",
	EffectPosterizeTime:      "templates/effect_adbe_posterize_time.bin",
	EffectSimpleChoker:       "templates/effect_adbe_simple_choker.bin",
	EffectMatteChoker:        "templates/effect_adbe_matte_choker.bin",
	EffectBulge:              "templates/effect_adbe_bulge.bin",
	EffectOffset:             "templates/effect_adbe_offset.bin",
	EffectMirror:             "templates/effect_adbe_mirror.bin",
	EffectFractal:            "templates/effect_adbe_fractal.bin",
	EffectWriteOn:            "templates/effect_adbe_write_on.bin",
	EffectScribble:           "templates/effect_adbe_scribble_fill.bin",
	EffectEyedropperFill:     "templates/effect_adbe_eyedropper_fill.bin",
	EffectAudioSpectrum:      "templates/effect_adbe_audspect.bin",
	EffectAudioWaveform:      "templates/effect_adbe_audwave.bin",
	EffectAutoLevels:         "templates/effect_adbe_autolevels.bin",
	EffectAutoColor:          "templates/effect_adbe_autocolor.bin",
	EffectAutoContrast:       "templates/effect_adbe_autocontrast.bin",
	EffectEqualize:           "templates/effect_adbe_equalize.bin",
	EffectLeaveColor:         "templates/effect_adbe_leave_color.bin",
	EffectChangeToColor:      "templates/effect_adbe_change_to_color.bin",
	EffectChangeColor:        "templates/effect_adbe_change_color.bin",
	EffectRadialShadow:       "templates/effect_adbe_radial_shadow.bin",
	EffectRemoveColorMatte:   "templates/effect_adbe_remove_color_matting.bin",
	EffectDustAndScratches:   "templates/effect_adbe_dust_scratches.bin",
	EffectNoiseAlpha:         "templates/effect_adbe_noise_alpha2.bin",
	EffectNoiseHLS:           "templates/effect_adbe_noise_hls2.bin",
	EffectRadialWipe:         "templates/effect_adbe_radial_wipe.bin",
	EffectBlockDissolve:      "templates/effect_adbe_block_dissolve.bin",
	EffectLumetri:             "templates/effect_adbe_lumetri.bin",
	EffectLightning:           "templates/effect_adbe_lightning.bin",
	EffectCCRadialFastBlur:    "templates/effect_cc_radial_fast_blur.bin",
	EffectCCRadialBlur:        "templates/effect_cc_radial_blur.bin",
	EffectCCCrossBlur:         "templates/effect_cs_crossblur.bin",
	EffectCCBendIt:            "templates/effect_cc_bend_it.bin",
	EffectCCBender:            "templates/effect_cc_bender.bin",
	EffectCCBlobbylize:        "templates/effect_cc_blobbylize.bin",
	EffectCCFloMotion:         "templates/effect_cc_flo_motion.bin",
	EffectCCGriddler:          "templates/effect_cc_griddler.bin",
	EffectCCLens:              "templates/effect_cc_lens.bin",
	EffectCCPageTurn:          "templates/effect_cc_page_turn.bin",
	EffectCCPowerPin:          "templates/effect_cc_power_pin.bin",
	EffectCCRipplePulse:       "templates/effect_cc_ripple_pulse.bin",
	EffectCCSlant:             "templates/effect_cc_slant.bin",
	EffectCCSmear:             "templates/effect_cc_smear.bin",
	EffectCCSplit:             "templates/effect_cc_split.bin",
	EffectCCSplit2:            "templates/effect_cc_split_2.bin",
	EffectCCTiler:             "templates/effect_cc_tiler.bin",
	EffectCCWarpoMatic:        "templates/effect_cc_warpomatic.bin",
	EffectCCLightBurst:        "templates/effect_cc_light_burst_2_5.bin",
	EffectCCLightRays:         "templates/effect_cc_light_rays.bin",
	EffectCCLightSweep:        "templates/effect_cc_light_sweep.bin",
	EffectCCThreads:           "templates/effect_cs_threads.bin",
	EffectCCCylinder:          "templates/effect_cc_cylinder.bin",
	EffectCCSphere:            "templates/effect_cc_sphere.bin",
	EffectCCSpotlight:         "templates/effect_cc_spotlight.bin",
	EffectCCGlass:             "templates/effect_cc_glass.bin",
	EffectCCHexTile:           "templates/effect_cs_hextile.bin",
	EffectCCKaleida:           "templates/effect_cc_kaleida.bin",
	EffectCCMrSmoothie:        "templates/effect_cc_mr_smoothie.bin",
	EffectCCPlastic:           "templates/effect_cc_plastic.bin",
	EffectCCRepeTile:          "templates/effect_cc_repetile.bin",
	EffectCCThreshold:         "templates/effect_cc_threshold.bin",
	EffectCCThresholdRGB:      "templates/effect_cc_threshold_rgb.bin",
	EffectCCPixelPolly:        "templates/effect_cc_pixel_polly.bin",
	EffectCCScatterize:        "templates/effect_cc_scatterize.bin",
	EffectCCStarBurst:         "templates/effect_cc_star_burst.bin",
	EffectCCForceMotionBlur:   "templates/effect_cc_force_motion_blur.bin",
	EffectCCWideTime:          "templates/effect_cc_wide_time.bin",
	EffectCCColorOffset:       "templates/effect_cc_color_offset.bin",
	EffectCCToner:             "templates/effect_cc_toner.bin",
	EffectCCBurnFilm:          "templates/effect_cc_burn_film.bin",
	EffectCCVignette:          "templates/effect_cs_vignette.bin",
	EffectCCSimpleWireRemoval: "templates/effect_cc_simple_wire_removal.bin",
	EffectDisplacementMap:     "templates/effect_adbe_displacement_map.bin",
	EffectCompoundBlur:        "templates/effect_adbe_compound_blur.bin",
	EffectCCVectorBlur:        "templates/effect_cc_vector_blur.bin",
	EffectBasic3D:               "templates/effect_adbe_basic_3d.bin",
	EffectBroadcastColors:       "templates/effect_adbe_broadcast_colors.bin",
	EffectChannelCombiner:       "templates/effect_adbe_channel_combiner.bin",
	EffectCineonConverter:       "templates/effect_adbe_cineon_converter2.bin",
	EffectColorKey:              "templates/effect_adbe_color_key.bin",
	EffectColorRange:            "templates/effect_adbe_color_range.bin",
	EffectExtract:               "templates/effect_adbe_extract.bin",
	EffectGeometryLegacy:        "templates/effect_adbe_geometry.bin",
	EffectGradientWipe:          "templates/effect_adbe_gradient_wipe.bin",
	EffectGrowBounds:            "templates/effect_adbe_grow_bounds.bin",
	EffectKeyCleaner:            "templates/effect_adbe_keycleaner.bin",
	EffectLayerControl:          "templates/effect_adbe_layer_control.bin",
	EffectLumaKey:               "templates/effect_adbe_luma_key.bin",
	EffectMedian:                "templates/effect_adbe_median.bin",
	EffectNoiseHLSAuto:          "templates/effect_adbe_noise_hls_auto2.bin",
	EffectColorProfileConverter: "templates/effect_adbe_profiletoprofile.bin",
	EffectTimeDisplacement:      "templates/effect_adbe_time_displacement.bin",
	EffectTimecode:              "templates/effect_adbe_timecode.bin",
	EffectCCBallAction:          "templates/effect_cc_ball_action.bin",
	EffectCCBubbles:             "templates/effect_cc_bubbles.bin",
	EffectCCComposite:           "templates/effect_cc_composite.bin",
	EffectCCDrizzle:             "templates/effect_cc_drizzle.bin",
	EffectCCEnvironment:         "templates/effect_cc_environment.bin",
	EffectCCGlassWipe:           "templates/effect_cc_glass_wipe.bin",
	EffectCCGlueGun:             "templates/effect_cc_glue_gun.bin",
	EffectCCGridWipe:            "templates/effect_cc_grid_wipe.bin",
	EffectCCHair:                "templates/effect_cc_hair.bin",
	EffectCCImageWipe:           "templates/effect_cc_image_wipe.bin",
	EffectCCJaws:                "templates/effect_cc_jaws.bin",
	EffectCCLightWipe:           "templates/effect_cc_light_wipe.bin",
	EffectCCMrMercury:           "templates/effect_cc_mr_mercury.bin",
	EffectCCParticleSystemsII:   "templates/effect_cc_particle_systems_ii.bin",
	EffectCCRadialScaleWipe:     "templates/effect_cc_radial_scalewipe.bin",
	EffectCCRain:                "templates/effect_cc_rain.bin",
	EffectCCScaleWipe:           "templates/effect_cc_scale_wipe.bin",
	EffectCCSnow:                "templates/effect_cc_snow.bin",
	EffectCCTwister:             "templates/effect_cc_twister.bin",
	EffectCCBlockLoad:           "templates/effect_cs_blockload.bin",
	EffectCCColorNeutralizer:    "templates/effect_cs_color_neutralizer.bin",
	EffectCCKernel:              "templates/effect_cs_kernel.bin",
	EffectCCLineSweep:           "templates/effect_cs_linesweep.bin",
	EffectCCRainfall:            "templates/effect_csrainfall.bin",
	EffectCCSnowfall:            "templates/effect_cssnowfall.bin",
	EffectAudioBackwards:    "templates/effect_adbe_aud_reverse.bin",
	EffectAudioBassTreble:   "templates/effect_adbe_aud_bt.bin",
	EffectAudioDelay:        "templates/effect_adbe_aud_delay.bin",
	EffectAudioFlangeChorus: "templates/effect_adbe_aud_flange.bin",
	EffectAudioHighLowPass:  "templates/effect_adbe_aud_hilo.bin",
	EffectAudioModulator:    "templates/effect_adbe_aud_modulator.bin",
	EffectAudioParametricEQ: "templates/effect_adbe_param_eq.bin",
	EffectAudioReverb:       "templates/effect_adbe_aud_reverb.bin",
	EffectAudioStereoMixer:  "templates/effect_adbe_aud_stereo_mixer.bin",
	EffectAudioTone:         "templates/effect_adbe_aud_tone.bin",
}

type cachedEffectTemplate struct {
	once  sync.Once
	chunk *rifx.Chunk
	err   error
}

var effectTemplateCache = map[string]*cachedEffectTemplate{}
var effectTemplateCacheMu sync.Mutex

// SupportedEffects returns the sorted set of effect match-names AddEffect can
// currently add from an embedded template.
func SupportedEffects() []string {
	names := make([]string, 0, len(effectTemplateFiles))
	for k := range effectTemplateFiles {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// cloneEffectTemplate parses the embedded template for matchName once (cached)
// and returns a deep clone of its (tdmn, sspc) pair so the spliced chunks never
// alias the cache (concurrent-mutate safe).
func cloneEffectTemplate(matchName string) (tdmn, sspc *rifx.Chunk, err error) {
	path, ok := effectTemplateFiles[matchName]
	if !ok {
		return nil, nil, fmt.Errorf("AddEffect: unsupported effect %q (supported: %v)", matchName, SupportedEffects())
	}
	effectTemplateCacheMu.Lock()
	ct := effectTemplateCache[matchName]
	if ct == nil {
		ct = &cachedEffectTemplate{}
		effectTemplateCache[matchName] = ct
	}
	effectTemplateCacheMu.Unlock()

	ct.once.Do(func() {
		raw, e := effectTemplateFS.ReadFile(path)
		if e != nil {
			ct.err = fmt.Errorf("read effect template %q (%s): %w", matchName, path, e)
			return
		}
		wrapper, e := rifx.ReadChunk(bytes.NewReader(raw))
		if e != nil {
			ct.err = fmt.Errorf("parse effect template %q: %w", matchName, e)
			return
		}
		if len(wrapper.Children) != 2 {
			ct.err = fmt.Errorf("effect template %q: want 2 wrapper children (tdmn, sspc), got %d", matchName, len(wrapper.Children))
			return
		}
		ct.chunk = wrapper
	})
	if ct.err != nil {
		return nil, nil, ct.err
	}
	return deepCloneChunk(ct.chunk.Children[0]), deepCloneChunk(ct.chunk.Children[1]), nil
}

// retargetEffectHostLayer rewrites every tdpi chunk (the 4-byte host-layer
// binding each effect param's tdbs carries) in the cloned template payload to
// the destination layer's ID. The embedded templates hold the extraction
// fixture's host layer id verbatim; AE validates the binding resolves on open
// and rejects the project with "cannot find layer ID=N in composition" when it
// dangles (RE'd 2026-06-10: re_effect_library host id 15, re_shape_effect host
// id 13 — tdpi tracks the host in both).
func retargetEffectHostLayer(c *rifx.Chunk, layerID uint32) {
	if c.ID == rifx.IDTdpi && len(c.Data) >= 4 {
		binary.BigEndian.PutUint32(c.Data[0:4], layerID)
	}
	for _, ch := range c.Children {
		retargetEffectHostLayer(ch, layerID)
	}
}

// RemoveEffect removes the effect at the given 0-based index from the layer's
// Effect Parade. Thin index-validated wrapper over RemovePropertyGroup (which is
// AE 2020 + AE 2025 ship-gate green for Effect-Parade child removal), giving
// AddEffect a symmetric inverse.
// (Full contract lives on the aep.RemoveEffect facade — docgen source.)
func RemoveEffect(layer *Layer, index int) error {
	if layer == nil {
		return fmt.Errorf("RemoveEffect: layer is nil")
	}
	parade := layer.EffectsParade()
	if parade == nil {
		return fmt.Errorf("RemoveEffect: layer %q has no Effect Parade group", layer.Name)
	}
	n := parade.NumProperties()
	if index < 0 || index >= n {
		return fmt.Errorf("RemoveEffect: index %d out of range (have %d effects)", index, n)
	}
	g, ok := parade.ChildByIndex(index).(*AEPropertyGroup)
	if !ok {
		return fmt.Errorf("RemoveEffect: effect at index %d is not a property group", index)
	}
	return RemovePropertyGroup(g)
}

// aeDefaultGroupName is the placeholder AE persists in a group's tdsn when the
// user never renamed it — "-_0_/-" observed on every AE-native Effect Parade
// (re_shape_effect.aep, re_effect_library.aep).
const aeDefaultGroupName = "-_0_/-"

// ensureEffectParade returns the layer's Effect Parade, splicing a fresh empty
// parade group (tdsb 0x01 + tdsn "-_0_/-" + Group End sentinel) into the Layr
// property tree when absent — immediately before "ADBE Transform Group",
// matching AE's emitted group order. The returned undo restores the pre-create
// chunk + scene-tree state (no-op when the parade already existed).
func ensureEffectParade(layer *Layer) (*AEPropertyGroup, func(), error) {
	if parade := layer.EffectsParade(); parade != nil {
		return parade, func() {}, nil
	}
	return spliceEmptyParade(layer, "AddEffect", MatchNameGroupEffectParade, []string{MatchNameGroupTransform})
}

// spliceEmptyParade splices a fresh empty parade group (tdsb 0x01 + tdsn
// "-_0_/-" + Group End sentinel) named paradeMatchName into the layer's Layr
// property tree, immediately before the first anchor group found (tried in
// anchorMatchNames order), mirroring the splice in the scene tree. The
// returned undo restores the pre-create chunk + scene-tree state.
func spliceEmptyParade(layer *Layer, opName, paradeMatchName string, anchorMatchNames []string) (*AEPropertyGroup, func(), error) {
	tree := layer.PropertyTree()
	if tree == nil {
		return nil, nil, fmt.Errorf("%s: layer %q was built outside the parser (no property tree to hold a %s); round-trip the project through aep.Reopen first, then mutate the re-parsed layer", opName, layer.Name, paradeMatchName)
	}
	lb := layerBack(layer)
	if lb == nil || lb.layrList == nil {
		return nil, nil, fmt.Errorf("%s: layer %q has no Layr chunk back-ref", opName, layer.Name)
	}
	var outer *rifx.Chunk
	for _, ch := range lb.layrList.Children {
		if ch.IsList() && ch.FormType == rifx.IDTdgp {
			outer = ch
			break
		}
	}
	if outer == nil {
		return nil, nil, fmt.Errorf("%s: layer %q has no property-group LIST in its Layr", opName, layer.Name)
	}
	anchor := -1
	for _, anchorName := range anchorMatchNames {
		for i, ch := range outer.Children {
			if ch.ID == rifx.IDTdmn && string(bytes.TrimRight(ch.Data, "\x00")) == anchorName {
				anchor = i
				break
			}
		}
		if anchor >= 0 {
			break
		}
	}
	if anchor < 0 {
		return nil, nil, fmt.Errorf("%s: layer %q has none of %v to anchor the %s position", opName, layer.Name, anchorMatchNames, paradeMatchName)
	}

	paradeTdgp := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp, Children: []*rifx.Chunk{
		makeTdsb(),
		makeTdsn(aeDefaultGroupName),
		makeTdmn("ADBE Group End"),
	}}

	oldOuterChildren := append([]*rifx.Chunk(nil), outer.Children...)
	oldTreeChildren := append([]PropertyBase(nil), tree.Children...)

	spliced := make([]*rifx.Chunk, 0, len(outer.Children)+2)
	spliced = append(spliced, outer.Children[:anchor]...)
	spliced = append(spliced, makeTdmn(paradeMatchName), paradeTdgp)
	spliced = append(spliced, outer.Children[anchor:]...)
	outer.Children = spliced

	parade := &AEPropertyGroup{MatchName: paradeMatchName, Name: paradeMatchName}
	scene.SetPropertyGroupParent(parade, tree)
	scene.SetPropertyGroupBack(parade, &propertyGroupBackrefs{chunk: paradeTdgp})
	treeAnchor := len(tree.Children)
	for _, anchorName := range anchorMatchNames {
		found := false
		for i, c := range tree.Children {
			if g, ok := c.(*AEPropertyGroup); ok && g.MatchName == anchorName {
				treeAnchor = i
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	newTreeChildren := make([]PropertyBase, 0, len(tree.Children)+1)
	newTreeChildren = append(newTreeChildren, tree.Children[:treeAnchor]...)
	newTreeChildren = append(newTreeChildren, parade)
	newTreeChildren = append(newTreeChildren, tree.Children[treeAnchor:]...)
	tree.Children = newTreeChildren

	return parade, func() {
		outer.Children = oldOuterChildren
		tree.Children = oldTreeChildren
	}, nil
}

// AddEffect appends an effect to the layer's Effect Parade (auto-creating the
// parade for effect-less parsed layers) and returns the parsed *Effect (so the
// caller can tune Effect.Parameters immediately).
// (Full contract + RE notes live on the aep.AddEffect facade — docgen source.)
func AddEffect(layer *Layer, effectMatchName string) (*Effect, error) {
	if layer == nil {
		return nil, fmt.Errorf("AddEffect: layer is nil")
	}
	if layer.Type == LayerTypeCamera || layer.Type == LayerTypeLight {
		return nil, fmt.Errorf("AddEffect: layer %q is a %s layer (AE does not allow effects on camera/light layers)", layer.Name, layer.Type)
	}
	parade, undoParadeCreate, err := ensureEffectParade(layer)
	if err != nil {
		return nil, err
	}
	pgb := propertyGroupBack(parade)
	if pgb == nil || pgb.chunk == nil {
		undoParadeCreate()
		return nil, fmt.Errorf("AddEffect: Effect Parade for layer %q has no chunk back-ref", layer.Name)
	}

	tdmnCh, sspcCh, err := cloneEffectTemplate(effectMatchName)
	if err != nil {
		undoParadeCreate()
		return nil, err
	}
	retargetEffectHostLayer(sspcCh, layer.ID)

	children := pgb.chunk.Children
	insertIdx := len(children)
	for i, ch := range children {
		if ch.ID == rifx.IDTdmn && string(bytes.TrimRight(ch.Data, "\x00")) == "ADBE Group End" {
			insertIdx = i
			break
		}
	}

	// Snapshot for atomic rollback.
	oldChunkChildren := append([]*rifx.Chunk(nil), children...)
	oldSceneChildren := append([]PropertyBase(nil), parade.Children...)
	oldEffects := append([]*Effect(nil), layer.Effects...)
	oldWarningsLen := warningsLen(layer)

	rollback := func() {
		pgb.chunk.Children = oldChunkChildren
		parade.Children = oldSceneChildren
		layer.Effects = oldEffects
		rollbackWarnings(layer, oldWarningsLen)
		undoParadeCreate()
	}

	// Chunk: splice (tdmn, sspc) in just before the Group End sentinel.
	spliced := make([]*rifx.Chunk, 0, len(children)+2)
	spliced = append(spliced, children[:insertIdx]...)
	spliced = append(spliced, tdmnCh, sspcCh)
	spliced = append(spliced, children[insertIdx:]...)
	pgb.chunk.Children = spliced

	// Scene: append a stand-in group node (the parade's scene children are the
	// effect groups in order; the Group End sentinel is chunk-only).
	cloneNode := &AEPropertyGroup{MatchName: effectMatchName, Name: effectMatchName}
	scene.SetPropertyGroupParent(cloneNode, parade)
	scene.SetPropertyGroupBack(cloneNode, &propertyGroupBackrefs{chunk: sspcCh})
	parade.Children = append(parade.Children, cloneNode)

	// Flat mirror: re-parse the spliced pair so the typed Effect's back-refs
	// point at the spliced chunks (never aliased to the template cache).
	var newEffect *Effect
	if comp := scene.LayerComp(layer); comp != nil && scene.CompositionProj(comp) != nil {
		ctx := newParseCtxFPS(comp.TickRate, comp.FrameRate, comp.Name, &scene.CompositionProj(comp).Warnings)
		tmpParade := &rifx.Chunk{ID: rifx.IDList, FormType: rifx.IDTdgp, Children: []*rifx.Chunk{tdmnCh, sspcCh}}
		var tmp []*Effect
		collectEffects(tmpParade, &tmp, ctx)
		if len(tmp) != 1 {
			rollback()
			return nil, fmt.Errorf("AddEffect: spliced effect re-parse produced %d effects (want 1)", len(tmp))
		}
		newEffect = tmp[0]
		layer.Effects = append(layer.Effects, newEffect)
	}

	if newWarn := newWarningsSince(layer, oldWarningsLen); len(newWarn) > 0 {
		rollback()
		return nil, fmt.Errorf("AddEffect: produced %d parser warning(s), rolled back: %v", len(newWarn), newWarn)
	}

	return newEffect, nil
}
