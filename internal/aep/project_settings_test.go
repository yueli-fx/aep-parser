package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// TestProjectSettings_Read verifies all reader helpers against
// re_cameralight.aep with values confirmed via tmp_debug dumps:
//   - Revision         = 4733 (head @0x12 uint16 BE)
//   - LinearBlending   = false (no lnrb chunk in fixture)
//   - LinearizeWorkingSpace = false (no lnrp)
//   - CompensateForSceneReferredProfiles = true (acer[0] = 0x01)
//   - AudioSampleRate  = 48000.0
//   - WorkingGamma     = 2.4 (dwga[0] = 0x01)
//   - GpuAccelType     = "7ee0ab59-822d-44cc-ac10-16279d041016"
//   - ExpressionEngine = "javascript-1.0"
func TestProjectSettings_Read(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}
	if got, want := proj.Revision(), uint16(4733); got != want {
		t.Errorf("Revision = %d, want %d", got, want)
	}
	if proj.LinearBlending() {
		t.Errorf("LinearBlending = true; want false (no lnrb)")
	}
	if proj.LinearizeWorkingSpace() {
		t.Errorf("LinearizeWorkingSpace = true; want false (no lnrp)")
	}
	if !proj.CompensateForSceneReferredProfiles() {
		t.Errorf("CompensateForSceneReferredProfiles = false; want true (acer[0]=1)")
	}
	if got, want := proj.AudioSampleRate(), 48000.0; got != want {
		t.Errorf("AudioSampleRate = %v, want %v", got, want)
	}
	if got, want := proj.WorkingGamma(), 2.4; got != want {
		t.Errorf("WorkingGamma = %v, want %v", got, want)
	}
	if got := proj.GpuAccelType(); got != "7ee0ab59-822d-44cc-ac10-16279d041016" {
		t.Errorf("GpuAccelType = %q", got)
	}
	if got, want := proj.ExpressionEngine(), "javascript-1.0"; got != want {
		t.Errorf("ExpressionEngine = %q, want %q", got, want)
	}
}

// TestProjectSettings_RoundtripWrite changes every setting, writes the
// project back to bytes, re-parses, and verifies the change took.
// Excludes Revision (R only) and the two toggle flag chunks (covered
// in a separate test).
func TestProjectSettings_RoundtripWrite(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}

	// Toggle each field to a non-default.
	if err := proj.SetCompensateForSceneReferredProfiles(false); err != nil {
		t.Fatalf("SetCompensateForSceneReferredProfiles: %v", err)
	}
	if err := proj.SetAudioSampleRate(44100); err != nil {
		t.Fatalf("SetAudioSampleRate(44100): %v", err)
	}
	if err := proj.SetWorkingGamma(2.2); err != nil {
		t.Fatalf("SetWorkingGamma(2.2): %v", err)
	}
	if err := proj.SetGpuAccelType("abcdef12-3456-7890-aaaa-bbbbccccdddd"); err != nil {
		t.Fatalf("SetGpuAccelType: %v", err)
	}
	if err := proj.SetExpressionEngine("extendscript"); err != nil {
		t.Fatalf("SetExpressionEngine: %v", err)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}

	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader after write: %v", err)
	}
	if proj2.CompensateForSceneReferredProfiles() {
		t.Errorf("after roundtrip: CompensateForSceneReferredProfiles = true; want false")
	}
	if got := proj2.AudioSampleRate(); got != 44100 {
		t.Errorf("after roundtrip: AudioSampleRate = %v; want 44100", got)
	}
	if got := proj2.WorkingGamma(); got != 2.2 {
		t.Errorf("after roundtrip: WorkingGamma = %v; want 2.2", got)
	}
	if got := proj2.GpuAccelType(); got != "abcdef12-3456-7890-aaaa-bbbbccccdddd" {
		t.Errorf("after roundtrip: GpuAccelType = %q", got)
	}
	if got := proj2.ExpressionEngine(); got != "extendscript" {
		t.Errorf("after roundtrip: ExpressionEngine = %q; want extendscript", got)
	}
}

// TestProjectSettings_ToggleFlagChunks adds + removes lnrb / lnrp,
// roundtrips, and verifies the chunk presence change persists.
func TestProjectSettings_ToggleFlagChunks(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}
	if proj.LinearBlending() {
		t.Skip("fixture already has lnrb; need clean baseline")
	}

	if err := proj.SetLinearBlending(true); err != nil {
		t.Fatalf("SetLinearBlending(true): %v", err)
	}
	if !proj.LinearBlending() {
		t.Errorf("after SetLinearBlending(true): still false")
	}
	if err := proj.SetLinearizeWorkingSpace(true); err != nil {
		t.Fatalf("SetLinearizeWorkingSpace(true): %v", err)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	if !proj2.LinearBlending() {
		t.Errorf("after roundtrip: LinearBlending = false; want true")
	}
	if !proj2.LinearizeWorkingSpace() {
		t.Errorf("after roundtrip: LinearizeWorkingSpace = false; want true")
	}

	// Toggle off and roundtrip again.
	if err := proj2.SetLinearBlending(false); err != nil {
		t.Fatalf("SetLinearBlending(false): %v", err)
	}
	var buf2 bytes.Buffer
	if err := proj2.WriteAEP(&buf2); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	proj3, err := aep.FromReader(bytes.NewReader(buf2.Bytes()))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	if proj3.LinearBlending() {
		t.Errorf("after toggle-off roundtrip: LinearBlending = true")
	}
	if !proj3.LinearizeWorkingSpace() {
		t.Errorf("after toggle-off roundtrip: LinearizeWorkingSpace = false (should still be set)")
	}
}

func TestProjectSettings_Validation(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}
	if err := proj.SetAudioSampleRate(12345); err == nil {
		t.Error("SetAudioSampleRate(12345): expected error (not in valid set)")
	}
	if err := proj.SetWorkingGamma(1.0); err == nil {
		t.Error("SetWorkingGamma(1.0): expected error (only 2.2 / 2.4)")
	}
	if err := proj.SetExpressionEngine("python"); err == nil {
		t.Error("SetExpressionEngine(python): expected error")
	}
}

func TestProjectSettings_StandaloneProject(t *testing.T) {
	p := &aep.Project{}
	if p.Revision() != 0 {
		t.Error("standalone Revision != 0")
	}
	if p.LinearBlending() {
		t.Error("standalone LinearBlending != false")
	}
	if p.CompensateForSceneReferredProfiles() {
		t.Error("standalone CompensateForSceneReferredProfiles != false")
	}
	if p.AudioSampleRate() != 0 {
		t.Error("standalone AudioSampleRate != 0")
	}
	if p.WorkingGamma() != 2.2 {
		t.Errorf("standalone WorkingGamma = %v, want 2.2 default", p.WorkingGamma())
	}
	if p.GpuAccelType() != "" {
		t.Error("standalone GpuAccelType != \"\"")
	}
	if got := p.ExpressionEngine(); got != "extendscript" {
		t.Errorf("standalone ExpressionEngine = %q, want extendscript default", got)
	}
	// Setters refuse on standalone project (no chunks to mutate).
	if err := p.SetLinearBlending(true); err == nil {
		t.Error("SetLinearBlending on standalone: expected error")
	}
	if err := p.SetAudioSampleRate(44100); err == nil {
		t.Error("SetAudioSampleRate on standalone: expected error")
	}
}

// TestProjectSettings_NnhdRead verifies nnhd field readers against
// re_cameralight.aep (AE 2020 project).
func TestProjectSettings_NnhdRead(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}

	// Verify nnhd fields have reasonable values (AE 2020 project)
	// Note: Actual values depend on the fixture; we just verify they're valid
	if got := proj.FeetFramesFilmType(); got != aep.FeetFramesFilmTypeMM35 && got != aep.FeetFramesFilmTypeMM16 {
		t.Errorf("FeetFramesFilmType = %v, want MM35 or MM16", got)
	}
	// FootageTimecodeDisplayStartType can be either value
	if got := proj.FootageTimecodeDisplayStartType(); got != aep.FootageTimecodeDisplayStartTypeStart0 && got != aep.FootageTimecodeDisplayStartTypeUseSourceMedia {
		t.Errorf("FootageTimecodeDisplayStartType = %v, want Start0 or UseSourceMedia", got)
	}
	if got := proj.TimecodeDefaultBase(); got < 1 || got > 999 {
		t.Errorf("TimecodeDefaultBase = %d, want 1-999", got)
	}
	if got := proj.FramesCountType(); got > 2 {
		t.Errorf("FramesCountType = %d, want 0-2", got)
	}
	if got := proj.DisplayStartFrame(); got != 0 && got != 1 {
		t.Errorf("DisplayStartFrame = %d, want 0 or 1", got)
	}
	if got := proj.TimeDisplayType(); got != aep.TimeDisplayTypeTimecode && got != aep.TimeDisplayTypeFrames {
		t.Errorf("TimeDisplayType = %v, want Timecode or Frames", got)
	}
}

// TestProjectSettings_NnhdRoundtrip verifies nnhd field writers roundtrip correctly.
func TestProjectSettings_NnhdRoundtrip(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}

	// Change all nnhd fields
	if err := proj.SetFeetFramesFilmType(aep.FeetFramesFilmTypeMM16); err != nil {
		t.Fatalf("SetFeetFramesFilmType: %v", err)
	}
	if err := proj.SetFootageTimecodeDisplayStartType(aep.FootageTimecodeDisplayStartTypeUseSourceMedia); err != nil {
		t.Fatalf("SetFootageTimecodeDisplayStartType: %v", err)
	}
	if err := proj.SetTimecodeDefaultBase(123); err != nil {
		t.Fatalf("SetTimecodeDefaultBase: %v", err)
	}
	if err := proj.SetFramesCountType(aep.FramesCountTypeStart1); err != nil {
		t.Fatalf("SetFramesCountType: %v", err)
	}
	if err := proj.SetDisplayStartFrame(1); err != nil {
		t.Fatalf("SetDisplayStartFrame: %v", err)
	}
	if err := proj.SetFramesUseFeetFrames(true); err != nil {
		t.Fatalf("SetFramesUseFeetFrames: %v", err)
	}
	if err := proj.SetTimeDisplayType(aep.TimeDisplayTypeFrames); err != nil {
		t.Fatalf("SetTimeDisplayType: %v", err)
	}
	if err := proj.SetTransparencyGridThumbnails(true); err != nil {
		t.Fatalf("SetTransparencyGridThumbnails: %v", err)
	}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}

	proj2, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("FromReader after write: %v", err)
	}

	// Verify roundtrip values
	if got := proj2.FeetFramesFilmType(); got != aep.FeetFramesFilmTypeMM16 {
		t.Errorf("after roundtrip: FeetFramesFilmType = %v, want MM16", got)
	}
	if got := proj2.FootageTimecodeDisplayStartType(); got != aep.FootageTimecodeDisplayStartTypeUseSourceMedia {
		t.Errorf("after roundtrip: FootageTimecodeDisplayStartType = %v, want UseSourceMedia", got)
	}
	if got := proj2.TimecodeDefaultBase(); got != 123 {
		t.Errorf("after roundtrip: TimecodeDefaultBase = %d, want 123", got)
	}
	if got := proj2.FramesCountType(); got != aep.FramesCountTypeStart1 {
		t.Errorf("after roundtrip: FramesCountType = %v, want Start1", got)
	}
	if got := proj2.DisplayStartFrame(); got != 1 {
		t.Errorf("after roundtrip: DisplayStartFrame = %d, want 1", got)
	}
	if got := proj2.FramesUseFeetFrames(); got != true {
		t.Errorf("after roundtrip: FramesUseFeetFrames = %v, want true", got)
	}
	if got := proj2.TimeDisplayType(); got != aep.TimeDisplayTypeFrames {
		t.Errorf("after roundtrip: TimeDisplayType = %v, want Frames", got)
	}
	if got := proj2.TransparencyGridThumbnails(); got != true {
		t.Errorf("after roundtrip: TransparencyGridThumbnails = %v, want true", got)
	}
}

// TestProjectSettings_NnhdValidation verifies nnhd field validation.
func TestProjectSettings_NnhdValidation(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}

	// TimecodeDefaultBase must be 1-999
	if err := proj.SetTimecodeDefaultBase(0); err == nil {
		t.Error("SetTimecodeDefaultBase(0): expected error")
	}
	if err := proj.SetTimecodeDefaultBase(1000); err == nil {
		t.Error("SetTimecodeDefaultBase(1000): expected error")
	}

	// DisplayStartFrame must be 0 or 1
	if err := proj.SetDisplayStartFrame(2); err == nil {
		t.Error("SetDisplayStartFrame(2): expected error")
	}
	if err := proj.SetDisplayStartFrame(-1); err == nil {
		t.Error("SetDisplayStartFrame(-1): expected error")
	}
}

// TestProjectSettings_NnhdStandalone verifies nnhd readers return defaults
// when nnhd chunk is absent (standalone project).
func TestProjectSettings_NnhdStandalone(t *testing.T) {
	p := &aep.Project{}
	if got := p.FeetFramesFilmType(); got != aep.FeetFramesFilmTypeMM35 {
		t.Errorf("standalone FeetFramesFilmType = %v, want MM35", got)
	}
	if got := p.FootageTimecodeDisplayStartType(); got != aep.FootageTimecodeDisplayStartTypeStart0 {
		t.Errorf("standalone FootageTimecodeDisplayStartType = %v, want Start0", got)
	}
	if got := p.TimecodeDefaultBase(); got != 0 {
		t.Errorf("standalone TimecodeDefaultBase = %d, want 0", got)
	}
	if got := p.FramesCountType(); got != aep.FramesCountTypeStart0 {
		t.Errorf("standalone FramesCountType = %v, want Start0", got)
	}
	if got := p.DisplayStartFrame(); got != 0 {
		t.Errorf("standalone DisplayStartFrame = %d, want 0", got)
	}
	if got := p.FramesUseFeetFrames(); got != false {
		t.Errorf("standalone FramesUseFeetFrames = %v, want false", got)
	}
	if got := p.TimeDisplayType(); got != aep.TimeDisplayTypeTimecode {
		t.Errorf("standalone TimeDisplayType = %v, want Timecode", got)
	}
	if got := p.TransparencyGridThumbnails(); got != false {
		t.Errorf("standalone TransparencyGridThumbnails = %v, want false", got)
	}

	// Setters refuse on standalone project (no nnhd chunk to mutate).
	if err := p.SetFeetFramesFilmType(aep.FeetFramesFilmTypeMM16); err == nil {
		t.Error("SetFeetFramesFilmType on standalone: expected error")
	}
	if err := p.SetTimecodeDefaultBase(123); err == nil {
		t.Error("SetTimecodeDefaultBase on standalone: expected error")
	}
}
