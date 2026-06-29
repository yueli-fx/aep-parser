package recipedoc

import "testing"

func TestReflectRecipeFieldsBuildsStablePaths(t *testing.T) {
	doc, err := BuildStructuralModel()
	if err != nil {
		t.Fatal(err)
	}
	requireField(t, doc, "schema_version")
	requireField(t, doc, "project.name")
	requireField(t, doc, "comps[].name")
	requireField(t, doc, "comps[].layers[].type")
	requireField(t, doc, "comps[].layers[].transform.position_keyframes[].in_ease.influence")
	requireField(t, doc, "comps[].layers[].shape.gradient_fill.color_stops[].offset")
	requireField(t, doc, "expected_profile.keyframes[].keyframes[].value")
}

func TestReflectRecipeFieldsMarksStructuralRequiredness(t *testing.T) {
	doc, err := BuildStructuralModel()
	if err != nil {
		t.Fatal(err)
	}
	width := requireField(t, doc, "comps[].width")
	if !width.StructuralRequired {
		t.Fatalf("comps[].width should be structurally required")
	}
	background := requireField(t, doc, "comps[].background_color")
	if background.StructuralRequired {
		t.Fatalf("comps[].background_color should not be structurally required")
	}
}

func requireField(t *testing.T, doc Document, path string) FieldModel {
	t.Helper()
	for _, field := range doc.Fields {
		if field.Path == path {
			return field
		}
	}
	t.Fatalf("field %q not found", path)
	return FieldModel{}
}
