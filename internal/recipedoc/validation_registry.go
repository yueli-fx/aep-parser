package recipedoc

var fieldValidation = map[string]FieldMeta{
	"schema_version": {
		Validation: "Must equal the supported recipe schema version.",
	},
	"project.target_version": {
		Validation: "Must be AE2020, AE2022, or AE2025. Omitted target_version defaults to AE2020.",
		Enum:       []string{"AE2020", "AE2022", "AE2025"},
	},
	"comps[].width": {
		Validation: "Pixel width must fit the composition writer's uint16 range.",
	},
	"comps[].height": {
		Validation: "Pixel height must fit the composition writer's uint16 range.",
	},
	"comps[].background_color": {
		Validation: "RGB color must contain exactly three channels in the 0..255 range.",
	},
	"comps[].motion_graphics_template_name": {
		Validation: "When present, must be non-empty. Empty strings are treated as omitted.",
	},
	"comps[].layers[].type": {
		Validation: "Supported values create the corresponding layer type.",
		Enum:       []string{"solid", "text", "shape", "camera", "light", "null", "adjustment"},
	},
	"comps[].layers[].text_animators[].property": {
		Validation: "Supported values create the corresponding text animator property.",
		Enum:       []string{"opacity", "position", "scale", "rotation", "color", "tracking", "character_offset", "fill_opacity", "stroke_opacity", "stroke_width", "skew", "rotation_x", "rotation_y", "stroke_color"},
	},
	"comps[].layers[].text_animators[].value": {
		Validation: "Opacity, rotation, tracking, character_offset, fill_opacity, stroke_opacity, stroke_width, skew, rotation_x, and rotation_y values must be numbers; position and scale values must be 3-number arrays; color and stroke_color values must be 3- or 4-number arrays.",
	},
	"comps[].layers[].text_animators[].range_start": {
		Validation: "Range selector start is required.",
	},
	"comps[].layers[].text_animators[].range_end": {
		Validation: "Range selector end is required.",
	},
	"comps[].layers[].text_animators[].range_offset": {
		Validation: "Range selector offset is required.",
	},
	"comps[].layers[].text_animators[].range_offset_keyframes[]": {
		Validation: "When present, requires at least 2 keyframes sorted by non-negative time.",
	},
	"comps[].layers[].text_animators[].value_keyframes[]": {
		Validation: "When present, requires at least 2 keyframes sorted by non-negative time. Value is a number for opacity, rotation, tracking, and character_offset; a 3-number array for position and scale; a 3- or 4-number color array for color.",
	},
	"comps[].layers[].text_style.font_size": {
		Validation: "Text font size must be positive.",
	},
	"comps[].layers[].text_style.fill_color": {
		Validation: "Text fill color must contain RGB or RGBA channels in the 0..255 range.",
	},
	"comps[].layers[].text_style.tsume": {
		Validation: "Text tsume must be between 0 and 100.",
	},
	"comps[].layers[].text_style.caps_option": {
		Validation: "Text caps option must use a supported value.",
		Enum:       []string{"normal", "small_caps", "all_caps", "all_small_caps"},
	},
	"comps[].layers[].text_style.baseline_option": {
		Validation: "Text baseline option must use a supported value.",
		Enum:       []string{"normal", "superscript", "subscript"},
	},
	"comps[].layers[].text_style.auto_kern_type": {
		Validation: "Text auto-kerning type must use a supported value.",
		Enum:       []string{"no_auto", "metric", "optical"},
	},
	"comps[].layers[].text_style.line_join_type": {
		Validation: "Text line-join type must use a supported value.",
		Enum:       []string{"miter", "round", "bevel"},
	},
	"comps[].layers[].text_style.digit_set": {
		Validation: "Text digit set must use a supported value.",
		Enum:       []string{"default", "arabic", "hindi", "farsi", "arabic_rtl"},
	},
	"comps[].layers[].text_style.stroke_color": {
		Validation: "Text stroke color must contain RGB or RGBA channels in the 0..255 range.",
	},
	"comps[].layers[].text_style.stroke_width": {
		Validation: "Text stroke width must be non-negative.",
	},
	"comps[].layers[].text_style.justification": {
		Validation: "Text justification must use a supported value.",
		Enum:       []string{"left", "right", "center"},
	},
	"comps[].layers[].matte": {
		Validation: "Requires project.target_version AE2025, a non-none track_matte mode, and a same-comp source layer name.",
	},
	"comps[].layers[].effects[].match_name": {
		Validation: "Must be a supported effect match name for recipe compilation.",
	},
	"comps[].layers[].effects[].params[].match_name": {
		Validation: "Must match a parameter exposed by the chosen effect.",
	},
	"comps[].layers[].effects[].params[].value": {
		Validation: "Supported values are numbers, booleans, or numeric arrays. Optional when keyframes are present.",
	},
	"comps[].layers[].effects[].params[].target_layer": {
		Validation: "For layer-reference effect parameters, must name a layer in the same comp. Cannot be combined with value, keyframes, or expression.",
	},
	"comps[].layers[].effects[].params[].keyframes[]": {
		Validation: "When present, requires at least 2 keyframes sorted by non-negative time. Values must be all scalar numbers or all 2-, 3-, or 4-number arrays.",
	},
	"comps[].layers[].effects[].params[].essential_graphics": {
		Validation: "First slice requires a static value on the same parameter and cannot be combined with target_layer, keyframes, or expression.",
	},
	"comps[].layers[].effects[].params[].essential_graphics.name": {
		Validation: "Optional controller display name. Empty uses the effect parameter's own name.",
	},
	"comps[].layers[].effects[].params[].expression.source": {
		Validation: "Expression source is applied only when non-empty.",
	},
	"expected_profile.comp_count": {
		Validation: "Expected count must be non-negative.",
	},
	"expected_profile.layer_count": {
		Validation: "Expected count must be non-negative.",
	},
	"expected_profile.text_layer_count": {
		Validation: "Expected count must be non-negative.",
	},
	"expected_profile.shape_layer_count": {
		Validation: "Expected count must be non-negative.",
	},
	"expected_profile.width": {
		Validation: "Expected width must be a positive integer.",
	},
	"expected_profile.height": {
		Validation: "Expected height must be a positive integer.",
	},
	"expected_profile.frame_rate": {
		Validation: "Expected frame rate must be positive.",
	},
	"expected_profile.duration": {
		Validation: "Expected duration must be positive.",
	},
	"expected_profile.label": {
		Validation: "Expected label must be a valid AE label index.",
	},
	"expected_profile.background_color": {
		Validation: "Expected RGB color must contain exactly three channels in the 0..255 range.",
	},
	"expected_profile.motion_graphics_template_name": {
		Validation: "When present, checks the first profiled composition's Motion Graphics template name.",
	},
	"expected_profile.resolution_factor": {
		Validation: "Expected resolution factor must contain two positive integer factors.",
	},
	"expected_profile.pixel_aspect": {
		Validation: "Expected pixel aspect ratio must be positive.",
	},
	"expected_profile.layers[].name": {
		Validation: "Layer profile checks require a target layer name.",
	},
	"expected_profile.layers[].label": {
		Validation: "Expected layer label must be a valid AE label index.",
	},
	"expected_profile.layers[].quality": {
		Validation: "Expected layer quality must use a supported recipe quality value.",
		Enum:       []string{"wireframe", "draft", "best"},
	},
	"expected_profile.layers[].blending_mode": {
		Validation: "Expected blending mode must use a supported recipe blending mode value.",
	},
	"expected_profile.layers[].track_matte": {
		Validation: "Expected track matte must use a supported recipe track matte value.",
		Enum:       []string{"none", "alpha", "alpha_inverse", "luma", "luma_inverse"},
	},
	"expected_profile.layers[].auto_orient": {
		Validation: "Expected auto-orient mode must use a supported recipe auto-orient value.",
		Enum:       []string{"none", "along_path", "camera_or_point_of_interest", "characters_toward_camera"},
	},
	"expected_profile.effects[].layer_name": {
		Validation: "Effect profile checks require a target layer name.",
	},
	"expected_profile.effects[].match_name": {
		Validation: "Effect profile checks require an effect match name.",
	},
	"expected_profile.effects[].params[]": {
		Validation: "Each expected effect parameter needs at least one of value, target_layer, expression, or expression_enabled.",
	},
	"expected_profile.effects[].params[].match_name": {
		Validation: "Expected effect parameter checks require a parameter match name.",
	},
	"expected_profile.effects[].params[].value": {
		Validation: "Expected parameter value must be a number, boolean, or numeric array.",
	},
	"expected_profile.effects[].params[].target_layer": {
		Validation: "Expected target layer checks compare against the profiled effect parameter layer reference name.",
	},
	"expected_profile.essential_graphics[].name": {
		Validation: "Essential Graphics controller checks require the expected controller name.",
	},
	"expected_profile.essential_graphics[].type": {
		Validation: "Expected controller type must be checkbox, slider, color, point, text, comment, multidimensional, group, dropdown, or unknown.",
		Enum:       []string{"checkbox", "slider", "color", "point", "text", "comment", "multidimensional", "group", "dropdown", "unknown"},
	},
	"expected_profile.properties[].layer_name": {
		Validation: "Property profile checks require a target layer name.",
	},
	"expected_profile.properties[].match_name": {
		Validation: "Property profile checks require a property match name.",
	},
	"expected_profile.properties[].value": {
		Validation: "Expected property value must be a number, boolean, or numeric array.",
	},
	"expected_profile.text_styles[].layer_name": {
		Validation: "Text style profile checks require a target layer name.",
	},
	"expected_profile.text_styles[].run_index": {
		Validation: "Expected text style run index must be non-negative.",
	},
	"expected_profile.text_styles[].paragraph_index": {
		Validation: "Expected text style paragraph index must be non-negative.",
	},
	"expected_profile.text_styles[].font_size": {
		Validation: "Expected font size must be positive.",
	},
	"expected_profile.text_styles[].fill_color": {
		Validation: "Expected fill color must contain RGB or RGBA channels in the 0..1 range.",
	},
	"expected_profile.text_styles[].tsume": {
		Validation: "Expected tsume must be between 0 and 100.",
	},
	"expected_profile.text_styles[].caps_option": {
		Validation: "Expected caps option must use a supported value.",
		Enum:       []string{"normal", "small_caps", "all_caps", "all_small_caps"},
	},
	"expected_profile.text_styles[].baseline_option": {
		Validation: "Expected baseline option must use a supported value.",
		Enum:       []string{"normal", "superscript", "subscript"},
	},
	"expected_profile.text_styles[].auto_kern_type": {
		Validation: "Expected auto-kerning type must use a supported value.",
		Enum:       []string{"no_auto", "metric", "optical"},
	},
	"expected_profile.text_styles[].line_join_type": {
		Validation: "Expected line-join type must use a supported value.",
		Enum:       []string{"miter", "round", "bevel"},
	},
	"expected_profile.text_styles[].digit_set": {
		Validation: "Expected digit set must use a supported value.",
		Enum:       []string{"default", "arabic", "hindi", "farsi", "arabic_rtl"},
	},
	"expected_profile.text_styles[].stroke_color": {
		Validation: "Expected stroke color must contain RGB or RGBA channels in the 0..1 range.",
	},
	"expected_profile.text_styles[].stroke_width": {
		Validation: "Expected stroke width must be non-negative.",
	},
	"expected_profile.text_styles[].justification": {
		Validation: "Expected justification must use a supported text justification value.",
		Enum:       []string{"left", "right", "center"},
	},
	"expected_profile.keyframes[].layer_name": {
		Validation: "Keyframe profile checks require a target layer name.",
	},
	"expected_profile.keyframes[].match_name": {
		Validation: "Keyframe profile checks require a property match name.",
	},
	"expected_profile.keyframes[].keyframes[]": {
		Validation: "Expected keyframes must contain at least one keyframe and be sorted by time.",
	},
	"expected_profile.keyframes[].keyframes[].time": {
		Validation: "Expected keyframe time must be non-negative and sorted in ascending order.",
	},
	"expected_profile.keyframes[].keyframes[].value": {
		Validation: "Expected keyframe value must be a number, boolean, or numeric array.",
	},
	"expected_profile.masks[].layer_name": {
		Validation: "Mask profile checks require a target layer name.",
	},
	"expected_profile.masks[].mode": {
		Validation: "Expected mask mode must use a supported mask mode value.",
		Enum:       []string{"add", "subtract", "intersect", "lighten", "darken", "difference", "none"},
	},
	"expected_profile.masks[].color": {
		Validation: "Expected mask color must contain exactly three channels in the 0..255 range.",
	},
	"expected_profile.masks[].motion_blur": {
		Validation: "Expected mask motion blur must use a supported mask motion blur value.",
		Enum:       []string{"same_as_layer", "on", "off"},
	},
	"expected_profile.masks[].feather_falloff": {
		Validation: "Expected mask feather falloff must use a supported feather falloff value.",
		Enum:       []string{"smooth", "linear"},
	},
	"expected_profile.masks[].opacity": {
		Validation: "Expected mask opacity must be between 0 and 1.",
	},
	"expected_profile.masks[].vertex_count": {
		Validation: "Expected mask vertex count must be non-negative.",
	},
	"expected_profile.masks[].path_keyframes[].time": {
		Validation: "Expected mask path keyframe time must be non-negative and sorted in ascending order.",
	},
	"expected_profile.masks[].path_keyframes[].vertex_count": {
		Validation: "Expected mask path keyframe vertex count must be non-negative.",
	},
}
