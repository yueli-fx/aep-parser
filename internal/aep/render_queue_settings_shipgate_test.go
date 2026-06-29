// internal/aep/render_queue_settings_shipgate_test.go
//
// AE ship gate (acceptance + resave-preservation) for the render-queue value
// setters — the ~38 RenderQueueItem render-settings + OutputModule setters that
// patch fixed-width binary fields and were tagged tier=alpha verify=roundtrip
// ("AE acceptance is presumed but not yet validated in-app"). This gate validates
// the presumption: it opens an AE-2020-native RQ fixture, sets every value setter
// to a NON-default, WriteAEP, and has AE open the result (acceptance — no corrupt/
// data-loss) and resave. The Go side then re-parses the resave and checks which
// fields AE preserved byte-stably. RQ binary settings have no reliable cross-
// version ScriptingAPI readback, so acceptance + resave-preservation is the
// ceiling (same model as the SetComment gate). Gated by AE_SHIP_GATE.
package aep_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

const rqSettingsFixture = "../../test_data/rq_ae2020_base.aep"

func runRQSettingsGate(t *testing.T, aeExe, ver string) {
	t.Helper()
	if os.Getenv("AE_SHIP_GATE") == "" {
		t.Skip("set AE_SHIP_GATE=1 with AE installed to run")
	}
	proj, err := aep.Open(rqSettingsFixture)
	if err != nil {
		t.Skipf("%s not present: %v", rqSettingsFixture, err)
	}
	if proj.RenderQueue == nil || proj.RenderQueue.NumItems() < 1 {
		t.Fatalf("%s: need >= 1 render queue item", rqSettingsFixture)
	}
	it := proj.RenderQueue.Items[0]
	if it.NumOutputModules() < 1 {
		t.Fatalf("%s: need >= 1 output module", rqSettingsFixture)
	}
	om := it.OutputModules[0]

	// --- set every value setter to a non-default ---
	it.SetQuality(0)
	it.SetColorDepth(2)
	it.SetEffects(0)
	it.SetFieldRender(1)
	it.SetPulldown(3)
	it.SetFrameBlending(0)
	it.SetMotionBlur(0)
	it.SetProxyUse(1)
	it.SetSoloSwitches(2)
	it.SetGuideLayers(2)
	it.SetDiskCache(2)
	it.SetFrameRate(1)
	it.SetResolution(2, 2)
	it.SetSkipExistingFiles(true)
	it.SetName("ShipGateTpl")
	it.SetQueueItemNotify(true)
	it.SetLogType(1)
	it.SetTimeSpanStart(0.5)
	it.SetTimeSpanDuration(1.5)

	om.SetChannels(1)
	om.SetResizeQuality(1)
	om.SetResize(true)
	om.SetLockAspectRatio(false)
	om.SetCrop(true)
	om.SetCropTop(4)
	om.SetCropLeft(8)
	om.SetCropBottom(12)
	om.SetCropRight(16)
	om.SetIncludeProjectLink(true)
	om.SetPostRenderAction(1)
	om.SetOutputAudio(1)
	om.SetConvertToLinear(1)
	om.SetUseCompFrameNumber(true)
	om.SetUseRegionOfInterest(true)
	om.SetIncludeSourceXMP(true)
	om.SetPreserveRGB(true)
	om.SetDepth(64)
	om.SetStartingNumber(42)

	const argsPath = `e:/projects/tools/aep-parser/test_data/rq_settings_args.json`
	const jsxPath = `E:/projects/tools/aep-parser/test_data/verify_rq_settings.jsx`
	toFwd := func(p string) string { return strings.ReplaceAll(p, `\`, `/`) }

	tempDir := t.TempDir()
	inputAEP := filepath.Join(tempDir, "rq_settings_in.aep")
	resavedAEP := filepath.Join(tempDir, "rq_settings_resaved.aep")
	doneFile := filepath.Join(tempDir, "rq_settings.done")

	out, err := os.Create(inputAEP)
	if err != nil {
		t.Fatal(err)
	}
	if err := proj.WriteAEP(out); err != nil {
		out.Close()
		t.Fatalf("WriteAEP: %v", err)
	}
	out.Close()

	argsJSON := fmt.Sprintf(`{"input":%q,"done":%q,"resaved":%q}`,
		toFwd(inputAEP), toFwd(doneFile), toFwd(resavedAEP))
	if err := os.WriteFile(argsPath, []byte(argsJSON), 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(argsPath)
	os.Remove(doneFile)

	runAeRunShipGate(t, aeExe, jsxPath, doneFile, 180)

	content, err := os.ReadFile(doneFile)
	if err != nil {
		t.Fatal(err)
	}
	body := string(content)
	t.Logf("rq settings %s AE open:\n%s", ver, body)
	if lines := strings.SplitN(body, "\n", 2); len(lines) == 0 || strings.TrimSpace(lines[0]) != "PASS" {
		t.Fatalf("rq settings %s acceptance FAIL:\n%s", ver, body)
	}

	// --- preservation: re-parse AE's resave, report expected vs got per field ---
	re, err := aep.Open(resavedAEP)
	if err != nil {
		t.Fatalf("re-open AE resave: %v", err)
	}
	if re.RenderQueue == nil || re.RenderQueue.NumItems() < 1 || re.RenderQueue.Items[0].NumOutputModules() < 1 {
		t.Fatalf("%s resave: RQ item/output module missing (AE dropped it)", ver)
	}
	ri := re.RenderQueue.Items[0]
	ro := ri.OutputModules[0]
	rs := ri.RenderSettings
	oms := ro.Settings

	var mism []string
	chk := func(name string, want, got any) {
		if fmt.Sprint(want) != fmt.Sprint(got) {
			mism = append(mism, fmt.Sprintf("%s: want %v got %v", name, want, got))
		}
	}
	chk("Quality", 0, rs.Quality)
	chk("ColorDepth", 2, rs.ColorDepth)
	chk("Effects", 0, rs.Effects)
	chk("FieldRender", 1, rs.FieldRender)
	chk("Pulldown", 3, rs.Pulldown)
	chk("FrameBlending", 0, rs.FrameBlending)
	chk("MotionBlur", 0, rs.MotionBlur)
	chk("ProxyUse", 1, rs.ProxyUse)
	chk("SoloSwitches", 2, rs.SoloSwitches)
	chk("GuideLayers", 2, rs.GuideLayers)
	chk("DiskCache", 2, rs.DiskCache)
	chk("FrameRate", 1, rs.FrameRate)
	chk("Resolution", [2]int{2, 2}, rs.Resolution)
	chk("SkipExistingFiles", true, rs.SkipExistingFiles)
	chk("Name", "ShipGateTpl", ri.Name)
	chk("LogType", uint16(1), ri.LogType)
	chk("TimeSpanStart", 0.5, ri.TimeSpanStart)
	chk("TimeSpanDuration", 1.5, ri.TimeSpanDuration)
	chk("Channels", 1, oms.Channels)
	chk("ResizeQuality", 1, oms.ResizeQuality)
	chk("Resize", true, oms.Resize)
	chk("LockAspectRatio", false, oms.LockAspectRatio)
	chk("Crop", true, oms.Crop)
	chk("CropTop", 4, oms.CropTop)
	chk("CropLeft", 8, oms.CropLeft)
	chk("CropBottom", 12, oms.CropBottom)
	chk("CropRight", 16, oms.CropRight)
	chk("IncludeProjectLink", true, oms.IncludeProjectLink)
	chk("PostRenderAction", uint32(1), oms.PostRenderAction)
	chk("OutputAudio", 1, oms.OutputAudio)
	chk("ConvertToLinear", 1, oms.ConvertToLinear)
	chk("UseCompFrameNumber", true, oms.UseCompFrameNumber)
	chk("UseRegionOfInterest", true, oms.UseRegionOfInterest)
	chk("IncludeSourceXMP", true, oms.IncludeSourceXMP)
	chk("Depth", 64, oms.Depth)
	chk("StartingNumber", uint32(42), oms.StartingNumber)
	// PreserveRGB and QueueItemNotify are NOT asserted — AE normalizes them on
	// resave (AE2025 clears PreserveRGB; AE2020 clears QueueItemNotify), an AE-
	// semantics reset gated by output color-management / queue context, not a write
	// bug. The bytes are still WRITTEN above (proving AE accepts a file carrying
	// them); both stay verify=roundtrip with a documented caveat. The other 36
	// fields preserve byte-stably through both versions' resave.
	if oms.PreserveRGB || ri.QueueItemNotify {
		t.Logf("rq settings %s: PreserveRGB=%v QueueItemNotify=%v (a normalizer was preserved — AE behavior changed?)", ver, oms.PreserveRGB, ri.QueueItemNotify)
	}

	if len(mism) > 0 {
		t.Errorf("rq settings %s: %d/36 fields NOT preserved through AE resave:\n  %s",
			ver, len(mism), strings.Join(mism, "\n  "))
	} else {
		t.Logf("rq settings %s: all 36 asserted fields preserved through AE resave", ver)
	}
}

func TestRenderQueueSettings_AEShipGate_AE2020(t *testing.T) {
	runRQSettingsGate(t, ae2020(), "AE2020")
}
func TestRenderQueueSettings_AEShipGate_AE2025(t *testing.T) {
	runRQSettingsGate(t, ae2025(), "AE2025")
}
