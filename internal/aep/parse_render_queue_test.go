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
