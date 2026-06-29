package serializer

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/codec"
	"github.com/yueli-fx/aep-parser/internal/scene"
)

// TestParseAttachCompleteness asserts that after Open() every scene object that
// should carry a back-ref has its back field set (non-nil). This is the M8 P2
// guard: written before the back-field inversion so that any "lost attach" bug
// introduced during P2 is caught immediately (C-4 / spec §2.3).
//
// The test PASSES today because Parse already attaches everything. If it ever
// FAILS, that means Parse has a real missing-attach bug — do not paper over.
func TestParseAttachCompleteness(t *testing.T) {
	proj, err := Open("../../test_data/re_compmarker.aep")
	if err != nil {
		t.Fatal(err)
	}
	walkAndAssertAttached(t, proj)
}

// walkAndAssertAttached traverses the reachable object graph and asserts each
// back field is non-nil. Uses t.Errorf so all failures are reported in one run.
func walkAndAssertAttached(t *testing.T, proj *Project) {
	t.Helper()

	nComps, nLayers, nProps, nKeyframes, nMarkers, nMasks, nFootage := 0, 0, 0, 0, 0, 0, 0

	if scene.ProjectBack(proj) == nil {
		t.Errorf("Project.back is nil")
	}

	for _, f := range proj.Footage {
		nFootage++
		if scene.FootageBack(f) == nil {
			t.Errorf("Footage %q (ID=%d): back is nil", f.Name, f.ID)
		}
	}

	for _, comp := range proj.Compositions {
		nComps++
		if scene.CompositionBack(comp) == nil {
			t.Errorf("Composition %q (ID=%d): back is nil", comp.Name, comp.ID)
		}

		for _, m := range comp.Markers {
			nMarkers++
			if scene.MarkerBack(m) == nil {
				t.Errorf("Composition %q marker @t=%.3fs: back is nil", comp.Name, m.Time)
			}
		}

		for _, layer := range comp.Layers {
			nLayers++
			if scene.LayerBack(layer) == nil {
				t.Errorf("Layer %q (ID=%d) in comp %q: back is nil", layer.Name, layer.ID, comp.Name)
			}

			for _, lm := range layer.Markers {
				nMarkers++
				if scene.MarkerBack(lm) == nil {
					t.Errorf("Layer %q marker @t=%.3fs: back is nil", layer.Name, lm.Time)
				}
			}

			for _, mask := range layer.Masks {
				nMasks++
				if scene.MaskBack(mask) == nil {
					t.Errorf("Layer %q mask %q (index=%d): back is nil", layer.Name, mask.Name, mask.Index)
				}
				for _, mp := range mask.Properties {
					nProps++
					if scene.PropertyBack(mp) == nil {
						t.Errorf("Layer %q mask %q property %q: back is nil", layer.Name, mask.Name, mp.MatchName)
					}
					for _, kf := range mp.Keyframes {
						nKeyframes++
						if scene.KeyframeBack(kf) == nil {
							t.Errorf("Layer %q mask %q property %q keyframe @t=%.3fs: back is nil",
								layer.Name, mask.Name, mp.MatchName, kf.Time)
						}
					}
				}
			}

			for _, prop := range layer.Properties {
				assertPropertyAttached(t, layer.Name, prop, &nProps, &nKeyframes)
			}
		}
	}

	t.Logf("coverage: %d comps, %d layers, %d properties, %d keyframes, %d markers, %d masks, %d footage",
		nComps, nLayers, nProps, nKeyframes, nMarkers, nMasks, nFootage)
}

// TestParseAttachCompletenessRenderQueue asserts the M8 P2 render-queue /
// output-module back-refs are attached after Open: each item's scene-owned
// settings copy and its chunk-aliasing back.settingsSlice are both present and
// equal length. syncRenderQueue silently no-ops on a length mismatch, so a lost
// attach would not be localized by the byte-identical round-trip alone.
func TestParseAttachCompletenessRenderQueue(t *testing.T) {
	proj, err := Open("../../test_data/rq_numitems_1.aep")
	if err != nil {
		t.Fatal(err)
	}
	rq := proj.RenderQueue
	if rq == nil || rq.NumItems() == 0 {
		t.Fatal("fixture has no render queue items")
	}
	for i, it := range rq.Items {
		rb := renderQueueItemBack(it)
		if rb == nil {
			t.Errorf("RenderQueueItem[%d]: back is nil", i)
			continue
		}
		if len(scene.RenderQueueItemSettings(it)) != codec.RenderSettingsItemSize {
			t.Errorf("RenderQueueItem[%d]: settingsBlock copy len=%d, want %d",
				i, len(scene.RenderQueueItemSettings(it)), codec.RenderSettingsItemSize)
		}
		if len(rb.settingsSlice) != len(scene.RenderQueueItemSettings(it)) {
			t.Errorf("RenderQueueItem[%d]: back.settingsSlice len=%d != copy len=%d",
				i, len(rb.settingsSlice), len(scene.RenderQueueItemSettings(it)))
		}
		for j, om := range it.OutputModules {
			ob := outputModuleBack(om)
			if ob == nil {
				t.Errorf("RenderQueueItem[%d] OutputModule[%d]: back is nil", i, j)
				continue
			}
			if len(scene.OutputModuleSettingsBlock(om)) != len(ob.settingsSlice) {
				t.Errorf("RenderQueueItem[%d] OutputModule[%d]: settings copy/alias len %d != %d",
					i, j, len(scene.OutputModuleSettingsBlock(om)), len(ob.settingsSlice))
			}
			if len(scene.OutputModuleRoouData(om)) != len(ob.roouSlice) {
				t.Errorf("RenderQueueItem[%d] OutputModule[%d]: roou copy/alias len %d != %d",
					i, j, len(scene.OutputModuleRoouData(om)), len(ob.roouSlice))
			}
		}
	}
}

// TestParseAttachCompletenessGuides asserts every parsed guide carries its
// scene-owned 16B copy (syncGuides pairs guides to the live Gide ldat by index
// at WriteAEP time — a missing copy would be silently skipped).
func TestParseAttachCompletenessGuides(t *testing.T) {
	proj, err := Open("../../test_data/guides.aep")
	if err != nil {
		t.Fatal(err)
	}
	total := 0
	for _, c := range proj.Compositions {
		for i, g := range c.Guides {
			total++
			if len(scene.GuideBlock(g)) != codec.GuideItemSize {
				t.Errorf("comp %q guide[%d]: block copy len=%d, want %d",
					c.Name, i, len(scene.GuideBlock(g)), codec.GuideItemSize)
			}
		}
	}
	if total == 0 {
		t.Fatal("fixture has no guides")
	}
}

// assertPropertyAttached recursively checks a Property and all its keyframes.
func assertPropertyAttached(t *testing.T, layerName string, prop *Property, nProps, nKeyframes *int) {
	t.Helper()
	*nProps++
	if scene.PropertyBack(prop) == nil {
		t.Errorf("Layer %q property %q: back is nil", layerName, prop.MatchName)
	}
	for _, kf := range prop.Keyframes {
		*nKeyframes++
		if scene.KeyframeBack(kf) == nil {
			t.Errorf("Layer %q property %q keyframe @t=%.3fs: back is nil",
				layerName, prop.MatchName, kf.Time)
		}
	}
}
