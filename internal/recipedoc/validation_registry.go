package recipedoc

var fieldValidation = map[string]FieldMeta{
	"schema_version": {
		Validation: "Must equal the supported recipe schema version.",
	},
	"project.target_version": {
		Validation: "Must be AE2020, AE2022, or AE2025. Omitted target_version defaults to AE2020.",
		Enum:       []string{"AE2020", "AE2022", "AE2025"},
	},
	"project.bits_per_channel": {
		Validation: "Must be 8, 16, or 32 bits per channel.",
		Enum:       []string{"8", "16", "32", "8bpc", "16bpc", "32bpc"},
	},
	"project.time_display_type": {
		Validation: "Must be timecode or frames.",
		Enum:       []string{"timecode", "frames"},
	},
	"project.frames_count_type": {
		Validation: "Must be start_0, start_1, or timecode_conversion.",
		Enum:       []string{"start_0", "start_1", "timecode_conversion"},
	},
	"project.feet_frames_film_type": {
		Validation: "Must be 35mm or 16mm.",
		Enum:       []string{"35mm", "16mm"},
	},
	"project.footage_timecode_display_start_type": {
		Validation: "Must be start_0 or source_media.",
		Enum:       []string{"start_0", "source_media"},
	},
	"project.expression_engine": {
		Validation: "Must be extendscript or javascript-1.0.",
		Enum:       []string{"extendscript", "javascript-1.0"},
	},
	"project.audio_sample_rate": {
		Validation: "Must be one of AE's supported audio sample rates.",
		Enum:       []string{"22050", "32000", "44100", "48000", "96000"},
	},
	"project.working_gamma": {
		Validation: "Must be 2.2 or 2.4.",
		Enum:       []string{"2.2", "2.4"},
	},
	"project.timecode_default_base": {
		Validation: "Must be an integer between 1 and 999.",
	},
	"comps[]": {
		Validation: "Recipe must include at least one composition.",
	},
	"comps[].name": {
		Validation: "Composition name is required and must be unique within comps[].",
	},
	"comps[].width": {
		Validation: "Composition width must be positive.",
	},
	"comps[].height": {
		Validation: "Composition height must be positive.",
	},
	"comps[].frame_rate": {
		Validation: "Composition frame rate must be positive.",
	},
	"comps[].duration": {
		Validation: "Composition duration must be positive.",
	},
	"comps[].background_color": {
		Validation: "RGB color must contain exactly three channels in the 0..255 range.",
	},
	"comps[].label": {
		Validation: "Composition label must be a valid AE label index.",
	},
	"comps[].motion_graphics_template_name": {
		Validation: "When present, must be non-empty. Empty strings are treated as omitted.",
	},
	"comps[].resolution_factor": {
		Validation: "Resolution factor must contain two positive integer factors.",
	},
	"comps[].pixel_aspect": {
		Validation: "Pixel aspect ratio must be positive.",
	},
	"comps[].display_start_time": {
		Validation: "Display start time must be non-negative.",
	},
	"comps[].motion_blur.shutter_angle": {
		Validation: "Motion blur shutter angle must be an integer between 0 and 720.",
	},
	"comps[].motion_blur.shutter_phase": {
		Validation: "Motion blur shutter phase must be an integer.",
	},
	"comps[].motion_blur.adaptive_sample_limit": {
		Validation: "Motion blur adaptive sample limit must be a non-negative integer.",
	},
	"comps[].motion_blur.samples_per_frame": {
		Validation: "Motion blur samples per frame must be a non-negative integer.",
	},
	"comps[].work_area": {
		Validation: "Work area start and end are required when work_area is present.",
	},
	"comps[].work_area.start": {
		Validation: "Work area start must be non-negative.",
	},
	"comps[].work_area.end": {
		Validation: "Work area end must satisfy start <= end <= comp duration.",
	},
	"comps[].layers[].type": {
		Validation: "Supported values create the corresponding layer type.",
		Enum:       []string{"solid", "text", "shape", "precomp", "camera", "light", "null", "adjustment"},
	},
	"comps[].layers[].name": {
		Validation: "Layer name is required.",
	},
	"comps[].layers[].source": {
		Validation: "Required for precomp layers and must name a different comp declared in comps[].",
	},
	"comps[].layers[].label": {
		Validation: "Layer label must be a valid AE label index.",
	},
	"comps[].layers[].quality": {
		Validation: "Layer quality must use a supported recipe quality value.",
		Enum:       []string{"wireframe", "draft", "best"},
	},
	"comps[].layers[].blending_mode": {
		Validation: "Layer blending mode must use a supported recipe blending mode value.",
	},
	"comps[].layers[].track_matte": {
		Validation: "Layer track matte must use a supported recipe track matte value.",
		Enum:       []string{"none", "alpha", "alpha_inverse", "luma", "luma_inverse"},
	},
	"comps[].layers[].auto_orient": {
		Validation: "Layer auto-orient mode must use a supported recipe auto-orient value.",
		Enum:       []string{"none", "along_path", "camera_or_point_of_interest", "characters_toward_camera"},
	},
	"comps[].layers[].in_point": {
		Validation: "Layer in point must be between 0 and comp duration.",
	},
	"comps[].layers[].out_point": {
		Validation: "Layer out point must be between 0 and comp duration and greater than or equal to in_point when both are present.",
	},
	"comps[].layers[].stretch": {
		Validation: "Layer stretch must be greater than 0.",
	},
	"comps[].layers[].camera": {
		Validation: "Camera options require a layer with type camera.",
	},
	"comps[].layers[].transform.position": {
		Validation: "When present, position must contain exactly two numeric values.",
	},
	"comps[].layers[].transform.scale": {
		Validation: "When present, scale must contain exactly two numeric values.",
	},
	"comps[].layers[].transform.anchor_point": {
		Validation: "When present, anchor_point must contain exactly two numeric values.",
	},
	"comps[].layers[].transform.position_keyframes[]": {
		Validation: "When present, keyframes must be sorted by time and each keyframe time must be within comp duration.",
	},
	"comps[].layers[].transform.position_keyframes[].time": {
		Validation: "Keyframe time must be non-negative, within comp duration, and sorted in ascending order.",
	},
	"comps[].layers[].transform.position_keyframes[].value": {
		Validation: "Position keyframe value must contain exactly two numeric values.",
	},
	"comps[].layers[].transform.position_keyframes[].in_ease.influence": {
		Validation: "Keyframe ease influence must be greater than 0 and at most 1.",
	},
	"comps[].layers[].transform.position_keyframes[].out_ease.influence": {
		Validation: "Keyframe ease influence must be greater than 0 and at most 1.",
	},
	"comps[].layers[].transform.anchor_point_keyframes[]": {
		Validation: "When present, keyframes must be sorted by time and each keyframe time must be within comp duration.",
	},
	"comps[].layers[].transform.anchor_point_keyframes[].time": {
		Validation: "Keyframe time must be non-negative, within comp duration, and sorted in ascending order.",
	},
	"comps[].layers[].transform.anchor_point_keyframes[].value": {
		Validation: "Anchor point keyframe value must contain exactly two numeric values.",
	},
	"comps[].layers[].transform.anchor_point_keyframes[].in_ease.influence": {
		Validation: "Keyframe ease influence must be greater than 0 and at most 1.",
	},
	"comps[].layers[].transform.anchor_point_keyframes[].out_ease.influence": {
		Validation: "Keyframe ease influence must be greater than 0 and at most 1.",
	},
	"comps[].layers[].transform.scale_keyframes[]": {
		Validation: "When present, keyframes must be sorted by time and each keyframe time must be within comp duration.",
	},
	"comps[].layers[].transform.scale_keyframes[].time": {
		Validation: "Keyframe time must be non-negative, within comp duration, and sorted in ascending order.",
	},
	"comps[].layers[].transform.scale_keyframes[].value": {
		Validation: "Scale keyframe value must contain exactly two numeric values.",
	},
	"comps[].layers[].transform.scale_keyframes[].in_ease.influence": {
		Validation: "Keyframe ease influence must be greater than 0 and at most 1.",
	},
	"comps[].layers[].transform.scale_keyframes[].out_ease.influence": {
		Validation: "Keyframe ease influence must be greater than 0 and at most 1.",
	},
	"comps[].layers[].transform.rotation_keyframes[]": {
		Validation: "When present, keyframes must be sorted by time and each keyframe time must be within comp duration.",
	},
	"comps[].layers[].transform.rotation_keyframes[].time": {
		Validation: "Keyframe time must be non-negative, within comp duration, and sorted in ascending order.",
	},
	"comps[].layers[].transform.rotation_keyframes[].in_ease.influence": {
		Validation: "Keyframe ease influence must be greater than 0 and at most 1.",
	},
	"comps[].layers[].transform.rotation_keyframes[].out_ease.influence": {
		Validation: "Keyframe ease influence must be greater than 0 and at most 1.",
	},
	"comps[].layers[].transform.opacity_keyframes[]": {
		Validation: "When present, keyframes must be sorted by time and each keyframe time must be within comp duration. Values must be between 0 and 100.",
	},
	"comps[].layers[].transform.opacity_keyframes[].time": {
		Validation: "Keyframe time must be non-negative, within comp duration, and sorted in ascending order.",
	},
	"comps[].layers[].transform.opacity_keyframes[].value": {
		Validation: "Opacity keyframe value must be between 0 and 100.",
	},
	"comps[].layers[].transform.opacity_keyframes[].in_ease.influence": {
		Validation: "Keyframe ease influence must be greater than 0 and at most 1.",
	},
	"comps[].layers[].transform.opacity_keyframes[].out_ease.influence": {
		Validation: "Keyframe ease influence must be greater than 0 and at most 1.",
	},
	"comps[].layers[].transform.expressions.position.source": {
		Validation: "Transform expression source is required when a position expression block is present.",
	},
	"comps[].layers[].transform.expressions.anchor_point.source": {
		Validation: "Transform expression source is required when an anchor_point expression block is present.",
	},
	"comps[].layers[].transform.expressions.scale.source": {
		Validation: "Transform expression source is required when a scale expression block is present.",
	},
	"comps[].layers[].transform.expressions.rotation.source": {
		Validation: "Transform expression source is required when a rotation expression block is present.",
	},
	"comps[].layers[].transform.expressions.opacity.source": {
		Validation: "Transform expression source is required when an opacity expression block is present.",
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
	"comps[].layers[].text_animators[].range_offset_keyframes[].time": {
		Validation: "Range offset keyframe time must be non-negative and sorted in ascending order.",
	},
	"comps[].layers[].text_animators[].range_offset_keyframes[].value": {
		Validation: "Range offset keyframe value must be numeric.",
	},
	"comps[].layers[].text_animators[].range_offset_keyframes[].in_ease.influence": {
		Validation: "Keyframe ease influence must be greater than 0 and at most 1.",
	},
	"comps[].layers[].text_animators[].range_offset_keyframes[].out_ease.influence": {
		Validation: "Keyframe ease influence must be greater than 0 and at most 1.",
	},
	"comps[].layers[].text_animators[].value_keyframes[]": {
		Validation: "When present, requires at least 2 keyframes sorted by non-negative time. Value is a number for opacity, rotation, tracking, and character_offset; a 3-number array for position and scale; a 3- or 4-number color array for color.",
	},
	"comps[].layers[].text_animators[].value_keyframes[].time": {
		Validation: "Text animator value keyframe time must be non-negative and sorted in ascending order.",
	},
	"comps[].layers[].text_animators[].value_keyframes[].value": {
		Validation: "Value type depends on the text animator property: scalar number, 3-number vector, or 3-/4-number color array.",
	},
	"comps[].layers[].text_animators[].value_keyframes[].in_ease.influence": {
		Validation: "Keyframe ease influence must be greater than 0 and at most 1.",
	},
	"comps[].layers[].text_animators[].value_keyframes[].out_ease.influence": {
		Validation: "Keyframe ease influence must be greater than 0 and at most 1.",
	},
	"comps[].layers[].text_style": {
		Validation: "Text style overrides are supported only on text layers.",
	},
	"comps[].layers[].text_style.run_index": {
		Validation: "Text style run_index must be non-negative.",
	},
	"comps[].layers[].text_style.paragraph_index": {
		Validation: "Text style paragraph_index must be non-negative.",
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
	"comps[].layers[].text_style.leading_type": {
		Validation: "Text paragraph leading type must use a supported value.",
		Enum:       []string{"roman", "japanese"},
	},
	"comps[].layers[].text_style.paragraph_direction": {
		Validation: "Text paragraph direction must use a supported value.",
		Enum:       []string{"ltr", "rtl"},
	},
	"comps[].layers[].text_animators[]": {
		Validation: "Text animators are supported only on text layers.",
	},
	"comps[].layers[].shape.kind": {
		Validation: "Shape kind must use a supported primitive type.",
		Enum:       []string{"rect", "ellipse", "star", "polygon"},
	},
	"comps[].layers[].shape.position": {
		Validation: "When present, shape position must contain exactly two numeric values.",
	},
	"comps[].layers[].shape.roundness": {
		Validation: "Shape roundness is supported only for rect shapes and must be non-negative.",
	},
	"comps[].layers[].shape.points": {
		Validation: "Star and polygon point count must be at least 3.",
	},
	"comps[].layers[].shape.inner_radius": {
		Validation: "Star inner_radius must be non-negative.",
	},
	"comps[].layers[].shape.outer_radius": {
		Validation: "Star outer_radius must be non-negative.",
	},
	"comps[].layers[].shape.fill_color": {
		Validation: "Fill color must contain RGB or RGBA channels in the 0..255 range.",
	},
	"comps[].layers[].shape.fill_opacity": {
		Validation: "Fill opacity must be between 0 and 100.",
	},
	"comps[].layers[].shape.fill_blend_mode": {
		Validation: "Fill blend mode must be an integer of at least 1.",
	},
	"comps[].layers[].shape.fill_composite_order": {
		Validation: "Fill composite order must use a supported value.",
		Enum:       []string{"above_previous", "below_previous"},
	},
	"comps[].layers[].shape.fill_rule": {
		Validation: "Fill rule must use a supported value.",
		Enum:       []string{"nonzero_winding", "even_odd"},
	},
	"comps[].layers[].shape.gradient_fill.type": {
		Validation: "Gradient fill type must use a supported value.",
		Enum:       []string{"linear", "radial"},
	},
	"comps[].layers[].shape.gradient_fill.start_point": {
		Validation: "Gradient fill start_point must contain exactly two numeric values.",
	},
	"comps[].layers[].shape.gradient_fill.end_point": {
		Validation: "Gradient fill end_point must contain exactly two numeric values.",
	},
	"comps[].layers[].shape.gradient_fill.highlight_length": {
		Validation: "Gradient fill highlight_length must be between -100 and 100.",
	},
	"comps[].layers[].shape.gradient_fill.color_stops[]": {
		Validation: "Gradient fill color_stops must include at least 2 stops.",
	},
	"comps[].layers[].shape.gradient_fill.color_stops[].offset": {
		Validation: "Gradient fill color stop offset must be between 0 and 1.",
	},
	"comps[].layers[].shape.gradient_fill.color_stops[].midpoint": {
		Validation: "Gradient fill color stop midpoint must be between 0 and 1.",
	},
	"comps[].layers[].shape.gradient_fill.color_stops[].color": {
		Validation: "Gradient fill color stop color must contain RGB or RGBA channels in the 0..255 range.",
	},
	"comps[].layers[].shape.gradient_fill.alpha_stops[]": {
		Validation: "Gradient fill alpha_stops must include at least 2 stops.",
	},
	"comps[].layers[].shape.gradient_fill.alpha_stops[].offset": {
		Validation: "Gradient fill alpha stop offset must be between 0 and 1.",
	},
	"comps[].layers[].shape.gradient_fill.alpha_stops[].midpoint": {
		Validation: "Gradient fill alpha stop midpoint must be between 0 and 1.",
	},
	"comps[].layers[].shape.gradient_fill.alpha_stops[].alpha": {
		Validation: "Gradient fill alpha stop alpha must be between 0 and 1.",
	},
	"comps[].layers[].shape.gradient_stroke.type": {
		Validation: "Gradient stroke type must use a supported value.",
		Enum:       []string{"linear", "radial"},
	},
	"comps[].layers[].shape.gradient_stroke.start_point": {
		Validation: "Gradient stroke start_point must contain exactly two numeric values.",
	},
	"comps[].layers[].shape.gradient_stroke.end_point": {
		Validation: "Gradient stroke end_point must contain exactly two numeric values.",
	},
	"comps[].layers[].shape.gradient_stroke.highlight_length": {
		Validation: "Gradient stroke highlight_length must be between -100 and 100.",
	},
	"comps[].layers[].shape.gradient_stroke.width": {
		Validation: "Gradient stroke width must be non-negative.",
	},
	"comps[].layers[].shape.gradient_stroke.line_cap": {
		Validation: "Gradient stroke line_cap must use a supported value.",
		Enum:       []string{"butt", "round", "projecting"},
	},
	"comps[].layers[].shape.gradient_stroke.line_join": {
		Validation: "Gradient stroke line_join must use a supported value.",
		Enum:       []string{"miter", "round", "bevel"},
	},
	"comps[].layers[].shape.gradient_stroke.miter_limit": {
		Validation: "Gradient stroke miter_limit must be at least 1.",
	},
	"comps[].layers[].shape.gradient_stroke.color_stops[]": {
		Validation: "Gradient stroke color_stops must include at least 2 stops.",
	},
	"comps[].layers[].shape.gradient_stroke.color_stops[].offset": {
		Validation: "Gradient stroke color stop offset must be between 0 and 1.",
	},
	"comps[].layers[].shape.gradient_stroke.color_stops[].midpoint": {
		Validation: "Gradient stroke color stop midpoint must be between 0 and 1.",
	},
	"comps[].layers[].shape.gradient_stroke.color_stops[].color": {
		Validation: "Gradient stroke color stop color must contain RGB or RGBA channels in the 0..255 range.",
	},
	"comps[].layers[].shape.gradient_stroke.alpha_stops[]": {
		Validation: "Gradient stroke alpha_stops must include at least 2 stops.",
	},
	"comps[].layers[].shape.gradient_stroke.alpha_stops[].offset": {
		Validation: "Gradient stroke alpha stop offset must be between 0 and 1.",
	},
	"comps[].layers[].shape.gradient_stroke.alpha_stops[].midpoint": {
		Validation: "Gradient stroke alpha stop midpoint must be between 0 and 1.",
	},
	"comps[].layers[].shape.gradient_stroke.alpha_stops[].alpha": {
		Validation: "Gradient stroke alpha stop alpha must be between 0 and 1.",
	},
	"comps[].layers[].shape.stroke.color": {
		Validation: "Stroke color must contain RGB or RGBA channels in the 0..255 range.",
	},
	"comps[].layers[].shape.stroke.width": {
		Validation: "Stroke width must be non-negative.",
	},
	"comps[].layers[].shape.stroke.opacity": {
		Validation: "Stroke opacity must be between 0 and 100.",
	},
	"comps[].layers[].shape.stroke.line_cap": {
		Validation: "Stroke line_cap must use a supported value.",
		Enum:       []string{"butt", "round", "projecting"},
	},
	"comps[].layers[].shape.stroke.line_join": {
		Validation: "Stroke line_join must use a supported value.",
		Enum:       []string{"miter", "round", "bevel"},
	},
	"comps[].layers[].shape.stroke.miter_limit": {
		Validation: "Stroke miter_limit must be at least 1.",
	},
	"comps[].layers[].shape.stroke.composite_order": {
		Validation: "Stroke composite_order must use a supported value.",
		Enum:       []string{"above_previous", "below_previous"},
	},
	"comps[].layers[].shape.stroke.dashes.dash": {
		Validation: "Stroke dash must be non-negative.",
	},
	"comps[].layers[].shape.stroke.dashes.gap": {
		Validation: "Stroke gap must be non-negative.",
	},
	"comps[].layers[].shape.trim.start": {
		Validation: "Trim start must be between 0 and 100.",
	},
	"comps[].layers[].shape.trim.end": {
		Validation: "Trim end must be between 0 and 100.",
	},
	"comps[].layers[].shape.round_corners.radius": {
		Validation: "Round corners radius must be non-negative.",
	},
	"comps[].layers[].shape.offset_paths.line_join": {
		Validation: "Offset paths line_join must use a supported value.",
		Enum:       []string{"miter", "round", "bevel"},
	},
	"comps[].layers[].shape.offset_paths.miter_limit": {
		Validation: "Offset paths miter_limit must be at least 1.",
	},
	"comps[].layers[].shape.offset_paths.copies": {
		Validation: "Offset paths copies must be at least 1.",
	},
	"comps[].layers[].shape.repeater.copies": {
		Validation: "Repeater copies must be at least 1.",
	},
	"comps[].layers[].shape.repeater.order": {
		Validation: "Repeater order must use a supported value.",
		Enum:       []string{"below", "above"},
	},
	"comps[].layers[].shape.repeater.anchor": {
		Validation: "Repeater anchor must contain exactly two numeric values.",
	},
	"comps[].layers[].shape.repeater.position": {
		Validation: "Repeater position must contain exactly two numeric values.",
	},
	"comps[].layers[].shape.repeater.scale": {
		Validation: "Repeater scale must contain exactly two numeric values.",
	},
	"comps[].layers[].shape.repeater.start_opacity": {
		Validation: "Repeater start_opacity must be between 0 and 100.",
	},
	"comps[].layers[].shape.repeater.end_opacity": {
		Validation: "Repeater end_opacity must be between 0 and 100.",
	},
	"comps[].layers[].shape.merge_paths.type": {
		Validation: "Merge paths type must use a supported value.",
		Enum:       []string{"merge", "add", "subtract", "intersect", "exclude"},
	},
	"comps[].layers[].shape.zigzag.size": {
		Validation: "Zigzag size must be non-negative.",
	},
	"comps[].layers[].shape.zigzag.detail": {
		Validation: "Zigzag detail must be non-negative.",
	},
	"comps[].layers[].shape.zigzag.points": {
		Validation: "Zigzag points must use a supported value.",
		Enum:       []string{"corner", "smooth"},
	},
	"comps[].layers[].shape.twist.center": {
		Validation: "Twist center must contain exactly two numeric values.",
	},
	"comps[].layers[].shape.wiggle_paths.points": {
		Validation: "Wiggle paths points must use a supported value.",
		Enum:       []string{"corner", "smooth"},
	},
	"comps[].layers[].shape.wiggle_paths.correlation": {
		Validation: "Wiggle paths correlation must be between 0 and 100.",
	},
	"comps[].layers[].shape.wiggle_transform.anchor": {
		Validation: "Wiggle transform anchor must contain exactly two numeric values.",
	},
	"comps[].layers[].shape.wiggle_transform.position": {
		Validation: "Wiggle transform position must contain exactly two numeric values.",
	},
	"comps[].layers[].shape.wiggle_transform.scale": {
		Validation: "Wiggle transform scale must contain exactly two numeric values.",
	},
	"comps[].layers[].shape.wiggle_transform.correlation": {
		Validation: "Wiggle transform correlation must be between 0 and 100.",
	},
	"comps[].layers[].masks[]": {
		Validation: "Masks are supported on AV, text, shape, solid, null, adjustment, and precomp layers; camera and light layers reject masks.",
	},
	"comps[].layers[].masks[].mode": {
		Validation: "Mask mode must use a supported AE mask mode value.",
		Enum:       []string{"add", "subtract", "intersect", "lighten", "darken", "difference", "none"},
	},
	"comps[].layers[].masks[].color": {
		Validation: "Mask color must contain exactly three channels in the 0..255 range.",
	},
	"comps[].layers[].masks[].motion_blur": {
		Validation: "Mask motion blur must use a supported value.",
		Enum:       []string{"same_as_layer", "on", "off"},
	},
	"comps[].layers[].masks[].feather_falloff": {
		Validation: "Mask feather falloff must use a supported value.",
		Enum:       []string{"smooth", "linear"},
	},
	"comps[].layers[].masks[].opacity": {
		Validation: "Mask opacity must be between 0 and 1.",
	},
	"comps[].layers[].masks[].feather": {
		Validation: "Mask feather must contain two non-negative numeric values.",
	},
	"comps[].layers[].masks[].vertices": {
		Validation: "Mask vertices must include at least 3 two-number points.",
	},
	"comps[].layers[].masks[].path_keyframes[]": {
		Validation: "When present, requires at least 2 keyframes sorted by time within comp duration.",
	},
	"comps[].layers[].masks[].path_keyframes[].time": {
		Validation: "Mask path keyframe time must be non-negative, within comp duration, and sorted in ascending order.",
	},
	"comps[].layers[].masks[].path_keyframes[].vertices": {
		Validation: "Mask path keyframe vertices must include at least 3 two-number points.",
	},
	"comps[].layers[].matte": {
		Validation: "Requires project.target_version AE2025, a non-none track_matte mode, and a same-comp source layer name.",
	},
	"comps[].layers[].light": {
		Validation: "Light options require a layer with type light.",
	},
	"comps[].layers[].light.kind": {
		Validation: "Light kind must use a supported recipe light type.",
		Enum:       []string{"parallel", "spot", "point", "ambient"},
	},
	"comps[].layers[].light.source_layer": {
		Validation: "Light source_layer must name a layer in the same comp.",
	},
	"comps[].layers[].light.color": {
		Validation: "Light color must contain RGB or RGBA channels in the 0..255 range.",
	},
	"comps[].layers[].effects[].match_name": {
		Validation: "Must be a supported effect match name for recipe compilation.",
	},
	"comps[].layers[].effects[]": {
		Validation: "Each effect requires a supported match_name.",
	},
	"comps[].layers[].effects[].params[].match_name": {
		Validation: "Must match a parameter exposed by the chosen effect.",
	},
	"comps[].layers[].effects[].params[]": {
		Validation: "Each effect parameter requires match_name. Layer-reference params cannot also set value, keyframes, or expression.",
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
	"comps[].layers[].effects[].params[].keyframes[].time": {
		Validation: "Effect parameter keyframe time must be non-negative and sorted in ascending order.",
	},
	"comps[].layers[].effects[].params[].keyframes[].value": {
		Validation: "Effect parameter keyframe values must be all scalar numbers or all 2-, 3-, or 4-number arrays.",
	},
	"comps[].layers[].effects[].params[].keyframes[].in_ease.influence": {
		Validation: "Keyframe ease influence must be greater than 0 and at most 1.",
	},
	"comps[].layers[].effects[].params[].keyframes[].out_ease.influence": {
		Validation: "Keyframe ease influence must be greater than 0 and at most 1.",
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
	"expected_profile.display_start_time": {
		Validation: "Expected display start time must be non-negative.",
	},
	"expected_profile.motion_blur.shutter_angle": {
		Validation: "Expected motion blur shutter angle must be an integer between 0 and 720.",
	},
	"expected_profile.motion_blur.shutter_phase": {
		Validation: "Expected motion blur shutter phase must be an integer.",
	},
	"expected_profile.motion_blur.adaptive_sample_limit": {
		Validation: "Expected motion blur adaptive sample limit must be a non-negative integer.",
	},
	"expected_profile.motion_blur.samples_per_frame": {
		Validation: "Expected motion blur samples per frame must be a non-negative integer.",
	},
	"expected_profile.work_area.start": {
		Validation: "Expected work area start must be non-negative.",
	},
	"expected_profile.work_area.end": {
		Validation: "Expected work area end must be non-negative and greater than or equal to start when both are present.",
	},
	"expected_profile.comps[].name": {
		Validation: "Expected composition checks require a target composition name.",
	},
	"expected_profile.comps[].width": {
		Validation: "Expected composition width must be a positive integer.",
	},
	"expected_profile.comps[].height": {
		Validation: "Expected composition height must be a positive integer.",
	},
	"expected_profile.comps[].frame_rate": {
		Validation: "Expected composition frame rate must be positive.",
	},
	"expected_profile.comps[].duration": {
		Validation: "Expected composition duration must be positive.",
	},
	"expected_profile.comps[].label": {
		Validation: "Expected composition label must be a valid AE label index.",
	},
	"expected_profile.comps[].background_color": {
		Validation: "Expected composition RGB color must contain exactly three channels in the 0..255 range.",
	},
	"expected_profile.comps[].resolution_factor": {
		Validation: "Expected composition resolution factor must contain two positive integer factors.",
	},
	"expected_profile.comps[].pixel_aspect": {
		Validation: "Expected composition pixel aspect ratio must be positive.",
	},
	"expected_profile.comps[].display_start_time": {
		Validation: "Expected composition display start time must be non-negative.",
	},
	"expected_profile.comps[].motion_blur.shutter_angle": {
		Validation: "Expected composition motion blur shutter angle must be an integer between 0 and 720.",
	},
	"expected_profile.comps[].motion_blur.shutter_phase": {
		Validation: "Expected composition motion blur shutter phase must be an integer.",
	},
	"expected_profile.comps[].motion_blur.adaptive_sample_limit": {
		Validation: "Expected composition motion blur adaptive sample limit must be a non-negative integer.",
	},
	"expected_profile.comps[].motion_blur.samples_per_frame": {
		Validation: "Expected composition motion blur samples per frame must be a non-negative integer.",
	},
	"expected_profile.comps[].work_area.start": {
		Validation: "Expected composition work area start must be non-negative.",
	},
	"expected_profile.comps[].work_area.end": {
		Validation: "Expected composition work area end must be non-negative and greater than or equal to start when both are present.",
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
	"expected_profile.layers[].timing.stretch": {
		Validation: "Expected layer stretch must be greater than 0.",
	},
	"expected_profile.layers[].timing": {
		Validation: "Expected layer timing values, when present, must satisfy non-negative time constraints.",
	},
	"expected_profile.layers[].timing.start_time": {
		Validation: "Expected layer start_time must be non-negative.",
	},
	"expected_profile.layers[].timing.in_point": {
		Validation: "Expected layer in_point must be non-negative.",
	},
	"expected_profile.layers[].timing.out_point": {
		Validation: "Expected layer out_point must be non-negative and greater than or equal to in_point when both are present.",
	},
	"expected_profile.layers[].timing.duration": {
		Validation: "Expected layer duration must be non-negative.",
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
	"expected_profile.effects[].params[].expression": {
		Validation: "Expected effect parameter expression checks compare the profiled expression source. Each expected param needs at least one assertion field.",
	},
	"expected_profile.effects[].params[].expression_enabled": {
		Validation: "Expected effect parameter expression_enabled checks compare the profiled expression enabled state. Each expected param needs at least one assertion field.",
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
	"expected_profile.properties[]": {
		Validation: "Each expected property needs at least one of value, expression, or expression_enabled.",
	},
	"expected_profile.properties[].value": {
		Validation: "Expected property value must be a number, boolean, or numeric array.",
	},
	"expected_profile.properties[].expression": {
		Validation: "Expected property expression checks compare the profiled expression source. Each expected property needs at least one assertion field.",
	},
	"expected_profile.properties[].expression_enabled": {
		Validation: "Expected property expression_enabled checks compare the profiled expression enabled state. Each expected property needs at least one assertion field.",
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
	"expected_profile.text_styles[].leading_type": {
		Validation: "Expected paragraph leading type must use a supported value.",
		Enum:       []string{"roman", "japanese"},
	},
	"expected_profile.text_styles[].paragraph_direction": {
		Validation: "Expected paragraph direction must use a supported value.",
		Enum:       []string{"ltr", "rtl"},
	},
	"expected_profile.bits_per_channel": {
		Validation: "Expected project color bit depth must be 8, 16, or 32 bits per channel.",
		Enum:       []string{"8", "16", "32", "8bpc", "16bpc", "32bpc"},
	},
	"expected_profile.time_display_type": {
		Validation: "Expected project time display mode must be supported.",
		Enum:       []string{"timecode", "frames"},
	},
	"expected_profile.frames_count_type": {
		Validation: "Expected project frame count mode must be supported.",
		Enum:       []string{"start_0", "start_1", "timecode_conversion"},
	},
	"expected_profile.feet_frames_film_type": {
		Validation: "Expected project feet+frames film type must be supported.",
		Enum:       []string{"35mm", "16mm"},
	},
	"expected_profile.footage_timecode_display_start_type": {
		Validation: "Expected project footage timecode display start mode must be supported.",
		Enum:       []string{"start_0", "source_media"},
	},
	"expected_profile.expression_engine": {
		Validation: "Expected project expression engine must be supported.",
		Enum:       []string{"extendscript", "javascript-1.0"},
	},
	"expected_profile.audio_sample_rate": {
		Validation: "Expected project audio sample rate must be one of AE's supported rates.",
		Enum:       []string{"22050", "32000", "44100", "48000", "96000"},
	},
	"expected_profile.working_gamma": {
		Validation: "Expected project working gamma must be 2.2 or 2.4.",
		Enum:       []string{"2.2", "2.4"},
	},
	"expected_profile.timecode_default_base": {
		Validation: "Expected project timecode default base must be an integer between 1 and 999.",
	},
	"expected_profile.keyframes[].layer_name": {
		Validation: "Keyframe profile checks require a target layer name.",
	},
	"expected_profile.keyframes[]": {
		Validation: "Each expected keyframed property requires a layer name, property match name, and at least one keyframe.",
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
	"expected_profile.masks[]": {
		Validation: "Each expected mask requires a layer name; optional mask fields are compared when present.",
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
	"expected_profile.masks[].feather": {
		Validation: "Expected mask feather must contain two non-negative numeric values.",
	},
	"expected_profile.masks[].vertex_count": {
		Validation: "Expected mask vertex count must be non-negative.",
	},
	"expected_profile.masks[].path_keyframes[]": {
		Validation: "Expected mask path keyframes must be sorted by non-negative time.",
	},
	"expected_profile.masks[].path_keyframes[].time": {
		Validation: "Expected mask path keyframe time must be non-negative and sorted in ascending order.",
	},
	"expected_profile.masks[].path_keyframes[].vertex_count": {
		Validation: "Expected mask path keyframe vertex count must be non-negative.",
	},
}
