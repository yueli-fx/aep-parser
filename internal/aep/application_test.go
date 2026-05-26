package aep_test

import (
	"os"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestApplication_Parse(t *testing.T) {
	app, err := aep.Parse("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}
	if app.Project == nil {
		t.Fatal("Application.Project == nil")
	}
	if len(app.Project.Compositions) == 0 {
		t.Error("Project.Compositions empty after Parse")
	}
}

func TestApplication_ParseReader(t *testing.T) {
	f, err := os.Open("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}
	defer f.Close()
	app, err := aep.ParseReader(f)
	if err != nil {
		t.Fatalf("ParseReader: %v", err)
	}
	if app.Project == nil {
		t.Fatal("Application.Project == nil")
	}
}

func TestApplication_Version(t *testing.T) {
	app, err := aep.Parse("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}
	v := app.Version()
	if v == "" {
		t.Skip("head chunk absent in this fixture")
	}
	// re_cameralight.aep was written by AE 2020 (per re_cameralight.jsx
	// AE 2020 ScriptingAPI convention). Major should be 17 (AE 2020).
	t.Logf("AE version: %s", v)
	// Format: "major.minorxbuild"
	if v == "0.0x0" {
		t.Errorf("version %q looks unparsed", v)
	}
	// Verify format contains exactly one '.' and one 'x'.
	dots, xs := 0, 0
	for _, c := range v {
		switch c {
		case '.':
			dots++
		case 'x':
			xs++
		default:
			if c < '0' || c > '9' {
				t.Errorf("non-numeric char %q in version %q", c, v)
				return
			}
		}
	}
	if dots != 1 || xs != 1 {
		t.Errorf("version %q expected one '.' and one 'x', got dots=%d xs=%d", v, dots, xs)
	}
}

func TestApplication_VersionNoProject(t *testing.T) {
	a := &aep.Application{}
	if v := a.Version(); v != "" {
		t.Errorf("Version on empty Application = %q, want \"\"", v)
	}
}
