package aep_test

import (
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

func TestImportPlaceholder_Basic(t *testing.T) {
	proj := aep.NewProject()

	// Create a placeholder
	footage, err := proj.ImportPlaceholder("Test Placeholder", 1920, 1080, 30.0, 5.0)
	if err != nil {
		t.Fatalf("ImportPlaceholder failed: %v", err)
	}

	// Verify the footage item
	if footage == nil {
		t.Fatal("footage is nil")
	}
	if footage.Name != "Test Placeholder" {
		t.Errorf("Name = %q, want %q", footage.Name, "Test Placeholder")
	}
	if footage.Width != 1920 {
		t.Errorf("Width = %d, want 1920", footage.Width)
	}
	if footage.Height != 1080 {
		t.Errorf("Height = %d, want 1080", footage.Height)
	}
	if !footage.IsPlaceholder {
		t.Error("IsPlaceholder = false, want true")
	}
	if footage.IsSolid {
		t.Error("IsSolid = true, want false")
	}
	if footage.AssetType() != "placeholder" {
		t.Errorf("AssetType = %q, want %q", footage.AssetType(), "placeholder")
	}

	// Verify it's in the project
	if len(proj.Footage) != 1 {
		t.Errorf("proj.Footage length = %d, want 1", len(proj.Footage))
	}
	if proj.Footage[0] != footage {
		t.Error("footage not in proj.Footage")
	}
}

func TestImportPlaceholder_NameNormalization(t *testing.T) {
	proj := aep.NewProject()

	// Empty name should default to "Placeholder"
	footage, err := proj.ImportPlaceholder("", 640, 480, 24.0, 10.0)
	if err != nil {
		t.Fatalf("ImportPlaceholder with empty name failed: %v", err)
	}
	if footage.Name != "Placeholder" {
		t.Errorf("Name with empty input = %q, want %q", footage.Name, "Placeholder")
	}
}

func TestImportPlaceholder_Validation(t *testing.T) {
	proj := aep.NewProject()

	cases := []struct {
		name    string
		width   int
		height  int
		fps     float64
		dur     float64
		wantErr string
	}{
		{"width too small", 3, 480, 24.0, 5.0, "width"},
		{"width too large", 30001, 480, 24.0, 5.0, "width"},
		{"height too small", 640, 3, 24.0, 5.0, "height"},
		{"height too large", 640, 30001, 24.0, 5.0, "height"},
		{"frameRate too low", 640, 480, 0.5, 5.0, "frameRate"},
		{"frameRate too high", 640, 480, 100.0, 5.0, "frameRate"},
		{"duration zero", 640, 480, 24.0, 0.0, "duration"},
		{"duration too large", 640, 480, 24.0, 20000.0, "duration"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := proj.ImportPlaceholder("Test", c.width, c.height, c.fps, c.dur)
			if err == nil {
				t.Fatalf("expected error for %s, got nil", c.name)
			}
			if !contains(err.Error(), c.wantErr) {
				t.Errorf("error = %q, want substring %q", err.Error(), c.wantErr)
			}
		})
	}
}

func TestImportPlaceholder_MultiplePlaceholders(t *testing.T) {
	proj := aep.NewProject()

	// Create multiple placeholders
	p1, err := proj.ImportPlaceholder("Placeholder 1", 1920, 1080, 30.0, 5.0)
	if err != nil {
		t.Fatalf("ImportPlaceholder 1 failed: %v", err)
	}
	p2, err := proj.ImportPlaceholder("Placeholder 2", 1280, 720, 24.0, 10.0)
	if err != nil {
		t.Fatalf("ImportPlaceholder 2 failed: %v", err)
	}

	// Verify they have different IDs
	if p1.ID == p2.ID {
		t.Errorf("Both placeholders have same ID: %d", p1.ID)
	}

	// Verify both are in the project
	if len(proj.Footage) != 2 {
		t.Errorf("proj.Footage length = %d, want 2", len(proj.Footage))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
