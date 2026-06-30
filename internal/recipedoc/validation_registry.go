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
	"comps[].layers[].type": {
		Validation: "Supported values create the corresponding layer type.",
		Enum:       []string{"solid", "text", "shape", "camera", "light", "null", "adjustment"},
	},
	"comps[].layers[].text_animators[].property": {
		Validation: "The first recipe slice supports opacity text animators only.",
		Enum:       []string{"opacity"},
	},
	"comps[].layers[].text_animators[].value": {
		Validation: "Opacity animator value must be a number.",
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
		Validation: "Supported values are numbers, booleans, or numeric arrays.",
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
		Validation: "Each expected effect parameter needs at least one of value, expression, or expression_enabled.",
	},
	"expected_profile.effects[].params[].match_name": {
		Validation: "Expected effect parameter checks require a parameter match name.",
	},
	"expected_profile.effects[].params[].value": {
		Validation: "Expected parameter value must be a number, boolean, or numeric array.",
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
