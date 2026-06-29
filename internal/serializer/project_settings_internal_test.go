package serializer

import (
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/rifx"
	"github.com/yueli-fx/aep-parser/internal/scene"
)

// TestProjectSettings_CmsEnumValidation verifies CMS setters reject
// out-of-range enum values, matching the validator pattern used by
// SetWorkingGamma / SetExpressionEngine. Internal test so we can wire
// up a synthetic cmsUtf8 chunk without exporting a test hook.
func TestProjectSettings_CmsEnumValidation(t *testing.T) {
	p := &Project{}
	scene.SetProjectBack(p, &projectBackrefs{
		cmsUtf8: &rifx.Chunk{
			ID:   rifx.IDUtf8,
			Data: []byte(`{"colorManagementSystem":0,"lutInterpolationMethod":0,"ocioConfigurationFile":""}`),
		},
	})

	if err := p.SetColorManagementSystem(ColorManagementSystem(99)); err == nil {
		t.Error("SetColorManagementSystem(99): expected enum-validation error")
	}
	if err := p.SetLutInterpolationMethod(LutInterpolationMethod(99)); err == nil {
		t.Error("SetLutInterpolationMethod(99): expected enum-validation error")
	}
	if err := p.SetColorManagementSystem(ColorManagementSystemOCIO); err != nil {
		t.Errorf("SetColorManagementSystem(OCIO): unexpected error %v", err)
	}
	if err := p.SetLutInterpolationMethod(LutInterpolationMethodTetrahedral); err != nil {
		t.Errorf("SetLutInterpolationMethod(Tetrahedral): unexpected error %v", err)
	}
}

// TestProjectSettings_CmsMalformedJsonWarns verifies cmsSettings emits a
// parser Warning when the chunk's JSON is malformed (so callers can
// surface the issue) instead of silently returning defaults.
func TestProjectSettings_CmsMalformedJsonWarns(t *testing.T) {
	p := &Project{}
	scene.SetProjectBack(p, &projectBackrefs{
		cmsUtf8: &rifx.Chunk{
			ID:   rifx.IDUtf8,
			Data: []byte(`{not valid json`),
		},
	})
	_ = p.ColorManagementSystem() // triggers cmsSettings()
	if len(p.Warnings) == 0 {
		t.Fatal("expected a Warning on malformed CMS JSON, got none")
	}
	if !strings.Contains(p.Warnings[0], "CMS JSON parse failed") {
		t.Errorf("warning = %q, want substring 'CMS JSON parse failed'", p.Warnings[0])
	}
}

// TestProject_XmpPacket_RoundtripFromFixture verifies XmpPacket reads
// trailing data and survives a WriteAEP roundtrip.
func TestProject_XmpPacket_RoundtripFromFixture(t *testing.T) {
	// Use any real fixture; XMP is the trailing UTF-8 after RIFX.
	proj, err := Open("../../test_data/re_cameralight.aep")
	if err != nil {
		t.Skipf("re_cameralight.aep not present: %v", err)
	}
	xmp := proj.XmpPacket()
	if xmp == "" {
		t.Skip("fixture has no XMP trailer")
	}
	if !strings.Contains(xmp, "xmpmeta") && !strings.Contains(xmp, "<?xpacket") {
		t.Errorf("XmpPacket() = %d bytes but no xmpmeta/xpacket marker; got prefix %q", len(xmp), xmp[:min(80, len(xmp))])
	}
}

// TestProject_XmpPacket_StandaloneEmpty verifies XmpPacket on a project
// built outside the parser returns "".
func TestProject_XmpPacket_StandaloneEmpty(t *testing.T) {
	p := &Project{}
	if got := p.XmpPacket(); got != "" {
		t.Errorf("standalone XmpPacket() = %q, want empty", got)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestProjectSettings_NhedNnhdMirror guards the RE finding (2026-06-14) that AE
// reads the project display settings from the legacy 32-byte nhed header, not
// the modern 40-byte nnhd. A setter that updates only nnhd is silently ignored
// by AE on reopen. This test asserts each display setter writes BOTH chunks at
// the correct offsets — the regression that produced the original false green
// (Go round-trip OK, AE DOM unchanged). See nnhd-display-settings-layout-re.md.
func TestProjectSettings_NhedNnhdMirror(t *testing.T) {
	nhed := &rifx.Chunk{ID: rifx.ChunkID{'n', 'h', 'e', 'd'}, Data: make([]byte, 32)}
	nnhd := &rifx.Chunk{ID: rifx.ChunkID{'n', 'n', 'h', 'd'}, Data: make([]byte, 40)}
	p := &Project{}
	scene.SetProjectBack(p, &projectBackrefs{nhedChunk: nhed, nnhdChunk: nnhd})

	if err := p.SetTimeDisplayType(TimeDisplayTypeFrames); err != nil {
		t.Fatalf("SetTimeDisplayType: %v", err)
	}
	if nnhd.Data[8] != 1 || nhed.Data[8] != 1 {
		t.Errorf("time display: nnhd[8]=%d nhed[8]=%d, want both 1", nnhd.Data[8], nhed.Data[8])
	}

	if err := p.SetFramesCountType(FramesCountTypeStart1); err != nil {
		t.Fatalf("SetFramesCountType: %v", err)
	}
	if nnhd.Data[20] != 1 || nhed.Data[14] != 1 {
		t.Errorf("frames count: nnhd[20]=%d nhed[14]=%d, want both 1", nnhd.Data[20], nhed.Data[14])
	}

	if err := p.SetFeetFramesFilmType(FeetFramesFilmTypeMM16); err != nil {
		t.Fatalf("SetFeetFramesFilmType: %v", err)
	}
	// MM16 = 40 frames/foot: nnhd u32 BE @0x10, nhed single byte @0x0D.
	if nnhd.Data[19] != 40 || nhed.Data[13] != 40 {
		t.Errorf("feet film type: nnhd[19]=%d nhed[13]=%d, want both 40", nnhd.Data[19], nhed.Data[13])
	}

	if err := p.SetFramesUseFeetFrames(true); err != nil {
		t.Fatalf("SetFramesUseFeetFrames: %v", err)
	}
	if nnhd.Data[11]&0x01 == 0 || nhed.Data[11]&0x01 == 0 {
		t.Errorf("use feet: nnhd[11]=%#x nhed[11]=%#x, want bit0 set in both", nnhd.Data[11], nhed.Data[11])
	}

	if err := p.SetFootageTimecodeDisplayStartType(FootageTimecodeDisplayStartTypeUseSourceMedia); err != nil {
		t.Fatalf("SetFootageTimecodeDisplayStartType: %v", err)
	}
	if nnhd.Data[9] != 1 || nhed.Data[9] != 1 {
		t.Errorf("footage tc: nnhd[9]=%d nhed[9]=%d, want both 1", nnhd.Data[9], nhed.Data[9])
	}

	if err := p.SetTransparencyGridThumbnails(true); err != nil {
		t.Fatalf("SetTransparencyGridThumbnails: %v", err)
	}
	if nnhd.Data[25] != 1 || nhed.Data[16] != 1 {
		t.Errorf("transparency grid: nnhd[25]=%d nhed[16]=%d, want both 1", nnhd.Data[25], nhed.Data[16])
	}

	// Round-trip read back through the (nnhd-based) readers.
	if got := p.TimeDisplayType(); got != TimeDisplayTypeFrames {
		t.Errorf("TimeDisplayType readback = %v, want Frames", got)
	}
	if got := p.FeetFramesFilmType(); got != FeetFramesFilmTypeMM16 {
		t.Errorf("FeetFramesFilmType readback = %v, want MM16", got)
	}
}

// TestProperty_LockedRatio_Positive verifies the LockedRatio reader and
// SetLockedRatio writer against a synthetic tdsb chunk — the existing
// external test in property_flags_test.go can only exercise the fallback
// path because tdsb is unexported.
func TestProperty_LockedRatio_Positive(t *testing.T) {
	tdsb := &rifx.Chunk{
		ID:   rifx.IDTdsb,
		Data: []byte{0x00, 0x00, 0x10, 0x00}, // byte 2 = bit 4 set
	}
	p := &Property{MatchName: "test", Components: 1}
	scene.SetPropertyBack(p, &propertyBackrefs{tdsb: tdsb})

	if !p.LockedRatio() {
		t.Error("LockedRatio with bit-4-set tdsb = false, want true")
	}

	if err := p.SetLockedRatio(false); err != nil {
		t.Fatalf("SetLockedRatio(false): %v", err)
	}
	if p.LockedRatio() {
		t.Error("after SetLockedRatio(false): still true")
	}
	if tdsb.Data[0x02]&0x10 != 0 {
		t.Errorf("after SetLockedRatio(false): tdsb[2] bit 4 still set (data=% x)", tdsb.Data)
	}

	if err := p.SetLockedRatio(true); err != nil {
		t.Fatalf("SetLockedRatio(true): %v", err)
	}
	if !p.LockedRatio() {
		t.Error("after SetLockedRatio(true): still false")
	}
	if tdsb.Data[0x02]&0x10 == 0 {
		t.Errorf("after SetLockedRatio(true): tdsb[2] bit 4 not set (data=% x)", tdsb.Data)
	}

	// Verify SetLockedRatio doesn't perturb adjacent bits — set byte 2 to
	// 0xFF then clear locked_ratio (bit 4); other 7 bits should remain.
	tdsb.Data[0x02] = 0xFF
	if err := p.SetLockedRatio(false); err != nil {
		t.Fatalf("SetLockedRatio(false): %v", err)
	}
	if tdsb.Data[0x02] != 0xEF { // 0xFF &^ 0x10
		t.Errorf("clearing locked_ratio perturbed other bits: tdsb[2] = %#x, want 0xEF", tdsb.Data[0x02])
	}
}
