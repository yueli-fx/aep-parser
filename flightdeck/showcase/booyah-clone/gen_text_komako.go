// flightdeck/showcase/booyah-clone/gen_text_komako.go — comp ② "テキスト変えるならココ！".
// One text layer "GLITCH" with a glitch-flicker layer transform (28-kf Position
// jitter + 24-kf Opacity flicker) and two text animators: Tracking Amount (2kf
// 135→0) + Character Offset (2kf 21→0). All values pulled from the original via
// the oracle; chunks built via our from-scratch text + animator + transform API.
//
// Two-phase: buildTextKomako creates the layer (NewTextLayer + SetText) before the
// project is reopened; finishTextKomako applies SetLayerTransform + the animators,
// which require a PARSED layer (see the Reopen in main). 30 fps (not the original's
// 29.97) for the same NewComposition fractional-fps cdta reason as comp ①.
package main

import (
	"fmt"

	aep "github.com/example/aep-parser/internal/aep"
)

const komakoCompName = "テキスト変えるならココ！"

func buildTextKomako(p *aep.Project, orc *oracle) {
	comp, err := aep.NewComposition(p, komakoCompName, 1920, 1080, 30, 6)
	must(err)
	tl, err := aep.NewTextLayer(comp, "GLITCH")
	must(err)
	must(tl.SetText("GLITCH"))
}

// finishTextKomako runs after the project is reopened: it operates on the parsed
// "GLITCH" layer (SetLayerTransform + AddText*Animator require a parsed layer).
func finishTextKomako(rp *aep.Project, orc *oracle) {
	comp := rp.CompositionByName(komakoCompName)
	if comp == nil {
		panic("comp ②: not found after reopen")
	}
	layer := comp.LayerByName("GLITCH")
	if layer == nil {
		panic("comp ②: GLITCH layer not found after reopen")
	}

	// Match the original GLITCH text style (RE'd from its text doc via AE DOM:
	// Industry-Demi 110pt, base tracking 65, faux italic, centre-justified). The
	// font + centre justification + size are what let the original's anchor
	// (copied below) place the text dead-centre — without them a default-font
	// left-justified layout lands low. Industry-Demi need not be installed on this
	// box (AE substitutes at render); it resolves on a machine that has it.
	fontIdx, err := layer.AddFont("Industry-Demi")
	must(err)
	must(layer.SetRunFontIndex(0, fontIdx))
	must(layer.SetRunFontSize(0, 110))
	must(layer.SetRunTracking(0, 65))
	must(layer.SetRunFauxItalic(0, true))
	must(layer.SetParagraphJustification(0, aep.TextJustifyCenter))

	orig := orc.mustComp(komakoCompName).Layers[0]
	otg := findGroup(orig.PropertyTree(), "ADBE Transform Group")

	// --- layer transform: anchor (static) + Position (28kf) + Opacity (24kf) ---
	tr := aep.NewLayerTransform()
	if ap := findProp(otg, "ADBE Anchor Point"); ap != nil && ap.StaticValue != nil {
		a := toFloats(ap.StaticValue)
		must(tr.AnchorPoint().SetStaticValue([2]float64{a[0], a[1]}))
	}
	for _, kf := range findProp(otg, "ADBE Position").Keyframes {
		v := toFloats(kf.Value)
		must(tr.Position().AddKeyframeLinear(kf.Time, [2]float64{v[0], v[1]}))
	}
	// Parser reports Opacity normalized 0..1; SetLayerTransform's Opacity is percent.
	for _, kf := range findProp(otg, "ADBE Opacity").Keyframes {
		must(tr.Opacity().AddKeyframeLinear(kf.Time, toScalar(kf.Value)*100))
	}
	must(aep.SetLayerTransform(layer, tr))

	// --- text animators: Tracking Amount + Character Offset (each 2kf, ease→linear) ---
	trackKfs := scalarKfsOf(orig, "ADBE Text Tracking Amount")
	offKfs := scalarKfsOf(orig, "ADBE Text Character Offset")

	_, err = aep.AddTextTrackingAnimator(layer, trackKfs[0].Value, 0, 100, 0)
	must(err)
	must(aep.AnimateTextTracking(layer, 0, trackKfs))

	_, err = aep.AddTextCharacterOffsetAnimator(layer, offKfs[0].Value, 0, 100, 0)
	must(err)
	must(aep.AnimateTextCharacterOffset(layer, 0, offKfs))

	fmt.Printf("  ② テキスト変えるならココ: GLITCH text + Position(%dkf)/Opacity(%dkf) + Tracking(%dkf)/CharOffset(%dkf)\n",
		len(findProp(otg, "ADBE Position").Keyframes), len(findProp(otg, "ADBE Opacity").Keyframes),
		len(trackKfs), len(offKfs))
}

// scalarKfsOf pulls a 1D-scalar animator leaf's keyframes (by match-name, anywhere
// under the layer's Text Animators) into []ScalarKeyframe (linear; the original's
// ease is approximated — key values/times are exact).
func scalarKfsOf(orig *aep.Layer, matchName string) []aep.ScalarKeyframe {
	tp := findGroup(orig.PropertyTree(), "ADBE Text Properties")
	animators := findGroup(tp, "ADBE Text Animators")
	var leaf *aep.Property
	for _, c := range animators.Children {
		ag, ok := c.(*aep.AEPropertyGroup)
		if !ok || ag.MatchName != "ADBE Text Animator" {
			continue
		}
		if p := findProp(findGroup(ag, "ADBE Text Animator Properties"), matchName); p != nil {
			leaf = p
			break
		}
	}
	if leaf == nil {
		panic("comp ②: animator leaf not found: " + matchName)
	}
	kfs := make([]aep.ScalarKeyframe, 0, len(leaf.Keyframes))
	for _, kf := range leaf.Keyframes {
		kfs = append(kfs, aep.ScalarKeyframe{Time: kf.Time, Value: toScalar(kf.Value)})
	}
	return kfs
}
