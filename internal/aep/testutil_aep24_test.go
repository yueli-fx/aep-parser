package aep_test

import (
	"testing"

	"github.com/yueli-fx/aep-parser/internal/aep"
)

// AE 24+ fixture openers. Each one skips the test when its .aep is
// missing so CI / fresh checkouts don't fail before the re_*_ae24.jsx
// outputs have been produced.

func openCapsAE24(t *testing.T) (*aep.Project, *aep.Composition, *aep.Composition) {
	t.Helper()
	proj, err := aep.Open("../../test_data/re_text_caps_ae24.aep")
	if err != nil {
		t.Skipf("re_text_caps_ae24.aep not present; rerun re_text_caps_ae24.jsx in AE 24+")
	}
	var compPt, compBx *aep.Composition
	for _, c := range proj.Compositions {
		switch c.Name {
		case "RE_CAPS_POINT":
			compPt = c
		case "RE_CAPS_BOX":
			compBx = c
		}
	}
	if compPt == nil || compBx == nil {
		t.Fatalf("RE_CAPS_POINT / RE_CAPS_BOX comps missing")
	}
	return proj, compPt, compBx
}

func openTrackMatteAE24(t *testing.T) (*aep.Project, *aep.Composition) {
	t.Helper()
	proj, err := aep.Open("../../test_data/re_trackmatte_ae24.aep")
	if err != nil {
		t.Skipf("re_trackmatte_ae24.aep not present; rerun re_trackmatte_ae24.jsx in AE 23+")
	}
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "RE_TRACKMATTE" {
			comp = c
			break
		}
	}
	if comp == nil {
		t.Fatalf("RE_TRACKMATTE comp missing")
	}
	return proj, comp
}

func openWave2AE24(t *testing.T) *aep.Project {
	t.Helper()
	proj, err := aep.Open("../../test_data/re_wave2_ae24.aep")
	if err != nil {
		t.Skipf("re_wave2_ae24.aep not present; rerun re_wave2_ae24.jsx in AE 24+")
	}
	return proj
}

func openTextAE24More(t *testing.T) *aep.Project {
	t.Helper()
	p, err := aep.Open("../../test_data/re_text_ae24_more.aep")
	if err != nil {
		t.Skipf("re_text_ae24_more.aep not present; rerun re_text_ae24_more.jsx in AE 24+")
	}
	return p
}

func openAlternateSourceAE24(t *testing.T) (*aep.Project, *aep.Composition) {
	t.Helper()
	proj, err := aep.Open("../../test_data/re_altsource_ae24.aep")
	if err != nil {
		t.Skipf("re_altsource_ae24.aep not present; rerun re_altsource_ae24.jsx in AE 18+")
	}
	var comp *aep.Composition
	for _, c := range proj.Compositions {
		if c.Name == "RE_ALT_SOURCE_MAIN" {
			comp = c
			break
		}
	}
	if comp == nil {
		t.Fatalf("RE_ALT_SOURCE_MAIN comp missing")
	}
	return proj, comp
}

// findLayerInComp walks all comps, returns the first layer in `compName`
// whose Name == `layerName`. Nil when not found.
func findLayerInComp(proj *aep.Project, compName, layerName string) *aep.Layer {
	for _, c := range proj.Compositions {
		if c.Name != compName {
			continue
		}
		for _, l := range c.Layers {
			if l.Name == layerName {
				return l
			}
		}
	}
	return nil
}

// layerByName returns the first layer in `comp` whose Name == n, or nil.
// Unlike layerBySourceName this matches the layer's own Utf8 chunk,
// which AE 25 writes for precomp-instance layers we addToMotionGraphics
// even though it omits it for solids/lights/nulls.
func layerByName(comp *aep.Composition, n string) *aep.Layer {
	for _, l := range comp.Layers {
		if l.Name == n {
			return l
		}
	}
	return nil
}

// layerBySourceName returns the layer whose source footage has the
// given name. Used to identify named solids when AE 25 omits Utf8.
func layerBySourceName(proj *aep.Project, comp *aep.Composition, footageName string) *aep.Layer {
	var fid uint32
	for _, f := range proj.Footage {
		if f.Name == footageName {
			fid = f.ID
			break
		}
	}
	if fid == 0 {
		return nil
	}
	for _, l := range comp.Layers {
		if l.SourceID == fid {
			return l
		}
	}
	return nil
}
