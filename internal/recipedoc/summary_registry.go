package recipedoc

var fieldSummaries = buildFieldSummaries()

// allowedMissingSemanticSummary is intentionally path-exact. Add a field here
// only when its semantics are known to be deferred to a later documentation pass.
var allowedMissingSemanticSummary = map[string]string{}

func buildFieldSummaries() map[string]string {
	out := map[string]string{
		"schema_version":         "Recipe schema version.",
		"project":                "Project-level metadata used when materializing the recipe.",
		"project.name":           "Project display name.",
		"project.target_version": "AE target template used when compiling the recipe.",
		"comps[]":                "Composition definitions to create or validate.",
		"expected_profile":       "Optional assertions used to compare the generated project against expected structure.",
	}

	addCompSummaries(out, "comps[]", "Composition")
	addLayerSummaries(out, "comps[].layers[]")
	addTextStyleSummaries(out, "comps[].layers[].text_style", "Layer text style")
	addCameraSummaries(out, "comps[].layers[].camera")
	addLightSummaries(out, "comps[].layers[].light")
	addShapeSummaries(out, "comps[].layers[].shape")
	addMaskSummaries(out, "comps[].layers[].masks[]", "Layer mask")
	addTransformSummaries(out, "comps[].layers[].transform")
	addEffectSummaries(out, "comps[].layers[].effects[]", "Layer effect")
	addExpectedProfileSummaries(out)
	return out
}

func addCompSummaries(out map[string]string, prefix, label string) {
	addSummaries(out, prefix, map[string]string{
		"name":                              label + " display name.",
		"width":                             label + " width in pixels.",
		"height":                            label + " height in pixels.",
		"frame_rate":                        label + " frame rate in frames per second.",
		"duration":                          label + " duration in seconds.",
		"background_color":                  label + " background color as RGB channels.",
		"label":                             label + " label color index.",
		"comment":                           label + " comment text.",
		"renderer":                          label + " renderer identifier.",
		"resolution_factor":                 label + " preview resolution factor.",
		"pixel_aspect":                      label + " pixel aspect ratio.",
		"display_start_time":                label + " display start time in seconds.",
		"frame_blending":                    label + " frame blending switch.",
		"draft_3d":                          label + " draft 3D switch.",
		"hide_shy_layers":                   label + " shy layer visibility switch.",
		"preserve_nested_frame_rate":        label + " nested frame-rate preservation switch.",
		"preserve_nested_resolution":        label + " nested resolution preservation switch.",
		"motion_blur":                       label + " motion blur settings.",
		"motion_blur.enabled":               label + " motion blur enable switch.",
		"motion_blur.shutter_angle":         label + " motion blur shutter angle.",
		"motion_blur.shutter_phase":         label + " motion blur shutter phase.",
		"motion_blur.adaptive_sample_limit": label + " motion blur adaptive sample limit.",
		"motion_blur.samples_per_frame":     label + " motion blur samples per frame.",
		"work_area":                         label + " work area range.",
		"work_area.start":                   label + " work area start time in seconds.",
		"work_area.end":                     label + " work area end time in seconds.",
		"layers[]":                          label + " layer definitions.",
	})
}

func addLayerSummaries(out map[string]string, prefix string) {
	addSummaries(out, prefix, map[string]string{
		"type":                     "Layer creation type.",
		"name":                     "Layer display name.",
		"label":                    "Layer label color index.",
		"comment":                  "Layer comment text.",
		"visible":                  "Layer video visibility switch.",
		"solo":                     "Layer solo switch.",
		"locked":                   "Layer lock switch.",
		"motion_blur":              "Layer motion blur switch.",
		"shy":                      "Layer shy switch.",
		"effects_enabled":          "Layer effects enable switch.",
		"audio_enabled":            "Layer audio enable switch.",
		"frame_blend_enabled":      "Layer frame blending switch.",
		"markers_locked":           "Layer marker lock switch.",
		"collapse_transform":       "Layer collapse transformations switch.",
		"is_3d":                    "Layer 3D switch.",
		"is_adjust":                "Adjustment-layer switch.",
		"is_null":                  "Null-layer switch.",
		"is_guide":                 "Guide-layer switch.",
		"sampling_bicubic":         "Bicubic sampling switch.",
		"frame_blend_pixel_motion": "Pixel-motion frame blending switch.",
		"preserve_transparency":    "Preserve transparency switch.",
		"quality":                  "Layer quality mode.",
		"blending_mode":            "Layer blending mode.",
		"track_matte":              "Layer track matte mode.",
		"matte":                    "Explicit track matte source layer name.",
		"auto_orient":              "Layer auto-orientation mode.",
		"start_time":               "Layer start time in seconds.",
		"in_point":                 "Layer in point in seconds.",
		"out_point":                "Layer out point in seconds.",
		"parent":                   "Parent layer name.",
		"text":                     "Source text for a text layer.",
		"text_style":               "Text layer style overrides.",
		"camera":                   "Camera layer options.",
		"light":                    "Light layer options.",
		"shape":                    "Shape layer primitive and operators.",
		"masks[]":                  "Layer masks.",
		"transform":                "Layer transform block.",
		"effects[]":                "Built-in effect instance to add to the layer.",
	})
}

func addTextStyleSummaries(out map[string]string, prefix, label string) {
	addSummaries(out, prefix, map[string]string{
		"run_index":       label + " text run index.",
		"paragraph_index": label + " paragraph index.",
		"font_size":       label + " font size.",
		"fill_color":      label + " fill color as RGB channels.",
		"tracking":        label + " character tracking amount.",
		"faux_bold":       label + " faux bold switch.",
		"faux_italic":     label + " faux italic switch.",
		"apply_stroke":    label + " stroke enable switch.",
		"stroke_color":    label + " stroke color as RGB channels.",
		"stroke_width":    label + " stroke width.",
		"justification":   label + " paragraph justification mode.",
	})
}

func addCameraSummaries(out map[string]string, prefix string) {
	addSummaries(out, prefix, map[string]string{
		"":                          "Camera layer options.",
		"zoom":                      "Camera zoom value.",
		"depth_of_field":            "Depth of field enable switch.",
		"focus_distance":            "Camera focus distance.",
		"aperture":                  "Camera aperture.",
		"blur_level":                "Camera blur level.",
		"iris_shape":                "Camera iris shape.",
		"iris_rotation":             "Camera iris rotation.",
		"iris_roundness":            "Camera iris roundness.",
		"iris_aspect_ratio":         "Camera iris aspect ratio.",
		"iris_diffraction_fringe":   "Camera iris diffraction fringe amount.",
		"iris_highlight_gain":       "Camera iris highlight gain.",
		"iris_highlight_threshold":  "Camera iris highlight threshold.",
		"iris_highlight_saturation": "Camera iris highlight saturation.",
	})
}

func addLightSummaries(out map[string]string, prefix string) {
	addSummaries(out, prefix, map[string]string{
		"":                 "Light layer options.",
		"kind":             "Light type.",
		"source_layer":     "Source layer for environment light data.",
		"intensity":        "Light intensity.",
		"color":            "Light color as RGB channels.",
		"casts_shadows":    "Light shadow-casting switch.",
		"shadow_darkness":  "Light shadow darkness.",
		"shadow_diffusion": "Light shadow diffusion.",
		"falloff_type":     "Light falloff type.",
		"falloff_start":    "Light falloff start distance.",
		"falloff_distance": "Light falloff distance.",
		"cone_angle":       "Spotlight cone angle.",
		"cone_feather":     "Spotlight cone feather amount.",
	})
}

func addShapeSummaries(out map[string]string, prefix string) {
	addSummaries(out, prefix, map[string]string{
		"":                     "Shape layer primitive and operators.",
		"kind":                 "Shape primitive type.",
		"size":                 "Shape size vector.",
		"position":             "Shape local position vector.",
		"roundness":            "Rectangle corner roundness.",
		"points":               "Star or polygon point count.",
		"rotation":             "Shape local rotation.",
		"inner_radius":         "Star inner radius.",
		"outer_radius":         "Star or polygon outer radius.",
		"inner_roundness":      "Star inner roundness.",
		"outer_roundness":      "Star or polygon outer roundness.",
		"fill_color":           "Solid fill color as RGB channels.",
		"fill_opacity":         "Solid fill opacity.",
		"fill_blend_mode":      "Solid fill blend mode.",
		"fill_composite_order": "Solid fill composite order.",
		"fill_rule":            "Solid fill rule.",
		"gradient_fill":        "Gradient fill settings.",
		"gradient_stroke":      "Gradient stroke settings.",
		"stroke":               "Solid stroke settings.",
		"trim":                 "Trim paths operator settings.",
		"round_corners":        "Round corners operator settings.",
		"offset_paths":         "Offset paths operator settings.",
		"repeater":             "Repeater operator settings.",
		"merge_paths":          "Merge paths operator settings.",
		"zigzag":               "Zig Zag operator settings.",
		"pucker_bloat":         "Pucker and Bloat operator settings.",
		"twist":                "Twist operator settings.",
		"wiggle_paths":         "Wiggle Paths operator settings.",
		"wiggle_transform":     "Wiggle Transform operator settings.",
	})
	addGradientFillSummaries(out, prefix+".gradient_fill")
	addGradientStrokeSummaries(out, prefix+".gradient_stroke")
	addStrokeSummaries(out, prefix+".stroke")
	addTrimSummaries(out, prefix+".trim")
	addSummaries(out, prefix+".round_corners", map[string]string{
		"":       "Round corners operator settings.",
		"radius": "Round corners radius.",
	})
	addSummaries(out, prefix+".offset_paths", map[string]string{
		"":            "Offset paths operator settings.",
		"amount":      "Offset paths amount.",
		"line_join":   "Offset paths line join mode.",
		"miter_limit": "Offset paths miter limit.",
		"copies":      "Offset paths copy count.",
		"copy_offset": "Offset paths copy offset.",
	})
	addRepeaterSummaries(out, prefix+".repeater")
	addSummaries(out, prefix+".merge_paths", map[string]string{
		"":     "Merge paths operator settings.",
		"type": "Merge paths mode.",
	})
	addSummaries(out, prefix+".zigzag", map[string]string{
		"":       "Zig Zag operator settings.",
		"size":   "Zig Zag size.",
		"detail": "Zig Zag detail.",
		"points": "Zig Zag point mode.",
	})
	addSummaries(out, prefix+".pucker_bloat", map[string]string{
		"":       "Pucker and Bloat operator settings.",
		"amount": "Pucker and Bloat amount.",
	})
	addSummaries(out, prefix+".twist", map[string]string{
		"":       "Twist operator settings.",
		"angle":  "Twist angle.",
		"center": "Twist center point.",
	})
	addWigglePathsSummaries(out, prefix+".wiggle_paths")
	addWiggleTransformSummaries(out, prefix+".wiggle_transform")
}

func addGradientFillSummaries(out map[string]string, prefix string) {
	addSummaries(out, prefix, map[string]string{
		"":                 "Gradient fill settings.",
		"type":             "Gradient fill type.",
		"start_point":      "Gradient fill start point.",
		"end_point":        "Gradient fill end point.",
		"highlight_length": "Radial gradient fill highlight length.",
		"highlight_angle":  "Radial gradient fill highlight angle.",
		"color_stops[]":    "Gradient fill color stops.",
		"alpha_stops[]":    "Gradient fill alpha stops.",
	})
	addGradientStopSummaries(out, prefix+".color_stops[]", "Gradient fill color stop")
	addGradientAlphaSummaries(out, prefix+".alpha_stops[]", "Gradient fill alpha stop")
}

func addGradientStrokeSummaries(out map[string]string, prefix string) {
	addSummaries(out, prefix, map[string]string{
		"":                 "Gradient stroke settings.",
		"type":             "Gradient stroke type.",
		"start_point":      "Gradient stroke start point.",
		"end_point":        "Gradient stroke end point.",
		"highlight_length": "Radial gradient stroke highlight length.",
		"highlight_angle":  "Radial gradient stroke highlight angle.",
		"width":            "Gradient stroke width.",
		"line_cap":         "Gradient stroke line cap mode.",
		"line_join":        "Gradient stroke line join mode.",
		"miter_limit":      "Gradient stroke miter limit.",
		"color_stops[]":    "Gradient stroke color stops.",
		"alpha_stops[]":    "Gradient stroke alpha stops.",
	})
	addGradientStopSummaries(out, prefix+".color_stops[]", "Gradient stroke color stop")
	addGradientAlphaSummaries(out, prefix+".alpha_stops[]", "Gradient stroke alpha stop")
}

func addGradientStopSummaries(out map[string]string, prefix, label string) {
	addSummaries(out, prefix, map[string]string{
		"offset":   label + " offset.",
		"midpoint": label + " midpoint.",
		"color":    label + " RGB color.",
	})
}

func addGradientAlphaSummaries(out map[string]string, prefix, label string) {
	addSummaries(out, prefix, map[string]string{
		"offset":   label + " offset.",
		"midpoint": label + " midpoint.",
		"alpha":    label + " opacity value.",
	})
}

func addStrokeSummaries(out map[string]string, prefix string) {
	addSummaries(out, prefix, map[string]string{
		"":                "Solid stroke settings.",
		"color":           "Stroke color as RGB channels.",
		"width":           "Stroke width.",
		"opacity":         "Stroke opacity.",
		"line_cap":        "Stroke line cap mode.",
		"line_join":       "Stroke line join mode.",
		"miter_limit":     "Stroke miter limit.",
		"composite_order": "Stroke composite order.",
		"taper":           "Stroke taper settings.",
		"wave":            "Stroke wave settings.",
		"dashes":          "Stroke dash pattern settings.",
	})
	addSummaries(out, prefix+".taper", map[string]string{
		"":             "Stroke taper settings.",
		"start_length": "Stroke taper start length.",
		"end_length":   "Stroke taper end length.",
		"start_width":  "Stroke taper start width.",
		"end_width":    "Stroke taper end width.",
		"start_ease":   "Stroke taper start ease.",
		"end_ease":     "Stroke taper end ease.",
	})
	addSummaries(out, prefix+".wave", map[string]string{
		"":           "Stroke wave settings.",
		"amount":     "Stroke wave amount.",
		"wavelength": "Stroke wave wavelength.",
		"phase":      "Stroke wave phase.",
	})
	addSummaries(out, prefix+".dashes", map[string]string{
		"":     "Stroke dash pattern settings.",
		"dash": "Stroke dash length.",
		"gap":  "Stroke gap length.",
	})
}

func addTrimSummaries(out map[string]string, prefix string) {
	addSummaries(out, prefix, map[string]string{
		"":       "Trim paths operator settings.",
		"start":  "Trim paths start percentage.",
		"end":    "Trim paths end percentage.",
		"offset": "Trim paths offset.",
	})
}

func addRepeaterSummaries(out map[string]string, prefix string) {
	addSummaries(out, prefix, map[string]string{
		"":              "Repeater operator settings.",
		"copies":        "Repeater copy count.",
		"offset":        "Repeater copy offset.",
		"order":         "Repeater composite order.",
		"anchor":        "Repeater transform anchor point.",
		"position":      "Repeater transform position.",
		"scale":         "Repeater transform scale.",
		"rotation":      "Repeater transform rotation.",
		"start_opacity": "Repeater starting opacity.",
		"end_opacity":   "Repeater ending opacity.",
	})
}

func addWigglePathsSummaries(out map[string]string, prefix string) {
	addSummaries(out, prefix, map[string]string{
		"":                   "Wiggle Paths operator settings.",
		"size":               "Wiggle Paths size.",
		"detail":             "Wiggle Paths detail.",
		"wiggles_per_second": "Wiggle Paths frequency.",
		"random_seed":        "Wiggle Paths random seed.",
		"points":             "Wiggle Paths point mode.",
		"correlation":        "Wiggle Paths correlation.",
		"temporal_phase":     "Wiggle Paths temporal phase.",
		"spatial_phase":      "Wiggle Paths spatial phase.",
	})
}

func addWiggleTransformSummaries(out map[string]string, prefix string) {
	addSummaries(out, prefix, map[string]string{
		"":                   "Wiggle Transform operator settings.",
		"anchor":             "Wiggle Transform anchor amount.",
		"position":           "Wiggle Transform position amount.",
		"scale":              "Wiggle Transform scale amount.",
		"rotation":           "Wiggle Transform rotation amount.",
		"wiggles_per_second": "Wiggle Transform frequency.",
		"random_seed":        "Wiggle Transform random seed.",
		"correlation":        "Wiggle Transform correlation.",
		"temporal_phase":     "Wiggle Transform temporal phase.",
		"spatial_phase":      "Wiggle Transform spatial phase.",
	})
}

func addMaskSummaries(out map[string]string, prefix, label string) {
	addSummaries(out, prefix, map[string]string{
		"name":                      label + " name.",
		"mode":                      label + " mode.",
		"inverted":                  label + " inverted switch.",
		"locked":                    label + " lock switch.",
		"color":                     label + " UI color as RGB channels.",
		"motion_blur":               label + " motion blur mode.",
		"feather_falloff":           label + " feather falloff mode.",
		"opacity":                   label + " opacity.",
		"feather":                   label + " feather vector.",
		"expansion":                 label + " expansion.",
		"closed":                    label + " closed path switch.",
		"vertices":                  label + " path vertices.",
		"path_keyframes[]":          label + " path keyframes.",
		"path_keyframes[].time":     label + " path keyframe time in seconds.",
		"path_keyframes[].vertices": label + " path keyframe vertices.",
	})
}

func addTransformSummaries(out map[string]string, prefix string) {
	addSummaries(out, prefix, map[string]string{
		"":             "Layer transform block.",
		"position":     "Static position value.",
		"scale":        "Static scale value.",
		"anchor_point": "Static anchor point value.",
		"rotation":     "Static rotation value.",
		"opacity":      "Static opacity value.",
		"expressions":  "Transform property expressions.",
	})
	addKeyframeSummaries(out, prefix+".position_keyframes[]", "Position keyframe", "vector")
	addKeyframeSummaries(out, prefix+".anchor_point_keyframes[]", "Anchor point keyframe", "vector")
	addKeyframeSummaries(out, prefix+".scale_keyframes[]", "Scale keyframe", "vector")
	addKeyframeSummaries(out, prefix+".rotation_keyframes[]", "Rotation keyframe", "scalar")
	addKeyframeSummaries(out, prefix+".opacity_keyframes[]", "Opacity keyframe", "scalar")
	addTransformExpressionSummaries(out, prefix+".expressions")
}

func addKeyframeSummaries(out map[string]string, prefix, label, valueKind string) {
	valueSummary := label + " value."
	if valueKind == "vector" {
		valueSummary = label + " vector value."
	}
	addSummaries(out, prefix, map[string]string{
		"":         label + " list.",
		"time":     label + " time in seconds.",
		"value":    valueSummary,
		"in_ease":  label + " incoming temporal ease.",
		"out_ease": label + " outgoing temporal ease.",
	})
	addTemporalEaseSummaries(out, prefix+".in_ease", label+" incoming ease")
	addTemporalEaseSummaries(out, prefix+".out_ease", label+" outgoing ease")
}

func addTemporalEaseSummaries(out map[string]string, prefix, label string) {
	addSummaries(out, prefix, map[string]string{
		"speed":     label + " speed.",
		"influence": label + " influence.",
	})
}

func addTransformExpressionSummaries(out map[string]string, prefix string) {
	addSummaries(out, prefix, map[string]string{
		"":             "Transform property expressions.",
		"position":     "Position expression settings.",
		"anchor_point": "Anchor point expression settings.",
		"scale":        "Scale expression settings.",
		"rotation":     "Rotation expression settings.",
		"opacity":      "Opacity expression settings.",
	})
	for _, name := range []string{"position", "anchor_point", "scale", "rotation", "opacity"} {
		addExpressionSummaries(out, prefix+"."+name, name+" transform expression")
	}
}

func addExpressionSummaries(out map[string]string, prefix, label string) {
	addSummaries(out, prefix, map[string]string{
		"source":  label + " source code.",
		"enabled": label + " enable switch.",
	})
}

func addEffectSummaries(out map[string]string, prefix, label string) {
	addSummaries(out, prefix, map[string]string{
		"match_name":          label + " match name.",
		"params[]":            label + " parameter assignments.",
		"params[].match_name": "Effect parameter match name.",
		"params[].value":      "Effect parameter value.",
		"params[].expression": "Effect parameter expression settings.",
	})
	addExpressionSummaries(out, prefix+".params[].expression", "effect parameter expression")
}

func addExpectedProfileSummaries(out map[string]string) {
	addSummaries(out, "expected_profile", map[string]string{
		"comp_count":        "Expected number of compositions.",
		"layer_count":       "Expected total number of layers.",
		"text_layer_count":  "Expected number of text layers.",
		"shape_layer_count": "Expected number of shape layers.",
		"effects[]":         "Expected effect checks.",
		"properties[]":      "Expected property checks.",
		"text_styles[]":     "Expected text style checks.",
		"keyframes[]":       "Expected keyframes for a layer property.",
		"masks[]":           "Expected mask checks.",
	})
	addCompSummaries(out, "expected_profile", "Expected composition")
	addExpectedLayerSummaries(out, "expected_profile.layers[]")
	addExpectedEffectSummaries(out, "expected_profile.effects[]")
	addExpectedPropertySummaries(out, "expected_profile.properties[]", "Expected property")
	addExpectedTextStyleSummaries(out, "expected_profile.text_styles[]")
	addExpectedKeyframeSummaries(out, "expected_profile.keyframes[]")
	addExpectedMaskSummaries(out, "expected_profile.masks[]")
}

func addExpectedLayerSummaries(out map[string]string, prefix string) {
	addSummaries(out, prefix, map[string]string{
		"name":              "Expected layer name.",
		"type":              "Expected layer type.",
		"quality":           "Expected layer quality mode.",
		"blending_mode":     "Expected layer blending mode.",
		"auto_orient":       "Expected layer auto-orientation mode.",
		"light_kind":        "Expected light type.",
		"source":            "Expected source item name.",
		"source_kind":       "Expected source item type.",
		"light_source":      "Expected source layer for light data.",
		"parent":            "Expected parent layer name.",
		"track_matte":       "Expected track matte mode.",
		"matte":             "Expected matte layer name.",
		"label":             "Expected layer label color index.",
		"comment":           "Expected layer comment text.",
		"timing":            "Expected layer timing checks.",
		"timing.start_time": "Expected layer start time in seconds.",
		"timing.in_point":   "Expected layer in point in seconds.",
		"timing.out_point":  "Expected layer out point in seconds.",
		"timing.duration":   "Expected layer duration in seconds.",
		"timing.stretch":    "Expected layer stretch percentage.",
		"flags":             "Expected layer switch checks.",
	})
	addExpectedFlagSummaries(out, prefix+".flags")
}

func addExpectedFlagSummaries(out map[string]string, prefix string) {
	addSummaries(out, prefix, map[string]string{
		"visible":                  "Expected layer video visibility switch.",
		"solo":                     "Expected layer solo switch.",
		"shy":                      "Expected layer shy switch.",
		"locked":                   "Expected layer lock switch.",
		"is_3d":                    "Expected layer 3D switch.",
		"is_adjustment":            "Expected adjustment-layer switch.",
		"is_null":                  "Expected null-layer switch.",
		"is_guide":                 "Expected guide-layer switch.",
		"motion_blur":              "Expected layer motion blur switch.",
		"effects_enabled":          "Expected layer effects enable switch.",
		"audio_enabled":            "Expected layer audio enable switch.",
		"frame_blend_enabled":      "Expected layer frame blending switch.",
		"markers_locked":           "Expected layer marker lock switch.",
		"frame_blend_pixel_motion": "Expected pixel-motion frame blending switch.",
		"collapse_transform":       "Expected collapse transformations switch.",
		"sampling_bicubic":         "Expected bicubic sampling switch.",
		"preserve_transparency":    "Expected preserve transparency switch.",
	})
}

func addExpectedEffectSummaries(out map[string]string, prefix string) {
	addSummaries(out, prefix, map[string]string{
		"layer_name":                  "Layer name to inspect for the expected effect.",
		"match_name":                  "Expected effect match name.",
		"params[]":                    "Expected effect parameter checks.",
		"params[].match_name":         "Expected effect parameter match name.",
		"params[].value":              "Expected effect parameter value.",
		"params[].expression":         "Expected effect parameter expression source.",
		"params[].expression_enabled": "Expected effect parameter expression enable switch.",
	})
}

func addExpectedPropertySummaries(out map[string]string, prefix, label string) {
	addSummaries(out, prefix, map[string]string{
		"layer_name":         label + " layer name.",
		"match_name":         label + " match name.",
		"value":              label + " value.",
		"expression":         label + " expression source.",
		"expression_enabled": label + " expression enable switch.",
	})
}

func addExpectedTextStyleSummaries(out map[string]string, prefix string) {
	addSummaries(out, prefix, map[string]string{
		"layer_name": "Layer name to inspect for the expected text style.",
	})
	addTextStyleSummaries(out, prefix, "Expected text style")
}

func addExpectedKeyframeSummaries(out map[string]string, prefix string) {
	addSummaries(out, prefix, map[string]string{
		"layer_name":        "Layer name to inspect for expected keyframes.",
		"match_name":        "Expected keyframed property match name.",
		"keyframes[]":       "Expected property keyframes.",
		"keyframes[].time":  "Expected property keyframe time in seconds.",
		"keyframes[].value": "Expected property keyframe value.",
	})
}

func addExpectedMaskSummaries(out map[string]string, prefix string) {
	addSummaries(out, prefix, map[string]string{
		"layer_name":                    "Layer name to inspect for the expected mask.",
		"name":                          "Expected mask name.",
		"mode":                          "Expected mask mode.",
		"inverted":                      "Expected mask inverted switch.",
		"locked":                        "Expected mask lock switch.",
		"color":                         "Expected mask UI color as RGB channels.",
		"motion_blur":                   "Expected mask motion blur mode.",
		"feather_falloff":               "Expected mask feather falloff mode.",
		"opacity":                       "Expected mask opacity.",
		"feather":                       "Expected mask feather vector.",
		"expansion":                     "Expected mask expansion.",
		"closed":                        "Expected mask closed path switch.",
		"vertex_count":                  "Expected mask vertex count.",
		"path_keyframes[]":              "Expected mask path keyframes.",
		"path_keyframes[].time":         "Expected mask path keyframe time in seconds.",
		"path_keyframes[].vertex_count": "Expected mask path keyframe vertex count.",
	})
}

func addSummaries(out map[string]string, prefix string, values map[string]string) {
	for name, summary := range values {
		path := prefix
		if name != "" {
			path += "." + name
		}
		out[path] = summary
	}
}
