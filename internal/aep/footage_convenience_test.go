package aep_test

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

func TestFootageConvenience_ReCameraLight(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}
	if len(proj.Footage) == 0 {
		t.Fatal("no footage in re_cameralight.aep")
	}

	// First footage: a JPG file → file asset_type, has path, no audio
	// (image), not solid/placeholder, footage missing depends on local
	// disk so we don't assert it.
	var file, solid *aep.Footage
	for _, f := range proj.Footage {
		switch {
		case file == nil && !f.IsSolid && !f.IsPlaceholder && f.Path != "":
			file = f
		case solid == nil && f.IsSolid:
			solid = f
		}
		if file != nil && solid != nil {
			break
		}
	}

	if file != nil {
		if got := file.AssetType(); got != "file" {
			t.Errorf("file footage AssetType = %q, want \"file\"", got)
		}
		if file.File() != file.Path {
			t.Errorf("file footage File() = %q, want Path %q", file.File(), file.Path)
		}
		// Image (JPG) → no audio.
		if file.HasAudio() {
			t.Errorf("JPG footage HasAudio = true; want false")
		}
		// Width/Height should be non-zero (bug fix: real AE sspc layout).
		if file.Width == 0 || file.Height == 0 {
			t.Errorf("file footage W/H = %d/%d, want non-zero", file.Width, file.Height)
		}
	} else {
		t.Log("no file footage to test")
	}

	if solid != nil {
		if got := solid.AssetType(); got != "solid" {
			t.Errorf("solid AssetType = %q, want \"solid\"", got)
		}
		if solid.File() != "" {
			t.Errorf("solid File() = %q, want empty", solid.File())
		}
		if solid.HasAudio() {
			t.Errorf("solid HasAudio = true; want false")
		}
		if solid.FootageMissing() {
			t.Errorf("solid FootageMissing = true; want false (solids never missing)")
		}
	} else {
		t.Log("no solid footage to test")
	}
}

func TestFootageConvenience_AssetType_Placeholder(t *testing.T) {
	f := &aep.Footage{IsPlaceholder: true, Name: "Missing"}
	if got := f.AssetType(); got != "placeholder" {
		t.Errorf("AssetType = %q, want \"placeholder\"", got)
	}
	if f.HasAudio() {
		t.Errorf("placeholder HasAudio = true; want false")
	}
	if f.FootageMissing() {
		t.Errorf("placeholder FootageMissing = true; want false (handled by AssetType)")
	}
}

func TestFootageConvenience_NoSspc(t *testing.T) {
	// Footage built without parser → no sspc; all sspc-derived helpers
	// return 0/false cleanly.
	f := &aep.Footage{Name: "Bare"}
	if f.StartFrame() != 0 {
		t.Errorf("bare StartFrame = %d, want 0", f.StartFrame())
	}
	if f.EndFrame() != 0 {
		t.Errorf("bare EndFrame = %d, want 0", f.EndFrame())
	}
	if f.HasAudio() {
		t.Error("bare HasAudio = true; want false")
	}
	if f.FootageMissing() {
		t.Error("bare FootageMissing = true; want false")
	}
	if got := f.AssetType(); got != "file" {
		t.Errorf("bare AssetType = %q, want \"file\" (default)", got)
	}
}
