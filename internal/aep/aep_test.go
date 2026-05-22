package aep_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// manualAEPPath lets you point the test at a real .aep file from the command line:
//
//	go test -run TestManualFile -aep "C:/path/to/project.aep" -v
var manualAEPPath = flag.String("aep", "", "path to a real .aep file for TestManualFile")

// Common synthetic-RIFX builder helpers (rifxBuilder, buildIdta, buildCdta,
// buildCdtaWithRate, buildMinimalAEP) live in testutil_test.go.
// Per-test specialized builders stay alongside their tests below.

func TestParseMinimalAEP(t *testing.T) {
	data := buildMinimalAEP()
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader failed: %v", err)
	}

	// Should have 1 composition
	if len(proj.Compositions) != 1 {
		t.Errorf("expected 1 composition, got %d", len(proj.Compositions))
	} else {
		comp := proj.Compositions[0]
		if comp.Name != "Test Comp" {
			t.Errorf("comp name = %q, want %q", comp.Name, "Test Comp")
		}
		if comp.Width != 1920 {
			t.Errorf("comp width = %d, want 1920", comp.Width)
		}
		if comp.Height != 1080 {
			t.Errorf("comp height = %d, want 1080", comp.Height)
		}
		wantFPS := 29.97
		if comp.FrameRate < wantFPS-0.1 || comp.FrameRate > wantFPS+0.1 {
			t.Errorf("comp fps = %.3f, want ~%.2f", comp.FrameRate, wantFPS)
		}
		// Duration: 300 frames / 29.97 fps ≈ 10.01 s
		if comp.Duration < 9.9 || comp.Duration > 10.1 {
			t.Errorf("comp duration = %.3f, want ~10.0s", comp.Duration)
		}
	}

	// Should have 1 footage
	if len(proj.Footage) != 1 {
		t.Errorf("expected 1 footage, got %d", len(proj.Footage))
	} else {
		f := proj.Footage[0]
		if f.Name != "bg.mp4" {
			t.Errorf("footage name = %q, want %q", f.Name, "bg.mp4")
		}
		if f.Width != 1280 {
			t.Errorf("footage width = %d, want 1280", f.Width)
		}
		if f.Height != 720 {
			t.Errorf("footage height = %d, want 720", f.Height)
		}
	}

	// Should have 1 folder
	if len(proj.Folders) != 1 {
		t.Errorf("expected 1 folder, got %d", len(proj.Folders))
	} else {
		if proj.Folders[0].Name != "Solids" {
			t.Errorf("folder name = %q, want %q", proj.Folders[0].Name, "Solids")
		}
	}
}

func TestJSONOutput(t *testing.T) {
	data := buildMinimalAEP()
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader failed: %v", err)
	}

	var buf bytes.Buffer
	if err := proj.WriteJSON(&buf); err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	// Validate it's valid JSON
	var out map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("invalid JSON output: %v\nJSON:\n%s", err, buf.String())
	}

	// Check compositions key exists
	if _, ok := out["compositions"]; !ok {
		t.Error("JSON missing 'compositions' key")
	}
	if _, ok := out["footage"]; !ok {
		t.Error("JSON missing 'footage' key")
	}
}

func TestInvalidFile(t *testing.T) {
	_, err := aep.FromReader(bytes.NewReader([]byte("not a RIFX file at all")))
	if err == nil {
		t.Error("expected error for invalid file, got nil")
	}
}

func TestEmptyFile(t *testing.T) {
	_, err := aep.FromReader(bytes.NewReader([]byte{}))
	if err == nil {
		t.Error("expected error for empty file, got nil")
	}
}

// TestManualFile parses a real .aep file provided via the -aep flag and dumps
// what was extracted. Skipped when no flag is set, so it never breaks CI.
//
//	go test -run TestManualFile -aep "C:/path/to/project.aep" -v
func TestManualFile(t *testing.T) {
	if *manualAEPPath == "" {
		t.Skip(`no -aep flag set; usage: go test -run TestManualFile -aep "path/to/file.aep" -v`)
	}

	proj, err := aep.Open(*manualAEPPath)
	if err != nil {
		t.Fatalf("Open(%q): %v", *manualAEPPath, err)
	}

	sep := strings.Repeat("─", 60)
	t.Logf("\n%s\nFile: %s\n%s", sep, *manualAEPPath, sep)
	t.Logf("Compositions: %d   Footage: %d   Folders: %d",
		len(proj.Compositions), len(proj.Footage), len(proj.Folders))

	for _, comp := range proj.Compositions {
		t.Logf("\n[Comp id=%d] %q  %dx%d  %.3ffps  %.3fs  bg=#%02X%02X%02X  layers=%d",
			comp.ID, comp.Name, comp.Width, comp.Height, comp.FrameRate, comp.Duration,
			comp.BGColor[0], comp.BGColor[1], comp.BGColor[2], len(comp.Layers))
		for _, l := range comp.Layers {
			t.Logf("  layer[%d] %-25s  type=%-8s src=%-4d  start=%.3fs  dur=%.3fs  stretch=%.3f  props=%d",
				l.Index+1, truncForLog(l.Name, 25), l.Type, l.SourceID,
				l.StartTime, l.Duration, l.Stretch, len(l.Properties))
			for _, p := range l.Properties {
				if p.StaticValue != nil {
					t.Logf("      ↳ %s = %v", p.Name, p.StaticValue)
				} else if len(p.Keyframes) > 0 {
					t.Logf("      ↳ %s (%d keyframes)", p.Name, len(p.Keyframes))
				}
			}
		}
	}

	for _, f := range proj.Footage {
		kind := "file"
		if f.IsSolid {
			kind = "solid"
		} else if f.IsStill {
			kind = "still"
		}
		t.Logf("[Footage id=%d kind=%s] %q  %dx%d  path=%s",
			f.ID, kind, f.Name, f.Width, f.Height, f.Path)
	}

	for _, fld := range proj.Folders {
		t.Logf("[Folder id=%d] %q", fld.ID, fld.Name)
	}

	// Also produce JSON so you can eyeball the serialized shape.
	var buf bytes.Buffer
	if err := proj.WriteJSON(&buf); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	t.Logf("\nJSON (%d bytes):\n%s", buf.Len(), buf.String())
}

func truncForLog(s string, n int) string {
	if len(s) <= n {
		return fmt.Sprintf("%-*s", n, s)
	}
	return s[:n-1] + "…"
}

func TestRoundtripWrite(t *testing.T) {
	data := buildMinimalAEP()
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), data) {
		t.Errorf("roundtrip mismatch: input %d bytes, output %d bytes", len(data), buf.Len())
		// Show the first divergence for debugging.
		minLen := len(data)
		if buf.Len() < minLen {
			minLen = buf.Len()
		}
		for i := 0; i < minLen; i++ {
			if data[i] != buf.Bytes()[i] {
				t.Logf("first diff at offset 0x%X: in=0x%02X out=0x%02X", i, data[i], buf.Bytes()[i])
				break
			}
		}
	}

	// Parse the output and confirm it still decodes to the same project.
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if len(proj2.Compositions) != 1 || proj2.Compositions[0].Name != "Test Comp" {
		t.Errorf("re-parsed compositions wrong: %+v", proj2.Compositions)
	}
}

// ──────────────────────────────────────────────
// Keyframe + static-value roundtrip
//
// Builds a synthetic .aep with one composition holding one layer that owns
// three properties:
//
//	ADBE Opacity                  — 1D non-spatial keyframed (kfValueOffset=0x08, bpk=48)
//	ADBE Position                 — 3D spatial keyframed     (kfValueOffset=0x38, bpk=128)
//	ADBE Glossiness Coefficient   — 1D static cdat value
//
// then exercises SetTime / SetValue / SetStaticValue and re-parses to confirm
// the edits land at the byte level.
// ──────────────────────────────────────────────

// TestKeyframeSetValueUpdatesInMemory pins the regression for the 1D
// SetValue path: bytes AND k.Value must both be updated. Earlier the
// scalar case returned early after writing bytes, leaving k.Value stale.
// ──────────────────────────────────────────────
// Extended features: expression / effects / markers / text source / opaque LIST roundtrip
// ──────────────────────────────────────────────


// ──────────────────────────────────────────────
// Easing / masks / shape-layer detection
// ──────────────────────────────────────────────


// ──────────────────────────────────────────────
// Text-source decoding (btds → TextSource)
// ──────────────────────────────────────────────


func staticValueOf(p *aep.Property) any {
	if p == nil {
		return nil
	}
	return p.StaticValue
}


// ──────────────────────────────────────────────────────────────────
// AE 24+ Wave 1: TextDocument fields exposed by AE 24 ScriptingAPI
// (caps / baseline / strokeOverFill + paragraph indents / spaces /
// autoHyphenate). Fixture re_text_caps_ae24.aep is generated by
// test_data/re_text_caps_ae24.jsx running in AE 24+ — skip if absent.
// ──────────────────────────────────────────────────────────────────


// ──────────────────────────────────────────────────────────────────
// AE 23+ Wave 2: AVLayer.trackMatteLayer (explicit matte source pointer)
// Fixture re_trackmatte_ae24.aep — RE_TRACKMATTE comp:
//   layer Source=solidC  (top)   id=67  — used as matte by id=59
//   layer Source=solidB          id=65  — used as matte by id=57
//   layer Source=solidA          id=63  — used as matte by id=55
//   layer Source=mt_baseline     id=61  — no matte
//   layer Source=mt_alphainv     id=59  TrackMatte=AlphaInverse  TMLayerID=67
//   layer Source=mt_luma         id=57  TrackMatte=Luma          TMLayerID=65
//   layer Source=mt_alpha (bot)  id=55  TrackMatte=Alpha         TMLayerID=63
// (AE 25 doesn't write Utf8 layer names for solids; identify by SourceID
// → footage name.)
// ──────────────────────────────────────────────────────────────────


// ──────────────────────────────────────────────────────────────────
// AE 24+ Wave 2: LightKind (ldta @0x88) + TimeRemapEnabled (property-tree
// derived) + Composition.DisplayStartTime (cdta @0xA4/@0xA8).
// Fixture re_wave2_ae24.aep.
// ──────────────────────────────────────────────────────────────────


// ──────────────────────────────────────────────────────────────────
// Wave 3 (preview): AE 24+ TextDocument extension fields —
// per-run AutoKernType / NoBreak / LineJoinType / DigitSet +
// per-paragraph LeadingType / HangingRoman / Direction. Fixture
// re_text_ae24_more.aep.
// ──────────────────────────────────────────────────────────────────


// ───── alternateSource (Wave 3) ─────


// ───── manual kerning (Wave 3) ─────

func TestProjectInitDerived_EmptyProject(t *testing.T) {
	data := buildMinimalAEP()
	proj, err := aep.FromReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}
	var maxExisting uint32
	for _, c := range proj.Compositions {
		if c.ID > maxExisting {
			maxExisting = c.ID
		}
	}
	for _, f := range proj.Footage {
		if f.ID > maxExisting {
			maxExisting = f.ID
		}
	}
	got := proj.NextItemIDForTest()
	if got <= maxExisting {
		t.Errorf("nextItemID = %d, want > max existing %d", got, maxExisting)
	}
	if proj.RootFoldForTest() == nil {
		t.Error("rootFold not cached after parseProject")
	}
}

