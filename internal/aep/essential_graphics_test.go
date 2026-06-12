package aep_test

import (
	"bytes"
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

// Rename round-trip on an AE-native fixture: the new name must land in all
// three persisted panel generations (CIFO/CIF2/CIF3 × CpS2+CapS = 6 slots) and
// the controller must survive untouched.
func TestSetMotionGraphicsTemplateName(t *testing.T) {
	proj, err := aep.Open("../../test_data/eg_custom_template_name.aep")
	if err != nil {
		t.Skipf("fixture not present: %v", err)
	}
	comp := proj.CompositionByName("primary")
	if comp == nil {
		t.Fatal("comp 'primary' not found")
	}
	if comp.MotionGraphicsTemplateName != "My Custom Template" {
		t.Fatalf("precondition: template name = %q", comp.MotionGraphicsTemplateName)
	}

	const newName = "GoRenamedTemplate"
	if err := comp.SetMotionGraphicsTemplateName(newName); err != nil {
		t.Fatalf("SetMotionGraphicsTemplateName: %v", err)
	}
	if comp.MotionGraphicsTemplateName != newName {
		t.Errorf("scene field = %q, want %q", comp.MotionGraphicsTemplateName, newName)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	raw := buf.Bytes()
	if got := bytes.Count(raw, []byte(newName)); got != 6 {
		t.Errorf("new name appears %d times in output, want 6 (3 generations x CpS2+CapS)", got)
	}
	if bytes.Contains(raw, []byte("My Custom Template")) {
		t.Error("old template name still present in output")
	}

	re, err := aep.FromReader(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	rc := re.CompositionByName("primary")
	if rc == nil {
		t.Fatal("re-parse: comp 'primary' not found")
	}
	if rc.MotionGraphicsTemplateName != newName {
		t.Errorf("re-parse template name = %q, want %q", rc.MotionGraphicsTemplateName, newName)
	}
	if got := rc.MotionGraphicsTemplateControllerCount(); got != 1 {
		t.Errorf("re-parse ControllerCount = %d, want 1", got)
	}
	if got := rc.EssentialGraphicsControllers[0].Name; got != "Fill Color" {
		t.Errorf("re-parse controller name = %q, want Fill Color", got)
	}
}

// A Go-built comp carries the EG panel shell from the embed template, so the
// rename works on it too; an empty name is refused.
func TestSetMotionGraphicsTemplateName_GoBuilt(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "built", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if err := comp.SetMotionGraphicsTemplateName(""); err == nil {
		t.Error("empty name accepted, want refusal")
	}
	if err := comp.SetMotionGraphicsTemplateName("BuiltTemplate"); err != nil {
		t.Fatalf("SetMotionGraphicsTemplateName: %v", err)
	}

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	rc := re.CompositionByName("built")
	if rc == nil {
		t.Fatal("re-parse: comp 'built' not found")
	}
	if rc.MotionGraphicsTemplateName != "BuiltTemplate" {
		t.Errorf("re-parse template name = %q, want BuiltTemplate", rc.MotionGraphicsTemplateName)
	}
}

// All-Go-built flow: build a project from scratch, add a Slider Control via
// AddEffect, expose its Slider param in the EG panel, and verify the
// controller round-trips (the path AE-native fixtures can't cover — no ID /
// shell coincidences possible).
func TestAddEssentialProperty_GoBuilt(t *testing.T) {
	p := aep.NewProject()
	comp, err := aep.NewComposition(p, "egcomp", 1920, 1080, 30, 5)
	if err != nil {
		t.Fatalf("NewComposition: %v", err)
	}
	if _, err := aep.NewSolidLayer(comp, "host", 1920, 1080, [3]float64{1, 0, 0}); err != nil {
		t.Fatalf("NewSolidLayer: %v", err)
	}
	p, err = aep.Reopen(p)
	if err != nil {
		t.Fatalf("Reopen: %v", err)
	}
	comp = p.CompositionByName("egcomp")
	if comp == nil || len(comp.Layers) == 0 {
		t.Fatal("re-parse: comp/layer missing")
	}
	layer := comp.Layers[0]
	fx, err := aep.AddEffect(layer, aep.EffectSliderControl)
	if err != nil {
		t.Fatalf("AddEffect: %v", err)
	}

	ctrl, err := aep.AddEssentialProperty(layer, fx, "ADBE Slider Control-0001", "Exposed Slider")
	if err != nil {
		t.Fatalf("AddEssentialProperty: %v", err)
	}
	if ctrl.Name != "Exposed Slider" || ctrl.Type != aep.EGSlider || len(ctrl.UUID) != 36 {
		t.Errorf("controller = %+v, want {Exposed Slider, slider, 36-char uuid}", ctrl)
	}
	if got := comp.MotionGraphicsTemplateControllerCount(); got != 1 {
		t.Errorf("ControllerCount = %d, want 1", got)
	}
	if err := comp.SetMotionGraphicsTemplateName("GoBuilt EG"); err != nil {
		t.Fatalf("SetMotionGraphicsTemplateName: %v", err)
	}

	var buf bytes.Buffer
	if err := p.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	if len(re.Warnings) != 0 {
		t.Fatalf("re-parse warnings: %v", re.Warnings)
	}
	rc := re.CompositionByName("egcomp")
	if rc == nil {
		t.Fatal("re-parse: comp missing")
	}
	if rc.MotionGraphicsTemplateName != "GoBuilt EG" {
		t.Errorf("re-parse template name = %q", rc.MotionGraphicsTemplateName)
	}
	if got := rc.MotionGraphicsTemplateControllerCount(); got != 1 {
		t.Fatalf("re-parse ControllerCount = %d, want 1", got)
	}
	g := rc.EssentialGraphicsControllers[0]
	if g.Name != "Exposed Slider" || g.Type != aep.EGSlider || g.UUID != ctrl.UUID {
		t.Errorf("re-parse controller = %+v, want name/type/uuid preserved (%s)", g, ctrl.UUID)
	}
}

// Fixture-mutate flow: append a controller to an AE-native panel that already
// has three, and verify count + order + existing controllers survive.
func TestAddEssentialProperty_FixtureAppend(t *testing.T) {
	proj, err := aep.Open("../../test_data/eg_multiple_controllers.aep")
	if err != nil {
		t.Skipf("fixture not present: %v", err)
	}
	comp := proj.CompositionByName("primary")
	if comp == nil {
		t.Fatal("comp 'primary' not found")
	}
	var layer *aep.Layer
	var fx *aep.Effect
	for _, l := range comp.Layers {
		for _, e := range l.Effects {
			if e.MatchName == "ADBE Brightness & Contrast 2" {
				layer, fx = l, e
			}
		}
	}
	if fx == nil {
		t.Fatal("Brightness & Contrast effect not found")
	}

	ctrl, err := aep.AddEssentialProperty(layer, fx, "ADBE Brightness & Contrast 2-0002", "Exposed Contrast")
	if err != nil {
		t.Fatalf("AddEssentialProperty: %v", err)
	}
	if got := comp.MotionGraphicsTemplateControllerCount(); got != 4 {
		t.Fatalf("ControllerCount = %d, want 4", got)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	rc := re.CompositionByName("primary")
	if got := rc.MotionGraphicsTemplateControllerCount(); got != 4 {
		t.Fatalf("re-parse ControllerCount = %d, want 4", got)
	}
	wantNames := []string{"Brightness", "Layer Opacity", "Background Color", "Exposed Contrast"}
	for i, want := range wantNames {
		if got := rc.EssentialGraphicsControllers[i].Name; got != want {
			t.Errorf("controller[%d].Name = %q, want %q", i, got, want)
		}
	}
	if rc.EssentialGraphicsControllers[3].UUID != ctrl.UUID {
		t.Errorf("appended controller UUID mismatch")
	}
}

// Refuse set: unknown param, color without a materialized value, effect not
// on the layer.
func TestAddEssentialProperty_Refusals(t *testing.T) {
	proj, err := aep.Open("../../test_data/eg_multiple_controllers.aep")
	if err != nil {
		t.Skipf("fixture not present: %v", err)
	}
	comp := proj.CompositionByName("primary")
	var layer *aep.Layer
	var bc, fill *aep.Effect
	for _, l := range comp.Layers {
		for _, e := range l.Effects {
			switch e.MatchName {
			case "ADBE Brightness & Contrast 2":
				layer, bc = l, e
			case "ADBE Fill":
				fill = e
			}
		}
	}
	if bc == nil || fill == nil {
		t.Fatal("expected effects not found")
	}
	if _, err := aep.AddEssentialProperty(layer, bc, "ADBE Brightness & Contrast 2-9999", ""); err == nil {
		t.Error("unknown param accepted, want refusal")
	}
	before := comp.MotionGraphicsTemplateControllerCount()
	if got := comp.MotionGraphicsTemplateControllerCount(); got != before {
		t.Errorf("refusal mutated controller count")
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
