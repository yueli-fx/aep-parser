package aep_test

import (
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// Golden cross-checked against py-aep samples/models/essential_graphics/*.json
// (motionGraphicsTemplateName / ...ControllerCount / ...ControllerNames).
// Controller types follow py-aep's CTyp enum. The EG panel lives on the
// "primary" comp; other comps default to "Untitled" / 0 controllers.
func TestEssentialGraphics(t *testing.T) {
	type ctrl struct {
		name string
		typ  aep.EGControllerType
	}
	cases := []struct {
		fixture  string
		template string
		ctrls    []ctrl
	}{
		{"eg_slider_controller", "Untitled", []ctrl{{"Intensity", aep.EGSlider}}},
		{"eg_color_controller", "Untitled", []ctrl{{"Custom Color", aep.EGColor}}},
		{"eg_checkbox_controller", "Untitled", []ctrl{{"Toggle Visibility", aep.EGCheckbox}}},
		{"eg_point_controller", "Untitled", []ctrl{{"Center Point", aep.EGPoint}}},
		{"eg_dropdown_controller", "Untitled", []ctrl{{"Style", aep.EGDropdown}}},
		{"eg_text_source_text", "Untitled", []ctrl{{"Title Text", aep.EGText}}},
		{"eg_custom_template_name", "My Custom Template", []ctrl{{"Fill Color", aep.EGColor}}},
		{"eg_controller_renamed", "Untitled", []ctrl{{"Renamed Color", aep.EGColor}}},
		{"eg_no_essential_properties", "Empty Template", nil},
		{"eg_multiple_controllers", "Untitled", []ctrl{
			{"Brightness", aep.EGSlider}, {"Layer Opacity", aep.EGSlider}, {"Background Color", aep.EGColor},
		}},
	}

	for _, c := range cases {
		t.Run(c.fixture, func(t *testing.T) {
			proj, err := aep.Open("../../test_data/" + c.fixture + ".aep")
			if err != nil {
				t.Skipf("%s.aep not present: %v", c.fixture, err)
			}
			comp := proj.CompositionByName("primary")
			if comp == nil {
				t.Fatal("comp 'primary' not found")
			}
			if comp.MotionGraphicsTemplateName != c.template {
				t.Errorf("MotionGraphicsTemplateName = %q, want %q", comp.MotionGraphicsTemplateName, c.template)
			}
			if got := comp.MotionGraphicsTemplateControllerCount(); got != len(c.ctrls) {
				t.Fatalf("ControllerCount = %d, want %d", got, len(c.ctrls))
			}
			for i, w := range c.ctrls {
				g := comp.EssentialGraphicsControllers[i]
				if g.Name != w.name {
					t.Errorf("controller[%d].Name = %q, want %q", i, g.Name, w.name)
				}
				if g.Type != w.typ {
					t.Errorf("controller[%d].Type = %d (%s), want %d (%s)", i, g.Type, g.Type, w.typ, w.typ)
				}
				if len(g.UUID) != 36 {
					t.Errorf("controller[%d].UUID = %q, want a 36-char GUID", i, g.UUID)
				}
			}
		})
	}
}

// A comp with no EG panel reports the AE default name and no controllers.
func TestEssentialGraphics_Default(t *testing.T) {
	proj, err := aep.Open("../../test_data/eg_slider_controller.aep")
	if err != nil {
		t.Skipf("fixture not present: %v", err)
	}
	main := proj.CompositionByName("main")
	if main == nil {
		t.Skip("comp 'main' not found")
	}
	if main.MotionGraphicsTemplateName != "Untitled" {
		t.Errorf("MotionGraphicsTemplateName = %q, want Untitled", main.MotionGraphicsTemplateName)
	}
	if main.MotionGraphicsTemplateControllerCount() != 0 {
		t.Errorf("ControllerCount = %d, want 0", main.MotionGraphicsTemplateControllerCount())
	}
	if names := main.MotionGraphicsTemplateControllerNames(); names != nil {
		t.Errorf("ControllerNames = %v, want nil", names)
	}
}

func TestEssentialGraphicsJSON(t *testing.T) {
	proj, err := aep.Open("../../test_data/eg_multiple_controllers.aep")
	if err != nil {
		t.Skipf("fixture not present: %v", err)
	}
	var sb strings.Builder
	if err := proj.WriteJSON(&sb); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	out := sb.String()
	for _, want := range []string{
		`"motion_graphics_template_name"`,
		`"Background Color"`,
		`"slider"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("JSON missing %s", want)
		}
	}
}
