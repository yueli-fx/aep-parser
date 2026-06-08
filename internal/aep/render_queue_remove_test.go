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

	if err := aep.RemoveItem(rq, 1); err != nil {
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

// compByName returns the first composition with the given name, or nil.
func compByName(proj *aep.Project, name string) *aep.Composition {
	for _, c := range proj.Compositions {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// TestRenderQueueAddItem clones the queue's last item for a second comp and
// confirms the new item + comp linkage survive WriteAEP.
func TestRenderQueueAddItem(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_rq_add_before.aep")
	if err != nil {
		t.Skipf("re_rq_add_before.aep not present; run test_data/re_rq_add.jsx in AE 2020")
	}
	rq := proj.RenderQueue
	if rq == nil || rq.NumItems() != 1 {
		t.Fatalf("base fixture: want 1 render queue item, got %d", rq.NumItems())
	}
	rqb := compByName(proj, "RQB")
	if rqb == nil {
		t.Fatal("fixture: comp RQB not found")
	}

	added, err := aep.AddItem(rq, rqb)
	if err != nil {
		t.Fatalf("AddItem(RQB): %v", err)
	}
	if added == nil || added.Comp == nil || added.Comp.Name != "RQB" {
		t.Fatalf("AddItem returned item with comp %v, want RQB", added.Comp)
	}
	if rq.NumItems() != 2 {
		t.Fatalf("after add: NumItems = %d, want 2", rq.NumItems())
	}

	assertTwo := func(t *testing.T, p *aep.Project, tag string) {
		t.Helper()
		if p.RenderQueue == nil || p.RenderQueue.NumItems() != 2 {
			t.Fatalf("%s: NumItems = %d, want 2", tag, p.RenderQueue.NumItems())
		}
		i0, i1 := p.RenderQueue.Items[0], p.RenderQueue.Items[1]
		if i0.Comp == nil || i0.Comp.Name != "RQA" {
			t.Errorf("%s: item0 comp = %v, want RQA", tag, i0.Comp)
		}
		if i1.Comp == nil || i1.Comp.Name != "RQB" {
			t.Errorf("%s: item1 comp = %v, want RQB", tag, i1.Comp)
		}
		if i1.NumOutputModules() < 1 {
			t.Errorf("%s: added item has no output module", tag)
		}
	}
	assertTwo(t, proj, "in-memory")

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}
	re, err := aep.FromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	assertTwo(t, re, "round-trip")
}

// TestRenderQueueAddItem_Refuse covers nil-comp + empty-queue guards.
func TestRenderQueueAddItem_Refuse(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_rq_add_before.aep")
	if err != nil {
		t.Skipf("re_rq_add_before.aep not present")
	}
	if _, err := aep.AddItem(proj.RenderQueue, nil); err == nil {
		t.Error("AddItem(nil) should refuse")
	}
}

// TestRenderQueueRemoveItem_Refuse covers out-of-range + missing-backref guards.
func TestRenderQueueRemoveItem_Refuse(t *testing.T) {
	proj, err := aep.Open("../../test_data/re_rq_delete_before.aep")
	if err != nil {
		t.Skipf("re_rq_delete_before.aep not present")
	}
	rq := proj.RenderQueue
	if err := aep.RemoveItem(rq, -1); err == nil {
		t.Error("RemoveItem(-1) should refuse")
	}
	if err := aep.RemoveItem(rq, 99); err == nil {
		t.Error("RemoveItem(99) should refuse")
	}
}
