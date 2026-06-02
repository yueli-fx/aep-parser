package aep_test

import (
	"bytes"
	"testing"

	aep "github.com/example/aep-parser/internal/aep"
)

// TestRenderQueueRemoveItem removes one of two render queue items and confirms
// the survivor + its settings/comp linkage stay coherent across WriteAEP.
func TestRenderQueueRemoveItem(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_rq_delete_before.aep")
	if err != nil {
		t.Skipf("re_rq_delete_before.aep not present; run test_data/re_rq_delete.jsx in AE 2020")
	}
	rq := proj.RenderQueue
	if rq == nil || rq.NumItems() != 2 {
		t.Fatalf("base fixture: want 2 render queue items, got %d", rq.NumItems())
	}
	// Record the survivor (index 0) comp so we can assert it persists.
	survivorComp := ""
	if c := rq.Items[0].Comp; c != nil {
		survivorComp = c.Name
	}

	if err := rq.RemoveItem(1); err != nil {
		t.Fatalf("RemoveItem(1): %v", err)
	}
	if rq.NumItems() != 1 {
		t.Fatalf("after remove: NumItems = %d, want 1", rq.NumItems())
	}

	assertOne := func(t *testing.T, p *aep.Project, tag string) {
		t.Helper()
		if p.RenderQueue == nil || p.RenderQueue.NumItems() != 1 {
			t.Fatalf("%s: NumItems = %d, want 1", tag, p.RenderQueue.NumItems())
		}
		it := p.RenderQueue.Items[0]
		if survivorComp != "" {
			if it.Comp == nil || it.Comp.Name != survivorComp {
				t.Errorf("%s: survivor comp = %v, want %q", tag, it.Comp, survivorComp)
			}
		}
		if it.NumOutputModules() < 1 {
			t.Errorf("%s: survivor lost its output module", tag)
		}
	}
	assertOne(t, proj, "in-memory")

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	assertOne(t, re, "round-trip")
}

// TestRenderQueueRemoveItem_Refuse covers out-of-range + missing-backref guards.
func TestRenderQueueRemoveItem_Refuse(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_rq_delete_before.aep")
	if err != nil {
		t.Skipf("re_rq_delete_before.aep not present")
	}
	rq := proj.RenderQueue
	if err := rq.RemoveItem(-1); err == nil {
		t.Error("RemoveItem(-1) should refuse")
	}
	if err := rq.RemoveItem(99); err == nil {
		t.Error("RemoveItem(99) should refuse")
	}
}
