package aep_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
	"github.com/yueli-fx/aep-parser/internal/aeptest"
)

// Tests reuse the DeleteLayer baseline fixture (re_delete_layer_baseline.aep,
// 3 solid layers L1_top / L2_mid / L3_bot, no refs) — pure structural
// reorder doesn't depend on layer types or references.

const moveLayerFixtureDir = "../../test_data/generated/fixtures"

func openMoveBaseline(t *testing.T) *aep.Project {
	t.Helper()
	path := filepath.Join(moveLayerFixtureDir, "re_delete_layer_baseline.aep")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Skipf("fixture missing: %s (3-solid baseline; produced by test_data/generators/re_delete_layer.jsx RE_DELETE_MODE=baseline)", path)
		return nil
	}
	proj, err := aep.Open(path)
	if err != nil {
		t.Fatalf("aep.Open(%s): %v", path, err)
	}
	return proj
}

func TestMoveLayer_RefuseFromOutOfRange(t *testing.T) {
	proj := openMoveBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	for _, from := range []int{-1, len(c.Layers), len(c.Layers) + 5} {
		if err := aep.MoveLayer(c, from, 0); err == nil {
			t.Errorf("MoveLayer(%d, 0): want error, got nil", from)
		}
	}
}

func TestMoveLayer_RefuseToOutOfRange(t *testing.T) {
	proj := openMoveBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	for _, to := range []int{-1, len(c.Layers), len(c.Layers) + 5} {
		if err := aep.MoveLayer(c, 0, to); err == nil {
			t.Errorf("MoveLayer(0, %d): want error, got nil", to)
		}
	}
}

func TestMoveLayer_RefuseMissingBackref(t *testing.T) {
	// Composition built outside the parser has c.back == nil.
	c := &aep.Composition{Name: "synthetic"}
	if err := aep.MoveLayer(c, 0, 1); err == nil {
		t.Error("MoveLayer on backref-less comp: want error, got nil")
	} else if !strings.Contains(err.Error(), "itemList") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestMoveLayer_NoOpSameIndex(t *testing.T) {
	proj := openMoveBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	pre := layerNames(c)
	preChildCount := len(aeptest.ItemList(c).Children)
	if err := aep.MoveLayer(c, 1, 1); err != nil {
		t.Fatalf("MoveLayer(1, 1): %v", err)
	}
	post := layerNames(c)
	if !equalStrings(pre, post) {
		t.Errorf("no-op changed layer order: pre=%v post=%v", pre, post)
	}
	if got := len(aeptest.ItemList(c).Children); got != preChildCount {
		t.Errorf("no-op changed itemList children count: pre=%d post=%d", preChildCount, got)
	}
}

func TestMoveLayer_HappyPath_FirstToLast(t *testing.T) {
	proj := openMoveBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	preChildCount := len(aeptest.ItemList(c).Children)
	pre := layerNames(c)
	if len(pre) != 3 {
		t.Fatalf("baseline must have 3 layers, got %d", len(pre))
	}

	if err := aep.MoveLayer(c, 0, 2); err != nil {
		t.Fatalf("MoveLayer(0, 2): %v", err)
	}

	want := []string{pre[1], pre[2], pre[0]}
	got := layerNames(c)
	if !equalStrings(got, want) {
		t.Errorf("layer order: got %v, want %v", got, want)
	}
	if postCount := len(aeptest.ItemList(c).Children); postCount != preChildCount {
		t.Errorf("itemList children count changed: pre=%d post=%d (move should preserve)", preChildCount, postCount)
	}
	for i, l := range c.Layers {
		if l.Index != i {
			t.Errorf("c.Layers[%d].Index = %d, want %d", i, l.Index, i)
		}
	}
}

func TestMoveLayer_HappyPath_LastToFirst(t *testing.T) {
	proj := openMoveBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	preChildCount := len(aeptest.ItemList(c).Children)
	pre := layerNames(c)

	if err := aep.MoveLayer(c, 2, 0); err != nil {
		t.Fatalf("MoveLayer(2, 0): %v", err)
	}

	want := []string{pre[2], pre[0], pre[1]}
	got := layerNames(c)
	if !equalStrings(got, want) {
		t.Errorf("layer order: got %v, want %v", got, want)
	}
	if postCount := len(aeptest.ItemList(c).Children); postCount != preChildCount {
		t.Errorf("itemList children count changed: pre=%d post=%d", preChildCount, postCount)
	}
	for i, l := range c.Layers {
		if l.Index != i {
			t.Errorf("c.Layers[%d].Index = %d, want %d", i, l.Index, i)
		}
	}
}

func TestMoveLayer_HappyPath_MidSwap(t *testing.T) {
	proj := openMoveBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	pre := layerNames(c)

	if err := aep.MoveLayer(c, 1, 0); err != nil {
		t.Fatalf("MoveLayer(1, 0): %v", err)
	}

	want := []string{pre[1], pre[0], pre[2]}
	got := layerNames(c)
	if !equalStrings(got, want) {
		t.Errorf("layer order: got %v, want %v", got, want)
	}
}

func TestMoveLayer_RoundTrip(t *testing.T) {
	proj := openMoveBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	pre := layerNames(c)

	if err := aep.MoveLayer(c, 0, 2); err != nil {
		t.Fatalf("MoveLayer(0, 2): %v", err)
	}
	want := []string{pre[1], pre[2], pre[0]}

	var buf bytes.Buffer
	if err := proj.WriteAEP(&buf); err != nil {
		t.Fatalf("WriteAEP: %v", err)
	}

	tmp := filepath.Join(t.TempDir(), "movelayer_rt.aep")
	if err := os.WriteFile(tmp, buf.Bytes(), 0o644); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	reopen, err := aep.Open(tmp)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if got := layerNames(reopen.Compositions[0]); !equalStrings(got, want) {
		t.Errorf("reopened layer order: got %v, want %v", got, want)
	}
}

func TestMoveLayer_ItemListChildrenIdentical(t *testing.T) {
	// Move must preserve the same set of chunk pointers in itemList.Children
	// (no chunks lost, no chunks created — just reordered).
	proj := openMoveBaseline(t)
	if proj == nil {
		return
	}
	c := proj.Compositions[0]
	preChildren := aeptest.ItemList(c).Children
	pre := make(map[any]struct{}, len(preChildren))
	for _, ch := range preChildren {
		pre[any(ch)] = struct{}{}
	}

	if err := aep.MoveLayer(c, 2, 0); err != nil {
		t.Fatalf("MoveLayer(2, 0): %v", err)
	}

	postChildren := aeptest.ItemList(c).Children
	post := make(map[any]struct{}, len(postChildren))
	for _, ch := range postChildren {
		post[any(ch)] = struct{}{}
	}
	if len(pre) != len(post) {
		t.Fatalf("chunk count: pre=%d post=%d", len(pre), len(post))
	}
	for k := range pre {
		if _, ok := post[k]; !ok {
			t.Error("chunk lost after move")
		}
	}
	for k := range post {
		if _, ok := pre[k]; !ok {
			t.Error("new chunk appeared after move")
		}
	}
}

func layerNames(c *aep.Composition) []string {
	out := make([]string, len(c.Layers))
	for i, l := range c.Layers {
		out[i] = l.Name
	}
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
