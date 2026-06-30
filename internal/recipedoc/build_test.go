package recipedoc

import (
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
