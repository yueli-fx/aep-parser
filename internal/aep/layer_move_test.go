package aep_test

import (
	"strings"
	"testing"
)

// Layer-level move wrappers delegate to Composition.MoveLayer; tests
// here verify the index-arithmetic in each wrapper rather than re-test
// the underlying splice machinery.

func TestLayerMoveToBeginning(t *testing.T) {
	proj := openMoveBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	pre := layerNames(c)
	if err := c.Layers[2].MoveToBeginning(); err != nil {
		t.Fatalf("MoveToBeginning: %v", err)
	}
	want := []string{pre[2], pre[0], pre[1]}
	if got := layerNames(c); !equalStrings(got, want) {
		t.Errorf("layer order: got %v, want %v", got, want)
	}
}

func TestLayerMoveToEnd(t *testing.T) {
	proj := openMoveBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	pre := layerNames(c)
	if err := c.Layers[0].MoveToEnd(); err != nil {
		t.Fatalf("MoveToEnd: %v", err)
	}
	want := []string{pre[1], pre[2], pre[0]}
	if got := layerNames(c); !equalStrings(got, want) {
		t.Errorf("layer order: got %v, want %v", got, want)
	}
}

func TestLayerMoveAfter_FromBefore(t *testing.T) {
	// l (idx 0) is before other (idx 2). After MoveAfter(other),
	// l lands at idx 2 (right after other's new position, which
	// shifts to idx 1 after the cut).
	proj := openMoveBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	pre := layerNames(c)
	l, other := c.Layers[0], c.Layers[2]
	if err := l.MoveAfter(other); err != nil {
		t.Fatalf("MoveAfter: %v", err)
	}
	want := []string{pre[1], pre[2], pre[0]}
	if got := layerNames(c); !equalStrings(got, want) {
		t.Errorf("layer order: got %v, want %v", got, want)
	}
}

func TestLayerMoveAfter_FromAfter(t *testing.T) {
	// l (idx 2) is after other (idx 0). After MoveAfter(other),
	// l lands at idx 1 (right after other; other stays at 0).
	proj := openMoveBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	pre := layerNames(c)
	l, other := c.Layers[2], c.Layers[0]
	if err := l.MoveAfter(other); err != nil {
		t.Fatalf("MoveAfter: %v", err)
	}
	want := []string{pre[0], pre[2], pre[1]}
	if got := layerNames(c); !equalStrings(got, want) {
		t.Errorf("layer order: got %v, want %v", got, want)
	}
}

func TestLayerMoveBefore_FromAfter(t *testing.T) {
	// l (idx 2) is after other (idx 0). After MoveBefore(other),
	// l lands at idx 0; other shifts to 1.
	proj := openMoveBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	pre := layerNames(c)
	l, other := c.Layers[2], c.Layers[0]
	if err := l.MoveBefore(other); err != nil {
		t.Fatalf("MoveBefore: %v", err)
	}
	want := []string{pre[2], pre[0], pre[1]}
	if got := layerNames(c); !equalStrings(got, want) {
		t.Errorf("layer order: got %v, want %v", got, want)
	}
}

func TestLayerMoveBefore_FromBefore(t *testing.T) {
	// l (idx 0) is before other (idx 2). After MoveBefore(other),
	// l lands at idx 1 (just before other's new position 2).
	proj := openMoveBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	pre := layerNames(c)
	l, other := c.Layers[0], c.Layers[2]
	if err := l.MoveBefore(other); err != nil {
		t.Fatalf("MoveBefore: %v", err)
	}
	want := []string{pre[1], pre[0], pre[2]}
	if got := layerNames(c); !equalStrings(got, want) {
		t.Errorf("layer order: got %v, want %v", got, want)
	}
}

func TestLayerMove_RefuseSelf(t *testing.T) {
	proj := openMoveBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	l := c.Layers[1]
	if err := l.MoveAfter(l); err == nil {
		t.Error("MoveAfter(self): want error, got nil")
	}
	if err := l.MoveBefore(l); err == nil {
		t.Error("MoveBefore(self): want error, got nil")
	}
}

func TestLayerMove_RefuseNilOther(t *testing.T) {
	proj := openMoveBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	l := c.Layers[1]
	if err := l.MoveAfter(nil); err == nil {
		t.Error("MoveAfter(nil): want error, got nil")
	}
	if err := l.MoveBefore(nil); err == nil {
		t.Error("MoveBefore(nil): want error, got nil")
	}
}

func TestLayerMove_RefuseCrossComp(t *testing.T) {
	projA := openMoveBaseline(t)
	if projA == nil {
		return
	}
	projB := openMoveBaseline(t)
	if projB == nil {
		return
	}
	la := projA.Compositions[0].Layers[0]
	lb := projB.Compositions[0].Layers[0]
	if err := la.MoveAfter(lb); err == nil {
		t.Error("MoveAfter cross-comp: want error, got nil")
	} else if !strings.Contains(err.Error(), "different comp") {
		t.Errorf("unexpected error: %v", err)
	}
}
