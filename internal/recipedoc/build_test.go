package recipedoc

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildDocumentRejectsUnknownRegistryPath(t *testing.T) {
	_, err := buildDocumentWithRegistries(registries{
		Summary: map[string]string{
			"comps[].does_not_exist": "bad path",
		},
	})
	if err == nil {
		t.Fatal("expected unknown registry path error")
	}
}

func TestBuildDocumentRejectsUnknownCapabilityKey(t *testing.T) {
	_, err := buildDocumentWithRegistries(registries{
		CapabilitiesByPath: map[string][]string{
			"comps[].background_color": {"missing.capability_key"},
		},
	})
	if err == nil {
		t.Fatal("expected unknown capability key error")
	}
}

func TestBuildDocumentJoinsFieldMetadata(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}
	field := requireField(t, doc, "comps[].background_color")
	if field.Summary == "" {
		t.Fatal("summary should be joined")
	}
	if len(field.Capabilities) == 0 || field.Capabilities[0].Key != "comp.set_background_color" {
		t.Fatalf("capabilities not joined: %+v", field.Capabilities)
	}
}

func TestBuildDocumentIncludesCoreCapabilityMetadata(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string]string{
		"comps[]":                                          "comp.create",
		"comps[].label":                                    "comp.set_label",
		"comps[].motion_blur.shutter_angle":                "comp.set_motion_blur_shutter_angle",
		"comps[].work_area":                                "comp.set_work_area",
		"comps[].layers[].visible":                         "layer.set_visible",
		"comps[].layers[].parent":                          "layer.set_parent",
		"comps[].layers[].start_time":                      "layer.set_start_time",
		"comps[].layers[].text":                            "layer.set_text",
		"comps[].layers[].text_style.font_size":            "text.set_run_font_size",
		"comps[].layers[].camera.zoom":                     "camera.set_zoom",
		"comps[].layers[].camera.iris_highlight_threshold": "camera.set_iris_highlight_threshold",
		"comps[].layers[].light.intensity":                 "light.set_intensity",
		"comps[].layers[].light.source_layer":              "light.set_source_layer",
	}
	for path, key := range tests {
		field := requireField(t, doc, path)
		if !hasCapability(field, key) {
			t.Fatalf("%s capabilities = %+v, want key %q", path, field.Capabilities, key)
		}
	}
}

func TestBuildDocumentIncludesShapeCapabilityMetadata(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string]string{
		"comps[].layers[].shape.size":                                "shape.rect.set_size",
		"comps[].layers[].shape.kind":                                "shape.star.add",
		"comps[].layers[].shape.fill_color":                          "shape.fill.set_color",
		"comps[].layers[].shape.gradient_fill.color_stops[]":         "shape.gradient_fill.set_color_stops",
		"comps[].layers[].shape.gradient_stroke.line_join":           "shape.gradient_stroke.set_line_join",
		"comps[].layers[].shape.stroke.taper.start_length":           "shape.stroke_taper.set_start_length",
		"comps[].layers[].shape.repeater.position":                   "shape.repeater_transform.set_position",
		"comps[].layers[].shape.merge_paths.type":                    "shape.merge_paths.set_type",
		"comps[].layers[].shape.wiggle_paths.wiggles_per_second":     "shape.wiggle_paths.set_wiggles_per_second",
		"comps[].layers[].shape.wiggle_transform.wiggles_per_second": "shape.wiggle_transform.set_wiggles_per_second",
	}
	for path, key := range tests {
		field := requireField(t, doc, path)
		if !hasCapability(field, key) {
			t.Fatalf("%s capabilities = %+v, want key %q", path, field.Capabilities, key)
		}
	}
}

func TestBuildDocumentIncludesMaskCapabilityMetadata(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string]string{
		"comps[].layers[].masks[]":                           "mask.add",
		"comps[].layers[].masks[].name":                      "mask.add",
		"comps[].layers[].masks[].mode":                      "mask.set_mode",
		"comps[].layers[].masks[].inverted":                  "mask.set_inverted",
		"comps[].layers[].masks[].locked":                    "mask.set_locked",
		"comps[].layers[].masks[].color":                     "mask.set_color",
		"comps[].layers[].masks[].motion_blur":               "mask.set_motion_blur",
		"comps[].layers[].masks[].feather_falloff":           "mask.set_feather_falloff",
		"comps[].layers[].masks[].opacity":                   "mask.set_opacity",
		"comps[].layers[].masks[].feather":                   "mask.set_feather",
		"comps[].layers[].masks[].expansion":                 "mask.set_expansion",
		"comps[].layers[].masks[].closed":                    "mask.add",
		"comps[].layers[].masks[].vertices":                  "mask.add",
		"comps[].layers[].masks[].path_keyframes[]":          "mask.set_path_keyframes",
		"comps[].layers[].masks[].path_keyframes[].time":     "mask.set_path_keyframes",
		"comps[].layers[].masks[].path_keyframes[].vertices": "mask.set_path_keyframes",
	}
	for path, key := range tests {
		field := requireField(t, doc, path)
		if !hasCapability(field, key) {
			t.Fatalf("%s capabilities = %+v, want key %q", path, field.Capabilities, key)
		}
	}
}

func TestBuildDocumentIncludesCompValidationMetadata(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}

	tests := []string{
		"comps[].name",
		"comps[].width",
		"comps[].height",
		"comps[].frame_rate",
		"comps[].duration",
		"comps[].background_color",
		"comps[].label",
		"comps[].motion_graphics_template_name",
		"comps[].resolution_factor",
		"comps[].pixel_aspect",
		"comps[].display_start_time",
		"comps[].motion_blur.shutter_angle",
		"comps[].motion_blur.shutter_phase",
		"comps[].motion_blur.adaptive_sample_limit",
		"comps[].motion_blur.samples_per_frame",
		"comps[].work_area",
		"comps[].work_area.start",
		"comps[].work_area.end",
	}
	for _, path := range tests {
		field := requireField(t, doc, path)
		if field.Validation == "" && len(field.Enum) == 0 {
			t.Fatalf("%s has no validation metadata: %+v", path, field)
		}
	}
}

func TestBuildDocumentIncludesLayerValidationMetadata(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}

	tests := []string{
		"comps[].layers[].name",
		"comps[].layers[].label",
		"comps[].layers[].quality",
		"comps[].layers[].blending_mode",
		"comps[].layers[].track_matte",
		"comps[].layers[].auto_orient",
		"comps[].layers[].in_point",
		"comps[].layers[].out_point",
		"comps[].layers[].stretch",
	}
	for _, path := range tests {
		field := requireField(t, doc, path)
		if field.Validation == "" && len(field.Enum) == 0 {
			t.Fatalf("%s has no validation metadata: %+v", path, field)
		}
	}
}

func TestBuildDocumentIncludesTransformValidationMetadata(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}

	tests := []string{
		"comps[].layers[].transform.position",
		"comps[].layers[].transform.scale",
		"comps[].layers[].transform.anchor_point",
		"comps[].layers[].transform.position_keyframes[]",
		"comps[].layers[].transform.position_keyframes[].time",
		"comps[].layers[].transform.position_keyframes[].value",
		"comps[].layers[].transform.position_keyframes[].in_ease.influence",
		"comps[].layers[].transform.position_keyframes[].out_ease.influence",
		"comps[].layers[].transform.anchor_point_keyframes[]",
		"comps[].layers[].transform.anchor_point_keyframes[].time",
		"comps[].layers[].transform.anchor_point_keyframes[].value",
		"comps[].layers[].transform.scale_keyframes[]",
		"comps[].layers[].transform.scale_keyframes[].time",
		"comps[].layers[].transform.scale_keyframes[].value",
		"comps[].layers[].transform.rotation_keyframes[]",
		"comps[].layers[].transform.rotation_keyframes[].time",
		"comps[].layers[].transform.opacity_keyframes[]",
		"comps[].layers[].transform.opacity_keyframes[].time",
		"comps[].layers[].transform.opacity_keyframes[].value",
		"comps[].layers[].transform.expressions.position.source",
		"comps[].layers[].transform.expressions.anchor_point.source",
		"comps[].layers[].transform.expressions.scale.source",
		"comps[].layers[].transform.expressions.rotation.source",
		"comps[].layers[].transform.expressions.opacity.source",
	}
	for _, path := range tests {
		field := requireField(t, doc, path)
		if field.Validation == "" && len(field.Enum) == 0 {
			t.Fatalf("%s has no validation metadata: %+v", path, field)
		}
	}
}

func TestBuildDocumentIncludesKeyframeValidationMetadata(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}

	tests := []string{
		"comps[].layers[].effects[].params[].keyframes[]",
		"comps[].layers[].effects[].params[].keyframes[].time",
		"comps[].layers[].effects[].params[].keyframes[].value",
		"comps[].layers[].effects[].params[].keyframes[].in_ease.influence",
		"comps[].layers[].effects[].params[].keyframes[].out_ease.influence",
		"comps[].layers[].text_animators[].range_offset_keyframes[]",
		"comps[].layers[].text_animators[].range_offset_keyframes[].time",
		"comps[].layers[].text_animators[].range_offset_keyframes[].value",
		"comps[].layers[].text_animators[].range_offset_keyframes[].in_ease.influence",
		"comps[].layers[].text_animators[].range_offset_keyframes[].out_ease.influence",
		"comps[].layers[].text_animators[].value_keyframes[]",
		"comps[].layers[].text_animators[].value_keyframes[].time",
		"comps[].layers[].text_animators[].value_keyframes[].value",
		"comps[].layers[].text_animators[].value_keyframes[].in_ease.influence",
		"comps[].layers[].text_animators[].value_keyframes[].out_ease.influence",
		"comps[].layers[].masks[].path_keyframes[]",
		"comps[].layers[].masks[].path_keyframes[].time",
		"comps[].layers[].masks[].path_keyframes[].vertices",
	}
	for _, path := range tests {
		field := requireField(t, doc, path)
		if field.Validation == "" && len(field.Enum) == 0 {
			t.Fatalf("%s has no validation metadata: %+v", path, field)
		}
	}
}

func TestBuildDocumentIncludesTextStyleValidationMetadata(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}

	tests := []string{
		"comps[].layers[].text_style.run_index",
		"comps[].layers[].text_style.paragraph_index",
		"comps[].layers[].text_style.font_size",
		"comps[].layers[].text_style.fill_color",
		"comps[].layers[].text_style.tsume",
		"comps[].layers[].text_style.stroke_width",
		"comps[].layers[].text_style.justification",
	}
	for _, path := range tests {
		field := requireField(t, doc, path)
		if field.Validation == "" && len(field.Enum) == 0 {
			t.Fatalf("%s has no validation metadata: %+v", path, field)
		}
	}
}

func TestBuildDocumentIncludesMaskValidationMetadata(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}

	tests := []string{
		"comps[].layers[].masks[]",
		"comps[].layers[].masks[].mode",
		"comps[].layers[].masks[].color",
		"comps[].layers[].masks[].motion_blur",
		"comps[].layers[].masks[].feather_falloff",
		"comps[].layers[].masks[].opacity",
		"comps[].layers[].masks[].feather",
		"comps[].layers[].masks[].vertices",
	}
	for _, path := range tests {
		field := requireField(t, doc, path)
		if field.Validation == "" && len(field.Enum) == 0 {
			t.Fatalf("%s has no validation metadata: %+v", path, field)
		}
	}
}

func TestBuildDocumentIncludesLightValidationMetadata(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}

	tests := []string{
		"comps[].layers[].light",
		"comps[].layers[].light.kind",
		"comps[].layers[].light.source_layer",
		"comps[].layers[].light.color",
	}
	for _, path := range tests {
		field := requireField(t, doc, path)
		if field.Validation == "" && len(field.Enum) == 0 {
			t.Fatalf("%s has no validation metadata: %+v", path, field)
		}
	}
}

func TestBuildDocumentIncludesShapePaintValidationMetadata(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}

	tests := []string{
		"comps[].layers[].shape.kind",
		"comps[].layers[].shape.position",
		"comps[].layers[].shape.roundness",
		"comps[].layers[].shape.points",
		"comps[].layers[].shape.inner_radius",
		"comps[].layers[].shape.outer_radius",
		"comps[].layers[].shape.fill_opacity",
		"comps[].layers[].shape.fill_blend_mode",
		"comps[].layers[].shape.fill_composite_order",
		"comps[].layers[].shape.fill_rule",
		"comps[].layers[].shape.gradient_fill.type",
		"comps[].layers[].shape.gradient_fill.start_point",
		"comps[].layers[].shape.gradient_fill.end_point",
		"comps[].layers[].shape.gradient_fill.highlight_length",
		"comps[].layers[].shape.gradient_fill.color_stops[]",
		"comps[].layers[].shape.gradient_fill.color_stops[].offset",
		"comps[].layers[].shape.gradient_fill.color_stops[].midpoint",
		"comps[].layers[].shape.gradient_fill.color_stops[].color",
		"comps[].layers[].shape.gradient_fill.alpha_stops[]",
		"comps[].layers[].shape.gradient_fill.alpha_stops[].offset",
		"comps[].layers[].shape.gradient_fill.alpha_stops[].midpoint",
		"comps[].layers[].shape.gradient_fill.alpha_stops[].alpha",
		"comps[].layers[].shape.gradient_stroke.type",
		"comps[].layers[].shape.gradient_stroke.start_point",
		"comps[].layers[].shape.gradient_stroke.end_point",
		"comps[].layers[].shape.gradient_stroke.highlight_length",
		"comps[].layers[].shape.gradient_stroke.width",
		"comps[].layers[].shape.gradient_stroke.line_cap",
		"comps[].layers[].shape.gradient_stroke.line_join",
		"comps[].layers[].shape.gradient_stroke.miter_limit",
		"comps[].layers[].shape.gradient_stroke.color_stops[]",
		"comps[].layers[].shape.gradient_stroke.color_stops[].offset",
		"comps[].layers[].shape.gradient_stroke.color_stops[].midpoint",
		"comps[].layers[].shape.gradient_stroke.color_stops[].color",
		"comps[].layers[].shape.gradient_stroke.alpha_stops[]",
		"comps[].layers[].shape.gradient_stroke.alpha_stops[].offset",
		"comps[].layers[].shape.gradient_stroke.alpha_stops[].midpoint",
		"comps[].layers[].shape.gradient_stroke.alpha_stops[].alpha",
		"comps[].layers[].shape.stroke.color",
		"comps[].layers[].shape.stroke.width",
		"comps[].layers[].shape.stroke.opacity",
		"comps[].layers[].shape.stroke.line_cap",
		"comps[].layers[].shape.stroke.line_join",
		"comps[].layers[].shape.stroke.miter_limit",
		"comps[].layers[].shape.stroke.composite_order",
		"comps[].layers[].shape.stroke.dashes.dash",
		"comps[].layers[].shape.stroke.dashes.gap",
	}
	for _, path := range tests {
		field := requireField(t, doc, path)
		if field.Validation == "" && len(field.Enum) == 0 {
			t.Fatalf("%s has no validation metadata: %+v", path, field)
		}
	}
}

func TestBuildDocumentIncludesShapeOperatorValidationMetadata(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}

	tests := []string{
		"comps[].layers[].shape.trim.start",
		"comps[].layers[].shape.trim.end",
		"comps[].layers[].shape.round_corners.radius",
		"comps[].layers[].shape.offset_paths.line_join",
		"comps[].layers[].shape.offset_paths.miter_limit",
		"comps[].layers[].shape.offset_paths.copies",
		"comps[].layers[].shape.repeater.copies",
		"comps[].layers[].shape.repeater.order",
		"comps[].layers[].shape.repeater.anchor",
		"comps[].layers[].shape.repeater.position",
		"comps[].layers[].shape.repeater.scale",
		"comps[].layers[].shape.repeater.start_opacity",
		"comps[].layers[].shape.repeater.end_opacity",
		"comps[].layers[].shape.merge_paths.type",
		"comps[].layers[].shape.zigzag.size",
		"comps[].layers[].shape.zigzag.detail",
		"comps[].layers[].shape.zigzag.points",
		"comps[].layers[].shape.twist.center",
		"comps[].layers[].shape.wiggle_paths.points",
		"comps[].layers[].shape.wiggle_paths.correlation",
		"comps[].layers[].shape.wiggle_transform.anchor",
		"comps[].layers[].shape.wiggle_transform.position",
		"comps[].layers[].shape.wiggle_transform.scale",
		"comps[].layers[].shape.wiggle_transform.correlation",
	}
	for _, path := range tests {
		field := requireField(t, doc, path)
		if field.Validation == "" && len(field.Enum) == 0 {
			t.Fatalf("%s has no validation metadata: %+v", path, field)
		}
	}
}

func TestBuildDocumentIncludesExpectedProfileValidationMetadata(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}

	tests := []string{
		"expected_profile.comp_count",
		"expected_profile.width",
		"expected_profile.background_color",
		"expected_profile.display_start_time",
		"expected_profile.motion_blur.shutter_angle",
		"expected_profile.motion_blur.shutter_phase",
		"expected_profile.motion_blur.adaptive_sample_limit",
		"expected_profile.motion_blur.samples_per_frame",
		"expected_profile.work_area.start",
		"expected_profile.work_area.end",
		"expected_profile.comps[].name",
		"expected_profile.comps[].width",
		"expected_profile.comps[].height",
		"expected_profile.comps[].frame_rate",
		"expected_profile.comps[].duration",
		"expected_profile.comps[].label",
		"expected_profile.comps[].background_color",
		"expected_profile.comps[].resolution_factor",
		"expected_profile.comps[].pixel_aspect",
		"expected_profile.comps[].display_start_time",
		"expected_profile.comps[].motion_blur.shutter_angle",
		"expected_profile.comps[].motion_blur.shutter_phase",
		"expected_profile.comps[].motion_blur.adaptive_sample_limit",
		"expected_profile.comps[].motion_blur.samples_per_frame",
		"expected_profile.comps[].work_area.start",
		"expected_profile.comps[].work_area.end",
		"expected_profile.layers[].quality",
		"expected_profile.layers[].track_matte",
		"expected_profile.effects[].params[]",
		"expected_profile.effects[].params[].value",
		"expected_profile.properties[].value",
		"expected_profile.text_styles[].justification",
		"expected_profile.keyframes[].keyframes[]",
		"expected_profile.masks[].opacity",
	}
	for _, path := range tests {
		field := requireField(t, doc, path)
		if field.Validation == "" && len(field.Enum) == 0 {
			t.Fatalf("%s has no validation metadata: %+v", path, field)
		}
	}
}

func TestBuildDocumentIncludesExpectedProfileObjectValidationMetadata(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}

	tests := []string{
		"expected_profile.layers[].timing",
		"expected_profile.layers[].timing.start_time",
		"expected_profile.layers[].timing.in_point",
		"expected_profile.layers[].timing.out_point",
		"expected_profile.layers[].timing.duration",
		"expected_profile.layers[].timing.stretch",
		"expected_profile.effects[].params[]",
		"expected_profile.effects[].params[].expression",
		"expected_profile.effects[].params[].expression_enabled",
		"expected_profile.properties[]",
		"expected_profile.properties[].expression",
		"expected_profile.properties[].expression_enabled",
		"expected_profile.keyframes[]",
		"expected_profile.masks[]",
		"expected_profile.masks[].feather",
		"expected_profile.masks[].path_keyframes[]",
	}
	for _, path := range tests {
		field := requireField(t, doc, path)
		if field.Validation == "" && len(field.Enum) == 0 {
			t.Fatalf("%s has no validation metadata: %+v", path, field)
		}
	}
}

func TestBuildDocumentIncludesExampleMetadata(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}

	tests := map[string]string{
		"comps[].background_color":                               "examples/recipes/minimal-comp-background-color.json",
		"comps[].layers[].transform.position_keyframes[]":        "examples/recipes/minimal-transform-keyframes.json",
		"comps[].layers[].transform.expressions.position.source": "examples/recipes/minimal-transform-expression.json",
		"comps[].layers[].effects[]":                             "examples/recipes/minimal-text-effect.json",
		"comps[].layers[].effects[].params[].expression.source":  "examples/recipes/minimal-effect-param-expression.json",
		"comps[].layers[].masks[]":                               "examples/recipes/minimal-layer-mask.json",
		"expected_profile.masks[]":                               "examples/recipes/minimal-layer-mask.json",
	}
	for path, want := range tests {
		field := requireField(t, doc, path)
		if field.Example != want {
			t.Fatalf("%s example = %q, want %q", path, field.Example, want)
		}
	}
}

func TestBuildDocumentWithExampleDirDiscoversAdditionalExamples(t *testing.T) {
	doc, err := BuildDocumentWithExampleDir(filepath.Join("..", "..", "examples", "recipes"), "examples/recipes")
	if err != nil {
		t.Fatal(err)
	}

	label := requireField(t, doc, "comps[].label")
	if label.Example != "examples/recipes/minimal-comp-label.json" {
		t.Fatalf("comps[].label example = %q", label.Example)
	}

	background := requireField(t, doc, "comps[].background_color")
	if background.Example != "examples/recipes/minimal-comp-background-color.json" {
		t.Fatalf("explicit background example was overwritten: %q", background.Example)
	}

	expectedCompFields := []string{
		"expected_profile.comps[].label",
		"expected_profile.comps[].comment",
		"expected_profile.comps[].motion_graphics_template_name",
		"expected_profile.comps[].background_color",
		"expected_profile.comps[].resolution_factor",
		"expected_profile.comps[].pixel_aspect",
		"expected_profile.comps[].display_start_time",
		"expected_profile.comps[].renderer",
		"expected_profile.comps[].draft_3d",
		"expected_profile.comps[].frame_blending",
		"expected_profile.comps[].hide_shy_layers",
		"expected_profile.comps[].preserve_nested_frame_rate",
		"expected_profile.comps[].preserve_nested_resolution",
		"expected_profile.comps[].motion_blur.enabled",
		"expected_profile.comps[].motion_blur.shutter_angle",
		"expected_profile.comps[].motion_blur.shutter_phase",
		"expected_profile.comps[].motion_blur.adaptive_sample_limit",
		"expected_profile.comps[].motion_blur.samples_per_frame",
		"expected_profile.comps[].work_area.start",
		"expected_profile.comps[].work_area.end",
	}
	for _, path := range expectedCompFields {
		field := requireField(t, doc, path)
		if field.Example == "" {
			t.Fatalf("%s example should be discovered", path)
		}
	}
}

func TestBuildDocumentWithExampleDirCoversExpectedProfileExamples(t *testing.T) {
	doc, err := BuildDocumentWithExampleDir(filepath.Join("..", "..", "examples", "recipes"), "examples/recipes")
	if err != nil {
		t.Fatal(err)
	}

	var missing []string
	for _, field := range doc.Fields {
		if strings.HasPrefix(field.Path, "expected_profile.") && field.Example == "" {
			missing = append(missing, field.Path)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("expected_profile fields without examples: %s", strings.Join(missing, ", "))
	}
}

func TestBuildDocumentWithExampleDirCoversAllRecipeExamples(t *testing.T) {
	doc, err := BuildDocumentWithExampleDir(filepath.Join("..", "..", "examples", "recipes"), "examples/recipes")
	if err != nil {
		t.Fatal(err)
	}

	var missing []string
	for _, field := range doc.Fields {
		if field.Example == "" {
			missing = append(missing, field.Path)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("recipe fields without examples: %s", strings.Join(missing, ", "))
	}
}

func TestBuildDocumentExampleReferencesExist(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}

	for _, field := range doc.Fields {
		if field.Example == "" {
			continue
		}
		path := filepath.Join("..", "..", filepath.FromSlash(field.Example))
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("%s example %q is not readable: %v", field.Path, field.Example, err)
		}
	}
}

func TestBuildDocumentExampleReferencesCoverFieldPath(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}

	for _, field := range doc.Fields {
		if field.Example == "" {
			continue
		}
		path := filepath.Join("..", "..", filepath.FromSlash(field.Example))
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s example %q: %v", field.Path, field.Example, err)
		}
		var recipe any
		if err := json.Unmarshal(data, &recipe); err != nil {
			t.Fatalf("parse %s example %q: %v", field.Path, field.Example, err)
		}
		if !jsonPathExists(recipe, field.Path) {
			t.Fatalf("%s example %q does not contain field path", field.Path, field.Example)
		}
	}
}

func TestJSONPathExistsHandlesRecipeArrayPaths(t *testing.T) {
	value := map[string]any{
		"comps": []any{
			map[string]any{
				"layers": []any{
					map[string]any{
						"transform": map[string]any{
							"position": []any{1.0, 2.0},
						},
					},
				},
			},
		},
	}
	if !jsonPathExists(value, "comps[].layers[].transform.position") {
		t.Fatal("expected nested array path to exist")
	}
	if jsonPathExists(value, "comps[].layers[].effects[]") {
		t.Fatal("missing nested array path should not exist")
	}
}

func TestNoUnexpectedMissingSemanticSummaries(t *testing.T) {
	doc, err := BuildDocument()
	if err != nil {
		t.Fatal(err)
	}

	var missing []string
	for _, field := range doc.Fields {
		if field.MissingSemanticSummary {
			if _, ok := allowedMissingSemanticSummary[field.Path]; !ok {
				missing = append(missing, field.Path)
			}
		}
	}
	if len(missing) > 0 {
		t.Fatalf("unexpected missing semantic summaries:\n%s", strings.Join(missing, "\n"))
	}
}

func hasCapability(field FieldModel, key string) bool {
	for _, capRef := range field.Capabilities {
		if capRef.Key == key {
			return true
		}
	}
	return false
}
