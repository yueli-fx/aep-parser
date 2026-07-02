package aepmigrate

import "testing"

func TestConvertProjectMaterializersAreExplicitAndOrdered(t *testing.T) {
	steps := convertProjectMaterializers()
	var names []string
	for _, step := range steps {
		if step.apply == nil {
			t.Fatalf("materializer %q has nil apply function", step.name)
		}
		names = append(names, step.name)
	}
	want := []string{
		"comp_metadata",
		"transform_expressions",
		"text_animators",
		"masks",
		"effects",
	}
	if len(names) != len(want) {
		t.Fatalf("materializer names = %+v, want %+v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("materializer names = %+v, want %+v", names, want)
		}
	}
}
