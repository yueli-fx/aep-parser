package aep_test

import (
	"bytes"
	"math"
	"os"
	"strings"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// Golden values cross-checked against py-aep
// flightdeck/charts/py-aep/samples/models/renderqueue/*.json
//
// Note on time span: numItems_1 uses time_span_source=WORK_AREA_ONLY, so its
// resolved duration depends on the composition's work area. py-aep's synthetic
// fixtures carry an inconsistent cdta (the dividend/divisor duration py-aep
// reads as 10s vs the frame-count field @0xB0 our parser reads as 15s); on
// real AE files these agree. So this test asserts the *resolution* against our
// own comp model rather than py-aep's golden seconds. Exact ldat decode is
// covered by the CUSTOM-source test below, which is comp-independent.
func TestRenderQueueReaderNumItems1(t *testing.T) {
	proj, err := aep.Open("../../test_data/rq_numitems_1.aep")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	rq := proj.RenderQueue
	if rq == nil {
		t.Fatal("proj.RenderQueue is nil, want non-nil")
	}
	if got := rq.NumItems(); got != 1 {
		t.Fatalf("NumItems() = %d, want 1", got)
	}

	item := rq.Items[0]
	if item.Comp == nil {
		t.Fatal("item.Comp is nil")
	}
	if item.Comp.Name != "TestComp" {
		t.Errorf("Comp.Name = %q, want TestComp", item.Comp.Name)
	}
	if item.Comment != "" {
		t.Errorf("Comment = %q, want empty", item.Comment)
	}
	// WORK_AREA_ONLY resolution: (WorkAreaStart, WorkAreaEnd-WorkAreaStart).
	if item.TimeSpanStart != item.Comp.WorkAreaStart {
		t.Errorf("TimeSpanStart = %v, want %v (comp work area start)", item.TimeSpanStart, item.Comp.WorkAreaStart)
	}
	wantDur := item.Comp.WorkAreaEnd - item.Comp.WorkAreaStart
	if item.TimeSpanDuration != wantDur {
		t.Errorf("TimeSpanDuration = %v, want %v (comp work area duration)", item.TimeSpanDuration, wantDur)
	}
	if got := item.NumOutputModules(); got != 1 {
		t.Fatalf("NumOutputModules() = %d, want 1", got)
	}

	om := item.OutputModules[0]
	if om.Name != "H.264 - Match Render Settings - 15 Mbps" {
		t.Errorf("OM.Name = %q, want H.264 - Match Render Settings - 15 Mbps", om.Name)
	}
	if om.FileTemplate != "[compName].[fileextension]" {
		t.Errorf("OM.FileTemplate = %q, want [compName].[fileextension]", om.FileTemplate)
	}
	if om.FullPath == "" {
		t.Error("OM.FullPath is empty, want a path from alas")
	}
}

// CUSTOM time span source: start/duration come straight from the settings
// ldat dividends, independent of comp parsing. 24s13f @ 24fps = 24 + 13/24.
func TestRenderQueueReaderCustomTimeSpan(t *testing.T) {
	proj, err := aep.Open("../../test_data/rq_custom_timespan.aep")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if proj.RenderQueue == nil || proj.RenderQueue.NumItems() != 1 {
		t.Fatalf("want 1 render queue item")
	}
	item := proj.RenderQueue.Items[0]
	if item.TimeSpanStart != 0 {
		t.Errorf("TimeSpanStart = %v, want 0", item.TimeSpanStart)
	}
	const want = 24 + 13.0/24.0 // 24.5416666...
	if math.Abs(item.TimeSpanDuration-want) > 1e-6 {
		t.Errorf("TimeSpanDuration = %v, want %v", item.TimeSpanDuration, want)
	}
}

func TestRenderQueueReaderComment(t *testing.T) {
	proj, err := aep.Open("../../test_data/rq_comment.aep")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if proj.RenderQueue == nil || proj.RenderQueue.NumItems() != 1 {
		t.Fatalf("want 1 render queue item")
	}
	if got := proj.RenderQueue.Items[0].Comment; got != "aaaaa" {
		t.Errorf("Comment = %q, want aaaaa", got)
	}
}

func TestRenderQueueReaderNumItems2(t *testing.T) {
	proj, err := aep.Open("../../test_data/rq_numitems_2.aep")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if proj.RenderQueue == nil {
		t.Fatal("proj.RenderQueue is nil")
	}
	if got := proj.RenderQueue.NumItems(); got != 2 {
		t.Errorf("NumItems() = %d, want 2", got)
	}
}

func TestRenderQueueJSON(t *testing.T) {
	proj, err := aep.Open("../../test_data/rq_numitems_1.aep")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	var sb strings.Builder
	if err := proj.WriteJSON(&sb); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	out := sb.String()
	for _, want := range []string{
		`"render_queue"`,
		`"num_items": 1`,
		`"comp_name": "TestComp"`,
		`"file_template": "[compName].[fileextension]"`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("JSON missing %s\n%s", want, out)
		}
	}
}

// Reading the render queue must not perturb byte-identical round-trip
// (opaque preservation): the LRdr subtree is untouched leaves/lists.
func TestRenderQueueRoundTripByteIdentical(t *testing.T) {
	const path = "../../test_data/rq_numitems_1.aep"
	orig, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	proj, err := aep.FromReader(bytes.NewReader(orig))
	if err != nil {
		t.Fatalf("FromReader: %v", err)
	}
	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	if !bytes.Equal(buf.Bytes(), orig) {
		t.Errorf("roundtrip not byte-identical: in=%d out=%d", len(orig), buf.Len())
	}
}

// Render settings (slice-2). Golden NUMBER values from
// renderqueue/numItems_1.json "settings" — these enum values equal the raw
// binary values (0xFFFF -> -1 sentinel for "current settings").
func TestRenderQueueRenderSettings(t *testing.T) {
	proj, err := aep.Open("../../test_data/rq_numitems_1.aep")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if proj.RenderQueue == nil || proj.RenderQueue.NumItems() != 1 {
		t.Fatalf("want 1 render queue item")
	}
	item := proj.RenderQueue.Items[0]
	rs := item.RenderSettings
	cases := []struct {
		name string
		got  int
		want int
	}{
		{"Quality", rs.Quality, 2},
		{"ColorDepth", rs.ColorDepth, -1},
		{"Effects", rs.Effects, 2},
		{"FieldRender", rs.FieldRender, 0},
		{"FrameBlending", rs.FrameBlending, 1},
		{"MotionBlur", rs.MotionBlur, 1},
		{"ProxyUse", rs.ProxyUse, 0},
		{"Pulldown", rs.Pulldown, 0},
		{"SoloSwitches", rs.SoloSwitches, 2},
		{"GuideLayers", rs.GuideLayers, 0},
		{"DiskCache", rs.DiskCache, 0},
		{"FrameRate", rs.FrameRate, 0},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("RenderSettings.%s = %d, want %d", c.name, c.got, c.want)
		}
	}
	if rs.Resolution != [2]int{1, 1} {
		t.Errorf("Resolution = %v, want [1 1]", rs.Resolution)
	}
	if rs.SkipExistingFiles {
		t.Error("SkipExistingFiles = true, want false")
	}
	if item.QueueItemNotify {
		t.Error("QueueItemNotify = true, want false")
	}
	if item.ElapsedSeconds != 0 {
		t.Errorf("ElapsedSeconds = %d, want 0", item.ElapsedSeconds)
	}
}

// Cross-check render-settings decode against fixtures with distinct values —
// guards against offset swaps (e.g. Quality vs ColorDepth).
func TestRenderQueueRenderSettingsVariants(t *testing.T) {
	cases := []struct {
		fixture                 string
		quality, colorDepth, mb int
		resolution              [2]int
	}{
		{"rq_quality_wireframe.aep", 0, -1, 1, [2]int{1, 1}},
		{"rq_color_depth_8.aep", 2, 0, 1, [2]int{1, 1}},
		{"rq_resolution_half.aep", 2, -1, 1, [2]int{2, 2}},
		{"rq_motion_blur_off.aep", 2, -1, 0, [2]int{1, 1}},
	}
	for _, c := range cases {
		t.Run(c.fixture, func(t *testing.T) {
			proj, err := aep.Open("../../test_data/" + c.fixture)
			if err != nil {
				t.Fatalf("Open: %v", err)
			}
			if proj.RenderQueue == nil || proj.RenderQueue.NumItems() == 0 {
				t.Fatal("no render queue items")
			}
			rs := proj.RenderQueue.Items[0].RenderSettings
			if rs.Quality != c.quality {
				t.Errorf("Quality = %d, want %d", rs.Quality, c.quality)
			}
			if rs.ColorDepth != c.colorDepth {
				t.Errorf("ColorDepth = %d, want %d", rs.ColorDepth, c.colorDepth)
			}
			if rs.MotionBlur != c.mb {
				t.Errorf("MotionBlur = %d, want %d", rs.MotionBlur, c.mb)
			}
			if rs.Resolution != c.resolution {
				t.Errorf("Resolution = %v, want %v", rs.Resolution, c.resolution)
			}
		})
	}
}

// Output module settings (slice-3): 128B OutputModuleSettingsItem + 154B Roou.
// Golden from renderqueue/numItems_1.json outputModules[0].settings.
func TestRenderQueueOutputModuleSettings(t *testing.T) {
	proj, err := aep.Open("../../test_data/rq_numitems_1.aep")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if proj.RenderQueue == nil || proj.RenderQueue.NumItems() != 1 {
		t.Fatalf("want 1 render queue item")
	}
	om := proj.RenderQueue.Items[0].OutputModules[0]
	s := om.Settings
	if s.Depth != 24 {
		t.Errorf("Depth = %d, want 24", s.Depth)
	}
	if s.AudioSampleRate != 48000 {
		t.Errorf("AudioSampleRate = %v, want 48000", s.AudioSampleRate)
	}
	if s.Channels != 0 {
		t.Errorf("Channels = %d, want 0", s.Channels)
	}
	if s.Crop || s.CropTop != 0 || s.CropLeft != 0 || s.CropBottom != 0 || s.CropRight != 0 {
		t.Errorf("Crop = %v top/left/bottom/right = %d/%d/%d/%d, want false 0/0/0/0",
			s.Crop, s.CropTop, s.CropLeft, s.CropBottom, s.CropRight)
	}
	if !s.LockAspectRatio {
		t.Error("LockAspectRatio = false, want true")
	}
	if !s.IncludeProjectLink {
		t.Error("IncludeProjectLink = false, want true")
	}
	if s.Resize {
		t.Error("Resize = true, want false")
	}
	if s.ResizeQuality != 1 {
		t.Errorf("ResizeQuality = %d, want 1", s.ResizeQuality)
	}
	if !s.UseCompFrameNumber {
		t.Error("UseCompFrameNumber = false, want true")
	}
	if s.UseRegionOfInterest {
		t.Error("UseRegionOfInterest = true, want false")
	}
	if s.IncludeSourceXMP {
		t.Error("IncludeSourceXMP = true, want false")
	}
	if s.PostRenderAction != 0 {
		t.Errorf("PostRenderAction = %d, want 0", s.PostRenderAction)
	}
	if s.StartingNumber != 0 {
		t.Errorf("StartingNumber = %d, want 0", s.StartingNumber)
	}
	if !s.VideoOutput {
		t.Error("VideoOutput = false, want true")
	}
	if s.FormatID != "H264" {
		t.Errorf("FormatID = %q, want H264", s.FormatID)
	}
}

// Orthogonal flag discriminators guard the 128B flag-byte bit decoding.
func TestRenderQueueOutputModuleFlags(t *testing.T) {
	cases := []struct {
		fixture            string
		lockAspect, useCFN bool
	}{
		{"rq_lock_aspect_off.aep", false, true},
		{"rq_use_comp_frame_off.aep", true, false},
	}
	for _, c := range cases {
		t.Run(c.fixture, func(t *testing.T) {
			proj, err := aep.Open("../../test_data/" + c.fixture)
			if err != nil {
				t.Fatalf("Open: %v", err)
			}
			s := proj.RenderQueue.Items[0].OutputModules[0].Settings
			if s.LockAspectRatio != c.lockAspect {
				t.Errorf("LockAspectRatio = %v, want %v", s.LockAspectRatio, c.lockAspect)
			}
			if s.UseCompFrameNumber != c.useCFN {
				t.Errorf("UseCompFrameNumber = %v, want %v", s.UseCompFrameNumber, c.useCFN)
			}
		})
	}
}

func TestRenderQueueReaderEmpty(t *testing.T) {
	proj, err := aep.Open("../../test_data/rq_empty.aep")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if proj.RenderQueue == nil {
		t.Fatal("proj.RenderQueue is nil, want non-nil with 0 items")
	}
	if got := proj.RenderQueue.NumItems(); got != 0 {
		t.Errorf("NumItems() = %d, want 0", got)
	}
}
